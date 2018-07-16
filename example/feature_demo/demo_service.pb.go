// Code generated by protoc-gen-go. DO NOT EDIT.
// source: example/feature_demo/demo_service.proto

package example

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
import _ "github.com/infobloxopen/protoc-gen-gorm/options"
import infoblox_api "github.com/infobloxopen/atlas-app-toolkit/query"

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
func (*IntPoint) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{0} }

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
func (*CreateIntPointRequest) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{1} }

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
func (*CreateIntPointResponse) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{2} }

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
func (*ReadIntPointRequest) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{3} }

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
func (*ReadIntPointResponse) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{4} }

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
func (*UpdateIntPointRequest) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{5} }

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
func (*UpdateIntPointResponse) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{6} }

func (m *UpdateIntPointResponse) GetResult() *IntPoint {
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
func (*DeleteIntPointRequest) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{7} }

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
func (*DeleteIntPointResponse) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{8} }

type ListIntPointResponse struct {
	// Note repeated field and plural name 'results'
	Results []*IntPoint `protobuf:"bytes,1,rep,name=results" json:"results,omitempty"`
}

func (m *ListIntPointResponse) Reset()                    { *m = ListIntPointResponse{} }
func (m *ListIntPointResponse) String() string            { return proto.CompactTextString(m) }
func (*ListIntPointResponse) ProtoMessage()               {}
func (*ListIntPointResponse) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{9} }

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
func (*Something) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{10} }

type ListIntPointRequest struct {
	Filter  *infoblox_api.Filtering      `protobuf:"bytes,1,opt,name=filter" json:"filter,omitempty"`
	OrderBy *infoblox_api.Sorting        `protobuf:"bytes,2,opt,name=order_by,json=orderBy" json:"order_by,omitempty"`
	Fields  *infoblox_api.FieldSelection `protobuf:"bytes,3,opt,name=fields" json:"fields,omitempty"`
	Paging  *infoblox_api.Pagination     `protobuf:"bytes,4,opt,name=paging" json:"paging,omitempty"`
}

func (m *ListIntPointRequest) Reset()                    { *m = ListIntPointRequest{} }
func (m *ListIntPointRequest) String() string            { return proto.CompactTextString(m) }
func (*ListIntPointRequest) ProtoMessage()               {}
func (*ListIntPointRequest) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{11} }

func (m *ListIntPointRequest) GetFilter() *infoblox_api.Filtering {
	if m != nil {
		return m.Filter
	}
	return nil
}

func (m *ListIntPointRequest) GetOrderBy() *infoblox_api.Sorting {
	if m != nil {
		return m.OrderBy
	}
	return nil
}

func (m *ListIntPointRequest) GetFields() *infoblox_api.FieldSelection {
	if m != nil {
		return m.Fields
	}
	return nil
}

func (m *ListIntPointRequest) GetPaging() *infoblox_api.Pagination {
	if m != nil {
		return m.Paging
	}
	return nil
}

func init() {
	proto.RegisterType((*IntPoint)(nil), "example.IntPoint")
	proto.RegisterType((*CreateIntPointRequest)(nil), "example.CreateIntPointRequest")
	proto.RegisterType((*CreateIntPointResponse)(nil), "example.CreateIntPointResponse")
	proto.RegisterType((*ReadIntPointRequest)(nil), "example.ReadIntPointRequest")
	proto.RegisterType((*ReadIntPointResponse)(nil), "example.ReadIntPointResponse")
	proto.RegisterType((*UpdateIntPointRequest)(nil), "example.UpdateIntPointRequest")
	proto.RegisterType((*UpdateIntPointResponse)(nil), "example.UpdateIntPointResponse")
	proto.RegisterType((*DeleteIntPointRequest)(nil), "example.DeleteIntPointRequest")
	proto.RegisterType((*DeleteIntPointResponse)(nil), "example.DeleteIntPointResponse")
	proto.RegisterType((*ListIntPointResponse)(nil), "example.ListIntPointResponse")
	proto.RegisterType((*Something)(nil), "example.Something")
	proto.RegisterType((*ListIntPointRequest)(nil), "example.ListIntPointRequest")
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
	List(ctx context.Context, in *ListIntPointRequest, opts ...grpc.CallOption) (*ListIntPointResponse, error)
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

func (c *intPointServiceClient) List(ctx context.Context, in *ListIntPointRequest, opts ...grpc.CallOption) (*ListIntPointResponse, error) {
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
	List(context.Context, *ListIntPointRequest) (*ListIntPointResponse, error)
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

func _IntPointService_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListIntPointRequest)
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
		return srv.(IntPointServiceServer).List(ctx, req.(*ListIntPointRequest))
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

func init() { proto.RegisterFile("example/feature_demo/demo_service.proto", fileDescriptor2) }

var fileDescriptor2 = []byte{
	// 702 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xdc, 0x96, 0xdf, 0x6e, 0xda, 0x30,
	0x14, 0xc6, 0x09, 0x65, 0x81, 0x1e, 0xba, 0x76, 0x73, 0x4b, 0x07, 0xac, 0xdb, 0xaa, 0x48, 0x53,
	0x3b, 0x55, 0x24, 0x15, 0xdb, 0x55, 0x2b, 0xed, 0x4f, 0xa1, 0x95, 0xb6, 0x6e, 0x52, 0x05, 0xdb,
	0xc5, 0x7a, 0x83, 0x02, 0x39, 0xa4, 0xd6, 0x42, 0xec, 0x26, 0x66, 0x82, 0xf7, 0xd8, 0xdd, 0x9e,
	0x04, 0x9e, 0x66, 0x8f, 0x32, 0x05, 0x07, 0x14, 0xd2, 0x44, 0x5b, 0xa5, 0x5e, 0x4c, 0xbb, 0x41,
	0xb2, 0xfd, 0x9d, 0xe3, 0xef, 0xd8, 0x3f, 0x1f, 0x02, 0x7b, 0x38, 0x32, 0x07, 0xdc, 0x41, 0xa3,
	0x8f, 0xa6, 0x18, 0x7a, 0xd8, 0xb1, 0x70, 0xc0, 0x8c, 0xe0, 0xa7, 0xe3, 0xa3, 0xf7, 0x9d, 0xf6,
	0x50, 0xe7, 0x1e, 0x13, 0x8c, 0xe4, 0x43, 0x61, 0xf5, 0xb1, 0xcd, 0x98, 0xed, 0xa0, 0x31, 0x9b,
	0xee, 0x0e, 0xfb, 0x06, 0x0e, 0xb8, 0x18, 0x4b, 0x55, 0xf5, 0xc8, 0xa6, 0xe2, 0x6a, 0xd8, 0xd5,
	0x7b, 0x6c, 0x60, 0x50, 0xb7, 0xcf, 0xba, 0x0e, 0x1b, 0x31, 0x8e, 0xae, 0x54, 0xf7, 0x6a, 0x36,
	0xba, 0x35, 0x9b, 0x79, 0x03, 0x83, 0x71, 0x41, 0x99, 0xeb, 0x1b, 0xc1, 0x20, 0x8c, 0xfd, 0x90,
	0x16, 0x6b, 0x0a, 0xc7, 0xf4, 0x6b, 0x26, 0xe7, 0x35, 0xc1, 0x98, 0xf3, 0x8d, 0x0a, 0xe3, 0x7a,
	0x88, 0xde, 0xd8, 0xe8, 0x31, 0xc7, 0xc1, 0x5e, 0x90, 0xa7, 0xc3, 0x38, 0x7a, 0xa6, 0x60, 0x9e,
	0x2f, 0x73, 0x69, 0xaf, 0xa1, 0xf0, 0xde, 0x15, 0x17, 0x8c, 0xba, 0x82, 0xac, 0x43, 0x96, 0x5a,
	0x65, 0x65, 0x57, 0xd9, 0xbf, 0xdf, 0xca, 0x52, 0x8b, 0xac, 0x81, 0x32, 0x2a, 0x67, 0x77, 0x95,
	0xfd, 0x7b, 0x2d, 0x65, 0x14, 0x8c, 0xc6, 0xe5, 0x15, 0x39, 0x1a, 0x1f, 0xa9, 0xd3, 0x49, 0x25,
	0x5b, 0x50, 0xb4, 0x26, 0x94, 0x1a, 0x1e, 0x9a, 0x02, 0xe7, 0x59, 0x5a, 0x78, 0x3d, 0x44, 0x5f,
	0x90, 0x03, 0xc8, 0x73, 0x73, 0xec, 0x30, 0x53, 0x66, 0x2c, 0xd6, 0x1f, 0xea, 0xe1, 0xc1, 0xe8,
	0x0b, 0xe9, 0x5c, 0xa1, 0x35, 0x60, 0x3b, 0x9e, 0xc5, 0xe7, 0xcc, 0xf5, 0x91, 0xbc, 0x00, 0xd5,
	0x43, 0x7f, 0xe8, 0x88, 0xf4, 0x2c, 0xa1, 0x40, 0x7b, 0x0e, 0x9b, 0x2d, 0x34, 0xad, 0xb8, 0x91,
	0x58, 0x55, 0xda, 0x3b, 0xd8, 0x5a, 0x96, 0xdd, 0x7e, 0xa7, 0x26, 0x94, 0xbe, 0x70, 0xeb, 0x0e,
	0x8a, 0x8e, 0x67, 0xb9, 0xbd, 0x95, 0x3d, 0x28, 0x35, 0xd1, 0xc1, 0x9b, 0x56, 0xe2, 0x65, 0x97,
	0x61, 0x3b, 0x2e, 0x94, 0xbb, 0x69, 0x0d, 0xd8, 0xfa, 0x48, 0x7d, 0x71, 0xc3, 0xc5, 0x01, 0xe4,
	0xe5, 0x26, 0x7e, 0x59, 0xd9, 0x5d, 0x49, 0x29, 0x26, 0x54, 0x68, 0x45, 0x58, 0x6d, 0xb3, 0x01,
	0x8a, 0x2b, 0xea, 0xda, 0xda, 0x2f, 0x05, 0x36, 0x97, 0x53, 0x4a, 0x4f, 0x06, 0xa8, 0x7d, 0xea,
	0x08, 0xf4, 0xc2, 0xba, 0x1e, 0xe9, 0x73, 0x7c, 0x75, 0x93, 0x53, 0xfd, 0x6c, 0xb6, 0x46, 0x5d,
	0xbb, 0x15, 0xca, 0xc8, 0x21, 0x14, 0x98, 0x67, 0xa1, 0xd7, 0xe9, 0x8e, 0x67, 0x20, 0x16, 0xeb,
	0xa5, 0xe5, 0x90, 0x36, 0xf3, 0x44, 0x10, 0x90, 0x9f, 0xc9, 0x4e, 0xc6, 0xe4, 0x55, 0xb0, 0x05,
	0x3a, 0x96, 0x3f, 0x43, 0xb5, 0x58, 0xdf, 0x89, 0x6f, 0x81, 0x8e, 0xd5, 0xc6, 0xf0, 0x35, 0xb4,
	0x42, 0x2d, 0x39, 0x04, 0x95, 0x9b, 0x36, 0x75, 0xed, 0x72, 0x6e, 0x16, 0x55, 0x5e, 0x8e, 0xba,
	0x08, 0xd6, 0x4c, 0x19, 0x21, 0x75, 0xf5, 0x9f, 0x39, 0xd8, 0x98, 0x97, 0xd7, 0x96, 0xef, 0x9f,
	0x9c, 0x83, 0x2a, 0x29, 0x26, 0x4f, 0x17, 0x27, 0x95, 0xf8, 0x38, 0xaa, 0xcf, 0x52, 0xd7, 0xc3,
	0x3b, 0xc9, 0x90, 0x53, 0xc8, 0x05, 0x98, 0x92, 0x9d, 0x85, 0x34, 0x01, 0xee, 0xea, 0x93, 0x94,
	0xd5, 0x45, 0x9a, 0x73, 0x50, 0x25, 0x64, 0x11, 0x4f, 0x89, 0xec, 0x46, 0x3c, 0x25, 0x53, 0x29,
	0x3d, 0x05, 0xd7, 0x1a, 0xf1, 0x94, 0x70, 0xcb, 0x11, 0x4f, 0x49, 0x58, 0x69, 0x19, 0xf2, 0x15,
	0x54, 0x89, 0x62, 0xc4, 0x53, 0x22, 0xc4, 0x11, 0x4f, 0x29, 0xec, 0xae, 0x4f, 0x27, 0x15, 0x88,
	0xb4, 0xb0, 0xb7, 0xb0, 0xd6, 0x18, 0xfa, 0x82, 0x0d, 0x3e, 0xa1, 0xb8, 0x62, 0x16, 0xd9, 0xd6,
	0x65, 0x13, 0xd6, 0xe7, 0x4d, 0x58, 0x3f, 0x0d, 0x9a, 0x70, 0x35, 0x65, 0x5e, 0xcb, 0x90, 0x63,
	0xd8, 0x90, 0x77, 0xb2, 0xc0, 0x99, 0x90, 0x85, 0x8b, 0xc5, 0x5c, 0x35, 0x61, 0x4e, 0xcb, 0x54,
	0xc3, 0xae, 0x58, 0xff, 0x91, 0x83, 0xe2, 0xdc, 0xd3, 0xe7, 0x91, 0xfb, 0xff, 0x93, 0xf1, 0x26,
	0x24, 0x23, 0xed, 0xbc, 0xff, 0xc8, 0xc4, 0xe5, 0xdd, 0x31, 0xb1, 0x31, 0x9d, 0x54, 0x8a, 0xb0,
	0x4a, 0x5d, 0xd1, 0xe1, 0xff, 0x02, 0x14, 0x85, 0xe9, 0xa4, 0x92, 0x2b, 0x28, 0x0f, 0x94, 0x93,
	0xb3, 0xcb, 0xe6, 0xdf, 0xfe, 0xed, 0x27, 0x7d, 0x6d, 0x1c, 0x87, 0x93, 0x5d, 0x75, 0xa6, 0x7e,
	0xf9, 0x3b, 0x00, 0x00, 0xff, 0xff, 0x5b, 0x88, 0x2b, 0xc7, 0x94, 0x08, 0x00, 0x00,
}
