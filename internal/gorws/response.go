package rws

import (
	"encoding/xml"
	"io"
)

type Response interface {
	Decode(reader io.Reader) error
}

type ErrorResp struct {
	ReasonCode                 string `xml:"ReasonCode,attr,omitempty"`
	ErrorOriginLocation        string `xml:"ErrorOriginLocation,attr,omitempty"`
	ErrorClientResponseMessage string `xml:"ErrorClientResponseMessage,attr,omitempty"`
}

func (resp *ErrorResp) Decode(reader io.Reader) error {
	if resp == nil {
		resp = &ErrorResp{}
	}
	return xml.NewDecoder(reader).Decode(resp)
}