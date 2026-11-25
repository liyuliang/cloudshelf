package odm

type MetaDataVersion struct {
	OID              string             `xml:"OID,attr,omitempty"`
	Name             string             `xml:"Name,attr,omitempty"`
	Description      string             `xml:"Description,attr,omitempty"`
	Include          *Include           `xml:"Include,omitempty"`
	Protocol         *Protocol          `xml:"Protocol,omitempty"`
	StudyEventDef    []StudyEventDef    `xml:"StudyEventDef,omitempty"`
	FormDef          []FormDef          `xml:"FormDef,omitempty"`
	ItemGroupDef     []ItemGroupDef     `xml:"ItemGroupDef,omitempty"`
	ItemDef          []ItemDef          `xml:"ItemDef,omitempty"`
	CodeList         []CodeList         `xml:"CodeList,omitempty"`
	ImputationMethod []ImputationMethod `xml:"ImputationMethod,omitempty"`
	Presentation     []Presentation     `xml:"Presentation,omitempty"`
	ConditionDef     []ConditionDef     `xml:"ConditionDef,omitempty"`
	MethodDef        []MethodDef        `xml:"MethodDef,omitempty"`
}

type Include struct {
	StudyOID           string `xml:"StudyOID,attr,omitempty"`
	MetaDataVersionOID string `xml:"MetaDataVersionOID,attr,omitempty"`
}

type Protocol struct {
	Description   Description   `xml:"Description,omitempty"`
	StudyEventRef StudyEventRef `xml:"StudyEventRef,omitempty"`
}

type StudyEventRef struct {
	OrderNumber                     int64  `xml:"OrderNumber,attr,omitempty"`
	Mandatory                       string `xml:"Mandatory,attr,omitempty"` // Yes or no
	CollectionExceptionConditionOID string `xml:"CollectionExceptionConditionOID,attr,omitempty"`
	StudyEventOID                   string `xml:"StudyEventOID,attr,omitempty"`
}

type StudyEventDef struct {
	Description Description `xml:"Description,omitempty"`
	FormRefs    []FormRef   `xml:"FormRef,omitempty"`
	OID         string      `xml:"OID,attr,omitempty"`
	Name        string      `xml:"Name,attr,omitempty"`
	Repeating   YesNoBool   `xml:"Repeating,attr,omitempty"` // Yes or No
	Type        string      `xml:"Type,attr,omitempty"`      // Scheduled or Unscheduled or Common
	Category    string      `xml:"Category,attr,omitempty"`
}

type FormDef struct {
	Description   Description     `xml:"Description,omitempty"`
	ItemGroupRef  []ItemGroupRef  `xml:"ItemGroupRef,omitempty"`
	ArchiveLayout []ArchiveLayout `xml:"ArchiveLayout,omitempty"`
	OID           string          `xml:"OID,attr,omitempty"`
	Name          string          `xml:"Name,attr,omitempty"`
	Repeating     YesNoBool       `xml:"Repeating,attr,omitempty"` // Yes or No
}

type ItemGroupRef struct {
	OrderNumber                     int64  `xml:"OrderNumber,attr,omitempty"`
	Mandatory                       YesNoBool `xml:"Mandatory,attr,omitempty"` // Yes or no
	CollectionExceptionConditionOID string    `xml:"CollectionExceptionConditionOID,attr,omitempty"`
	ItemGroupOID                    string    `xml:"ItemGroupOID,attr,omitempty"`
}

type ArchiveLayout struct {
	OID             string `xml:"OID,attr,omitempty"`
	PdfFileName     string `xml:"PdfFileName,attr,omitempty"`
	PresentationOID string `xml:"PresentationOID,attr,omitempty"`
}

type ItemGroupDef struct {
	Description     Description `xml:"Description,omitempty"`
	ItemRef         []ItemRef   `xml:"ItemRef,omitempty"`
	Alias           []Alias     `xml:"Alias,omitempty"`
	OID             string      `xml:"OID,attr,omitempty"`
	Name            string      `xml:"Name,attr,omitempty"`
	Repeating       YesNoBool   `xml:"Repeating,attr,omitempty"`       // Yes or no
	IsReferenceData YesNoBool   `xml:"IsReferenceData,attr,omitempty"` // Yes or no
	SASDatasetName  string      `xml:"SASDatasetName,attr,omitempty"`
	Domain          string      `xml:"Domain,attr,omitempty"`
	Origin          string      `xml:"Origin,attr,omitempty"`
	Role            string      `xml:"Role,attr,omitempty"`
	Purpose         string      `xml:"Purpose,attr,omitempty"`
	Comment         string      `xml:"Comment,attr,omitempty"`
}

type ItemRef struct {
	OrderNumber                     int64    `xml:"OrderNumber,attr,omitempty"`
	Mandatory                       YesNoBool `xml:"Mandatory,attr,omitempty"` // Yes or no
	CollectionExceptionConditionOID string    `xml:"CollectionExceptionConditionOID,attr,omitempty"`
	ItemOID                         string    `xml:"ItemOID,attr,omitempty"`
	KeySequence                     int64    `xml:"KeySequence,attr,omitempty"`
	MethodOID                       string   `xml:"MethodOID,attr,omitempty"`
	ImputationMethodOID             string   `xml:"ImputationMethodOID,attr,omitempty"`
	Role                            []string `xml:"Role,attr,omitempty"`
	RoleCodeListOID                 string   `xml:"RoleCodeListOID,attr,omitempty"`
}

type Alias struct {
	Context string `xml:"Context,attr,omitempty"`
	Name    string `xml:"Name,attr,omitempty"`
}

type ItemDef struct {
	Description        *Description         `xml:"Description,omitempty"`
	Question           *Question            `xml:"Question,omitempty"`
	ExternalQuestion   *ExternalQuestion    `xml:"ExternalQuestion,omitempty"`
	MeasurementUnitRef []MeasurementUnitRef `xml:"MeasurementUnitRef,omitempty"`
	RangeCheck         []RangeCheck         `xml:"RangeCheck,omitempty"`
	CodeListRef        CodeListRef          `xml:"CodeListRef,omitempty"`
	Role               []Role               `xml:"Role,omitempty"`
	Alias              *Alias               `xml:"Alias,omitempty"`
	OID                string               `xml:"OID,attr,omitempty"`
	Name               string               `xml:"Name,attr,omitempty"`
	DataType           string               `xml:"DataType,attr,omitempty"` // enum: integer, float, date, datetime, time, text, string, double, boolean, hexBinary, base64Binary, hexFloat, base64Float, partialDate, partialTime, partialDatetime, durationDatetime, intervalDatetime, incompleteDatetime
	Length             int64                `xml:"Length,attr,omitempty"`
	SignificantDigits  int64                `xml:"SignificantDigits,attr,omitempty"`
	SASFieldName       string               `xml:"SASFieldName,attr,omitempty"`
	SDSVarName         string               `xml:"SDSVarName,attr,omitempty"`
	Origin             string               `xml:"Origin,attr,omitempty"`
	Comment            string               `xml:"Comment,attr,omitempty"`
}

type Question struct {
	TranslatedText []TranslatedText `xml:"TranslatedText,omitempty"`
	QuestionNumber string           `xml:"QuestionNumber,omitempty"`
}

type ExternalQuestion struct {
	Dictionary string `xml:"Dictionary,attr,omitempty"`
	Version    string `xml:"Version,attr,omitempty"`
	Code       string `xml:"Code,attr,omitempty"`
}

type MeasurementUnitRef struct {
	MeasurementUnitOID string `xml:"MeasurementUnitOID,attr,omitempty"`
}

type RangeCheck struct {
	CheckValue         []string            `xml:"CheckValue,omitempty"`
	FormalExpression   *FormalExpression   `xml:"FormalExpression,omitempty"`
	MeasurementUnitRef *MeasurementUnitRef `xml:"MeasurementUnitRef,omitempty"`
	ErrorMessage       *ErrorMessage       `xml:"ErrorMessage,omitempty"`
	Comparator         string              `xml:"Comparator,attr,omitempty"` // LT, LE, GT, GE, EQ, NE, IN, NOTIN
	SoftHard           string              `xml:"SoftHard,attr,omitempty"`   // Soft or Hard
}

// type CheckValue struct {
// }

type FormalExpression struct {
	Value   string `xml:",chardata,omitempty"`
	Context string `xml:"Context,attr,omitempty"`
}

type ErrorMessage struct {
	TranslatedText []TranslatedText `xml:"TranslatedText,omitempty"`
}

type CodeListRef struct {
	CodeListOID string `xml:"CodeListOID,attr,omitempty"`
}

type Role struct {
	Value string `xml:",chardata,omitempty"`
}

type CodeList struct {
	CodeListItem     []CodeListItem    `xml:"CodeListItem,omitempty"`
	ExternalCodeList *ExternalCodeList `xml:"ExternalCodeList,omitempty"`
	EnumeratedItem   []EnumeratedItem  `xml:"EnumeratedItem,omitempty"`
	OID              string            `xml:"OID,attr,omitempty"`
	Name             string            `xml:"Name,attr,omitempty"`
	DataType         string            `xml:"DataType,attr,omitempty"` // integer, float, text, string
	SASFormatName    string            `xml:"SASFormatName,attr,omitempty"`
}

type CodeListItem struct {
	Decode     *Decode `xml:"Decode,omitempty"`
	CodedValue string  `xml:"CodedValue,attr,omitempty"`
	Rank       float64 `xml:"Rank,attr,omitempty"`
}

type Decode struct {
	TranslatedText []TranslatedText `xml:"TranslatedText,omitempty"`
}

type ExternalCodeList struct {
	Dictionary string `xml:"Dictionary,attr,omitempty"`
	Version    string `xml:"Version,attr,omitempty"`
}

type EnumeratedItem struct {
	CodedValue string  `xml:"CodedValue,attr,omitempty"`
	Rank       float64 `xml:"Rank,attr,omitempty"`
}

type ImputationMethod struct {
	Value string `xml:",chardata,omitempty"`
	OID   string `xml:"OID,attr,omitempty"`
}

type Presentation struct {
	Value    string `xml:",chardata,omitempty"`
	OID      string `xml:"OID,attr,omitempty"`
	Language string `xml:"xml:lang,attr,omitempty"`
}

type ConditionDef struct {
	Description      *Description      `xml:"Description,omitempty"`
	FormalExpression *FormalExpression `xml:"FormalExpression,omitempty"`
	OID              string            `xml:"OID,attr,omitempty"`
	Name             string            `xml:"Name,attr,omitempty"`
}

type MethodDef struct {
	Description      *Description      `xml:"Description,omitempty"`
	FormalExpression *FormalExpression `xml:"FormalExpression,omitempty"`
	OID              string            `xml:"OID,attr,omitempty"`
	Name             string            `xml:"Name,attr,omitempty"`
	Type             string            `xml:"Type,attr,omitempty"` // Computation, Imputation, Transpose, Other
}

type Description struct {
	TranslatedText []TranslatedText `xml:"TranslatedText,omitempty"`
}

type FormRef struct {
	OrderNumber                     int64  `xml:"OrderNumber,attr,omitempty"`
	Mandatory                       YesNoBool `xml:"Mandatory,attr,omitempty"` // Yes or no
	CollectionExceptionConditionOID string    `xml:"CollectionExceptionConditionOID,attr,omitempty"`
	FormOID                         string    `xml:"FormOID,attr,omitempty"`
}
