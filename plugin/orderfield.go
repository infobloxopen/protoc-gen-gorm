package plugin

import (
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

type OrmOrderField struct{}

func (o *OrmOrderField) getOrderedFieldNames(ormable *OrmableType, message *generator.Descriptor) (fields []string) {

	for _, v := range message.GetField() {
		fieldName := generator.CamelCase(*v.Name)
		fields = append(fields, fieldName)
	}

	ormFields := o.getOrmableFieldNames(ormable.Fields)
	for _, ormField := range ormFields {
		if !o.searchField(fields, ormField) {
			fields = append(fields, ormField)
		}
	}

	return fields
}

func (o *OrmOrderField) searchField(fields []string, searchName string) bool {
	for _, fieldName := range fields {
		if searchName == fieldName {
			return true
		}
	}
	return false
}

func (o *OrmOrderField) getOrmableFieldNames(fields map[string]*Field) (keys []string) {
	for k := range fields {
		keys = append(keys, k)
	}
	return keys
}
