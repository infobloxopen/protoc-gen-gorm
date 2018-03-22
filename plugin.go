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
		p.generateMessages(file, msg)
		p.generateMapFunctions(msg)
	}

	p.P()
	p.P(`////////////////////////// CURDL for objects`)
	p.generateDefaultHandlers(file)
	p.P()
	if len(file.GetService()) > 0 {
		p.P(`////////////////////////// Handlers for RPCs`)
		p.generateDefaultServiceHandlers(file)
	}
}

func (p *ormPlugin) generateMessages(file *generator.FileDescriptor, message *generator.Descriptor) {
	ccTypeNamePb := generator.CamelCaseSlice(message.TypeName())
	ccTypeName := fmt.Sprintf("%sORM", lintName(ccTypeNamePb))

	// Check for a comment, generate a default if none is provided
	comment := p.Comments(message.Path())
	commentStart := strings.Split(strings.Trim(comment, " "), " ")[0]
	if generator.CamelCase(commentStart) == ccTypeNamePb || commentStart == ccTypeName {
		comment = strings.Replace(comment, commentStart, ccTypeName, 1)
	} else if len(comment) == 0 {
		comment = fmt.Sprintf(" %s no comment was provided for message type", ccTypeName)
	} else {
		comment = fmt.Sprintf(" %s ... %s", ccTypeName, comment)
	}
	p.P(`//`, comment)
	p.P(`type `, ccTypeName, ` struct {`)
	p.In()
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
	p.Out()
	p.P(`}`)

	// Set TableName back to gorm default to remove "ORM" suffix
	p.P(`func (`, ccTypeName, `) TableName() string {`)
	p.In()

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
	p.Out()
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
	p.In()
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
	p.Out()
	p.P(`}`)

	p.P()
	///// To Pb
	p.P(`// Convert`, ccTypeNameBase, `FromORM takes an orm object and returns a pb object`)
	p.P(`func Convert`, ccTypeNameBase, `FromORM (from `, ccTypeNameOrm, `) `,
		ccTypeNamePb, ` {`)
	p.In()
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
	p.Out()
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
			p.In()
			if toORM {
				p.P(`to.`, fieldName, ` = int32(from.`, fromName, `)`)
			} else {
				fieldType, _ := p.GoType(message, field)
				p.P(`to.`, fieldName, ` = `, fieldType, `(from.`, fromName, `)`)
			}
			p.Out()
			p.P(`}`) // end for
		} else if *(field.Type) == typeMessage { // WKT or custom type (hopefully)
			//Check for WKTs
			p.P(`for _, v := range from.`, fromName, ` {`)
			p.In()
			fieldType, _ := p.GoType(message, field)
			parts := strings.Split(fieldType, ".")
			coreType := parts[len(parts)-1]
			// Type is a WKT, convert to/from as ptr to base type
			if _, exists := wellKnownTypes[coreType]; exists {
				if toORM {
					p.P(`if `, fieldName, ` != nil {`)
					p.In()
					p.P(`temp := from.`, fromName, `.Value`)
					p.P(`to.`, fieldName, ` = append(to.`, fieldName, `, &temp`)
					p.Out()
					p.P(`} else {`)
					p.In()
					p.P(`to.`, fieldName, ` = append(nil)`)
					p.Out()
					p.P(`}`)
				} else {
					p.P(`if from.`, fromName, ` != nil {`)
					p.In()
					p.P(`to.`, fieldName, ` = append(t.`, fieldName, `, &`, p.wktPkgName, ".", coreType,
						`{Value: *from.`, fromName, `}`)
					p.Out()
					p.P(`} else {`)
					p.In()
					p.P(`to.`, fieldName, ` = append(to.`, fieldName, `, nil)`)
					p.Out()
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
					p.In()
					p.P(`temp`, lintName(fieldName), ` := Convert`, fieldType, dir, `ORM (*v)`)
					p.P(`to.`, fieldName, ` = append(to.`, fieldName, `, &temp`, lintName(fieldName), `)`)
					p.Out()
					p.P(`} else {`)
					p.In()
					p.P(`to.`, fieldName, ` = append(to.`, fieldName, `, nil)`)
					p.Out()
					p.P(`}`)
				} else {
					p.P(`to.`, fieldName, ` = Convert`, fieldType, dir, `ORM (from.`, fromName, `)`)
				}
				p.Out()
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
				p.In()
				p.P(`v := from.`, fromName, `.Value`)
				p.P(`to.`, fieldName, ` = &v`)
				p.Out()
				p.P(`}`)
			} else {
				p.P(`if from.`, fromName, ` != nil {`)
				p.In()
				p.P(`to.`, fieldName, ` = &`, p.wktPkgName, ".", coreType, `{Value: *from.`, fromName, `}`)
				p.Out()
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
					p.In()
					p.P(`temp`, lintName(fieldName), ` := Convert`, fieldType, dir, `ORM (*from.`, fromName, `)`)
					p.P(`to.`, fieldName, ` = &temp`, lintName(fieldName))
					p.Out()
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
		pkgGORM := p.NewImport("github.com/jinzhu/gorm")
		pkgGORM.Use()
		p.gormPkgAlias = pkgGORM.Name()
		pkgContext := p.NewImport("context")
		pkgContext.Use()
		p.generateCreateHandler(file, message)
		p.generateReadHandler(file, message)
		p.generateUpdateHandler(file, message)
		p.generateDeleteHandler(file, message)
		// p.generateListHandler(file, service, method)
	}
}

func (p *ormPlugin) generateCreateHandler(file *generator.FileDescriptor, message *generator.Descriptor) {
	typeNamePb := generator.CamelCaseSlice(message.TypeName())
	typeName := lintName(typeNamePb)
	p.P(`// DefaultCreate`, typeName, ` executes a basic gorm create call`)
	p.P(`func DefaultCreate`, typeName, `(ctx context.Context, in *`,
		typeNamePb, `, db *`, p.gormPkgAlias, `.DB) (*`, typeNamePb, `, error) {`)
	p.In()
	p.P(`if in == nil {`)
	p.In()
	p.P(`return nil, fmt.Errorf("Nil argument to DefaultCreate`, typeName, `")`)
	p.Out()
	p.P(`}`)
	p.P(`ormObj := Convert`, typeName, `ToORM(*in)`)
	p.P(`if err := db.Create(&ormObj).Error; err != nil {`)
	p.In()
	p.P(`return nil, err`)
	p.Out()
	p.P(`}`)
	p.P(`pbResponse := Convert`, typeName, `FromORM(ormObj)`)
	p.P(`return &pbResponse, nil`)
	p.Out()
	p.P(`}`)
	p.P()
}

func (p *ormPlugin) generateReadHandler(file *generator.FileDescriptor, message *generator.Descriptor) {
	typeNamePb := generator.CamelCaseSlice(message.TypeName())
	typeName := lintName(typeNamePb)
	p.P(`// DefaultRead`, typeName, ` executes a basic gorm read call`)
	p.P(`func DefaultRead`, typeName, `(ctx context.Context, in *`,
		typeNamePb, `, db *`, p.gormPkgAlias, `.DB) (*`, typeNamePb, `, error) {`)
	p.In()
	p.P(`if in == nil {`)
	p.In()
	p.P(`return nil, fmt.Errorf("Nil argument to DefaultRead`, typeName, `")`)
	p.Out()
	p.P(`}`)
	p.P(`ormParams := Convert`, typeName, `ToORM(*in)`)
	p.P(`ormResponse := `, typeName, `ORM{}`)
	p.P(`if err := db.Set("gorm:auto_preload", true).Where(&ormParams).First(&ormResponse).Error; err != nil {`)
	p.In()
	p.P(`return nil, err`)
	p.Out()
	p.P(`}`)
	p.P(`pbResponse := Convert`, typeName, `FromORM(ormResponse)`)
	p.P(`return &pbResponse, nil`)
	p.Out()
	p.P(`}`)
	p.P()
}

func (p *ormPlugin) generateUpdateHandler(file *generator.FileDescriptor, message *generator.Descriptor) {
	typeNamePb := generator.CamelCaseSlice(message.TypeName())
	typeName := lintName(typeNamePb)
	p.P(`// DefaultUpdate`, typeName, ` executes a basic gorm update call`)
	p.P(`func DefaultUpdate`, typeName, `(ctx context.Context, in *`,
		typeNamePb, `, db *`, p.gormPkgAlias, `.DB) (*`, typeNamePb, `, error) {`)
	p.In()
	p.P(`if in == nil {`)
	p.In()
	p.P(`return nil, fmt.Errorf("Nil argument to DefaultUpdate`, typeName, `")`)
	p.Out()
	p.P(`}`)
	p.P(`ormObj := Convert`, typeName, `ToORM(*in)`)
	p.P(`if err := db.Save(&ormObj).Error; err != nil {`)
	p.In()
	p.P(`return nil, err`)
	p.Out()
	p.P(`}`)
	p.P(`pbResponse := Convert`, typeName, `FromORM(ormObj)`)
	p.P(`return &pbResponse, nil`)
	p.Out()
	p.P(`}`)
	p.P()
}

func (p *ormPlugin) generateDeleteHandler(file *generator.FileDescriptor, message *generator.Descriptor) {
	typeNamePb := generator.CamelCaseSlice(message.TypeName())
	typeName := lintName(typeNamePb)
	p.P(`// DefaultDelete`, typeName, ` executes a basic gorm delete call`)
	p.P(`func DefaultDelete`, typeName, `(ctx context.Context, in *`,
		typeNamePb, `, db *`, p.gormPkgAlias, `.DB) error {`)
	p.In()
	p.P(`if in == nil {`)
	p.In()
	p.P(`return fmt.Errorf("Nil argument to DefaultDelete`, typeName, `")`)
	p.Out()
	p.P(`}`)
	p.P(`ormObj := Convert`, typeName, `ToORM(*in)`)
	p.P(`err := db.Where(&ormObj).Delete(&`, typeName, `ORM{}).Error`)
	p.P(`return err`)
	p.Out()
	p.P(`}`)
	p.P()
}

/*************************** Service Based Generation *************/
func (p *ormPlugin) generateDefaultServiceHandlers(file *generator.FileDescriptor) {
	for _, service := range file.GetService() {
		svcName := lintName(generator.CamelCase(service.GetName()))

		// All the default handler has is a db connection
		p.P(`type `, svcName, `DefaultHandler struct {`)
		p.In()
		p.RecordTypeUse(`github.com/jinzhu/gorm`)
		p.P(`DB `, p.gormPkgAlias, `.DB`)
		p.Out()
		p.P(`}`)
		for _, method := range service.GetMethod() {
			methodName := generator.CamelCase(method.GetName())
			if strings.HasPrefix(methodName, "Create") {
				p.generateCreateServiceHandler(file, service, method)
			} else if strings.HasPrefix(methodName, "Read") {
				p.generateReadServiceHandler(file, service, method)
			} else if strings.HasPrefix(methodName, "Update") {
				p.generateUpdateServiceHandler(file, service, method)
			} else if strings.HasPrefix(methodName, "Delete") {
				p.generateDeleteServiceHandler(file, service, method)
			} else if strings.HasPrefix(methodName, "List") {
				p.generateListServiceHandler(file, service, method)
			} else {
				p.P(`// You'll have to create the `, methodName, ` handler function yourself`)
				p.P()
			}
		}
	}
}

// TODO if type is from same package, need to not output package name or it
// causes "import loop" error
func (p *ormPlugin) generateDefaultRPCHandlerSignature(inType, outType generator.Object, methodName, svcName string) {

	p.P(`// `, methodName, ` ...`) // Golint comment, should check for proto file comments
	p.P(`func (m *`, svcName, `DefaultHandler) `, methodName, ` (ctx context.Context, in *`,
		inType.PackageName(), ".", generator.CamelCaseSlice(inType.TypeName()),
		`, opts ...grpc.CallOption) (*`, outType.PackageName(), ".",
		generator.CamelCaseSlice(outType.TypeName()), `, error) {`)
}

// Fetches and formats some names
func (p *ormPlugin) getNames(file *generator.FileDescriptor,
	service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) (generator.Object, generator.Object, string, string) {

	inType := p.ObjectNamed(method.GetInputType())
	p.RecordTypeUse(method.GetInputType())
	outType := p.ObjectNamed(method.GetOutputType())
	p.RecordTypeUse(method.GetOutputType())
	methodName := generator.CamelCase(method.GetName())
	svcName := lintName(generator.CamelCase(service.GetName()))
	return inType, outType, methodName, svcName
}

// Enforces convention of <Operation><Object Type>, e.g. CreateTag, UpdateContact
// outputting an object that is or contains the Object Type in the name
func validateORMableOutputType(operation string, method *descriptor.MethodDescriptorProto,
	outputType generator.Object) error {

	typeInMethodName := strings.TrimPrefix(generator.CamelCase(method.GetName()), operation)
	validOutputType := false
	outputTypeName := generator.CamelCaseSlice(outputType.TypeName())
	if outputTypeName == typeInMethodName || strings.TrimSuffix(strings.TrimPrefix(outputTypeName, operation), "Response") == typeInMethodName {
		validOutputType = true
	} else if typeMsg, exists := typeNames[outputTypeName]; exists {
		// Check subfields for an ormable object that matches name with the method
		for _, field := range typeMsg.Field {
			rawType := strings.Split(field.GetTypeName(), ".")
			if generator.CamelCase(rawType[len(rawType)-1]) == typeInMethodName {
				// fmt.Fprintf(os.Stderr, "Match found for type %s in %s\n", generator.CamelCase(rawType[len(rawType)-1]), typeMsg.GetName())
				validOutputType = true
				break
			}
		}
	}
	if !validOutputType {
		return fmt.Errorf(`Your choice of output type (%s) in %s breaks convention`,
			generator.CamelCaseSlice(outputType.TypeName()), generator.CamelCase(method.GetName()))
	}
	return nil
}

// Just checks if the method name is <Operation><Ormable Type>
func checkTypeInMethodName(operation, method string) bool {
	typeInMethodName := strings.TrimPrefix(method, operation)
	_, exists := convertibleTypes[typeInMethodName]
	return exists
}

func (p *ormPlugin) generateCreateServiceHandler(file *generator.FileDescriptor,
	service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {

	inType, outType, methodName, svcName := p.getNames(file, service, method)
	if !checkTypeInMethodName("Create", methodName) {
		p.P(`// Cannot autogen create function `, methodName, `: unrecognized anticipated type `)
		return
	}
	if err := validateORMableOutputType("Create", method, outType); err != nil {
		p.Fail(err.Error())
	}
	p.P()
	p.generateDefaultRPCHandlerSignature(inType, outType, methodName, svcName)
	p.In()
	p.P(`return nil, nil`)
	p.Out()
	p.P(`}`)
}

func (p *ormPlugin) generateReadServiceHandler(file *generator.FileDescriptor,
	service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {

	inType, outType, methodName, svcName := p.getNames(file, service, method)
	if !checkTypeInMethodName("Read", methodName) {
		p.P(`// Cannot autogen read function `, methodName, `: unrecognized anticipated type `)
		return
	}
	if err := validateORMableOutputType("Read", method, outType); err != nil {
		p.Fail(err.Error())
	}
	//typeInMethodName := strings.TrimPrefix(strings.TrimPrefix(methodName, "Read"), "read")
	p.generateDefaultRPCHandlerSignature(inType, outType, methodName, svcName)
	p.In()
	p.P(`return nil, nil`)
	p.Out()
	p.P(`}`)
}

func (p *ormPlugin) generateUpdateServiceHandler(file *generator.FileDescriptor,
	service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {

	inType, outType, methodName, svcName := p.getNames(file, service, method)
	if !checkTypeInMethodName("Update", methodName) {
		p.P(`// Cannot autogen update function `, methodName, `: unrecognized anticipated type `)
		return
	}
	if err := validateORMableOutputType("Update", method, outType); err != nil {
		p.Fail(err.Error())
	}
	p.generateDefaultRPCHandlerSignature(inType, outType, methodName, svcName)
	p.In()
	p.P(`return nil, nil`)
	p.Out()
	p.P(`}`)
}

func (p *ormPlugin) generateDeleteServiceHandler(file *generator.FileDescriptor,
	service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {

	inType, outType, methodName, svcName := p.getNames(file, service, method)
	if !checkTypeInMethodName("Delete", methodName) {
		p.P(`// Cannot autogen read function `, methodName, `: unrecognized anticipated type `)
		return
	}
	p.generateDefaultRPCHandlerSignature(inType, outType, methodName, svcName)
	p.In()
	p.P(`return nil, nil`)
	p.Out()
	p.P(`}`)
}

func (p *ormPlugin) generateListServiceHandler(file *generator.FileDescriptor,
	service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {

	inType, outType, methodName, svcName := p.getNames(file, service, method)
	typeInMethodName := strings.TrimPrefix(methodName, "List")
	if _, exists := convertibleTypes[typeInMethodName]; !exists {
		p.P(`// Cannot autogen list function for unrecognized anticipated type `, typeInMethodName)
		return
	}
	p.generateDefaultRPCHandlerSignature(inType, outType, methodName, svcName)
	p.In()
	p.P(`return nil, nil`)
	p.Out()
	p.P(`}`)
}
