// Code generated by protoc-gen-go. DO NOT EDIT.
// source: example/user/user.proto

package user // import "github.com/infobloxopen/protoc-gen-gorm/example/user"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import timestamp "github.com/golang/protobuf/ptypes/timestamp"
import resourcepb "github.com/infobloxopen/atlas-app-toolkit/rpc/resource/resourcepb"
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

type User struct {
	Id                   *resourcepb.Identifier `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	CreatedAt            *timestamp.Timestamp   `protobuf:"bytes,2,opt,name=created_at,json=createdAt" json:"created_at,omitempty"`
	UpdatedAt            *timestamp.Timestamp   `protobuf:"bytes,3,opt,name=updated_at,json=updatedAt" json:"updated_at,omitempty"`
	Birthday             *timestamp.Timestamp   `protobuf:"bytes,4,opt,name=birthday" json:"birthday,omitempty"`
	Age                  uint32                 `protobuf:"varint,5,opt,name=age" json:"age,omitempty"`
	Num                  uint32                 `protobuf:"varint,6,opt,name=num" json:"num,omitempty"`
	CreditCard           *CreditCard            `protobuf:"bytes,7,opt,name=credit_card,json=creditCard" json:"credit_card,omitempty"`
	Emails               []*Email               `protobuf:"bytes,8,rep,name=emails" json:"emails,omitempty"`
	Tasks                []*Task                `protobuf:"bytes,9,rep,name=tasks" json:"tasks,omitempty"`
	BillingAddress       *Address               `protobuf:"bytes,10,opt,name=billing_address,json=billingAddress" json:"billing_address,omitempty"`
	ShippingAddress      *Address               `protobuf:"bytes,11,opt,name=shipping_address,json=shippingAddress" json:"shipping_address,omitempty"`
	Languages            []*Language            `protobuf:"bytes,12,rep,name=languages" json:"languages,omitempty"`
	Friends              []*User                `protobuf:"bytes,13,rep,name=friends" json:"friends,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *User) Reset()         { *m = User{} }
func (m *User) String() string { return proto.CompactTextString(m) }
func (*User) ProtoMessage()    {}
func (*User) Descriptor() ([]byte, []int) {
	return fileDescriptor_user_963ef7b99018bc7c, []int{0}
}
func (m *User) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_User.Unmarshal(m, b)
}
func (m *User) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_User.Marshal(b, m, deterministic)
}
func (dst *User) XXX_Merge(src proto.Message) {
	xxx_messageInfo_User.Merge(dst, src)
}
func (m *User) XXX_Size() int {
	return xxx_messageInfo_User.Size(m)
}
func (m *User) XXX_DiscardUnknown() {
	xxx_messageInfo_User.DiscardUnknown(m)
}

var xxx_messageInfo_User proto.InternalMessageInfo

func (m *User) GetId() *resourcepb.Identifier {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *User) GetCreatedAt() *timestamp.Timestamp {
	if m != nil {
		return m.CreatedAt
	}
	return nil
}

func (m *User) GetUpdatedAt() *timestamp.Timestamp {
	if m != nil {
		return m.UpdatedAt
	}
	return nil
}

func (m *User) GetBirthday() *timestamp.Timestamp {
	if m != nil {
		return m.Birthday
	}
	return nil
}

func (m *User) GetAge() uint32 {
	if m != nil {
		return m.Age
	}
	return 0
}

func (m *User) GetNum() uint32 {
	if m != nil {
		return m.Num
	}
	return 0
}

func (m *User) GetCreditCard() *CreditCard {
	if m != nil {
		return m.CreditCard
	}
	return nil
}

func (m *User) GetEmails() []*Email {
	if m != nil {
		return m.Emails
	}
	return nil
}

func (m *User) GetTasks() []*Task {
	if m != nil {
		return m.Tasks
	}
	return nil
}

func (m *User) GetBillingAddress() *Address {
	if m != nil {
		return m.BillingAddress
	}
	return nil
}

func (m *User) GetShippingAddress() *Address {
	if m != nil {
		return m.ShippingAddress
	}
	return nil
}

func (m *User) GetLanguages() []*Language {
	if m != nil {
		return m.Languages
	}
	return nil
}

func (m *User) GetFriends() []*User {
	if m != nil {
		return m.Friends
	}
	return nil
}

type Email struct {
	Id                   *resourcepb.Identifier `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Email                string                 `protobuf:"bytes,2,opt,name=email" json:"email,omitempty"`
	Subscribed           bool                   `protobuf:"varint,3,opt,name=subscribed" json:"subscribed,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *Email) Reset()         { *m = Email{} }
func (m *Email) String() string { return proto.CompactTextString(m) }
func (*Email) ProtoMessage()    {}
func (*Email) Descriptor() ([]byte, []int) {
	return fileDescriptor_user_963ef7b99018bc7c, []int{1}
}
func (m *Email) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Email.Unmarshal(m, b)
}
func (m *Email) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Email.Marshal(b, m, deterministic)
}
func (dst *Email) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Email.Merge(dst, src)
}
func (m *Email) XXX_Size() int {
	return xxx_messageInfo_Email.Size(m)
}
func (m *Email) XXX_DiscardUnknown() {
	xxx_messageInfo_Email.DiscardUnknown(m)
}

var xxx_messageInfo_Email proto.InternalMessageInfo

func (m *Email) GetId() *resourcepb.Identifier {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *Email) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *Email) GetSubscribed() bool {
	if m != nil {
		return m.Subscribed
	}
	return false
}

type Address struct {
	Id                   *resourcepb.Identifier `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Address_1            string                 `protobuf:"bytes,2,opt,name=address_1,json=address1" json:"address_1,omitempty"`
	Address_2            string                 `protobuf:"bytes,3,opt,name=address_2,json=address2" json:"address_2,omitempty"`
	Post                 string                 `protobuf:"bytes,4,opt,name=post" json:"post,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *Address) Reset()         { *m = Address{} }
func (m *Address) String() string { return proto.CompactTextString(m) }
func (*Address) ProtoMessage()    {}
func (*Address) Descriptor() ([]byte, []int) {
	return fileDescriptor_user_963ef7b99018bc7c, []int{2}
}
func (m *Address) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Address.Unmarshal(m, b)
}
func (m *Address) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Address.Marshal(b, m, deterministic)
}
func (dst *Address) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Address.Merge(dst, src)
}
func (m *Address) XXX_Size() int {
	return xxx_messageInfo_Address.Size(m)
}
func (m *Address) XXX_DiscardUnknown() {
	xxx_messageInfo_Address.DiscardUnknown(m)
}

var xxx_messageInfo_Address proto.InternalMessageInfo

func (m *Address) GetId() *resourcepb.Identifier {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *Address) GetAddress_1() string {
	if m != nil {
		return m.Address_1
	}
	return ""
}

func (m *Address) GetAddress_2() string {
	if m != nil {
		return m.Address_2
	}
	return ""
}

func (m *Address) GetPost() string {
	if m != nil {
		return m.Post
	}
	return ""
}

type Language struct {
	Id                   *resourcepb.Identifier `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Name                 string                 `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	Code                 string                 `protobuf:"bytes,3,opt,name=code" json:"code,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *Language) Reset()         { *m = Language{} }
func (m *Language) String() string { return proto.CompactTextString(m) }
func (*Language) ProtoMessage()    {}
func (*Language) Descriptor() ([]byte, []int) {
	return fileDescriptor_user_963ef7b99018bc7c, []int{3}
}
func (m *Language) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Language.Unmarshal(m, b)
}
func (m *Language) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Language.Marshal(b, m, deterministic)
}
func (dst *Language) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Language.Merge(dst, src)
}
func (m *Language) XXX_Size() int {
	return xxx_messageInfo_Language.Size(m)
}
func (m *Language) XXX_DiscardUnknown() {
	xxx_messageInfo_Language.DiscardUnknown(m)
}

var xxx_messageInfo_Language proto.InternalMessageInfo

func (m *Language) GetId() *resourcepb.Identifier {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *Language) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Language) GetCode() string {
	if m != nil {
		return m.Code
	}
	return ""
}

type CreditCard struct {
	Id                   *resourcepb.Identifier `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	CreatedAt            *timestamp.Timestamp   `protobuf:"bytes,2,opt,name=created_at,json=createdAt" json:"created_at,omitempty"`
	UpdatedAt            *timestamp.Timestamp   `protobuf:"bytes,3,opt,name=updated_at,json=updatedAt" json:"updated_at,omitempty"`
	Number               string                 `protobuf:"bytes,4,opt,name=number" json:"number,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *CreditCard) Reset()         { *m = CreditCard{} }
func (m *CreditCard) String() string { return proto.CompactTextString(m) }
func (*CreditCard) ProtoMessage()    {}
func (*CreditCard) Descriptor() ([]byte, []int) {
	return fileDescriptor_user_963ef7b99018bc7c, []int{4}
}
func (m *CreditCard) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreditCard.Unmarshal(m, b)
}
func (m *CreditCard) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreditCard.Marshal(b, m, deterministic)
}
func (dst *CreditCard) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreditCard.Merge(dst, src)
}
func (m *CreditCard) XXX_Size() int {
	return xxx_messageInfo_CreditCard.Size(m)
}
func (m *CreditCard) XXX_DiscardUnknown() {
	xxx_messageInfo_CreditCard.DiscardUnknown(m)
}

var xxx_messageInfo_CreditCard proto.InternalMessageInfo

func (m *CreditCard) GetId() *resourcepb.Identifier {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *CreditCard) GetCreatedAt() *timestamp.Timestamp {
	if m != nil {
		return m.CreatedAt
	}
	return nil
}

func (m *CreditCard) GetUpdatedAt() *timestamp.Timestamp {
	if m != nil {
		return m.UpdatedAt
	}
	return nil
}

func (m *CreditCard) GetNumber() string {
	if m != nil {
		return m.Number
	}
	return ""
}

type Task struct {
	Name        string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Description string `protobuf:"bytes,2,opt,name=description" json:"description,omitempty"`
	Priority    int64  `protobuf:"varint,3,opt,name=priority" json:"priority,omitempty"`
	// The user_id is a "external" reference to the User resource.
	// The "external" means this identifier is not coupled to Task resource
	// or User resource explicitly.
	// The generated code does not asserts that this identifier is belongs to
	// the User resource type.
	UserId               *resourcepb.Identifier `protobuf:"bytes,4,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *Task) Reset()         { *m = Task{} }
func (m *Task) String() string { return proto.CompactTextString(m) }
func (*Task) ProtoMessage()    {}
func (*Task) Descriptor() ([]byte, []int) {
	return fileDescriptor_user_963ef7b99018bc7c, []int{5}
}
func (m *Task) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Task.Unmarshal(m, b)
}
func (m *Task) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Task.Marshal(b, m, deterministic)
}
func (dst *Task) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Task.Merge(dst, src)
}
func (m *Task) XXX_Size() int {
	return xxx_messageInfo_Task.Size(m)
}
func (m *Task) XXX_DiscardUnknown() {
	xxx_messageInfo_Task.DiscardUnknown(m)
}

var xxx_messageInfo_Task proto.InternalMessageInfo

func (m *Task) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Task) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *Task) GetPriority() int64 {
	if m != nil {
		return m.Priority
	}
	return 0
}

func (m *Task) GetUserId() *resourcepb.Identifier {
	if m != nil {
		return m.UserId
	}
	return nil
}

func init() {
	proto.RegisterType((*User)(nil), "user.User")
	proto.RegisterType((*Email)(nil), "user.Email")
	proto.RegisterType((*Address)(nil), "user.Address")
	proto.RegisterType((*Language)(nil), "user.Language")
	proto.RegisterType((*CreditCard)(nil), "user.CreditCard")
	proto.RegisterType((*Task)(nil), "user.Task")
}

func init() { proto.RegisterFile("example/user/user.proto", fileDescriptor_user_963ef7b99018bc7c) }

var fileDescriptor_user_963ef7b99018bc7c = []byte{
	// 692 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xcc, 0x55, 0xcd, 0x6e, 0xdb, 0x38,
	0x10, 0x8e, 0xfc, 0x2b, 0x8f, 0xf3, 0x63, 0x10, 0x8b, 0x5d, 0xc5, 0x0b, 0xec, 0x1a, 0xde, 0x8b,
	0xb1, 0xa8, 0x25, 0xc4, 0x2d, 0x8a, 0x26, 0x01, 0x8a, 0x26, 0x41, 0x0f, 0x01, 0x7a, 0x28, 0x84,
	0xf4, 0xd2, 0x8b, 0x41, 0x89, 0xb4, 0x42, 0x44, 0x12, 0x59, 0x92, 0x02, 0x92, 0x17, 0xe8, 0xa5,
	0xcf, 0xd0, 0x07, 0x49, 0x1e, 0xa2, 0xcf, 0x54, 0x88, 0xa2, 0x14, 0x5d, 0xd2, 0x24, 0xb7, 0x5e,
	0x84, 0xe1, 0xf0, 0x9b, 0x99, 0x8f, 0x9f, 0x3e, 0x4a, 0xf0, 0x17, 0xbd, 0xc6, 0x99, 0x48, 0x69,
	0x50, 0x28, 0x2a, 0xcd, 0xc3, 0x17, 0x92, 0x6b, 0x8e, 0x7a, 0x65, 0x3c, 0x3d, 0x4a, 0x98, 0xbe,
	0x2c, 0x22, 0x3f, 0xe6, 0x59, 0xc0, 0xf2, 0x0d, 0x8f, 0x52, 0x7e, 0xcd, 0x05, 0xcd, 0x03, 0x03,
	0x8a, 0x97, 0x09, 0xcd, 0x97, 0x09, 0x97, 0x59, 0xc0, 0x85, 0x66, 0x3c, 0x57, 0x41, 0xb9, 0xa8,
	0x3a, 0x4c, 0xff, 0x4d, 0x38, 0x4f, 0x52, 0x5a, 0x41, 0xa3, 0x62, 0x13, 0x68, 0x96, 0x51, 0xa5,
	0x71, 0x26, 0x2c, 0xe0, 0xe3, 0x43, 0xcd, 0xb1, 0x4e, 0xb1, 0x5a, 0x62, 0x21, 0x96, 0x9a, 0xf3,
	0xf4, 0x8a, 0xe9, 0x40, 0x8a, 0x38, 0x90, 0x54, 0xf1, 0x42, 0xc6, 0xb4, 0x09, 0x44, 0xd4, 0x84,
	0x55, 0xc7, 0xf9, 0xd7, 0x3e, 0xf4, 0x3e, 0x29, 0x2a, 0xd1, 0x02, 0x3a, 0x8c, 0x78, 0xce, 0xcc,
	0x59, 0x8c, 0x57, 0x9e, 0x5f, 0x37, 0xf7, 0xa5, 0x88, 0xfd, 0x73, 0x42, 0x73, 0xcd, 0x36, 0x8c,
	0xca, 0xb0, 0xc3, 0x08, 0x3a, 0x04, 0x88, 0x25, 0xc5, 0x9a, 0x92, 0x35, 0xd6, 0x5e, 0xc7, 0x54,
	0x4c, 0xfd, 0x8a, 0xba, 0x5f, 0x53, 0xf7, 0x2f, 0x6a, 0xea, 0xe1, 0xc8, 0xa2, 0x4f, 0x74, 0x59,
	0x5a, 0x08, 0x52, 0x97, 0x76, 0x1f, 0x2f, 0xb5, 0xe8, 0x13, 0x8d, 0x5e, 0x83, 0x1b, 0x31, 0xa9,
	0x2f, 0x09, 0xbe, 0xf1, 0x7a, 0x8f, 0x16, 0x36, 0x58, 0xe4, 0x41, 0x17, 0x27, 0xd4, 0xeb, 0xcf,
	0x9c, 0xc5, 0xce, 0xe9, 0xe0, 0xee, 0x76, 0xbf, 0x33, 0x71, 0xc2, 0x32, 0x85, 0x26, 0xd0, 0xcd,
	0x8b, 0xcc, 0x1b, 0x94, 0x3b, 0x61, 0x19, 0xa2, 0x03, 0x18, 0xc7, 0x92, 0x12, 0xa6, 0xd7, 0x31,
	0x96, 0xc4, 0x1b, 0x9a, 0x31, 0x13, 0xdf, 0xbc, 0xe3, 0x33, 0xb3, 0x71, 0x86, 0x25, 0x09, 0x21,
	0x6e, 0x62, 0xf4, 0x1f, 0x0c, 0x68, 0x86, 0x59, 0xaa, 0x3c, 0x77, 0xd6, 0x5d, 0x8c, 0x57, 0xe3,
	0x0a, 0xfd, 0xbe, 0xcc, 0x85, 0x76, 0x0b, 0xad, 0xa0, 0xaf, 0xb1, 0xba, 0x52, 0xde, 0xc8, 0x60,
	0xa0, 0xc2, 0x5c, 0x60, 0x75, 0x75, 0x3a, 0xb9, 0xbb, 0xdd, 0xdf, 0xfe, 0x1f, 0xe6, 0xae, 0x90,
	0x8c, 0x4b, 0xa6, 0x6f, 0xc2, 0x0a, 0x8a, 0xde, 0xc2, 0x5e, 0xc4, 0xd2, 0x94, 0xe5, 0xc9, 0x1a,
	0x13, 0x22, 0xa9, 0x52, 0x1e, 0x18, 0x3e, 0x3b, 0x55, 0xf5, 0x49, 0x95, 0xac, 0x8e, 0x34, 0xdf,
	0x0a, 0x77, 0x2d, 0xda, 0xe6, 0xd1, 0x3b, 0x98, 0xa8, 0x4b, 0x26, 0x44, 0xbb, 0xc1, 0xf8, 0x57,
	0x0d, 0xf6, 0x6a, 0x78, 0xdd, 0xe1, 0x15, 0x8c, 0x52, 0x9c, 0x27, 0x05, 0x4e, 0xa8, 0xf2, 0xb6,
	0x0d, 0xf3, 0xdd, 0xaa, 0xf4, 0x83, 0x4d, 0x57, 0xb5, 0xab, 0xad, 0xf0, 0x1e, 0x88, 0x5e, 0xc0,
	0x70, 0x23, 0x19, 0xcd, 0x89, 0xf2, 0x76, 0xda, 0xa7, 0x2d, 0x4d, 0xd6, 0xe0, 0x6b, 0xc8, 0x91,
	0x7b, 0x77, 0xbb, 0xdf, 0x73, 0x9d, 0x99, 0x33, 0xff, 0x02, 0x7d, 0x23, 0xda, 0x33, 0x8c, 0xf8,
	0x07, 0xf4, 0x8d, 0xc0, 0xc6, 0x83, 0xa3, 0xb0, 0x5a, 0xa0, 0x7f, 0x00, 0x54, 0x11, 0xa9, 0x58,
	0xb2, 0x88, 0x12, 0xe3, 0x31, 0x37, 0x6c, 0x65, 0x5a, 0x23, 0xbf, 0x39, 0x30, 0xac, 0x0f, 0xfb,
	0xf4, 0xa9, 0x7f, 0xc3, 0xc8, 0xea, 0xb9, 0x3e, 0xb0, 0x93, 0x5d, 0x9b, 0x38, 0x68, 0x6f, 0xae,
	0xcc, 0xec, 0xfb, 0xcd, 0x15, 0x42, 0xd0, 0x13, 0x5c, 0x69, 0x63, 0xdf, 0x51, 0x68, 0xe2, 0x16,
	0x9b, 0x0d, 0xb8, 0xb5, 0xae, 0xcf, 0x60, 0x83, 0xa0, 0x97, 0xe3, 0x8c, 0x5a, 0x22, 0x26, 0x2e,
	0x73, 0x31, 0x27, 0xd4, 0xce, 0x37, 0x71, 0x6b, 0xce, 0x0f, 0x07, 0xe0, 0xde, 0xcc, 0xbf, 0xfd,
	0xbd, 0xff, 0x13, 0x06, 0x79, 0x91, 0x45, 0x54, 0x5a, 0xd9, 0xec, 0xaa, 0x75, 0xa0, 0xef, 0x0e,
	0xf4, 0xca, 0xbb, 0xd4, 0x68, 0xe1, 0xb4, 0xb4, 0x98, 0xc1, 0x98, 0xd0, 0xf2, 0xd5, 0x9b, 0xaf,
	0xad, 0x95, 0xa9, 0x9d, 0x42, 0x53, 0x68, 0xee, 0x9e, 0x61, 0xd6, 0x0d, 0x9b, 0x35, 0x3a, 0x84,
	0x61, 0x69, 0xde, 0x35, 0x23, 0xf6, 0x9b, 0xf3, 0xa0, 0x42, 0x95, 0xb5, 0xdf, 0x38, 0xe1, 0xa0,
	0x2c, 0x38, 0x6f, 0xd9, 0xec, 0xf4, 0xf8, 0xf3, 0xe1, 0x53, 0xff, 0x09, 0xed, 0x5f, 0xcb, 0x71,
	0xf9, 0x88, 0x06, 0x06, 0xf2, 0xf2, 0x67, 0x00, 0x00, 0x00, 0xff, 0xff, 0x27, 0xe9, 0xc9, 0xa9,
	0x76, 0x06, 0x00, 0x00,
}
