package odm

type Study struct {
	GlobalVariables  *GlobalVariables  `xml:"GlobalVariables,omitempty"`
	BasicDefinitions *BasicDefinitions `xml:"BasicDefinitions,omitempty"`
	MetaDataVersions []MetaDataVersion `xml:"MetaDataVersion,omitempty"`
	OID              string            `xml:"OID,attr,omitempty"`
}

type GlobalVariables struct {
	StudyName        string `xml:"StudyName,omitempty"`
	StudyDescription string `xml:"StudyDescription,omitempty"`
	ProtocolName     string `xml:"ProtocolName,omitempty"`
}

type BasicDefinitions struct {
	MeasurementUnits []MeasurementUnit `xml:"MeasurementUnits,omitempty"`
}

type MeasurementUnit struct {
	Symbol Symbol `xml:"Symbol,omitempty"`
	OID    string `xml:"OID,attr,omitempty"`
	Name   string `xml:"Name,attr,omitempty"`
}

type Symbol struct {
	TranslatedText []TranslatedText `xml:"TranslatedText,omitempty"`
}

type TranslatedText struct {
	Value    string `xml:",chardata,omitempty"`
	Language string `xml:"xml:lang,omitempty"`
}
