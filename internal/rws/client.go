// Package rws 封装 go-rws SDK，为 CloudShelf 提供 Rave Web Services 数据浏览能力。
//
// Account 字段映射：
//   - AccessKeyID     → RWS 用户名
//   - SecretAccessKey → RWS 密码（数据库中加密存储）
//   - Region          → RWS Domain（如 janussandbox、innovate）
//   - Endpoint        → Host 模板覆盖（留空使用默认 https://%s.mdsol.com）
//   - SiteURL         → VirtualDir 覆盖（留空使用默认 RaveWebServices）
package rws

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	gorws "codeup.aliyun.com/60b733f02c5969c370c5e0ab/sdk/go-rws"
	"codeup.aliyun.com/60b733f02c5969c370c5e0ab/sdk/go-rws/odm"

	"cloudshelf/internal/model"
)

// newClient 从 Account 构建 go-rws Client。
func newClient(a *model.Account) *gorws.Client {
	var opts []gorws.ClientOption

	// Domain（必填）
	if a.Region != "" {
		opts = append(opts, gorws.WithDomain(a.Region))
	}
	// Host 模板覆盖（可选）
	if a.Endpoint != "" {
		opts = append(opts, gorws.WithHost(a.Endpoint))
	}
	// VirtualDir 覆盖（可选）
	if a.SiteURL != "" {
		opts = append(opts, gorws.WithVirtualDir(a.SiteURL))
	}

	return gorws.NewClient(a.AccessKeyID, a.SecretAccessKey, opts...)
}

// TestConnection 测试 RWS 账号连通性（尝试获取 Study 列表）。
func TestConnection(a *model.Account) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	studies, err := newClient(a).Study().List(ctx)
	if err != nil {
		return fmt.Errorf("RWS 连接失败: %w", err)
	}
	_ = studies
	return nil
}

// ---- Study ----

type StudyInfo struct {
	OID         string
	Name        string
	Description string
}

// ListStudies 返回账号下的所有 Study。
func ListStudies(a *model.Account) ([]StudyInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	studies, err := newClient(a).Study().List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]StudyInfo, 0, len(studies))
	for _, s := range studies {
		info := StudyInfo{OID: s.OID}
		if s.GlobalVariables != nil {
			info.Name = s.GlobalVariables.StudyName
			info.Description = s.GlobalVariables.StudyDescription
		}
		if info.Name == "" {
			info.Name = s.OID
		}
		out = append(out, info)
	}
	return out, nil
}

// ---- Subject ----

type SubjectInfo struct {
	SubjectKey string
	SiteRef    string
	Status     string
}

// ListSubjects 返回指定 Study 下的所有受试者。
func ListSubjects(a *model.Account, studyOID string) ([]SubjectInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	clinicalData, err := newClient(a).Subject().List(ctx, studyOID)
	if err != nil {
		return nil, err
	}
	var out []SubjectInfo
	for _, cd := range clinicalData {
		for _, sd := range cd.SubjectData {
			info := SubjectInfo{
				SubjectKey: sd.SubjectKey,
				Status:     sd.Status,
			}
			if sd.SiteRef != nil {
				info.SiteRef = sd.SiteRef.LocationOID
			}
			out = append(out, info)
		}
	}
	return out, nil
}

// ---- Site ----

type SiteInfo struct {
	OID              string
	Name             string
	StudySiteNumber  string
}

// ListSites 返回指定 Study 下的所有中心（Site）。
func ListSites(a *model.Account, studyOID string) ([]SiteInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	adminData, err := newClient(a).Site().List(ctx, studyOID)
	if err != nil {
		return nil, err
	}
	var out []SiteInfo
	for _, ad := range adminData {
		for _, loc := range ad.Location {
			info := SiteInfo{
				OID:  loc.OID,
				Name: loc.Name,
			}
			if len(loc.MetaDataVersionRef) > 0 {
				info.StudySiteNumber = loc.MetaDataVersionRef[0].StudySiteNumber
			}
			out = append(out, info)
		}
	}
	return out, nil
}

// ---- Dataset ----

type DatasetField struct {
	OID      string
	Name     string
	DataType string
	Length   int64
}

type DatasetInfo struct {
	OID    string
	Name   string
	Domain string
	Type   string // "raw" or "regular"
	Fields []DatasetField
}

// ListDatasets 返回指定 Study 的元数据（数据集定义）。
// 同时请求 raw 和 regular 两种类型，合并去重后返回，每条记录标注其类型。
func ListDatasets(a *model.Account, studyOID string) ([]DatasetInfo, error) {
	client := newClient(a)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var allDatasets []DatasetInfo

	// 优先加载 raw（大多数 CSV 是 raw 类型）
	for _, typ := range []string{"raw", "regular"} {
		odmData, err := client.DatasetsMetadata(ctx, studyOID, typ)
		if err != nil {
			// 某一类型元数据请求失败不影响另一类型
			continue
		}
		ds := parseDatasets(odmData)
		for i := range ds {
			ds[i].Type = typ
		}
		allDatasets = append(allDatasets, ds...)
	}

	// 去重合并：同名 dataset 合并类型（如 raw+regular）
	seen := make(map[string]int) // name → index in merged
	var merged []DatasetInfo
	for _, d := range allDatasets {
		if idx, ok := seen[d.Name]; ok {
			// 已存在同名数据集，合并类型
			existing := &merged[idx]
			if !strings.Contains(existing.Type, d.Type) {
				existing.Type = existing.Type + "+" + d.Type
			}
		} else {
			seen[d.Name] = len(merged)
			merged = append(merged, d)
		}
	}
	return merged, nil
}

// datasetName 从 ItemGroupDef 推导 CSV 端点所需的数据集名称。
// 优先用 SASDatasetName（最精确），其次去掉 _LOG_LINE / _NON_LOG_LINE 后缀的 Name，
// 最后才回退到 Domain（多个数据集可能共享同一 Domain，如 DS_IC/DS_W/DS 都是 DS 域）。
func datasetName(ig odm.ItemGroupDef) string {
	if ig.SASDatasetName != "" {
		return ig.SASDatasetName
	}
	// 去掉常见后缀
	name := ig.Name
	for _, suffix := range []string{"_NON_LOG_LINE", "_LOG_LINE"} {
		if trimmed, ok := strings.CutSuffix(name, suffix); ok {
			return trimmed
		}
	}
	// Domain 作为最后手段（可能多条 ItemGroupDef 共享同一 Domain）
	if ig.Domain != "" {
		return ig.Domain
	}
	return name
}

func parseDatasets(odmData *odm.ODM) []DatasetInfo {
	if odmData == nil {
		return nil
	}
	var out []DatasetInfo
	for _, study := range odmData.Studies {
		for _, mdv := range study.MetaDataVersions {
			for _, ig := range mdv.ItemGroupDef {
				dsName := datasetName(ig)
				ds := DatasetInfo{
					OID:    dsName,
					Name:   dsName,
					Domain: ig.Domain,
				}
				// 通过 ItemRef 查找 ItemDef 填充字段
				itemDefMap := make(map[string]odm.ItemDef, len(mdv.ItemDef))
				for _, id := range mdv.ItemDef {
					itemDefMap[id.OID] = id
				}
				for _, ir := range ig.ItemRef {
					if def, ok := itemDefMap[ir.ItemOID]; ok {
						ds.Fields = append(ds.Fields, DatasetField{
							OID:      def.OID,
							Name:     def.Name,
							DataType: def.DataType,
							Length:   def.Length,
						})
					}
				}
				out = append(out, ds)
			}
		}
	}
	return out
}

// ---- Dataset CSV Type Probe ----

// buildBaseURL 复制 go-rws SDK 的 baseUrl 构建逻辑，用于轻量级 HTTP 探测。
func buildBaseURL(a *model.Account) string {
	domain := a.Region
	if domain == "" {
		domain = "innovate"
	}
	host := "https://%s.mdsol.com"
	if a.Endpoint != "" {
		host = a.Endpoint
	}
	virtualDir := "RaveWebServices"
	if a.SiteURL != "" {
		virtualDir = a.SiteURL
	}

	if strings.HasPrefix(strings.ToLower(domain), "http") {
		u, _ := url.JoinPath(domain, virtualDir)
		return u
	}
	h := strings.ToLower(host)
	if !strings.Contains(h, "%s") {
		u, _ := url.JoinPath(host, virtualDir)
		return u
	}
	u, _ := url.JoinPath(fmt.Sprintf(h, domain), virtualDir)
	return u
}

// probeCSV 用轻量级 HTTP GET 检查指定 CSV 端点是否可用（200），
// 只读取极少字节后立即关闭连接，避免下载完整文件。
func probeCSV(ctx context.Context, baseURL, user, pass, studyOID, typ, dataset string) bool {
	csvPath := fmt.Sprintf("studies/%s/datasets/%s/%s.csv",
		url.PathEscape(studyOID), url.PathEscape(typ), url.PathEscape(dataset))
	u := baseURL + "/" + csvPath

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return false
	}
	req.SetBasicAuth(user, pass)
	req.Header.Set("Content-Type", "text/xml")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	// 只读 1 字节确认有数据，立即关闭
	_, _ = io.ReadAll(io.LimitReader(resp.Body, 1))
	_ = resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// ProbeDatasetCSVTypes 并发探测每个数据集在 raw/regular 两种类型下的 CSV 可用性。
// 返回 map[datasetName]type，type 为 "raw"、"regular" 或 "raw+regular"。
func ProbeDatasetCSVTypes(a *model.Account, studyOID string, datasets []string) map[string]string {
	baseURL := buildBaseURL(a)
	user := a.AccessKeyID
	pass := a.SecretAccessKey

	type result struct {
		name string
		typ  string
		ok   bool
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ch := make(chan result, len(datasets)*2)
	var wg sync.WaitGroup

	// 限制并发数，避免过多请求
	sem := make(chan struct{}, 6)

	for _, ds := range datasets {
		for _, typ := range []string{"raw", "regular"} {
			wg.Add(1)
			go func(name, t string) {
				defer wg.Done()
				sem <- struct{}{}
				ok := probeCSV(ctx, baseURL, user, pass, studyOID, t, name)
				<-sem
				ch <- result{name: name, typ: t, ok: ok}
			}(ds, typ)
		}
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	// 收集结果
	avail := make(map[string]map[string]bool) // name → {raw: true, regular: true}
	for r := range ch {
		if !r.ok {
			continue
		}
		if avail[r.name] == nil {
			avail[r.name] = make(map[string]bool)
		}
		avail[r.name][r.typ] = true
	}

	out := make(map[string]string, len(avail))
	for name, types := range avail {
		hasRaw := types["raw"]
		hasReg := types["regular"]
		switch {
		case hasRaw && hasReg:
			out[name] = "raw+regular"
		case hasRaw:
			out[name] = "raw"
		case hasReg:
			out[name] = "regular"
		}
	}
	return out
}

// ---- Dataset Data (CSV) ----

type DatasetData struct {
	Headers []string
	Rows    [][]string
	Type    string // "raw" or "regular" — 实际成功访问的类型
}

// GetDatasetCSV 获取指定 Study 的某个 Dataset 的 CSV 数据。
// hintType 是列表中探测到的类型提示（可选），优先尝试该类型；
// 若为空则默认先 raw 再 regular。
func GetDatasetCSV(a *model.Account, studyOID, hintType, dataset string) (*DatasetData, error) {
	client := newClient(a)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 确定尝试顺序：有提示就先用提示类型，否则 raw 优先
	tryOrder := []string{"raw", "regular"}
	if hintType == "regular" {
		tryOrder = []string{"regular", "raw"}
	} else if hintType == "raw" {
		tryOrder = []string{"raw", "regular"}
	}

	var lastErr error
	for _, typ := range tryOrder {
		resp, err := client.DatasetListCsv(ctx, studyOID, typ, dataset)
		if err != nil {
			lastErr = err
			// HTTP 状态码错误（4xx/5xx）→ 尝试下一个类型
			if strings.Contains(err.Error(), "http status code:") {
				continue
			}
			// 非 HTTP 错误（网络不可达、超时、认证失败等）→ 直接返回
			return nil, err
		}
		if len(resp.Data) == 0 {
			return &DatasetData{Type: typ}, nil
		}
		return &DatasetData{
			Headers: resp.Data[0],
			Rows:    resp.Data[1:],
			Type:    typ,
		}, nil
	}
	return nil, lastErr
}
