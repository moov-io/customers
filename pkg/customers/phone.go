package customers

import (
	"fmt"
	"strings"

	"github.com/moov-io/customers/pkg/client"
)

func phoneTypeToModel(v client.PhoneType) (PhoneType, error) {
	v = client.PhoneType(strings.ToLower(string(v)))
	m := map[client.PhoneType]PhoneType{
		client.PHONETYPE_HOME:   PhoneType_Home,
		client.PHONETYPE_MOBILE: PhoneType_Mobile,
		client.PHONETYPE_WORK:   PhoneType_Work,
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

func (a PhoneType) Common() client.PhoneType {
	switch a {
	case PhoneType_Home:
		return client.PHONETYPE_HOME
	case PhoneType_Mobile:
		return client.PHONETYPE_MOBILE
	case PhoneType_Work:
		return client.PHONETYPE_WORK
	}
	return ""
}

func (a PhoneType) String() string {
	return fmt.Sprintf("'%s'", a.Common())
}
