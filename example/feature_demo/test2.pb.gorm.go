// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: example/feature_demo/test2.proto

package example

import context "context"
import gorm "github.com/jinzhu/gorm"
import ops "github.com/Infoblox-CTO/ngp.api.toolkit/op/gorm"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

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
func ConvertIntPointToORM(from IntPoint) IntPointORM {
	to := IntPointORM{}
	to.ID = from.Id
	to.X = from.X
	to.Y = from.Y
	return to
}

// ConvertIntPointFromORM takes an orm object and returns a pb object
func ConvertIntPointFromORM(from IntPointORM) IntPoint {
	to := IntPoint{}
	to.Id = from.ID
	to.X = from.X
	to.Y = from.Y
	return to
}

////////////////////////// CURDL for objects
// DefaultCreateIntPoint executes a basic gorm create call
func DefaultCreateIntPoint(ctx context.Context, in *IntPoint, db *gorm.DB) (*IntPoint, error) {
	if in == nil {
		return nil, fmt.Errorf("Nil argument to DefaultCreateIntPoint")
	}
	ormObj := ConvertIntPointToORM(*in)
	if err := db.Create(&ormObj).Error; err != nil {
		return nil, err
	}
	pbResponse := ConvertIntPointFromORM(ormObj)
	return &pbResponse, nil
}

// DefaultReadIntPoint executes a basic gorm read call
func DefaultReadIntPoint(ctx context.Context, in *IntPoint, db *gorm.DB) (*IntPoint, error) {
	if in == nil {
		return nil, fmt.Errorf("Nil argument to DefaultReadIntPoint")
	}
	ormParams := ConvertIntPointToORM(*in)
	ormResponse := IntPointORM{}
	if err := db.Set("gorm:auto_preload", true).Where(&ormParams).First(&ormResponse).Error; err != nil {
		return nil, err
	}
	pbResponse := ConvertIntPointFromORM(ormResponse)
	return &pbResponse, nil
}

// DefaultUpdateIntPoint executes a basic gorm update call
func DefaultUpdateIntPoint(ctx context.Context, in *IntPoint, db *gorm.DB) (*IntPoint, error) {
	if in == nil {
		return nil, fmt.Errorf("Nil argument to DefaultUpdateIntPoint")
	}
	ormObj := ConvertIntPointToORM(*in)
	if err := db.Save(&ormObj).Error; err != nil {
		return nil, err
	}
	pbResponse := ConvertIntPointFromORM(ormObj)
	return &pbResponse, nil
}

// DefaultDeleteIntPoint executes a basic gorm delete call
func DefaultDeleteIntPoint(ctx context.Context, in *IntPoint, db *gorm.DB) error {
	if in == nil {
		return fmt.Errorf("Nil argument to DefaultDeleteIntPoint")
	}
	ormObj := ConvertIntPointToORM(*in)
	err := db.Where(&ormObj).Delete(&IntPointORM{}).Error
	return err
}

// DefaultListIntPoint executes a basic gorm find call
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
		temp := ConvertIntPointFromORM(responseEntry)
		pbResponse = append(pbResponse, &temp)
	}
	return pbResponse, nil
}
