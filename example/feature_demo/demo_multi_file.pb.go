// Code generated by protoc-gen-go. DO NOT EDIT.
// source: example/feature_demo/demo_multi_file.proto

/*
Package example is a generated protocol buffer package.

It is generated from these files:
	example/feature_demo/demo_multi_file.proto
	example/feature_demo/demo_types.proto
	example/feature_demo/demo_service.proto

It has these top-level messages:
	ExternalChild
	TestTypes
	TypeWithID
	MultiaccountTypeWithID
	MultiaccountTypeWithoutID
	APIOnlyType
	PrimaryUUIDType
	PrimaryStringType
	PrimaryIncluded
	IntPoint
	CreateIntPointRequest
	CreateIntPointResponse
	ReadIntPointRequest
	ReadIntPointResponse
	UpdateIntPointRequest
	UpdateIntPointResponse
	DeleteIntPointRequest
	DeleteIntPointResponse
	ListIntPointResponse
	Something
	ListIntPointRequest
*/
package example

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/infobloxopen/protoc-gen-gorm/options"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type ExternalChild struct {
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
}

func (m *ExternalChild) Reset()                    { *m = ExternalChild{} }
func (m *ExternalChild) String() string            { return proto.CompactTextString(m) }
func (*ExternalChild) ProtoMessage()               {}
func (*ExternalChild) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *ExternalChild) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func init() {
	proto.RegisterType((*ExternalChild)(nil), "example.ExternalChild")
}

func init() { proto.RegisterFile("example/feature_demo/demo_multi_file.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 179 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xd2, 0x4a, 0xad, 0x48, 0xcc,
	0x2d, 0xc8, 0x49, 0xd5, 0x4f, 0x4b, 0x4d, 0x2c, 0x29, 0x2d, 0x4a, 0x8d, 0x4f, 0x49, 0xcd, 0xcd,
	0xd7, 0x07, 0x11, 0xf1, 0xb9, 0xa5, 0x39, 0x25, 0x99, 0xf1, 0x69, 0x99, 0x39, 0xa9, 0x7a, 0x05,
	0x45, 0xf9, 0x25, 0xf9, 0x42, 0xec, 0x50, 0xb5, 0x52, 0x56, 0xe9, 0x99, 0x25, 0x19, 0xa5, 0x49,
	0x7a, 0xc9, 0xf9, 0xb9, 0xfa, 0x99, 0x79, 0x69, 0xf9, 0x49, 0x39, 0xf9, 0x15, 0xf9, 0x05, 0xa9,
	0x79, 0xfa, 0x60, 0x75, 0xc9, 0xba, 0xe9, 0xa9, 0x79, 0xba, 0xe9, 0xf9, 0x45, 0xb9, 0xfa, 0xf9,
	0x05, 0x25, 0x99, 0xf9, 0x79, 0xc5, 0xfa, 0x20, 0x0e, 0xc4, 0x10, 0x25, 0x75, 0x2e, 0x5e, 0xd7,
	0x8a, 0x92, 0xd4, 0xa2, 0xbc, 0xc4, 0x1c, 0xe7, 0x8c, 0xcc, 0x9c, 0x14, 0x21, 0x3e, 0x2e, 0xa6,
	0xcc, 0x14, 0x09, 0x46, 0x05, 0x46, 0x0d, 0xce, 0x20, 0xa6, 0xcc, 0x14, 0x2b, 0xb6, 0x5d, 0x3b,
	0x25, 0x99, 0x38, 0x18, 0x9d, 0xdc, 0xa2, 0x5c, 0x88, 0xb5, 0x06, 0x9b, 0x1f, 0xac, 0xa1, 0x82,
	0x49, 0x6c, 0x60, 0xd5, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0x53, 0xb9, 0x8f, 0x17, 0xea,
	0x00, 0x00, 0x00,
}
