package rws

import (
	"context"
	"fmt"
	"net/http"

	"cloudshelf/internal/gorws/odm"
)

func (c Client) Study() Study {
	return &study{
		HTTPRequester: &baseReq{
			baseUrl: c.baseURL,
			headers: c.defaultHeaders(),
		},
	}
}

type Study interface {
	List(ctx context.Context) ([]*odm.Study, error)
}

type study struct {
	HTTPRequester *baseReq
}

func (c *study) List(ctx context.Context) ([]*odm.Study, error) {

	path := fmt.Sprintf(`/studies`)
	resp := &ODMResp{}
	if err := c.HTTPRequester.SendRequest(ctx, http.MethodGet, path, nil, resp); err != nil {
		return nil, err
	}
	return resp.Studies, nil
}
