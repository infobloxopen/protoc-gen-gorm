package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/infobloxopen/protoc-gen-gorm/orm"
)

// ORMable types
var convertibleTypes = make(map[string]struct{})

// All message objects
var typeNames = make(map[string]generator.Descriptor)

type ormPlugin struct {
	*generator.Generator
	generator.PluginImports

	wktPkgName      string
	emptyPkgName    string
	pkgParent       generator.Single
	originalPackage string
	newPackage      string
}

func (p *ormPlugin) Name() string {
	return "gorm"
}

func (p *ormPlugin) Init(g *generator.Generator) {
	p.Generator = g
}

/*func (p *ormPlugin) GenerateImports(file *generator.FileDescriptor) {
	if p.newPackage != "" {
		p.PrintImport(p.originalPackage, p.originalPackage)
	}
}*/

func (p *ormPlugin) Generate(file *generator.FileDescriptor) {

	p.PluginImports = generator.NewPluginImports(p.Generator)
	//p.pkgWKT = p.NewImport("github.com/golang/protobuf/ptypes/wrappers")
	p.originalPackage = file.PackageName()
	p.pkgParent = p.NewImport(fmt.Sprintf("%s/%s", path.Dir(*file.Name), p.originalPackage))
	p.pkgParent.Use()
	if file.Options != nil {
		v, err := proto.GetExtension(file.Options, orm.E_Package)
		if err == nil && v != nil {
			p.newPackage = *(v.(*string))
		} else {
			p.P("//has file options, but no package extension")
			p.P()
		}
	}
	// Preload just the types we'll be creating
	for _, msg := range file.Messages() {
		// We don't want to bother with the MapEntry stuff
		if msg.DescriptorProto.GetOptions().GetMapEntry() {
			continue
		}
		unlintedTypeName := generator.CamelCaseSlice(msg.TypeName())
		typeNames[unlintedTypeName] = *msg
		if msg.Options != nil {
			v, err := proto.GetExtension(msg.Options, orm.E_Opts)
			opts := v.(*orm.OrmMessageOptions)
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
		p.generateMessages(file, msg)
		p.generateMapFunctions(msg)
	}

	p.P()

	p.generateDefaultHandlers(file)
}

const typeMessage = 11
const typeEnum = 14

var wellKnownTypes = map[string]string{
	"StringValue": "*string",
	"DoubleValue": "*double",
	"FloatValue":  "*float",
	"Int32Value":  "*int32",
	"Int64Value":  "*int64",
	"Uint32Value": "*uint32",
	"UInt64Value": "*uint64",
	"BoolValue":   "*bool",
	//  "BytesValue" : "*[]byte",
}

func getTagString(tags []*orm.Tag) string {
	var tagString bytes.Buffer
	if len(tags) != 0 {
		tagString.WriteString("`")
		for _, tag := range tags {
			tagString.WriteString(fmt.Sprintf("%s:\"", *tag.Pkg))
			for i, tagVal := range tag.Values {
				tagString.WriteString(tagVal)
				if i < len(tag.Values)-1 {
					tagString.WriteString(",")
				}
			}
			tagString.WriteString("\"")
		}
		tagString.WriteString("`")
	}
	return tagString.String()
}

func (p *ormPlugin) generateMessages(file *generator.FileDescriptor, message *generator.Descriptor) {
	ccTypeNamePb := generator.CamelCaseSlice(message.TypeName())
	ccTypeName := lintName(ccTypeNamePb)
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
	if message.Options != nil {
		v, err := proto.GetExtension(message.Options, orm.E_Opts)
		opts := v.(*orm.OrmMessageOptions)
		if err == nil && opts != nil {
			for _, field := range opts.Include {
				tagString := getTagString(field.Tags)
				p.P(lintName(generator.CamelCase(*field.Name)), ` `, field.Type, ` `, tagString)
			}
		}
	}

	for _, field := range message.Field {
		fieldName := p.GetOneOfFieldName(message, field)
		fieldType, _ := p.GoType(message, field)
		var tagString bytes.Buffer
		if field.Options != nil {
			v, err := proto.GetExtension(field.Options, orm.E_Field)
			if err == nil && v.(*orm.OrmOptions) != nil {
				if v.(*orm.OrmOptions).Drop != nil && *v.(*orm.OrmOptions).Drop {
					p.P(`// Skipping field: `, fieldName)
					continue
				}
				tags := v.(*orm.OrmOptions).Tags
				if len(tags) != 0 {
					tagString.WriteString("`")
					for _, tag := range tags {
						tagString.WriteString(fmt.Sprintf("%s:\"", *tag.Pkg))
						for i, tagVal := range tag.Values {
							tagString.WriteString(tagVal)
							if i < len(tag.Values)-1 {
								tagString.WriteString(",")
							}
						}
						tagString.WriteString("\"")
					}
					tagString.WriteString("`")
				}
			} else if err != nil {
				p.P("//", err.Error())
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
			} else if _, exists := convertibleTypes[strings.Trim(fieldType, "[]*")]; !exists {
				p.P("// Can't work with type ", fieldType, ", not tagged as ormable")
				continue
			} else if parts[len(parts)-1] == "Empty" {
				p.RecordTypeUse(".google.protobuf.Empty")
			} else {
				fieldType = lintName(fieldType)
			}
		}
		p.P(lintName(fieldName), " ", fieldType, tagString.String())
	}
	p.Out()
	p.P(`}`)
}

func (p *ormPlugin) generateMapFunctions(message *generator.Descriptor) {
	ccTypeNamePb := generator.CamelCaseSlice(message.TypeName())
	ccTypeNameOrm := lintName(ccTypeNamePb)
	///// To Orm
	p.P(`// ConvertTo`, ccTypeNameOrm, ` takes a pb object and returns an orm object`)
	p.P(`func ConvertTo`, ccTypeNameOrm, `(from `, p.pkgParent.Name(), ".",
		ccTypeNamePb, `) `, ccTypeNameOrm, ` {`)
	p.In()
	p.P(`to := `, ccTypeNameOrm, `{}`)
	for _, field := range message.Field {
		if field.Options != nil {
			v, err := proto.GetExtension(field.Options, orm.E_Field)
			if err == nil && v.(*orm.OrmOptions) != nil {
				if v.(*orm.OrmOptions).Drop != nil && *v.(*orm.OrmOptions).Drop {
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
	p.P(`// ConvertFrom`, ccTypeNameOrm, ` takes an orm object and returns a pb object`)
	p.P(`func ConvertFrom`, ccTypeNameOrm, `(from `, ccTypeNameOrm, `) `,
		p.pkgParent.Name(), ".", ccTypeNamePb, ` {`)
	p.In()
	p.P(`to := `, p.pkgParent.Name(), ".", ccTypeNamePb, `{}`)
	for _, field := range message.Field {
		if field.Options != nil {
			v, err := proto.GetExtension(field.Options, orm.E_Field)
			if err == nil && v.(*orm.OrmOptions) != nil {
				if v.(*orm.OrmOptions).Drop != nil && *v.(*orm.OrmOptions).Drop {
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
		p.P(`for _, v := range from.`, fromName, ` {`)
		p.In()
		if *(field.Type) == typeEnum {
			if toORM {
				p.P(`to.`, fieldName, ` = int32(from.`, fromName, `)`)
			} else {
				fieldType, _ := p.GoType(message, field)
				p.P(`to.`, fieldName, ` = `, p.pkgParent.Name(), ".", fieldType, `(from.`, fromName, `)`)
			}
		} else if *(field.Type) == typeMessage { // WKT or custom type (hopefully)
			//Check for WKTs
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
					p.P(`temp`, lintName(fieldName), ` := Convert`, dir, fieldType, `(*v)`)
					p.P(`to.`, fieldName, ` = append(to.`, fieldName, `, &temp`, lintName(fieldName), `)`)
					p.Out()
					p.P(`} else {`)
					p.In()
					p.P(`to.`, fieldName, ` = append(to.`, fieldName, `, nil)`)
					p.Out()
					p.P(`}`)
				} else {
					p.P(`to.`, fieldName, ` = Convert`, dir, fieldType, `(from.`, fromName, `)`)
				}
			} else {
				p.P(`Type `, fieldType, ` is not an ORMable message type`)
			}
		} else { // Raw type (actually can't be pointer type)
			p.P(`to.`, fieldName, ` = append(to.`, fieldName, `, v)`)
		}
		p.Out()
		p.P(`}`)
	} else if *(field.Type) == typeEnum { // Enum, which is an int32 ------------
		if toORM {
			p.P(`to.`, fieldName, ` = int32(from.`, fromName, `)`)
		} else {
			fieldType, _ := p.GoType(message, field)
			p.P(`to.`, fieldName, ` = `, p.pkgParent.Name(), ".", fieldType, `(from.`, fromName, `)`)
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
					p.P(`temp`, lintName(fieldName), ` := Convert`, dir, fieldType, `(*from.`, fromName, `)`)
					p.P(`to.`, fieldName, ` = &temp`, lintName(fieldName))
					p.Out()
					p.P(`}`)
				} else {
					p.P(`to.`, fieldName, ` = Convert`, dir, fieldType, `(from.`, fromName, `)`)
				}
			}
		}
	} else { // Singular raw ----------------------------------------------------
		p.P(`to.`, fieldName, ` = from.`, fromName)
	}
	return nil
}

func (p *ormPlugin) generateDefaultHandlers(file *generator.FileDescriptor) {
	for _, service := range file.GetService() {
		svcName := lintName(generator.CamelCase(service.GetName()))

		p.P(`type `, svcName, `DefaultHandler struct {`)
		p.In()
		p.P(`DB gorm.DB`)
		p.RecordTypeUse(`github.com/jinzhu/gorm`)
		p.Out()
		p.P(`}`)
		for _, method := range service.GetMethod() {
			methodName := generator.CamelCase(method.GetName())
			if strings.HasPrefix(methodName, "Create") {
				p.generateCreateHandler(file, service, method)
			} else if strings.HasPrefix(methodName, "Read") {
				p.generateReadHandler(file, service, method)
			} else if strings.HasPrefix(methodName, "Update") {
				p.generateUpdateHandler(file, service, method)
			} else if strings.HasPrefix(methodName, "Delete") {
				p.generateDeleteHandler(file, service, method)
			} else if strings.HasPrefix(methodName, "List") {
				p.generateListHandler(file, service, method)
			} else {
				p.P(`// You'll have to create the `, methodName, ` handler function yourself`)
				p.P()
			}
		}
	}
}

func (p *ormPlugin) generateDefaultFunctionSignature(inType, outType generator.Object, methodName string) {
	//p.RecordTypeUse("golang.org/x/net/context")
	p.P(`// `, methodName, ` ...`)
	p.P(`func `, methodName, `Handler(ctx context.Context, in *`,
		inType.PackageName(), ".", generator.CamelCaseSlice(inType.TypeName()),
		`, db gorm.DB) (`, `*`, outType.PackageName(), ".",
		generator.CamelCaseSlice(outType.TypeName()), `, error) {`)

}

func (p *ormPlugin) generateDefaultHandler(inType, outType generator.Object, methodName, svcName string) {

	p.P(`// `, methodName, ` ...`)
	p.P(`func (m *`, svcName, `DefaultHandler) `, methodName, ` (ctx context.Context, in *`,
		inType.PackageName(), ".", generator.CamelCaseSlice(inType.TypeName()),
		`, opts ...grpc.CallOption) (*`, outType.PackageName(), ".",
		generator.CamelCaseSlice(outType.TypeName()), `, error) {`)
	p.In()
	p.P(`return `, methodName, `Handler(ctx, in, m.DB)`)
	p.Out()
	p.P(`}`)
}

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
				//_, typeExists := convertibleTypes[generator.CamelCase(rawType[len(rawType)-1])]; typeExists {
				fmt.Fprintf(os.Stderr, "Match found for type %s in %s\n", generator.CamelCase(rawType[len(rawType)-1]), typeMsg.GetName())
				validOutputType = true
				break
			}
		}
	}
	if !validOutputType {
		return fmt.Errorf(`I find your choice of output type (%s) in %s disturbing`,
			generator.CamelCaseSlice(outputType.TypeName()), generator.CamelCase(method.GetName()))
	}
	return nil
}

func checkTypeInMethodName(operation, method string) bool {
	typeInMethodName := strings.TrimPrefix(method, operation)
	_, exists := convertibleTypes[typeInMethodName]
	return exists
}

func (p *ormPlugin) generateCreateHandler(file *generator.FileDescriptor,
	service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {

	inType, outType, methodName, svcName := p.getNames(file, service, method)
	p.P(`//`, generator.CamelCaseSlice(outType.TypeName()))
	if !checkTypeInMethodName("Create", methodName) {
		p.P(`// Cannot autogen create function `, methodName, `: unrecognized anticipated type `)
		return
	}
	if err := validateORMableOutputType("Create", method, outType); err != nil {
		p.Fail(err.Error())
	}

	p.generateDefaultFunctionSignature(inType, outType, methodName)
	p.In()

	p.P(`out := &`, outType.PackageName(), ".", generator.CamelCaseSlice(outType.TypeName()), `{}`)
	var pbDestField, pbDestType string
	if _, exists := convertibleTypes[generator.CamelCaseSlice(outType.TypeName())]; exists {
		pbDestField = `out`
		pbDestType = generator.CamelCaseSlice(outType.TypeName())
	} else if typeMsg, exists := typeNames[generator.CamelCaseSlice(outType.TypeName())]; exists {
		// Check subfields for an ormable object that matches name with the method
		for _, field := range typeMsg.Field {
			rawType := strings.Split(field.GetTypeName(), ".")
			//p.P(`// `, generator.CamelCase(rawType[len(rawType)-1]))

			if _, typeEx := convertibleTypes[generator.CamelCase(rawType[len(rawType)-1])]; typeEx {
				pbDestType = generator.CamelCase(rawType[len(rawType)-1])
				pbDestField = fmt.Sprintf("out.%s", field.GetName())
			}
		}
	}
	if pbDestField == "" {
		p.Fail(method.GetName(), ` output type was not ormable, and had no ormable components`)
	}
	p.P(`tempORM := ConvertTo`, pbDestType, `(in)`)
	p.P(`err := db.Create(&tempORM)`)
	p.P(pbDestField, ` = ConvertFrom`, pbDestType, `(tempORM)`)
	p.P(`return out, err`)
	p.Out()
	p.P(`}`)
	p.P()
	p.generateDefaultHandler(inType, outType, methodName, svcName)
}

func (p *ormPlugin) generateReadHandler(file *generator.FileDescriptor,
	service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {

	inType, outType, methodName, svcName := p.getNames(file, service, method)
	if checkTypeInMethodName("Read", methodName) {
		p.P(`// Cannot autogen read function "%s": unrecognized anticipated type `, methodName)
		return
	}
	if err := validateORMableOutputType("Read", method, outType); err != nil {
		p.Fail(err.Error())
	}
	//typeInMethodName := strings.TrimPrefix(strings.TrimPrefix(methodName, "Read"), "read")
	p.generateDefaultFunctionSignature(inType, outType, methodName)
	p.In()
	p.P(`return nil, nil`)
	p.Out()
	p.P(`}`)
	p.generateDefaultHandler(inType, outType, methodName, svcName)
}

func (p *ormPlugin) generateUpdateHandler(file *generator.FileDescriptor,
	service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {

	inType, outType, methodName, svcName := p.getNames(file, service, method)
	if checkTypeInMethodName("Update", methodName) {
		p.P(`// Cannot autogen update function "%s": unrecognized anticipated type `, methodName)
		return
	}
	if err := validateORMableOutputType("Update", method, outType); err != nil {
		p.Fail(err.Error())
	}
	p.generateDefaultFunctionSignature(inType, outType, methodName)
	p.In()
	p.P(`return nil, nil`)
	p.Out()
	p.P(`}`)
	p.generateDefaultHandler(inType, outType, methodName, svcName)
}

func (p *ormPlugin) generateDeleteHandler(file *generator.FileDescriptor,
	service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {

	inType, outType, methodName, svcName := p.getNames(file, service, method)
	if checkTypeInMethodName("Delete", methodName) {
		p.P(`// Cannot autogen delete function "%s": unrecognized anticipated type `, methodName)
		return
	}
	p.generateDefaultFunctionSignature(inType, outType, methodName)
	p.In()
	p.P(`return nil, nil`)
	p.Out()
	p.P(`}`)
	p.generateDefaultHandler(inType, outType, methodName, svcName)
}

func (p *ormPlugin) generateListHandler(file *generator.FileDescriptor,
	service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {

	inType, outType, methodName, svcName := p.getNames(file, service, method)
	typeInMethodName := strings.TrimPrefix(methodName, "List")
	if _, exists := convertibleTypes[typeInMethodName]; !exists {
		p.P(`// Cannot autogen list function for unrecognized anticipated type `, typeInMethodName)
		return
	}
	p.generateDefaultFunctionSignature(inType, outType, methodName)
	p.In()
	p.P(`return nil, nil`)
	p.Out()
	p.P(`}`)
	p.generateDefaultHandler(inType, outType, methodName, svcName)
}
