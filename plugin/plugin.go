package plugin

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	gorm "github.com/infobloxopen/protoc-gen-gorm/options"
	jgorm "github.com/jinzhu/gorm"
	"github.com/jinzhu/inflection"
)

// ORMable types, only for existence checks
var convertibleTypes = make(map[string]struct{})

// All message objects
var typeNames = make(map[string]*generator.Descriptor)

const (
	typeMessage = 11
	typeEnum    = 14

	protoTypeTimestamp = "Timestamp" // last segment, first will be *google_protobufX
	protoTypeUUID      = "UUIDValue"
)

var wellKnownTypes = map[string]string{
	"StringValue": "*string",
	"DoubleValue": "*double",
	"FloatValue":  "*float",
	"Int32Value":  "*int32",
	"Int64Value":  "*int64",
	"UInt32Value": "*uint32",
	"UInt64Value": "*uint64",
	"BoolValue":   "*bool",
	//  "BytesValue" : "*[]byte",
}

type OrmableType struct {
	Name         string
	MultiAccount bool
	Fields       map[string]*Field
}

type Field struct {
	Type    string
	Options *gorm.GormFieldOptions
}

var ormableTypes = make(map[string]*OrmableType)

// OrmPlugin implements the plugin interface and creates GORM code from .protos
type OrmPlugin struct {
	*generator.Generator
	wktPkgName  string
	gormPkgName string
	lftPkgName  string // 'Locally Famous Types', used for collection operators
	usingUUID   bool
	usingTime   bool
	usingAuth   bool
	usingGRPC   bool
}

func (p *OrmPlugin) parseBasicFields(msg *generator.Descriptor) {
	typeName := generator.CamelCaseSlice(msg.TypeName())
	ormable := ormableTypes[typeName]
	ormable.Name = fmt.Sprintf("%sORM", typeName)
	if getMessageOptions(msg).GetMultiAccount() {
		ormable.MultiAccount = true
	}
	for _, field := range msg.GetField() {
		fieldOpts := getFieldOptions(field)
		if fieldOpts.GetDrop() {
			continue
		}
		fieldName := generator.CamelCase(field.GetName())
		fieldType, _ := p.GoType(msg, field)
		if *(field.Type) == typeEnum {
			fieldType = "int32"
		} else if *(field.Type) == typeMessage {
			//Check for WKTs or fields of nonormable types
			parts := strings.Split(fieldType, ".")
			rawType := parts[len(parts)-1]
			if v, exists := wellKnownTypes[rawType]; exists {
				p.RecordTypeUse(".google.protobuf.StringValue")
				p.wktPkgName = strings.Trim(parts[0], "*")
				fieldType = v
			} else if rawType == protoTypeUUID {
				fieldType = "*uuid.UUID"
				p.usingUUID = true
			} else if rawType == protoTypeTimestamp {
				p.usingTime = true
				fieldType = "time.Time"
			} else {
				continue
			}
		}
		ormable.Fields[fieldName] = &Field{Type: fieldType, Options: fieldOpts}
	}
	if ormable.MultiAccount {
		if accID, ok := ormable.Fields["AccountID"]; !ok {
			ormable.Fields["AccountID"] = &Field{Type: "string"}
		} else {
			if accID.Type != "string" {
				p.Fail("Cannot include AccountID field into ", ormable.Name, " as it already exists there with a different type.")
			}
		}
	}
	for _, field := range getMessageOptions(msg).GetInclude() {
		if _, ok := ormable.Fields[field.GetName()]; !ok {
			ormable.Fields[field.GetName()] = &Field{Type: field.GetType(), Options: &gorm.GormFieldOptions{Tag: field.GetTag()}}
		} else {
			p.Fail("Cannot include ", field.GetName(), " field into ", ormable.Name, " as it aready exists there.")
		}
	}
}

func (p *OrmPlugin) isOrmable(typeName string) bool {
	_, ok := ormableTypes[strings.Trim(typeName, "[]*")]
	return ok
}

func (p *OrmPlugin) parseAssociations(msg *generator.Descriptor) {
	typeName := generator.CamelCaseSlice(msg.TypeName())
	ormable := ormableTypes[typeName]
	for _, field := range msg.GetField() {
		fieldOpts := getFieldOptions(field)
		if fieldOpts == nil {
			fieldOpts = &gorm.GormFieldOptions{}
		}
		if fieldOpts.GetDrop() {
			continue
		}
		fieldName := generator.CamelCase(field.GetName())
		fieldType, _ := p.GoType(msg, field)
		if p.isOrmable(fieldType) {
			if field.IsRepeated() {
				assocType := strings.Trim(fieldType, "[]*")
				hasMany := fieldOpts.GetHasMany()
				if hasMany == nil {
					hasMany = &gorm.HasManyOptions{}
					fieldOpts.HasMany = hasMany
				}
				var assocKey *Field
				var assocKeyName string
				if assocKeyName = hasMany.GetAssociationForeignkey(); assocKeyName == "" {
					assocKeyName, assocKey = p.findPrimaryKey(ormable)
					hasMany.AssociationForeignkey = &assocKeyName
				} else {
					assocKey = ormable.Fields[assocKeyName]
				}
				foreignKey := &Field{Type: assocKey.Type}
				var foreignKeyName string
				if foreignKeyName = hasMany.GetForeignkey(); foreignKeyName == "" {
					foreignKeyName = fmt.Sprintf(typeName + assocKeyName)
					hasMany.Foreignkey = &foreignKeyName
				}
				if exField, ok := ormableTypes[assocType].Fields[foreignKeyName]; !ok {
					ormableTypes[assocType].Fields[foreignKeyName] = foreignKey
				} else {
					if exField.Type != foreignKey.Type {
						p.Fail("Cannot include ", foreignKeyName, " field into ", assocType, " as it already exists there with a different type.")
					}
				}
				if posField := hasMany.GetPositionField(); posField != "" {
					if exField, ok := ormableTypes[assocType].Fields[posField]; !ok {
						ormableTypes[assocType].Fields[posField] = &Field{Type: "int"}
					} else {
						if exField.Type != "int" {
							p.Fail("Cannot include ", posField, " field into ", assocType, " as it already exists there with a different type.")
						}
					}
				}
				fieldType = fmt.Sprintf("[]*%sORM", strings.Trim(fieldType, "[]*"))

			} else {
				assocType := strings.Trim(fieldType, "[]*")
				hasOne := fieldOpts.GetHasOne()
				if hasOne == nil {
					hasOne = &gorm.HasOneOptions{}
					fieldOpts.HasOne = hasOne
				}
				var assocKey *Field
				var assocKeyName string
				if assocKeyName = hasOne.GetAssociationForeignkey(); assocKeyName == "" {
					assocKeyName, assocKey = p.findPrimaryKey(ormable)
					hasOne.AssociationForeignkey = &assocKeyName
				} else {
					assocKey = ormable.Fields[assocKeyName]
				}
				foreignKey := &Field{Type: assocKey.Type}
				var foreignKeyName string
				if foreignKeyName = hasOne.GetForeignkey(); foreignKeyName == "" {
					foreignKeyName = fmt.Sprintf(typeName + assocKeyName)
					hasOne.Foreignkey = &foreignKeyName
				}
				if exField, ok := ormableTypes[assocType].Fields[foreignKeyName]; !ok {
					ormableTypes[assocType].Fields[foreignKeyName] = foreignKey
				} else {
					if exField.Type != foreignKey.Type {
						p.Fail("Cannot include ", foreignKeyName, " field into ", assocType, " as it already exists there with a different type.")
					}
				}
				fieldType = fmt.Sprintf("*%sORM", strings.Trim(fieldType, "*"))
			}
			ormable.Fields[fieldName] = &Field{Type: fieldType, Options: fieldOpts}
		}
	}
}

func (p *OrmPlugin) findPrimaryKey(ormable *OrmableType) (string, *Field) {
	for fieldName, field := range ormable.Fields {
		if field.Options != nil && field.Options.Tag != nil {
			if field.Options.Tag.GetPrimaryKey() {
				return fieldName, field
			}
		}
	}
	for fieldName, field := range ormable.Fields {
		if strings.ToLower(fieldName) == "id" {
			return fieldName, field
		}
	}
	return "", nil
}

func (p *OrmPlugin) generateOrmable(ormable *OrmableType) {
	p.P(`type `, ormable.Name, ` struct {`)
	for fieldName, field := range ormable.Fields {
		p.P(fieldName, ` `, field.Type, p.renderGormTag(field.Options))
	}
	p.P(`}`)
}

func (p *OrmPlugin) renderGormTag(opts *gorm.GormFieldOptions) string {
	if opts == nil {
		return ""
	}
	res := ""
	if tag := opts.GetTag(); tag != nil {
		if tag.Column != nil {
			res += fmt.Sprintf("column:%s;", tag.GetColumn())
		}
		if tag.Type != nil {
			res += fmt.Sprintf("type:%s;", string(tag.GetType()))
		}
		if tag.Size_ != nil {
			res += fmt.Sprintf("size:%s;", string(tag.GetSize_()))
		}
		if tag.Precision != nil {
			res += fmt.Sprintf("precision:%s;", string(tag.GetPrecision()))
		}
		if tag.GetPrimaryKey() {
			res += "primary_key;"
		}
		if tag.GetUnique() {
			res += "unique;"
		}
		if tag.Default != nil {
			res += fmt.Sprintf("column:%s;", tag.GetDefault())
		}
		if tag.GetNotNull() {
			res += "not null;"
		}
		if tag.GetAutoIncrement() {
			res += "auto_increment;"
		}
		if tag.Index != nil {
			if tag.GetIndex() == "" {
				res += "index;"
			} else {
				res += fmt.Sprintf("index:%s;", tag.GetIndex())
			}
		}
		if tag.UniqueIndex != nil {
			if tag.GetUniqueIndex() == "" {
				res += "unique_index;"
			} else {
				res += fmt.Sprintf("unique_index:%s;", tag.GetUniqueIndex())
			}
		}
		if tag.GetEmbedded() {
			res += "embedded;"
		}
		if tag.EmbeddedPrefix != nil {
			res += fmt.Sprintf("embedded_prefix:%s;", tag.GetEmbeddedPrefix())
		}
		if tag.GetIgnore() {
			res += "-;"
		}
	}
	if hasMany := opts.GetHasMany(); hasMany != nil {
		if hasMany.Foreignkey != nil {
			res += fmt.Sprintf("foreignkey:%s;", hasMany.GetForeignkey())
		}
		if hasMany.AssociationForeignkey != nil {
			res += fmt.Sprintf("association_foreignkey:%s;", hasMany.GetAssociationForeignkey())
		}
	} else if hasOne := opts.GetHasOne(); hasOne != nil {
		if hasOne.Foreignkey != nil {
			res += fmt.Sprintf("foreignkey:%s;", hasOne.GetForeignkey())
		}
		if hasOne.AssociationForeignkey != nil {
			res += fmt.Sprintf("association_foreignkey:%s;", hasOne.GetAssociationForeignkey())
		}
	}
	if res == "" {
		return ""
	}
	return fmt.Sprintf("`gorm:\"%s\"`", res)
}

// Name identifies the plugin
func (p *OrmPlugin) Name() string {
	return "gorm"
}

// Init is called once after data structures are built but before
// code generation begins.
func (p *OrmPlugin) Init(g *generator.Generator) {
	p.Generator = g
}

// Generate produces the code generated by the plugin for this file,
// except for the imports, by calling the generator's methods P, In, and Out.
func (p *OrmPlugin) Generate(file *generator.FileDescriptor) {
	p.resetImports()
	// Preload just the types we'll be creating
	for _, msg := range file.Messages() {
		// We don't want to bother with the MapEntry stuff
		if msg.DescriptorProto.GetOptions().GetMapEntry() {
			continue
		}
		typeName := generator.CamelCaseSlice(msg.TypeName())
		typeNames[typeName] = msg
		if opts := getMessageOptions(msg); opts != nil && *opts.Ormable {
			convertibleTypes[typeName] = struct{}{}
			ormable := new(OrmableType)
			ormable.Fields = make(map[string]*Field)
			ormableTypes[typeName] = ormable
		}
	}
	for _, msg := range file.Messages() {
		typeName := generator.CamelCaseSlice(msg.TypeName())
		if _, exists := convertibleTypes[typeName]; !exists {
			continue
		}
		p.parseBasicFields(msg)
	}
	for _, msg := range file.Messages() {
		typeName := generator.CamelCaseSlice(msg.TypeName())
		if _, exists := convertibleTypes[typeName]; !exists {
			continue
		}
		p.parseAssociations(msg)
	}
	for _, msg := range file.Messages() {
		// We don't want to bother with the MapEntry stuff
		if msg.DescriptorProto.GetOptions().GetMapEntry() {
			continue
		}
		typeName := generator.CamelCaseSlice(msg.TypeName())
		if _, exists := convertibleTypes[typeName]; !exists {
			continue
		}
		// Create the orm object definitions and the converter functions
		p.generateOrmable(ormableTypes[typeName])
		p.generateTableNameFunction(msg)
		p.generateConvertFunctions(msg)
		p.generateHookInterfaces(msg)
	}

	p.P()
	p.P(`////////////////////////// CURDL for objects`)
	p.generateDefaultHandlers(file)
	p.P()

	p.generateDefaultServer(file)
}

// generateTableNameFunction the function to set the gorm table name
// back to gorm default, removing "ORM" suffix
func (p *OrmPlugin) generateTableNameFunction(message *generator.Descriptor) {
	typeName := p.TypeName(message)

	p.P(`// TableName overrides the default tablename generated by GORM`)
	p.P(`func (`, typeName, `ORM) TableName() string {`)

	tableName := inflection.Plural(jgorm.ToDBName(message.GetName()))
	if opts := getMessageOptions(message); opts != nil && opts.Table != nil {
		tableName = opts.GetTable()
	}
	p.P(`return "`, tableName, `"`)
	p.P(`}`)
}

// generateMapFunctions creates the converter functions
func (p *OrmPlugin) generateConvertFunctions(message *generator.Descriptor) {
	typeName := p.TypeName(message)

	///// To Orm
	p.P(`// ToORM runs the BeforeToORM hook if present, converts the fields of this`)
	p.P(`// object to ORM format, runs the AfterToORM hook, then returns the ORM object`)
	p.P(`func (m *`, typeName, `) ToORM (ctx context.Context) (`, typeName, `ORM, error) {`)
	p.P(`to := `, typeName, `ORM{}`)
	p.P(`var err error`)
	p.P(`if prehook, ok := interface{}(m).(`, typeName, `WithBeforeToORM); ok {`)
	p.P(`if err = prehook.BeforeToORM(ctx, &to); err != nil {`)
	p.P(`return to, err`)
	p.P(`}`)
	p.P(`}`)
	for _, field := range message.Field {
		// Checking if field is skipped
		if field.Options != nil {
			v, err := proto.GetExtension(field.Options, gorm.E_Field)
			opts, valid := v.(*gorm.GormFieldOptions)
			if err == nil && valid && opts != nil {
				if opts.Drop != nil && *opts.Drop {
					p.P(`// Skipping field: `, generator.CamelCase(field.GetName()))
					continue
				}
			}
		}
		p.generateFieldConversion(message, field, true)
	}
	p.P(`if posthook, ok := interface{}(m).(`, typeName, `WithAfterToORM); ok {`)
	p.P(`err = posthook.AfterToORM(ctx, &to)`)
	p.P(`}`)
	p.P(`return to, err`)
	p.P(`}`)

	p.P()
	///// To Pb
	p.P(`// ToPB runs the BeforeToPB hook if present, converts the fields of this`)
	p.P(`// object to PB format, runs the AfterToPB hook, then returns the PB object`)
	p.P(`func (m *`, typeName, `ORM) ToPB (ctx context.Context) (`,
		typeName, `, error) {`)
	p.P(`to := `, typeName, `{}`)
	p.P(`var err error`)
	p.P(`if prehook, ok := interface{}(m).(`, typeName, `WithBeforeToPB); ok {`)
	p.P(`if err = prehook.BeforeToPB(ctx, &to); err != nil {`)
	p.P(`return to, err`)
	p.P(`}`)
	p.P(`}`)
	for _, field := range message.Field {
		// Checking if field is skipped
		if field.Options != nil {
			v, err := proto.GetExtension(field.Options, gorm.E_Field)
			opts, valid := v.(*gorm.GormFieldOptions)
			if err == nil && valid && opts != nil {
				if opts.Drop != nil && *opts.Drop {
					p.P(`// Skipping field: `, generator.CamelCase(field.GetName()))
					continue
				}
			}
		}
		p.generateFieldConversion(message, field, false)
	}
	p.P(`if posthook, ok := interface{}(m).(`, typeName, `WithAfterToPB); ok {`)
	p.P(`err = posthook.AfterToPB(ctx, &to)`)
	p.P(`}`)
	p.P(`return to, err`)
	p.P(`}`)
}

// Output code that will convert a field to/from orm.
func (p *OrmPlugin) generateFieldConversion(message *generator.Descriptor, field *descriptor.FieldDescriptorProto, toORM bool) error {
	fieldName := generator.CamelCase(field.GetName())
	fieldType, _ := p.GoType(message, field)
	if field.IsRepeated() { // Repeated Object ----------------------------------
		if _, exists := convertibleTypes[strings.Trim(fieldType, "[]*")]; exists { // Repeated ORMable type
			//fieldType = strings.Trim(fieldType, "[]*")

			p.P(`for _, v := range m.`, fieldName, ` {`)
			p.P(`if v != nil {`)
			if toORM {
				p.P(`if temp`, fieldName, `, cErr := v.ToORM(ctx); cErr == nil {`)
			} else {
				p.P(`if temp`, fieldName, `, cErr := v.ToPB(ctx); cErr == nil {`)
			}
			p.P(`to.`, fieldName, ` = append(to.`, fieldName, `, &temp`, fieldName, `)`)
			p.P(`} else {`)
			p.P(`return to, cErr`)
			p.P(`}`)
			p.P(`} else {`)
			p.P(`to.`, fieldName, ` = append(to.`, fieldName, `, nil)`)
			p.P(`}`)
			p.P(`}`) // end repeated for
		} else {
			p.P(`// Repeated type `, fieldType, ` is not an ORMable message type`)
		}
	} else if *(field.Type) == typeEnum { // Singular Enum, which is an int32 ---
		if toORM {
			p.P(`to.`, fieldName, ` = int32(m.`, fieldName, `)`)
		} else {
			p.P(`to.`, fieldName, ` = `, fieldType, `(m.`, fieldName, `)`)
		}
	} else if *(field.Type) == typeMessage { // Singular Object -------------
		//Check for WKTs
		parts := strings.Split(fieldType, ".")
		coreType := parts[len(parts)-1]
		// Type is a WKT, convert to/from as ptr to base type
		if _, exists := wellKnownTypes[coreType]; exists { // Singular WKT -----
			if toORM {
				p.P(`if m.`, fieldName, ` != nil {`)
				p.P(`v := m.`, fieldName, `.Value`)
				p.P(`to.`, fieldName, ` = &v`)
				p.P(`}`)
			} else {
				p.P(`if m.`, fieldName, ` != nil {`)
				p.P(`to.`, fieldName, ` = &`, p.wktPkgName, ".", coreType,
					`{Value: *m.`, fieldName, `}`)
				p.P(`}`)
			}
		} else if coreType == protoTypeUUID { // Singular UUID type ------------
			if toORM {
				p.P(`if m.`, fieldName, ` != nil {`)
				p.P(`tempUUID, uErr := uuid.FromString(m.`, fieldName, `.Value)`)
				p.P(`if uErr != nil {`)
				p.P(`return to, uErr`)
				p.P(`}`)
				p.P(`to.`, fieldName, ` = &tempUUID`)
				p.P(`}`)
			} else {
				p.P(`if m.`, fieldName, ` != nil {`)
				p.P(`to.`, fieldName, ` = &gtypes.UUIDValue{Value: m.`, fieldName, `.String()}`)
				p.P(`}`)
			}
		} else if coreType == protoTypeTimestamp { // Singular WKT Timestamp ---
			if toORM {
				p.P(`if m.`, fieldName, ` != nil {`)
				p.P(`if to.`, fieldName, `, err = ptypes.Timestamp(m.`, fieldName, `); err != nil {`)
				p.P(`return to, err`)
				p.P(`}`)
				p.P(`}`)
			} else {
				p.P(`if to.`, fieldName, `, err = ptypes.TimestampProto(m.`, fieldName, `); err != nil {`)
				p.P(`return to, err`)
				p.P(`}`)
			}
		} else if _, exists := convertibleTypes[strings.Trim(fieldType, "[]*")]; exists {
			// Not a WKT, but a type we're building converters for
			fieldType = strings.Trim(fieldType, "*")
			p.P(`if m.`, fieldName, ` != nil {`)
			if toORM {
				p.P(`temp`, fieldType, `, err := m.`, fieldName, `.ToORM (ctx)`)
			} else {
				p.P(`temp`, fieldType, `, err := m.`, fieldName, `.ToPB (ctx)`)
			}
			p.P(`if err != nil {`)
			p.P(`return to, err`)
			p.P(`}`)
			p.P(`to.`, fieldName, ` = &temp`, fieldType)
			p.P(`}`)
		}
	} else { // Singular raw ----------------------------------------------------
		p.P(`to.`, fieldName, ` = m.`, fieldName)
	}
	return nil
}

func (p *OrmPlugin) generateHookInterfaces(message *generator.Descriptor) {
	typeName := p.TypeName(message)
	p.P(`// The following are interfaces you can implement for special behavior during ORM/PB conversions`)
	p.P(`// of type `, typeName, ` the arg will be the target, the caller the one being converted from`)
	p.P()
	for _, desc := range [][]string{
		{"BeforeToORM", typeName + "ORM", " called before default ToORM code"},
		{"AfterToORM", typeName + "ORM", " called after default ToORM code"},
		{"BeforeToPB", typeName, " called before default ToPB code"},
		{"AfterToPB", typeName, " called after default ToPB code"},
	} {
		p.P(`// `, typeName, desc[0], desc[2])
		p.P(`type `, typeName, `With`, desc[0], ` interface {`)
		p.P(desc[0], `(context.Context, *`, desc[1], `) error`)
		p.P(`}`)
		p.P()
	}
}
