package plugin

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/infobloxopen/protoc-gen-gorm/options"
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
		unlintedTypeName := generator.CamelCaseSlice(msg.TypeName())
		typeNames[unlintedTypeName] = msg
		if opts := getMessageOptions(msg); opts != nil && *opts.Ormable {
			convertibleTypes[unlintedTypeName] = struct{}{}
		}
	}
	for _, msg := range file.Messages() {
		// We don't want to bother with the MapEntry stuff
		if msg.DescriptorProto.GetOptions().GetMapEntry() {
			continue
		}
		unlintedTypeName := generator.CamelCaseSlice(msg.TypeName())
		if _, exists := convertibleTypes[unlintedTypeName]; !exists {
			continue
		}
		// Create the orm object definitions and the converter functions
		p.generateMessages(msg)
		p.generateConvertFunctions(msg)
	}

	p.P()
	p.P(`////////////////////////// CURDL for objects`)
	p.generateDefaultHandlers(file)
	p.P()

	p.generateDefaultServer(file)
}

func (p *OrmPlugin) generateMessages(message *generator.Descriptor) {
	typeName := p.TypeName(message)
	p.generateMessageHead(message)
	for _, field := range message.Field {
		fieldName := generator.CamelCase(field.GetName())
		fieldType, _ := p.GoType(message, field)
		var tagString string
		if field.Options != nil {
			v, err := proto.GetExtension(field.Options, gorm.E_Field)
			opts, valid := v.(*gorm.GormFieldOptions)
			if err == nil && valid && opts != nil {
				if opts.Drop != nil && *opts.Drop {
					p.P(`// Skipping field from proto option: `, fieldName)
					continue
				}
				tags := v.(*gorm.GormFieldOptions).Tags
				if tags != nil {
					tagString = fmt.Sprintf("`%s`", *tags)
				}
			}
		}
		if _, exists := convertibleTypes[strings.Trim(fieldType, "[]*")]; field.IsRepeated() && !exists {
			p.P(`// The non-ORMable repeated field "`, fieldName, `" can't be included`)
			continue
		} else if *(field.Type) == typeEnum {
			fieldType = "int32"
		} else if *(field.Type) == typeMessage {
			//Check for WKTs or fields of nonormable types
			parts := strings.Split(fieldType, ".")
			rawType := parts[len(parts)-1]
			if v, exists := wellKnownTypes[rawType]; exists {
				p.RecordTypeUse(".google.protobuf.StringValue")
				p.wktPkgName = strings.Trim(parts[0], "*")
				fieldType = v
			} else if rawType == "Empty" {
				p.P("// Empty type has no ORM equivalency")
				continue
			} else if rawType == protoTypeUUID {
				fieldType = "*uuid.UUID"
				p.usingUUID = true
				if tagString == "" {
					tagString = "`sql:\"type:uuid\"`"
				} else if strings.Contains(strings.ToLower(tagString), "sql:") &&
					!strings.Contains(strings.ToLower(tagString), "type:") { // sql tag already there

					index := strings.Index(tagString, "sql:")
					tagString = fmt.Sprintf("%stype:uuid;%s", tagString[:index+6],
						tagString[index+6:])
				} else { // no sql tag yet
					tagString = fmt.Sprintf("`sql:\"type:uuid\" %s", tagString[1:])
				}

			} else if rawType == protoTypeTimestamp {
				p.usingTime = true
				fieldType = "time.Time"
			} else if _, exists := convertibleTypes[strings.Trim(fieldType, "[]*")]; !exists {
				p.P("// Skipping type ", fieldType, ", not tagged as ormable")
				continue
			} else {
				if field.IsRepeated() {
					fieldType = fmt.Sprintf("[]*%sORM", strings.Trim(fieldType, "[]*"))
				} else {
					fieldType = fmt.Sprintf("*%sORM", strings.Trim(fieldType, "*"))
				}
				// Insert the foreign key if not present,
				if tagString == "" {
					tagString = fmt.Sprintf("`gorm:\"foreignkey:%sId\"`", typeName)
				} else if !strings.Contains(strings.ToLower(tagString), "foreignkey") {
					if strings.Contains(strings.ToLower(tagString), "gorm:") { // gorm tag already there
						index := strings.Index(tagString, "gorm:")
						tagString = fmt.Sprintf("%sforeignkey:%sId;%s", tagString[:index+6],
							typeName, tagString[index+6:])
					} else { // no gorm tag yet
						tagString = fmt.Sprintf("`gorm:\"foreignkey:%sId\" %s", typeName, tagString[1:])
					}
				}
			}
		}
		p.P(fieldName, " ", fieldType, tagString)
	}
	p.P(`}`)

	p.generateTableNameFunction(message)
}

// generateMessageComment pulls from the proto file comment or creates a
// default comment if none is present there, and writes the signature and
// fields from the proto file options
func (p *OrmPlugin) generateMessageHead(message *generator.Descriptor) {
	typeName := p.TypeName(message)
	typeNameOrm := fmt.Sprintf("%sORM", typeName)
	// Check for a comment, generate a default if none is provided
	comment := p.Comments(message.Path())
	commentStart := strings.Split(strings.Trim(comment, " "), " ")[0]
	if generator.CamelCase(commentStart) == typeName || commentStart == typeNameOrm {
		comment = strings.Replace(comment, commentStart, typeNameOrm, 1)
	} else if len(comment) == 0 {
		comment = fmt.Sprintf(" %s no comment was provided for message type", typeNameOrm)
	} else if len(strings.Replace(comment, " ", "", -1)) > 0 {
		comment = fmt.Sprintf(" %s %s", typeNameOrm, comment)
	} else {
		comment = fmt.Sprintf(" %s no comment provided", typeNameOrm)
	}
	p.P(`//`, comment)
	p.P(`type `, typeNameOrm, ` struct {`)
	// Checking for any ORM only fields specified by option (gorm.opts).include
	if opts := getMessageOptions(message); opts != nil {
		for _, field := range opts.Include {
			tagString := ""
			if field.Tags != nil {
				tagString = fmt.Sprintf("`%s`", field.GetTags())
			}
			p.P(generator.CamelCase(*field.Name), ` `, field.Type, ` `, tagString)
		}
		if opts.GetMultiTenant() {
			p.P("TenantID string")
		}
	}
}

// generateTableNameFunction the function to set the gorm table name
// back to gorm default, removing "ORM" suffix
func (p *OrmPlugin) generateTableNameFunction(message *generator.Descriptor) {
	typeName := p.TypeName(message)

	p.P(`// TableName overrides the default tablename generated by GORM`)
	p.P(`func (`, typeName, `ORM) TableName() string {`)

	tableName := inflection.Plural(jgorm.ToDBName(message.GetName()))
	if opts := getMessageOptions(message); opts == nil {
		tableName = opts.GetTable()
	}
	p.P(`return "`, tableName, `"`)
	p.P(`}`)
}

// generateMapFunctions creates the converter functions
func (p *OrmPlugin) generateConvertFunctions(message *generator.Descriptor) {
	typeName := p.TypeName(message)

	///// To Orm
	p.P(`// Convert`, typeName, `ToORM takes a pb object and returns an orm object`)
	p.P(`func Convert`, typeName, `ToORM (from `,
		typeName, `) (`, typeName, `ORM, error) {`)
	p.P(`to := `, typeName, `ORM{}`)
	p.P(`var err error`)
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
	p.P(`return to, err`)
	p.P(`}`)

	p.P()
	///// To Pb
	p.P(`// Convert`, typeName, `FromORM takes an orm object and returns a pb object`)
	p.P(`func Convert`, typeName, `FromORM (from `, typeName, `ORM) (`,
		typeName, `, error) {`)
	p.P(`to := `, typeName, `{}`)
	p.P(`var err error`)
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
	p.P(`return to, err`)
	p.P(`}`)
}

// Output code that will convert a field to/from orm.
func (p *OrmPlugin) generateFieldConversion(message *generator.Descriptor, field *descriptor.FieldDescriptorProto, toORM bool) error {
	fieldName := generator.CamelCase(field.GetName())
	fieldType, _ := p.GoType(message, field)
	if field.IsRepeated() { // Repeated Object ----------------------------------
		if _, exists := convertibleTypes[strings.Trim(fieldType, "[]*")]; exists { // Repeated ORMable type
			fieldType = strings.Trim(fieldType, "[]*")
			dir := "From"
			if toORM {
				dir = "To"
			}

			p.P(`for _, v := range from.`, fieldName, ` {`)
			p.P(`if v != nil {`)
			p.P(`if temp`, fieldName, `, cErr := Convert`, fieldType, dir, `ORM (*v); cErr == nil {`)
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
			p.P(`to.`, fieldName, ` = int32(from.`, fieldName, `)`)
		} else {
			p.P(`to.`, fieldName, ` = `, fieldType, `(from.`, fieldName, `)`)
		}
	} else if *(field.Type) == typeMessage { // Singular Object -------------
		//Check for WKTs
		parts := strings.Split(fieldType, ".")
		coreType := parts[len(parts)-1]
		// Type is a WKT, convert to/from as ptr to base type
		if _, exists := wellKnownTypes[coreType]; exists { // Singular WKT -----
			if toORM {
				p.P(`if from.`, fieldName, ` != nil {`)
				p.P(`v := from.`, fieldName, `.Value`)
				p.P(`to.`, fieldName, ` = &v`)
				p.P(`}`)
			} else {
				p.P(`if from.`, fieldName, ` != nil {`)
				p.P(`to.`, fieldName, ` = &`, p.wktPkgName, ".", coreType,
					`{Value: *from.`, fieldName, `}`)
				p.P(`}`)
			}
		} else if coreType == protoTypeUUID { // Singular UUID type ------------
			if toORM {
				p.P(`if from.`, fieldName, ` != nil {`)
				p.P(`tempUUID, uErr := uuid.FromString(from.`, fieldName, `.Value)`)
				p.P(`if uErr != nil {`)
				p.P(`return to, uErr`)
				p.P(`}`)
				p.P(`to.`, fieldName, ` = &tempUUID`)
				p.P(`}`)
			} else {
				p.P(`if from.`, fieldName, ` != nil {`)
				p.P(`to.`, fieldName, ` = &gtypes.UUIDValue{Value: from.`, fieldName, `.String()}`)
				p.P(`}`)
			}
		} else if coreType == protoTypeTimestamp { // Singular WKT Timestamp ---
			if toORM {
				p.P(`if from.`, fieldName, ` != nil {`)
				p.P(`if to.`, fieldName, `, err = ptypes.Timestamp(from.`, fieldName, `); err != nil {`)
				p.P(`return to, err`)
				p.P(`}`)
				p.P(`}`)
			} else {
				p.P(`if to.`, fieldName, `, err = ptypes.TimestampProto(from.`, fieldName, `); err != nil {`)
				p.P(`return to, err`)
				p.P(`}`)
			}
		} else if _, exists := convertibleTypes[strings.Trim(fieldType, "[]*")]; exists {
			// Not a WKT, but a type we're building converters for
			fieldType = strings.Trim(fieldType, "*")
			dir := "From"
			if toORM {
				dir = "To"
			}
			p.P(`if from.`, fieldName, ` != nil {`)
			p.P(`temp`, fieldType, `, err := Convert`, fieldType, dir, `ORM (*from.`, fieldName, `)`)
			p.P(`if err != nil {`)
			p.P(`return to, err`)
			p.P(`}`)
			p.P(`to.`, fieldName, ` = &temp`, fieldType)
			p.P(`}`)
		}
	} else { // Singular raw ----------------------------------------------------
		p.P(`to.`, fieldName, ` = from.`, fieldName)
	}
	return nil
}
