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
		fieldType = strings.Trim(fieldType, "[]*")
		if p.isOrmable(fieldType) {
			if fieldOpts == nil {
				fieldOpts = &gorm.GormFieldOptions{}
			}
			assocOrmable := p.getOrmable(fieldType)
			if field.IsRepeated() {
				// if fieldOpts.GetManyToMany() != nil {
				// }
				p.parseHasMany(msg, ormable, fieldName, fieldType, assocOrmable, fieldOpts)
				fieldType = fmt.Sprintf("[]*%s", assocOrmable.Name)

			} else {
				if fieldOpts.GetBelongsTo() != nil {
					p.parseBelongsTo(msg, ormable, fieldName, fieldType, assocOrmable, fieldOpts)
				} else {
					p.parseHasOne(msg, ormable, fieldName, fieldType, assocOrmable, fieldOpts)
				}
				fieldType = fmt.Sprintf("*%s", assocOrmable.Name)
			}
			ormable.Fields[fieldName] = &Field{Type: fieldType, GormFieldOptions: fieldOpts}
		}
	}
}

func (p *OrmPlugin) countHasAssociationDimension(msg *generator.Descriptor, typeName string) int {
	dim := 0
	for _, field := range msg.GetField() {
		fieldOpts := getFieldOptions(field)
		if fieldOpts.GetDrop() {
			continue
		}
		fieldType, _ := p.GoType(msg, field)
		if fieldOpts.GetManyToMany() == nil && fieldOpts.GetBelongsTo() == nil {
			if strings.Trim(typeName, "[]*") == strings.Trim(fieldType, "[]*") {
				dim++
			}
		}
	}
	return dim
}

func (p *OrmPlugin) countBelongsToAssociationDimension(msg *generator.Descriptor, typeName string) int {
	dim := 0
	for _, field := range msg.GetField() {
		fieldOpts := getFieldOptions(field)
		if fieldOpts.GetDrop() {
			continue
		}
		fieldType, _ := p.GoType(msg, field)
		if fieldOpts.GetBelongsTo() != nil {
			if strings.Trim(typeName, "[]*") == strings.Trim(fieldType, "[]*") {
				dim++
			}
		}
	}
	return dim
}

func (p *OrmPlugin) parseHasMany(msg *generator.Descriptor, parent *OrmableType, fieldName string, fieldType string, child *OrmableType, opts *gorm.GormFieldOptions) {
	typeName := generator.CamelCaseSlice(msg.TypeName())
	hasMany := opts.GetHasMany()
	if hasMany == nil {
		hasMany = &gorm.HasManyOptions{}
		opts.Association = &gorm.GormFieldOptions_HasMany{hasMany}
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
		if p.countHasAssociationDimension(msg, fieldType) == 1 {
			foreignKeyName = fmt.Sprintf(typeName + assocKeyName)
		} else {
			foreignKeyName = fmt.Sprintf(fieldName + typeName + assocKeyName)
		}
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

func (p *OrmPlugin) parseHasOne(msg *generator.Descriptor, parent *OrmableType, fieldName string, fieldType string, child *OrmableType, opts *gorm.GormFieldOptions) {
	typeName := generator.CamelCaseSlice(msg.TypeName())
	hasOne := opts.GetHasOne()
	if hasOne == nil {
		hasOne = &gorm.HasOneOptions{}
		opts.Association = &gorm.GormFieldOptions_HasOne{hasOne}
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
		if p.countHasAssociationDimension(msg, fieldType) == 1 {
			foreignKeyName = fmt.Sprintf(typeName + assocKeyName)
		} else {
			foreignKeyName = fmt.Sprintf(fieldName + typeName + assocKeyName)
		}
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

func (p *OrmPlugin) parseBelongsTo(msg *generator.Descriptor, child *OrmableType, fieldName string, fieldType string, parent *OrmableType, opts *gorm.GormFieldOptions) {
	belongsTo := opts.GetBelongsTo()
	if belongsTo == nil {
		belongsTo = &gorm.BelongsToOptions{}
		opts.Association = &gorm.GormFieldOptions_BelongsTo{belongsTo}
	}
	var assocKey *Field
	var assocKeyName string
	if assocKeyName = generator.CamelCase(belongsTo.GetAssociationForeignkey()); assocKeyName == "" {
		assocKeyName, assocKey = p.findPrimaryKey(parent)
	} else {
		var ok bool
		assocKey, ok = parent.Fields[assocKeyName]
		if !ok {
			p.Fail("Missing", assocKeyName, "field in", parent.Name, ".")
		}
	}
	belongsTo.AssociationForeignkey = &assocKeyName
	foreignKey := &Field{Type: assocKey.Type, GormFieldOptions: &gorm.GormFieldOptions{Tag: belongsTo.GetForeignkeyTag()}}
	var foreignKeyName string
	if foreignKeyName = generator.CamelCase(belongsTo.GetForeignkey()); foreignKeyName == "" {
		if p.countBelongsToAssociationDimension(msg, fieldType) == 1 {
			foreignKeyName = fmt.Sprintf(fieldType + assocKeyName)
		} else {
			foreignKeyName = fmt.Sprintf(fieldName + assocKeyName)
		}
	}
	belongsTo.Foreignkey = &foreignKeyName
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

func (p *OrmPlugin) hasPrimaryKey(ormable *OrmableType) bool {
	for _, field := range ormable.Fields {
		if field.GetTag().GetPrimaryKey() {
			return true
		}
	}
	for fieldName, _ := range ormable.Fields {
		if strings.ToLower(fieldName) == "id" {
			return true
		}
	}
	return false
}
