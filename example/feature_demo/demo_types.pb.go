// Code generated by protoc-gen-go.
// source: example/feature_demo/demo_types.proto
// DO NOT EDIT!

/*
Package example is a generated protocol buffer package.

It is generated from these files:
	example/feature_demo/demo_types.proto
	example/feature_demo/demo_service.proto

It has these top-level messages:
	TestTypes
	TypeWithID
	MultiaccountTypeWithID
	MultiaccountTypeWithoutID
	APIOnlyType
	IntPoint
	CreateIntPointRequest
	CreateIntPointResponse
	ReadIntPointRequest
	ReadIntPointResponse
	UpdateIntPointRequest
	UpdateIntPointResponse
	PatchIntPointRequest
	PatchIntPointResponse
	DeleteIntPointRequest
	DeleteIntPointResponse
	ListIntPointResponse
	Something
*/
package example

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/infobloxopen/protoc-gen-gorm/options"
import gorm_types "github.com/infobloxopen/protoc-gen-gorm/types"
import google_protobuf1 "github.com/golang/protobuf/ptypes/wrappers"
import google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
import google_protobuf3 "github.com/golang/protobuf/ptypes/timestamp"
import user "github.com/infobloxopen/protoc-gen-gorm/example/user"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// enums are mapped to the their underlying numeric value in the db.
// This is practical from an API perspective, but tougher for debugging.
// Strings with validation constraints can be used instead if desired
type TestTypesStatus int32

const (
	TestTypes_UNKNOWN TestTypesStatus = 0
	TestTypes_GOOD    TestTypesStatus = 1
	TestTypes_BAD     TestTypesStatus = 2
)

var TestTypesStatus_name = map[int32]string{
	0: "UNKNOWN",
	1: "GOOD",
	2: "BAD",
}
var TestTypesStatus_value = map[string]int32{
	"UNKNOWN": 0,
	"GOOD":    1,
	"BAD":     2,
}

func (x TestTypesStatus) String() string {
	return proto.EnumName(TestTypesStatus_name, int32(x))
}
func (TestTypesStatus) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

// test_types is a message that includes a representative sample of the
// available types
type TestTypes struct {
	// the (gorm.field).drop option allows for setting a field to be API only
	ApiOnlyString string `protobuf:"bytes,1,opt,name=api_only_string,json=apiOnlyString" json:"api_only_string,omitempty"`
	// repeated raw types are currently unsupported, so this field will be dropped
	// at the ORM level
	Numbers []int32 `protobuf:"varint,2,rep,packed,name=numbers" json:"numbers,omitempty"`
	// a StringValue represents a Nullable string
	OptionalString *google_protobuf1.StringValue `protobuf:"bytes,3,opt,name=optional_string,json=optionalString" json:"optional_string,omitempty"`
	BecomesInt     TestTypesStatus               `protobuf:"varint,4,opt,name=becomes_int,json=becomesInt,enum=example.TestTypesStatus" json:"becomes_int,omitempty"`
	// The Empty type serves no purpose outside of rpc calls and is dropped
	// automatically from objects
	Nothingness *google_protobuf2.Empty `protobuf:"bytes,5,opt,name=nothingness" json:"nothingness,omitempty"`
	// The UUID custom type should act like a StringValue at the API level, but is
	// automatically converted to and from a uuid.UUID (github.com/satori/go.uuid)
	Uuid *gorm_types.UUID `protobuf:"bytes,6,opt,name=uuid" json:"uuid,omitempty"`
	// Timestamps convert to golang's time.Time type, and created_at and
	// updated_at values are automatically filled by GORM
	CreatedAt *google_protobuf3.Timestamp `protobuf:"bytes,7,opt,name=created_at,json=createdAt" json:"created_at,omitempty"`
	// This represents a foreign key to the 'type_with_id' type for associations
	// This could be hidden from the API (or soon autogenerated).
	TypeWithIdId uint32 `protobuf:"varint,8,opt,name=type_with_id_id,json=typeWithIdId" json:"type_with_id_id,omitempty"`
	// This is an arbitrary JSON string that is marshalled and unmarshalled
	// specially in grpc-gateway as a JSON object
	JsonField *gorm_types.JSONValue `protobuf:"bytes,9,opt,name=json_field,json=jsonField" json:"json_field,omitempty"`
	// The UUIDValue custom type should act like a StringValue at the API level, but is
	// automatically converted to and from a *uuid.UUID (github.com/satori/go.uuid)
	NullableUuid *gorm_types.UUIDValue `protobuf:"bytes,10,opt,name=nullable_uuid,json=nullableUuid" json:"nullable_uuid,omitempty"`
}

func (m *TestTypes) Reset()                    { *m = TestTypes{} }
func (m *TestTypes) String() string            { return proto.CompactTextString(m) }
func (*TestTypes) ProtoMessage()               {}
func (*TestTypes) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *TestTypes) GetApiOnlyString() string {
	if m != nil {
		return m.ApiOnlyString
	}
	return ""
}

func (m *TestTypes) GetNumbers() []int32 {
	if m != nil {
		return m.Numbers
	}
	return nil
}

func (m *TestTypes) GetOptionalString() *google_protobuf1.StringValue {
	if m != nil {
		return m.OptionalString
	}
	return nil
}

func (m *TestTypes) GetBecomesInt() TestTypesStatus {
	if m != nil {
		return m.BecomesInt
	}
	return TestTypes_UNKNOWN
}

func (m *TestTypes) GetNothingness() *google_protobuf2.Empty {
	if m != nil {
		return m.Nothingness
	}
	return nil
}

func (m *TestTypes) GetUuid() *gorm_types.UUID {
	if m != nil {
		return m.Uuid
	}
	return nil
}

func (m *TestTypes) GetCreatedAt() *google_protobuf3.Timestamp {
	if m != nil {
		return m.CreatedAt
	}
	return nil
}

func (m *TestTypes) GetTypeWithIdId() uint32 {
	if m != nil {
		return m.TypeWithIdId
	}
	return 0
}

func (m *TestTypes) GetJsonField() *gorm_types.JSONValue {
	if m != nil {
		return m.JsonField
	}
	return nil
}

func (m *TestTypes) GetNullableUuid() *gorm_types.UUIDValue {
	if m != nil {
		return m.NullableUuid
	}
	return nil
}

// TypeWithID demonstrates some basic assocation behavior
type TypeWithID struct {
	// any field named 'id' is assumed by gorm to be the primary key for the
	// object.
	Id uint32 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	// The field option also allows arbitrary tag setting, such as informing
	// gorm of a primary key, different column names or different types in the db
	Ip string `protobuf:"bytes,2,opt,name=ip" json:"ip,omitempty"`
	// A default has-many relationship, will error on generation if no FK field,
	// convention {typename}_id, is present. These FK fields will be automatically
	// populated on create and update.
	Things []*TestTypes `protobuf:"bytes,3,rep,name=things" json:"things,omitempty"`
	// A default has-one relationship, will error as above
	ANestedObject *TestTypes `protobuf:"bytes,4,opt,name=a_nested_object,json=aNestedObject" json:"a_nested_object,omitempty"`
	// An in-package and cross-package imported type (in-package can use any
	// association type, cross-package is limited to belongs_to and many_to_many)
	Point   *IntPoint             `protobuf:"bytes,5,opt,name=point" json:"point,omitempty"`
	User    *user.User            `protobuf:"bytes,6,opt,name=user" json:"user,omitempty"`
	Address *gorm_types.InetValue `protobuf:"bytes,7,opt,name=address" json:"address,omitempty"`
}

func (m *TypeWithID) Reset()                    { *m = TypeWithID{} }
func (m *TypeWithID) String() string            { return proto.CompactTextString(m) }
func (*TypeWithID) ProtoMessage()               {}
func (*TypeWithID) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *TypeWithID) GetId() uint32 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *TypeWithID) GetIp() string {
	if m != nil {
		return m.Ip
	}
	return ""
}

func (m *TypeWithID) GetThings() []*TestTypes {
	if m != nil {
		return m.Things
	}
	return nil
}

func (m *TypeWithID) GetANestedObject() *TestTypes {
	if m != nil {
		return m.ANestedObject
	}
	return nil
}

func (m *TypeWithID) GetPoint() *IntPoint {
	if m != nil {
		return m.Point
	}
	return nil
}

func (m *TypeWithID) GetUser() *user.User {
	if m != nil {
		return m.User
	}
	return nil
}

func (m *TypeWithID) GetAddress() *gorm_types.InetValue {
	if m != nil {
		return m.Address
	}
	return nil
}

// MultiaccountTypeWithID demonstrates the generated multi-account support
type MultiaccountTypeWithID struct {
	Id        uint64 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	SomeField string `protobuf:"bytes,2,opt,name=some_field,json=someField" json:"some_field,omitempty"`
}

func (m *MultiaccountTypeWithID) Reset()                    { *m = MultiaccountTypeWithID{} }
func (m *MultiaccountTypeWithID) String() string            { return proto.CompactTextString(m) }
func (*MultiaccountTypeWithID) ProtoMessage()               {}
func (*MultiaccountTypeWithID) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *MultiaccountTypeWithID) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *MultiaccountTypeWithID) GetSomeField() string {
	if m != nil {
		return m.SomeField
	}
	return ""
}

type MultiaccountTypeWithoutID struct {
	SomeField string `protobuf:"bytes,1,opt,name=some_field,json=someField" json:"some_field,omitempty"`
}

func (m *MultiaccountTypeWithoutID) Reset()                    { *m = MultiaccountTypeWithoutID{} }
func (m *MultiaccountTypeWithoutID) String() string            { return proto.CompactTextString(m) }
func (*MultiaccountTypeWithoutID) ProtoMessage()               {}
func (*MultiaccountTypeWithoutID) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *MultiaccountTypeWithoutID) GetSomeField() string {
	if m != nil {
		return m.SomeField
	}
	return ""
}

type APIOnlyType struct {
	// here the ormable flag is not used, so nothing will be generated for this
	// object at the ORM level, and when this type is used as a field or
	// repeated field in another message that field will be dropped in the Orm
	// model, and would have to be set by hook
	Contents string `protobuf:"bytes,1,opt,name=contents" json:"contents,omitempty"`
}

func (m *APIOnlyType) Reset()                    { *m = APIOnlyType{} }
func (m *APIOnlyType) String() string            { return proto.CompactTextString(m) }
func (*APIOnlyType) ProtoMessage()               {}
func (*APIOnlyType) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *APIOnlyType) GetContents() string {
	if m != nil {
		return m.Contents
	}
	return ""
}

func init() {
	proto.RegisterType((*TestTypes)(nil), "example.TestTypes")
	proto.RegisterType((*TypeWithID)(nil), "example.TypeWithID")
	proto.RegisterType((*MultiaccountTypeWithID)(nil), "example.MultiaccountTypeWithID")
	proto.RegisterType((*MultiaccountTypeWithoutID)(nil), "example.MultiaccountTypeWithoutID")
	proto.RegisterType((*APIOnlyType)(nil), "example.APIOnlyType")
	proto.RegisterEnum("example.TestTypesStatus", TestTypesStatus_name, TestTypesStatus_value)
}

func init() { proto.RegisterFile("example/feature_demo/demo_types.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 842 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x94, 0xdf, 0x8e, 0xdb, 0x44,
	0x14, 0xc6, 0x6b, 0xe7, 0xff, 0x71, 0xb3, 0x49, 0x47, 0xa2, 0x38, 0x29, 0xb0, 0x51, 0x44, 0x69,
	0x40, 0xaa, 0x2d, 0xa5, 0x5c, 0xd0, 0x70, 0x81, 0x76, 0x95, 0x16, 0x05, 0x44, 0x52, 0xdc, 0x0d,
	0x95, 0xb8, 0xb1, 0xc6, 0xf6, 0x24, 0x3b, 0x95, 0x3d, 0x63, 0x3c, 0x63, 0xda, 0xbc, 0x0c, 0xef,
	0x91, 0x3c, 0x05, 0x8f, 0x84, 0x66, 0x6c, 0xaf, 0xc2, 0x66, 0x2b, 0xb5, 0x37, 0x89, 0xec, 0xf3,
	0x3b, 0xdf, 0xf1, 0x39, 0xe7, 0x9b, 0x81, 0xc7, 0xe4, 0x3d, 0x4e, 0xd2, 0x98, 0xb8, 0x1b, 0x82,
	0x65, 0x9e, 0x11, 0x3f, 0x22, 0x09, 0x77, 0xd5, 0x8f, 0x2f, 0x77, 0x29, 0x11, 0x4e, 0x9a, 0x71,
	0xc9, 0x51, 0xab, 0xc4, 0x86, 0xb3, 0x2d, 0x95, 0xd7, 0x79, 0xe0, 0x84, 0x3c, 0x71, 0x29, 0xdb,
	0xf0, 0x20, 0xe6, 0xef, 0x79, 0x4a, 0x98, 0xab, 0xb9, 0xf0, 0xe9, 0x96, 0xb0, 0xa7, 0x5b, 0x9e,
	0x25, 0x2e, 0x4f, 0x25, 0xe5, 0x4c, 0xb8, 0xea, 0xa1, 0x10, 0x19, 0x3e, 0xff, 0xd8, 0x5c, 0x5d,
	0xd9, 0x3d, 0xaa, 0x3f, 0xfc, 0x6a, 0xcb, 0xf9, 0x36, 0x26, 0x05, 0x19, 0xe4, 0x1b, 0xf7, 0x5d,
	0x86, 0xd3, 0x94, 0x64, 0x55, 0xfc, 0xd1, 0xed, 0x38, 0x49, 0x52, 0xb9, 0x2b, 0x83, 0xe7, 0xb7,
	0x83, 0x92, 0x26, 0x44, 0x48, 0x9c, 0xa4, 0x25, 0xf0, 0xe4, 0xc3, 0x43, 0x10, 0x24, 0xfb, 0x9b,
	0x86, 0xa4, 0x04, 0x7f, 0xfa, 0xd8, 0x0e, 0x2a, 0xc1, 0x5c, 0x90, 0x4c, 0xff, 0x14, 0x02, 0xe3,
	0x7f, 0x1a, 0xd0, 0xb9, 0x22, 0x42, 0x5e, 0xa9, 0xde, 0x90, 0x03, 0x3d, 0x9c, 0x52, 0x9f, 0xb3,
	0x78, 0xe7, 0x0b, 0x99, 0x51, 0xb6, 0xb5, 0x8d, 0x91, 0x31, 0xe9, 0x5c, 0x36, 0x0f, 0xfb, 0x81,
	0xd9, 0x37, 0xbc, 0x2e, 0x4e, 0xe9, 0x8a, 0xc5, 0xbb, 0xd7, 0x3a, 0x88, 0x6c, 0x68, 0xb1, 0x3c,
	0x09, 0x48, 0x26, 0x6c, 0x73, 0x54, 0x9b, 0x34, 0xbc, 0xea, 0x11, 0xbd, 0x80, 0x5e, 0x31, 0x70,
	0x1c, 0x57, 0x4a, 0xb5, 0x91, 0x31, 0xb1, 0xa6, 0x5f, 0x38, 0x45, 0xf3, 0x4e, 0xd5, 0xbc, 0x53,
	0x68, 0xfd, 0x81, 0xe3, 0x9c, 0x78, 0x67, 0x55, 0x52, 0x59, 0x60, 0x06, 0x56, 0x40, 0x42, 0x9e,
	0x10, 0xe1, 0x53, 0x26, 0xed, 0xfa, 0xc8, 0x98, 0x9c, 0x4d, 0x07, 0x4e, 0xd9, 0x8d, 0x73, 0xf3,
	0xe5, 0x8e, 0x90, 0x58, 0xe6, 0xc2, 0x83, 0x92, 0x5e, 0x30, 0x89, 0x7e, 0x00, 0x8b, 0x71, 0x79,
	0x4d, 0xd9, 0x96, 0x11, 0x21, 0xec, 0x86, 0x2e, 0xff, 0xf0, 0xa4, 0xfc, 0x0b, 0xb5, 0x18, 0xef,
	0x18, 0x45, 0x5f, 0x43, 0x3d, 0xcf, 0x69, 0x64, 0x37, 0x75, 0x4a, 0xdf, 0xd1, 0x96, 0x29, 0xb6,
	0xbf, 0x5e, 0x2f, 0xe6, 0x9e, 0x8e, 0xa2, 0xe7, 0x00, 0x61, 0x46, 0xb0, 0x24, 0x91, 0x8f, 0xa5,
	0xdd, 0xd2, 0xec, 0xf0, 0x44, 0xfe, 0xaa, 0x5a, 0xad, 0xd7, 0x29, 0xe9, 0x0b, 0x89, 0x1e, 0x43,
	0x4f, 0xc9, 0xf9, 0xef, 0xa8, 0xbc, 0xf6, 0x69, 0xe4, 0xd3, 0xc8, 0x6e, 0x8f, 0x8c, 0x49, 0xd7,
	0xbb, 0xaf, 0x5e, 0xbf, 0xa1, 0xf2, 0x7a, 0x11, 0x2d, 0x22, 0xf4, 0x3d, 0xc0, 0x5b, 0xc1, 0x99,
	0xbf, 0xa1, 0x24, 0x8e, 0xec, 0x8e, 0xae, 0xf0, 0xd9, 0xf1, 0xd7, 0xfc, 0xf2, 0x7a, 0xb5, 0x2c,
	0x06, 0xd7, 0x51, 0xe0, 0x4b, 0xc5, 0xa1, 0x19, 0x74, 0x59, 0x1e, 0xc7, 0x38, 0x88, 0x89, 0xaf,
	0xdb, 0x80, 0xd3, 0x44, 0xd5, 0x46, 0x91, 0x78, 0xbf, 0x62, 0xd7, 0x39, 0x8d, 0xc6, 0x13, 0x68,
	0x16, 0x93, 0x44, 0x16, 0xb4, 0xd6, 0xcb, 0x5f, 0x97, 0xab, 0x37, 0xcb, 0xfe, 0x3d, 0xd4, 0x86,
	0xfa, 0xcf, 0xab, 0xd5, 0xbc, 0x6f, 0xa0, 0x16, 0xd4, 0x2e, 0x2f, 0xe6, 0x7d, 0x73, 0xb6, 0x39,
	0xec, 0x07, 0x41, 0xdb, 0x40, 0x4f, 0xc0, 0x2a, 0x76, 0x75, 0x91, 0x65, 0x78, 0x87, 0x1a, 0x58,
	0xfd, 0x8d, 0x1f, 0x1c, 0xf9, 0x32, 0xa6, 0x81, 0x9b, 0xfe, 0x85, 0x26, 0xff, 0x07, 0x9b, 0x1a,
	0x9c, 0xde, 0x41, 0x0e, 0x2d, 0x91, 0xf0, 0x6c, 0x8b, 0x45, 0xc0, 0xb3, 0x68, 0xfc, 0xaf, 0x09,
	0x70, 0x55, 0x0d, 0x65, 0x8e, 0xce, 0xc0, 0xa4, 0x91, 0x36, 0x65, 0xd7, 0x33, 0x69, 0x84, 0xce,
	0xc1, 0xa4, 0xa9, 0x6d, 0x6a, 0x93, 0xf6, 0x0e, 0xfb, 0x81, 0x05, 0x1d, 0x68, 0xd1, 0xd4, 0xc7,
	0x51, 0x94, 0x79, 0x26, 0x4d, 0xd1, 0x77, 0xd0, 0xd4, 0x8b, 0x15, 0x76, 0x6d, 0x54, 0x9b, 0x58,
	0x53, 0x74, 0x6a, 0x1e, 0xaf, 0x24, 0xd0, 0x0c, 0x7a, 0xd8, 0x67, 0x44, 0xa8, 0x95, 0xf2, 0xe0,
	0x2d, 0x09, 0x0b, 0xc7, 0xdd, 0x9d, 0xd4, 0xc5, 0x4b, 0x4d, 0xae, 0x34, 0x88, 0x5c, 0x68, 0xa4,
	0x5c, 0x79, 0xb4, 0xf0, 0xd9, 0x83, 0x9b, 0x8c, 0x05, 0x93, 0xaf, 0x54, 0xa0, 0x38, 0x43, 0xe3,
	0x7b, 0x5e, 0xc1, 0xa1, 0x6f, 0xa0, 0xae, 0xce, 0x61, 0x69, 0x32, 0x70, 0xf4, 0xa1, 0x5c, 0x0b,
	0x92, 0xdd, 0x80, 0x3a, 0x8e, 0x5c, 0x68, 0xa9, 0x66, 0x94, 0x85, 0x5b, 0xa7, 0x8b, 0x5c, 0x30,
	0x22, 0x8b, 0x45, 0x56, 0xd4, 0xec, 0xfc, 0xb0, 0x1f, 0x3c, 0x6a, 0x1b, 0xe8, 0x73, 0x68, 0x50,
	0x26, 0x9f, 0x4d, 0x11, 0x08, 0x12, 0x66, 0x44, 0xaa, 0x13, 0x34, 0x34, 0x53, 0x63, 0xfc, 0x3b,
	0x3c, 0xfc, 0x2d, 0x8f, 0x25, 0xc5, 0x61, 0xc8, 0x73, 0x26, 0xef, 0x9c, 0x6e, 0x5d, 0x4f, 0xf7,
	0x4b, 0x00, 0xc1, 0x13, 0x52, 0x1a, 0x50, 0x4f, 0xd9, 0xeb, 0xa8, 0x37, 0xda, 0x69, 0xb3, 0xf6,
	0x61, 0x3f, 0xa8, 0xb7, 0x8d, 0x91, 0x31, 0x9e, 0xc3, 0xe0, 0x2e, 0x49, 0x9e, 0xcb, 0xc5, 0xfc,
	0x96, 0x8a, 0xf1, 0x61, 0x95, 0x6f, 0xc1, 0xba, 0x78, 0xb5, 0x50, 0xf7, 0x8b, 0x12, 0x40, 0x43,
	0x68, 0x87, 0x9c, 0x49, 0xc2, 0xa4, 0x28, 0xb3, 0x6e, 0x9e, 0x2f, 0x5f, 0xfe, 0x39, 0xff, 0xd4,
	0xab, 0xef, 0xf8, 0x2e, 0xfd, 0xb1, 0x7c, 0x19, 0x34, 0x35, 0xfd, 0xec, 0xbf, 0x00, 0x00, 0x00,
	0xff, 0xff, 0x2e, 0x5e, 0xd2, 0x54, 0x77, 0x06, 0x00, 0x00,
}
