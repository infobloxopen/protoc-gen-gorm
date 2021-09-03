package types

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/golang/protobuf/jsonpb"
)

// This Regex should match valid UUID format, in 8-4-4-4-12 or straight 32 byte format
var validChars = regexp.MustCompile("^[0-9a-f]{8}-?[0-9a-f]{4}-?[1-5][0-9a-f]{3}-?[89ab][0-9a-f]{3}-?[0-9a-f]{12}$")

// ZeroUUID The Zero value used for non-nil, but uninitialized UUID type
const ZeroUUID = "00000000-0000-0000-0000-000000000000"

// MarshalJSONPB overloads UUID's standard PB -> JSON conversion
func (m *UUID) MarshalJSONPB(*jsonpb.Marshaler) ([]byte, error) {
	if len(m.Value) == 0 {
		return []byte(fmt.Sprintf(`%q`, ZeroUUID)), nil
	}
	return []byte(fmt.Sprintf(`%q`, m.Value)), nil
}

// UnmarshalJSONPB overloads UUID's standard JSON -> PB conversion. If
// data is null, can't create nil object, but will marshal as null later
func (m *UUID) UnmarshalJSONPB(_ *jsonpb.Unmarshaler, data []byte) error {
	if string(data) == "null" {
		m.Value = ZeroUUID
		return nil
	}
	m.Value = strings.Trim(string(data), `"`)
	if !validChars.Match([]byte(m.Value)) {
		return fmt.Errorf(`invalid uuid '%s' does not match accepted format`, m.Value)
	}
	return nil
}

// MarshalJSONPB overloads UUIDValue's standard PB -> JSON conversion
func (m *UUIDValue) MarshalJSONPB(*jsonpb.Marshaler) ([]byte, error) {
	if len(m.Value) == 0 {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`%q`, m.Value)), nil
}

// UnmarshalJSONPB overloads UUIDValue's standard JSON -> PB conversion. If
// data is null, can't create nil object, but will marshal as null later
func (m *UUIDValue) UnmarshalJSONPB(_ *jsonpb.Unmarshaler, data []byte) error {
	if string(data) == "null" {
		m.Value = ""
		return nil
	}
	m.Value = strings.Trim(string(data), `"`)
	if !validChars.Match([]byte(m.Value)) {
		return fmt.Errorf(`invalid uuid '%s' does not match accepted format`, m.Value)
	}
	return nil
}

// MarshalJSONPB overloads JSONValue's standard PB -> JSON conversion
func (m *JSONValue) MarshalJSONPB(*jsonpb.Marshaler) ([]byte, error) {
	if len(m.Value) == 0 {
		return []byte("null"), nil
	}
	return []byte(m.Value), nil
}

// UnmarshalJSONPB overloads JSONValue's standard JSON -> PB conversion. If
// data is null, can't create nil object, but will marshal as null later
func (m *JSONValue) UnmarshalJSONPB(_ *jsonpb.Unmarshaler, data []byte) error {
	if string(data) == "null" {
		m.Value = ""
		return nil
	}
	m.Value = string(data)
	return nil
}

// MarshalJSONPB overloads InetValue's standard PB -> JSON conversion
func (m *InetValue) MarshalJSONPB(*jsonpb.Marshaler) ([]byte, error) {
	if len(m.Value) == 0 {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`%q`, m.Value)), nil
}

// UnmarshalJSONPB overloads InetValue's standard JSON -> PB conversion. If
// data is null, can't create nil object, but will marshal as null later
func (m *InetValue) UnmarshalJSONPB(_ *jsonpb.Unmarshaler, data []byte) error {
	if string(data) == "null" {
		m.Value = ""
		return nil
	}
	// Very minimal check for validity, if not a quoted string fails
	// Additional validation as a valid inet done in conversion to ORM type or
	// must be performed manually
	if data[0] != '"' && data[len(data)-1] != '"' {
		return fmt.Errorf(`invalid inet '%s' does not match accepted format`, data)
	}
	m.Value = strings.Trim(string(data), `"`)
	return nil
}

func (t *TimeOnly) MarshalJSONPB(*jsonpb.Marshaler) ([]byte, error) {
	timeStr, err := t.StringRepresentation()
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf(`%q`, timeStr)), nil
}

func (t *TimeOnly) UnmarshalJSONPB(_ *jsonpb.Unmarshaler, data []byte) error {
	if data[0] == '"' && data[len(data)-1] == '"' {
		data = data[1 : len(data)-1]
	}
	strTime := string(data)
	timeOnly, err := TimeOnlyByString(strTime)
	if err != nil {
		return err
	}
	t.Value = timeOnly.Value
	return nil
}
