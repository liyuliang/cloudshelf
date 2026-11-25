package rws

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type HTTPRequester interface {
	SendRequest(ctx context.Context, method, path string, body io.Reader, response Response) error
}

type baseReq struct {
	baseUrl string
	headers http.Header
}

func (req *baseReq) SendRequest(ctx context.Context, method, path string, body io.Reader, resp Response) error {

	u := req.baseUrl + "/" + path

	request, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return err
	}
	request.Header = req.headers

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		// 读取响应体以获取服务端错误信息
		errBody, _ := io.ReadAll(response.Body)
		errMsg := string(errBody)
		if len(errMsg) > 500 {
			errMsg = errMsg[:500]
		}
		if errMsg != "" {
			return fmt.Errorf("http status code: %d, url: %s, response: %s", response.StatusCode, u, errMsg)
		}
		return fmt.Errorf("http status code: %d, url: %s", response.StatusCode, u)
	}
	return resp.Decode(response.Body)
}
