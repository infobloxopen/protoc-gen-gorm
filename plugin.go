package main

import (
	"fmt"
	"reflect"
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

type ormPlugin struct {
	*generator.Generator
	wktPkgName  string
	gormPkgName string
	lftPkgName  string // 'Locally Famous Types', used for collection operators
	usingUUID   bool
	usingTime   bool
	usingAuth   bool
}

func (p *ormPlugin) GenerateImports(file *generator.FileDescriptor) {
	if p.gormPkgName != "" {
		p.PrintImport("context", "context")
		p.PrintImport("errors", "errors")
		p.PrintImport(p.gormPkgName, "github.com/jinzhu/gorm")
		p.PrintImport(p.lftPkgName, "github.com/Infoblox-CTO/ngp.api.toolkit/op/gorm")
	}
	if p.usingUUID {
		p.PrintImport("uuid", "github.com/satori/go.uuid")
		p.PrintImport("gtypes", "github.com/infobloxopen/protoc-gen-gorm/types")
	}
	if p.usingTime {
		p.PrintImport("time", "time")
		p.PrintImport("ptypes", "github.com/golang/protobuf/ptypes")
	}
	if p.usingAuth {
		p.PrintImport("auth", "github.com/Infoblox-CTO/ngp.api.toolkit/mw/auth")
	}
}

func (p *ormPlugin) Name() string {
	return "gorm"
}

func (p *ormPlugin) Init(g *generator.Generator) {
	p.Generator = g
}

func (p *ormPlugin) Generate(file *generator.FileDescriptor) {

	// Preload just the types we'll be creating
	for _, msg := range file.Messages() {
		// We don't want to bother with the MapEntry stuff
		if msg.DescriptorProto.GetOptions().GetMapEntry() {
			continue
		}
		unlintedTypeName := generator.CamelCaseSlice(msg.TypeName())
		typeNames[unlintedTypeName] = msg
		if msg.Options != nil {
			v, err := proto.GetExtension(msg.Options, gorm.E_Opts)
			opts := v.(*gorm.GormMessageOptions)
			if err == nil && opts != nil && *opts.Ormable {
				convertibleTypes[unlintedTypeName] = struct{}{}
			}
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
}

func (p *ormPlugin) generateMessages(message *generator.Descriptor) {
	typeNamePb, _, _ := getTypeNames(message)
	p.generateMessageComment(message)
	for _, field := range message.Field {
		fieldName := generator.CamelCase(field.GetName())
		fieldType, _ := p.GoType(message, field)
		var tagString string
		if field.Options != nil {
			v, _ := proto.GetExtension(field.Options, gorm.E_Field)
			if v != nil && v.(*gorm.GormFieldOptions) != nil {
				if v.(*gorm.GormFieldOptions).Drop != nil && *v.(*gorm.GormFieldOptions).Drop {
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
				p.RecordTypeUse(".google.protobuf.Empty")
				p.P("// Empty type has no ORM equivalency")
				continue
			} else if rawType == protoTypeUUID {
				fieldType = "uuid.UUID"
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
				fieldType = fmt.Sprintf("%sORM", lintName(fieldType))
				// Insert the foreign key if not present,
				if tagString == "" {
					tagString = fmt.Sprintf("`gorm:\"foreignkey:%sID\"`", lintName(typeNamePb))
				} else if !strings.Contains(strings.ToLower(tagString), "foreignkey") {
					if strings.Contains(strings.ToLower(tagString), "gorm:") { // gorm tag already there
						index := strings.Index(tagString, "gorm:")
						tagString = fmt.Sprintf("%sforeignkey:%sID;%s", tagString[:index+6],
							lintName(typeNamePb), tagString[index+6:])
					} else { // no gorm tag yet
						tagString = fmt.Sprintf("`gorm:\"foreignkey:%sID\" %s", lintName(typeNamePb), tagString[1:])
					}
				}
			}
		}
		p.P(lintName(fieldName), " ", fieldType, tagString)
	}
	p.P(`}`)

	p.generateTableNameFunction(message)
}

// Returns the pb, linted, and linted+ORM suffix type names for a given object
func getTypeNames(desc *generator.Descriptor) (string, string, string) {
	typeNamePb := generator.CamelCaseSlice(desc.TypeName())
	typeName := lintName(typeNamePb)
	typeNameOrm := fmt.Sprintf("%sORM", typeName)
	return typeNamePb, typeName, typeNameOrm
}

// generateMessageComment pulls from the proto file comment or creates a
// default comment if none is present there
func (p *ormPlugin) generateMessageComment(message *generator.Descriptor) {
	typeNamePb, _, typeNameOrm := getTypeNames(message)
	// Check for a comment, generate a default if none is provided
	comment := p.Comments(message.Path())
	commentStart := strings.Split(strings.Trim(comment, " "), " ")[0]
	if generator.CamelCase(commentStart) == typeNamePb || commentStart == typeNameOrm {
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
	if message.Options != nil {
		v, err := proto.GetExtension(message.Options, gorm.E_Opts)
		opts := v.(*gorm.GormMessageOptions)
		if err == nil && opts != nil {
			for _, field := range opts.Include {
				tagString := ""
				if field.Tags != nil {
					tagString = fmt.Sprintf("`%s`", field.GetTags())
				}
				p.P(lintName(generator.CamelCase(*field.Name)), ` `, field.Type, ` `, tagString)
			}
			if opts.GetMultiTenant() {
				p.P("TenantID string")
			}
		}
	}
}

// generateTableNameFunction the function to set the gorm table name
// back to gorm default, removing "ORM" suffix
func (p *ormPlugin) generateTableNameFunction(message *generator.Descriptor) {
	_, _, typeNameOrm := getTypeNames(message)

	p.P(`// TableName overrides the default tablename generated by GORM`)
	p.P(`func (`, typeNameOrm, `) TableName() string {`)

	tableName := inflection.Plural(jgorm.ToDBName(message.GetName()))
	if message.Options != nil {
		v, _ := proto.GetExtension(message.Options, gorm.E_Opts)
		if v != nil {
			opts := v.(*gorm.GormMessageOptions)
			if opts != nil && opts.Table != nil {
				tableName = opts.GetTable()
			}
		}
	}
	p.P(`return "`, tableName, `"`)
	p.P(`}`)
}

// generateMapFunctions creates the converter functions
func (p *ormPlugin) generateConvertFunctions(message *generator.Descriptor) {
	typeNamePb, typeNameBase, typeNameOrm := getTypeNames(message)

	///// To Orm
	p.P(`// Convert`, typeNameBase, `ToORM takes a pb object and returns an orm object`)
	p.P(`func Convert`, typeNameBase, `ToORM (from `,
		typeNamePb, `) (`, typeNameOrm, `, error) {`)
	p.P(`to := `, typeNameOrm, `{}`)
	p.P(`var err error`)
	for _, field := range message.Field {
		// Checking if field is skipped
		if field.Options != nil {
			v, err := proto.GetExtension(field.Options, gorm.E_Field)
			if err == nil && v.(*gorm.GormFieldOptions) != nil {
				if v.(*gorm.GormFieldOptions).Drop != nil && *v.(*gorm.GormFieldOptions).Drop {
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
	p.P(`// Convert`, typeNameBase, `FromORM takes an orm object and returns a pb object`)
	p.P(`func Convert`, typeNameBase, `FromORM (from `, typeNameOrm, `) (`,
		typeNamePb, `, error) {`)
	p.P(`to := `, typeNamePb, `{}`)
	p.P(`var err error`)
	for _, field := range message.Field {
		// Checking if field is skipped
		if field.Options != nil {
			v, err := proto.GetExtension(field.Options, gorm.E_Field)
			if err == nil && v.(*gorm.GormFieldOptions) != nil {
				if v.(*gorm.GormFieldOptions).Drop != nil && *v.(*gorm.GormFieldOptions).Drop {
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
func (p *ormPlugin) generateFieldConversion(message *generator.Descriptor, field *descriptor.FieldDescriptorProto, toORM bool) error {
	fieldName := generator.CamelCase(field.GetName())
	fromName := fieldName
	if toORM {
		fieldName = lintName(fromName)
	} else {
		fromName = lintName(fromName)
	}
	fieldType, _ := p.GoType(message, field)
	if field.IsRepeated() { // Repeated Object ----------------------------------
		if _, exists := convertibleTypes[strings.Trim(fieldType, "[]*")]; exists { // Repeated ORMable type
			p.P(`for _, v := range from.`, fromName, ` {`)

			fieldType = strings.Trim(fieldType, "[]*")
			fieldType = lintName(fieldType)
			dir := "From"
			if toORM {
				dir = "To"
			}
			p.P(`if v != nil {`)
			p.P(`if temp`, lintName(fieldName), `, cErr := Convert`, fieldType, dir, `ORM (*v); cErr == nil {`)
			p.P(`to.`, fieldName, ` = append(to.`, fieldName, `, &temp`, lintName(fieldName), `)`)
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
			p.P(`to.`, fieldName, ` = int32(from.`, fromName, `)`)
		} else {
			p.P(`to.`, fieldName, ` = `, fieldType, `(from.`, fromName, `)`)
		}
	} else if *(field.Type) == typeMessage { // Singular Object -------------
		//Check for WKTs
		parts := strings.Split(fieldType, ".")
		coreType := parts[len(parts)-1]
		// Type is a WKT, convert to/from as ptr to base type
		if _, exists := wellKnownTypes[coreType]; exists { // Singular WKT -----
			if toORM {
				p.P(`if from.`, fromName, ` != nil {`)
				p.P(`v := from.`, fromName, `.Value`)
				p.P(`to.`, fieldName, ` = &v`)
				p.P(`}`)
			} else {
				p.P(`if from.`, fromName, ` != nil {`)
				p.P(`to.`, fieldName, ` = &`, p.wktPkgName, ".", coreType,
					`{Value: *from.`, fromName, `}`)
				p.P(`}`)
			}
		} else if coreType == protoTypeUUID { // Singular UUID type ------------
			if toORM {
				p.P(`if from.`, fromName, ` != nil {`)
				p.P(`if to.`, fieldName, `, err = uuid.FromString(from.`,
					fromName, `.Value); err != nil {`)
				p.P(`return to, err`)
				p.P(`}`)
				p.P(`}`)
			} else {
				p.P(`to.`, fieldName, ` = &gtypes.UUIDValue{Value: from.`, fromName, `.String()}`)
			}
		} else if coreType == protoTypeTimestamp { // Singular WKT Timestamp ---
			if toORM {
				p.P(`if from.`, fromName, ` != nil {`)
				p.P(`if to.`, fieldName, `, err = ptypes.Timestamp(from.`, fromName, `); err != nil {`)
				p.P(`return to, err`)
				p.P(`}`)
				p.P(`}`)
			} else {
				p.P(`if to.`, fieldName, `, err = ptypes.TimestampProto(from.`, fromName, `); err != nil {`)
				p.P(`return to, err`)
				p.P(`}`)
			}
		} else if _, exists := convertibleTypes[strings.Trim(fieldType, "[]*")]; exists {
			// Not a WKT, but a type we're building converters for
			fieldType = strings.Trim(fieldType, "*")
			fieldType = lintName(fieldType)
			dir := "From"
			if toORM {
				dir = "To"
			}
			p.P(`if from.`, fromName, ` != nil {`)
			p.P(`if to.`, fieldName, `, err = Convert`, fieldType, dir, `ORM (from.`, fromName, `); err != nil {`)
			p.P(`return to, err`)
			p.P(`}`)
			p.P(`}`)
		}
	} else { // Singular raw ----------------------------------------------------
		p.P(`to.`, fieldName, ` = from.`, fromName)
	}
	return nil
}

func (p *ormPlugin) generateDefaultHandlers(file *generator.FileDescriptor) {
	for _, message := range file.Messages() {
		if message.Options != nil {
			v, err := proto.GetExtension(message.Options, gorm.E_Opts)
			if err != nil {
				continue
			}
			if opts := v.(*gorm.GormMessageOptions); opts == nil || !*opts.Ormable {
				continue
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

func (p *ormPlugin) GetMessageOptions(message *generator.Descriptor) *gorm.GormMessageOptions {
	if message.Options == nil {
		return nil
	}
	v, err := proto.GetExtension(message.Options, gorm.E_Opts)
	if err != nil {
		return nil
	}
	return v.(*gorm.GormMessageOptions)
}

func (p *ormPlugin) generateCreateHandler(message *generator.Descriptor) {
	typeNamePb, typeName, _ := getTypeNames(message)
	p.P(`// DefaultCreate`, typeName, ` executes a basic gorm create call`)
	p.P(`func DefaultCreate`, typeName, `(ctx context.Context, in *`,
		typeNamePb, `, db *`, p.gormPkgName, `.DB) (*`, typeNamePb, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, errors.New("Nil argument to DefaultCreate`, typeName, `")`)
	p.P(`}`)
	p.P(`ormObj, err := Convert`, typeName, `ToORM(*in)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	if opts := p.GetMessageOptions(message); opts != nil && opts.GetMultiTenant() {
		p.P("tenantID, tIDErr := auth.GetTenantID(ctx)")
		p.P("if tIDErr != nil {")
		p.P("return nil, tIDErr")
		p.P("}")
		p.P("ormObj.TenantID = tenantID")
	}
	p.P(`if err = db.Create(&ormObj).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse, err := Convert`, typeName, `FromORM(ormObj)`)
	p.P(`return &pbResponse, err`)
	p.P(`}`)
	p.P()
}

func (p *ormPlugin) generateReadHandler(message *generator.Descriptor) {
	typeNamePb, typeName, typeNameOrm := getTypeNames(message)
	p.P(`// DefaultRead`, typeName, ` executes a basic gorm read call`)
	p.P(`func DefaultRead`, typeName, `(ctx context.Context, in *`,
		typeNamePb, `, db *`, p.gormPkgName, `.DB) (*`, typeNamePb, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, errors.New("Nil argument to DefaultRead`, typeName, `")`)
	p.P(`}`)
	p.P(`ormParams, err := Convert`, typeName, `ToORM(*in)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	if opts := p.GetMessageOptions(message); opts != nil && opts.GetMultiTenant() {
		p.P("tenantID, tIDErr := auth.GetTenantID(ctx)")
		p.P("if tIDErr != nil {")
		p.P("return nil, tIDErr")
		p.P("}")
		p.P("ormParams.TenantID = tenantID")
	}
	p.P(`ormResponse := `, typeNameOrm, `{}`)
	p.P(`if err = db.Set("gorm:auto_preload", true).Where(&ormParams).First(&ormResponse).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse, err := Convert`, typeName, `FromORM(ormResponse)`)
	p.P(`return &pbResponse, err`)
	p.P(`}`)
	p.P()
}

func (p *ormPlugin) generateUpdateHandler(message *generator.Descriptor) {
	typeNamePb, typeName, _ := getTypeNames(message)

	hasIDField := false
	for _, field := range message.GetField() {
		if strings.ToLower(field.GetName()) == "id" {
			hasIDField = true
			break
		}
	}
	isMultiTenant := false
	if opts := p.GetMessageOptions(message); opts != nil && opts.GetMultiTenant() {
		isMultiTenant = true
	}

	if isMultiTenant && !hasIDField {
		p.P(fmt.Sprintf("// Cannot autogen DefaultUpdate%s: this is a multi-tenant table without an \"id\" field in the message.\n", typeName))
		return
	}

	p.P(`// DefaultUpdate`, typeName, ` executes a basic gorm update call`)
	p.P(`func DefaultUpdate`, typeName, `(ctx context.Context, in *`,
		typeNamePb, `, db *`, p.gormPkgName, `.DB) (*`, typeNamePb, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, errors.New("Nil argument to DefaultUpdate`, typeName, `")`)
	p.P(`}`)
	if isMultiTenant {
		p.P(fmt.Sprintf("if exists, err := DefaultRead%s(ctx, &%s{Id: in.GetId()}, db); err != nil {",
			typeName, typeName))
		p.P("return nil, err")
		p.P("} else if exists == nil {")
		p.P(fmt.Sprintf("return nil, errors.New(\"%s not found\")", typeName))
		p.P("}")
	}

	p.P(`ormObj, err := Convert`, typeName, `ToORM(*in)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`if err = db.Save(&ormObj).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse, err := Convert`, typeName, `FromORM(ormObj)`)
	p.P(`return &pbResponse, err`)
	p.P(`}`)
	p.P()
}

func (p *ormPlugin) generateDeleteHandler(message *generator.Descriptor) {
	typeNamePb, typeName, _ := getTypeNames(message)
	p.P(`func DefaultDelete`, typeName, `(ctx context.Context, in *`,
		typeNamePb, `, db *`, p.gormPkgName, `.DB) error {`)
	p.P(`if in == nil {`)
	p.P(`return errors.New("Nil argument to DefaultDelete`, typeName, `")`)
	p.P(`}`)
	p.P(`ormObj, err := Convert`, typeName, `ToORM(*in)`)
	p.P(`if err != nil {`)
	p.P(`return err`)
	p.P(`}`)
	if opts := p.GetMessageOptions(message); opts != nil && opts.GetMultiTenant() {
		p.P("tenantID, tIDErr := auth.GetTenantID(ctx)")
		p.P("if tIDErr != nil {")
		p.P("return tIDErr")
		p.P("}")
		p.P("ormObj.TenantID = tenantID")
	}
	p.P(`err = db.Where(&ormObj).Delete(&`, typeName, `ORM{}).Error`)
	p.P(`return err`)
	p.P(`}`)
	p.P()
}

func (p *ormPlugin) generateListHandler(message *generator.Descriptor) {
	typeNamePb, typeName, typeNameOrm := getTypeNames(message)

	p.P(`// DefaultList`, typeName, ` executes a gorm list call`)
	p.P(`func DefaultList`, typeName, `(ctx context.Context, db *`, p.gormPkgName,
		`.DB) ([]*`, typeName, `, error) {`)
	p.P(`ormResponse := []`, typeName, `ORM{}`)
	p.P(`db, err := `, p.lftPkgName, `.ApplyCollectionOperators(db, ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	if opts := p.GetMessageOptions(message); opts != nil && opts.GetMultiTenant() {
		p.P("tenantID, tIDErr := auth.GetTenantID(ctx)")
		p.P("if tIDErr != nil {")
		p.P("return nil, tIDErr")
		p.P("}")
		p.P(`db = db.Where(&`, typeNameOrm, `{TenantID: tenantID})`)
	}
	p.P(`if err := db.Set("gorm:auto_preload", true).Find(&ormResponse).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse := []*`, typeNamePb, `{}`)
	p.P(`for _, responseEntry := range ormResponse {`)
	p.P(`temp, err := Convert`, typeName, `FromORM(responseEntry)`)
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
func (p *ormPlugin) findAssociationKeys(parent *generator.Descriptor,
	child *generator.Descriptor, field *descriptor.FieldDescriptorProto) map[string]string {
	// Check for gorm tags
	_, parentTypeName, _ := getTypeNames(parent)
	keyMap := make(map[string]string)
	defaultKeyMap := map[string]string{fmt.Sprintf("%sID", parentTypeName): "ID"}
	childFields := []string{fmt.Sprintf("%sID", parentTypeName)}
	parentFields := []string{"ID"}
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
				childFields = append(childFields, lintName(generator.CamelCase(fName)))
			}
		} else if tagType[0] == "association_foreignkey" {
			parentFields = []string{}
			for _, fName := range strings.Split(tagType[1], ",") { // for compound key
				parentFields = append(parentFields, lintName(generator.CamelCase(fName)))
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

func (p *ormPlugin) removeChildAssociations(message *generator.Descriptor) bool {
	_, typeName, _ := getTypeNames(message)
	usedTenantID := false
	if _, exists := typeNames[typeName]; !exists {
		return usedTenantID
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
		p.P(`filterObj`, rawFieldType, ` := `, rawFieldType, `ORM{}`)
		for k, v := range keys {
			for _, childField := range typeNames[rawFieldType].GetField() {
				if strings.EqualFold(childField.GetName(), k) {
					childFKeyTypeName, _ = p.GoType(message, childField)
					break
				}
			}
			// If we accidentally delete without a set PK in our filter, everything might go
			if strings.Contains(childFKeyTypeName, "*") {
				p.P(`if ormObj.`, v, ` == nil || *ormObj.`, v, ` == `,
					guessZeroValue(childFKeyTypeName), ` {`)
			} else {
				p.P(`if ormObj.`, v, ` == `, guessZeroValue(childFKeyTypeName), ` {`)
			}
			p.P(`return nil, errors.New("Can't do overwriting update with no '`, v,
				`' value for FK of field '`, lintName(generator.CamelCase(field.GetName())), `'")`)
			p.P(`}`)

			p.P(`filterObj`, rawFieldType, `.`, k, ` = ormObj.`, v)
		}
		if opts := p.GetMessageOptions(typeNames[rawFieldType]); opts != nil && opts.GetMultiTenant() {
			p.P("tenantID, tIDErr := auth.GetTenantID(ctx)")
			p.P("if tIDErr != nil {")
			p.P("return nil, tIDErr")
			p.P("}")
			p.P(`filterObj`, rawFieldType, `.TenantID = tenantID`)
			usedTenantID = true
		}

		p.P(`if err = db.Where(filterObj`, rawFieldType, `).Delete(`,
			strings.Trim(fieldType, "[]*"), `{}).Error; err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
	}
	return usedTenantID
}

func (p *ormPlugin) generateStrictUpdateHandler(message *generator.Descriptor) {
	typeNamePb, typeName, typeNameOrm := getTypeNames(message)
	p.P(`// DefaultStrictUpdate`, typeName, ` clears first level 1:many children and then executes a gorm update call`)
	p.P(`func DefaultStrictUpdate`, typeName, `(ctx context.Context, in *`,
		typeNamePb, `, db *`, p.gormPkgName, `.DB) (`, `*`, typeNamePb, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, fmt.Errorf("Nil argument to DefaultCascadedUpdate`, typeName, `")`)
	p.P(`}`)
	p.P(`ormObj, err := Convert`, typeName, `ToORM(*in)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	usedTenantID := p.removeChildAssociations(message)
	if opts := p.GetMessageOptions(message); opts != nil && opts.GetMultiTenant() {
		if !usedTenantID {
			p.P("tenantID, tIDErr := auth.GetTenantID(ctx)")
			p.P("if tIDErr != nil {")
			p.P("return nil, tIDErr")
			p.P("}")
		}
		p.P(`db = db.Where(&`, typeNameOrm, `{TenantID: tenantID})`)
	}
	p.P(`if err = db.Save(&ormObj).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse, err := Convert`, typeName, `FromORM(ormObj)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`return &pbResponse, nil`)
	p.P(`}`)
	p.P()
}
