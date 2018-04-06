// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: example/feature_demo/test2.proto

package example

import context "context"
import errors "errors"
import gorm "github.com/jinzhu/gorm"
import ops "github.com/Infoblox-CTO/ngp.api.toolkit/op/gorm"
import grpc "google.golang.org/grpc"
import uuid "github.com/satori/go.uuid"
import gtypes "github.com/infobloxopen/protoc-gen-gorm/types"
import time "time"
import ptypes "github.com/golang/protobuf/ptypes"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
import _ "github.com/infobloxopen/protoc-gen-gorm/options"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
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

type PointDefaultServer struct {
	DB *gorm.DB
}

// CreateIntPoint ...
func (m *PointDefaultServer) CreateIntPoint(ctx context.Context, in *CreateIntPointRequest, opts ...grpc.CallOption) (*CreateIntPointResponse, error) {
	var out CreateIntPointResponse
	res, err := DefaultCreateIntPoint(ctx, in.GetPayload(), db)
	out.Result = res
	return &out, err
}

// ReadIntPoint ...
func (m *PointDefaultServer) ReadIntPoint(ctx context.Context, in *ReadIntPointRequest, opts ...grpc.CallOption) (*ReadIntPointResponse, error) {
	var out ReadIntPointResponse
	res, err := DefaultReadIntPoint(ctx, in.GetPayload(), db)
	out.Result = res
	return &out, err
}

// UpdateIntPoint ...
func (m *PointDefaultServer) UpdateIntPoint(ctx context.Context, in *UpdateIntPointRequest, opts ...grpc.CallOption) (*UpdateIntPointResponse, error) {
	var out UpdateIntPointResponse
	res, err := DefaultUpdateIntPoint(ctx, in.GetPayload(), db)
	out.Result = res
	return &out, err
}

// ListIntPoint ...
func (m *PointDefaultServer) ListIntPoint(ctx context.Context, in *google_protobuf2.Empty, opts ...grpc.CallOption) (*ListIntPointResponse, error) {
	var out ListIntPointResponse
	res, err := DefaultListIntPoint(ctx, db)
	l.Results = res
	return &out, err
}

// DeleteIntPoint ...
func (m *PointDefaultServer) DeleteIntPoint(ctx context.Context, in *DeleteIntPointRequest, opts ...grpc.CallOption) (*google_protobuf2.Empty, error) {
	return nil, DefaultDeleteIntPoint(ctx, in.GetPayload(), db)
}

// CustomMethod ...
func (m *PointDefaultServer) CustomMethod(ctx context.Context, in *google_protobuf2.Empty, opts ...grpc.CallOption) (*google_protobuf2.Empty, error) {
	return nil, nil
}
