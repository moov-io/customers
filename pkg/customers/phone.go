package customers

import (
	"fmt"
	"strings"
)

func phoneTypeToModel(v string) (PhoneType, error) {
	v = strings.ToLower(v)
	m := map[string]PhoneType{
		"home":   PhoneType_Home,
		"mobile": PhoneType_Mobile,
		"work":   PhoneType_Work,
	}

	phoneType, ok := m[v]
	if !ok {
		return 0, fmt.Errorf("unknown phone type: '%s'", v)
	}
	return phoneType, nil
}

type PhoneType int

const (
	PhoneType_Home   PhoneType = 1
	PhoneType_Mobile PhoneType = 2
	PhoneType_Work   PhoneType = 3
)

func (a PhoneType) Common() string {
	switch a {
	case PhoneType_Home:
		return "home"
	case PhoneType_Mobile:
		return "mobile"
	case PhoneType_Work:
		return "work"
	}
	return ""
}

func (a PhoneType) String() string {
	return fmt.Sprintf("'%s'", a.Common())
}
