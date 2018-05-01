// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: example/contacts/contacts.proto

/*
Package contacts is a generated protocol buffer package.

It is generated from these files:
	example/contacts/contacts.proto

It has these top-level messages:
	Contact
	ContactPage
	SearchRequest
	GetRequest
	SMSRequest
*/
package contacts

import context "context"
import errors "errors"

import auth "github.com/infobloxopen/atlas-app-toolkit/mw/auth"
import gorm "github.com/jinzhu/gorm"
import ops "github.com/infobloxopen/atlas-app-toolkit/op/gorm"

import fmt "fmt"
import math "math"
import _ "github.com/golang/protobuf/ptypes/empty"
import _ "google.golang.org/genproto/googleapis/api/annotations"
import _ "github.com/lyft/protoc-gen-validate/validate"
import _ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger/options"

// Reference imports to suppress errors if they are not otherwise used.
var _ = fmt.Errorf
var _ = math.Inf

// ContactORM no comment was provided for message type
type ContactORM struct {
	AccountID    string
	Id           uint64
	FirstName    string
	MiddleName   string
	LastName     string
	EmailAddress string
}

// TableName overrides the default tablename generated by GORM
func (ContactORM) TableName() string {
	return "contacts"
}

// ToORM adds a pb object function that returns an orm object
func (m *Contact) ToORM() (ContactORM, error) {
	to := ContactORM{}
	if prehook, ok := interface{}(m).(ContactWithBeforeToORM); ok {
		prehook.BeforeToORM(to)
	}
	var err error
	to.Id = m.Id
	to.FirstName = m.FirstName
	to.MiddleName = m.MiddleName
	to.LastName = m.LastName
	to.EmailAddress = m.EmailAddress
	if posthook, ok := interface{}(m).(ContactWithAfterToORM); ok {
		posthook.AfterToORM(to)
	}
	return to, err
}

// FromORM returns a pb object
func (m *ContactORM) ToPB() (Contact, error) {
	to := Contact{}
	if prehook, ok := interface{}(m).(ContactWithBeforeToPB); ok {
		prehook.BeforeToPB(to)
	}
	var err error
	to.Id = m.Id
	to.FirstName = m.FirstName
	to.MiddleName = m.MiddleName
	to.LastName = m.LastName
	to.EmailAddress = m.EmailAddress
	if posthook, ok := interface{}(m).(ContactWithAfterToPB); ok {
		posthook.AfterToPB(to)
	}
	return to, err
}

// The following are interfaces you can implement for special behavior during ORM/PB conversions
// of type Contact the arg will be the target, the caller the one being converted from

// ContactBeforeToORM called before default ToORM code
type ContactWithBeforeToORM interface {
	BeforeToORM(ContactORM)
}

// ContactAfterToORM called after default ToORM code
type ContactWithAfterToORM interface {
	AfterToORM(ContactORM)
}

// ContactBeforeToPB called before default ToPB code
type ContactWithBeforeToPB interface {
	BeforeToPB(Contact)
}

// ContactAfterToPB called after default ToPB code
type ContactWithAfterToPB interface {
	AfterToPB(Contact)
}

////////////////////////// CURDL for objects
// DefaultCreateContact executes a basic gorm create call
func DefaultCreateContact(ctx context.Context, in *Contact, db *gorm.DB) (*Contact, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultCreateContact")
	}
	ormObj, err := in.ToORM()
	if err != nil {
		return nil, err
	}
	accountID, err := auth.GetAccountID(ctx, nil)
	if err != nil {
		return nil, err
	}
	ormObj.AccountID = accountID
	if err = db.Create(&ormObj).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ormObj.ToPB()
	return &pbResponse, err
}

// DefaultReadContact executes a basic gorm read call
func DefaultReadContact(ctx context.Context, in *Contact, db *gorm.DB) (*Contact, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultReadContact")
	}
	ormParams, err := in.ToORM()
	if err != nil {
		return nil, err
	}
	accountID, err := auth.GetAccountID(ctx, nil)
	if err != nil {
		return nil, err
	}
	ormParams.AccountID = accountID
	ormResponse := ContactORM{}
	if err = db.Set("gorm:auto_preload", true).Where(&ormParams).First(&ormResponse).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ormResponse.ToPB()
	return &pbResponse, err
}

// DefaultUpdateContact executes a basic gorm update call
func DefaultUpdateContact(ctx context.Context, in *Contact, db *gorm.DB) (*Contact, error) {
	if in == nil {
		return nil, errors.New("Nil argument to DefaultUpdateContact")
	}
	if exists, err := DefaultReadContact(ctx, &Contact{Id: in.GetId()}, db); err != nil {
		return nil, err
	} else if exists == nil {
		return nil, errors.New("Contact not found")
	}
	ormObj, err := in.ToORM()
	if err != nil {
		return nil, err
	}
	if err = db.Save(&ormObj).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ormObj.ToPB()
	return &pbResponse, err
}

func DefaultDeleteContact(ctx context.Context, in *Contact, db *gorm.DB) error {
	if in == nil {
		return errors.New("Nil argument to DefaultDeleteContact")
	}
	ormObj, err := in.ToORM()
	if err != nil {
		return err
	}
	accountID, err := auth.GetAccountID(ctx, nil)
	if err != nil {
		return err
	}
	ormObj.AccountID = accountID
	err = db.Where(&ormObj).Delete(&ContactORM{}).Error
	return err
}

// DefaultListContact executes a gorm list call
func DefaultListContact(ctx context.Context, db *gorm.DB) ([]*Contact, error) {
	ormResponse := []ContactORM{}
	db, err := ops.ApplyCollectionOperators(db, ctx)
	if err != nil {
		return nil, err
	}
	accountID, err := auth.GetAccountID(ctx, nil)
	if err != nil {
		return nil, err
	}
	db = db.Where(&ContactORM{AccountID: accountID})
	if err := db.Set("gorm:auto_preload", true).Find(&ormResponse).Error; err != nil {
		return nil, err
	}
	pbResponse := []*Contact{}
	for _, responseEntry := range ormResponse {
		temp, err := responseEntry.ToPB()
		if err != nil {
			return nil, err
		}
		pbResponse = append(pbResponse, &temp)
	}
	return pbResponse, nil
}

// DefaultStrictUpdateContact clears first level 1:many children and then executes a gorm update call
func DefaultStrictUpdateContact(ctx context.Context, in *Contact, db *gorm.DB) (*Contact, error) {
	if in == nil {
		return nil, fmt.Errorf("Nil argument to DefaultCascadedUpdateContact")
	}
	ormObj, err := in.ToORM()
	if err != nil {
		return nil, err
	}
	accountID, err := auth.GetAccountID(ctx, nil)
	if err != nil {
		return nil, err
	}
	db = db.Where(&ContactORM{AccountID: accountID})
	if err = db.Save(&ormObj).Error; err != nil {
		return nil, err
	}
	pbResponse, err := ormObj.ToPB()
	if err != nil {
		return nil, err
	}
	return &pbResponse, nil
}
