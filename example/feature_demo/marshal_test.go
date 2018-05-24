package example

import (
	"reflect"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/infobloxopen/protoc-gen-gorm/types"
)

func TestUnmarshalTypes(t *testing.T) {
	ts, _ := ptypes.TimestampProto(time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC))
	marshaller := &runtime.JSONPb{OrigName: true, EmitDefaults: true}
	for in, expected := range map[string]TestTypes{
		`{}`: {},
		`{"api_only_string":"important text"}`:            {ApiOnlyString: "important text"},
		`{"numbers":[1,2,3,4]}`:                           {Numbers: []int32{1, 2, 3, 4}},
		`{"optional_string":"something real"}`:            {OptionalString: &wrappers.StringValue{Value: "something real"}},
		`{"becomes_int":"GOOD"}`:                          {BecomesInt: 1},
		`{"uuid":"123abc-invalid-uuid"}`:                  {Uuid: &types.UUIDValue{Value: "123abc-invalid-uuid"}},
		`{"created_at":"2009-11-17T20:34:58.651387237Z"}`: {CreatedAt: ts},
		`{"type_with_id_id":4}`:                           {TypeWithIdId: 4},
		`{"json_field":{"top":[{"something":1},2]}}`:      {JsonField: &types.JSONValue{Value: `{"top":[{"something":1},2]}`}},
	} {
		tt := &TestTypes{}
		err := marshaller.Unmarshal([]byte(in), tt)
		if err != nil {
			t.Error(err.Error())
		}
		if !reflect.DeepEqual(*tt, expected) {
			t.Errorf("Expected unmarshaled output '%+v' did not match actual output '%+v'",
				expected, *tt)
		}
	}
}

func TestMarshalTypes(t *testing.T) {
	ts, _ := ptypes.TimestampProto(time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC))
	marshaller := &runtime.JSONPb{OrigName: true, EmitDefaults: true}
	for expected, in := range map[string]TestTypes{
		`{"api_only_string":"","numbers":[],"optional_string":null,"becomes_int":"UNKNOWN","nothingness":null,"uuid":null,"created_at":null,"type_with_id_id":0,"json_field":null}`:                             {},
		`{"api_only_string":"Something","numbers":[],"optional_string":null,"becomes_int":"UNKNOWN","nothingness":null,"uuid":null,"created_at":null,"type_with_id_id":0,"json_field":null}`:                    {ApiOnlyString: "Something"},
		`{"api_only_string":"","numbers":[0,1,2,3],"optional_string":null,"becomes_int":"UNKNOWN","nothingness":null,"uuid":null,"created_at":null,"type_with_id_id":0,"json_field":null}`:                      {Numbers: []int32{0, 1, 2, 3}},
		`{"api_only_string":"","numbers":[],"optional_string":"Not nothing","becomes_int":"UNKNOWN","nothingness":null,"uuid":null,"created_at":null,"type_with_id_id":0,"json_field":null}`:                    {OptionalString: &wrappers.StringValue{Value: "Not nothing"}},
		`{"api_only_string":"","numbers":[],"optional_string":null,"becomes_int":"GOOD","nothingness":null,"uuid":null,"created_at":null,"type_with_id_id":0,"json_field":null}`:                                {BecomesInt: 1},
		`{"api_only_string":"","numbers":[],"optional_string":null,"becomes_int":"UNKNOWN","nothingness":null,"uuid":"123abc-invalid-uuid","created_at":null,"type_with_id_id":0,"json_field":null}`:            {Uuid: &types.UUIDValue{Value: "123abc-invalid-uuid"}},
		`{"api_only_string":"","numbers":[],"optional_string":null,"becomes_int":"UNKNOWN","nothingness":null,"uuid":null,"created_at":"2009-11-17T20:34:58.651387237Z","type_with_id_id":0,"json_field":null}`: {CreatedAt: ts},
		`{"api_only_string":"","numbers":[],"optional_string":null,"becomes_int":"UNKNOWN","nothingness":null,"uuid":null,"created_at":null,"type_with_id_id":2,"json_field":null}`:                             {TypeWithIdId: 2},
		`{"api_only_string":"","numbers":[],"optional_string":null,"becomes_int":"UNKNOWN","nothingness":null,"uuid":null,"created_at":null,"type_with_id_id":0,"json_field":{"text":[]}}`:                      {JsonField: &types.JSONValue{Value: `{"text":[]}`}},
	} {
		out, err := marshaller.Marshal(&in)
		if err != nil {
			t.Error(err.Error())
		}
		if string(out) != expected {
			t.Errorf("Expected marshaled output '%s' did not match actual output '%s'",
				expected, out)
		}
	}

}
