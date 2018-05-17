package plugin

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	gorm "github.com/infobloxopen/protoc-gen-gorm/options"
)

func (p *OrmPlugin) parseAssociations(msg *generator.Descriptor) {
	typeName := generator.CamelCaseSlice(msg.TypeName())
	ormable := p.getOrmable(typeName)
	for _, field := range msg.GetField() {
		fieldOpts := getFieldOptions(field)
		if fieldOpts.GetDrop() {
			continue
		}
		fieldName := generator.CamelCase(field.GetName())
		fieldType, _ := p.GoType(msg, field)
		if p.isOrmable(fieldType) {
			if fieldOpts == nil {
				fieldOpts = &gorm.GormFieldOptions{}
			}
			childOrmable := p.getOrmable(fieldType)
			if field.IsRepeated() {
				// if fieldOpts.GetManyToMany() != nil {
				// }
				p.parseHasMany(typeName, ormable, fieldName, childOrmable, fieldOpts)
				fieldType = fmt.Sprintf("[]*%s", childOrmable.Name)

			} else {
				// if fieldOpts.GetBelongsTo() != nil {
				// }
				p.parseHasOne(typeName, ormable, fieldName, childOrmable, fieldOpts)
				fieldType = fmt.Sprintf("*%s", childOrmable.Name)
			}
			ormable.Fields[fieldName] = &Field{Type: fieldType, GormFieldOptions: fieldOpts}
		}
	}
}

func (p *OrmPlugin) parseHasMany(typeName string, parent *OrmableType, fieldName string, child *OrmableType, opts *gorm.GormFieldOptions) {
	hasMany := opts.GetHasMany()
	if hasMany == nil {
		hasMany = &gorm.HasManyOptions{}
		opts.HasMany = hasMany
	}
	var assocKey *Field
	var assocKeyName string
	if assocKeyName = generator.CamelCase(hasMany.GetAssociationForeignkey()); assocKeyName == "" {
		assocKeyName, assocKey = p.findPrimaryKey(parent)
	} else {
		var ok bool
		assocKey, ok = parent.Fields[assocKeyName]
		if !ok {
			p.Fail("Missing", assocKeyName, "field in", parent.Name, ".")
		}
	}
	hasMany.AssociationForeignkey = &assocKeyName
	foreignKey := &Field{Type: assocKey.Type, GormFieldOptions: &gorm.GormFieldOptions{Tag: hasMany.GetForeignkeyTag()}}
	var foreignKeyName string
	if foreignKeyName = hasMany.GetForeignkey(); foreignKeyName == "" {
		foreignKeyName = fmt.Sprintf(fieldName + typeName + assocKeyName)
	}
	hasMany.Foreignkey = &foreignKeyName
	if exField, ok := child.Fields[foreignKeyName]; !ok {
		child.Fields[foreignKeyName] = foreignKey
	} else {
		if exField.Type != foreignKey.Type {
			p.Fail("Cannot include", foreignKeyName, "field into", child.Name, "as it already exists there with a different type.")
		}
	}
	var posField string
	if posField = generator.CamelCase(hasMany.GetPositionField()); posField != "" {
		if exField, ok := child.Fields[posField]; !ok {
			child.Fields[posField] = &Field{Type: "int", GormFieldOptions: &gorm.GormFieldOptions{Tag: hasMany.GetPositionFieldTag()}}
		} else {
			if strings.Contains(exField.Type, "int") {
				p.Fail("Cannot include", posField, "field into", child.Name, "as it already exists there with a different type.")
			}
		}
		hasMany.PositionField = &posField
	}
}

func (p *OrmPlugin) parseHasOne(typeName string, parent *OrmableType, fieldName string, child *OrmableType, opts *gorm.GormFieldOptions) {
	hasOne := opts.GetHasOne()
	if hasOne == nil {
		hasOne = &gorm.HasOneOptions{}
		opts.HasOne = hasOne
	}
	var assocKey *Field
	var assocKeyName string
	if assocKeyName = generator.CamelCase(hasOne.GetAssociationForeignkey()); assocKeyName == "" {
		assocKeyName, assocKey = p.findPrimaryKey(parent)
	} else {
		var ok bool
		assocKey, ok = parent.Fields[assocKeyName]
		if !ok {
			p.Fail("Missing", assocKeyName, "field in", parent.Name, ".")
		}
	}
	hasOne.AssociationForeignkey = &assocKeyName
	foreignKey := &Field{Type: assocKey.Type, GormFieldOptions: &gorm.GormFieldOptions{Tag: hasOne.GetForeignkeyTag()}}
	var foreignKeyName string
	if foreignKeyName = generator.CamelCase(hasOne.GetForeignkey()); foreignKeyName == "" {
		foreignKeyName = fmt.Sprintf(fieldName + typeName + assocKeyName)
	}
	hasOne.Foreignkey = &foreignKeyName
	if exField, ok := child.Fields[foreignKeyName]; !ok {
		child.Fields[foreignKeyName] = foreignKey
	} else {
		if exField.Type != foreignKey.Type {
			p.Fail("Cannot include", foreignKeyName, "field into", child.Name, "as it already exists there with a different type.")
		}
	}
}

func (p *OrmPlugin) findPrimaryKey(ormable *OrmableType) (string, *Field) {
	for fieldName, field := range ormable.Fields {
		if field.GetTag().GetPrimaryKey() {
			return fieldName, field
		}
	}
	for fieldName, field := range ormable.Fields {
		if strings.ToLower(fieldName) == "id" {
			return fieldName, field
		}
	}
	p.Fail("Primary key cannot be found in", ormable.Name, ".")
	return "", nil
}
