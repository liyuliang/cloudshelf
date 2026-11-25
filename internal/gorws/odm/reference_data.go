package odm

type ReferenceData struct {
	ItemGroupData      []ItemGroupData `xml:"ItemGroupData,omitempty"`
	StudyOID           string          `xml:"StudyOID,attr,omitempty"`
	MetaDataVersionOID string          `xml:"MetaDataVersionOID,attr,omitempty"`
}

type ItemGroupData struct {
	AuditRecord                *AuditRecord `xml:"AuditRecord,omitempty"`
	Signature                  *Signature   `xml:"Signature,omitempty"`
	Annotation                 []Annotation `xml:"Annotation,omitempty"`
	ItemData                   []ItemData   `xml:"ItemData,omitempty"`
	ItemDataInteger            []string     `xml:"ItemDataInteger,omitempty"`
	ItemDataBoolean            []string     `xml:"ItemDataBoolean,omitempty"`
	ItemDataFloat              []string     `xml:"ItemDataFloat,omitempty"`
	ItemDataHexFloat           []string     `xml:"ItemDataHexFloat,omitempty"`
	ItemDataAny                []string     `xml:"ItemDataAny,omitempty"`
	ItemDataPartialDatetime    []string     `xml:"ItemDataPartialDatetime,omitempty"`
	ItemDataDatetime           []string     `xml:"ItemDataDatetime,omitempty"`
	ItemDataIntervalDatetime   []string     `xml:"ItemDataIntervalDatetime,omitempty"`
	ItemDataDouble             []string     `xml:"ItemDataDouble,omitempty"`
	ItemDataPartialTime        []string     `xml:"ItemDataPartialTime,omitempty"`
	ItemDataBase64Binary       []string     `xml:"ItemDataBase64Binary,omitempty"`
	ItemDataPartialDate        []string     `xml:"ItemDataPartialDate,omitempty"`
	ItemDataBase64Float        []string     `xml:"ItemDataBase64Float,omitempty"`
	ItemDataHexBinary          []string     `xml:"ItemDataHexBinary,omitempty"`
	ItemDataDurationDatetime   []string     `xml:"ItemDataDurationDatetime,omitempty"`
	ItemDataDate               []string     `xml:"ItemDataDate,omitempty"`
	ItemDataURI                []string     `xml:"ItemDataURI,omitempty"`
	ItemDataIncompleteDatetime []string     `xml:"ItemDataIncompleteDatetime,omitempty"`
	ItemDataTime               []string     `xml:"ItemDataTime,omitempty"`
	ItemDataString             []string     `xml:"ItemDataString,omitempty"`
	ItemGroupOID               string       `xml:"ItemGroupOID,attr,omitempty"`
	ItemGroupRepeatKey         string       `xml:"ItemGroupRepeatKey,attr,omitempty"`
	TransactionType            string       `xml:"TransactionType,attr,omitempty"` // Insert, Update, Remove, Upsert, Context
}

type AuditRecord struct {
	UserRef              *UserRef     `xml:"UserRef,omitempty"`
	LocationRef          *LocationRef `xml:"LocationRef,omitempty"`
	DateTimeStamp        string       `xml:"DateTimeStamp,omitempty"`
	ReasonForChange      string       `xml:"ReasonForChange,omitempty"`
	SourceID             string       `xml:"SourceID,omitempty"`
	EditPoint            string       `xml:"EditPoint,attr,omitempty"`            // Monitoring, DataManagement, DBAudit
	UsedImputationMethod YesNoBool    `xml:"UsedImputationMethod,attr,omitempty"` // Yes or No
	ID                   string       `xml:"ID,attr,omitempty"`
}

type UserRef struct {
	UserOID string `xml:"UserOID,attr,omitempty"`
}

type Annotation struct {
	Comment         Comment `xml:"Comment,omitempty"`
	Flag            Flag    `xml:"Flag,omitempty"`
	SeqNum          int64   `xml:"SeqNum,attr,omitempty"`
	TransactionType string  `xml:"TransactionType,attr,omitempty"` // Insert, Update, Remove, Upsert, Context
	ID              string  `xml:"ID,attr,omitempty"`
}

type Comment struct {
	Value         string `xml:"Value,attr,omitempty"`
	SponsorOrSite string `xml:"SponsorOrSite,attr,omitempty"` //  Sponsor or Site
}

type Flag struct {
	FlagValue FlagValue `xml:"FlagValue,omitempty"`
	FlagType  FlagType  `xml:"FlagType,omitempty"`
}

type FlagValue struct {
	Value       string `xml:"Value,attr,omitempty"`
	CodeListOID string `xml:"CodeListOID,attr,omitempty"`
}

type FlagType struct {
	Value       string `xml:"Value,attr,omitempty"`
	CodeListOID string `xml:"CodeListOID,attr,omitempty"`
}

type ItemData struct {
	AuditRecord        *AuditRecord        `xml:"AuditRecord,omitempty"`
	Signature          *Signature          `xml:"Signature,omitempty"`
	MeasurementUnitRef *MeasurementUnitRef `xml:"MeasurementUnitRef,omitempty"`
	Annotation         []Annotation        `xml:"Annotation,omitempty"`
	AuditLogs          []AuditLogs         `xml:"AuditLogs,omitempty"`
	DiscrepancyNotes   []DiscrepancyNotes  `xml:"DiscrepancyNotes,omitempty"`
	Value              string              `xml:"Value,attr"`
	SpecifyValue       string              `xml:"mdsol:SpecifyValue,attr,omitempty"`
	SubmitMarkingOnly  string              `xml:"mdsol:SubmitMarkingOnly,attr,omitempty"`
	ReasonForNull      string              `xml:"ReasonForNull,attr,omitempty"`
	ItemOID            string              `xml:"ItemOID,attr,omitempty"`
	TransactionType    string              `xml:"TransactionType,attr,omitempty"` // Insert, Update, Remove, Upsert, Context
	IsNull             string              `xml:"IsNull,attr,omitempty"`          // yes only
	Query              *Query              `xml:"http://www.mdsol.com/ns/odm/metadata Query,omitempty"`
}

type Query struct {
	ID                      string `xml:"ID,attr,omitempty"`
	FieldOID                string `xml:"FieldOID,attr,omitempty"`
	Recipient               string `xml:"Recipient,attr,omitempty"`
	Value                   string `xml:"Value,attr,omitempty"`
	QueryRepeatKey          string `xml:"QueryRepeatKey,attr,omitempty"`
	Status                  string `xml:"Status,attr,omitempty"`
	Response                string `xml:"Response,attr,omitempty"`
	RequiresResponse        string `xml:"RequiresResponse,attr,omitempty"`
	PrecedingQueryRepeatKey string `xml:"PrecedingQueryRepeatKey,attr,omitempty"`
}

type AuditLogs struct {
	AuditLog []AuditLog `xml:"AuditLog,omitempty"`
	EntityID string     `xml:"EntityID,attr,omitempty"`
}

type AuditLog struct {
	ID              string `xml:"ID,attr,omitempty"`
	UserID          string `xml:"UserID,attr,omitempty"`
	DateTimeStamp   string `xml:"DateTimeStamp,attr,omitempty"`
	Type            string `xml:"Type,attr,omitempty"`
	ReasonForChange string `xml:"ReasonForChange,attr,omitempty"`
	OldValue        string `xml:"OldValue,attr,omitempty"`
	NewValue        string `xml:"NewValue,attr,omitempty"`
}

type DiscrepancyNotes struct {
	DiscrepancyNote []DiscrepancyNote `xml:"DiscrepancyNote,omitempty"`
	EntityID        string            `xml:"EntityID,attr,omitempty"`
}

type DiscrepancyNote struct {
	ChildNote          []ChildNote `xml:"ChildNote,omitempty"`
	ID                 string      `xml:"ID,attr,omitempty"`
	NoteType           string      `xml:"NoteType,attr,omitempty"`
	Status             string      `xml:"Status,attr,omitempty"`
	DateUpdated        string      `xml:"DateUpdated,attr,omitempty"`
	NumberOfChildNotes int64       `xml:"NumberOfChildNotes,attr,omitempty"`
}

type ChildNote struct {
	Description  []string  `xml:"Description,omitempty"`
	DetailedNote []string  `xml:"DetailedNote,omitempty"`
	UserRef      []UserRef `xml:"UserRef,omitempty"`
	ID           string    `xml:"ID,attr,omitempty"`
	Status       string    `xml:"Status,attr,omitempty"`
	DateCreated  string    `xml:"DateCreated,attr,omitempty"`
}
