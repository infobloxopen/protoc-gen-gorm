package types

import (
	"github.com/golang/protobuf/jsonpb"
)

// MarshalJSONPB overloads UUIDValue's standard PB -> JSON conversion
func (m *UUIDValue) MarshalJSONPB(*jsonpb.Marshaler) ([]byte, error) {
	return []byte(m.Value), nil
}

// UnmarshalJSONPB overloads UUIDValue's standard JSON -> PB conversion
func (m *UUIDValue) UnmarshalJSONPB(_ *jsonpb.Unmarshaler, data []byte) error {
	m.Value = string(data)
	return nil
}

// MarshalJSONPB overloads JSONValue's standard PB -> JSON conversion
func (m *JSONValue) MarshalJSONPB(*jsonpb.Marshaler) ([]byte, error) {
	return []byte(m.Value), nil
}

// UnmarshalJSONPB overloads JSONValue's standard JSON -> PB conversion
func (m *JSONValue) UnmarshalJSONPB(_ *jsonpb.Unmarshaler, data []byte) error {
	m.Value = string(data)
	return nil
}
