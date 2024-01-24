package example

import (
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/go-cmp/cmp"
	"github.com/infobloxopen/protoc-gen-gorm/types"
	"google.golang.org/protobuf/testing/protocmp"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func TestSuccessfulUnmarshalTypes(t *testing.T) {
	unmarshaler := &jsonpb.Unmarshaler{}
	for in, expected := range map[string]TestTypes{
		`{}`:                                              {},
		`{"api_only_string":"important text"}`:            {ApiOnlyString: "important text"},
		`{"numbers":null}`:                                {},
		`{"numbers":[]}`:                                  {Numbers: []int32{}},
		`{"numbers":[1,2,3,4]}`:                           {Numbers: []int32{1, 2, 3, 4}},
		`{"optional_string":null}`:                        {},
		`{"optional_string":""}`:                          {OptionalString: &wrappers.StringValue{Value: ""}},
		`{"optional_string":"something real"}`:            {OptionalString: &wrappers.StringValue{Value: "something real"}},
		`{"becomes_int":"UNKNOWN"}`:                       {},
		`{"becomes_int":"GOOD"}`:                          {BecomesInt: TestTypes_GOOD},
		`{"becomes_int":"BAD"}`:                           {BecomesInt: TestTypes_BAD},
		`{"uuid":"6ba7b810-9dad-11d1-80b4-00c04fd430c8"}`: {Uuid: &types.UUID{Value: "6ba7b810-9dad-11d1-80b4-00c04fd430c8"}},
		`{"created_at":"2009-11-17T20:34:58.651387237Z"}`: {CreatedAt: timestamppb.New(time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC))},
		`{"type_with_id_id":4}`:                           {TypeWithIdId: 4},
		`{"json_field":{"top":[{"something":1},2]}}`:      {JsonField: &types.JSONValue{Value: `{"top":[{"something":1},2]}`}},
		`{"json_field":
  {"top":
    [
      {"something":1}
      ,2
    ]
  }
}`: {JsonField: &types.JSONValue{Value: `{"top":
    [
      {"something":1}
      ,2
    ]
  }`}},
	} {
		tt := &TestTypes{}
		err := unmarshaler.Unmarshal(strings.NewReader(in), tt)
		if err != nil {
			t.Error(err.Error())
		}
		if !cmp.Equal(*tt, expected, protocmp.Transform()) {
			t.Errorf("Expected unmarshaled output '%+v' did not match actual output '%+v'",
				expected, *tt)
		}
	}
}

func TestBrokenUnmarshalTypes(t *testing.T) {
	unmarshaler := &jsonpb.Unmarshaler{}
	for in, expected := range map[string]string{
		// A subset of possible broken inputs
		`{"}`:                                 "unexpected EOF",
		`{"becomes_int":"NOT_AN_ENUM_VALUE"}`: `unknown value "\"NOT_AN_ENUM_VALUE\"" for enum example.TestTypes.status`,
		`{"numbers":[1,2,3,4,]}`:              "invalid character ']' looking for beginning of value",
		`{"json_field":{"top":{"something":1},2]}}`: "invalid character '2' looking for beginning of object key string",
		`{"uuid":""}`: "invalid uuid '' does not match accepted format",
		`{"uuid":"   6ba7b810-9dad-11d1-80b4-00c04fd430c8"}`: "invalid uuid '   6ba7b810-9dad-11d1-80b4-00c04fd430c8' does not match accepted format",
		`{"time_only":"24:00:00"}`:                           "Hours value outside expected range: 24",
	} {
		t.Run(in, func(t *testing.T) {
			err := unmarshaler.Unmarshal(strings.NewReader(in), &TestTypes{})
			if err == nil {
				t.Errorf("No error, expected: %q", expected)
				return
			}
			if err.Error() != expected {
				t.Errorf("\ngot:    %s\nwanted: %s", err, expected)
			}
		})
	}
}

func TestMarshalTypes(t *testing.T) {
	// Will marshal with snake_case names and default values included
	marshaler := &jsonpb.Marshaler{OrigName: true, EmitDefaults: true}
	for expected, in := range map[string]TestTypes{
		`{"api_only_string":"","numbers":[],"optional_string":null,"becomes_int":"UNKNOWN","nothingness":null,"uuid":null,"created_at":null,"duration":null,"type_with_id_id":0,"json_field":null,"nullable_uuid":null,"time_only":null,"bigint":null,"several_values":[]}`:                                   {},
		`{"api_only_string":"Something","numbers":[],"optional_string":null,"becomes_int":"UNKNOWN","nothingness":null,"uuid":null,"created_at":null,"duration":null,"type_with_id_id":0,"json_field":null,"nullable_uuid":null,"time_only":null,"bigint":null,"several_values":[]}`:                          {ApiOnlyString: "Something"},
		`{"api_only_string":"","numbers":[0,1,2,3],"optional_string":null,"becomes_int":"UNKNOWN","nothingness":null,"uuid":null,"created_at":null,"duration":null,"type_with_id_id":0,"json_field":null,"nullable_uuid":null,"time_only":null,"bigint":null,"several_values":[]}`:                            {Numbers: []int32{0, 1, 2, 3}},
		`{"api_only_string":"","numbers":[],"optional_string":"Not nothing","becomes_int":"UNKNOWN","nothingness":null,"uuid":null,"created_at":null,"duration":null,"type_with_id_id":0,"json_field":null,"nullable_uuid":null,"time_only":null,"bigint":null,"several_values":[]}`:                          {OptionalString: &wrappers.StringValue{Value: "Not nothing"}},
		`{"api_only_string":"","numbers":[],"optional_string":null,"becomes_int":"GOOD","nothingness":null,"uuid":null,"created_at":null,"duration":null,"type_with_id_id":0,"json_field":null,"nullable_uuid":null,"time_only":null,"bigint":null,"several_values":[]}`:                                      {BecomesInt: TestTypes_GOOD},
		`{"api_only_string":"","numbers":[],"optional_string":null,"becomes_int":"UNKNOWN","nothingness":null,"uuid":"6ba7b810-9dad-11d1-80b4-00c04fd430c8","created_at":null,"duration":null,"type_with_id_id":0,"json_field":null,"nullable_uuid":null,"time_only":null,"bigint":null,"several_values":[]}`: {Uuid: &types.UUID{Value: "6ba7b810-9dad-11d1-80b4-00c04fd430c8"}},
		`{"api_only_string":"","numbers":[],"optional_string":null,"becomes_int":"UNKNOWN","nothingness":null,"uuid":null,"created_at":"2009-11-17T20:34:58.651387237Z","duration":null,"type_with_id_id":0,"json_field":null,"nullable_uuid":null,"time_only":null,"bigint":null,"several_values":[]}`:       {CreatedAt: timestamppb.New(time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC))},
		`{"api_only_string":"","numbers":[],"optional_string":null,"becomes_int":"UNKNOWN","nothingness":null,"uuid":null,"created_at":null,"duration":"3600s","type_with_id_id":0,"json_field":null,"nullable_uuid":null,"time_only":null,"bigint":null,"several_values":[]}`:                                {Duration: durationpb.New(time.Hour)},
		`{"api_only_string":"","numbers":[],"optional_string":null,"becomes_int":"UNKNOWN","nothingness":null,"uuid":null,"created_at":null,"duration":null,"type_with_id_id":2,"json_field":null,"nullable_uuid":null,"time_only":null,"bigint":null,"several_values":[]}`:                                   {TypeWithIdId: 2},
		`{"api_only_string":"","numbers":[],"optional_string":null,"becomes_int":"UNKNOWN","nothingness":null,"uuid":null,"created_at":null,"duration":null,"type_with_id_id":0,"json_field":{"text":[]},"nullable_uuid":null,"time_only":null,"bigint":null,"several_values":[]}`:                            {JsonField: &types.JSONValue{Value: `{"text":[]}`}},
		`{"api_only_string":"","numbers":[],"optional_string":null,"becomes_int":"UNKNOWN","nothingness":null,"uuid":null,"created_at":null,"duration":null,"type_with_id_id":0,"json_field":null,"nullable_uuid":null,"time_only":"01:59:18","bigint":null,"several_values":[]}`:                             {TimeOnly: &types.TimeOnly{Value: 7158}},
		`{"api_only_string":"","numbers":[],"optional_string":null,"becomes_int":"UNKNOWN","nothingness":null,"uuid":null,"created_at":null,"duration":null,"type_with_id_id":0,"json_field":null,"nullable_uuid":null,"time_only":null,"bigint":"7158","several_values":[]}`:                                 {Bigint: &types.BigInt{Value: "7158"}},
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
	// Will marshal with snake_case names, but not default values
	marshaller := &jsonpb.Marshaler{OrigName: true}
	for expected, in := range map[string]TestTypes{
		`{}`:                                {},
		`{"api_only_string":"Something"}`:   {ApiOnlyString: "Something"},
		`{"numbers":[0,1,2,3]}`:             {Numbers: []int32{0, 1, 2, 3}},
		`{"optional_string":"Not nothing"}`: {OptionalString: &wrappers.StringValue{Value: "Not nothing"}},
		`{"becomes_int":"GOOD"}`:            {BecomesInt: 1},
		`{"uuid":"6ba7b810-9dad-11d1-80b4-00c04fd430c8"}`: {Uuid: &types.UUID{Value: "6ba7b810-9dad-11d1-80b4-00c04fd430c8"}},
		`{"created_at":"2009-11-17T20:34:58.651387237Z"}`: {CreatedAt: timestamppb.New(time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC))},
		`{"type_with_id_id":2}`:                           {TypeWithIdId: 2},
		`{"json_field":{"text":[]}}`:                      {JsonField: &types.JSONValue{Value: `{"text":[]}`}},
		`{"time_only":"00:00:00"}`:                        {TimeOnly: &types.TimeOnly{Value: 0}},
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
