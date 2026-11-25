package odm

type ClinicalData struct {
	SubjectData        []*SubjectData  `xml:"SubjectData,omitempty"`
	AuditRecords       []*AuditRecords `xml:"AuditRecords,omitempty"`
	Signatures         []*Signatures   `xml:"Signatures,omitempty"`
	Annotations        []*Annotations  `xml:"Annotations,omitempty"`
	StudyOID           string          `xml:"StudyOID,attr,omitempty"`
	MetaDataVersionOID string          `xml:"MetaDataVersionOID,attr,omitempty"`
}

type SubjectData struct {
	AuditRecord      *AuditRecord      `xml:"AuditRecord,omitempty"`
	Signature        *Signature        `xml:"Signature,omitempty"`
	InvestigatorRef  *InvestigatorRef  `xml:"InvestigatorRef,omitempty"`
	SiteRef          *SiteRef          `xml:"SiteRef,omitempty"`
	Annotation       []*Annotation     `xml:"Annotation,omitempty"`
	StudyEventData   *StudyEventData   `xml:"StudyEventData,omitempty"`
	SubjectGroupData *SubjectGroupData `xml:"SubjectGroupData,omitempty"`
	AuditLogs        *AuditLogs        `xml:"AuditLogs,omitempty"`
	DiscrepancyNotes *DiscrepancyNotes `xml:"DiscrepancyNotes,omitempty"`
	StudySubjectID   string            `xml:"StudySubjectID,omitempty"`
	UniqueIdentifier string            `xml:"UniqueIdentifier,omitempty"`
	Status           string            `xml:"Status,omitempty"`
	SecondaryID      string            `xml:"SecondaryID,omitempty"`
	DateOfBirth      string            `xml:"DateOfBirth,omitempty"`
	YearOfBirth      int64             `xml:"YearOfBirth,omitempty"`
	Sex              string            `xml:"Sex,omitempty"`
	SubjectKey       string            `xml:"SubjectKey,attr,omitempty"`
	TransactionType  string            `xml:"TransactionType,attr,omitempty"`
	SubjectKeyType   string            `xml:"http://www.mdsol.com/ns/odm/metadata SubjectKeyType,attr,omitempty"`
}

type InvestigatorRef struct {
	UserOID string `xml:"UserOID,attr,omitempty"`
}

type SiteRef struct {
	LocationOID        string `xml:"LocationOID,attr,omitempty"`
	StudyEnvSiteNumber string `xml:"http://www.mdsol.com/ns/odm/metadata StudyEnvSiteNumber,attr,omitempty"`
}

type StudyEventData struct {
	AuditRecord         *AuditRecord      `xml:"AuditRecord,omitempty"`
	Signature           *Signature        `xml:"Signature,omitempty"`
	Annotation          []Annotation      `xml:"Annotation,omitempty"`
	FormData            *FormData         `xml:"FormData,omitempty"`
	AuditLogs           *AuditLogs        `xml:"AuditLogs,omitempty"`
	DiscrepancyNotes    *DiscrepancyNotes `xml:"DiscrepancyNotes,omitempty"`
	StudyEventOID       string            `xml:"StudyEventOID,attr,omitempty"`
	StudyEventRepeatKey string            `xml:"StudyEventRepeatKey,attr,omitempty"`
	TransactionType     string            `xml:"TransactionType,attr,omitempty"`
	StudyEventLocation  string            `xml:"StudyEventLocation,attr,omitempty"`
	StartDate           string            `xml:"StartDate,attr,omitempty"`
	EndDate             string            `xml:"EndDate,attr,omitempty"`
	Status              string            `xml:"Status,attr,omitempty"`
	SubjectAgeAtEvent   int64             `xml:"SubjectAgeAtEvent,attr,omitempty"`
}

type FormData struct {
	AuditRecord      *AuditRecord      `xml:"AuditRecord,omitempty"`
	Signature        *Signature        `xml:"Signature,omitempty"`
	ArchiveLayoutRef *ArchiveLayoutRef `xml:"ArchiveLayoutRef,omitempty"`
	Annotation       []Annotation      `xml:"Annotation,omitempty"`
	ItemGroupData    *ItemGroupData    `xml:"ItemGroupData,omitempty"`
	AuditLogs        *AuditLogs        `xml:"AuditLogs,omitempty"`
	DiscrepancyNotes *DiscrepancyNotes `xml:"DiscrepancyNotes,omitempty"`
	Version          string            `xml:"Version,attr,omitempty"`
	InterviewerName  string            `xml:"InterviewerName,attr,omitempty"`
	InterviewDate    string            `xml:"InterviewDate,attr,omitempty"`
	Status           string            `xml:"Status,attr,omitempty"`
	FormOID          string            `xml:"FormOID,attr,omitempty"`
	FormRepeatKey    string            `xml:"FormRepeatKey,attr,omitempty"`
	TransactionType  string            `xml:"TransactionType,attr,omitempty"`
}

type ArchiveLayoutRef struct {
	ArchiveLayoutOID string `xml:"ArchiveLayoutOID,attr,omitempty"`
}

type SubjectGroupData struct {
	StudyGroupClassID   string `xml:"StudyGroupClassID,attr,omitempty"`
	StudyGroupClassName string `xml:"StudyGroupClassName,attr,omitempty"`
	StudyGroupName      string `xml:"StudyGroupName,attr,omitempty"`
}

type AuditRecords struct {
	AuditLog []AuditLog `xml:"AuditLog,omitempty"`
	EntityID string     `xml:"EntityID,attr,omitempty"`
}

type Signatures struct {
	Signature []Signature `xml:"Signature,omitempty"`
}

type Annotations struct {
	Annotation []Annotation `xml:"Annotation,omitempty"`
}
