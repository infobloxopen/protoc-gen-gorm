package gorm

import (
	"context"
	"strings"

	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/query"
	"github.com/jinzhu/gorm"
)

func ApplyCollectionOperators(db *gorm.DB, ctx context.Context) (*gorm.DB, error) {
	// ApplyCollectionOperators applies query operators taken from context ctx to gorm instance db.
	f, err := gateway.Filtering(ctx)
	if err != nil {
		return nil, err
	}
	db, err = ApplyFiltering(db, f)
	if err != nil {
		return nil, err
	}

	var s *query.Sorting
	s, err = gateway.Sorting(ctx)
	if err != nil {
		return nil, err
	}
	db = ApplySorting(db, s)

	var p *query.Pagination
	p, err = gateway.Pagination(ctx)
	if err != nil {
		return nil, err
	}
	db = ApplyPagination(db, p)

	fs := gateway.FieldSelection(ctx)
	if err != nil {
		return nil, err
	}
	db = ApplyFieldSelection(db, fs)

	return db, nil
}

// ApplyFiltering applies filtering operator f to gorm instance db.
func ApplyFiltering(db *gorm.DB, f *query.Filtering) (*gorm.DB, error) {
	str, args, err := FilteringToGorm(f)
	if err != nil {
		return nil, err
	}
	if str != "" {
		return db.Where(str, args...), nil
	}
	return db, nil
}

// ApplySorting applies sorting operator s to gorm instance db.
func ApplySorting(db *gorm.DB, s *query.Sorting) *gorm.DB {
	var crs []string
	for _, cr := range s.GetCriterias() {
		if cr.IsDesc() {
			crs = append(crs, cr.GetTag()+" desc")
		} else {
			crs = append(crs, cr.GetTag())
		}
	}
	if len(crs) > 0 {
		return db.Order(strings.Join(crs, ","))
	}
	return db
}

// ApplyPagination applies pagination operator p to gorm instance db.
func ApplyPagination(db *gorm.DB, p *query.Pagination) *gorm.DB {
	return db.Offset(p.GetOffset()).Limit(p.DefaultLimit())
}

// ApplyFieldSelection applies field selection operator fs to gorm instance db.
func ApplyFieldSelection(db *gorm.DB, fs *query.FieldSelection) *gorm.DB {
	var fields []string
	for _, f := range fs.GetFields() {
		fields = append(fields, f.GetName())
	}
	if len(fields) > 0 {
		return db.Select(fields)
	}
	return db
}
