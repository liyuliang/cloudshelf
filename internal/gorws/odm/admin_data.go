package odm

type AdminData struct {
	User         []User         `xml:"User,omitempty"`
	Location     []Location     `xml:"Location,omitempty"`
	SignatureDef []SignatureDef `xml:"SignatureDef,omitempty"`
	StudyOID     string         `xml:"StudyOID,attr,omitempty"`
}

type User struct {
	LoginName    *UserLoginName    `xml:"LoginName,omitempty"`
	DisplayName  *UserDisplayName  `xml:"DisplayName,omitempty"`
	FullName     *UserFullName     `xml:"FullName,omitempty"`
	FirstName    *UserFirstName    `xml:"FirstName,omitempty"`
	LastName     *UserLastName     `xml:"LastName,omitempty"`
	Organization *UserOrganization `xml:"Organization,omitempty"`
	Address      []Address         `xml:"Address,omitempty"`
	Email        []UserEmail       `xml:"Email,omitempty"`
	Picture      *Picture          `xml:"Picture,omitempty"`
	Pager        *UserPager        `xml:"Pager,omitempty"`
	Fax          []UserFax         `xml:"Fax,omitempty"`
	Phone        []UserPhone       `xml:"Phone,omitempty"`
	LocationRef  []LocationRef     `xml:"LocationRef,omitempty"`
	Certificate  []UserCertificate `xml:"Certificate,omitempty"`
	OID          string            `xml:"OID,attr,omitempty"`
}

type UserLoginName struct {
	Value string `xml:",chardata,omitempty"`
}
type UserDisplayName struct {
	Value string `xml:",chardata,omitempty"`
}
type UserFullName struct {
	Value string `xml:",chardata,omitempty"`
}
type UserFirstName struct {
	Value string `xml:",chardata,omitempty"`
}
type UserLastName struct {
	Value string `xml:",chardata,omitempty"`
}
type UserOrganization struct {
	Value string `xml:",chardata,omitempty"`
}
type UserEmail struct {
	Value string `xml:",chardata,omitempty"`
}
type UserPager struct {
	Value string `xml:",chardata,omitempty"`
}
type UserFax struct {
	Value string `xml:",chardata,omitempty"`
}
type UserPhone struct {
	Value string `xml:",chardata,omitempty"`
}
type UserCertificate struct {
	Value string `xml:",chardata,omitempty"`
}

type Address struct {
	StreetName []AddressStreetName `xml:"StreetName,omitempty"`
	City       []AddressCity       `xml:"City,omitempty"`
	StateProv  *AddressStateProv   `xml:"StateProv,omitempty"`
	Country    *AddressCountry     `xml:"Country,omitempty"`
	PostalCode *AddressPostalCode  `xml:"PostalCode,omitempty"`
	OtherText  *AddressOtherText   `xml:"OtherText,omitempty"`
}

type AddressStreetName struct {
	Value string `xml:",chardata,omitempty"`
}
type AddressCity struct {
	Value string `xml:",chardata,omitempty"`
}
type AddressStateProv struct {
	Value string `xml:",chardata,omitempty"`
}
type AddressCountry struct {
	Value string `xml:",chardata,omitempty"`
}
type AddressPostalCode struct {
	Value string `xml:",chardata,omitempty"`
}
type AddressOtherText struct {
	Value string `xml:",chardata,omitempty"`
}

type Picture struct {
	PictureFileName string `xml:"PictureFileName,attr,omitempty"`
	ImageType       string `xml:"ImageType,attr,omitempty"`
}
type LocationRef struct {
	LocationOID string `xml:"LocationOID,attr,omitempty"`
}
type Location struct {
	MetaDataVersionRef []MetaDataVersionRef `xml:"MetaDataVersionRef,omitempty"`
	OID                string               `xml:"OID,attr,omitempty"`
	Name               string               `xml:"Name,attr,omitempty"`
	LocationType       string               `xml:"LocationType,attr,omitempty"` // Sponsor, Site, CRO, Lab, Other
}

type MetaDataVersionRef struct {
	StudyOID           string `xml:"StudyOID,attr,omitempty"`
	MetaDataVersionOID string `xml:"MetaDataVersionOID,attr,omitempty"`
	EffectiveDate      string `xml:"EffectiveDate,attr,omitempty"` // date?
	StudySiteNumber    string `xml:"StudySiteNumber,attr"`
}
type SignatureDef struct {
	Meaning     *SignatureDefMeaning     `xml:"Meaning,omitempty"`
	LegalReason *SignatureDefLegalReason `xml:"LegalReason,omitempty"`
	OID         string                   `xml:"OID,attr,omitempty"`
	Methodology string                   `xml:"Methodology,attr,omitempty"` // Digital, Electronic
}

type SignatureDefMeaning struct {
	Value string `xml:",chardata,omitempty"`
}
type SignatureDefLegalReason struct {
	Value string `xml:",chardata,omitempty"`
}
