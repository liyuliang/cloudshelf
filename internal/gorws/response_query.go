package rws

import (
	"cloudshelf/internal/gorws/odm"
	"encoding/xml"
	"errors"
	"io"
)

type OpenQueryResp struct {
	XMLName                 xml.Name    `xml:"Response"`
	ReferenceNumber         string      `xml:"ReferenceNumber,attr"`
	InboundODMFileOID       string      `xml:"InboundODMFileOID,attr"`
	IsTransactionSuccessful string      `xml:"IsTransactionSuccessful,attr"`
	SuccessStatistics       string      `xml:"SuccessStatistics,attr"`
	NewRecords              string      `xml:"NewRecords,attr"`
	Queries                 []odm.Query `xml:"Query"`
	RequestXml              string
	ErrorResp
}

func (resp *OpenQueryResp) Decode(reader io.Reader) error {
	if resp == nil {
		resp = &OpenQueryResp{}
	}
	return xml.NewDecoder(reader).Decode(resp)
}

func (resp *OpenQueryResp) QueryIds() ([]string, error) {
	if resp == nil {
		// TODO
	}

	if resp.ErrorResp.ErrorClientResponseMessage != "" {
		return nil, errors.New(resp.ErrorResp.ErrorClientResponseMessage)
	}

	var ids = make([]string, len(resp.Queries))
	for i, q := range resp.Queries {
		ids[i] = q.ID
	}
	return ids, nil
}

type CloseQueryResp struct {
	RequestXml string
	ErrorResp
}

type AnswerQueryResp struct {
	RequestXml string
	ErrorResp
}

type CancelQueryResp struct {
	RequestXml string
	ErrorResp
}

type ReQueryResp struct {
	XMLName                 xml.Name    `xml:"Response"`
	ReferenceNumber         string      `xml:"ReferenceNumber,attr"`
	InboundODMFileOID       string      `xml:"InboundODMFileOID,attr"`
	IsTransactionSuccessful string      `xml:"IsTransactionSuccessful,attr"`
	SuccessStatistics       string      `xml:"SuccessStatistics,attr"`
	NewRecords              string      `xml:"NewRecords,attr"`
	Queries                 []odm.Query `xml:"Query"`
	RequestXml              string
	ErrorResp
}

func (resp *ReQueryResp) Decode(reader io.Reader) error {
	if resp == nil {
		resp = &ReQueryResp{}
	}
	return xml.NewDecoder(reader).Decode(resp)
}

func (resp *ReQueryResp) QueryIds() ([]string, error) {
	if resp == nil {
		// TODO
	}

	if resp.ErrorResp.ErrorClientResponseMessage != "" {
		return nil, errors.New(resp.ErrorResp.ErrorClientResponseMessage)
	}

	var ids = make([]string, len(resp.Queries))
	for i, q := range resp.Queries {
		ids[i] = q.ID
	}
	return ids, nil
}
