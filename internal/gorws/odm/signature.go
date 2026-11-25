package odm

type Signature struct {
	SignedInfo     SignedInfo
	SignatureValue string
	KeyInfo        string
	Object         string
	ID             string // Id or ID
}

type SignedInfo struct {
	CanonicalizationMethod CanonicalizationMethod
	SignatureMethod        SignatureMethod
	Reference              Reference
	ID                     string // Id or ID
}

type CanonicalizationMethod struct {
	Content   []any
	Algorithm string
}
type SignatureMethod struct {
	Content   []any
	Algorithm string
}
type Reference struct {
	Transforms   Transforms
	DigestMethod DigestMethod
	DigestValue  []byte
	ID           string
	URI          string
	Type         string
}

type Transforms struct {
	Transform []Transform
}
type Transform struct {
	Content   []any
	Algorithm string
}
type DigestMethod struct {
	Content   []any
	Algorithm string
}
