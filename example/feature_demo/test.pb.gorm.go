// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: example/feature_demo/test.proto

package example

import context "context"
import errors "errors"
import gorm "github.com/jinzhu/gorm"
import ops "github.com/Infoblox-CTO/ngp.api.toolkit/op/gorm"
import uuid "github.com/satori/go.uuid"
import gtypes "github.com/infobloxopen/protoc-gen-gorm/types"
import time "time"
import ptypes "github.com/golang/protobuf/ptypes"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/infobloxopen/protoc-gen-gorm/types"
import google_protobuf1 "github.com/golang/protobuf/ptypes/wrappers"
import google_protobuf2 "github.com/golang/protobuf/ptypes/empty"
import _ "github.com/golang/protobuf/ptypes/timestamp"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// TestTypesORM is a message that serves as an example
type TestTypesORM struct {
	// Skipping field from proto option: ApiOnlyString
	// The non-ORMable repeated field "Numbers" can't be included
	OptionalString *string
	BecomesInt     int32
	// Empty type has no ORM equivalency
	UUID      uuid.UUID `sql:"type:uuid"`
	CreatedAt time.Time
}

// TableName overrides the default tablename generated by GORM
func (TestTypesORM) TableName() string {
	return "smorgasbord"
}

// ConvertTestTypesToORM takes a pb object and returns an orm object
func ConvertTestTypesToORM(from TestTypes) (TestTypesORM, error) {
	to := TestTypesORM{}
	var err error
	// Skipping field: ApiOnlyString
	// Repeated type []int32 is not an ORMable message type
	if from.OptionalString != nil {
		v := from.OptionalString.Value
		to.OptionalString = &v
	}
	to.BecomesInt = int32(from.BecomesInt)
	if from.Uuid != nil {
		if to.UUID, err = uuid.FromString(from.Uuid.Value); err != nil {
			return to, err
		}
	}
	if from.CreatedAt != nil {
		if to.CreatedAt, err = ptypes.Timestamp(from.CreatedAt); err != nil {
			return to, err
		}
	}
	return to, err
}

// ConvertTestTypesFromORM takes an orm object and returns a pb object
func ConvertTestTypesFromORM(from TestTypesORM) (TestTypes, error) {
	to := TestTypes{}
	var err error
	// Skipping field: ApiOnlyString
	// Repeated type []int32 is not an ORMable message type
	if from.OptionalString != nil {
		to.OptionalString = &google_protobuf1.StringValue{Value: *from.OptionalString}
	}
	to.BecomesInt = TestTypesStatus(from.BecomesInt)
	to.Uuid = &gtypes.UUIDValue{Value: from.UUID.String()}
	if to.CreatedAt, err = ptypes.TimestampProto(from.CreatedAt); err != nil {
		return to, err
	}
	return to, err
}

// TypeWithIDORM no comment was provided for message type
type TypeWithIDORM struct {
	UUID          int32           `gorm:"primary_key"`
	IP            string          `gorm:"ip_addr"`
	Things        []*TestTypesORM `gorm:"foreignkey:TypeWithIDID"`
	ANestedObject *TestTypesORM   `gorm:"foreignkey:TypeWithIDID"`
}

// TableName overrides the default tablename generated by GORM
func (TypeWithIDORM) TableName() string {
	return "type_with_ids"
}

// ConvertTypeWithIDToORM takes a pb object and returns an orm object
func ConvertTypeWithIDToORM(from TypeWithId) (TypeWithIDORM, error) {
	to := TypeWithIDORM{}
	var err error
	to.IP = from.Ip
	for _, v := range from.Things {
		if v != nil {
			if tempThings, cErr := ConvertTestTypesToORM(*v); cErr == nil {
				to.Things = append(to.Things, &tempThings)
			} else {
				return to, cErr
			}
		} else {
			to.Things = append(to.Things, nil)
		}
	}
	if from.ANestedObject != nil {
		if to.ANestedObject, err = ConvertTestTypesToORM(from.ANestedObject); err != nil {
			return to, err
		}
	}
	return to, err
}

// ConvertTypeWithIDFromORM takes an orm object and returns a pb object
func ConvertTypeWithIDFromORM(from TypeWithIDORM) (TypeWithId, error) {
	to := TypeWithId{}
	var err error
	to.Ip = from.IP
	for _, v := range from.Things {
		if v != nil {
			if tempThings, cErr := ConvertTestTypesFromORM(*v); cErr == nil {
				to.Things = append(to.Things, &tempThings)
			} else {
				return to, cErr
			}
		} else {
			to.Things = append(to.Things, nil)
		}
	}
	if from.ANestedObject != nil {
		if to.ANestedObject, err = ConvertTestTypesFromORM(from.ANestedObject); err != nil {
			return to, err
		}
	}
	return to, err
}

// MultitenantTypeWithIDORM no comment was provided for message type
type MultitenantTypeWithIDORM struct {
	TenantID  string
	ID        uint64
	SomeField string
}

// TableName overrides the default tablename generated by GORM
func (MultitenantTypeWithIDORM) TableName() string {
	return "multitenant_type_with_ids"
}

// ConvertMultitenantTypeWithIDToORM takes a pb object and returns an orm object
func ConvertMultitenantTypeWithIDToORM(from MultitenantTypeWithId) (MultitenantTypeWithIDORM, error) {
	to := MultitenantTypeWithIDORM{}
	var err error
	to.ID = from.Id
	to.SomeField = from.SomeField
	return to, err
}

// ConvertMultitenantTypeWithIDFromORM takes an orm object and returns a pb object
func ConvertMultitenantTypeWithIDFromORM(from MultitenantTypeWithIDORM) (MultitenantTypeWithId, error) {
	to := MultitenantTypeWithId{}
	var err error
	to.Id = from.ID
	to.SomeField = from.SomeField
	return to, err
}

// MultitenantTypeWithoutIDORM no comment was provided for message type
type MultitenantTypeWithoutIDORM struct {
	TenantID  string
	SomeField string
}

// TableName overrides the default tablename generated by GORM
func (MultitenantTypeWithoutIDORM) TableName() string {
	return "multitenant_type_without_ids"
}

// ConvertMultitenantTypeWithoutIDToORM takes a pb object and returns an orm object
func ConvertMultitenantTypeWithoutIDToORM(from MultitenantTypeWithoutId) (MultitenantTypeWithoutIDORM, error) {
	to := MultitenantTypeWithoutIDORM{}
	var err error
	to.SomeField = from.SomeField
	return to, err
}

// ConvertMultitenantTypeWithoutIDFromORM takes an orm object and returns a pb object
func ConvertMultitenantTypeWithoutIDFromORM(from MultitenantTypeWithoutIDORM) (MultitenantTypeWithoutId, error) {
	to := MultitenantTypeWithoutId{}
	var err error
	to.SomeField = from.SomeField
	return to, err
}

// TypeBecomesEmptyORM no comment was provided for message type
type TypeBecomesEmptyORM struct {
	// Skipping type *ApiOnlyType, not tagged as ormable
}

// TableName overrides the default tablename generated by GORM
func (TypeBecomesEmptyORM) TableName() string {
	return "type_becomes_empties"
}

// ConvertTypeBecomesEmptyToORM takes a pb object and returns an orm object
func ConvertTypeBecomesEmptyToORM(from TypeBecomesEmpty) (TypeBecomesEmptyORM, error) {
	to := TypeBecomesEmptyORM{}
	var err error
	return to, err
}

// ConvertTypeBecomesEmptyFromORM takes an orm object and returns a pb object
func ConvertTypeBecomesEmptyFromORM(from TypeBecomesEmptyORM) (TypeBecomesEmpty, error) {
	to := TypeBecomesEmpty{}
	var err error
	return to, err
}

////////////////////////// CURDL for objects
// DefaultCreateTestTypes executes a basic gorm create call
func DefaultCreateTestTypes(ctx context.Context, in *TestTypes, db *gorm.DB) (*TestTypes, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultCreateTestTypes")
	}
	ormObj, err := ConvertTestTypesToORM(*in)
	if err != nil {
		return nil, err
	}
	if err = db.Create(&ormObj).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ConvertTestTypesFromORM(ormObj)
	return &pbResponse, err
}

// DefaultReadTestTypes executes a basic gorm read call
func DefaultReadTestTypes(ctx context.Context, in *TestTypes, db *gorm.DB) (*TestTypes, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultReadTestTypes")
	}
	ormParams, err := ConvertTestTypesToORM(*in)
	if err != nil {
		return nil, err
	}
	ormResponse := TestTypesORM{}
	if err = db.Set("gorm:auto_preload", true).Where(&ormParams).First(&ormResponse).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ConvertTestTypesFromORM(ormResponse)
	return &pbResponse, err
}

// DefaultUpdateTestTypes executes a basic gorm update call
func DefaultUpdateTestTypes(ctx context.Context, in *TestTypes, db *gorm.DB) (*TestTypes, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultUpdateTestTypes")
	}
	ormObj, err := ConvertTestTypesToORM(*in)
	if err != nil {
		return nil, err
	}
	if err = db.Save(&ormObj).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ConvertTestTypesFromORM(ormObj)
	return &pbResponse, err
}

func DefaultDeleteTestTypes(ctx context.Context, in *TestTypes, db *gorm.DB) error {
	if in == nil {
		return errors.New("Nil argument to DefaultDeleteTestTypes")
	}
	ormObj, err := ConvertTestTypesToORM(*in)
	if err != nil {
		return err
	}
	err = db.Where(&ormObj).Delete(&TestTypesORM{}).Error
	return err
}

// DefaultListTestTypes executes a basic gorm find call
func DefaultListTestTypes(ctx context.Context, db *gorm.DB) ([]*TestTypes, error) {
	ormResponse := []TestTypesORM{}
	db, err := ops.ApplyCollectionOperators(db, ctx)
	if err != nil {
		return nil, err
	}
	if err := db.Set("gorm:auto_preload", true).Find(&ormResponse).Error; err != nil {
		return nil, err
	}
	pbResponse := []*TestTypes{}
	for _, responseEntry := range ormResponse {
		temp, err := ConvertTestTypesFromORM(responseEntry)
		if err != nil {
			return nil, err
		}
		pbResponse = append(pbResponse, &temp)
	}
	return pbResponse, nil
}

// DefaultUpdateTestTypes executes a basic gorm update call
func DefaultCascadedUpdateTestTypes(ctx context.Context, in *TestTypes, db *gorm.DB) (*TestTypes, error) {
	if in == nil {
		return nil, fmt.Errorf("Nil argument to DefaultCascadedUpdateTestTypes")
	}
	ormObj := ConvertTestTypesToORM(*in)
	tx := db.Begin()
	if err := tx.Save(&ormObj).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	pbResponse := ConvertTestTypesFromORM(ormObj)
	tx.Commit()
	return &pbResponse, nil
}

// DefaultCreateTypeWithID executes a basic gorm create call
func DefaultCreateTypeWithID(ctx context.Context, in *TypeWithId, db *gorm.DB) (*TypeWithId, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultCreateTypeWithID")
	}
	ormObj, err := ConvertTypeWithIDToORM(*in)
	if err != nil {
		return nil, err
	}
	if err = db.Create(&ormObj).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ConvertTypeWithIDFromORM(ormObj)
	return &pbResponse, err
}

// DefaultReadTypeWithID executes a basic gorm read call
func DefaultReadTypeWithID(ctx context.Context, in *TypeWithId, db *gorm.DB) (*TypeWithId, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultReadTypeWithID")
	}
	ormParams, err := ConvertTypeWithIDToORM(*in)
	if err != nil {
		return nil, err
	}
	ormResponse := TypeWithIDORM{}
	if err = db.Set("gorm:auto_preload", true).Where(&ormParams).First(&ormResponse).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ConvertTypeWithIDFromORM(ormResponse)
	return &pbResponse, err
}

// DefaultUpdateTypeWithID executes a basic gorm update call
func DefaultUpdateTypeWithID(ctx context.Context, in *TypeWithId, db *gorm.DB) (*TypeWithId, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultUpdateTypeWithID")
	}
	ormObj, err := ConvertTypeWithIDToORM(*in)
	if err != nil {
		return nil, err
	}
	if err = db.Save(&ormObj).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ConvertTypeWithIDFromORM(ormObj)
	return &pbResponse, err
}

func DefaultDeleteTypeWithID(ctx context.Context, in *TypeWithId, db *gorm.DB) error {
	if in == nil {
		return errors.New("Nil argument to DefaultDeleteTypeWithID")
	}
	ormObj, err := ConvertTypeWithIDToORM(*in)
	if err != nil {
		return err
	}
	err = db.Where(&ormObj).Delete(&TypeWithIDORM{}).Error
	return err
}

// DefaultListTypeWithID executes a basic gorm find call
func DefaultListTypeWithID(ctx context.Context, db *gorm.DB) ([]*TypeWithId, error) {
	ormResponse := []TypeWithIDORM{}
	db, err := ops.ApplyCollectionOperators(db, ctx)
	if err != nil {
		return nil, err
	}
	if err := db.Set("gorm:auto_preload", true).Find(&ormResponse).Error; err != nil {
		return nil, err
	}
	pbResponse := []*TypeWithId{}
	for _, responseEntry := range ormResponse {
		temp, err := ConvertTypeWithIDFromORM(responseEntry)
		if err != nil {
			return nil, err
		}
		pbResponse = append(pbResponse, &temp)
	}
	return pbResponse, nil
}

// DefaultUpdateTypeWithID executes a basic gorm update call
func DefaultCascadedUpdateTypeWithID(ctx context.Context, in *TypeWithId, db *gorm.DB) (*TypeWithId, error) {
	if in == nil {
		return nil, fmt.Errorf("Nil argument to DefaultCascadedUpdateTypeWithID")
	}
	ormObj := ConvertTypeWithIDToORM(*in)
	tx := db.Begin()
	if err := tx.Save(&ormObj).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	pbResponse := ConvertTypeWithIDFromORM(ormObj)
	tx.Commit()
	return &pbResponse, nil
}

// DefaultCreateMultitenantTypeWithID executes a basic gorm create call
func DefaultCreateMultitenantTypeWithID(ctx context.Context, in *MultitenantTypeWithId, db *gorm.DB) (*MultitenantTypeWithId, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultCreateMultitenantTypeWithID")
	}
	ormObj, err := ConvertMultitenantTypeWithIDToORM(*in)
	if err != nil {
		return nil, err
	}
	tenantID, tIDErr := auth.GetTenantID(ctx)
	if tIDErr != nil {
		return nil, tIDErr
	}
	ormObj.TenantID = tenantID
	if err = db.Create(&ormObj).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ConvertMultitenantTypeWithIDFromORM(ormObj)
	return &pbResponse, err
}

// DefaultReadMultitenantTypeWithID executes a basic gorm read call
func DefaultReadMultitenantTypeWithID(ctx context.Context, in *MultitenantTypeWithId, db *gorm.DB) (*MultitenantTypeWithId, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultReadMultitenantTypeWithID")
	}
	ormParams, err := ConvertMultitenantTypeWithIDToORM(*in)
	if err != nil {
		return nil, err
	}
	tenantID, tIDErr := auth.GetTenantID(ctx)
	if tIDErr != nil {
		return nil, tIDErr
	}
	ormParams.TenantID = tenantID
	ormResponse := MultitenantTypeWithIDORM{}
	if err = db.Set("gorm:auto_preload", true).Where(&ormParams).First(&ormResponse).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ConvertMultitenantTypeWithIDFromORM(ormResponse)
	return &pbResponse, err
}

// DefaultUpdateMultitenantTypeWithID executes a basic gorm update call
func DefaultUpdateMultitenantTypeWithID(ctx context.Context, in *MultitenantTypeWithId, db *gorm.DB) (*MultitenantTypeWithId, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultUpdateMultitenantTypeWithID")
	}
	if exists, err := DefaultReadMultitenantTypeWithID(ctx, &MultitenantTypeWithID{Id: in.GetId()}, db); err != nil {
		return nil, err
	} else if exists == nil {
		return nil, errors.New("MultitenantTypeWithID not found")
	}
	ormObj, err := ConvertMultitenantTypeWithIDToORM(*in)
	if err != nil {
		return nil, err
	}
	if err = db.Save(&ormObj).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ConvertMultitenantTypeWithIDFromORM(ormObj)
	return &pbResponse, err
}

func DefaultDeleteMultitenantTypeWithID(ctx context.Context, in *MultitenantTypeWithId, db *gorm.DB) error {
	if in == nil {
		return errors.New("Nil argument to DefaultDeleteMultitenantTypeWithID")
	}
	ormObj, err := ConvertMultitenantTypeWithIDToORM(*in)
	if err != nil {
		return err
	}
	tenantID, tIDErr := auth.GetTenantID(ctx)
	if tIDErr != nil {
		return tIDErr
	}
	ormObj.TenantID = tenantID
	err = db.Where(&ormObj).Delete(&MultitenantTypeWithIDORM{}).Error
	return err
}

// DefaultListMultitenantTypeWithID executes a basic gorm find call
func DefaultListMultitenantTypeWithID(ctx context.Context, db *gorm.DB) ([]*MultitenantTypeWithId, error) {
	ormResponse := []MultitenantTypeWithIDORM{}
	db, err := ops.ApplyCollectionOperators(db, ctx)
	if err != nil {
		return nil, err
	}
	tenantID, tIDErr := auth.GetTenantID(ctx)
	if tIDErr != nil {
		return nil, tIDErr
	}
	db = db.Where(&ContactORM{TenantID: tenantID})
	if err := db.Set("gorm:auto_preload", true).Find(&ormResponse).Error; err != nil {
		return nil, err
	}
	pbResponse := []*MultitenantTypeWithId{}
	for _, responseEntry := range ormResponse {
		temp, err := ConvertMultitenantTypeWithIDFromORM(responseEntry)
		if err != nil {
			return nil, err
		}
		pbResponse = append(pbResponse, &temp)
	}
	return pbResponse, nil
}

// DefaultUpdateMultitenantTypeWithID executes a basic gorm update call
func DefaultCascadedUpdateMultitenantTypeWithID(ctx context.Context, in *MultitenantTypeWithId, db *gorm.DB) (*MultitenantTypeWithId, error) {
	if in == nil {
		return nil, fmt.Errorf("Nil argument to DefaultCascadedUpdateMultitenantTypeWithID")
	}
	ormObj := ConvertMultitenantTypeWithIDToORM(*in)
	tx := db.Begin()
	if err := tx.Save(&ormObj).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	pbResponse := ConvertMultitenantTypeWithIDFromORM(ormObj)
	tx.Commit()
	return &pbResponse, nil
}

// DefaultCreateMultitenantTypeWithoutID executes a basic gorm create call
func DefaultCreateMultitenantTypeWithoutID(ctx context.Context, in *MultitenantTypeWithoutId, db *gorm.DB) (*MultitenantTypeWithoutId, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultCreateMultitenantTypeWithoutID")
	}
	ormObj, err := ConvertMultitenantTypeWithoutIDToORM(*in)
	if err != nil {
		return nil, err
	}
	tenantID, tIDErr := auth.GetTenantID(ctx)
	if tIDErr != nil {
		return nil, tIDErr
	}
	ormObj.TenantID = tenantID
	if err = db.Create(&ormObj).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ConvertMultitenantTypeWithoutIDFromORM(ormObj)
	return &pbResponse, err
}

// DefaultReadMultitenantTypeWithoutID executes a basic gorm read call
func DefaultReadMultitenantTypeWithoutID(ctx context.Context, in *MultitenantTypeWithoutId, db *gorm.DB) (*MultitenantTypeWithoutId, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultReadMultitenantTypeWithoutID")
	}
	ormParams, err := ConvertMultitenantTypeWithoutIDToORM(*in)
	if err != nil {
		return nil, err
	}
	tenantID, tIDErr := auth.GetTenantID(ctx)
	if tIDErr != nil {
		return nil, tIDErr
	}
	ormParams.TenantID = tenantID
	ormResponse := MultitenantTypeWithoutIDORM{}
	if err = db.Set("gorm:auto_preload", true).Where(&ormParams).First(&ormResponse).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ConvertMultitenantTypeWithoutIDFromORM(ormResponse)
	return &pbResponse, err
}

// Cannot autogen DefaultUpdateMultitenantTypeWithoutID: this is a multi-tenant table without an "id" field in the message.

func DefaultDeleteMultitenantTypeWithoutID(ctx context.Context, in *MultitenantTypeWithoutId, db *gorm.DB) error {
	if in == nil {
		return errors.New("Nil argument to DefaultDeleteMultitenantTypeWithoutID")
	}
	ormObj, err := ConvertMultitenantTypeWithoutIDToORM(*in)
	if err != nil {
		return err
	}
	tenantID, tIDErr := auth.GetTenantID(ctx)
	if tIDErr != nil {
		return tIDErr
	}
	ormObj.TenantID = tenantID
	err = db.Where(&ormObj).Delete(&MultitenantTypeWithoutIDORM{}).Error
	return err
}

// DefaultListMultitenantTypeWithoutID executes a basic gorm find call
func DefaultListMultitenantTypeWithoutID(ctx context.Context, db *gorm.DB) ([]*MultitenantTypeWithoutId, error) {
	ormResponse := []MultitenantTypeWithoutIDORM{}
	db, err := ops.ApplyCollectionOperators(db, ctx)
	if err != nil {
		return nil, err
	}
	tenantID, tIDErr := auth.GetTenantID(ctx)
	if tIDErr != nil {
		return nil, tIDErr
	}
	db = db.Where(&ContactORM{TenantID: tenantID})
	if err := db.Set("gorm:auto_preload", true).Find(&ormResponse).Error; err != nil {
		return nil, err
	}
	pbResponse := []*MultitenantTypeWithoutId{}
	for _, responseEntry := range ormResponse {
		temp, err := ConvertMultitenantTypeWithoutIDFromORM(responseEntry)
		if err != nil {
			return nil, err
		}
		pbResponse = append(pbResponse, &temp)
	}
	return pbResponse, nil
}

// DefaultUpdateMultitenantTypeWithoutID executes a basic gorm update call
func DefaultCascadedUpdateMultitenantTypeWithoutID(ctx context.Context, in *MultitenantTypeWithoutId, db *gorm.DB) (*MultitenantTypeWithoutId, error) {
	if in == nil {
		return nil, fmt.Errorf("Nil argument to DefaultCascadedUpdateMultitenantTypeWithoutID")
	}
	ormObj := ConvertMultitenantTypeWithoutIDToORM(*in)
	tx := db.Begin()
	if err := tx.Save(&ormObj).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	pbResponse := ConvertMultitenantTypeWithoutIDFromORM(ormObj)
	tx.Commit()
	return &pbResponse, nil
}

// DefaultCreateTypeBecomesEmpty executes a basic gorm create call
func DefaultCreateTypeBecomesEmpty(ctx context.Context, in *TypeBecomesEmpty, db *gorm.DB) (*TypeBecomesEmpty, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultCreateTypeBecomesEmpty")
	}
	ormObj, err := ConvertTypeBecomesEmptyToORM(*in)
	if err != nil {
		return nil, err
	}
	if err = db.Create(&ormObj).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ConvertTypeBecomesEmptyFromORM(ormObj)
	return &pbResponse, err
}

// DefaultReadTypeBecomesEmpty executes a basic gorm read call
func DefaultReadTypeBecomesEmpty(ctx context.Context, in *TypeBecomesEmpty, db *gorm.DB) (*TypeBecomesEmpty, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultReadTypeBecomesEmpty")
	}
	ormParams, err := ConvertTypeBecomesEmptyToORM(*in)
	if err != nil {
		return nil, err
	}
	ormResponse := TypeBecomesEmptyORM{}
	if err = db.Set("gorm:auto_preload", true).Where(&ormParams).First(&ormResponse).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ConvertTypeBecomesEmptyFromORM(ormResponse)
	return &pbResponse, err
}

// DefaultUpdateTypeBecomesEmpty executes a basic gorm update call
func DefaultUpdateTypeBecomesEmpty(ctx context.Context, in *TypeBecomesEmpty, db *gorm.DB) (*TypeBecomesEmpty, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultUpdateTypeBecomesEmpty")
	}
	ormObj, err := ConvertTypeBecomesEmptyToORM(*in)
	if err != nil {
		return nil, err
	}
	if err = db.Save(&ormObj).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ConvertTypeBecomesEmptyFromORM(ormObj)
	return &pbResponse, err
}

func DefaultDeleteTypeBecomesEmpty(ctx context.Context, in *TypeBecomesEmpty, db *gorm.DB) error {
	if in == nil {
		return errors.New("Nil argument to DefaultDeleteTypeBecomesEmpty")
	}
	ormObj, err := ConvertTypeBecomesEmptyToORM(*in)
	if err != nil {
		return err
	}
	err = db.Where(&ormObj).Delete(&TypeBecomesEmptyORM{}).Error
	return err
}

// DefaultListTypeBecomesEmpty executes a basic gorm find call
func DefaultListTypeBecomesEmpty(ctx context.Context, db *gorm.DB) ([]*TypeBecomesEmpty, error) {
	ormResponse := []TypeBecomesEmptyORM{}
	db, err := ops.ApplyCollectionOperators(db, ctx)
	if err != nil {
		return nil, err
	}
	if err := db.Set("gorm:auto_preload", true).Find(&ormResponse).Error; err != nil {
		return nil, err
	}
	pbResponse := []*TypeBecomesEmpty{}
	for _, responseEntry := range ormResponse {
		temp, err := ConvertTypeBecomesEmptyFromORM(responseEntry)
		if err != nil {
			return nil, err
		}
		pbResponse = append(pbResponse, &temp)
	}
	return pbResponse, nil
}

// DefaultUpdateTypeBecomesEmpty executes a basic gorm update call
func DefaultCascadedUpdateTypeBecomesEmpty(ctx context.Context, in *TypeBecomesEmpty, db *gorm.DB) (*TypeBecomesEmpty, error) {
	if in == nil {
		return nil, fmt.Errorf("Nil argument to DefaultCascadedUpdateTypeBecomesEmpty")
	}
	ormObj := ConvertTypeBecomesEmptyToORM(*in)
	tx := db.Begin()
	if err := tx.Save(&ormObj).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	pbResponse := ConvertTypeBecomesEmptyFromORM(ormObj)
	tx.Commit()
	return &pbResponse, nil
}
