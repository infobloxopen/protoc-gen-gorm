package plugin

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	jgorm "github.com/jinzhu/gorm"
	"github.com/jinzhu/inflection"

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
		parts := strings.Split(fieldType, ".")
		fieldTypeShort := parts[len(parts)-1]
		if p.isOrmable(fieldType) {
			if fieldOpts == nil {
				fieldOpts = &gorm.GormFieldOptions{}
			}
			assocOrmable := p.getOrmable(fieldType)
			if field.IsRepeated() {
				if fieldOpts.GetManyToMany() != nil {
					p.parseManyToMany(msg, ormable, fieldName, fieldTypeShort, assocOrmable, fieldOpts)
				} else {
					p.parseHasMany(msg, ormable, fieldName, fieldTypeShort, assocOrmable, fieldOpts)
				}
				fieldType = fmt.Sprintf("[]*%sORM", fieldType)
			} else {
				if fieldOpts.GetBelongsTo() != nil {
					p.parseBelongsTo(msg, ormable, fieldName, fieldTypeShort, assocOrmable, fieldOpts)
				} else {
					p.parseHasOne(msg, ormable, fieldName, fieldTypeShort, assocOrmable, fieldOpts)
				}
				fieldType = fmt.Sprintf("*%sORM", fieldType)
			}
			// Register type used, in case it's an imported type from another package
			p.GetFileImports().typesToRegister = append(p.GetFileImports().typesToRegister, field.GetTypeName())
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

func (p *OrmPlugin) countManyToManyAssociationDimension(msg *generator.Descriptor, typeName string) int {
	dim := 0
	for _, field := range msg.GetField() {
		fieldOpts := getFieldOptions(field)
		if fieldOpts.GetDrop() {
			continue
		}
		fieldType, _ := p.GoType(msg, field)
		if fieldOpts.GetManyToMany() != nil {
			if strings.Trim(typeName, "[]*") == strings.Trim(fieldType, "[]*") {
				dim++
			}
		}
	}
	return dim
}

func (p *OrmPlugin) resolveAliasName(goType, goPackage string, file *generator.FileDescriptor) string {
	originFile := p.currentFile
	p.setFile(file)
	isPointer := strings.HasPrefix(goType, "*")
	typeParts := strings.Split(goType, ".")
	if len(typeParts) == 2 {
		var newType string
		if strings.Contains(goPackage, "github.com") {
			newType = p.Import(goPackage) + "." + typeParts[1]
		} else {
			p.UsingGoImports(goPackage)
			packageParts := strings.Split(goPackage, "/")
			newType = packageParts[len(packageParts)-1] + "." + typeParts[1]
		}
		if isPointer {
			return "*" + newType
		}
		return newType
	}
	p.setFile(originFile)
	return goType
}

func (p *OrmPlugin) sameType(field1 *Field, field2 *Field) bool {
	isPointer1 := strings.HasPrefix(field1.Type, "*")
	typeParts1 := strings.Split(field1.Type, ".")
	if len(typeParts1) == 2 {
		isPointer2 := strings.HasPrefix(field2.Type, "*")
		typeParts2 := strings.Split(field2.Type, ".")
		if len(typeParts2) == 2 && isPointer1 == isPointer2 && typeParts1[1] == typeParts2[1] && field1.Package == field2.Package {
			return true
		}
		return false
	}
	return field1.Type == field2.Type
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
	var foreignKeyType string
	if hasMany.GetForeignkeyTag().GetNotNull() {
		foreignKeyType = strings.TrimPrefix(assocKey.Type, "*")
	} else if strings.HasPrefix(assocKey.Type, "*") {
		foreignKeyType = assocKey.Type
	} else if strings.Contains(assocKey.Type, "[]byte") {
		foreignKeyType = assocKey.Type
	} else {
		foreignKeyType = "*" + assocKey.Type
	}
	foreignKeyType = p.resolveAliasName(foreignKeyType, assocKey.Package, child.File)
	foreignKey := &Field{Type: foreignKeyType, Package: assocKey.Package, GormFieldOptions: &gorm.GormFieldOptions{Tag: hasMany.GetForeignkeyTag()}}
	var foreignKeyName string
	if foreignKeyName = hasMany.GetForeignkey(); foreignKeyName == "" {
		if p.countHasAssociationDimension(msg, fieldType) == 1 {
			foreignKeyName = fmt.Sprintf(typeName + assocKeyName)
		} else {
			foreignKeyName = fmt.Sprintf(fieldName + typeName + assocKeyName)
		}
	}
	hasMany.Foreignkey = &foreignKeyName
	if _, ok := child.Fields[foreignKeyName]; child.Package != parent.Package && !ok {
		p.Fail(`Object`, child.Name, `from package`, child.Package, `cannot be used for has-many in`, parent.Name, `since it`,
			`does not have FK`, foreignKeyName, `defined. Manually define the key, or switch to many-to-many`)
	}
	if exField, ok := child.Fields[foreignKeyName]; !ok {
		child.Fields[foreignKeyName] = foreignKey
	} else {
		if exField.Type == "interface{}" {
			exField.Type = foreignKey.Type
		} else if !p.sameType(exField, foreignKey) {
			p.Fail("Cannot include", foreignKeyName, "field into", child.Name, "as it already exists there with a different type:", exField.Type, foreignKey.Type)
		}
	}
	child.Fields[foreignKeyName].ParentOriginName = parent.OriginName

	var posField string
	if posField = generator.CamelCase(hasMany.GetPositionField()); posField != "" {
		if exField, ok := child.Fields[posField]; !ok {
			child.Fields[posField] = &Field{Type: "int", GormFieldOptions: &gorm.GormFieldOptions{Tag: hasMany.GetPositionFieldTag()}}
		} else {
			if !strings.Contains(exField.Type, "int") {
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
	var foreignKeyType string
	if hasOne.GetForeignkeyTag().GetNotNull() {
		foreignKeyType = strings.TrimPrefix(assocKey.Type, "*")
	} else if strings.HasPrefix(assocKey.Type, "*") {
		foreignKeyType = assocKey.Type
	} else if strings.Contains(assocKey.Type, "[]byte") {
		foreignKeyType = assocKey.Type
	} else {
		foreignKeyType = "*" + assocKey.Type
	}
	foreignKeyType = p.resolveAliasName(foreignKeyType, assocKey.Package, child.File)
	foreignKey := &Field{Type: foreignKeyType, Package: assocKey.Package, GormFieldOptions: &gorm.GormFieldOptions{Tag: hasOne.GetForeignkeyTag()}}
	var foreignKeyName string
	if foreignKeyName = generator.CamelCase(hasOne.GetForeignkey()); foreignKeyName == "" {
		if p.countHasAssociationDimension(msg, fieldType) == 1 {
			foreignKeyName = fmt.Sprintf(typeName + assocKeyName)
		} else {
			foreignKeyName = fmt.Sprintf(fieldName + typeName + assocKeyName)
		}
	}
	hasOne.Foreignkey = &foreignKeyName
	if _, ok := child.Fields[foreignKeyName]; child.Package != parent.Package && !ok {
		p.Fail(`Object`, child.Name, `from package`, child.Package, `cannot be used for has-one in`, parent.Name, `since it`,
			`does not have FK field`, foreignKeyName, `defined. Manually define the key, or switch to belongs-to`)
	}
	if exField, ok := child.Fields[foreignKeyName]; !ok {
		child.Fields[foreignKeyName] = foreignKey
	} else {
		if exField.Type == "interface{}" {
			exField.Type = foreignKey.Type
		} else if !p.sameType(exField, foreignKey) {
			p.Fail("Cannot include", foreignKeyName, "field into", child.Name, "as it already exists there with a different type:", exField.Type, foreignKey.Type)
		}
	}
	child.Fields[foreignKeyName].ParentOriginName = parent.OriginName
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
	var foreignKeyType string
	if belongsTo.GetForeignkeyTag().GetNotNull() {
		foreignKeyType = strings.TrimPrefix(assocKey.Type, "*")
	} else if strings.HasPrefix(assocKey.Type, "*") {
		foreignKeyType = assocKey.Type
	} else if strings.Contains(assocKey.Type, "[]byte") {
		foreignKeyType = assocKey.Type
	} else {
		foreignKeyType = "*" + assocKey.Type
	}
	foreignKeyType = p.resolveAliasName(foreignKeyType, assocKey.Package, child.File)
	foreignKey := &Field{Type: foreignKeyType, Package: assocKey.Package, GormFieldOptions: &gorm.GormFieldOptions{Tag: belongsTo.GetForeignkeyTag()}}
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
		if exField.Type == "interface{}" {
			exField.Type = foreignKeyType
		} else if !p.sameType(exField, foreignKey) {
			p.Fail("Cannot include", foreignKeyName, "field into", child.Name, "as it already exists there with a different type:", exField.Type, foreignKey.Type)
		}
	}
	child.Fields[foreignKeyName].ParentOriginName = parent.OriginName
}

func (p *OrmPlugin) parseManyToMany(msg *generator.Descriptor, ormable *OrmableType, fieldName string, fieldType string, assoc *OrmableType, opts *gorm.GormFieldOptions) {
	typeName := generator.CamelCaseSlice(msg.TypeName())
	mtm := opts.GetManyToMany()
	if mtm == nil {
		mtm = &gorm.ManyToManyOptions{}
		opts.Association = &gorm.GormFieldOptions_ManyToMany{mtm}
	}

	var foreignKeyName string
	if foreignKeyName = generator.CamelCase(mtm.GetForeignkey()); foreignKeyName == "" {
		foreignKeyName, _ = p.findPrimaryKey(ormable)
	} else {
		var ok bool
		_, ok = ormable.Fields[foreignKeyName]
		if !ok {
			p.Fail("Missing", foreignKeyName, "field in", ormable.Name, ".")
		}
	}
	mtm.Foreignkey = &foreignKeyName
	var assocKeyName string
	if assocKeyName = generator.CamelCase(mtm.GetAssociationForeignkey()); assocKeyName == "" {
		assocKeyName, _ = p.findPrimaryKey(assoc)
	} else {
		var ok bool
		_, ok = assoc.Fields[assocKeyName]
		if !ok {
			p.Fail("Missing", assocKeyName, "field in", assoc.Name, ".")
		}
	}
	mtm.AssociationForeignkey = &assocKeyName
	var jt string
	if jt = jgorm.ToDBName(mtm.GetJointable()); jt == "" {
		if p.countManyToManyAssociationDimension(msg, fieldType) == 1 && typeName != fieldType {
			jt = jgorm.ToDBName(typeName + inflection.Plural(fieldType))
		} else {
			jt = jgorm.ToDBName(typeName + inflection.Plural(fieldName))
		}
	}
	mtm.Jointable = &jt
	var jtForeignKey string
	if jtForeignKey = generator.CamelCase(mtm.GetJointableForeignkey()); jtForeignKey == "" {
		jtForeignKey = jgorm.ToDBName(typeName + foreignKeyName)
	}
	mtm.JointableForeignkey = &jtForeignKey
	var jtAssocForeignKey string
	if jtAssocForeignKey = generator.CamelCase(mtm.GetAssociationJointableForeignkey()); jtAssocForeignKey == "" {
		if typeName == fieldType {
			jtAssocForeignKey = jgorm.ToDBName(inflection.Singular(fieldName) + assocKeyName)
		} else {
			jtAssocForeignKey = jgorm.ToDBName(fieldType + assocKeyName)
		}
	}
	mtm.AssociationJointableForeignkey = &jtAssocForeignKey
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
