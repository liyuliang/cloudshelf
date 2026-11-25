package rws

import (
	"cloudshelf/internal/gorws/odm"
	"encoding/xml"
	"io"
)

type ODMResp struct {
	*odm.ODM
}

func (resp *ODMResp) Decode(reader io.Reader) error {
	if resp.ODM == nil {
		resp.ODM = &odm.ODM{}
	}
	return xml.NewDecoder(reader).Decode(resp.ODM)
}
