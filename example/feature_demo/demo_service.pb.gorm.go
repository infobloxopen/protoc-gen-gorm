// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: example/feature_demo/demo_service.proto

package example

import context "context"
import errors "errors"

import gorm "github.com/jinzhu/gorm"
import ops "github.com/infobloxopen/atlas-app-toolkit/gorm"

import fmt "fmt"
import math "math"
import google_protobuf2 "github.com/golang/protobuf/ptypes/empty"

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
func DefaultCreateIntPoint(ctx context.Context, in *IntPoint, db *gorm.DB) (*IntPoint, error) {
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
func DefaultReadIntPoint(ctx context.Context, in *IntPoint, db *gorm.DB) (*IntPoint, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultReadIntPoint")
	}
	ormParams, err := in.ToORM(ctx)
	if err != nil {
		return nil, err
	}
	ormResponse := IntPointORM{}
	if err = db.Set("gorm:auto_preload", true).Where(&ormParams).First(&ormResponse).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ormResponse.ToPB(ctx)
	return &pbResponse, err
}

// DefaultUpdateIntPoint executes a basic gorm update call
func DefaultUpdateIntPoint(ctx context.Context, in *IntPoint, db *gorm.DB) (*IntPoint, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultUpdateIntPoint")
	}
	if exists, err := DefaultReadIntPoint(ctx, &IntPoint{Id: in.GetId()}, db); err != nil {
		return nil, err
	} else if exists == nil {
		return nil, errors.New("IntPoint not found")
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

func DefaultDeleteIntPoint(ctx context.Context, in *IntPoint, db *gorm.DB) error {
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
func DefaultStrictUpdateIntPoint(ctx context.Context, in *IntPoint, db *gorm.DB) (*IntPoint, error) {
	if in == nil {
		return nil, fmt.Errorf("Nil argument to DefaultCascadedUpdateIntPoint")
	}
	ormObj, err := in.ToORM(ctx)
	if err != nil {
		return nil, err
	}
	if exists, err := DefaultReadIntPoint(ctx, &IntPoint{Id: in.GetId()}, db); err != nil {
		return nil, err
	} else if exists == nil {
		return nil, errors.New("IntPoint not found")
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

// DefaultListIntPoint executes a gorm list call
func DefaultListIntPoint(ctx context.Context, db *gorm.DB) ([]*IntPoint, error) {
	ormResponse := []IntPointORM{}
	db, err := ops.ApplyCollectionOperators(db, ctx)
	if err != nil {
		return nil, err
	}
	if err := db.Set("gorm:auto_preload", true).Find(&ormResponse).Error; err != nil {
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

type IntPointDefaultServer struct {
	DB *gorm.DB
}
type IntPointCreateCustomHandler interface {
	CustomCreate(context.Context, *CreateIntPointRequest) (*CreateIntPointResponse, error)
}

// Create ...
func (m *IntPointDefaultServer) Create(ctx context.Context, in *CreateIntPointRequest) (*CreateIntPointResponse, error) {
	if custom, ok := interface{}(m).(IntPointCreateCustomHandler); ok {
		return custom.CustomCreate(ctx, in)
	}
	res, err := DefaultCreateIntPoint(ctx, in.GetPayload(), m.DB)
	if err != nil {
		return nil, err
	}
	return &CreateIntPointResponse{Result: res}, nil
}

type IntPointReadCustomHandler interface {
	CustomRead(context.Context, *ReadIntPointRequest) (*ReadIntPointResponse, error)
}

// Read ...
func (m *IntPointDefaultServer) Read(ctx context.Context, in *ReadIntPointRequest) (*ReadIntPointResponse, error) {
	if custom, ok := interface{}(m).(IntPointReadCustomHandler); ok {
		return custom.CustomRead(ctx, in)
	}
	res, err := DefaultReadIntPoint(ctx, &IntPoint{Id: in.GetId()}, m.DB)
	if err != nil {
		return nil, err
	}
	return &ReadIntPointResponse{Result: res}, nil
}

type IntPointUpdateCustomHandler interface {
	CustomUpdate(context.Context, *UpdateIntPointRequest) (*UpdateIntPointResponse, error)
}

// Update ...
func (m *IntPointDefaultServer) Update(ctx context.Context, in *UpdateIntPointRequest) (*UpdateIntPointResponse, error) {
	if custom, ok := interface{}(m).(IntPointUpdateCustomHandler); ok {
		return custom.CustomUpdate(ctx, in)
	}
	res, err := DefaultStrictUpdateIntPoint(ctx, in.GetPayload(), m.DB)
	if err != nil {
		return nil, err
	}
	return &UpdateIntPointResponse{Result: res}, nil
}

type IntPointListCustomHandler interface {
	CustomList(context.Context, *google_protobuf2.Empty) (*ListIntPointResponse, error)
}

// List ...
func (m *IntPointDefaultServer) List(ctx context.Context, in *google_protobuf2.Empty) (*ListIntPointResponse, error) {
	if custom, ok := interface{}(m).(IntPointListCustomHandler); ok {
		return custom.CustomList(ctx, in)
	}
	res, err := DefaultListIntPoint(ctx, m.DB)
	if err != nil {
		return nil, err
	}
	return &ListIntPointResponse{Results: res}, nil
}

type IntPointDeleteCustomHandler interface {
	CustomDelete(context.Context, *DeleteIntPointRequest) (*DeleteIntPointResponse, error)
}

// Delete ...
func (m *IntPointDefaultServer) Delete(ctx context.Context, in *DeleteIntPointRequest) (*DeleteIntPointResponse, error) {
	if custom, ok := interface{}(m).(IntPointDeleteCustomHandler); ok {
		return custom.CustomDelete(ctx, in)
	}
	return &DeleteIntPointResponse{}, DefaultDeleteIntPoint(ctx, &IntPoint{Id: in.GetId()}, m.DB)
}

type IntPointCustomMethodCustomHandler interface {
	CustomCustomMethod(context.Context, *google_protobuf2.Empty) (*google_protobuf2.Empty, error)
}

// CustomMethod ...
func (m *IntPointDefaultServer) CustomMethod(ctx context.Context, in *google_protobuf2.Empty) (*google_protobuf2.Empty, error) {
	if custom, ok := interface{}(m).(IntPointCustomMethodCustomHandler); ok {
		return custom.CustomCustomMethod(ctx, in)
	}
	return &google_protobuf2.Empty{}, nil
}

type IntPointCreateSomethingCustomHandler interface {
	CustomCreateSomething(context.Context, *Something) (*Something, error)
}

// CreateSomething ...
func (m *IntPointDefaultServer) CreateSomething(ctx context.Context, in *Something) (*Something, error) {
	if custom, ok := interface{}(m).(IntPointCreateSomethingCustomHandler); ok {
		return custom.CustomCreateSomething(ctx, in)
	}
	return &Something{}, nil
}
