package rws

import (
	"cloudshelf/internal/gorws/odm"
	"context"
	"fmt"
	"net/http"
)

func (c Client) Site() Site {
	return &site{
		HTTPRequester: &baseReq{
			baseUrl: c.baseURL,
			headers: c.defaultHeaders(),
		},
	}
}

type Site interface {
	List(ctx context.Context, studyOId string) ([]*odm.AdminData, error)
}

type site struct {
	HTTPRequester *baseReq
}

// List 获取site列表
func (c *site) List(ctx context.Context, studyOId string) ([]*odm.AdminData, error) {
	studyOId = EnsureURLEncoded(studyOId) //studyOId 需要传入url encode的值

	path := fmt.Sprintf(`/datasets/Sites.odm?studyoid=%s`, studyOId)
	resp := &ODMResp{}
	if err := c.HTTPRequester.SendRequest(ctx, http.MethodGet, path, nil, resp); err != nil {
		return nil, err
	}
	return resp.AdminData, nil
}
