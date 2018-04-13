// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: example/feature_demo/test2.proto

package example

import context "context"
import errors "errors"

import gorm "github.com/jinzhu/gorm"
import ops "github.com/infobloxopen/atlas-app-toolkit/op/gorm"

import fmt "fmt"
import math "math"
import google_protobuf2 "github.com/golang/protobuf/ptypes/empty"

// Reference imports to suppress errors if they are not otherwise used.
var _ = fmt.Errorf
var _ = math.Inf

// IntPointORM no comment was provided for message type
type IntPointORM struct {
	ID uint32
	X  int32
	Y  int32
}

// TableName overrides the default tablename generated by GORM
func (IntPointORM) TableName() string {
	return "int_points"
}

// ConvertIntPointToORM takes a pb object and returns an orm object
func ConvertIntPointToORM(from IntPoint) (IntPointORM, error) {
	to := IntPointORM{}
	var err error
	to.ID = from.Id
	to.X = from.X
	to.Y = from.Y
	return to, err
}

// ConvertIntPointFromORM takes an orm object and returns a pb object
func ConvertIntPointFromORM(from IntPointORM) (IntPoint, error) {
	to := IntPoint{}
	var err error
	to.Id = from.ID
	to.X = from.X
	to.Y = from.Y
	return to, err
}

////////////////////////// CURDL for objects
// DefaultCreateIntPoint executes a basic gorm create call
func DefaultCreateIntPoint(ctx context.Context, in *IntPoint, db *gorm.DB) (*IntPoint, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultCreateIntPoint")
	}
	ormObj, err := ConvertIntPointToORM(*in)
	if err != nil {
		return nil, err
	}
	if err = db.Create(&ormObj).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ConvertIntPointFromORM(ormObj)
	return &pbResponse, err
}

// DefaultReadIntPoint executes a basic gorm read call
func DefaultReadIntPoint(ctx context.Context, in *IntPoint, db *gorm.DB) (*IntPoint, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultReadIntPoint")
	}
	ormParams, err := ConvertIntPointToORM(*in)
	if err != nil {
		return nil, err
	}
	ormResponse := IntPointORM{}
	if err = db.Set("gorm:auto_preload", true).Where(&ormParams).First(&ormResponse).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ConvertIntPointFromORM(ormResponse)
	return &pbResponse, err
}

// DefaultUpdateIntPoint executes a basic gorm update call
func DefaultUpdateIntPoint(ctx context.Context, in *IntPoint, db *gorm.DB) (*IntPoint, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultUpdateIntPoint")
	}
	ormObj, err := ConvertIntPointToORM(*in)
	if err != nil {
		return nil, err
	}
	if err = db.Save(&ormObj).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ConvertIntPointFromORM(ormObj)
	return &pbResponse, err
}

func DefaultDeleteIntPoint(ctx context.Context, in *IntPoint, db *gorm.DB) error {
	if in == nil {
		return errors.New("Nil argument to DefaultDeleteIntPoint")
	}
	ormObj, err := ConvertIntPointToORM(*in)
	if err != nil {
		return err
	}
	err = db.Where(&ormObj).Delete(&IntPointORM{}).Error
	return err
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
		temp, err := ConvertIntPointFromORM(responseEntry)
		if err != nil {
			return nil, err
		}
		pbResponse = append(pbResponse, &temp)
	}
	return pbResponse, nil
}

// DefaultStrictUpdateIntPoint clears first level 1:many children and then executes a gorm update call
func DefaultStrictUpdateIntPoint(ctx context.Context, in *IntPoint, db *gorm.DB) (*IntPoint, error) {
	if in == nil {
		return nil, fmt.Errorf("Nil argument to DefaultCascadedUpdateIntPoint")
	}
	ormObj, err := ConvertIntPointToORM(*in)
	if err != nil {
		return nil, err
	}
	if err = db.Save(&ormObj).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ConvertIntPointFromORM(ormObj)
	if err != nil {
		return nil, err
	}
	return &pbResponse, nil
}

type IntPointDefaultServer struct {
	DB *gorm.DB
}

// Create ...
func (m *IntPointDefaultServer) Create(ctx context.Context, in *CreateIntPointRequest) (*CreateIntPointResponse, error) {
	res, err := DefaultCreateIntPoint(ctx, in.GetPayload(), m.DB)
	if err != nil {
		return nil, err
	}
	return &CreateIntPointResponse{Result: res}, nil
}

// Read ...
func (m *IntPointDefaultServer) Read(ctx context.Context, in *ReadIntPointRequest) (*ReadIntPointResponse, error) {
	res, err := DefaultReadIntPoint(ctx, &IntPoint{Id: in.GetId()}, m.DB)
	if err != nil {
		return nil, err
	}
	return &ReadIntPointResponse{Result: res}, nil
}

// Update ...
func (m *IntPointDefaultServer) Update(ctx context.Context, in *UpdateIntPointRequest) (*UpdateIntPointResponse, error) {
	res, err := DefaultUpdateIntPoint(ctx, in.GetPayload(), m.DB)
	if err != nil {
		return nil, err
	}
	return &UpdateIntPointResponse{Result: res}, nil
}

// List ...
func (m *IntPointDefaultServer) List(ctx context.Context, in *google_protobuf2.Empty) (*ListIntPointResponse, error) {
	res, err := DefaultListIntPoint(ctx, m.DB)
	if err != nil {
		return nil, err
	}
	return &ListIntPointResponse{Results: res}, nil
}

// Delete ...
func (m *IntPointDefaultServer) Delete(ctx context.Context, in *DeleteIntPointRequest) (*google_protobuf2.Empty, error) {
	return &google_protobuf2.Empty{}, DefaultDeleteIntPoint(ctx, &IntPoint{Id: in.GetId()}, m.DB)
}

// CustomMethod ...
func (m *IntPointDefaultServer) CustomMethod(ctx context.Context, in *google_protobuf2.Empty) (*google_protobuf2.Empty, error) {
	return &google_protobuf2.Empty{}, nil
}

// CreateSomething ...
func (m *IntPointDefaultServer) CreateSomething(ctx context.Context, in *Something) (*Something, error) {
	return &Something{}, nil
}
