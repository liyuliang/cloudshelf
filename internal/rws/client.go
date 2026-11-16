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
	Fields []DatasetField
}

// ListDatasets 返回指定 Study 的元数据（数据集定义）。
// typ 可选 "regular" 或 "raw"。
func ListDatasets(a *model.Account, studyOID, typ string) ([]DatasetInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	odmData, err := newClient(a).DatasetsMetadata(ctx, studyOID, typ)
	if err != nil {
		return nil, err
	}
	return parseDatasets(odmData), nil
}

func parseDatasets(odmData *odm.ODM) []DatasetInfo {
	if odmData == nil {
		return nil
	}
	var out []DatasetInfo
	for _, study := range odmData.Studies {
		for _, mdv := range study.MetaDataVersions {
			for _, ig := range mdv.ItemGroupDef {
				ds := DatasetInfo{
					OID:    ig.OID,
					Name:   ig.Name,
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

// ---- Dataset Data (CSV) ----

type DatasetData struct {
	Headers []string
	Rows    [][]string
}

// GetDatasetCSV 获取指定 Study 的某个 Dataset 的 CSV 数据。
// typ 可选 "regular" 或 "raw"。
func GetDatasetCSV(a *model.Account, studyOID, typ, dataset string) (*DatasetData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	resp, err := newClient(a).DatasetListCsv(ctx, studyOID, typ, dataset)
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return &DatasetData{}, nil
	}
	return &DatasetData{
		Headers: resp.Data[0],
		Rows:    resp.Data[1:],
	}, nil
}
