package odm

import (
	"encoding/xml"
	"strings"
)

// YesNoBool 是一个自定义布尔类型，用于解析 ODM XML 中的 "Yes"/"No" 属性值。
// Go 标准库的 xml.Unmarshal 对 bool 类型只接受 "true"/"false"/"1"/"0"，
// 但 CDISC ODM 规范使用 "Yes"/"No"，因此需要自定义解析。
type YesNoBool bool

// UnmarshalXMLAttr 实现 xml.UnmarshalerAttr 接口，支持解析以下值为 true：
//
//	"Yes", "yes", "YES", "true", "True", "TRUE", "1"
//
// 其余值均视为 false。
func (b *YesNoBool) UnmarshalXMLAttr(attr xml.Attr) error {
	switch strings.ToLower(attr.Value) {
	case "yes", "true", "1":
		*b = true
	default:
		*b = false
	}
	return nil
}

// MarshalXMLAttr 实现 xml.MarshalerAttr 接口，将布尔值序列化为 "Yes"/"No"。
func (b YesNoBool) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	val := "No"
	if b {
		val = "Yes"
	}
	return xml.Attr{Name: name, Value: val}, nil
}
