package types

import (
	"reflect"
	"strings"
	"testing"

	"github.com/golang/protobuf/jsonpb"
)

// WrapperMessage implements protobuf.Message but is not a normal generated message type.
type WrapperMessage struct {
	JSON      *JSONValue `protobuf:"bytes,1,opt,name=json,json=json" json:"json,omitempty"`
	UUIDValue *UUIDValue `protobuf:"bytes,2,opt,name=uuid_value,json=uuid_value" json:"uuid_value,omitempty"`
	UUID      *UUID      `protobuf:"bytes,3,opt,name=uuid,json=uuid" json:"uuid,omitempty"`
	Inet      *InetValue `protobuf:"bytes,4,opt,name=inet,json=inet" json:"inet,omitempty"`
}

func (m *WrapperMessage) Reset() {
	m.JSON = nil
}

func (m *WrapperMessage) String() string {
	return "null"
}

func (m *WrapperMessage) ProtoMessage() {
}

func TestSuccessfulUnmarshalTypes(t *testing.T) {
	unmarshaler := &jsonpb.Unmarshaler{}
	for in, expected := range map[string]WrapperMessage{
		`{}`: {JSON: nil, UUID: nil},
		// Can't unmarshal 'null' to nil like a WKT, only an invalid, empty state
		// which will be remarshalled to 'null'
		`{"json":null}`:       {JSON: &JSONValue{}},
		`{"uuid_value":null}`: {UUIDValue: &UUIDValue{}},
		// Still can't unmarshal 'null' to nil, but will initialize to zero-UUID
		`{"uuid":null}`:                                            {UUID: &UUID{Value: "00000000-0000-0000-0000-000000000000"}},
		`{"json":    {"key": "value"}}`:                            {JSON: &JSONValue{Value: `{"key": "value"}`}},
		`{"uuid_value":  "6ba7b810-9dad-11d1-80b4-00c04fd430c8" }`: {UUIDValue: &UUIDValue{Value: `6ba7b810-9dad-11d1-80b4-00c04fd430c8`}},
		`{"uuid_value":  "6ba7b8109dad11d180b400c04fd430c8" }`:     {UUIDValue: &UUIDValue{Value: `6ba7b8109dad11d180b400c04fd430c8`}},
		`{"inet":  "1.2.3.4"}`:                                     {Inet: &InetValue{Value: `1.2.3.4`}},
		`{"inet":null}`:                                            {Inet: &InetValue{Value: ""}},
	} {
		jv := &WrapperMessage{}
		err := unmarshaler.Unmarshal(strings.NewReader(in), jv)
		if err != nil {
			t.Error(err.Error())
		}
		if !reflect.DeepEqual(*jv, expected) {
			t.Errorf("Expected unmarshaled output '%+v' did not match actual output '%+v'",
				expected, *jv)
		}
	}
}

func TestBrokenUnmarshalTypes(t *testing.T) {
	unmarshaler := &jsonpb.Unmarshaler{}
	for in, expected := range map[string]string{
		// A couple cases to demo standard json unmarshaling handling
		`{"}`: "unexpected EOF",
		`{"uuid":"6ba7b810-9dad-11d1-80b4-00c04fd430c8}`:        "unexpected EOF",
		`{"json":[1,2,3,4,`:                                     "unexpected EOF",
		`{"json":}`:                                             "invalid character '}' looking for beginning of value",
		`{"json":[1,2,3,4,]}`:                                   "invalid character ']' looking for beginning of value",
		`{"json":{"top":{"something":1},2]}}`:                   "invalid character '2' looking for beginning of object key string",
		`{"uuid_value":{"top":{"something":1}}}`:                "invalid uuid '{\"top\":{\"something\":1}}' does not match accepted format",
		`{"uuid_value":"6ba7b810-9dad-11d1-80b4-00c04fdX30c8"}`: "invalid uuid '6ba7b810-9dad-11d1-80b4-00c04fdX30c8' does not match accepted format",
		`{"uuid_value":6ba7b810-9dad-11d1-80b4-00c04fd430c8}`:   "invalid character 'b' after object key:value pair",
		`{"uuid_value":ba67b810-9dad-11d1-80b4-00c04fd430c8}`:   "invalid character 'b' looking for beginning of value",
		`{"inet": 1.2.3.4}`:                                     "invalid character '.' after object key:value pair",
		`{"inet": 1}`:                                           "invalid inet '1' does not match accepted format",
	} {
		err := unmarshaler.Unmarshal(strings.NewReader(in), &WrapperMessage{})
		if err == nil || err.Error() != expected {
			if err == nil {
				t.Errorf("Expected error %q, but got no error", expected)
			} else {
				t.Errorf("Expected error %q, but got %q", expected, err.Error())
			}
		}
	}
}

func TestMarshalTypes(t *testing.T) {
	marshaler := &jsonpb.Marshaler{OrigName: true, EmitDefaults: true}
	for expected, in := range map[string]WrapperMessage{
		`{"json":null,"uuid_value":null,"uuid":null,"inet":null}`:                                                     {},
		`{"json":null,"uuid_value":null,"uuid":"00000000-0000-0000-0000-000000000000","inet":null}`:                   {UUID: &UUID{}},
		`{"json":{"key": "value"},"uuid_value":"6ba7b810-9dad-11d1-80b4-00c04fd430c8","uuid":null,"inet":null}`:       {JSON: &JSONValue{Value: `{"key": "value"}`}, UUIDValue: &UUIDValue{Value: `6ba7b810-9dad-11d1-80b4-00c04fd430c8`}},
		`{"json":{"key": "value"},"uuid_value":"6ba7b810-9dad-11d1-80b4-00c04fd430c8","uuid":null,"inet":"10.0.0.1"}`: {JSON: &JSONValue{Value: `{"key": "value"}`}, UUIDValue: &UUIDValue{Value: `6ba7b810-9dad-11d1-80b4-00c04fd430c8`}, Inet: &InetValue{Value: `10.0.0.1`}},
		`{"json":null,"uuid_value":"00000000-0000-0000-0000-000000000000","uuid":null,"inet":null}`:                   {UUIDValue: &UUIDValue{Value: "00000000-0000-0000-0000-000000000000"}},
	} {
		out, err := marshaler.MarshalToString(&in)
		if err != nil {
			t.Error(err.Error())
		}
		if string(out) != expected {
			t.Errorf("Expected marshaled output '%s' did not match actual output '%s'",
				expected, out)
		}
	}
}

func TestMarshalTypesOmitEmpty(t *testing.T) {
	marshaller := &jsonpb.Marshaler{OrigName: true}
	for expected, in := range map[string]WrapperMessage{
		`{}`:                                                                            {},
		`{"json":null}`:                                                                 {JSON: &JSONValue{}},
		`{"uuid_value":null}`:                                                           {UUIDValue: &UUIDValue{}},
		`{"json":{"key": "value"}}`:                                                     {JSON: &JSONValue{Value: `{"key": "value"}`}},
		`{"uuid_value":"6ba7b810-9dad-11d1-80b4-00c04fd430c8"}`:                         {UUIDValue: &UUIDValue{Value: `6ba7b810-9dad-11d1-80b4-00c04fd430c8`}},
		`{"json":{"key": "value"},"uuid_value":"6ba7b810-9dad-11d1-80b4-00c04fd430c8"}`: {JSON: &JSONValue{Value: `{"key": "value"}`}, UUIDValue: &UUIDValue{Value: `6ba7b810-9dad-11d1-80b4-00c04fd430c8`}},
		`{"inet":null}`: {Inet: &InetValue{}},
	} {
		out, err := marshaller.MarshalToString(&in)
		if err != nil {
			t.Error(err.Error())
		}
		if string(out) != expected {
			t.Errorf("Expected marshaled output '%s' did not match actual output '%s'",
				expected, out)
		}
	}
}
