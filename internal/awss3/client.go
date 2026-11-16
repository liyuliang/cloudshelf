package awss3

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"cloudshelf/internal/model"
	"cloudshelf/internal/proxy"
)

// awsHTTPClient 缓存带代理配置的 HTTP 客户端：代理配置不变时复用同一个实例，
// 从而共享连接池（省去重复的 TCP/TLS 握手）；配置一改就重建。
var (
	awsHTTPMu     sync.Mutex
	awsHTTPClient aws.HTTPClient
	awsHTTPConfig proxy.Config
)

func awsHTTPClientFor(c proxy.Config) aws.HTTPClient {
	awsHTTPMu.Lock()
	defer awsHTTPMu.Unlock()
	if awsHTTPClient == nil || awsHTTPConfig != c {
		awsHTTPClient = awshttp.NewBuildableClient().WithTransportOptions(func(tr *http.Transport) {
			tr.Proxy = proxy.ProxyFuncFor(c)
		})
		awsHTTPConfig = c
	}
	return awsHTTPClient
}

func newClient(a *model.Account) *s3.Client {
	return newClientWithProxy(a, proxy.Current())
}

// newClientWithProxy 用指定的代理配置构造 S3 客户端（newClient 传当前配置，测试传固定配置）。
func newClientWithProxy(a *model.Account, pc proxy.Config) *s3.Client {
	creds := credentials.NewStaticCredentialsProvider(a.AccessKeyID, a.SecretAccessKey, "")
	region := a.Region
	if region == "" {
		region = "us-east-1"
	}
	return s3.New(s3.Options{
		Region:       region,
		Credentials:  creds,
		UsePathStyle: a.UsePathStyle,
		BaseEndpoint: endpointPtr(a.Endpoint),
		HTTPClient:   awsHTTPClientFor(pc),
	})
}

func endpointPtr(e string) *string {
	if e == "" {
		return nil
	}
	return &e
}

// TestConnection 通过 ListBuckets 验证账号凭证是否有效。
func TestConnection(a *model.Account) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	client := newClient(a)
	_, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	return err
}

type Object struct {
	Key          string
	Name         string
	IsDir        bool
	Size         int64
	Type         string
	LastModified time.Time
}

// fileType 根据文件名返回一个简明的类型描述。
func fileType(name string) string {
	ext := strings.TrimPrefix(filepath.Ext(name), ".")
	if ext == "" {
		return "文件"
	}
	return strings.ToUpper(ext)
}

// ListObjects 单层列出目录内容。S3 单次 ListObjectsV2 最多返回 1000 个 key
// （CommonPrefixes 与 Contents 合计）；这里按约定只取第一页、不翻页，
// 首页之外的文件通过 SearchObjects 的服务端搜索定位。
// 返回的 bool 表示结果被截断（该目录还有更多条目）。
func ListObjects(a *model.Account, bucket, prefix string) ([]Object, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return listObjects(ctx, newClient(a), bucket, prefix, "")
}

// SearchObjects 按名称前缀在服务端搜索：把关键字拼到当前 prefix 之后作为
// ListObjectsV2 的 Prefix 并去掉 Delimiter，因此能命中列表首页之外、
// 以及子目录里的文件。S3 只支持前缀匹配，无法做任意位置包含。
func SearchObjects(a *model.Account, bucket, prefix, keyword string) ([]Object, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return listObjects(ctx, newClient(a), bucket, prefix, keyword)
}

func listObjects(ctx context.Context, client *s3.Client, bucket, prefix, keyword string) ([]Object, bool, error) {
	in := &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: aws.String(prefix + keyword),
	}
	if keyword == "" {
		in.Delimiter = aws.String("/")
	}
	out, err := client.ListObjectsV2(ctx, in)
	if err != nil {
		return nil, false, err
	}
	var objects []Object
	for _, cp := range out.CommonPrefixes {
		p := aws.ToString(cp.Prefix)
		name := p[len(prefix):]
		objects = append(objects, Object{Key: p, Name: name, IsDir: true, Type: "文件夹"})
	}
	for _, obj := range out.Contents {
		k := aws.ToString(obj.Key)
		if k == prefix || strings.HasSuffix(k, "/") {
			continue // 跳过目录占位对象
		}
		name := k[len(prefix):]
		objects = append(objects, Object{
			Key:          k,
			Name:         name,
			IsDir:        false,
			Size:         aws.ToInt64(obj.Size),
			Type:         fileType(name),
			LastModified: aws.ToTime(obj.LastModified),
		})
	}
	sortObjects(objects)
	return objects, aws.ToBool(out.IsTruncated), nil
}

// ResolvePathDir 按路径中的目录段逐级下钻，返回最后一个「存在」的目录前缀
// （以 / 结尾，空串表示 bucket 根）；目录存在指其下至少有一个对象。
// 有多少个目录段就尝试进入多少级，某一级找不到就停在上一级，不跳到空目录。
func ResolvePathDir(a *model.Account, bucket string, dirs []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return resolvePathDir(ctx, newClient(a), bucket, dirs)
}

func resolvePathDir(ctx context.Context, client *s3.Client, bucket string, dirs []string) (string, error) {
	prefix := ""
	for _, d := range dirs {
		next := prefix + d + "/"
		ok, err := dirExists(ctx, client, bucket, next)
		if err != nil {
			return "", err
		}
		if !ok {
			break
		}
		prefix = next
	}
	return prefix, nil
}

// dirExists 用 MaxKeys=1 只探测 prefix 下是否有任意对象，避免整页列表的 1000 上限影响判断。
func dirExists(ctx context.Context, client *s3.Client, bucket, prefix string) (bool, error) {
	out, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  &bucket,
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		return false, err
	}
	return len(out.Contents) > 0 || len(out.CommonPrefixes) > 0, nil
}

// sortObjects 把文件列表按修改时间倒序显示：文件夹固定在前面便于下钻，
// 文件夹与文件各自按修改时间倒序（S3 的 CommonPrefixes 没有时间，保持接口原始顺序）。
func sortObjects(objects []Object) {
	sort.SliceStable(objects, func(i, j int) bool {
		a, b := objects[i], objects[j]
		if a.IsDir != b.IsDir {
			return a.IsDir
		}
		return a.LastModified.After(b.LastModified)
	})
}

// ListObjectsRecursive 递归列出 prefix 下的所有文件对象（不再按 / 分层），用于整个文件夹下载。
// 返回的 Name 是相对 prefix 的路径（可能包含子目录）。
func ListObjectsRecursive(a *model.Account, bucket, prefix string) ([]Object, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	client := newClient(a)
	var objects []Object
	var token *string
	for {
		out, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            &bucket,
			Prefix:            &prefix,
			ContinuationToken: token,
		})
		if err != nil {
			return nil, err
		}
		for _, obj := range out.Contents {
			k := aws.ToString(obj.Key)
			if k == prefix || strings.HasSuffix(k, "/") {
				continue // 跳过自身与文件夹占位对象
			}
			name := strings.TrimPrefix(k, prefix)
			if name == "" {
				name = k[strings.LastIndex(k, "/")+1:]
			}
			objects = append(objects, Object{
				Key:          k,
				Name:         name,
				Size:         aws.ToInt64(obj.Size),
				Type:         fileType(name),
				LastModified: aws.ToTime(obj.LastModified),
			})
		}
		if !aws.ToBool(out.IsTruncated) {
			break
		}
		token = out.NextContinuationToken
	}
	return objects, nil
}

// progressReader 包装 io.ReadSeeker，每次读取后回调进度（用于上传）。
type progressReader struct {
	r     io.ReadSeeker
	done  int64
	total int64
	cb    func(done, total int64)
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	p.done += int64(n)
	if p.cb != nil {
		p.cb(p.done, p.total)
	}
	return n, err
}

func (p *progressReader) Seek(offset int64, whence int) (int64, error) {
	pos, err := p.r.Seek(offset, whence)
	if whence == io.SeekStart && offset == 0 {
		p.done = 0
	}
	return pos, err
}

// progressWriter 包装 io.Writer，每次写入后回调进度（用于下载）。
type progressWriter struct {
	w     io.Writer
	done  int64
	total int64
	cb    func(done, total int64)
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n, err := p.w.Write(b)
	p.done += int64(n)
	if p.cb != nil {
		p.cb(p.done, p.total)
	}
	return n, err
}

// ErrReadOnly 表示当前账号是只读分享，不允许写入。
var ErrReadOnly = errors.New("只读分享配置，禁止写入操作")

// checkWritable 拦截只读账号的写操作（分享配置加载的账号只读）。
func checkWritable(a *model.Account) error {
	if a != nil && a.ReadOnly {
		return ErrReadOnly
	}
	return nil
}

// UploadFile 上传本地文件到 bucket/key，progress 回调可为 nil。
func UploadFile(a *model.Account, bucket, key, localPath string, progress func(done, total int64)) error {
	if err := checkWritable(a); err != nil {
		return err
	}
	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	total := info.Size()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	client := newClient(a)

	body := &progressReader{r: f, total: total, cb: progress}
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        &bucket,
		Key:           &key,
		Body:          body,
		ContentLength: aws.Int64(total),
	})
	return err
}

// DownloadFile 下载 bucket/key 到本地 localPath，progress 回调可为 nil。
func DownloadFile(a *model.Account, bucket, key, localPath string, progress func(done, total int64)) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	client := newClient(a)

	out, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return err
	}
	defer out.Body.Close()

	// 下载整个文件夹时对象 key 含子目录，需先建立本地父目录。
	if dir := filepath.Dir(localPath); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	f, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	pw := &progressWriter{w: f, total: aws.ToInt64(out.ContentLength), cb: progress}
	_, err = io.Copy(pw, out.Body)
	return err
}

// DeleteObject 删除 bucket/key 指向的对象。
func DeleteObject(a *model.Account, bucket, key string) error {
	if err := checkWritable(a); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client := newClient(a)
	_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	return err
}

// CreateFolder 在 bucket 的 prefix 下创建一个名为 name 的"文件夹"（S3 用以 / 结尾的空对象表示）。
func CreateFolder(a *model.Account, bucket, prefix, name string) error {
	if err := checkWritable(a); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client := newClient(a)
	key := prefix + name
	if key == "" || key[len(key)-1] != '/' {
		key += "/"
	}
	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	return err
}

// PresignURL 为 bucket/key 生成一个有效期为 expiry 的预签名 GET URL。
func PresignURL(a *model.Account, bucket, key string, expiry time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client := newClient(a)
	ps := s3.NewPresignClient(client)
	req, err := ps.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

type BucketInfo struct {
	Name       string
	Accessible bool
}

// HeadBucket 检查是否有权限访问该 bucket。
// 先尝试 HeadBucket API；若失败则回退到 ListObjectsV2(MaxKeys=0)，
// 后者对 S3 兼容服务（MinIO、Ceph 等）以及仅授予 s3:ListBucket 的 IAM 策略兼容性更好。
func HeadBucket(a *model.Account, bucket string) error {
	client := newClient(a)

	ctx1, cancel1 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel1()
	_, err := client.HeadBucket(ctx1, &s3.HeadBucketInput{Bucket: &bucket})
	if err == nil {
		return nil
	}

	// 回退：用 ListObjectsV2 探测访问权限（MaxKeys=1 只返回至多一个对象，开销极小）。
	// 注意：MaxKeys 必须 ≥ 1，某些 S3 兼容服务（MinIO、Ceph 等）会拒绝 MaxKeys=0。
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	one := int32(1)
	_, err2 := client.ListObjectsV2(ctx2, &s3.ListObjectsV2Input{
		Bucket:  &bucket,
		MaxKeys: &one,
	})
	return err2
}

// ListBucketsWithAccess 列出账号下的桶，并用 HeadBucket 并发检查每个桶的访问权限。
func ListBucketsWithAccess(a *model.Account) ([]BucketInfo, error) {
	names, err := ListBuckets(a)
	if err != nil {
		return nil, err
	}
	infos := make([]BucketInfo, len(names))
	var wg sync.WaitGroup
	for i, name := range names {
		infos[i].Name = name
		wg.Add(1)
		go func(idx int, b string) {
			defer wg.Done()
			infos[idx].Accessible = HeadBucket(a, b) == nil
		}(i, name)
	}
	wg.Wait()
	return infos, nil
}

// ListBuckets 返回账号下的桶名列表。
func ListBuckets(a *model.Account) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	client := newClient(a)
	out, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(out.Buckets))
	for _, b := range out.Buckets {
		names = append(names, aws.ToString(b.Name))
	}
	return names, nil
}
