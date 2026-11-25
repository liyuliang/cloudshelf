package rws

import (
	"context"
	"fmt"
	"net/http"

	"cloudshelf/internal/gorws/odm"
)

func (c Client) Subject() Subject {
	return &subject{
		HTTPRequester: &baseReq{
			baseUrl: c.baseURL,
			headers: c.defaultHeaders(),
		},
	}
}

type Subject interface {
	List(ctx context.Context, studyOId string) ([]*odm.ClinicalData, error)
}

type subject struct {
	HTTPRequester *baseReq
}

func (c *subject) List(ctx context.Context, studyOId string) ([]*odm.ClinicalData, error) {
	studyOId = EnsureURLEncoded(studyOId) //studyOId 需要传入url encode的值

	path := fmt.Sprintf(`/studies/%s/Subjects?status=all`, studyOId)
	resp := &ODMResp{}
	if err := c.HTTPRequester.SendRequest(ctx, http.MethodGet, path, nil, resp); err != nil {
		return nil, err
	}
	return resp.ClinicalData, nil
}
