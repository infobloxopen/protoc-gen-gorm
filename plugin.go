package main

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
var typeNames = make(map[string]generator.Descriptor)

const typeMessage = 11
const typeEnum = 14

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
	generator.PluginImports
	wktPkgName   string
	gormPkgAlias string
	lftPkgName   string // 'Locally Famous Types', used for collection operators
}

func (p *ormPlugin) GenerateImports(file *generator.FileDescriptor) {
	if p.gormPkgAlias != "" {
		p.PrintImport("context", "context")
		p.PrintImport(p.gormPkgAlias, "github.com/jinzhu/gorm")
		p.PrintImport("grpc", "google.golang.org/grpc")
		p.PrintImport(p.lftPkgName, "github.com/Infoblox-CTO/ngp.api.toolkit/op/gorm")
	}
}

func (p *ormPlugin) Name() string {
	return "gorm"
}

func (p *ormPlugin) Init(g *generator.Generator) {
	p.Generator = g
}

func (p *ormPlugin) Generate(file *generator.FileDescriptor) {

	p.PluginImports = generator.NewPluginImports(p.Generator)

	// Preload just the types we'll be creating
	for _, msg := range file.Messages() {
		// We don't want to bother with the MapEntry stuff
		if msg.DescriptorProto.GetOptions().GetMapEntry() {
			continue
		}
		unlintedTypeName := generator.CamelCaseSlice(msg.TypeName())
		typeNames[unlintedTypeName] = *msg
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
		p.generateMapFunctions(msg)
	}

	p.P()
	p.P(`////////////////////////// CURDL for objects`)
	p.generateDefaultHandlers(file)
	p.P()

	p.generateDefaultServer(file)
}

func (p *ormPlugin) generateMessages(message *generator.Descriptor) {
	ccTypeNamePb := generator.CamelCaseSlice(message.TypeName())
	ccTypeName := fmt.Sprintf("%sORM", lintName(ccTypeNamePb))

	// Check for a comment, generate a default if none is provided
	comment := p.Comments(message.Path())
	commentStart := strings.Split(strings.Trim(comment, " "), " ")[0]
	if generator.CamelCase(commentStart) == ccTypeNamePb || commentStart == ccTypeName {
		comment = strings.Replace(comment, commentStart, ccTypeName, 1)
	} else if len(comment) == 0 {
		comment = fmt.Sprintf(" %s no comment was provided for message type", ccTypeName)
	} else if len(strings.Replace(comment, " ", "", -1)) > 0 {
		comment = fmt.Sprintf(" %s %s", ccTypeName, comment)
	} else {
		comment = fmt.Sprintf(" %s no comment provided", ccTypeName)
	}
	p.P(`//`, comment)
	p.P(`type `, ccTypeName, ` struct {`)
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

	for _, field := range message.Field {
		fieldName := p.GetOneOfFieldName(message, field)
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
		if *(field.Type) == typeEnum {
			fieldType = "int32"
		} else if *(field.Type) == typeMessage {
			//Check for WKTs or fields of nonormable types
			parts := strings.Split(fieldType, ".")
			if v, exists := wellKnownTypes[parts[len(parts)-1]]; exists {
				p.RecordTypeUse(".google.protobuf.StringValue")
				p.wktPkgName = strings.Trim(parts[0], "*")
				fieldType = v
			} else if parts[len(parts)-1] == "Empty" {
				p.RecordTypeUse(".google.protobuf.Empty")
				p.P("// Empty type has no ORM equivalency")
				continue
			} else if _, exists := convertibleTypes[strings.Trim(fieldType, "[]*")]; !exists {
				p.P("// Skipping type ", fieldType, ", not tagged as ormable")
				continue
			} else {
				fieldType = fmt.Sprintf("%sORM", lintName(fieldType))
				// Insert the foreign key if not present,
				if tagString == "" {
					tagString = fmt.Sprintf("`gorm:\"foreignkey:%sID\"`", lintName(ccTypeNamePb))
				} else if !strings.Contains(strings.ToLower(tagString), "foreignkey") {
					if strings.Contains(strings.ToLower(tagString), "gorm:") { // gorm tag already there
						index := strings.Index(tagString, "gorm:")
						tagString = fmt.Sprintf("%sforeignkey:%sID;%s", tagString[:index+6],
							lintName(ccTypeNamePb), tagString[index+6:])
					} else { // no gorm tag yet
						tagString = fmt.Sprintf("`gorm:\"foreignkey:%sID\" %s", lintName(ccTypeNamePb), tagString[1:])
					}
				}
			}
		} else if field.IsRepeated() {
			p.P(`// A repeated raw type is not supported by gORM`)
			continue
		}
		p.P(lintName(fieldName), " ", fieldType, tagString)
	}
	p.P(`}`)

	// Set TableName back to gorm default to remove "ORM" suffix
	p.P(`// TableName overrides the default tablename generated by GORM`)
	p.P(`func (`, ccTypeName, `) TableName() string {`)

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
func (p *ormPlugin) generateMapFunctions(message *generator.Descriptor) {
	ccTypeNamePb := generator.CamelCaseSlice(message.TypeName())
	ccTypeNameBase := lintName(ccTypeNamePb)
	ccTypeNameOrm := fmt.Sprintf("%sORM", ccTypeNameBase)
	///// To Orm
	p.P(`// Convert`, ccTypeNameBase, `ToORM takes a pb object and returns an orm object`)
	p.P(`func Convert`, ccTypeNameBase, `ToORM (from `,
		ccTypeNamePb, `) `, ccTypeNameOrm, ` {`)
	p.P(`to := `, ccTypeNameOrm, `{}`)
	for _, field := range message.Field {
		// Checking if field is skipped
		if field.Options != nil {
			v, err := proto.GetExtension(field.Options, gorm.E_Field)
			if err == nil && v.(*gorm.GormFieldOptions) != nil {
				if v.(*gorm.GormFieldOptions).Drop != nil && *v.(*gorm.GormFieldOptions).Drop {
					p.P(`// Skipping field: `, p.GetOneOfFieldName(message, field))
					continue
				}
			}
		}
		p.generateFieldMap(message, field, true)
	}
	p.P(`return to`)
	p.P(`}`)

	p.P()
	///// To Pb
	p.P(`// Convert`, ccTypeNameBase, `FromORM takes an orm object and returns a pb object`)
	p.P(`func Convert`, ccTypeNameBase, `FromORM (from `, ccTypeNameOrm, `) `,
		ccTypeNamePb, ` {`)
	p.P(`to := `, ccTypeNamePb, `{}`)
	for _, field := range message.Field {
		// Checking if field is skipped
		if field.Options != nil {
			v, err := proto.GetExtension(field.Options, gorm.E_Field)
			if err == nil && v.(*gorm.GormFieldOptions) != nil {
				if v.(*gorm.GormFieldOptions).Drop != nil && *v.(*gorm.GormFieldOptions).Drop {
					p.P(`// Skipping field: `, p.GetOneOfFieldName(message, field))
					continue
				}
			}
		}
		p.generateFieldMap(message, field, false)
	}
	p.P(`return to`)
	p.P(`}`)
}

func (p *ormPlugin) generateFieldMap(message *generator.Descriptor, field *descriptor.FieldDescriptorProto, toORM bool) error {
	fieldName := p.GetOneOfFieldName(message, field)
	fromName := fieldName
	if toORM {
		fieldName = lintName(fromName)
	} else {
		fromName = lintName(fromName)
	}
	if field.IsRepeated() { // Repeated Object ----------------------------------
		if *(field.Type) == typeEnum {
			p.P(`for _, v := range from.`, fromName, ` {`)
			if toORM {
				p.P(`to.`, fieldName, ` = int32(from.`, fromName, `)`)
			} else {
				fieldType, _ := p.GoType(message, field)
				p.P(`to.`, fieldName, ` = `, fieldType, `(from.`, fromName, `)`)
			}
			p.P(`}`) // end for
		} else if *(field.Type) == typeMessage { // WKT or custom type (hopefully)
			//Check for WKTs
			p.P(`for _, v := range from.`, fromName, ` {`)
			fieldType, _ := p.GoType(message, field)
			parts := strings.Split(fieldType, ".")
			coreType := parts[len(parts)-1]
			// Type is a WKT, convert to/from as ptr to base type
			if _, exists := wellKnownTypes[coreType]; exists {
				if toORM {
					p.P(`if `, fieldName, ` != nil {`)
					p.P(`temp := from.`, fromName, `.Value`)
					p.P(`to.`, fieldName, ` = append(to.`, fieldName, `, &temp`)
					p.P(`} else {`)
					p.P(`to.`, fieldName, ` = append(nil)`)
					p.P(`}`)
				} else {
					p.P(`if from.`, fromName, ` != nil {`)
					p.P(`to.`, fieldName, ` = append(t.`, fieldName, `, &`, p.wktPkgName, ".", coreType,
						`{Value: *from.`, fromName, `}`)
					p.P(`} else {`)
					p.P(`to.`, fieldName, ` = append(to.`, fieldName, `, nil)`)
					p.P(`}`)
				}
			} else if _, exists := convertibleTypes[strings.Trim(fieldType, "[]*")]; exists {
				isPtr := strings.Contains(fieldType, "*") // Really, it should always be a ptr

				fieldType = strings.Trim(fieldType, "[]*")
				fieldType = lintName(fieldType)
				dir := "From"
				if toORM {
					dir = "To"
				}
				if isPtr {
					p.P(`if from.`, fromName, ` != nil {`)
					p.P(`temp`, lintName(fieldName), ` := Convert`, fieldType, dir, `ORM (*v)`)
					p.P(`to.`, fieldName, ` = append(to.`, fieldName, `, &temp`, lintName(fieldName), `)`)
					p.P(`} else {`)
					p.P(`to.`, fieldName, ` = append(to.`, fieldName, `, nil)`)
					p.P(`}`)
				} else {
					p.P(`to.`, fieldName, ` = Convert`, fieldType, dir, `ORM (from.`, fromName, `)`)
				}
				p.P(`}`) // end for
			} else {
				p.P(`// Type `, fieldType, ` is not an ORMable message type`)
			}
		} else { // Raw type, actually ORM won't support slice of raw type, no relational data
			//p.P(`to.`, fieldName, ` = append(to.`, fieldName, `, v)`)
		}
	} else if *(field.Type) == typeEnum { // Enum, which is an int32 ------------
		if toORM {
			p.P(`to.`, fieldName, ` = int32(from.`, fromName, `)`)
		} else {
			fieldType, _ := p.GoType(message, field)
			p.P(`to.`, fieldName, ` = `, fieldType, `(from.`, fromName, `)`)
		}
	} else if *(field.Type) == typeMessage { // Singular Object -----------------
		//Check for WKTs
		fieldType, _ := p.GoType(message, field)
		parts := strings.Split(fieldType, ".")
		coreType := parts[len(parts)-1]
		// Type is a WKT, convert to/from as ptr to base type
		if _, exists := wellKnownTypes[coreType]; exists {
			if toORM {
				p.P(`if from.`, fromName, ` != nil {`)
				p.P(`v := from.`, fromName, `.Value`)
				p.P(`to.`, fieldName, ` = &v`)
				p.P(`}`)
			} else {
				p.P(`if from.`, fromName, ` != nil {`)
				p.P(`to.`, fieldName, ` = &`, p.wktPkgName, ".", coreType, `{Value: *from.`, fromName, `}`)
				p.P(`}`)
			}
		} else { // Not a WKT, but a type we're building converters for
			if _, exists := convertibleTypes[strings.Trim(fieldType, "[]*")]; exists {
				isPtr := strings.Contains(fieldType, "*")

				fieldType = strings.Trim(fieldType, "*")
				fieldType = lintName(fieldType)
				dir := "From"
				if toORM {
					dir = "To"
				}
				if isPtr {
					p.P(`if from.`, fromName, ` != nil {`)
					p.P(`temp`, lintName(fieldName), ` := Convert`, fieldType, dir, `ORM (*from.`, fromName, `)`)
					p.P(`to.`, fieldName, ` = &temp`, lintName(fieldName))
					p.P(`}`)
				} else {
					p.P(`to.`, fieldName, ` = Convert`, fieldType, dir, `ORM (from.`, fromName, `)`)
				}
			}
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
		p.gormPkgAlias = "gorm"
		p.lftPkgName = "ops"

		p.generateCreateHandler(message)
		p.generateReadHandler(message)
		p.generateUpdateHandler(message)
		p.generateDeleteHandler(message)
		p.generateListHandler(message)
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
	typeNamePb := generator.CamelCaseSlice(message.TypeName())
	typeName := lintName(typeNamePb)
	p.P(`// DefaultCreate`, typeName, ` executes a basic gorm create call`)
	p.P(`func DefaultCreate`, typeName, `(ctx context.Context, in *`,
		typeNamePb, `, db *`, p.gormPkgAlias, `.DB) (*`, typeNamePb, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, fmt.Errorf("Nil argument to DefaultCreate`, typeName, `")`)
	p.P(`}`)
	p.P(`ormObj := Convert`, typeName, `ToORM(*in)`)
	if opts := p.GetMessageOptions(message); opts != nil && opts.GetMultiTenant() {
		p.P("tenantID, tIDErr := auth.GetTenantID(ctx)")
		p.P("if tIDErr != nil {")
		p.P("return nil, tIDErr")
		p.P("}")
		p.P("ormObj.TenantID = tenantID")
	}
	p.P(`if err := db.Create(&ormObj).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse := Convert`, typeName, `FromORM(ormObj)`)
	p.P(`return &pbResponse, nil`)
	p.P(`}`)
	p.P()
}

func (p *ormPlugin) generateReadHandler(message *generator.Descriptor) {
	typeNamePb := generator.CamelCaseSlice(message.TypeName())
	typeName := lintName(typeNamePb)
	p.P(`// DefaultRead`, typeName, ` executes a basic gorm read call`)
	p.P(`func DefaultRead`, typeName, `(ctx context.Context, in *`,
		typeNamePb, `, db *`, p.gormPkgAlias, `.DB) (*`, typeNamePb, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, fmt.Errorf("Nil argument to DefaultRead`, typeName, `")`)
	p.P(`}`)
	p.P(`ormParams := Convert`, typeName, `ToORM(*in)`)
	if opts := p.GetMessageOptions(message); opts != nil && opts.GetMultiTenant() {
		p.P("tenantID, tIDErr := auth.GetTenantID(ctx)")
		p.P("if tIDErr != nil {")
		p.P("return nil, tIDErr")
		p.P("}")
		p.P("ormParams.TenantID = tenantID")
	}
	p.P(`ormResponse := `, typeName, `ORM{}`)
	p.P(`if err := db.Set("gorm:auto_preload", true).Where(&ormParams).First(&ormResponse).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse := Convert`, typeName, `FromORM(ormResponse)`)
	p.P(`return &pbResponse, nil`)
	p.P(`}`)
	p.P()
}

func (p *ormPlugin) generateUpdateHandler(message *generator.Descriptor) {
	typeNamePb := generator.CamelCaseSlice(message.TypeName())
	typeName := lintName(typeNamePb)

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
		typeNamePb, `, db *`, p.gormPkgAlias, `.DB) (*`, typeNamePb, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, fmt.Errorf("Nil argument to DefaultUpdate`, typeName, `")`)
	p.P(`}`)
	if isMultiTenant {
		p.P(fmt.Sprintf("if exists, err := DefaultRead%s(ctx, &%s{Id: in.GetId()}, db); err != nil {", typeName, typeName))
		p.P("return nil, err")
		p.P("} else if exists == nil {")
		p.P(fmt.Sprintf("return nil, errors.New(\"%s not found\")", typeName))
		p.P("}")
	}

	p.P(`ormObj := Convert`, typeName, `ToORM(*in)`)
	p.P(`if err := db.Save(&ormObj).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse := Convert`, typeName, `FromORM(ormObj)`)
	p.P(`return &pbResponse, nil`)
	p.P(`}`)
	p.P()
}

func (p *ormPlugin) generateDeleteHandler(message *generator.Descriptor) {
	typeNamePb := generator.CamelCaseSlice(message.TypeName())
	typeName := lintName(typeNamePb)
	p.P(`// DefaultDelete`, typeName, ` executes a basic gorm delete call`)
	p.P(`func DefaultDelete`, typeName, `(ctx context.Context, in *`,
		typeNamePb, `, db *`, p.gormPkgAlias, `.DB) error {`)
	p.P(`if in == nil {`)
	p.P(`return fmt.Errorf("Nil argument to DefaultDelete`, typeName, `")`)
	p.P(`}`)
	p.P(`ormObj := Convert`, typeName, `ToORM(*in)`)
	if opts := p.GetMessageOptions(message); opts != nil && opts.GetMultiTenant() {
		p.P("tenantID, tIDErr := auth.GetTenantID(ctx)")
		p.P("if tIDErr != nil {")
		p.P("return tIDErr")
		p.P("}")
		p.P("ormObj.TenantID = tenantID")
	}
	p.P(`err := db.Where(&ormObj).Delete(&`, typeName, `ORM{}).Error`)
	p.P(`return err`)
	p.P(`}`)
	p.P()
}

func (p *ormPlugin) generateListHandler(message *generator.Descriptor) {
	typeNamePb := generator.CamelCaseSlice(message.TypeName())
	typeName := lintName(typeNamePb)

	p.P(`// DefaultList`, typeName, ` executes a basic gorm find call`)
	p.P(`func DefaultList`, typeName, `(ctx context.Context, db *`, p.gormPkgAlias, `.DB) ([]*`, typeName, `, error) {`)
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
		p.P("db = db.Where(&ContactORM{TenantID: tenantID})")
	}
	p.P(`if err := db.Set("gorm:auto_preload", true).Find(&ormResponse).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse := []*`, typeName, `{}`)
	p.P(`for _, responseEntry := range ormResponse {`)
	p.P(`temp := Convert`, typeName, `FromORM(responseEntry)`)
	p.P(`pbResponse = append(pbResponse, &temp)`)
	p.P(`}`)
	p.P(`return pbResponse, nil`)
	p.P(`}`)
	p.P()
}

func (p *ormPlugin) generateDefaultServer(file *generator.FileDescriptor) {
	for _, service := range file.GetService() {
		svcName := lintName(generator.CamelCase(service.GetName()))
		if service.Options != nil {
			v, err := proto.GetExtension(service.GetOptions(), gorm.E_Server)
			opts := v.(*gorm.AutoServerOptions)
			if err == nil && opts != nil && *opts.Autogen {
				// All the default server has is a db connection
				p.P(`type `, svcName, `DefaultServer struct {`)
				p.P(`DB `, p.gormPkgAlias, `.DB`)
				p.P(`}`)
				for _, method := range service.GetMethod() {
					methodName := lintName(generator.CamelCase(method.GetName()))
					if strings.HasPrefix(methodName, "Create") {
						p.generateCreateServerMethod(file, service, method)
					} else if strings.HasPrefix(methodName, "Read") {
						p.generateReadServerMethod(file, service, method)
					} else if strings.HasPrefix(methodName, "Update") {
						p.generateUpdateServerMethod(file, service, method)
					} else if strings.HasPrefix(methodName, "Delete") {
						p.generateDeleteServerMethod(file, service, method)
					} else if strings.HasPrefix(methodName, "List") {
						p.generateListServerMethod(file, service, method)
					}
				}
			}
		}
	}
}

func (p *ormPlugin) generateCreateServerMethod(file *generator.FileDescriptor,
	service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	p.P(`return DefaultCreate`, p.TypeName(inType), `(ctx, in, db)`)
	p.P(`}`)
}

func (p *ormPlugin) generateReadServerMethod(file *generator.FileDescriptor,
	service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	p.P(`return DefaultRead`, p.TypeName(inType), `(ctx, in, db)`)
	p.P(`}`)
}

func (p *ormPlugin) generateUpdateServerMethod(file *generator.FileDescriptor,
	service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	p.P(`return DefaultUpdate`, p.TypeName(inType), `(ctx, in, db)`)
	p.P(`}`)
}

func (p *ormPlugin) generateDeleteServerMethod(file *generator.FileDescriptor,
	service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	p.P(`return nil, DefaultDelete`, p.TypeName(inType), `(ctx, in, db)`)
	p.P(`}`)
}

func (p *ormPlugin) generateListServerMethod(file *generator.FileDescriptor,
	service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	p.P(`var l `, p.TypeName(outType))
	p.P(`res, err := DefaultList`, strings.TrimPrefix(methodName, "List"), `(ctx, db)`)
	p.P(`l.Results = res`)
	p.P(`return &l, err`)
	p.P(`}`)
}

func (p *ormPlugin) generateMethodSignature(inType, outType generator.Object, methodName, svcName string) {
	p.P(`// `, methodName, ` ...`)
	p.P(`func (m *`, svcName, `DefaultServer) `, methodName, ` (ctx context.Context, in *`,
		p.TypeName(inType), `, opts ...grpc.CallOption) (*`, p.TypeName(outType), `, error) {`)
}

func (p *ormPlugin) getMethodProps(service *descriptor.ServiceDescriptorProto,
	method *descriptor.MethodDescriptorProto) (generator.Object, generator.Object, string, string) {
	inType := p.ObjectNamed(method.GetInputType())
	p.RecordTypeUse(method.GetInputType())
	outType := p.ObjectNamed(method.GetOutputType())
	p.RecordTypeUse(method.GetOutputType())
	methodName := lintName(generator.CamelCase(method.GetName()))
	svcName := lintName(generator.CamelCase(service.GetName()))
	return inType, outType, methodName, svcName
}
