package plugin

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	gorm "github.com/infobloxopen/protoc-gen-gorm/options"
)

func (p *OrmPlugin) generateDefaultHandlers(file *generator.FileDescriptor) {
	for _, message := range file.Messages() {
		if message.Options != nil {
			if opts := getMessageOptions(message); opts == nil || !*opts.Ormable {
				continue
			} else if opts.GetMultiAccount() {
				p.usingAuth = true
			}
		} else {
			continue
		}
		p.gormPkgName = "gorm"
		p.lftPkgName = "ops"

		p.generateCreateHandler(message)
		p.generateReadHandler(message)
		p.generateUpdateHandler(message)
		p.generateDeleteHandler(message)
		p.generateListHandler(message)
		p.generateStrictUpdateHandler(message)
	}
}

func (p *OrmPlugin) generateCreateHandler(message *generator.Descriptor) {
	typeName := p.TypeName(message)
	p.P(`// DefaultCreate`, typeName, ` executes a basic gorm create call`)
	p.P(`func DefaultCreate`, typeName, `(ctx context.Context, in *`,
		typeName, `, db *`, p.gormPkgName, `.DB) (*`, typeName, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, errors.New("Nil argument to DefaultCreate`, typeName, `")`)
	p.P(`}`)
	p.P(`ormObj, err := in.ToOrm()`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	if opts := getMessageOptions(message); opts != nil && opts.GetMultiAccount() {
		p.P("accountID, err := auth.GetAccountID(ctx, nil)")
		p.P("if err != nil {")
		p.P("return nil, err")
		p.P("}")
		p.P("ormObj.AccountID = accountID")
	}
	p.P(`if err = db.Create(&ormObj).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse, err := ormObj.ToPB()`)
	p.P(`return &pbResponse, err`)
	p.P(`}`)
	p.P()
}

func (p *OrmPlugin) generateReadHandler(message *generator.Descriptor) {
	typeName := p.TypeName(message)
	p.P(`// DefaultRead`, typeName, ` executes a basic gorm read call`)
	p.P(`func DefaultRead`, typeName, `(ctx context.Context, in *`,
		typeName, `, db *`, p.gormPkgName, `.DB) (*`, typeName, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, errors.New("Nil argument to DefaultRead`, typeName, `")`)
	p.P(`}`)
	p.P(`ormParams, err := in.ToOrm()`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	if opts := getMessageOptions(message); opts != nil && opts.GetMultiAccount() {
		p.P("accountID, err := auth.GetAccountID(ctx, nil)")
		p.P("if err != nil {")
		p.P("return nil, err")
		p.P("}")
		p.P("ormParams.AccountID = accountID")
	}
	p.P(`ormResponse := `, typeName, `ORM{}`)
	p.P(`if err = db.Set("gorm:auto_preload", true).Where(&ormParams).First(&ormResponse).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse, err := ormResponse.ToPB()`)
	p.P(`return &pbResponse, err`)
	p.P(`}`)
	p.P()
}

func (p *OrmPlugin) generateUpdateHandler(message *generator.Descriptor) {
	typeName := p.TypeName(message)

	hasIDField := false
	for _, field := range message.GetField() {
		if strings.ToLower(field.GetName()) == "id" {
			hasIDField = true
			break
		}
	}
	isMultiAccount := false
	if opts := getMessageOptions(message); opts != nil && opts.GetMultiAccount() {
		isMultiAccount = true
	}

	if isMultiAccount && !hasIDField {
		p.P(fmt.Sprintf("// Cannot autogen DefaultUpdate%s: this is a multi-account table without an \"id\" field in the message.\n", typeName))
		return
	}

	p.P(`// DefaultUpdate`, typeName, ` executes a basic gorm update call`)
	p.P(`func DefaultUpdate`, typeName, `(ctx context.Context, in *`,
		typeName, `, db *`, p.gormPkgName, `.DB) (*`, typeName, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, errors.New("Nil argument to DefaultUpdate`, typeName, `")`)
	p.P(`}`)
	if isMultiAccount {
		p.P(fmt.Sprintf("if exists, err := DefaultRead%s(ctx, &%s{Id: in.GetId()}, db); err != nil {",
			typeName, typeName))
		p.P("return nil, err")
		p.P("} else if exists == nil {")
		p.P(fmt.Sprintf("return nil, errors.New(\"%s not found\")", typeName))
		p.P("}")
	}

	p.P(`ormObj, err := in.ToOrm()`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`if err = db.Save(&ormObj).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse, err := ormObj.ToPB()`)
	p.P(`return &pbResponse, err`)
	p.P(`}`)
	p.P()
}

func (p *OrmPlugin) generateDeleteHandler(message *generator.Descriptor) {
	typeName := p.TypeName(message)
	p.P(`func DefaultDelete`, typeName, `(ctx context.Context, in *`,
		typeName, `, db *`, p.gormPkgName, `.DB) error {`)
	p.P(`if in == nil {`)
	p.P(`return errors.New("Nil argument to DefaultDelete`, typeName, `")`)
	p.P(`}`)
	p.P(`ormObj, err := in.ToOrm()`)
	p.P(`if err != nil {`)
	p.P(`return err`)
	p.P(`}`)
	if opts := getMessageOptions(message); opts != nil && opts.GetMultiAccount() {
		p.P("accountID, err := auth.GetAccountID(ctx, nil)")
		p.P("if err != nil {")
		p.P("return err")
		p.P("}")
		p.P("ormObj.AccountID = accountID")
	}
	p.P(`err = db.Where(&ormObj).Delete(&`, typeName, `ORM{}).Error`)
	p.P(`return err`)
	p.P(`}`)
	p.P()
}

func (p *OrmPlugin) generateListHandler(message *generator.Descriptor) {
	typeName := p.TypeName(message)

	p.P(`// DefaultList`, typeName, ` executes a gorm list call`)
	p.P(`func DefaultList`, typeName, `(ctx context.Context, db *`, p.gormPkgName,
		`.DB) ([]*`, typeName, `, error) {`)
	p.P(`ormResponse := []`, typeName, `ORM{}`)
	p.P(`db, err := `, p.lftPkgName, `.ApplyCollectionOperators(db, ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	if opts := getMessageOptions(message); opts != nil && opts.GetMultiAccount() {
		p.P("accountID, err := auth.GetAccountID(ctx, nil)")
		p.P("if err != nil {")
		p.P("return nil, err")
		p.P("}")
		p.P(`db = db.Where(&`, typeName, `ORM{AccountID: accountID})`)
	}
	p.P(`if err := db.Set("gorm:auto_preload", true).Find(&ormResponse).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse := []*`, typeName, `{}`)
	p.P(`for _, responseEntry := range ormResponse {`)
	p.P(`temp, err := responseEntry.ToPB()`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse = append(pbResponse, &temp)`)
	p.P(`}`)
	p.P(`return pbResponse, nil`)
	p.P(`}`)
	p.P()
}

/////////// For association removal during update

// findAssociationKeys should return a map[childFK]parentKeyField
func (p *OrmPlugin) findAssociationKeys(parent *generator.Descriptor,
	child *generator.Descriptor, field *descriptor.FieldDescriptorProto) map[string]string {
	// Check for gorm tags
	parentTypeName := p.TypeName(parent)
	keyMap := make(map[string]string)
	defaultKeyMap := map[string]string{fmt.Sprintf("%sId", parentTypeName): "Id"}
	childFields := []string{fmt.Sprintf("%sId", parentTypeName)}
	parentFields := []string{"Id"}
	if field.Options == nil {
		return defaultKeyMap
	}
	v, err := proto.GetExtension(field.Options, gorm.E_Field)
	if err != nil {
		return defaultKeyMap
	}
	gfOptions, ok := v.(*gorm.GormFieldOptions)
	if !ok || gfOptions.Tags == nil {
		return defaultKeyMap
	}
	tag := gfOptions.GetTags()
	value, ok := reflect.StructTag(tag).Lookup("gorm")
	if !ok {
		return defaultKeyMap
	}
	tagParts := strings.Split(value, ";") // Can have multiple ';' separated tags
	for _, arg := range tagParts {
		tagType := strings.Split(arg, ":") // tags follow key:value convention
		tagType[0] = strings.ToLower(tagType[0])
		if tagType[0] == "many2many" {
			// Not there just yet
		} else if tagType[0] == "foreignkey" {
			childFields = []string{}
			for _, fName := range strings.Split(tagType[1], ",") { // for compound key
				childFields = append(childFields, generator.CamelCase(fName))
			}
		} else if tagType[0] == "association_foreignkey" {
			parentFields = []string{}
			for _, fName := range strings.Split(tagType[1], ",") { // for compound key
				parentFields = append(parentFields, generator.CamelCase(fName))
			}
		}
	}

	if len(childFields) != len(parentFields) {
		p.Fail(`Number of foreign keys and association foreign keys between type `,
			parentTypeName, ` and `, field.GetName(), ` didn't match`)
	} else {
		for i, child := range childFields {
			keyMap[child] = parentFields[i]
		}
	}
	return keyMap
}

// guessZeroValue of the input type, so that we can check if a (key) value is set or not
func guessZeroValue(typeName string) string {
	typeName = strings.ToLower(typeName)
	if strings.Contains(typeName, "string") {
		return `""`
	}
	if strings.Contains(typeName, "int") {
		return `0`
	}
	if strings.Contains(typeName, "uuid") {
		return `uuid.Nil`
	}
	return ``
}

func (p *OrmPlugin) removeChildAssociations(message *generator.Descriptor) bool {
	typeName := p.TypeName(message)
	usedAccountID := false
	if _, exists := typeNames[typeName]; !exists {
		return usedAccountID
	}
	for _, field := range message.Field {
		// Only looking at slices
		if !field.IsRepeated() {
			continue
		}
		fieldType, _ := p.GoType(message, field)
		rawFieldType := strings.Trim(fieldType, "[]*")
		// Has to be ORMable
		if _, exists := convertibleTypes[rawFieldType]; !exists {
			continue
		}

		// Prep the filter for the child objects of this type
		keys := p.findAssociationKeys(message, typeNames[typeName], field)
		childFKeyTypeName := ""
		fieldTypeName := p.TypeName(typeNames[rawFieldType])
		p.P(`filterObj`, rawFieldType, ` := `, fieldTypeName, `ORM{}`)
		for k, v := range keys {
			for _, childField := range typeNames[rawFieldType].GetField() {
				if strings.EqualFold(generator.CamelCase(childField.GetName()), k) {
					childFKeyTypeName, _ = p.GoType(message, childField)
					k = generator.CamelCase(childField.GetName())
					break
				}
			}
			if typeNames[rawFieldType].Options != nil {
				if opts := getMessageOptions(typeNames[rawFieldType]); opts != nil {
					for _, field := range opts.Include {
						if strings.EqualFold(generator.CamelCase(field.GetName()), k) {
							childFKeyTypeName = field.GetType()
							k = generator.CamelCase(field.GetName())
							break
						}
					}
					if opts.GetMultiAccount() && k == "AccountID" {
						childFKeyTypeName = "string"
					}
				}
			}
			if childFKeyTypeName == "" {
				p.Fail(`Child type`, rawFieldType, `seems to have no foreign key field`,
					`linking it to parent type`, typeName, `: expected field`, k, `in`,
					rawFieldType, `associated with field`, v, `in`, typeName)
			}
			// If we accidentally delete without a set PK in our filter, everything might go
			if strings.Contains(childFKeyTypeName, "*") {
				p.P(`if ormObj.`, v, ` == nil || *ormObj.`, v, ` == `,
					guessZeroValue(childFKeyTypeName), ` {`)
			} else {
				p.P(`if ormObj.`, v, ` == `, guessZeroValue(childFKeyTypeName), ` {`)
			}
			p.P(`return nil, errors.New("Can't do overwriting update with no '`, v,
				`' value for FK of field '`, generator.CamelCase(field.GetName()), `'")`)
			p.P(`}`)

			p.P(`filterObj`, rawFieldType, `.`, k, ` = ormObj.`, v)
		}
		if opts := getMessageOptions(typeNames[rawFieldType]); opts != nil && opts.GetMultiAccount() {
			p.P("accountID, err := auth.GetAccountID(ctx, nil)")
			p.P("if err != nil {")
			p.P("return nil, err")
			p.P("}")
			p.P(`filterObj`, rawFieldType, `.AccountID = accountID`)
			usedAccountID = true
		}

		p.P(`if err = db.Where(filterObj`, rawFieldType, `).Delete(`,
			strings.Trim(fieldType, "[]*"), `{}).Error; err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
	}
	return usedAccountID
}

func (p *OrmPlugin) generateStrictUpdateHandler(message *generator.Descriptor) {
	typeName := p.TypeName(message)
	p.P(`// DefaultStrictUpdate`, typeName, ` clears first level 1:many children and then executes a gorm update call`)
	p.P(`func DefaultStrictUpdate`, typeName, `(ctx context.Context, in *`,
		typeName, `, db *`, p.gormPkgName, `.DB) (*`, typeName, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, fmt.Errorf("Nil argument to DefaultCascadedUpdate`, typeName, `")`)
	p.P(`}`)
	p.P(`ormObj, err := in.ToOrm()`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	usedAccountID := p.removeChildAssociations(message)
	if opts := getMessageOptions(message); opts != nil && opts.GetMultiAccount() {
		if !usedAccountID {
			p.P("accountID, err := auth.GetAccountID(ctx, nil)")
			p.P("if err != nil {")
			p.P("return nil, err")
			p.P("}")
		}
		p.P(`db = db.Where(&`, typeName, `ORM{AccountID: accountID})`)
	}
	p.P(`if err = db.Save(&ormObj).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse, err := ormObj.ToPB()`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`return &pbResponse, nil`)
	p.P(`}`)
	p.P()
}
