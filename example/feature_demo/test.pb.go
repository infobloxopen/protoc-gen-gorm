// Code generated by protoc-gen-go. DO NOT EDIT.
// source: example/feature_demo/test.proto

/*
Package example is a generated protocol buffer package.

It is generated from these files:
	example/feature_demo/test.proto
	example/feature_demo/test2.proto

It has these top-level messages:
	TestTypes
	TypeWithID
	MultiaccountTypeWithID
	MultiaccountTypeWithoutID
	APIOnlyType
	TypeBecomesEmpty
	IntPoint
	CreateIntPointRequest
	CreateIntPointResponse
	ReadIntPointRequest
	ReadIntPointResponse
	UpdateIntPointRequest
	UpdateIntPointResponse
	DeleteIntPointRequest
	ListIntPointResponse
	Something
*/
package example

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/infobloxopen/protoc-gen-gorm/options"
import gormable_types "github.com/infobloxopen/protoc-gen-gorm/types"
import google_protobuf1 "github.com/golang/protobuf/ptypes/wrappers"
import google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
import google_protobuf3 "github.com/golang/protobuf/ptypes/timestamp"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

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

// test_types is a message that serves as an example
type TestTypes struct {
	ApiOnlyString  string                        `protobuf:"bytes,1,opt,name=api_only_string,json=apiOnlyString" json:"api_only_string,omitempty"`
	Numbers        []int32                       `protobuf:"varint,2,rep,packed,name=numbers" json:"numbers,omitempty"`
	OptionalString *google_protobuf1.StringValue `protobuf:"bytes,3,opt,name=optional_string,json=optionalString" json:"optional_string,omitempty"`
	BecomesInt     TestTypesStatus               `protobuf:"varint,4,opt,name=becomes_int,json=becomesInt,enum=example.TestTypesStatus" json:"becomes_int,omitempty"`
	Nothingness    *google_protobuf2.Empty       `protobuf:"bytes,5,opt,name=nothingness" json:"nothingness,omitempty"`
	Uuid           *gormable_types.UUIDValue     `protobuf:"bytes,6,opt,name=uuid" json:"uuid,omitempty"`
	CreatedAt      *google_protobuf3.Timestamp   `protobuf:"bytes,7,opt,name=created_at,json=createdAt" json:"created_at,omitempty"`
	TypeWithIdId   uint32                        `protobuf:"varint,8,opt,name=type_with_id_id,json=typeWithIdId" json:"type_with_id_id,omitempty"`
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

func (m *TestTypes) GetUuid() *gormable_types.UUIDValue {
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

type TypeWithID struct {
	Ip            string       `protobuf:"bytes,1,opt,name=ip" json:"ip,omitempty"`
	Things        []*TestTypes `protobuf:"bytes,3,rep,name=things" json:"things,omitempty"`
	ANestedObject *TestTypes   `protobuf:"bytes,4,opt,name=a_nested_object,json=aNestedObject" json:"a_nested_object,omitempty"`
	Id            uint32       `protobuf:"varint,5,opt,name=id" json:"id,omitempty"`
}

func (m *TypeWithID) Reset()                    { *m = TypeWithID{} }
func (m *TypeWithID) String() string            { return proto.CompactTextString(m) }
func (*TypeWithID) ProtoMessage()               {}
func (*TypeWithID) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

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

func (m *TypeWithID) GetId() uint32 {
	if m != nil {
		return m.Id
	}
	return 0
}

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

type TypeBecomesEmpty struct {
	AThing *APIOnlyType `protobuf:"bytes,1,opt,name=a_thing,json=aThing" json:"a_thing,omitempty"`
}

func (m *TypeBecomesEmpty) Reset()                    { *m = TypeBecomesEmpty{} }
func (m *TypeBecomesEmpty) String() string            { return proto.CompactTextString(m) }
func (*TypeBecomesEmpty) ProtoMessage()               {}
func (*TypeBecomesEmpty) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *TypeBecomesEmpty) GetAThing() *APIOnlyType {
	if m != nil {
		return m.AThing
	}
	return nil
}

func init() {
	proto.RegisterType((*TestTypes)(nil), "example.TestTypes")
	proto.RegisterType((*TypeWithID)(nil), "example.TypeWithID")
	proto.RegisterType((*MultiaccountTypeWithID)(nil), "example.MultiaccountTypeWithID")
	proto.RegisterType((*MultiaccountTypeWithoutID)(nil), "example.MultiaccountTypeWithoutID")
	proto.RegisterType((*APIOnlyType)(nil), "example.APIOnlyType")
	proto.RegisterType((*TypeBecomesEmpty)(nil), "example.TypeBecomesEmpty")
	proto.RegisterEnum("example.TestTypesStatus", TestTypesStatus_name, TestTypesStatus_value)
}

func init() { proto.RegisterFile("example/feature_demo/test.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 730 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x94, 0xdf, 0x8e, 0x9b, 0x46,
	0x14, 0xc6, 0x0b, 0xf6, 0xda, 0xec, 0xa1, 0xbb, 0xb6, 0xa6, 0x51, 0x04, 0xee, 0x9f, 0xb8, 0xa8,
	0x91, 0x9c, 0x4a, 0xc6, 0x92, 0x73, 0xd3, 0xd0, 0xab, 0xb5, 0x9c, 0x54, 0x56, 0x55, 0xbb, 0xa5,
	0xde, 0x46, 0xea, 0x0d, 0x1a, 0x60, 0x8c, 0xa7, 0x85, 0x19, 0xc4, 0x0c, 0x4a, 0xfc, 0x68, 0xf6,
	0x7b, 0xf4, 0x05, 0xfa, 0x24, 0xd5, 0x0c, 0x78, 0xb5, 0xda, 0xdd, 0x4a, 0xb9, 0x41, 0x9a, 0x39,
	0xbf, 0xf3, 0x9d, 0x73, 0xbe, 0x03, 0xc0, 0x0b, 0xf2, 0x11, 0x17, 0x65, 0x4e, 0x66, 0x3b, 0x82,
	0x65, 0x5d, 0x91, 0x28, 0x25, 0x05, 0x9f, 0x49, 0x22, 0xa4, 0x5f, 0x56, 0x5c, 0x72, 0xd4, 0x6f,
	0x81, 0x51, 0x90, 0x51, 0xb9, 0xaf, 0x63, 0x3f, 0xe1, 0xc5, 0x8c, 0xb2, 0x1d, 0x8f, 0x73, 0xfe,
	0x91, 0x97, 0x84, 0xcd, 0x34, 0x97, 0x4c, 0x33, 0xc2, 0xa6, 0x19, 0xaf, 0x8a, 0x19, 0x2f, 0x25,
	0xe5, 0x4c, 0xcc, 0xd4, 0xa1, 0x11, 0x19, 0xbd, 0xf9, 0xd4, 0x5c, 0x79, 0x28, 0x89, 0x68, 0x9e,
	0x6d, 0xea, 0x37, 0x19, 0xe7, 0x59, 0x4e, 0x1a, 0x32, 0xae, 0x77, 0xb3, 0x0f, 0x15, 0x2e, 0x4b,
	0x52, 0x9d, 0xe3, 0x5f, 0x3e, 0x8c, 0x93, 0xa2, 0x94, 0x87, 0x36, 0xf8, 0xe2, 0x61, 0x50, 0xd2,
	0x82, 0x08, 0x89, 0x8b, 0xb2, 0x01, 0xbc, 0x7f, 0x3b, 0x70, 0xb9, 0x25, 0x42, 0x6e, 0x55, 0x45,
	0xe4, 0xc3, 0x00, 0x97, 0x34, 0xe2, 0x2c, 0x3f, 0x44, 0x42, 0x56, 0x94, 0x65, 0x8e, 0x31, 0x36,
	0x26, 0x97, 0x8b, 0xde, 0xe9, 0xe8, 0x9a, 0x43, 0x23, 0xbc, 0xc2, 0x25, 0xdd, 0xb0, 0xfc, 0xf0,
	0xbb, 0x0e, 0x22, 0x07, 0xfa, 0xac, 0x2e, 0x62, 0x52, 0x09, 0xc7, 0x1c, 0x77, 0x26, 0x17, 0xe1,
	0xf9, 0x88, 0xde, 0xc2, 0xa0, 0xb1, 0x01, 0xe7, 0x67, 0xa5, 0xce, 0xd8, 0x98, 0xd8, 0xf3, 0xaf,
	0xfc, 0xa6, 0x25, 0xff, 0xdc, 0x92, 0xdf, 0x68, 0xfd, 0x81, 0xf3, 0x9a, 0x84, 0xd7, 0xe7, 0xa4,
	0xb6, 0x40, 0x00, 0x76, 0x4c, 0x12, 0x5e, 0x10, 0x11, 0x51, 0x26, 0x9d, 0xee, 0xd8, 0x98, 0x5c,
	0xcf, 0x5d, 0xbf, 0x5d, 0x89, 0x7f, 0xd7, 0xb9, 0x2f, 0x24, 0x96, 0xb5, 0x08, 0xa1, 0xa5, 0x57,
	0x4c, 0xa2, 0x1f, 0xc0, 0x66, 0x5c, 0xee, 0x29, 0xcb, 0x18, 0x11, 0xc2, 0xb9, 0xd0, 0xe5, 0x9f,
	0x3f, 0x2a, 0xff, 0x56, 0xd9, 0x15, 0xde, 0x47, 0xd1, 0x14, 0xba, 0x75, 0x4d, 0x53, 0xa7, 0xa7,
	0x53, 0x5c, 0x5f, 0x6d, 0x06, 0xc7, 0x39, 0x89, 0x9a, 0xbd, 0xdc, 0xde, 0xae, 0x96, 0x4d, 0xbb,
	0x1a, 0x43, 0x6f, 0x00, 0x92, 0x8a, 0x60, 0x49, 0xd2, 0x08, 0x4b, 0xa7, 0xaf, 0x93, 0x46, 0x8f,
	0xea, 0x6c, 0xcf, 0xce, 0x87, 0x97, 0x2d, 0x7d, 0x23, 0xd1, 0x4b, 0x18, 0x28, 0xcd, 0xe8, 0x03,
	0x95, 0xfb, 0x88, 0xa6, 0x11, 0x4d, 0x1d, 0x6b, 0x6c, 0x4c, 0xae, 0xc2, 0xcf, 0xd5, 0xf5, 0x7b,
	0x2a, 0xf7, 0xab, 0x74, 0x95, 0x7a, 0x13, 0xe8, 0x35, 0x03, 0x22, 0x1b, 0xfa, 0xb7, 0xeb, 0x9f,
	0xd7, 0x9b, 0xf7, 0xeb, 0xe1, 0x67, 0xc8, 0x82, 0xee, 0x4f, 0x9b, 0xcd, 0x72, 0x68, 0xa0, 0x3e,
	0x74, 0x16, 0x37, 0xcb, 0xa1, 0x19, 0x7c, 0x71, 0x3a, 0xba, 0x03, 0xcb, 0x18, 0xd9, 0xa2, 0xe0,
	0x55, 0x86, 0x45, 0xcc, 0xab, 0xd4, 0xfb, 0xc7, 0x00, 0xd8, 0x9e, 0xf5, 0x96, 0xe8, 0x3b, 0x30,
	0x69, 0xd9, 0x2e, 0xf6, 0xd9, 0xe9, 0xe8, 0x0e, 0xe1, 0x5a, 0x4d, 0x18, 0x78, 0xb4, 0x8c, 0x70,
	0x9a, 0x56, 0x5e, 0x68, 0xd2, 0x12, 0x7d, 0x0f, 0x3d, 0xed, 0x88, 0x70, 0x3a, 0xe3, 0xce, 0xc4,
	0x9e, 0xa3, 0xc7, 0xae, 0x87, 0x2d, 0x81, 0x02, 0x18, 0xe0, 0x88, 0x11, 0xa1, 0x2c, 0xe0, 0xf1,
	0x5f, 0x24, 0x69, 0x56, 0xf5, 0x74, 0xd2, 0x15, 0x5e, 0x6b, 0x72, 0xa3, 0x41, 0x74, 0x0d, 0x26,
	0x4d, 0xf5, 0x76, 0xae, 0x42, 0x93, 0xa6, 0xc1, 0xab, 0xd3, 0xd1, 0x7d, 0x69, 0x19, 0xe8, 0x5b,
	0xb8, 0xa0, 0x4c, 0xbe, 0x9e, 0x23, 0x6d, 0xf2, 0x08, 0x35, 0x2d, 0x96, 0x15, 0x2d, 0x70, 0x75,
	0x88, 0xfe, 0x26, 0x07, 0xcf, 0xfb, 0x0d, 0x9e, 0xff, 0x52, 0xe7, 0x92, 0xe2, 0x24, 0xe1, 0x35,
	0x93, 0xf7, 0x46, 0x6c, 0x44, 0xd5, 0x88, 0x5d, 0x25, 0x8a, 0xbe, 0x06, 0x10, 0xbc, 0x20, 0xd1,
	0x8e, 0x92, 0x3c, 0x75, 0x4c, 0x35, 0x7a, 0x78, 0xa9, 0x6e, 0xde, 0xa9, 0x8b, 0xc0, 0x3a, 0x1d,
	0xdd, 0xae, 0x65, 0x8c, 0x0d, 0x6f, 0x09, 0xee, 0x53, 0x92, 0xbc, 0x96, 0xab, 0xe5, 0x03, 0x15,
	0xe3, 0xff, 0x55, 0x5e, 0x81, 0x7d, 0xf3, 0xeb, 0x4a, 0x7d, 0x28, 0x4a, 0x00, 0x8d, 0xc0, 0x4a,
	0x38, 0x93, 0x84, 0x49, 0xd1, 0x66, 0xdd, 0x9d, 0xbd, 0x15, 0x0c, 0x15, 0xb3, 0x68, 0xde, 0x5b,
	0xfd, 0x32, 0xa2, 0x29, 0xf4, 0x71, 0xa4, 0xad, 0xd5, 0xb8, 0x3d, 0x7f, 0x76, 0x67, 0xe3, 0x3d,
	0xd9, 0xb0, 0x87, 0xb7, 0x8a, 0x09, 0xf4, 0xc7, 0x69, 0x19, 0x8b, 0x77, 0x7f, 0x2e, 0x3f, 0xf5,
	0x37, 0xf3, 0xd4, 0x4f, 0xef, 0xc7, 0xf6, 0x32, 0xee, 0x69, 0xfa, 0xf5, 0x7f, 0x01, 0x00, 0x00,
	0xff, 0xff, 0x82, 0x12, 0x03, 0xb0, 0x1b, 0x05, 0x00, 0x00,
}
