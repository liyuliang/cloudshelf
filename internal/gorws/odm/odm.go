package odm

type ODM struct {
	Studies             []*Study         `xml:"Study,omitempty"`
	AdminData           []*AdminData     `xml:"AdminData,omitempty"`
	ReferenceData       []*ReferenceData `xml:"ReferenceData,omitempty"`
	ClinicalData        []*ClinicalData  `xml:"ClinicalData,omitempty"`
	Association         *Association     `xml:"Association,omitempty"`
	Signature           *Signature       `xml:"Signature,omitempty"`
	Description         string           `xml:"Description,attr,omitempty"`
	FileType            string           `xml:"FileType,attr,omitempty"`    // Snapshot, Transactional
	Granularity         string           `xml:"Granularity,attr,omitempty"` // All, Metadata, AdminData, ReferenceData, AllClinicalData, SingleSite, SingleSubject
	Archival            string           `xml:"Archival,attr,omitempty"`    // Yes only
	FileOID             string           `xml:"FileOID,attr,omitempty"`
	CreationDateTime    string           `xml:"CreationDateTime,attr,omitempty"`
	PriorFileOID        string           `xml:"PriorFileOID,attr,omitempty"`
	AsOfDateTime        string           `xml:"AsOfDateTime,attr,omitempty"`
	ODMVersion          string           `xml:"ODMVersion,attr,omitempty"`
	Originator          string           `xml:"Originator,attr,omitempty"`
	SourceSystem        string           `xml:"SourceSystem,attr,omitempty"`
	SourceSystemVersion string           `xml:"SourceSystemVersion,attr,omitempty"`
	ID                  string           `xml:"ID,attr,omitempty"` // Id or ID
	XMLNs               string           `xml:"xmlns,attr,omitempty"`
	XMLNsMdsol          string           `xml:"xmlns:mdsol,attr,omitempty"`
	XMLNsXlink          string           `xml:"xmlns:xlink,attr,omitempty"`
}
