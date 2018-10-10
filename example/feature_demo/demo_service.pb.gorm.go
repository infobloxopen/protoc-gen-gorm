// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: example/feature_demo/demo_service.proto

package example

import context "context"
import errors "errors"

import field_mask1 "google.golang.org/genproto/protobuf/field_mask"
import gateway1 "github.com/infobloxopen/atlas-app-toolkit/gateway"
import gorm1 "github.com/jinzhu/gorm"
import gorm2 "github.com/infobloxopen/atlas-app-toolkit/gorm"
import query1 "github.com/infobloxopen/atlas-app-toolkit/query"

import fmt "fmt"
import math "math"
import google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
import _ "google.golang.org/genproto/protobuf/field_mask"
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

type SomethingORM struct {
	Field string
}

// TableName overrides the default tablename generated by GORM
func (SomethingORM) TableName() string {
	return "somethings"
}

// ToORM runs the BeforeToORM hook if present, converts the fields of this
// object to ORM format, runs the AfterToORM hook, then returns the ORM object
func (m *Something) ToORM(ctx context.Context) (SomethingORM, error) {
	to := SomethingORM{}
	var err error
	if prehook, ok := interface{}(m).(SomethingWithBeforeToORM); ok {
		if err = prehook.BeforeToORM(ctx, &to); err != nil {
			return to, err
		}
	}
	to.Field = m.Field
	if posthook, ok := interface{}(m).(SomethingWithAfterToORM); ok {
		err = posthook.AfterToORM(ctx, &to)
	}
	return to, err
}

// ToPB runs the BeforeToPB hook if present, converts the fields of this
// object to PB format, runs the AfterToPB hook, then returns the PB object
func (m *SomethingORM) ToPB(ctx context.Context) (Something, error) {
	to := Something{}
	var err error
	if prehook, ok := interface{}(m).(SomethingWithBeforeToPB); ok {
		if err = prehook.BeforeToPB(ctx, &to); err != nil {
			return to, err
		}
	}
	to.Field = m.Field
	if posthook, ok := interface{}(m).(SomethingWithAfterToPB); ok {
		err = posthook.AfterToPB(ctx, &to)
	}
	return to, err
}

// The following are interfaces you can implement for special behavior during ORM/PB conversions
// of type Something the arg will be the target, the caller the one being converted from

// SomethingBeforeToORM called before default ToORM code
type SomethingWithBeforeToORM interface {
	BeforeToORM(context.Context, *SomethingORM) error
}

// SomethingAfterToORM called after default ToORM code
type SomethingWithAfterToORM interface {
	AfterToORM(context.Context, *SomethingORM) error
}

// SomethingBeforeToPB called before default ToPB code
type SomethingWithBeforeToPB interface {
	BeforeToPB(context.Context, *Something) error
}

// SomethingAfterToPB called after default ToPB code
type SomethingWithAfterToPB interface {
	AfterToPB(context.Context, *Something) error
}

type CircleORM struct {
	R uint32
}

// TableName overrides the default tablename generated by GORM
func (CircleORM) TableName() string {
	return "circles"
}

// ToORM runs the BeforeToORM hook if present, converts the fields of this
// object to ORM format, runs the AfterToORM hook, then returns the ORM object
func (m *Circle) ToORM(ctx context.Context) (CircleORM, error) {
	to := CircleORM{}
	var err error
	if prehook, ok := interface{}(m).(CircleWithBeforeToORM); ok {
		if err = prehook.BeforeToORM(ctx, &to); err != nil {
			return to, err
		}
	}
	to.R = m.R
	if posthook, ok := interface{}(m).(CircleWithAfterToORM); ok {
		err = posthook.AfterToORM(ctx, &to)
	}
	return to, err
}

// ToPB runs the BeforeToPB hook if present, converts the fields of this
// object to PB format, runs the AfterToPB hook, then returns the PB object
func (m *CircleORM) ToPB(ctx context.Context) (Circle, error) {
	to := Circle{}
	var err error
	if prehook, ok := interface{}(m).(CircleWithBeforeToPB); ok {
		if err = prehook.BeforeToPB(ctx, &to); err != nil {
			return to, err
		}
	}
	to.R = m.R
	if posthook, ok := interface{}(m).(CircleWithAfterToPB); ok {
		err = posthook.AfterToPB(ctx, &to)
	}
	return to, err
}

// The following are interfaces you can implement for special behavior during ORM/PB conversions
// of type Circle the arg will be the target, the caller the one being converted from

// CircleBeforeToORM called before default ToORM code
type CircleWithBeforeToORM interface {
	BeforeToORM(context.Context, *CircleORM) error
}

// CircleAfterToORM called after default ToORM code
type CircleWithAfterToORM interface {
	AfterToORM(context.Context, *CircleORM) error
}

// CircleBeforeToPB called before default ToPB code
type CircleWithBeforeToPB interface {
	BeforeToPB(context.Context, *Circle) error
}

// CircleAfterToPB called after default ToPB code
type CircleWithAfterToPB interface {
	AfterToPB(context.Context, *Circle) error
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
	if hook, ok := interface{}(&ormObj).(IntPointORMWithBeforeCreate); ok {
		if db, err = hook.BeforeCreate(ctx, db); err != nil {
			return nil, err
		}
	}
	if err = db.Create(&ormObj).Error; err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(IntPointORMWithAfterCreate); ok {
		if err = hook.AfterCreate(ctx, db); err != nil {
			return nil, err
		}
	}
	pbResponse, err := ormObj.ToPB(ctx)
	return &pbResponse, err
}

type IntPointORMWithBeforeCreate interface {
	BeforeCreate(context.Context, *gorm1.DB) (*gorm1.DB, error)
}
type IntPointORMWithAfterCreate interface {
	AfterCreate(context.Context, *gorm1.DB) error
}

// DefaultReadIntPoint executes a basic gorm read call
func DefaultReadIntPoint(ctx context.Context, in *IntPoint, db *gorm1.DB, fs *query1.FieldSelection) (*IntPoint, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultReadIntPoint")
	}
	ormObj, err := in.ToORM(ctx)
	if err != nil {
		return nil, err
	}
	if ormObj.Id == 0 {
		return nil, errors.New("DefaultReadIntPoint requires a non-zero primary key")
	}
	if hook, ok := interface{}(&ormObj).(IntPointORMWithBeforeReadApplyQuery); ok {
		if db, err = hook.BeforeReadApplyQuery(ctx, db, fs); err != nil {
			return nil, err
		}
	}
	if db, err = gorm2.ApplyFieldSelection(ctx, db, fs, &IntPointORM{}); err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(IntPointORMWithBeforeReadFind); ok {
		if db, err = hook.BeforeReadFind(ctx, db, fs); err != nil {
			return nil, err
		}
	}
	ormResponse := IntPointORM{}
	if err = db.Where(&ormObj).First(&ormResponse).Error; err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormResponse).(IntPointORMWithAfterReadFind); ok {
		if err = hook.AfterReadFind(ctx, db, fs); err != nil {
			return nil, err
		}
	}
	pbResponse, err := ormResponse.ToPB(ctx)
	return &pbResponse, err
}

type IntPointORMWithBeforeReadApplyQuery interface {
	BeforeReadApplyQuery(context.Context, *gorm1.DB, *query1.FieldSelection) (*gorm1.DB, error)
}
type IntPointORMWithBeforeReadFind interface {
	BeforeReadFind(context.Context, *gorm1.DB, *query1.FieldSelection) (*gorm1.DB, error)
}
type IntPointORMWithAfterReadFind interface {
	AfterReadFind(context.Context, *gorm1.DB, *query1.FieldSelection) error
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
	if hook, ok := interface{}(&ormObj).(IntPointORMWithBeforeDelete); ok {
		if db, err = hook.BeforeDelete(ctx, db); err != nil {
			return err
		}
	}
	err = db.Where(&ormObj).Delete(&IntPointORM{}).Error
	if err != nil {
		return err
	}
	if hook, ok := interface{}(&ormObj).(IntPointORMWithAfterDelete); ok {
		err = hook.AfterDelete(ctx, db)
	}
	return err
}

type IntPointORMWithBeforeDelete interface {
	BeforeDelete(context.Context, *gorm1.DB) (*gorm1.DB, error)
}
type IntPointORMWithAfterDelete interface {
	AfterDelete(context.Context, *gorm1.DB) error
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
	if hook, ok := interface{}(&ormObj).(IntPointORMWithBeforeStrictUpdateSave); ok {
		if db, err = hook.BeforeStrictUpdateSave(ctx, db); err != nil {
			return nil, err
		}
	}
	if err = db.Set("gorm:association_autoupdate", false).Set("gorm:association_autocreate", false).Set("gorm:association_save_reference", false).Save(&ormObj).Error; err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(IntPointORMWithBeforeStrictUpdateCleanup); ok {
		if db, err = hook.BeforeStrictUpdateCleanup(ctx, db); err != nil {
			return nil, err
		}
	}
	if err = db.Save(&ormObj).Error; err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(IntPointORMWithAfterStrictUpdateSave); ok {
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

type IntPointORMWithBeforeStrictUpdateCleanup interface {
	BeforeStrictUpdateCleanup(context.Context, *gorm1.DB) (*gorm1.DB, error)
}
type IntPointORMWithBeforeStrictUpdateSave interface {
	BeforeStrictUpdateSave(context.Context, *gorm1.DB) (*gorm1.DB, error)
}
type IntPointORMWithAfterStrictUpdateSave interface {
	AfterStrictUpdateSave(context.Context, *gorm1.DB) error
}

// DefaultPatchIntPoint executes a basic gorm update call with patch behavior
func DefaultPatchIntPoint(ctx context.Context, in *IntPoint, updateMask *field_mask1.FieldMask, db *gorm1.DB) (*IntPoint, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultPatchIntPoint")
	}
	var pbObj IntPoint
	var err error
	if hook, ok := interface{}(&pbObj).(IntPointWithBeforePatchRead); ok {
		if db, err = hook.BeforePatchRead(ctx, in, updateMask, db); err != nil {
			return nil, err
		}
	}
	pbReadRes, err := DefaultReadIntPoint(ctx, &IntPoint{Id: in.GetId()}, db, nil)
	if err != nil {
		return nil, err
	}
	pbObj = *pbReadRes
	if hook, ok := interface{}(&pbObj).(IntPointWithBeforePatchApplyFieldMask); ok {
		if db, err = hook.BeforePatchApplyFieldMask(ctx, in, updateMask, db); err != nil {
			return nil, err
		}
	}
	if _, err := DefaultApplyFieldMaskIntPoint(ctx, &pbObj, in, updateMask, "", db); err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&pbObj).(IntPointWithBeforePatchSave); ok {
		if db, err = hook.BeforePatchSave(ctx, in, updateMask, db); err != nil {
			return nil, err
		}
	}
	pbResponse, err := DefaultStrictUpdateIntPoint(ctx, &pbObj, db)
	if err != nil {
		return nil, err
	}
	if hook, ok := interface{}(pbResponse).(IntPointWithAfterPatchSave); ok {
		if err = hook.AfterPatchSave(ctx, in, updateMask, db); err != nil {
			return nil, err
		}
	}
	return pbResponse, nil
}

type IntPointWithBeforePatchRead interface {
	BeforePatchRead(context.Context, *IntPoint, *field_mask1.FieldMask, *gorm1.DB) (*gorm1.DB, error)
}
type IntPointWithBeforePatchApplyFieldMask interface {
	BeforePatchApplyFieldMask(context.Context, *IntPoint, *field_mask1.FieldMask, *gorm1.DB) (*gorm1.DB, error)
}
type IntPointWithBeforePatchSave interface {
	BeforePatchSave(context.Context, *IntPoint, *field_mask1.FieldMask, *gorm1.DB) (*gorm1.DB, error)
}
type IntPointWithAfterPatchSave interface {
	AfterPatchSave(context.Context, *IntPoint, *field_mask1.FieldMask, *gorm1.DB) error
}

// DefaultApplyFieldMaskIntPoint patches an pbObject with patcher according to a field mask.
func DefaultApplyFieldMaskIntPoint(ctx context.Context, patchee *IntPoint, patcher *IntPoint, updateMask *field_mask1.FieldMask, prefix string, db *gorm1.DB) (*IntPoint, error) {
	if patcher == nil {
		return nil, nil
	} else if patchee == nil {
		return nil, errors.New("Patchee inputs to DefaultApplyFieldMaskIntPoint must be non-nil")
	}
	var err error
	for _, f := range updateMask.Paths {
		if f == prefix+"Id" {
			patchee.Id = patcher.Id
			continue
		}
		if f == prefix+"X" {
			patchee.X = patcher.X
			continue
		}
		if f == prefix+"Y" {
			patchee.Y = patcher.Y
			continue
		}
	}
	if err != nil {
		return nil, err
	}
	return patchee, nil
}

// DefaultListIntPoint executes a gorm list call
func DefaultListIntPoint(ctx context.Context, db *gorm1.DB, f *query1.Filtering, s *query1.Sorting, p *query1.Pagination, fs *query1.FieldSelection) ([]*IntPoint, error) {
	in := IntPoint{}
	ormObj, err := in.ToORM(ctx)
	if err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(IntPointORMWithBeforeListApplyQuery); ok {
		if db, err = hook.BeforeListApplyQuery(ctx, db, f, s, p, fs); err != nil {
			return nil, err
		}
	}
	db, err = gorm2.ApplyCollectionOperators(ctx, db, &IntPointORM{}, &IntPoint{}, f, s, p, fs)
	if err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(IntPointORMWithBeforeListFind); ok {
		if db, err = hook.BeforeListFind(ctx, db, f, s, p, fs); err != nil {
			return nil, err
		}
	}
	db = db.Order("id")
	ormResponse := []IntPointORM{}
	if err := db.Find(&ormResponse).Error; err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(IntPointORMWithAfterListFind); ok {
		if err = hook.AfterListFind(ctx, db, &ormResponse, f, s, p, fs); err != nil {
			return nil, err
		}
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

type IntPointORMWithBeforeListApplyQuery interface {
	BeforeListApplyQuery(context.Context, *gorm1.DB, *query1.Filtering, *query1.Sorting, *query1.Pagination, *query1.FieldSelection) (*gorm1.DB, error)
}
type IntPointORMWithBeforeListFind interface {
	BeforeListFind(context.Context, *gorm1.DB, *query1.Filtering, *query1.Sorting, *query1.Pagination, *query1.FieldSelection) (*gorm1.DB, error)
}
type IntPointORMWithAfterListFind interface {
	AfterListFind(context.Context, *gorm1.DB, *[]IntPointORM, *query1.Filtering, *query1.Sorting, *query1.Pagination, *query1.FieldSelection) error
}

// DefaultCreateSomething executes a basic gorm create call
func DefaultCreateSomething(ctx context.Context, in *Something, db *gorm1.DB) (*Something, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultCreateSomething")
	}
	ormObj, err := in.ToORM(ctx)
	if err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(SomethingORMWithBeforeCreate); ok {
		if db, err = hook.BeforeCreate(ctx, db); err != nil {
			return nil, err
		}
	}
	if err = db.Create(&ormObj).Error; err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(SomethingORMWithAfterCreate); ok {
		if err = hook.AfterCreate(ctx, db); err != nil {
			return nil, err
		}
	}
	pbResponse, err := ormObj.ToPB(ctx)
	return &pbResponse, err
}

type SomethingORMWithBeforeCreate interface {
	BeforeCreate(context.Context, *gorm1.DB) (*gorm1.DB, error)
}
type SomethingORMWithAfterCreate interface {
	AfterCreate(context.Context, *gorm1.DB) error
}

// DefaultApplyFieldMaskSomething patches an pbObject with patcher according to a field mask.
func DefaultApplyFieldMaskSomething(ctx context.Context, patchee *Something, patcher *Something, updateMask *field_mask1.FieldMask, prefix string, db *gorm1.DB) (*Something, error) {
	if patcher == nil {
		return nil, nil
	} else if patchee == nil {
		return nil, errors.New("Patchee inputs to DefaultApplyFieldMaskSomething must be non-nil")
	}
	var err error
	for _, f := range updateMask.Paths {
		if f == prefix+"Field" {
			patchee.Field = patcher.Field
			continue
		}
	}
	if err != nil {
		return nil, err
	}
	return patchee, nil
}

// DefaultListSomething executes a gorm list call
func DefaultListSomething(ctx context.Context, db *gorm1.DB) ([]*Something, error) {
	in := Something{}
	ormObj, err := in.ToORM(ctx)
	if err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(SomethingORMWithBeforeListApplyQuery); ok {
		if db, err = hook.BeforeListApplyQuery(ctx, db); err != nil {
			return nil, err
		}
	}
	db, err = gorm2.ApplyCollectionOperators(ctx, db, &SomethingORM{}, &Something{}, nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(SomethingORMWithBeforeListFind); ok {
		if db, err = hook.BeforeListFind(ctx, db); err != nil {
			return nil, err
		}
	}
	ormResponse := []SomethingORM{}
	if err := db.Find(&ormResponse).Error; err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(SomethingORMWithAfterListFind); ok {
		if err = hook.AfterListFind(ctx, db, &ormResponse); err != nil {
			return nil, err
		}
	}
	pbResponse := []*Something{}
	for _, responseEntry := range ormResponse {
		temp, err := responseEntry.ToPB(ctx)
		if err != nil {
			return nil, err
		}
		pbResponse = append(pbResponse, &temp)
	}
	return pbResponse, nil
}

type SomethingORMWithBeforeListApplyQuery interface {
	BeforeListApplyQuery(context.Context, *gorm1.DB) (*gorm1.DB, error)
}
type SomethingORMWithBeforeListFind interface {
	BeforeListFind(context.Context, *gorm1.DB) (*gorm1.DB, error)
}
type SomethingORMWithAfterListFind interface {
	AfterListFind(context.Context, *gorm1.DB, *[]SomethingORM) error
}

// DefaultCreateCircle executes a basic gorm create call
func DefaultCreateCircle(ctx context.Context, in *Circle, db *gorm1.DB) (*Circle, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultCreateCircle")
	}
	ormObj, err := in.ToORM(ctx)
	if err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(CircleORMWithBeforeCreate); ok {
		if db, err = hook.BeforeCreate(ctx, db); err != nil {
			return nil, err
		}
	}
	if err = db.Create(&ormObj).Error; err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(CircleORMWithAfterCreate); ok {
		if err = hook.AfterCreate(ctx, db); err != nil {
			return nil, err
		}
	}
	pbResponse, err := ormObj.ToPB(ctx)
	return &pbResponse, err
}

type CircleORMWithBeforeCreate interface {
	BeforeCreate(context.Context, *gorm1.DB) (*gorm1.DB, error)
}
type CircleORMWithAfterCreate interface {
	AfterCreate(context.Context, *gorm1.DB) error
}

// DefaultApplyFieldMaskCircle patches an pbObject with patcher according to a field mask.
func DefaultApplyFieldMaskCircle(ctx context.Context, patchee *Circle, patcher *Circle, updateMask *field_mask1.FieldMask, prefix string, db *gorm1.DB) (*Circle, error) {
	if patcher == nil {
		return nil, nil
	} else if patchee == nil {
		return nil, errors.New("Patchee inputs to DefaultApplyFieldMaskCircle must be non-nil")
	}
	var err error
	for _, f := range updateMask.Paths {
		if f == prefix+"R" {
			patchee.R = patcher.R
			continue
		}
	}
	if err != nil {
		return nil, err
	}
	return patchee, nil
}

// DefaultListCircle executes a gorm list call
func DefaultListCircle(ctx context.Context, db *gorm1.DB) ([]*Circle, error) {
	in := Circle{}
	ormObj, err := in.ToORM(ctx)
	if err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(CircleORMWithBeforeListApplyQuery); ok {
		if db, err = hook.BeforeListApplyQuery(ctx, db); err != nil {
			return nil, err
		}
	}
	db, err = gorm2.ApplyCollectionOperators(ctx, db, &CircleORM{}, &Circle{}, nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(CircleORMWithBeforeListFind); ok {
		if db, err = hook.BeforeListFind(ctx, db); err != nil {
			return nil, err
		}
	}
	ormResponse := []CircleORM{}
	if err := db.Find(&ormResponse).Error; err != nil {
		return nil, err
	}
	if hook, ok := interface{}(&ormObj).(CircleORMWithAfterListFind); ok {
		if err = hook.AfterListFind(ctx, db, &ormResponse); err != nil {
			return nil, err
		}
	}
	pbResponse := []*Circle{}
	for _, responseEntry := range ormResponse {
		temp, err := responseEntry.ToPB(ctx)
		if err != nil {
			return nil, err
		}
		pbResponse = append(pbResponse, &temp)
	}
	return pbResponse, nil
}

type CircleORMWithBeforeListApplyQuery interface {
	BeforeListApplyQuery(context.Context, *gorm1.DB) (*gorm1.DB, error)
}
type CircleORMWithBeforeListFind interface {
	BeforeListFind(context.Context, *gorm1.DB) (*gorm1.DB, error)
}
type CircleORMWithAfterListFind interface {
	AfterListFind(context.Context, *gorm1.DB, *[]CircleORM) error
}
type IntPointServiceDefaultServer struct {
	DB *gorm1.DB
}

// Create ...
func (m *IntPointServiceDefaultServer) Create(ctx context.Context, in *CreateIntPointRequest) (*CreateIntPointResponse, error) {
	db := m.DB
	if custom, ok := interface{}(in).(IntPointServiceIntPointWithBeforeCreate); ok {
		var err error
		if db, err = custom.BeforeCreate(ctx, db); err != nil {
			return nil, err
		}
	}
	res, err := DefaultCreateIntPoint(ctx, in.GetPayload(), db)
	if err != nil {
		return nil, err
	}
	out := &CreateIntPointResponse{Result: res}
	if custom, ok := interface{}(in).(IntPointServiceIntPointWithAfterCreate); ok {
		var err error
		if err = custom.AfterCreate(ctx, out, db); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// IntPointServiceIntPointWithBeforeCreate called before DefaultCreateIntPoint in the default Create handler
type IntPointServiceIntPointWithBeforeCreate interface {
	BeforeCreate(context.Context, *gorm1.DB) (*gorm1.DB, error)
}

// IntPointServiceIntPointWithAfterCreate called before DefaultCreateIntPoint in the default Create handler
type IntPointServiceIntPointWithAfterCreate interface {
	AfterCreate(context.Context, *CreateIntPointResponse, *gorm1.DB) error
}

// Read ...
func (m *IntPointServiceDefaultServer) Read(ctx context.Context, in *ReadIntPointRequest) (*ReadIntPointResponse, error) {
	db := m.DB
	if custom, ok := interface{}(in).(IntPointServiceIntPointWithBeforeRead); ok {
		var err error
		if db, err = custom.BeforeRead(ctx, db); err != nil {
			return nil, err
		}
	}
	res, err := DefaultReadIntPoint(ctx, &IntPoint{Id: in.GetId()}, db, in.Fields)
	if err != nil {
		return nil, err
	}
	out := &ReadIntPointResponse{Result: res}
	if custom, ok := interface{}(in).(IntPointServiceIntPointWithAfterRead); ok {
		var err error
		if err = custom.AfterRead(ctx, out, db); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// IntPointServiceIntPointWithBeforeRead called before DefaultReadIntPoint in the default Read handler
type IntPointServiceIntPointWithBeforeRead interface {
	BeforeRead(context.Context, *gorm1.DB) (*gorm1.DB, error)
}

// IntPointServiceIntPointWithAfterRead called before DefaultReadIntPoint in the default Read handler
type IntPointServiceIntPointWithAfterRead interface {
	AfterRead(context.Context, *ReadIntPointResponse, *gorm1.DB) error
}

// Update ...
func (m *IntPointServiceDefaultServer) Update(ctx context.Context, in *UpdateIntPointRequest) (*UpdateIntPointResponse, error) {
	var err error
	var res *IntPoint
	db := m.DB
	if custom, ok := interface{}(in).(IntPointServiceIntPointWithBeforeUpdate); ok {
		var err error
		if db, err = custom.BeforeUpdate(ctx, db); err != nil {
			return nil, err
		}
	}
	if in.GetGerogeriGegege() == nil {
		res, err = DefaultStrictUpdateIntPoint(ctx, in.GetPayload(), db)
	} else {
		res, err = DefaultPatchIntPoint(ctx, in.GetPayload(), in.GetGerogeriGegege(), db)
	}
	if err != nil {
		return nil, err
	}
	out := &UpdateIntPointResponse{Result: res}
	if custom, ok := interface{}(in).(IntPointServiceIntPointWithAfterUpdate); ok {
		var err error
		if err = custom.AfterUpdate(ctx, out, db); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// IntPointServiceIntPointWithBeforeUpdate called before DefaultUpdateIntPoint in the default Update handler
type IntPointServiceIntPointWithBeforeUpdate interface {
	BeforeUpdate(context.Context, *gorm1.DB) (*gorm1.DB, error)
}

// IntPointServiceIntPointWithAfterUpdate called before DefaultUpdateIntPoint in the default Update handler
type IntPointServiceIntPointWithAfterUpdate interface {
	AfterUpdate(context.Context, *UpdateIntPointResponse, *gorm1.DB) error
}

// List ...
func (m *IntPointServiceDefaultServer) List(ctx context.Context, in *ListIntPointRequest) (*ListIntPointResponse, error) {
	db := m.DB
	if custom, ok := interface{}(in).(IntPointServiceIntPointWithBeforeList); ok {
		var err error
		if db, err = custom.BeforeList(ctx, db); err != nil {
			return nil, err
		}
	}
	pagedRequest := false
	if in.GetPaging().GetLimit() >= 1 {
		in.Paging.Limit++
		pagedRequest = true
	}
	res, err := DefaultListIntPoint(ctx, db, in.Filter, in.OrderBy, in.Paging, in.Fields)
	if err != nil {
		return nil, err
	}
	var resPaging *query1.PageInfo
	if pagedRequest {
		var offset int32
		var size int32 = int32(len(res))
		if size == in.GetPaging().GetLimit() {
			size--
			res = res[:size]
			offset = in.GetPaging().GetOffset() + size
		}
		resPaging = &query1.PageInfo{Offset: offset}
	}
	out := &ListIntPointResponse{Results: res, PageInfo: resPaging}
	if custom, ok := interface{}(in).(IntPointServiceIntPointWithAfterList); ok {
		var err error
		if err = custom.AfterList(ctx, out, db); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// IntPointServiceIntPointWithBeforeList called before DefaultListIntPoint in the default List handler
type IntPointServiceIntPointWithBeforeList interface {
	BeforeList(context.Context, *gorm1.DB) (*gorm1.DB, error)
}

// IntPointServiceIntPointWithAfterList called before DefaultListIntPoint in the default List handler
type IntPointServiceIntPointWithAfterList interface {
	AfterList(context.Context, *ListIntPointResponse, *gorm1.DB) error
}

// ListSomething ...
func (m *IntPointServiceDefaultServer) ListSomething(ctx context.Context, in *google_protobuf2.Empty) (*ListSomethingResponse, error) {
	return &ListSomethingResponse{}, nil
}

// Delete ...
func (m *IntPointServiceDefaultServer) Delete(ctx context.Context, in *DeleteIntPointRequest) (*DeleteIntPointResponse, error) {
	db := m.DB
	if custom, ok := interface{}(in).(IntPointServiceIntPointWithBeforeDelete); ok {
		var err error
		if db, err = custom.BeforeDelete(ctx, db); err != nil {
			return nil, err
		}
	}
	err := DefaultDeleteIntPoint(ctx, &IntPoint{Id: in.GetId()}, db)
	if err != nil {
		return nil, err
	}
	out := &DeleteIntPointResponse{}
	if custom, ok := interface{}(in).(IntPointServiceIntPointWithAfterDelete); ok {
		var err error
		if err = custom.AfterDelete(ctx, out, db); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// IntPointServiceIntPointWithBeforeDelete called before DefaultDeleteIntPoint in the default Delete handler
type IntPointServiceIntPointWithBeforeDelete interface {
	BeforeDelete(context.Context, *gorm1.DB) (*gorm1.DB, error)
}

// IntPointServiceIntPointWithAfterDelete called before DefaultDeleteIntPoint in the default Delete handler
type IntPointServiceIntPointWithAfterDelete interface {
	AfterDelete(context.Context, *DeleteIntPointResponse, *gorm1.DB) error
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
		if db, err = custom.BeforeCreate(ctx, db); err != nil {
			return nil, err
		}
	}
	res, err := DefaultCreateIntPoint(ctx, in.GetPayload(), db)
	if err != nil {
		return nil, err
	}
	out := &CreateIntPointResponse{Result: res}
	if custom, ok := interface{}(in).(IntPointTxnIntPointWithAfterCreate); ok {
		var err error
		if err = custom.AfterCreate(ctx, out, db); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// IntPointTxnIntPointWithBeforeCreate called before DefaultCreateIntPoint in the default Create handler
type IntPointTxnIntPointWithBeforeCreate interface {
	BeforeCreate(context.Context, *gorm1.DB) (*gorm1.DB, error)
}

// IntPointTxnIntPointWithAfterCreate called before DefaultCreateIntPoint in the default Create handler
type IntPointTxnIntPointWithAfterCreate interface {
	AfterCreate(context.Context, *CreateIntPointResponse, *gorm1.DB) error
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
		if db, err = custom.BeforeRead(ctx, db); err != nil {
			return nil, err
		}
	}
	res, err := DefaultReadIntPoint(ctx, &IntPoint{Id: in.GetId()}, db, in.Fields)
	if err != nil {
		return nil, err
	}
	out := &ReadIntPointResponse{Result: res}
	if custom, ok := interface{}(in).(IntPointTxnIntPointWithAfterRead); ok {
		var err error
		if err = custom.AfterRead(ctx, out, db); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// IntPointTxnIntPointWithBeforeRead called before DefaultReadIntPoint in the default Read handler
type IntPointTxnIntPointWithBeforeRead interface {
	BeforeRead(context.Context, *gorm1.DB) (*gorm1.DB, error)
}

// IntPointTxnIntPointWithAfterRead called before DefaultReadIntPoint in the default Read handler
type IntPointTxnIntPointWithAfterRead interface {
	AfterRead(context.Context, *ReadIntPointResponse, *gorm1.DB) error
}

// Update ...
func (m *IntPointTxnDefaultServer) Update(ctx context.Context, in *UpdateIntPointRequest) (*UpdateIntPointResponse, error) {
	var err error
	var res *IntPoint
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
		if db, err = custom.BeforeUpdate(ctx, db); err != nil {
			return nil, err
		}
	}
	if in.GetGerogeriGegege() == nil {
		res, err = DefaultStrictUpdateIntPoint(ctx, in.GetPayload(), db)
	} else {
		res, err = DefaultPatchIntPoint(ctx, in.GetPayload(), in.GetGerogeriGegege(), db)
	}
	if err != nil {
		return nil, err
	}
	out := &UpdateIntPointResponse{Result: res}
	if custom, ok := interface{}(in).(IntPointTxnIntPointWithAfterUpdate); ok {
		var err error
		if err = custom.AfterUpdate(ctx, out, db); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// IntPointTxnIntPointWithBeforeUpdate called before DefaultUpdateIntPoint in the default Update handler
type IntPointTxnIntPointWithBeforeUpdate interface {
	BeforeUpdate(context.Context, *gorm1.DB) (*gorm1.DB, error)
}

// IntPointTxnIntPointWithAfterUpdate called before DefaultUpdateIntPoint in the default Update handler
type IntPointTxnIntPointWithAfterUpdate interface {
	AfterUpdate(context.Context, *UpdateIntPointResponse, *gorm1.DB) error
}

// List ...
func (m *IntPointTxnDefaultServer) List(ctx context.Context, in *ListIntPointRequest) (*ListIntPointResponse, error) {
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
		if db, err = custom.BeforeList(ctx, db); err != nil {
			return nil, err
		}
	}
	pagedRequest := false
	if in.GetPaging().GetLimit() >= 1 {
		in.Paging.Limit++
		pagedRequest = true
	}
	res, err := DefaultListIntPoint(ctx, db, in.Filter, in.OrderBy, in.Paging, in.Fields)
	if err != nil {
		return nil, err
	}
	var resPaging *query1.PageInfo
	if pagedRequest {
		var offset int32
		var size int32 = int32(len(res))
		if size == in.GetPaging().GetLimit() {
			size--
			res = res[:size]
			offset = in.GetPaging().GetOffset() + size
		}
		resPaging = &query1.PageInfo{Offset: offset}
	}
	out := &ListIntPointResponse{Results: res, PageInfo: resPaging}
	if custom, ok := interface{}(in).(IntPointTxnIntPointWithAfterList); ok {
		var err error
		if err = custom.AfterList(ctx, out, db); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// IntPointTxnIntPointWithBeforeList called before DefaultListIntPoint in the default List handler
type IntPointTxnIntPointWithBeforeList interface {
	BeforeList(context.Context, *gorm1.DB) (*gorm1.DB, error)
}

// IntPointTxnIntPointWithAfterList called before DefaultListIntPoint in the default List handler
type IntPointTxnIntPointWithAfterList interface {
	AfterList(context.Context, *ListIntPointResponse, *gorm1.DB) error
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
		if db, err = custom.BeforeDelete(ctx, db); err != nil {
			return nil, err
		}
	}
	err := DefaultDeleteIntPoint(ctx, &IntPoint{Id: in.GetId()}, db)
	if err != nil {
		return nil, err
	}
	out := &DeleteIntPointResponse{}
	if custom, ok := interface{}(in).(IntPointTxnIntPointWithAfterDelete); ok {
		var err error
		if err = custom.AfterDelete(ctx, out, db); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// IntPointTxnIntPointWithBeforeDelete called before DefaultDeleteIntPoint in the default Delete handler
type IntPointTxnIntPointWithBeforeDelete interface {
	BeforeDelete(context.Context, *gorm1.DB) (*gorm1.DB, error)
}

// IntPointTxnIntPointWithAfterDelete called before DefaultDeleteIntPoint in the default Delete handler
type IntPointTxnIntPointWithAfterDelete interface {
	AfterDelete(context.Context, *DeleteIntPointResponse, *gorm1.DB) error
}

// CustomMethod ...
func (m *IntPointTxnDefaultServer) CustomMethod(ctx context.Context, in *google_protobuf2.Empty) (*google_protobuf2.Empty, error) {
	return &google_protobuf2.Empty{}, nil
}

// CreateSomething ...
func (m *IntPointTxnDefaultServer) CreateSomething(ctx context.Context, in *Something) (*Something, error) {
	return &Something{}, nil
}

type CircleServiceDefaultServer struct {
	DB *gorm1.DB
}

// List ...
func (m *CircleServiceDefaultServer) List(ctx context.Context, in *ListCircleRequest) (*ListCircleResponse, error) {
	db := m.DB
	if custom, ok := interface{}(in).(CircleServiceCircleWithBeforeList); ok {
		var err error
		if db, err = custom.BeforeList(ctx, db); err != nil {
			return nil, err
		}
	}
	res, err := DefaultListCircle(ctx, db)
	if err != nil {
		return nil, err
	}
	out := &ListCircleResponse{Results: res}
	if custom, ok := interface{}(in).(CircleServiceCircleWithAfterList); ok {
		var err error
		if err = custom.AfterList(ctx, out, db); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// CircleServiceCircleWithBeforeList called before DefaultListCircle in the default List handler
type CircleServiceCircleWithBeforeList interface {
	BeforeList(context.Context, *gorm1.DB) (*gorm1.DB, error)
}

// CircleServiceCircleWithAfterList called before DefaultListCircle in the default List handler
type CircleServiceCircleWithAfterList interface {
	AfterList(context.Context, *ListCircleResponse, *gorm1.DB) error
}
