package main

import (
	"bytes"
	"fmt"
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

	pkgWKT          generator.Single
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
	p.pkgWKT = p.NewImport("github.com/golang/protobuf/ptypes/wrappers")
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
				p.pkgWKT.Use()
				fieldType = v
			} else if _, exists := convertibleTypes[strings.Trim(fieldType, "[]*")]; !exists {
				p.P("// Can't work with type ", fieldType, ", not tagged as ormable")
				continue
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
					p.P(`to.`, fieldName, ` = append(t.`, fieldName, `, &`, p.pkgWKT.Name(), ".", coreType,
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
				p.P(`to.`, fieldName, ` = &`, p.pkgWKT.Name(), ".", coreType, `{Value: *from.`, fromName, `}`)
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
