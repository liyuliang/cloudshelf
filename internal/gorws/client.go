package rws

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

func NewClient(username string, password string, options ...ClientOption) *Client {
	opt := newClientOptions()
	for _, option := range options {
		option(opt)
	}

	client := &Client{
		username: username,
		password: password,
		baseURL:  opt.baseUrl(),
		client:   http.DefaultClient,
	}

	return client
}

func (c Client) get(ctx context.Context, path string, vals url.Values, resp Response) error {
	p := path
	if len(vals) > 0 {
		p = fmt.Sprintf("%s?%s", path, vals.Encode())
	}

	req := &baseReq{
		baseUrl: c.baseURL,
		headers: c.defaultHeaders(),
	}
	return req.SendRequest(ctx, http.MethodGet, p, nil, resp)
}

func (c Client) post(ctx context.Context, path string, vals url.Values, body io.Reader, resp Response) error {
	p := path
	if len(vals) > 0 {
		p = fmt.Sprintf("%s?%s", path, vals.Encode())
	}

	req := &baseReq{
		baseUrl: c.baseURL,
		headers: c.defaultHeaders(),
	}
	return req.SendRequest(ctx, http.MethodPost, p, body, resp)
}

func (c Client) getAuthHeader() http.Header {
	req := &http.Request{
		Header: http.Header{},
	}
	req.SetBasicAuth(c.username, c.password)
	return req.Header
}

func (c Client) defaultHeaders() http.Header {
	h := c.getAuthHeader()
	h.Set("Content-Type", "text/xml")
	return h
}

// EnsureURLEncoded 确保输入字符串经过 URL 编码
func EnsureURLEncoded(input string) string {
	// 尝试解码输入字符串
	decoded, err := url.QueryUnescape(input)
	if err != nil {
		// 如果解码失败，说明输入可能已经是未编码的
		decoded = input
	}
	// 重新编码字符串
	encoded := url.QueryEscape(decoded)
	return encoded
}
