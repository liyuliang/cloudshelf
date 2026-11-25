package rws

import (
	"bytes"
	"cloudshelf/internal/gorws/odm"
	"context"
	"encoding/xml"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"time"
)

func (c Client) Query() Query {
	return &query{
		HTTPRequester: &baseReq{
			baseUrl: c.baseURL,
			headers: c.defaultHeaders(),
		},
	}
}

type Query interface {
	Open(ctx context.Context, param OpenQueryParam) (*OpenQueryResp, error)
	Cancel(ctx context.Context, param CancelQueryParam) (*CancelQueryResp, error)
	Answer(ctx context.Context, param AnswerQueryParam) (*AnswerQueryResp, error)
	Close(ctx context.Context, param CloseQueryParam) (*CloseQueryResp, error)
	ReQuery(ctx context.Context, param ReQueryParam) (*ReQueryResp, error)
}

func (c query) Open(ctx context.Context, param OpenQueryParam) (*OpenQueryResp, error) {
	body, err := xml.MarshalIndent(param.toODM(), "", "  ")
	if err != nil {
		return nil, err
	}

	resp := &OpenQueryResp{
		RequestXml: string(body),
	}
	err = c.HTTPRequester.SendRequest(ctx, http.MethodPost, "webservice.aspx?PostODMClinicalData", bytes.NewReader(body), resp)
	if err != nil {
		return resp, err
	}
	//if len(resp.Queries) == 0 {
	//	return nil, fmt.Errorf("%s", resp.ErrorClientResponseMessage)
	//}
	return resp, nil
}
func (c query) Cancel(ctx context.Context, param CancelQueryParam) (*CancelQueryResp, error) {
	body, err := xml.MarshalIndent(param.toODM(), "", "  ")
	if err != nil {
		return nil, err
	}

	resp := &CancelQueryResp{
		RequestXml: string(body),
	}
	err = c.HTTPRequester.SendRequest(ctx, http.MethodPost, "webservice.aspx?PostODMClinicalData", bytes.NewReader(body), resp)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (c query) Answer(ctx context.Context, param AnswerQueryParam) (*AnswerQueryResp, error) {
	body, err := xml.MarshalIndent(param.toODM(), "", "  ")
	if err != nil {
		return nil, err
	}

	resp := &AnswerQueryResp{
		RequestXml: string(body),
	}
	err = c.HTTPRequester.SendRequest(ctx, http.MethodPost, "webservice.aspx?PostODMClinicalData", bytes.NewReader(body), resp)
	if err != nil {
		return resp, err
	}
	return resp, nil
}
func (c query) Close(ctx context.Context, param CloseQueryParam) (*CloseQueryResp, error) {
	body, err := xml.MarshalIndent(param.toODM(), "", "  ")
	if err != nil {
		return nil, err
	}

	resp := &CloseQueryResp{
		RequestXml: string(body),
	}
	err = c.HTTPRequester.SendRequest(ctx, http.MethodPost, "webservice.aspx?PostODMClinicalData", bytes.NewReader(body), resp)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (c query) ReQuery(ctx context.Context, param ReQueryParam) (*ReQueryResp, error) {
	body, err := xml.MarshalIndent(param.toODM(), "", "  ")
	if err != nil {
		return nil, err
	}

	resp := &ReQueryResp{
		RequestXml: string(body),
	}
	err = c.HTTPRequester.SendRequest(ctx, http.MethodPost, "webservice.aspx?PostODMClinicalData", bytes.NewReader(body), resp)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

type OpenQueryParam struct {
	PublicQueryParams
	Query               string
	StudyEventRepeatKey string
	FormRepeatKey       string
}
type CancelQueryParam struct {
	PublicQueryParams
	QueryId string
}
type AnswerQueryParam struct {
	PublicQueryParams
	QueryId string
	Reason  string
}
type CloseQueryParam struct {
	PublicQueryParams
	QueryId string
}

type ReQueryParam struct {
	PublicQueryParams
	QueryId   string
	QueryText string
}

const (
	QueryStatusOpen      = "Open"
	QueryStatusCancelled = "Cancelled"
	QueryStatusAnswered  = "Answered"
	QueryStatusClosed    = "Closed"
)

func (p OpenQueryParam) toODM() odm.ODM {
	return p.PublicQueryParams.ToODM(&odm.Query{
		Recipient:        p.PublicQueryParams.GetQueryRole(),
		Status:           QueryStatusOpen,
		RequiresResponse: "Yes",
		Value:            p.Query,
	})
}
func (p CancelQueryParam) toODM() odm.ODM {
	return p.PublicQueryParams.ToODM(&odm.Query{
		Recipient:      p.PublicQueryParams.GetQueryRole(),
		Status:         QueryStatusCancelled,
		QueryRepeatKey: p.QueryId,
	})
}
func (p AnswerQueryParam) toODM() odm.ODM {
	return p.PublicQueryParams.ToODM(&odm.Query{
		Recipient:      p.PublicQueryParams.GetQueryRole(),
		Status:         QueryStatusAnswered,
		QueryRepeatKey: p.QueryId,
		Response:       p.Reason,
	})
}
func (p CloseQueryParam) toODM() odm.ODM {
	return p.PublicQueryParams.ToODM(&odm.Query{
		Recipient:      p.PublicQueryParams.GetQueryRole(),
		Status:         QueryStatusClosed,
		QueryRepeatKey: p.QueryId,
	})
}
func (p ReQueryParam) toODM() odm.ODM {
	return p.PublicQueryParams.ToODM(&odm.Query{
		Recipient:               p.PublicQueryParams.GetQueryRole(),
		Status:                  QueryStatusOpen,
		PrecedingQueryRepeatKey: p.QueryId,
		RequiresResponse:        "Yes",
		Value:                   p.QueryText,
	})
}

type query struct {
	HTTPRequester
}

type PublicQueryParams struct {
	Study              string
	CRFVersion         string
	Site               string
	Subject            string
	VisitOId           string
	VisitRepeatKey     string
	FormOId            string
	FormRepeatKey      string
	ItemGroupRepeatKey string
	FieldOId           string
	FieldValue         string
	FieldSpecifyValue  string
	FieldIsNull        string
	FileOId            string `json:"FileOID,omitempty"`
	QueryRole          string `json:"Recipient,omitempty"`
	TransactionType    string
	IsLogLine          bool
	SubmitMarkingOnly  string
}

const (
	QueryMarkingGroupNameMM = "Site from MM"

	QuerySubmitMarkingOnlyYes = "Yes"
)

func (p PublicQueryParams) GetQueryRole() string {
	if p.QueryRole == "" {
		return QueryMarkingGroupNameMM
	}
	return p.QueryRole
}

func (p PublicQueryParams) ItemOID() string {
	return fmt.Sprintf("%s.%s", p.FormOId, p.FieldOId)
}

func (p PublicQueryParams) ItemGroupOId() string {
	if p.IsLogLine {
		return fmt.Sprintf("%s%s", p.FormOId, "_LOG_LINE")
	} else {
		return p.FormOId
	}
}

func (p PublicQueryParams) ItemData(q *odm.Query) []odm.ItemData {
	var itemData []odm.ItemData
	if p.SubmitMarkingOnly == QuerySubmitMarkingOnlyYes {
		itemData = []odm.ItemData{
			{
				Value:             "",
				SpecifyValue:      "",
				TransactionType:   p.TransactionType,
				SubmitMarkingOnly: p.SubmitMarkingOnly,
				ItemOID:           p.ItemOID(),
				Query:             q,
			},
		}
	} else {
		itemData = []odm.ItemData{
			{
				Value:           p.FieldValue,
				SpecifyValue:    p.FieldSpecifyValue,
				TransactionType: p.TransactionType,
				ItemOID:         p.ItemOID(),
				Query:           q,
			},
		}
	}
	return itemData
}

func (p PublicQueryParams) GenerateFileOID() string {
	id, err := uuid.NewV7()
	if err != nil {
		return "4063cfea-dc23-4fe7-8865-920780be1ce4"
	}
	return id.String()
}

func (p PublicQueryParams) ToODM(q *odm.Query) odm.ODM {
	var (
		xmlns      = "http://www.cdisc.org/ns/odm/v1.3"
		xmlnsmdsol = "http://www.mdsol.com/ns/odm/metadata"
		xmlnsxlink = "http://www.w3.org/1999/xlink"
		FileType   = "Transactional"
		ODMVer     = "1.3"
		now        = time.Now().UTC()
	)

	return odm.ODM{
		FileType:         FileType,
		FileOID:          p.GenerateFileOID(),
		CreationDateTime: now.Format("2006-01-02T15:04:05.000-07:00"),
		ODMVersion:       ODMVer,
		XMLNs:            xmlns,
		XMLNsMdsol:       xmlnsmdsol,
		XMLNsXlink:       xmlnsxlink,
		ClinicalData: []*odm.ClinicalData{
			{
				StudyOID:           p.Study,
				MetaDataVersionOID: p.CRFVersion,
				SubjectData: []*odm.SubjectData{
					{
						SubjectKey: p.Subject,
						SiteRef: &odm.SiteRef{
							LocationOID: p.Site,
						},
						StudyEventData: &odm.StudyEventData{
							StudyEventOID:       p.VisitOId,
							StudyEventRepeatKey: p.VisitRepeatKey,
							FormData: &odm.FormData{
								FormOID:       p.FormOId,
								FormRepeatKey: p.FormRepeatKey,
								ItemGroupData: &odm.ItemGroupData{
									ItemGroupRepeatKey: p.ItemGroupRepeatKey,
									ItemGroupOID:       p.ItemGroupOId(),
									ItemData:           p.ItemData(q),
								},
							},
						},
					},
				},
			},
		},
	}
}
