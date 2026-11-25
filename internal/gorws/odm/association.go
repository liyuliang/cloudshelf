package odm

type Association struct {
	KeySet             KeySet       `xml:"KeySet,omitempty"`
	Annotation         []Annotation `xml:"Annotation,omitempty"`
	StudyOID           string       `xml:"StudyOID,attr,omitempty"`
	MetaDataVersionOID string       `xml:"MetaDataVersionOID,attr,omitempty"`
}

type KeySet struct {
	StudyOID            string `xml:"StudyOID,attr,omitempty"`
	SubjectKey          string `xml:"SubjectKey,attr,omitempty"`
	StudyEventOID       string `xml:"StudyEventOID,attr,omitempty"`
	StudyEventRepeatKey string `xml:"StudyEventRepeatKey,attr,omitempty"`
	FormOID             string `xml:"FormOID,attr,omitempty"`
	FormRepeatKey       string `xml:"FormRepeatKey,attr,omitempty"`
	ItemGroupOID        string `xml:"ItemGroupOID,attr,omitempty"`
	ItemGroupRepeatKey  string `xml:"ItemGroupRepeatKey,attr,omitempty"`
	ItemOID             string `xml:"ItemOID,attr,omitempty"`
	OID                 string `xml:"OID,attr,omitempty"`
}
