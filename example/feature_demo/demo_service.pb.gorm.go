// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: example/feature_demo/demo_service.proto

package example

import context "context"
import errors "errors"

import gateway1 "github.com/infobloxopen/atlas-app-toolkit/gateway"
import gorm1 "github.com/jinzhu/gorm"
import gorm2 "github.com/infobloxopen/atlas-app-toolkit/gorm"
import query1 "github.com/infobloxopen/atlas-app-toolkit/query"

import fmt "fmt"
import math "math"
import google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
import _ "github.com/infobloxopen/atlas-app-toolkit/query"

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
		return nil, fmt.Errorf("Nil argument to DefaultStrictUpdateIntPoint")
	}
	ormObj, err := in.ToORM(ctx)
	if err != nil {
		return nil, err
	}
	count := 1
	err = db.Model(&ormObj).Where("id=?", ormObj.Id).Count(&count).Error
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
	if count == 0 {
		err = gateway1.SetCreated(ctx, "")
	}
	return &pbResponse, err
}

// getCollectionOperators takes collection operator values from corresponding message fields
func getCollectionOperators(in interface{}) (*query1.Filtering, *query1.Sorting, *query1.Pagination, *query1.FieldSelection, error) {
	f := &query1.Filtering{}
	err := gateway1.GetCollectionOp(in, f)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	s := &query1.Sorting{}
	err = gateway1.GetCollectionOp(in, s)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	p := &query1.Pagination{}
	err = gateway1.GetCollectionOp(in, p)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	fs := &query1.FieldSelection{}
	err = gateway1.GetCollectionOp(in, fs)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return f, s, p, fs, nil
}

// DefaultListIntPoint executes a gorm list call
func DefaultListIntPoint(ctx context.Context, db *gorm1.DB, req interface{}) ([]*IntPoint, error) {
	ormResponse := []IntPointORM{}
	f, s, p, fs, err := getCollectionOperators(req)
	if err != nil {
		return nil, err
	}
	db, err = gorm2.ApplyCollectionOperators(db, &IntPointORM{}, f, s, p, fs)
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

// Create ...
func (m *IntPointServiceDefaultServer) Create(ctx context.Context, in *CreateIntPointRequest) (*CreateIntPointResponse, error) {
	db := m.DB
	if custom, ok := interface{}(in).(IntPointServiceIntPointWithBeforeCreate); ok {
		var err error
		ctx, db, err = custom.BeforeCreate(ctx, in, db)
		if err != nil {
			return nil, err
		}
	}
	res, err := DefaultCreateIntPoint(ctx, in.GetPayload(), db)
	if err != nil {
		return nil, err
	}
	return &CreateIntPointResponse{Result: res}, nil
}

// IntPointServiceIntPointWithBeforeCreate called before DefaultCreateIntPoint in the default Create handler
type IntPointServiceIntPointWithBeforeCreate interface {
	BeforeCreate(context.Context, *CreateIntPointRequest, *gorm1.DB) (context.Context, *gorm1.DB, error)
}

// Read ...
func (m *IntPointServiceDefaultServer) Read(ctx context.Context, in *ReadIntPointRequest) (*ReadIntPointResponse, error) {
	db := m.DB
	if custom, ok := interface{}(in).(IntPointServiceIntPointWithBeforeRead); ok {
		var err error
		ctx, db, err = custom.BeforeRead(ctx, in, db)
		if err != nil {
			return nil, err
		}
	}
	res, err := DefaultReadIntPoint(ctx, &IntPoint{Id: in.GetId()}, db)
	if err != nil {
		return nil, err
	}
	return &ReadIntPointResponse{Result: res}, nil
}

// IntPointServiceIntPointWithBeforeRead called before DefaultCreateIntPoint in the default Create handler
type IntPointServiceIntPointWithBeforeRead interface {
	BeforeRead(context.Context, *ReadIntPointRequest, *gorm1.DB) (context.Context, *gorm1.DB, error)
}

// Update ...
func (m *IntPointServiceDefaultServer) Update(ctx context.Context, in *UpdateIntPointRequest) (*UpdateIntPointResponse, error) {
	db := m.DB
	if custom, ok := interface{}(in).(IntPointServiceIntPointWithBeforeUpdate); ok {
		var err error
		ctx, db, err = custom.BeforeUpdate(ctx, in, db)
		if err != nil {
			return nil, err
		}
	}
	res, err := DefaultStrictUpdateIntPoint(ctx, in.GetPayload(), db)
	if err != nil {
		return nil, err
	}
	return &UpdateIntPointResponse{Result: res}, nil
}

// IntPointServiceIntPointWithBeforeUpdate called before DefaultCreateIntPoint in the default Create handler
type IntPointServiceIntPointWithBeforeUpdate interface {
	BeforeUpdate(context.Context, *UpdateIntPointRequest, *gorm1.DB) (context.Context, *gorm1.DB, error)
}

// List ...
func (m *IntPointServiceDefaultServer) List(ctx context.Context, in *ListIntPointRequest) (*ListIntPointResponse, error) {
	db := m.DB
	if custom, ok := interface{}(in).(IntPointServiceIntPointWithBeforeList); ok {
		var err error
		ctx, db, err = custom.BeforeList(ctx, in, db)
		if err != nil {
			return nil, err
		}
	}
	res, err := DefaultListIntPoint(ctx, db, in)
	if err != nil {
		return nil, err
	}
	return &ListIntPointResponse{Results: res}, nil
}

// IntPointServiceIntPointWithBeforeList called before DefaultCreateIntPoint in the default Create handler
type IntPointServiceIntPointWithBeforeList interface {
	BeforeList(context.Context, *ListIntPointRequest, *gorm1.DB) (context.Context, *gorm1.DB, error)
}

// Delete ...
func (m *IntPointServiceDefaultServer) Delete(ctx context.Context, in *DeleteIntPointRequest) (*DeleteIntPointResponse, error) {
	db := m.DB
	if custom, ok := interface{}(in).(IntPointServiceIntPointWithBeforeDelete); ok {
		var err error
		ctx, db, err = custom.BeforeDelete(ctx, in, db)
		if err != nil {
			return nil, err
		}
	}
	return &DeleteIntPointResponse{}, DefaultDeleteIntPoint(ctx, &IntPoint{Id: in.GetId()}, db)
}

// IntPointServiceIntPointWithBeforeDelete called before DefaultCreateIntPoint in the default Create handler
type IntPointServiceIntPointWithBeforeDelete interface {
	BeforeDelete(context.Context, *DeleteIntPointRequest, *gorm1.DB) (context.Context, *gorm1.DB, error)
}

// CustomMethod ...
func (m *IntPointServiceDefaultServer) CustomMethod(ctx context.Context, in *google_protobuf2.Empty) (*google_protobuf2.Empty, error) {
	return &google_protobuf2.Empty{}, nil
}

// CreateSomething ...
func (m *IntPointServiceDefaultServer) CreateSomething(ctx context.Context, in *Something) (*Something, error) {
	return &Something{}, nil
}

type IntPointTxnDefaultServer struct {
}

// Create ...
func (m *IntPointTxnDefaultServer) Create(ctx context.Context, in *CreateIntPointRequest) (*CreateIntPointResponse, error) {
	txn, ok := gorm2.FromContext(ctx)
	if !ok {
		return nil, errors.New("Database Transaction For Request Missing")
	}
	db := txn.Begin()
	if db.Error != nil {
		return nil, db.Error
	}
	if custom, ok := interface{}(in).(IntPointTxnIntPointWithBeforeCreate); ok {
		var err error
		ctx, db, err = custom.BeforeCreate(ctx, in, db)
		if err != nil {
			return nil, err
		}
	}
	res, err := DefaultCreateIntPoint(ctx, in.GetPayload(), db)
	if err != nil {
		return nil, err
	}
	return &CreateIntPointResponse{Result: res}, nil
}

// IntPointTxnIntPointWithBeforeCreate called before DefaultCreateIntPoint in the default Create handler
type IntPointTxnIntPointWithBeforeCreate interface {
	BeforeCreate(context.Context, *CreateIntPointRequest, *gorm1.DB) (context.Context, *gorm1.DB, error)
}

// Read ...
func (m *IntPointTxnDefaultServer) Read(ctx context.Context, in *ReadIntPointRequest) (*ReadIntPointResponse, error) {
	txn, ok := gorm2.FromContext(ctx)
	if !ok {
		return nil, errors.New("Database Transaction For Request Missing")
	}
	db := txn.Begin()
	if db.Error != nil {
		return nil, db.Error
	}
	if custom, ok := interface{}(in).(IntPointTxnIntPointWithBeforeRead); ok {
		var err error
		ctx, db, err = custom.BeforeRead(ctx, in, db)
		if err != nil {
			return nil, err
		}
	}
	res, err := DefaultReadIntPoint(ctx, &IntPoint{Id: in.GetId()}, db)
	if err != nil {
		return nil, err
	}
	return &ReadIntPointResponse{Result: res}, nil
}

// IntPointTxnIntPointWithBeforeRead called before DefaultCreateIntPoint in the default Create handler
type IntPointTxnIntPointWithBeforeRead interface {
	BeforeRead(context.Context, *ReadIntPointRequest, *gorm1.DB) (context.Context, *gorm1.DB, error)
}

// Update ...
func (m *IntPointTxnDefaultServer) Update(ctx context.Context, in *UpdateIntPointRequest) (*UpdateIntPointResponse, error) {
	txn, ok := gorm2.FromContext(ctx)
	if !ok {
		return nil, errors.New("Database Transaction For Request Missing")
	}
	db := txn.Begin()
	if db.Error != nil {
		return nil, db.Error
	}
	if custom, ok := interface{}(in).(IntPointTxnIntPointWithBeforeUpdate); ok {
		var err error
		ctx, db, err = custom.BeforeUpdate(ctx, in, db)
		if err != nil {
			return nil, err
		}
	}
	res, err := DefaultStrictUpdateIntPoint(ctx, in.GetPayload(), db)
	if err != nil {
		return nil, err
	}
	return &UpdateIntPointResponse{Result: res}, nil
}

// IntPointTxnIntPointWithBeforeUpdate called before DefaultCreateIntPoint in the default Create handler
type IntPointTxnIntPointWithBeforeUpdate interface {
	BeforeUpdate(context.Context, *UpdateIntPointRequest, *gorm1.DB) (context.Context, *gorm1.DB, error)
}

// List ...
func (m *IntPointTxnDefaultServer) List(ctx context.Context, in *google_protobuf2.Empty) (*ListIntPointResponse, error) {
	txn, ok := gorm2.FromContext(ctx)
	if !ok {
		return nil, errors.New("Database Transaction For Request Missing")
	}
	db := txn.Begin()
	if db.Error != nil {
		return nil, db.Error
	}
	if custom, ok := interface{}(in).(IntPointTxnIntPointWithBeforeList); ok {
		var err error
		ctx, db, err = custom.BeforeList(ctx, in, db)
		if err != nil {
			return nil, err
		}
	}
	res, err := DefaultListIntPoint(ctx, db, in)
	if err != nil {
		return nil, err
	}
	return &ListIntPointResponse{Results: res}, nil
}

// IntPointTxnIntPointWithBeforeList called before DefaultCreateIntPoint in the default Create handler
type IntPointTxnIntPointWithBeforeList interface {
	BeforeList(context.Context, *google_protobuf2.Empty, *gorm1.DB) (context.Context, *gorm1.DB, error)
}

// Delete ...
func (m *IntPointTxnDefaultServer) Delete(ctx context.Context, in *DeleteIntPointRequest) (*DeleteIntPointResponse, error) {
	txn, ok := gorm2.FromContext(ctx)
	if !ok {
		return nil, errors.New("Database Transaction For Request Missing")
	}
	db := txn.Begin()
	if db.Error != nil {
		return nil, db.Error
	}
	if custom, ok := interface{}(in).(IntPointTxnIntPointWithBeforeDelete); ok {
		var err error
		ctx, db, err = custom.BeforeDelete(ctx, in, db)
		if err != nil {
			return nil, err
		}
	}
	return &DeleteIntPointResponse{}, DefaultDeleteIntPoint(ctx, &IntPoint{Id: in.GetId()}, db)
}

// IntPointTxnIntPointWithBeforeDelete called before DefaultCreateIntPoint in the default Create handler
type IntPointTxnIntPointWithBeforeDelete interface {
	BeforeDelete(context.Context, *DeleteIntPointRequest, *gorm1.DB) (context.Context, *gorm1.DB, error)
}

// CustomMethod ...
func (m *IntPointTxnDefaultServer) CustomMethod(ctx context.Context, in *google_protobuf2.Empty) (*google_protobuf2.Empty, error) {
	return &google_protobuf2.Empty{}, nil
}

// CreateSomething ...
func (m *IntPointTxnDefaultServer) CreateSomething(ctx context.Context, in *Something) (*Something, error) {
	return &Something{}, nil
}
