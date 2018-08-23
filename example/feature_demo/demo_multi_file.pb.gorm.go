// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: example/feature_demo/demo_multi_file.proto

/*
Package example is a generated protocol buffer package.

It is generated from these files:
	example/feature_demo/demo_multi_file.proto
	example/feature_demo/demo_types.proto
	example/feature_demo/demo_service.proto

It has these top-level messages:
	ExternalChild
	TestTypes
	TypeWithID
	MultiaccountTypeWithID
	MultiaccountTypeWithoutID
	APIOnlyType
	PrimaryUUIDType
	PrimaryStringType
	PrimaryIncluded
	IntPoint
	CreateIntPointRequest
	CreateIntPointResponse
	ReadIntPointRequest
	ReadIntPointResponse
	UpdateIntPointRequest
	UpdateIntPointResponse
	DeleteIntPointRequest
	DeleteIntPointResponse
	ListIntPointResponse
	Something
	ListIntPointRequest
*/
package example

import context "context"
import errors "errors"

import field_mask1 "google.golang.org/genproto/protobuf/field_mask"
import gateway1 "github.com/infobloxopen/atlas-app-toolkit/gateway"
import go_uuid1 "github.com/satori/go.uuid"
import gorm1 "github.com/jinzhu/gorm"
import gorm2 "github.com/infobloxopen/atlas-app-toolkit/gorm"

import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = fmt.Errorf
var _ = math.Inf

type ExternalChildORM struct {
	Id                  string
	PrimaryIncludedId   *go_uuid1.UUID
	PrimaryStringTypeId *string
	PrimaryUUIDTypeId   *go_uuid1.UUID
}

// TableName overrides the default tablename generated by GORM
func (ExternalChildORM) TableName() string {
	return "external_children"
}

// ToORM runs the BeforeToORM hook if present, converts the fields of this
// object to ORM format, runs the AfterToORM hook, then returns the ORM object
func (m *ExternalChild) ToORM(ctx context.Context) (ExternalChildORM, error) {
	to := ExternalChildORM{}
	var err error
	if prehook, ok := interface{}(m).(ExternalChildWithBeforeToORM); ok {
		if err = prehook.BeforeToORM(ctx, &to); err != nil {
			return to, err
		}
	}
	to.Id = m.Id
	if posthook, ok := interface{}(m).(ExternalChildWithAfterToORM); ok {
		err = posthook.AfterToORM(ctx, &to)
	}
	return to, err
}

// ToPB runs the BeforeToPB hook if present, converts the fields of this
// object to PB format, runs the AfterToPB hook, then returns the PB object
func (m *ExternalChildORM) ToPB(ctx context.Context) (ExternalChild, error) {
	to := ExternalChild{}
	var err error
	if prehook, ok := interface{}(m).(ExternalChildWithBeforeToPB); ok {
		if err = prehook.BeforeToPB(ctx, &to); err != nil {
			return to, err
		}
	}
	to.Id = m.Id
	if posthook, ok := interface{}(m).(ExternalChildWithAfterToPB); ok {
		err = posthook.AfterToPB(ctx, &to)
	}
	return to, err
}

// The following are interfaces you can implement for special behavior during ORM/PB conversions
// of type ExternalChild the arg will be the target, the caller the one being converted from

// ExternalChildBeforeToORM called before default ToORM code
type ExternalChildWithBeforeToORM interface {
	BeforeToORM(context.Context, *ExternalChildORM) error
}

// ExternalChildAfterToORM called after default ToORM code
type ExternalChildWithAfterToORM interface {
	AfterToORM(context.Context, *ExternalChildORM) error
}

// ExternalChildBeforeToPB called before default ToPB code
type ExternalChildWithBeforeToPB interface {
	BeforeToPB(context.Context, *ExternalChild) error
}

// ExternalChildAfterToPB called after default ToPB code
type ExternalChildWithAfterToPB interface {
	AfterToPB(context.Context, *ExternalChild) error
}

// DefaultCreateExternalChild executes a basic gorm create call
func DefaultCreateExternalChild(ctx context.Context, in *ExternalChild, db *gorm1.DB) (*ExternalChild, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultCreateExternalChild")
	}
	ormObj, err := in.ToORM(ctx)
	if err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(ExternalChildORMWithBeforeCreate); ok {
		if db, err = hook.BeforeCreate(ctx, db); err != nil {
			return nil, err
		}
	}
	if err = db.Create(&ormObj).Error; err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(ExternalChildORMWithAfterCreate); ok {
		if err = hook.AfterCreate(ctx, db); err != nil {
			return nil, err
		}
	}
	pbResponse, err := ormObj.ToPB(ctx)
	return &pbResponse, err
}

type ExternalChildORMWithBeforeCreate interface {
	BeforeCreate(context.Context, *gorm1.DB) (*gorm1.DB, error)
}
type ExternalChildORMWithAfterCreate interface {
	AfterCreate(context.Context, *gorm1.DB) error
}

// DefaultReadExternalChild executes a basic gorm read call
func DefaultReadExternalChild(ctx context.Context, in *ExternalChild, db *gorm1.DB) (*ExternalChild, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultReadExternalChild")
	}
	ormObj, err := in.ToORM(ctx)
	if err != nil {
		return nil, err
	}
	if ormObj.Id == "" {
		return nil, errors.New("DefaultReadExternalChild requires a non-zero primary key")
	}
	if hook, ok := interface{}(&ormObj).(ExternalChildORMWithBeforeReadApplyQuery); ok {
		if db, err = hook.BeforeReadApplyQuery(ctx, db); err != nil {
			return nil, err
		}
	}
	if db, err = gorm2.ApplyFieldSelection(ctx, db, nil, &ExternalChildORM{}); err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(ExternalChildORMWithBeforeReadFind); ok {
		if db, err = hook.BeforeReadFind(ctx, db); err != nil {
			return nil, err
		}
	}
	ormResponse := ExternalChildORM{}
	if err = db.Where(&ormObj).First(&ormResponse).Error; err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormResponse).(ExternalChildORMWithAfterReadFind); ok {
		if err = hook.AfterReadFind(ctx, db); err != nil {
			return nil, err
		}
	}
	pbResponse, err := ormResponse.ToPB(ctx)
	return &pbResponse, err
}

type ExternalChildORMWithBeforeReadApplyQuery interface {
	BeforeReadApplyQuery(context.Context, *gorm1.DB) (*gorm1.DB, error)
}
type ExternalChildORMWithBeforeReadFind interface {
	BeforeReadFind(context.Context, *gorm1.DB) (*gorm1.DB, error)
}
type ExternalChildORMWithAfterReadFind interface {
	AfterReadFind(context.Context, *gorm1.DB) error
}

func DefaultDeleteExternalChild(ctx context.Context, in *ExternalChild, db *gorm1.DB) error {
	if in == nil {
		return errors.New("Nil argument to DefaultDeleteExternalChild")
	}
	ormObj, err := in.ToORM(ctx)
	if err != nil {
		return err
	}
	if ormObj.Id == "" {
		return errors.New("A non-zero ID value is required for a delete call")
	}
	if hook, ok := interface{}(&ormObj).(ExternalChildORMWithBeforeDelete); ok {
		if db, err = hook.BeforeDelete(ctx, db); err != nil {
			return err
		}
	}
	err = db.Where(&ormObj).Delete(&ExternalChildORM{}).Error
	if err != nil {
		return err
	}
	if hook, ok := interface{}(&ormObj).(ExternalChildORMWithAfterDelete); ok {
		err = hook.AfterDelete(ctx, db)
	}
	return err
}

type ExternalChildORMWithBeforeDelete interface {
	BeforeDelete(context.Context, *gorm1.DB) (*gorm1.DB, error)
}
type ExternalChildORMWithAfterDelete interface {
	AfterDelete(context.Context, *gorm1.DB) error
}

// DefaultStrictUpdateExternalChild clears first level 1:many children and then executes a gorm update call
func DefaultStrictUpdateExternalChild(ctx context.Context, in *ExternalChild, db *gorm1.DB) (*ExternalChild, error) {
	if in == nil {
		return nil, fmt.Errorf("Nil argument to DefaultStrictUpdateExternalChild")
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
	if hook, ok := interface{}(&ormObj).(ExternalChildORMWithBeforeStrictUpdateCleanup); ok {
		if db, err = hook.BeforeStrictUpdateCleanup(ctx, db); err != nil {
			return nil, err
		}
	}
	if hook, ok := interface{}(&ormObj).(ExternalChildORMWithBeforeStrictUpdateSave); ok {
		if db, err = hook.BeforeStrictUpdateSave(ctx, db); err != nil {
			return nil, err
		}
	}
	if err = db.Save(&ormObj).Error; err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(ExternalChildORMWithAfterStrictUpdateSave); ok {
		if err = hook.AfterStrictUpdateSave(ctx, db); err != nil {
			return nil, err
		}
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

type ExternalChildORMWithBeforeStrictUpdateCleanup interface {
	BeforeStrictUpdateCleanup(context.Context, *gorm1.DB) (*gorm1.DB, error)
}
type ExternalChildORMWithBeforeStrictUpdateSave interface {
	BeforeStrictUpdateSave(context.Context, *gorm1.DB) (*gorm1.DB, error)
}
type ExternalChildORMWithAfterStrictUpdateSave interface {
	AfterStrictUpdateSave(context.Context, *gorm1.DB) error
}

// DefaultPatchExternalChild executes a basic gorm update call with patch behavior
func DefaultPatchExternalChild(ctx context.Context, in *ExternalChild, updateMask *field_mask1.FieldMask, db *gorm1.DB) (*ExternalChild, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultPatchExternalChild")
	}
	var pbObj ExternalChild
	var err error
	if hook, ok := interface{}(&pbObj).(ExternalChildWithBeforePatchRead); ok {
		if db, err = hook.BeforePatchRead(ctx, in, updateMask, db); err != nil {
			return nil, err
		}
	}
	pbReadRes, err := DefaultReadExternalChild(ctx, &ExternalChild{Id: in.GetId()}, db)
	if err != nil {
		return nil, err
	}
	pbObj = *pbReadRes
	if hook, ok := interface{}(&pbObj).(ExternalChildWithBeforePatchApplyFieldMask); ok {
		if db, err = hook.BeforePatchApplyFieldMask(ctx, in, updateMask, db); err != nil {
			return nil, err
		}
	}
	if _, err := DefaultApplyFieldMaskExternalChild(ctx, &pbObj, in, updateMask, "", db); err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&pbObj).(ExternalChildWithBeforePatchSave); ok {
		if db, err = hook.BeforePatchSave(ctx, in, updateMask, db); err != nil {
			return nil, err
		}
	}
	pbResponse, err := DefaultStrictUpdateExternalChild(ctx, &pbObj, db)
	if err != nil {
		return nil, err
	}
	if hook, ok := interface{}(pbResponse).(ExternalChildWithAfterPatchSave); ok {
		if err = hook.AfterPatchSave(ctx, in, updateMask, db); err != nil {
			return nil, err
		}
	}
	return pbResponse, nil
}

type ExternalChildWithBeforePatchRead interface {
	BeforePatchRead(context.Context, *ExternalChild, *field_mask1.FieldMask, *gorm1.DB) (*gorm1.DB, error)
}
type ExternalChildWithBeforePatchApplyFieldMask interface {
	BeforePatchApplyFieldMask(context.Context, *ExternalChild, *field_mask1.FieldMask, *gorm1.DB) (*gorm1.DB, error)
}
type ExternalChildWithBeforePatchSave interface {
	BeforePatchSave(context.Context, *ExternalChild, *field_mask1.FieldMask, *gorm1.DB) (*gorm1.DB, error)
}
type ExternalChildWithAfterPatchSave interface {
	AfterPatchSave(context.Context, *ExternalChild, *field_mask1.FieldMask, *gorm1.DB) error
}

// DefaultApplyFieldMaskExternalChild patches an pbObject with patcher according to a field mask.
func DefaultApplyFieldMaskExternalChild(ctx context.Context, patchee *ExternalChild, patcher *ExternalChild, updateMask *field_mask1.FieldMask, prefix string, db *gorm1.DB) (*ExternalChild, error) {
	if patcher == nil {
		return nil, nil
	} else if patchee == nil {
		return nil, errors.New("Patchee inputs to DefaultApplyFieldMaskExternalChild must be non-nil")
	}
	var err error
	for _, f := range updateMask.Paths {
		if f == prefix+"Id" {
			patchee.Id = patcher.Id
			continue
		}
	}
	if err != nil {
		return nil, err
	}
	return patchee, nil
}

// DefaultListExternalChild executes a gorm list call
func DefaultListExternalChild(ctx context.Context, db *gorm1.DB) ([]*ExternalChild, error) {
	in := ExternalChild{}
	ormObj, err := in.ToORM(ctx)
	if err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(ExternalChildORMWithBeforeListApplyQuery); ok {
		if db, err = hook.BeforeListApplyQuery(ctx, db); err != nil {
			return nil, err
		}
	}
	db, err = gorm2.ApplyCollectionOperators(ctx, db, &ExternalChildORM{}, &ExternalChild{}, nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(ExternalChildORMWithBeforeListFind); ok {
		if db, err = hook.BeforeListFind(ctx, db); err != nil {
			return nil, err
		}
	}
	db = db.Where(&ormObj)
	db = db.Order("id")
	ormResponse := []ExternalChildORM{}
	if err := db.Find(&ormResponse).Error; err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(ExternalChildORMWithAfterListFind); ok {
		if err = hook.AfterListFind(ctx, db, &ormResponse); err != nil {
			return nil, err
		}
	}
	pbResponse := []*ExternalChild{}
	for _, responseEntry := range ormResponse {
		temp, err := responseEntry.ToPB(ctx)
		if err != nil {
			return nil, err
		}
		pbResponse = append(pbResponse, &temp)
	}
	return pbResponse, nil
}

type ExternalChildORMWithBeforeListApplyQuery interface {
	BeforeListApplyQuery(context.Context, *gorm1.DB) (*gorm1.DB, error)
}
type ExternalChildORMWithBeforeListFind interface {
	BeforeListFind(context.Context, *gorm1.DB) (*gorm1.DB, error)
}
type ExternalChildORMWithAfterListFind interface {
	AfterListFind(context.Context, *gorm1.DB, *[]ExternalChildORM) error
}
