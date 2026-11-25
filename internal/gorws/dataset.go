package rws

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"cloudshelf/internal/gorws/odm"
)

// pathEscape 对路径段做 URL 编码，但保留 / 分隔符不编码。
func pathEscape(s string) string {
	parts := strings.Split(s, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}

// DatasetsMetadata typ = regular/raw
func (c Client) DatasetsMetadata(ctx context.Context, studyID, typ string) (*odm.ODM, error) {
	var resp ODMResp
	// regular 和 raw 分别使用各自的 metadata 端点
	metaType := "regular"
	if typ == "raw" {
		metaType = "raw"
	}
	u := fmt.Sprintf("studies/%s/datasets/metadata/%s", pathEscape(studyID), metaType)
	err := c.get(ctx, u, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.ODM, nil
}

// DatasetsRawData 源数据
func (c Client) DatasetsRawData(ctx context.Context, studyID string) (*odm.ODM, error) {
	var resp ODMResp
	err := c.get(ctx, fmt.Sprintf("studies/%s/datasets/raw", pathEscape(studyID)), nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.ODM, nil
}

// func (c Client) DatasetsList(ctx context.Context, studyID, subjectID, typ string) (*odm.ODM, error) {
// 	var resp ODMResp
// 	err := c.get(ctx, fmt.Sprintf("studies/%s/subjects/%s/datasets/%s", studyID, subjectID, typ), nil, &resp)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return resp.ODM, nil
// }

// DatasetList typ = regular/raw
func (c Client) DatasetList(ctx context.Context, studyID, subjectID, typ, dataset string) (*odm.ODM, error) {
	var resp ODMResp
	err := c.get(ctx, fmt.Sprintf("studies/%s/subjects/%s/datasets/%s/%s", pathEscape(studyID), pathEscape(subjectID), url.PathEscape(typ), url.PathEscape(dataset)), nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.ODM, nil
}

// DatasetListCsv typ = regular/raw
func (c Client) DatasetListCsv(ctx context.Context, studyID, typ, dataset string) (*CsvResp, error) {
	var resp CsvResp
	err := c.get(ctx, fmt.Sprintf("studies/%s/datasets/%s/%s.csv", pathEscape(studyID), url.PathEscape(typ), url.PathEscape(dataset)), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ClinicalAuditRecordList .
func (c Client) ClinicalAuditRecordList(ctx context.Context, studyoid string, startId, perPage int) (*odm.ODM, error) {
	var resp ODMResp

	query := url.Values{}
	query.Set("studyoid", studyoid)
	query.Set("startid", strconv.Itoa(startId))
	query.Set("per_page", strconv.Itoa(perPage))

	err := c.get(ctx, "datasets/ClinicalAuditRecords.odm", query, &resp)
	if err != nil {
		return nil, err
	}
	return resp.ODM, nil
}
