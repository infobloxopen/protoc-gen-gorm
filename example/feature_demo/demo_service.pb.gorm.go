// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: example/feature_demo/demo_service.proto

package example

import context "context"
import errors "errors"

import field_mask1 "google.golang.org/genproto/protobuf/field_mask"
import gorm1 "github.com/jinzhu/gorm"
import gorm2 "github.com/infobloxopen/atlas-app-toolkit/gorm"

import fmt "fmt"
import math "math"
import google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
import _ "google.golang.org/genproto/protobuf/field_mask"

// Reference imports to suppress errors if they are not otherwise used.
var _ = fmt.Errorf
var _ = math.Inf

type IntPointORM struct {
	Id uint32
	X  int32
	Y  int32
}

// TableName overrides the default tablename generated by GORM
func (IntPointORM) TableName() string {
	return "int_points"
}

// ToORM runs the BeforeToORM hook if present, converts the fields of this
// object to ORM format, runs the AfterToORM hook, then returns the ORM object
func (m *IntPoint) ToORM(ctx context.Context) (IntPointORM, error) {
	to := IntPointORM{}
	var err error
	if prehook, ok := interface{}(m).(IntPointWithBeforeToORM); ok {
		if err = prehook.BeforeToORM(ctx, &to); err != nil {
			return to, err
		}
	}
	to.Id = m.Id
	to.X = m.X
	to.Y = m.Y
	if posthook, ok := interface{}(m).(IntPointWithAfterToORM); ok {
		err = posthook.AfterToORM(ctx, &to)
	}
	return to, err
}

// ToPB runs the BeforeToPB hook if present, converts the fields of this
// object to PB format, runs the AfterToPB hook, then returns the PB object
func (m *IntPointORM) ToPB(ctx context.Context) (IntPoint, error) {
	to := IntPoint{}
	var err error
	if prehook, ok := interface{}(m).(IntPointWithBeforeToPB); ok {
		if err = prehook.BeforeToPB(ctx, &to); err != nil {
			return to, err
		}
	}
	to.Id = m.Id
	to.X = m.X
	to.Y = m.Y
	if posthook, ok := interface{}(m).(IntPointWithAfterToPB); ok {
		err = posthook.AfterToPB(ctx, &to)
	}
	return to, err
}

// The following are interfaces you can implement for special behavior during ORM/PB conversions
// of type IntPoint the arg will be the target, the caller the one being converted from

// IntPointBeforeToORM called before default ToORM code
type IntPointWithBeforeToORM interface {
	BeforeToORM(context.Context, *IntPointORM) error
}

// IntPointAfterToORM called after default ToORM code
type IntPointWithAfterToORM interface {
	AfterToORM(context.Context, *IntPointORM) error
}

// IntPointBeforeToPB called before default ToPB code
type IntPointWithBeforeToPB interface {
	BeforeToPB(context.Context, *IntPoint) error
}

// IntPointAfterToPB called after default ToPB code
type IntPointWithAfterToPB interface {
	AfterToPB(context.Context, *IntPoint) error
}

// DefaultCreateIntPoint executes a basic gorm create call
func DefaultCreateIntPoint(ctx context.Context, in *IntPoint, db *gorm1.DB) (*IntPoint, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultCreateIntPoint")
	}
	ormObj, err := in.ToORM(ctx)
	if err != nil {
		return nil, err
	}
	if err = db.Create(&ormObj).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ormObj.ToPB(ctx)
	return &pbResponse, err
}

// DefaultReadIntPoint executes a basic gorm read call
func DefaultReadIntPoint(ctx context.Context, in *IntPoint, db *gorm1.DB) (*IntPoint, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultReadIntPoint")
	}
	ormParams, err := in.ToORM(ctx)
	if err != nil {
		return nil, err
	}
	ormResponse := IntPointORM{}
	if err = db.Where(&ormParams).First(&ormResponse).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ormResponse.ToPB(ctx)
	return &pbResponse, err
}

// DefaultUpdateIntPoint executes a basic gorm update call
func DefaultUpdateIntPoint(ctx context.Context, in *IntPoint, db *gorm1.DB) (*IntPoint, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultUpdateIntPoint")
	}
	ormObj, err := in.ToORM(ctx)
	if err != nil {
		return nil, err
	}
	if err = db.Save(&ormObj).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ormObj.ToPB(ctx)
	return &pbResponse, err
}

func DefaultDeleteIntPoint(ctx context.Context, in *IntPoint, db *gorm1.DB) error {
	if in == nil {
		return errors.New("Nil argument to DefaultDeleteIntPoint")
	}
	ormObj, err := in.ToORM(ctx)
	if err != nil {
		return err
	}
	if ormObj.Id == 0 {
		return errors.New("A non-zero ID value is required for a delete call")
	}
	err = db.Where(&ormObj).Delete(&IntPointORM{}).Error
	return err
}

// DefaultStrictUpdateIntPoint clears first level 1:many children and then executes a gorm update call
func DefaultStrictUpdateIntPoint(ctx context.Context, in *IntPoint, db *gorm1.DB) (*IntPoint, error) {
	if in == nil {
		return nil, fmt.Errorf("Nil argument to DefaultCascadedUpdateIntPoint")
	}
	ormObj, err := in.ToORM(ctx)
	if err != nil {
		return nil, err
	}
	if err = db.Save(&ormObj).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ormObj.ToPB(ctx)
	if err != nil {
		return nil, err
	}
	return &pbResponse, nil
}

// DefaultPatchIntPoint executes a basic gorm update call with patch behavior
func DefaultPatchIntPoint(ctx context.Context, in *IntPoint, updateMask *field_mask1.FieldMask, db *gorm1.DB) (*IntPoint, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultPatchIntPoint")
	}
	ormParams, err := (&IntPoint{Id: in.GetId()}).ToORM(ctx)
	if err != nil {
		return nil, err
	}
	ormObj := IntPointORM{}
	if err := db.Where(&ormParams).First(&ormObj).Error; err != nil {
		return nil, err
	}
	pbObj, err := ormObj.ToPB(ctx)
	if err != nil {
		return nil, err
	}
	for _, f := range updateMask.GetPaths() {
		if f == "Id" {
			pbObj.Id = in.Id
		}
		if f == "X" {
			pbObj.X = in.X
		}
		if f == "Y" {
			pbObj.Y = in.Y
		}
	}
	ormObj, err = pbObj.ToORM(ctx)
	if err != nil {
		return nil, err
	}
	if err = db.Save(&ormObj).Error; err != nil {
		return nil, err
	}
	pbObj, err = ormObj.ToPB(ctx)
	if err != nil {
		return nil, err
	}
	return &pbObj, err
}

// DefaultListIntPoint executes a gorm list call
func DefaultListIntPoint(ctx context.Context, db *gorm1.DB) ([]*IntPoint, error) {
	ormResponse := []IntPointORM{}
	db, err := gorm2.ApplyCollectionOperators(db, ctx)
	if err != nil {
		return nil, err
	}
	in := IntPoint{}
	ormParams, err := in.ToORM(ctx)
	if err != nil {
		return nil, err
	}
	db = db.Where(&ormParams)
	db = db.Order("id")
	if err := db.Find(&ormResponse).Error; err != nil {
		return nil, err
	}
	pbResponse := []*IntPoint{}
	for _, responseEntry := range ormResponse {
		temp, err := responseEntry.ToPB(ctx)
		if err != nil {
			return nil, err
		}
		pbResponse = append(pbResponse, &temp)
	}
	return pbResponse, nil
}

type IntPointServiceDefaultServer struct {
	DB *gorm1.DB
}
type IntPointServiceCreateCustomHandler interface {
	CustomCreate(context.Context, *CreateIntPointRequest) (*CreateIntPointResponse, error)
}

// Create ...
func (m *IntPointServiceDefaultServer) Create(ctx context.Context, in *CreateIntPointRequest) (*CreateIntPointResponse, error) {
	if custom, ok := interface{}(m).(IntPointServiceCreateCustomHandler); ok {
		return custom.CustomCreate(ctx, in)
	}
	db := m.DB
	res, err := DefaultCreateIntPoint(ctx, in.GetPayload(), db)
	if err != nil {
		return nil, err
	}
	return &CreateIntPointResponse{Result: res}, nil
}

type IntPointServiceReadCustomHandler interface {
	CustomRead(context.Context, *ReadIntPointRequest) (*ReadIntPointResponse, error)
}

// Read ...
func (m *IntPointServiceDefaultServer) Read(ctx context.Context, in *ReadIntPointRequest) (*ReadIntPointResponse, error) {
	if custom, ok := interface{}(m).(IntPointServiceReadCustomHandler); ok {
		return custom.CustomRead(ctx, in)
	}
	db := m.DB
	res, err := DefaultReadIntPoint(ctx, &IntPoint{Id: in.GetId()}, db)
	if err != nil {
		return nil, err
	}
	return &ReadIntPointResponse{Result: res}, nil
}

type IntPointServiceUpdateCustomHandler interface {
	CustomUpdate(context.Context, *UpdateIntPointRequest) (*UpdateIntPointResponse, error)
}

// Update ...
func (m *IntPointServiceDefaultServer) Update(ctx context.Context, in *UpdateIntPointRequest) (*UpdateIntPointResponse, error) {
	if custom, ok := interface{}(m).(IntPointServiceUpdateCustomHandler); ok {
		return custom.CustomUpdate(ctx, in)
	}
	db := m.DB
	res, err := DefaultStrictUpdateIntPoint(ctx, in.GetPayload(), db)
	if err != nil {
		return nil, err
	}
	return &UpdateIntPointResponse{Result: res}, nil
}

type IntPointServicePatchCustomHandler interface {
	CustomPatch(context.Context, *PatchIntPointRequest) (*PatchIntPointResponse, error)
}

// Patch ...
func (m *IntPointServiceDefaultServer) Patch(ctx context.Context, in *PatchIntPointRequest) (*PatchIntPointResponse, error) {
	if custom, ok := interface{}(m).(IntPointServicePatchCustomHandler); ok {
		return custom.CustomPatch(ctx, in)
	}
	db := m.DB
	res, err := DefaultPatchIntPoint(ctx, in.GetPayload(), in.GetUpdateMask(), db)
	if err != nil {
		return nil, err
	}
	return &PatchIntPointResponse{Result: res}, nil
}

type IntPointServiceListCustomHandler interface {
	CustomList(context.Context, *google_protobuf2.Empty) (*ListIntPointResponse, error)
}

// List ...
func (m *IntPointServiceDefaultServer) List(ctx context.Context, in *google_protobuf2.Empty) (*ListIntPointResponse, error) {
	if custom, ok := interface{}(m).(IntPointServiceListCustomHandler); ok {
		return custom.CustomList(ctx, in)
	}
	db := m.DB
	res, err := DefaultListIntPoint(ctx, db)
	if err != nil {
		return nil, err
	}
	return &ListIntPointResponse{Results: res}, nil
}

type IntPointServiceDeleteCustomHandler interface {
	CustomDelete(context.Context, *DeleteIntPointRequest) (*DeleteIntPointResponse, error)
}

// Delete ...
func (m *IntPointServiceDefaultServer) Delete(ctx context.Context, in *DeleteIntPointRequest) (*DeleteIntPointResponse, error) {
	if custom, ok := interface{}(m).(IntPointServiceDeleteCustomHandler); ok {
		return custom.CustomDelete(ctx, in)
	}
	db := m.DB
	return &DeleteIntPointResponse{}, DefaultDeleteIntPoint(ctx, &IntPoint{Id: in.GetId()}, db)
}

type IntPointServiceCustomMethodCustomHandler interface {
	CustomCustomMethod(context.Context, *google_protobuf2.Empty) (*google_protobuf2.Empty, error)
}

// CustomMethod ...
func (m *IntPointServiceDefaultServer) CustomMethod(ctx context.Context, in *google_protobuf2.Empty) (*google_protobuf2.Empty, error) {
	if custom, ok := interface{}(m).(IntPointServiceCustomMethodCustomHandler); ok {
		return custom.CustomCustomMethod(ctx, in)
	}
	return &google_protobuf2.Empty{}, nil
}

type IntPointServiceCreateSomethingCustomHandler interface {
	CustomCreateSomething(context.Context, *Something) (*Something, error)
}

// CreateSomething ...
func (m *IntPointServiceDefaultServer) CreateSomething(ctx context.Context, in *Something) (*Something, error) {
	if custom, ok := interface{}(m).(IntPointServiceCreateSomethingCustomHandler); ok {
		return custom.CustomCreateSomething(ctx, in)
	}
	return &Something{}, nil
}

type IntPointTxnDefaultServer struct {
}
type IntPointTxnCreateCustomHandler interface {
	CustomCreate(context.Context, *CreateIntPointRequest) (*CreateIntPointResponse, error)
}

// Create ...
func (m *IntPointTxnDefaultServer) Create(ctx context.Context, in *CreateIntPointRequest) (*CreateIntPointResponse, error) {
	if custom, ok := interface{}(m).(IntPointTxnCreateCustomHandler); ok {
		return custom.CustomCreate(ctx, in)
	}
	txn, ok := gorm2.FromContext(ctx)
	if !ok {
		return nil, errors.New("Database Transaction For Request Missing")
	}
	db := txn.Begin()
	if db.Error != nil {
		return nil, db.Error
	}
	res, err := DefaultCreateIntPoint(ctx, in.GetPayload(), db)
	if err != nil {
		return nil, err
	}
	return &CreateIntPointResponse{Result: res}, nil
}

type IntPointTxnReadCustomHandler interface {
	CustomRead(context.Context, *ReadIntPointRequest) (*ReadIntPointResponse, error)
}

// Read ...
func (m *IntPointTxnDefaultServer) Read(ctx context.Context, in *ReadIntPointRequest) (*ReadIntPointResponse, error) {
	if custom, ok := interface{}(m).(IntPointTxnReadCustomHandler); ok {
		return custom.CustomRead(ctx, in)
	}
	txn, ok := gorm2.FromContext(ctx)
	if !ok {
		return nil, errors.New("Database Transaction For Request Missing")
	}
	db := txn.Begin()
	if db.Error != nil {
		return nil, db.Error
	}
	res, err := DefaultReadIntPoint(ctx, &IntPoint{Id: in.GetId()}, db)
	if err != nil {
		return nil, err
	}
	return &ReadIntPointResponse{Result: res}, nil
}

type IntPointTxnUpdateCustomHandler interface {
	CustomUpdate(context.Context, *UpdateIntPointRequest) (*UpdateIntPointResponse, error)
}

// Update ...
func (m *IntPointTxnDefaultServer) Update(ctx context.Context, in *UpdateIntPointRequest) (*UpdateIntPointResponse, error) {
	if custom, ok := interface{}(m).(IntPointTxnUpdateCustomHandler); ok {
		return custom.CustomUpdate(ctx, in)
	}
	txn, ok := gorm2.FromContext(ctx)
	if !ok {
		return nil, errors.New("Database Transaction For Request Missing")
	}
	db := txn.Begin()
	if db.Error != nil {
		return nil, db.Error
	}
	res, err := DefaultStrictUpdateIntPoint(ctx, in.GetPayload(), db)
	if err != nil {
		return nil, err
	}
	return &UpdateIntPointResponse{Result: res}, nil
}

type IntPointTxnPatchCustomHandler interface {
	CustomPatch(context.Context, *PatchIntPointRequest) (*PatchIntPointResponse, error)
}

// Patch ...
func (m *IntPointTxnDefaultServer) Patch(ctx context.Context, in *PatchIntPointRequest) (*PatchIntPointResponse, error) {
	if custom, ok := interface{}(m).(IntPointTxnPatchCustomHandler); ok {
		return custom.CustomPatch(ctx, in)
	}
	txn, ok := gorm2.FromContext(ctx)
	if !ok {
		return nil, errors.New("Database Transaction For Request Missing")
	}
	db := txn.Begin()
	if db.Error != nil {
		return nil, db.Error
	}
	res, err := DefaultPatchIntPoint(ctx, in.GetPayload(), in.GetUpdateMask(), db)
	if err != nil {
		return nil, err
	}
	return &PatchIntPointResponse{Result: res}, nil
}

type IntPointTxnListCustomHandler interface {
	CustomList(context.Context, *google_protobuf2.Empty) (*ListIntPointResponse, error)
}

// List ...
func (m *IntPointTxnDefaultServer) List(ctx context.Context, in *google_protobuf2.Empty) (*ListIntPointResponse, error) {
	if custom, ok := interface{}(m).(IntPointTxnListCustomHandler); ok {
		return custom.CustomList(ctx, in)
	}
	txn, ok := gorm2.FromContext(ctx)
	if !ok {
		return nil, errors.New("Database Transaction For Request Missing")
	}
	db := txn.Begin()
	if db.Error != nil {
		return nil, db.Error
	}
	res, err := DefaultListIntPoint(ctx, db)
	if err != nil {
		return nil, err
	}
	return &ListIntPointResponse{Results: res}, nil
}

type IntPointTxnDeleteCustomHandler interface {
	CustomDelete(context.Context, *DeleteIntPointRequest) (*DeleteIntPointResponse, error)
}

// Delete ...
func (m *IntPointTxnDefaultServer) Delete(ctx context.Context, in *DeleteIntPointRequest) (*DeleteIntPointResponse, error) {
	if custom, ok := interface{}(m).(IntPointTxnDeleteCustomHandler); ok {
		return custom.CustomDelete(ctx, in)
	}
	txn, ok := gorm2.FromContext(ctx)
	if !ok {
		return nil, errors.New("Database Transaction For Request Missing")
	}
	db := txn.Begin()
	if db.Error != nil {
		return nil, db.Error
	}
	return &DeleteIntPointResponse{}, DefaultDeleteIntPoint(ctx, &IntPoint{Id: in.GetId()}, db)
}

type IntPointTxnCustomMethodCustomHandler interface {
	CustomCustomMethod(context.Context, *google_protobuf2.Empty) (*google_protobuf2.Empty, error)
}

// CustomMethod ...
func (m *IntPointTxnDefaultServer) CustomMethod(ctx context.Context, in *google_protobuf2.Empty) (*google_protobuf2.Empty, error) {
	if custom, ok := interface{}(m).(IntPointTxnCustomMethodCustomHandler); ok {
		return custom.CustomCustomMethod(ctx, in)
	}
	return &google_protobuf2.Empty{}, nil
}

type IntPointTxnCreateSomethingCustomHandler interface {
	CustomCreateSomething(context.Context, *Something) (*Something, error)
}

// CreateSomething ...
func (m *IntPointTxnDefaultServer) CreateSomething(ctx context.Context, in *Something) (*Something, error) {
	if custom, ok := interface{}(m).(IntPointTxnCreateSomethingCustomHandler); ok {
		return custom.CustomCreateSomething(ctx, in)
	}
	return &Something{}, nil
}
