package main

import (
	"cloudshelf/internal/rws"
)

// ---- RWS DTO ----

type RWSStudyDTO struct {
	OID         string
	Name        string
	Description string
}

type RWSSubjectDTO struct {
	SubjectKey string
	SiteRef    string
	Status     string
}

type RWSSiteDTO struct {
	OID             string
	Name            string
	StudySiteNumber string
}

type RWSDatasetDTO struct {
	OID    string
	Name   string
	Domain string
	Fields []RWSDatasetFieldDTO
}

type RWSDatasetFieldDTO struct {
	OID      string
	Name     string
	DataType string
	Length   int64
}

type RWSDatasetDataDTO struct {
	Headers []string
	Rows    [][]string
}

// ---- RWS Study 列表 ----

func (a *App) RWSListStudies(acc AccountDTO) ([]RWSStudyDTO, error) {
	studies, err := rws.ListStudies(toAccount(acc))
	if err != nil {
		return nil, err
	}
	out := make([]RWSStudyDTO, 0, len(studies))
	for _, s := range studies {
		out = append(out, RWSStudyDTO{
			OID:         s.OID,
			Name:        s.Name,
			Description: s.Description,
		})
	}
	return out, nil
}

// ---- RWS Subject 列表 ----

func (a *App) RWSListSubjects(acc AccountDTO, studyOID string) ([]RWSSubjectDTO, error) {
	subjects, err := rws.ListSubjects(toAccount(acc), studyOID)
	if err != nil {
		return nil, err
	}
	out := make([]RWSSubjectDTO, 0, len(subjects))
	for _, s := range subjects {
		out = append(out, RWSSubjectDTO{
			SubjectKey: s.SubjectKey,
			SiteRef:    s.SiteRef,
			Status:     s.Status,
		})
	}
	return out, nil
}

// ---- RWS Site 列表 ----

func (a *App) RWSListSites(acc AccountDTO, studyOID string) ([]RWSSiteDTO, error) {
	sites, err := rws.ListSites(toAccount(acc), studyOID)
	if err != nil {
		return nil, err
	}
	out := make([]RWSSiteDTO, 0, len(sites))
	for _, s := range sites {
		out = append(out, RWSSiteDTO{
			OID:             s.OID,
			Name:            s.Name,
			StudySiteNumber: s.StudySiteNumber,
		})
	}
	return out, nil
}

// ---- RWS Dataset 元数据 ----

func (a *App) RWSListDatasets(acc AccountDTO, studyOID, typ string) ([]RWSDatasetDTO, error) {
	datasets, err := rws.ListDatasets(toAccount(acc), studyOID, typ)
	if err != nil {
		return nil, err
	}
	out := make([]RWSDatasetDTO, 0, len(datasets))
	for _, ds := range datasets {
		dto := RWSDatasetDTO{
			OID:    ds.OID,
			Name:   ds.Name,
			Domain: ds.Domain,
		}
		for _, f := range ds.Fields {
			dto.Fields = append(dto.Fields, RWSDatasetFieldDTO{
				OID:      f.OID,
				Name:     f.Name,
				DataType: f.DataType,
				Length:   f.Length,
			})
		}
		out = append(out, dto)
	}
	return out, nil
}

// ---- RWS Dataset CSV 数据 ----

func (a *App) RWSGetDatasetCSV(acc AccountDTO, studyOID, typ, dataset string) (RWSDatasetDataDTO, error) {
	data, err := rws.GetDatasetCSV(toAccount(acc), studyOID, typ, dataset)
	if err != nil {
		return RWSDatasetDataDTO{}, err
	}
	return RWSDatasetDataDTO{
		Headers: data.Headers,
		Rows:    data.Rows,
	}, nil
}
