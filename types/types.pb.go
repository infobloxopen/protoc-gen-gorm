// Code generated by protoc-gen-go. DO NOT EDIT.
// source: types/types.proto

/*
Package types is a generated protocol buffer package.

It is generated from these files:
	types/types.proto

It has these top-level messages:
	UUIDValue
	JSONValue
	UUID
	InetValue
*/
package types

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type UUIDValue struct {
	Value string `protobuf:"bytes,1,opt,name=value" json:"value,omitempty"`
}

func (m *UUIDValue) Reset()                    { *m = UUIDValue{} }
func (m *UUIDValue) String() string            { return proto.CompactTextString(m) }
func (*UUIDValue) ProtoMessage()               {}
func (*UUIDValue) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *UUIDValue) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type JSONValue struct {
	Value string `protobuf:"bytes,1,opt,name=value" json:"value,omitempty"`
}

func (m *JSONValue) Reset()                    { *m = JSONValue{} }
func (m *JSONValue) String() string            { return proto.CompactTextString(m) }
func (*JSONValue) ProtoMessage()               {}
func (*JSONValue) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *JSONValue) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type UUID struct {
	Value string `protobuf:"bytes,1,opt,name=value" json:"value,omitempty"`
}

func (m *UUID) Reset()                    { *m = UUID{} }
func (m *UUID) String() string            { return proto.CompactTextString(m) }
func (*UUID) ProtoMessage()               {}
func (*UUID) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *UUID) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type InetValue struct {
	Value string `protobuf:"bytes,1,opt,name=value" json:"value,omitempty"`
}

func (m *InetValue) Reset()                    { *m = InetValue{} }
func (m *InetValue) String() string            { return proto.CompactTextString(m) }
func (*InetValue) ProtoMessage()               {}
func (*InetValue) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *InetValue) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

func init() {
	proto.RegisterType((*UUIDValue)(nil), "gorm.types.UUIDValue")
	proto.RegisterType((*JSONValue)(nil), "gorm.types.JSONValue")
	proto.RegisterType((*UUID)(nil), "gorm.types.UUID")
	proto.RegisterType((*InetValue)(nil), "gorm.types.InetValue")
}

func init() { proto.RegisterFile("types/types.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 147 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2c, 0xa9, 0x2c, 0x48,
	0x2d, 0xd6, 0x07, 0x93, 0x7a, 0x05, 0x45, 0xf9, 0x25, 0xf9, 0x42, 0x5c, 0xe9, 0xf9, 0x45, 0xb9,
	0x7a, 0x60, 0x11, 0x25, 0x45, 0x2e, 0xce, 0xd0, 0x50, 0x4f, 0x97, 0xb0, 0xc4, 0x9c, 0xd2, 0x54,
	0x21, 0x11, 0x2e, 0xd6, 0x32, 0x10, 0x43, 0x82, 0x51, 0x81, 0x51, 0x83, 0x33, 0x08, 0xc2, 0x01,
	0x29, 0xf1, 0x0a, 0xf6, 0xf7, 0xc3, 0xa7, 0x44, 0x86, 0x8b, 0x05, 0x64, 0x0a, 0x6e, 0x03, 0x3c,
	0xf3, 0x52, 0x4b, 0xf0, 0x18, 0xe0, 0x64, 0x1a, 0x65, 0x9c, 0x9e, 0x59, 0x92, 0x51, 0x9a, 0xa4,
	0x97, 0x9c, 0x9f, 0xab, 0x9f, 0x99, 0x97, 0x96, 0x9f, 0x94, 0x93, 0x5f, 0x91, 0x5f, 0x90, 0x9a,
	0xa7, 0x0f, 0x76, 0x73, 0xb2, 0x6e, 0x7a, 0x6a, 0x9e, 0x2e, 0xc8, 0xdd, 0x10, 0x9f, 0x58, 0x83,
	0xc9, 0x24, 0x36, 0xb0, 0xa4, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0xe7, 0x8b, 0xcb, 0x94, 0xe5,
	0x00, 0x00, 0x00,
}
