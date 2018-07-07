// Code generated by protoc-gen-go.
// source: example/feature_demo/demo_service.proto
// DO NOT EDIT!

package example

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
import _ "github.com/infobloxopen/protoc-gen-gorm/options"
import google_protobuf4 "google.golang.org/genproto/protobuf/field_mask"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// IntPoint is a basic message type representing a single cartesian point
// that we want to store in a database
type IntPoint struct {
	Id uint32 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	X  int32  `protobuf:"varint,2,opt,name=x" json:"x,omitempty"`
	Y  int32  `protobuf:"varint,3,opt,name=y" json:"y,omitempty"`
}

func (m *IntPoint) Reset()                    { *m = IntPoint{} }
func (m *IntPoint) String() string            { return proto.CompactTextString(m) }
func (*IntPoint) ProtoMessage()               {}
func (*IntPoint) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

func (m *IntPoint) GetId() uint32 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *IntPoint) GetX() int32 {
	if m != nil {
		return m.X
	}
	return 0
}

func (m *IntPoint) GetY() int32 {
	if m != nil {
		return m.Y
	}
	return 0
}

type CreateIntPointRequest struct {
	// Convention dictates that this field be of the given type, and be
	// named 'payload' in order to autogenerate the handler
	Payload *IntPoint `protobuf:"bytes,1,opt,name=payload" json:"payload,omitempty"`
}

func (m *CreateIntPointRequest) Reset()                    { *m = CreateIntPointRequest{} }
func (m *CreateIntPointRequest) String() string            { return proto.CompactTextString(m) }
func (*CreateIntPointRequest) ProtoMessage()               {}
func (*CreateIntPointRequest) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{1} }

func (m *CreateIntPointRequest) GetPayload() *IntPoint {
	if m != nil {
		return m.Payload
	}
	return nil
}

type CreateIntPointResponse struct {
	// Convention also requires that the return type be the same and named 'result'
	Result *IntPoint `protobuf:"bytes,1,opt,name=result" json:"result,omitempty"`
}

func (m *CreateIntPointResponse) Reset()                    { *m = CreateIntPointResponse{} }
func (m *CreateIntPointResponse) String() string            { return proto.CompactTextString(m) }
func (*CreateIntPointResponse) ProtoMessage()               {}
func (*CreateIntPointResponse) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{2} }

func (m *CreateIntPointResponse) GetResult() *IntPoint {
	if m != nil {
		return m.Result
	}
	return nil
}

type ReadIntPointRequest struct {
	// For a read request, the id field is the only to be specified
	Id uint32 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
}

func (m *ReadIntPointRequest) Reset()                    { *m = ReadIntPointRequest{} }
func (m *ReadIntPointRequest) String() string            { return proto.CompactTextString(m) }
func (*ReadIntPointRequest) ProtoMessage()               {}
func (*ReadIntPointRequest) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{3} }

func (m *ReadIntPointRequest) GetId() uint32 {
	if m != nil {
		return m.Id
	}
	return 0
}

type ReadIntPointResponse struct {
	// Again the type with 'result' name
	Result *IntPoint `protobuf:"bytes,1,opt,name=result" json:"result,omitempty"`
}

func (m *ReadIntPointResponse) Reset()                    { *m = ReadIntPointResponse{} }
func (m *ReadIntPointResponse) String() string            { return proto.CompactTextString(m) }
func (*ReadIntPointResponse) ProtoMessage()               {}
func (*ReadIntPointResponse) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{4} }

func (m *ReadIntPointResponse) GetResult() *IntPoint {
	if m != nil {
		return m.Result
	}
	return nil
}

type UpdateIntPointRequest struct {
	Payload *IntPoint `protobuf:"bytes,1,opt,name=payload" json:"payload,omitempty"`
}

func (m *UpdateIntPointRequest) Reset()                    { *m = UpdateIntPointRequest{} }
func (m *UpdateIntPointRequest) String() string            { return proto.CompactTextString(m) }
func (*UpdateIntPointRequest) ProtoMessage()               {}
func (*UpdateIntPointRequest) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{5} }

func (m *UpdateIntPointRequest) GetPayload() *IntPoint {
	if m != nil {
		return m.Payload
	}
	return nil
}

type UpdateIntPointResponse struct {
	Result *IntPoint `protobuf:"bytes,1,opt,name=result" json:"result,omitempty"`
}

func (m *UpdateIntPointResponse) Reset()                    { *m = UpdateIntPointResponse{} }
func (m *UpdateIntPointResponse) String() string            { return proto.CompactTextString(m) }
func (*UpdateIntPointResponse) ProtoMessage()               {}
func (*UpdateIntPointResponse) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{6} }

func (m *UpdateIntPointResponse) GetResult() *IntPoint {
	if m != nil {
		return m.Result
	}
	return nil
}

type PatchIntPointRequest struct {
	Payload    *IntPoint                   `protobuf:"bytes,1,opt,name=payload" json:"payload,omitempty"`
	UpdateMask *google_protobuf4.FieldMask `protobuf:"bytes,2,opt,name=update_mask,json=updateMask" json:"update_mask,omitempty"`
}

func (m *PatchIntPointRequest) Reset()                    { *m = PatchIntPointRequest{} }
func (m *PatchIntPointRequest) String() string            { return proto.CompactTextString(m) }
func (*PatchIntPointRequest) ProtoMessage()               {}
func (*PatchIntPointRequest) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{7} }

func (m *PatchIntPointRequest) GetPayload() *IntPoint {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (m *PatchIntPointRequest) GetUpdateMask() *google_protobuf4.FieldMask {
	if m != nil {
		return m.UpdateMask
	}
	return nil
}

type PatchIntPointResponse struct {
	Result *IntPoint `protobuf:"bytes,1,opt,name=result" json:"result,omitempty"`
}

func (m *PatchIntPointResponse) Reset()                    { *m = PatchIntPointResponse{} }
func (m *PatchIntPointResponse) String() string            { return proto.CompactTextString(m) }
func (*PatchIntPointResponse) ProtoMessage()               {}
func (*PatchIntPointResponse) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{8} }

func (m *PatchIntPointResponse) GetResult() *IntPoint {
	if m != nil {
		return m.Result
	}
	return nil
}

type DeleteIntPointRequest struct {
	// Only the id is needed for a delete request
	Id uint32 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
}

func (m *DeleteIntPointRequest) Reset()                    { *m = DeleteIntPointRequest{} }
func (m *DeleteIntPointRequest) String() string            { return proto.CompactTextString(m) }
func (*DeleteIntPointRequest) ProtoMessage()               {}
func (*DeleteIntPointRequest) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{9} }

func (m *DeleteIntPointRequest) GetId() uint32 {
	if m != nil {
		return m.Id
	}
	return 0
}

// By convention, on DELETE no response data is given, so either a
// google.protobuf.empty, or an empty struct is sufficient
type DeleteIntPointResponse struct {
}

func (m *DeleteIntPointResponse) Reset()                    { *m = DeleteIntPointResponse{} }
func (m *DeleteIntPointResponse) String() string            { return proto.CompactTextString(m) }
func (*DeleteIntPointResponse) ProtoMessage()               {}
func (*DeleteIntPointResponse) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{10} }

type ListIntPointResponse struct {
	// Note repeated field and plural name 'results'
	Results []*IntPoint `protobuf:"bytes,1,rep,name=results" json:"results,omitempty"`
}

func (m *ListIntPointResponse) Reset()                    { *m = ListIntPointResponse{} }
func (m *ListIntPointResponse) String() string            { return proto.CompactTextString(m) }
func (*ListIntPointResponse) ProtoMessage()               {}
func (*ListIntPointResponse) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{11} }

func (m *ListIntPointResponse) GetResults() []*IntPoint {
	if m != nil {
		return m.Results
	}
	return nil
}

// A dummy type to demo an rpc that can't be autogenerated
type Something struct {
}

func (m *Something) Reset()                    { *m = Something{} }
func (m *Something) String() string            { return proto.CompactTextString(m) }
func (*Something) ProtoMessage()               {}
func (*Something) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{12} }

func init() {
	proto.RegisterType((*IntPoint)(nil), "example.IntPoint")
	proto.RegisterType((*CreateIntPointRequest)(nil), "example.CreateIntPointRequest")
	proto.RegisterType((*CreateIntPointResponse)(nil), "example.CreateIntPointResponse")
	proto.RegisterType((*ReadIntPointRequest)(nil), "example.ReadIntPointRequest")
	proto.RegisterType((*ReadIntPointResponse)(nil), "example.ReadIntPointResponse")
	proto.RegisterType((*UpdateIntPointRequest)(nil), "example.UpdateIntPointRequest")
	proto.RegisterType((*UpdateIntPointResponse)(nil), "example.UpdateIntPointResponse")
	proto.RegisterType((*PatchIntPointRequest)(nil), "example.PatchIntPointRequest")
	proto.RegisterType((*PatchIntPointResponse)(nil), "example.PatchIntPointResponse")
	proto.RegisterType((*DeleteIntPointRequest)(nil), "example.DeleteIntPointRequest")
	proto.RegisterType((*DeleteIntPointResponse)(nil), "example.DeleteIntPointResponse")
	proto.RegisterType((*ListIntPointResponse)(nil), "example.ListIntPointResponse")
	proto.RegisterType((*Something)(nil), "example.Something")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for IntPointService service

type IntPointServiceClient interface {
	// The convention requires the rpc names have Create/Read/Update/List/Delete
	// as a prefix. The type is inferred from the response (except for delete),
	// so multiple objects can have CURDL handlers in the same service, provided
	// they are given unique suffixes
	Create(ctx context.Context, in *CreateIntPointRequest, opts ...grpc.CallOption) (*CreateIntPointResponse, error)
	Read(ctx context.Context, in *ReadIntPointRequest, opts ...grpc.CallOption) (*ReadIntPointResponse, error)
	Update(ctx context.Context, in *UpdateIntPointRequest, opts ...grpc.CallOption) (*UpdateIntPointResponse, error)
	Patch(ctx context.Context, in *PatchIntPointRequest, opts ...grpc.CallOption) (*PatchIntPointResponse, error)
	List(ctx context.Context, in *google_protobuf2.Empty, opts ...grpc.CallOption) (*ListIntPointResponse, error)
	Delete(ctx context.Context, in *DeleteIntPointRequest, opts ...grpc.CallOption) (*DeleteIntPointResponse, error)
	// CustomMethod can't be autogenerated as it matches no conventions, it will
	// become a stub
	CustomMethod(ctx context.Context, in *google_protobuf2.Empty, opts ...grpc.CallOption) (*google_protobuf2.Empty, error)
	// CreateSomething also doesn't match conventions and will become a stub
	CreateSomething(ctx context.Context, in *Something, opts ...grpc.CallOption) (*Something, error)
}

type intPointServiceClient struct {
	cc *grpc.ClientConn
}

func NewIntPointServiceClient(cc *grpc.ClientConn) IntPointServiceClient {
	return &intPointServiceClient{cc}
}

func (c *intPointServiceClient) Create(ctx context.Context, in *CreateIntPointRequest, opts ...grpc.CallOption) (*CreateIntPointResponse, error) {
	out := new(CreateIntPointResponse)
	err := grpc.Invoke(ctx, "/example.IntPointService/Create", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *intPointServiceClient) Read(ctx context.Context, in *ReadIntPointRequest, opts ...grpc.CallOption) (*ReadIntPointResponse, error) {
	out := new(ReadIntPointResponse)
	err := grpc.Invoke(ctx, "/example.IntPointService/Read", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *intPointServiceClient) Update(ctx context.Context, in *UpdateIntPointRequest, opts ...grpc.CallOption) (*UpdateIntPointResponse, error) {
	out := new(UpdateIntPointResponse)
	err := grpc.Invoke(ctx, "/example.IntPointService/Update", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *intPointServiceClient) Patch(ctx context.Context, in *PatchIntPointRequest, opts ...grpc.CallOption) (*PatchIntPointResponse, error) {
	out := new(PatchIntPointResponse)
	err := grpc.Invoke(ctx, "/example.IntPointService/Patch", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *intPointServiceClient) List(ctx context.Context, in *google_protobuf2.Empty, opts ...grpc.CallOption) (*ListIntPointResponse, error) {
	out := new(ListIntPointResponse)
	err := grpc.Invoke(ctx, "/example.IntPointService/List", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *intPointServiceClient) Delete(ctx context.Context, in *DeleteIntPointRequest, opts ...grpc.CallOption) (*DeleteIntPointResponse, error) {
	out := new(DeleteIntPointResponse)
	err := grpc.Invoke(ctx, "/example.IntPointService/Delete", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *intPointServiceClient) CustomMethod(ctx context.Context, in *google_protobuf2.Empty, opts ...grpc.CallOption) (*google_protobuf2.Empty, error) {
	out := new(google_protobuf2.Empty)
	err := grpc.Invoke(ctx, "/example.IntPointService/CustomMethod", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *intPointServiceClient) CreateSomething(ctx context.Context, in *Something, opts ...grpc.CallOption) (*Something, error) {
	out := new(Something)
	err := grpc.Invoke(ctx, "/example.IntPointService/CreateSomething", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for IntPointService service

type IntPointServiceServer interface {
	// The convention requires the rpc names have Create/Read/Update/List/Delete
	// as a prefix. The type is inferred from the response (except for delete),
	// so multiple objects can have CURDL handlers in the same service, provided
	// they are given unique suffixes
	Create(context.Context, *CreateIntPointRequest) (*CreateIntPointResponse, error)
	Read(context.Context, *ReadIntPointRequest) (*ReadIntPointResponse, error)
	Update(context.Context, *UpdateIntPointRequest) (*UpdateIntPointResponse, error)
	Patch(context.Context, *PatchIntPointRequest) (*PatchIntPointResponse, error)
	List(context.Context, *google_protobuf2.Empty) (*ListIntPointResponse, error)
	Delete(context.Context, *DeleteIntPointRequest) (*DeleteIntPointResponse, error)
	// CustomMethod can't be autogenerated as it matches no conventions, it will
	// become a stub
	CustomMethod(context.Context, *google_protobuf2.Empty) (*google_protobuf2.Empty, error)
	// CreateSomething also doesn't match conventions and will become a stub
	CreateSomething(context.Context, *Something) (*Something, error)
}

func RegisterIntPointServiceServer(s *grpc.Server, srv IntPointServiceServer) {
	s.RegisterService(&_IntPointService_serviceDesc, srv)
}

func _IntPointService_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateIntPointRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntPointServiceServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/example.IntPointService/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntPointServiceServer).Create(ctx, req.(*CreateIntPointRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntPointService_Read_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadIntPointRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntPointServiceServer).Read(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/example.IntPointService/Read",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntPointServiceServer).Read(ctx, req.(*ReadIntPointRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntPointService_Update_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateIntPointRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntPointServiceServer).Update(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/example.IntPointService/Update",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntPointServiceServer).Update(ctx, req.(*UpdateIntPointRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntPointService_Patch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PatchIntPointRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntPointServiceServer).Patch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/example.IntPointService/Patch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntPointServiceServer).Patch(ctx, req.(*PatchIntPointRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntPointService_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(google_protobuf2.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntPointServiceServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/example.IntPointService/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntPointServiceServer).List(ctx, req.(*google_protobuf2.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntPointService_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteIntPointRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntPointServiceServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/example.IntPointService/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntPointServiceServer).Delete(ctx, req.(*DeleteIntPointRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntPointService_CustomMethod_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(google_protobuf2.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntPointServiceServer).CustomMethod(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/example.IntPointService/CustomMethod",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntPointServiceServer).CustomMethod(ctx, req.(*google_protobuf2.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntPointService_CreateSomething_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Something)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntPointServiceServer).CreateSomething(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/example.IntPointService/CreateSomething",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntPointServiceServer).CreateSomething(ctx, req.(*Something))
	}
	return interceptor(ctx, in, info, handler)
}

var _IntPointService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "example.IntPointService",
	HandlerType: (*IntPointServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Create",
			Handler:    _IntPointService_Create_Handler,
		},
		{
			MethodName: "Read",
			Handler:    _IntPointService_Read_Handler,
		},
		{
			MethodName: "Update",
			Handler:    _IntPointService_Update_Handler,
		},
		{
			MethodName: "Patch",
			Handler:    _IntPointService_Patch_Handler,
		},
		{
			MethodName: "List",
			Handler:    _IntPointService_List_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _IntPointService_Delete_Handler,
		},
		{
			MethodName: "CustomMethod",
			Handler:    _IntPointService_CustomMethod_Handler,
		},
		{
			MethodName: "CreateSomething",
			Handler:    _IntPointService_CreateSomething_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "example/feature_demo/demo_service.proto",
}

// Client API for IntPointTxn service

type IntPointTxnClient interface {
	// The convention requires the rpc names have Create/Read/Update/List/Delete
	// as a prefix. The type is inferred from the response (except for delete),
	// so multiple objects can have CURDL handlers in the same service, provided
	// they are given unique suffixes
	Create(ctx context.Context, in *CreateIntPointRequest, opts ...grpc.CallOption) (*CreateIntPointResponse, error)
	Read(ctx context.Context, in *ReadIntPointRequest, opts ...grpc.CallOption) (*ReadIntPointResponse, error)
	Update(ctx context.Context, in *UpdateIntPointRequest, opts ...grpc.CallOption) (*UpdateIntPointResponse, error)
	Patch(ctx context.Context, in *PatchIntPointRequest, opts ...grpc.CallOption) (*PatchIntPointResponse, error)
	List(ctx context.Context, in *google_protobuf2.Empty, opts ...grpc.CallOption) (*ListIntPointResponse, error)
	Delete(ctx context.Context, in *DeleteIntPointRequest, opts ...grpc.CallOption) (*DeleteIntPointResponse, error)
	// CustomMethod can't be autogenerated as it matches no conventions, it will
	// become a stub
	CustomMethod(ctx context.Context, in *google_protobuf2.Empty, opts ...grpc.CallOption) (*google_protobuf2.Empty, error)
	// CreateSomething also doesn't match conventions and will become a stub
	CreateSomething(ctx context.Context, in *Something, opts ...grpc.CallOption) (*Something, error)
}

type intPointTxnClient struct {
	cc *grpc.ClientConn
}

func NewIntPointTxnClient(cc *grpc.ClientConn) IntPointTxnClient {
	return &intPointTxnClient{cc}
}

func (c *intPointTxnClient) Create(ctx context.Context, in *CreateIntPointRequest, opts ...grpc.CallOption) (*CreateIntPointResponse, error) {
	out := new(CreateIntPointResponse)
	err := grpc.Invoke(ctx, "/example.IntPointTxn/Create", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *intPointTxnClient) Read(ctx context.Context, in *ReadIntPointRequest, opts ...grpc.CallOption) (*ReadIntPointResponse, error) {
	out := new(ReadIntPointResponse)
	err := grpc.Invoke(ctx, "/example.IntPointTxn/Read", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *intPointTxnClient) Update(ctx context.Context, in *UpdateIntPointRequest, opts ...grpc.CallOption) (*UpdateIntPointResponse, error) {
	out := new(UpdateIntPointResponse)
	err := grpc.Invoke(ctx, "/example.IntPointTxn/Update", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *intPointTxnClient) Patch(ctx context.Context, in *PatchIntPointRequest, opts ...grpc.CallOption) (*PatchIntPointResponse, error) {
	out := new(PatchIntPointResponse)
	err := grpc.Invoke(ctx, "/example.IntPointTxn/Patch", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *intPointTxnClient) List(ctx context.Context, in *google_protobuf2.Empty, opts ...grpc.CallOption) (*ListIntPointResponse, error) {
	out := new(ListIntPointResponse)
	err := grpc.Invoke(ctx, "/example.IntPointTxn/List", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *intPointTxnClient) Delete(ctx context.Context, in *DeleteIntPointRequest, opts ...grpc.CallOption) (*DeleteIntPointResponse, error) {
	out := new(DeleteIntPointResponse)
	err := grpc.Invoke(ctx, "/example.IntPointTxn/Delete", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *intPointTxnClient) CustomMethod(ctx context.Context, in *google_protobuf2.Empty, opts ...grpc.CallOption) (*google_protobuf2.Empty, error) {
	out := new(google_protobuf2.Empty)
	err := grpc.Invoke(ctx, "/example.IntPointTxn/CustomMethod", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *intPointTxnClient) CreateSomething(ctx context.Context, in *Something, opts ...grpc.CallOption) (*Something, error) {
	out := new(Something)
	err := grpc.Invoke(ctx, "/example.IntPointTxn/CreateSomething", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for IntPointTxn service

type IntPointTxnServer interface {
	// The convention requires the rpc names have Create/Read/Update/List/Delete
	// as a prefix. The type is inferred from the response (except for delete),
	// so multiple objects can have CURDL handlers in the same service, provided
	// they are given unique suffixes
	Create(context.Context, *CreateIntPointRequest) (*CreateIntPointResponse, error)
	Read(context.Context, *ReadIntPointRequest) (*ReadIntPointResponse, error)
	Update(context.Context, *UpdateIntPointRequest) (*UpdateIntPointResponse, error)
	Patch(context.Context, *PatchIntPointRequest) (*PatchIntPointResponse, error)
	List(context.Context, *google_protobuf2.Empty) (*ListIntPointResponse, error)
	Delete(context.Context, *DeleteIntPointRequest) (*DeleteIntPointResponse, error)
	// CustomMethod can't be autogenerated as it matches no conventions, it will
	// become a stub
	CustomMethod(context.Context, *google_protobuf2.Empty) (*google_protobuf2.Empty, error)
	// CreateSomething also doesn't match conventions and will become a stub
	CreateSomething(context.Context, *Something) (*Something, error)
}

func RegisterIntPointTxnServer(s *grpc.Server, srv IntPointTxnServer) {
	s.RegisterService(&_IntPointTxn_serviceDesc, srv)
}

func _IntPointTxn_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateIntPointRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntPointTxnServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/example.IntPointTxn/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntPointTxnServer).Create(ctx, req.(*CreateIntPointRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntPointTxn_Read_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadIntPointRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntPointTxnServer).Read(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/example.IntPointTxn/Read",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntPointTxnServer).Read(ctx, req.(*ReadIntPointRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntPointTxn_Update_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateIntPointRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntPointTxnServer).Update(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/example.IntPointTxn/Update",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntPointTxnServer).Update(ctx, req.(*UpdateIntPointRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntPointTxn_Patch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PatchIntPointRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntPointTxnServer).Patch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/example.IntPointTxn/Patch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntPointTxnServer).Patch(ctx, req.(*PatchIntPointRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntPointTxn_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(google_protobuf2.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntPointTxnServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/example.IntPointTxn/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntPointTxnServer).List(ctx, req.(*google_protobuf2.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntPointTxn_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteIntPointRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntPointTxnServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/example.IntPointTxn/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntPointTxnServer).Delete(ctx, req.(*DeleteIntPointRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntPointTxn_CustomMethod_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(google_protobuf2.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntPointTxnServer).CustomMethod(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/example.IntPointTxn/CustomMethod",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntPointTxnServer).CustomMethod(ctx, req.(*google_protobuf2.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntPointTxn_CreateSomething_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Something)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntPointTxnServer).CreateSomething(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/example.IntPointTxn/CreateSomething",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntPointTxnServer).CreateSomething(ctx, req.(*Something))
	}
	return interceptor(ctx, in, info, handler)
}

var _IntPointTxn_serviceDesc = grpc.ServiceDesc{
	ServiceName: "example.IntPointTxn",
	HandlerType: (*IntPointTxnServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Create",
			Handler:    _IntPointTxn_Create_Handler,
		},
		{
			MethodName: "Read",
			Handler:    _IntPointTxn_Read_Handler,
		},
		{
			MethodName: "Update",
			Handler:    _IntPointTxn_Update_Handler,
		},
		{
			MethodName: "Patch",
			Handler:    _IntPointTxn_Patch_Handler,
		},
		{
			MethodName: "List",
			Handler:    _IntPointTxn_List_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _IntPointTxn_Delete_Handler,
		},
		{
			MethodName: "CustomMethod",
			Handler:    _IntPointTxn_CustomMethod_Handler,
		},
		{
			MethodName: "CreateSomething",
			Handler:    _IntPointTxn_CreateSomething_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "example/feature_demo/demo_service.proto",
}

func init() { proto.RegisterFile("example/feature_demo/demo_service.proto", fileDescriptor1) }

var fileDescriptor1 = []byte{
	// 630 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xec, 0x96, 0x5f, 0x6f, 0x93, 0x50,
	0x18, 0xc6, 0x4b, 0xd7, 0xb1, 0xee, 0x65, 0x6e, 0x7a, 0x5c, 0x97, 0x16, 0x75, 0x36, 0x27, 0x31,
	0x9b, 0x59, 0x06, 0x49, 0xbd, 0x5b, 0x13, 0xff, 0xac, 0xdd, 0xa2, 0x99, 0x4b, 0x96, 0x4e, 0x2f,
	0xdc, 0x4d, 0x43, 0xcb, 0x29, 0x25, 0x03, 0x0e, 0xc2, 0xc1, 0xb4, 0x77, 0x7e, 0xad, 0xf6, 0xc3,
	0xe8, 0x57, 0x31, 0x70, 0x28, 0x22, 0x85, 0xc4, 0xc6, 0xc5, 0x78, 0xe1, 0xcd, 0xb2, 0xf3, 0x9e,
	0xa7, 0xcf, 0x79, 0x5e, 0xde, 0x1f, 0x27, 0xc0, 0x01, 0x99, 0x68, 0xb6, 0x6b, 0x11, 0x75, 0x44,
	0x34, 0x16, 0x78, 0xa4, 0xaf, 0x13, 0x9b, 0xaa, 0xe1, 0x9f, 0xbe, 0x4f, 0xbc, 0x2f, 0xe6, 0x90,
	0x28, 0xae, 0x47, 0x19, 0x45, 0x1b, 0xb1, 0x50, 0x7e, 0x64, 0x50, 0x6a, 0x58, 0x44, 0x8d, 0xca,
	0x83, 0x60, 0xa4, 0x12, 0xdb, 0x65, 0x53, 0xae, 0x92, 0x4f, 0x0c, 0x93, 0x8d, 0x83, 0x81, 0x32,
	0xa4, 0xb6, 0x6a, 0x3a, 0x23, 0x3a, 0xb0, 0xe8, 0x84, 0xba, 0xc4, 0xe1, 0xea, 0xe1, 0xb1, 0x41,
	0x9c, 0x63, 0x83, 0x7a, 0xb6, 0x4a, 0x5d, 0x66, 0x52, 0xc7, 0x57, 0xc3, 0x45, 0xfc, 0xdb, 0x66,
	0xd6, 0x78, 0x64, 0x12, 0x4b, 0xef, 0xdb, 0x9a, 0x7f, 0xcb, 0x15, 0xf8, 0x25, 0x54, 0xdf, 0x39,
	0xec, 0x8a, 0x9a, 0x0e, 0x43, 0xdb, 0x50, 0x36, 0xf5, 0xba, 0xd0, 0x14, 0x0e, 0xef, 0xf5, 0xca,
	0xa6, 0x8e, 0xb6, 0x40, 0x98, 0xd4, 0xcb, 0x4d, 0xe1, 0x70, 0xbd, 0x27, 0x4c, 0xc2, 0xd5, 0xb4,
	0xbe, 0xc6, 0x57, 0xd3, 0x13, 0x71, 0x3e, 0x6b, 0x94, 0xab, 0x02, 0xee, 0x42, 0xad, 0xe3, 0x11,
	0x8d, 0x91, 0x85, 0x4b, 0x8f, 0x7c, 0x0e, 0x88, 0xcf, 0xd0, 0x11, 0x6c, 0xb8, 0xda, 0xd4, 0xa2,
	0x1a, 0x77, 0x94, 0x5a, 0x0f, 0x94, 0xb8, 0x5d, 0x25, 0x91, 0x2e, 0x14, 0xb8, 0x03, 0x7b, 0x59,
	0x17, 0xdf, 0xa5, 0x8e, 0x4f, 0xd0, 0x73, 0x10, 0x3d, 0xe2, 0x07, 0x16, 0x2b, 0x76, 0x89, 0x05,
	0xf8, 0x19, 0x3c, 0xec, 0x11, 0x4d, 0xcf, 0x06, 0xc9, 0x74, 0x85, 0xdf, 0xc0, 0xee, 0xaf, 0xb2,
	0xd5, 0x4f, 0xea, 0x42, 0xed, 0xa3, 0xab, 0xdf, 0x41, 0xd3, 0x59, 0x97, 0xd5, 0xa3, 0x7c, 0x15,
	0x60, 0xf7, 0x4a, 0x63, 0xc3, 0xf1, 0x9f, 0x44, 0x41, 0x6d, 0x90, 0x82, 0x28, 0x4a, 0x84, 0x46,
	0x34, 0x73, 0xa9, 0x25, 0x2b, 0x9c, 0x1e, 0x65, 0x41, 0x8f, 0x72, 0x1e, 0xd2, 0x73, 0xa9, 0xf9,
	0xb7, 0x3d, 0xe0, 0xf2, 0xf0, 0x7f, 0x7c, 0x0a, 0xb5, 0x4c, 0x82, 0xd5, 0xdb, 0x38, 0x80, 0x5a,
	0x97, 0x58, 0x64, 0xf9, 0x89, 0x66, 0xa7, 0x57, 0x87, 0xbd, 0xac, 0x90, 0x9f, 0x86, 0x3b, 0xb0,
	0xfb, 0xde, 0xf4, 0xd9, 0x52, 0x8a, 0x23, 0xd8, 0xe0, 0x87, 0xf8, 0x75, 0xa1, 0xb9, 0x56, 0xf0,
	0x20, 0x62, 0x05, 0x96, 0x60, 0xf3, 0x9a, 0xda, 0x84, 0x8d, 0x4d, 0xc7, 0x68, 0x7d, 0xaf, 0xc0,
	0xce, 0x42, 0x72, 0xcd, 0xdf, 0x5c, 0x74, 0x01, 0x22, 0x27, 0x15, 0xed, 0x27, 0x36, 0xb9, 0x2f,
	0x80, 0xfc, 0xb4, 0x70, 0x3f, 0x0e, 0x5c, 0x42, 0x67, 0x50, 0x09, 0x51, 0x44, 0x8f, 0x13, 0x69,
	0x0e, 0xc0, 0xf2, 0x93, 0x82, 0xdd, 0xc4, 0xe6, 0x02, 0x44, 0x0e, 0x52, 0x2a, 0x53, 0x2e, 0x9f,
	0xa9, 0x4c, 0xf9, 0xe4, 0xe1, 0x12, 0x7a, 0x0b, 0xeb, 0xd1, 0x34, 0xd1, 0xcf, 0x63, 0xf3, 0xf8,
	0x92, 0xf7, 0x8b, 0xb6, 0x13, 0xa7, 0x57, 0x50, 0x09, 0x07, 0x82, 0xf6, 0x96, 0x38, 0x3a, 0x0b,
	0xaf, 0xb7, 0x54, 0x5f, 0x79, 0x73, 0xc3, 0x25, 0xf4, 0x09, 0x44, 0x3e, 0xeb, 0x54, 0x5f, 0xb9,
	0x94, 0xa4, 0xfa, 0x2a, 0x80, 0x63, 0x7b, 0x3e, 0x6b, 0x40, 0xea, 0xaa, 0x7b, 0x0d, 0x5b, 0x9d,
	0xc0, 0x67, 0xd4, 0xbe, 0x24, 0x6c, 0x4c, 0xf5, 0xc2, 0x8c, 0x05, 0x75, 0x5c, 0x42, 0x6d, 0xd8,
	0xe1, 0x73, 0x4d, 0x78, 0x41, 0x28, 0x49, 0x91, 0xd4, 0xe4, 0x9c, 0x1a, 0x2e, 0xc9, 0xf1, 0xed,
	0xd9, 0xfa, 0x56, 0x01, 0x69, 0x91, 0xe9, 0xc3, 0xc4, 0xf9, 0x4f, 0xd7, 0x5f, 0xa4, 0xeb, 0xe6,
	0xee, 0xe8, 0xda, 0x99, 0xcf, 0x1a, 0x12, 0x6c, 0x9a, 0x0e, 0xeb, 0xbb, 0xff, 0x02, 0x5e, 0xd5,
	0xf9, 0xac, 0x51, 0xa9, 0x0a, 0xf7, 0x85, 0xd3, 0xf3, 0x9b, 0xee, 0xef, 0x7e, 0x3e, 0xe4, 0x7d,
	0xb5, 0xb4, 0xe3, 0xe2, 0x40, 0x8c, 0xd4, 0x2f, 0x7e, 0x04, 0x00, 0x00, 0xff, 0xff, 0x93, 0xf7,
	0x7b, 0x4e, 0xdc, 0x08, 0x00, 0x00,
}
