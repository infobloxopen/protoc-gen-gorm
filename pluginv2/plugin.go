package pluginv2

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	gorm "github.com/infobloxopen/protoc-gen-gorm/options"
	jgorm "github.com/jinzhu/gorm"
	"github.com/jinzhu/inflection"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

const (
	createService    = "Create"
	readService      = "Read"
	updateService    = "Update"
	updateSetService = "UpdateSet"
	deleteService    = "Delete"
	deleteSetService = "DeleteSet"
	listService      = "List"
)

var (
	gormImport         = "github.com/jinzhu/gorm"
	tkgormImport       = "github.com/infobloxopen/atlas-app-toolkit/gorm"
	uuidImport         = "github.com/satori/go.uuid"
	authImport         = "github.com/infobloxopen/atlas-app-toolkit/auth"
	gormpqImport       = "github.com/jinzhu/gorm/dialects/postgres"
	gtypesImport       = "github.com/infobloxopen/protoc-gen-gorm/types"
	ptypesImport       = "github.com/golang/protobuf/ptypes"
	wktImport          = "github.com/golang/protobuf/ptypes/wrappers"
	resourceImport     = "github.com/infobloxopen/atlas-app-toolkit/gorm/resource"
	fmImport           = "google.golang.org/genproto/protobuf/field_mask"
	queryImport        = "github.com/infobloxopen/atlas-app-toolkit/query"
	ocTraceImport      = "go.opencensus.io/trace"
	gatewayImport      = "github.com/infobloxopen/atlas-app-toolkit/gateway"
	pqImport           = "github.com/lib/pq"
	gerrorsImport      = "github.com/infobloxopen/protoc-gen-gorm/errors"
	timestampImport    = "github.com/golang/protobuf/ptypes/timestamp"
	stdFmtImport       = "fmt"
	stdCtxImport       = "context"
	stdStringsImport   = "strings"
	stdTimeImport      = "time"
	encodingJsonImport = "encoding/json"
)

var builtinTypes = map[string]struct{}{
	"bool":    {},
	"int":     {},
	"int8":    {},
	"int16":   {},
	"int32":   {},
	"int64":   {},
	"uint":    {},
	"uint8":   {},
	"uint16":  {},
	"uint32":  {},
	"uint64":  {},
	"uintptr": {},
	"float32": {},
	"float64": {},
	"string":  {},
	"[]byte":  {},
}

var wellKnownTypes = map[string]string{
	"StringValue": "*string",
	"DoubleValue": "*float64",
	"FloatValue":  "*float32",
	"Int32Value":  "*int32",
	"Int64Value":  "*int64",
	"UInt32Value": "*uint32",
	"UInt64Value": "*uint64",
	"BoolValue":   "*bool",
	//  "BytesValue" : "*[]byte",
}

const (
	protoTypeTimestamp = "Timestamp" // last segment, first will be *google_protobufX
	protoTypeJSON      = "JSONValue"
	protoTypeUUID      = "UUID"
	protoTypeUUIDValue = "UUIDValue"
	protoTypeResource  = "Identifier"
	protoTypeInet      = "InetValue"
	protoTimeOnly      = "TimeOnly"
)

// DB Engine Enum
const (
	ENGINE_UNSET = iota
	ENGINE_POSTGRES
)

type ORMBuilder struct {
	plugin         *protogen.Plugin
	ormableTypes   map[string]*OrmableType
	messages       map[string]struct{}
	fileImports    map[string]*fileImports // TODO: populate
	currentFile    string                  // TODO populate
	currentPackage string
	dbEngine       int
	stringEnums    bool
	gateway        bool
	suppressWarn   bool
}

func New(opts protogen.Options, request *pluginpb.CodeGeneratorRequest) (*ORMBuilder, error) {
	plugin, err := opts.New(request)
	if err != nil {
		return nil, err
	}

	builder := &ORMBuilder{
		plugin:       plugin,
		ormableTypes: make(map[string]*OrmableType),
		messages:     make(map[string]struct{}),
		fileImports:  make(map[string]*fileImports),
	}

	params := parseParameter(request.GetParameter())

	if strings.EqualFold(params["engine"], "postgres") {
		builder.dbEngine = ENGINE_POSTGRES
	} else {
		builder.dbEngine = ENGINE_UNSET
	}

	if strings.EqualFold(params["enums"], "string") {
		builder.stringEnums = true
	}

	if _, ok := params["gateway"]; ok {
		builder.gateway = true
	}

	if _, ok := params["quiet"]; ok {
		builder.suppressWarn = true
	}

	return builder, nil
}

func parseParameter(param string) map[string]string {
	paramMap := make(map[string]string)

	params := strings.Split(param, ",")
	for _, param := range params {
		if strings.Contains(param, "=") {
			kv := strings.Split(param, "=")
			paramMap[kv[0]] = kv[1]
			continue
		}
		paramMap[param] = ""
	}

	return paramMap
}

type OrmableType struct {
	File       *protogen.File
	Fields     map[string]*Field
	Methods    map[string]*autogenMethod
	Name       string
	OriginName string
	Package    string
}

func NewOrmableType(originalName string, pkg string, file *protogen.File) *OrmableType {
	return &OrmableType{
		OriginName: originalName,
		Package:    pkg,
		File:       file,
		Fields:     make(map[string]*Field),
		Methods:    make(map[string]*autogenMethod),
	}
}

type Field struct {
	*gorm.GormFieldOptions
	ParentGoType   string
	Type           string
	Package        string
	ParentOrigName string
}

type autogenMethod struct {
	*protogen.Method
	ccName            string
	verb              string
	followsConvention bool
	baseType          string
	fieldMaskName     string
	inType *protogen.Message
	outType *protogen.Message
}

type fileImports struct {
	wktPkgName      string
	packages        map[string]*pkgImport
	typesToRegister []string
	stdImports      []string
}

type autogenService struct {
	*protogen.Service
	ccName string
	file *protogen.File
	usesTxnMiddleware bool
	methods           []autogenMethod
	autogen           bool
}

func newFileImports() *fileImports {
	return &fileImports{packages: make(map[string]*pkgImport)}
}

type pkgImport struct {
	packagePath string
	alias       string
}

func (b *ORMBuilder) Generate() (*pluginpb.CodeGeneratorResponse, error) {
	genFileMap := make(map[string]*protogen.GeneratedFile)

	for _, protoFile := range b.plugin.Files {
		fileName := protoFile.GeneratedFilenamePrefix + ".pb.gorm.go"
		g := b.plugin.NewGeneratedFile(fileName, ".")
		genFileMap[fileName] = g

		// TODO: set current file and newFileImport
		b.fileImports[*protoFile.Proto.Name] = newFileImports()

		// first traverse: preload the messages
		for _, message := range protoFile.Messages {
			if message.Desc.IsMapEntry() {
				continue
			}

			typeName := string(message.Desc.Name())
			b.messages[typeName] = struct{}{}

			if isOrmable(message) {
				ormable := NewOrmableType(typeName, string(protoFile.GoPackageName), protoFile)
				// TODO: for some reason pluginv1 thinks that we can
				// override values in this map
				b.ormableTypes[typeName] = ormable
			}
		}

		// second traverse: parse basic fields
		for _, message := range protoFile.Messages {
			if isOrmable(message) {
				b.parseBasicFields(message, g)
			}
		}

		// third traverse: build associations
		// TODO: implement functions, simple example will not have any associations
		for _, message := range protoFile.Messages {
			typeName := string(message.Desc.Name())
			if isOrmable(message) {
				b.parseAssociations(message)
				o := b.getOrmable(typeName)
				if b.hasPrimaryKey(o) {
					_, fd := b.findPrimaryKey(o)
					fd.ParentOrigName = o.OriginName
				}
			}
		}

		// Debug
		// ---------------

		// for _, ot := range b.ormableTypes {
		// 	fmt.Fprintf(os.Stderr, "ormable type: %+v\n", ot.Name)
		// 	for name, field := range ot.Fields {
		// 		fmt.Fprintf(os.Stderr, "name: %s, field: %+v\n", name, field.Type)
		// 	}
		// }

		// // dumb files
		// filename := protoFile.GeneratedFilenamePrefix + ".gorm.go"
		// gormFile := b.plugin.NewGeneratedFile(filename, ".")
		// gormFile.P("// this file is generated")
	}

	for _, protoFile := range b.plugin.Files {
		b.parseServices(protoFile)
	}

	for _, protoFile := range b.plugin.Files {
		// generate actual code
		fileName := protoFile.GeneratedFilenamePrefix + ".pb.gorm.go"
		g, ok := genFileMap[fileName]
		if !ok {
			panic("generated file should be present")
		}
		g.P("package ", protoFile.GoPackageName)

		for _, message := range protoFile.Messages {
			if isOrmable(message) {
				b.generateOrmable(g, message)
				b.generateTableNameFunctions(g, message)
				b.generateConvertFunctions(g, message)
				b.generateHookInterfaces(g, message)
			}
		}

		b.generateDefaultHandlers(protoFile, g)
		b.generateDefaultServer(protoFile, g)
	}

	return b.plugin.Response(), nil
}

func (b *ORMBuilder) generateConvertFunctions(g *protogen.GeneratedFile, message *protogen.Message) {
	typeName := string(message.Desc.Name())
	ormable := b.getOrmable(camelCase(typeName))

	///// To Orm
	g.P(`// ToORM runs the BeforeToORM hook if present, converts the fields of this`)
	g.P(`// object to ORM format, runs the AfterToORM hook, then returns the ORM object`)
	g.P(`func (m *`, typeName, `) ToORM (ctx `, generateImport("Context", "context", g), `) (`, typeName, `ORM, error) {`)
	g.P(`to := `, typeName, `ORM{}`)
	g.P(`var err error`)
	g.P(`if prehook, ok := interface{}(m).(`, typeName, `WithBeforeToORM); ok {`)
	g.P(`if err = prehook.BeforeToORM(ctx, &to); err != nil {`)
	g.P(`return to, err`)
	g.P(`}`)
	g.P(`}`)
	for _, field := range message.Fields {
		// Checking if field is skipped
		options := field.Desc.Options().(*descriptorpb.FieldOptions)
		fieldOpts := getFieldOptions(options)
		if fieldOpts.GetDrop() {
			continue
		}

		ofield := ormable.Fields[camelCase(field.GoName)]
		b.generateFieldConversion(message, field, true, ofield, g)
	}
	if getMessageOptions(message).GetMultiAccount() {
		g.P("accountID, err := ", generateImport("GetAccountID", authImport, g), "(ctx, nil)")
		g.P("if err != nil {")
		g.P("return to, err")
		g.P("}")
		g.P("to.AccountID = accountID")
	}
	b.setupOrderedHasMany(message, g)
	g.P(`if posthook, ok := interface{}(m).(`, typeName, `WithAfterToORM); ok {`)
	g.P(`err = posthook.AfterToORM(ctx, &to)`)
	g.P(`}`)
	g.P(`return to, err`)
	g.P(`}`)

	g.P()
	///// To Pb
	g.P(`// ToPB runs the BeforeToPB hook if present, converts the fields of this`)
	g.P(`// object to PB format, runs the AfterToPB hook, then returns the PB object`)
	g.P(`func (m *`, typeName, `ORM) ToPB (ctx context.Context) (`,
		typeName, `, error) {`)
	g.P(`to := `, typeName, `{}`)
	g.P(`var err error`)
	g.P(`if prehook, ok := interface{}(m).(`, typeName, `WithBeforeToPB); ok {`)
	g.P(`if err = prehook.BeforeToPB(ctx, &to); err != nil {`)
	g.P(`return to, err`)
	g.P(`}`)
	g.P(`}`)
	for _, field := range message.Fields {
		// Checking if field is skipped
		options := field.Desc.Options().(*descriptorpb.FieldOptions)
		fieldOpts := getFieldOptions(options)
		if fieldOpts.GetDrop() {
			continue
		}
		ofield := ormable.Fields[generator.CamelCase(field.GoName)]
		b.generateFieldConversion(message, field, false, ofield, g)
	}
	g.P(`if posthook, ok := interface{}(m).(`, typeName, `WithAfterToPB); ok {`)
	g.P(`err = posthook.AfterToPB(ctx, &to)`)
	g.P(`}`)
	g.P(`return to, err`)
	g.P(`}`)
}

func (b *ORMBuilder) generateTableNameFunctions(g *protogen.GeneratedFile, message *protogen.Message) {
	typeName := string(message.Desc.Name())
	msgName := string(message.Desc.Name())

	g.P(`// TableName overrides the default tablename generated by GORM`)
	g.P(`func (`, typeName, `ORM) TableName() string {`)

	tableName := inflection.Plural(jgorm.ToDBName(msgName))
	if opts := getMessageOptions(message); opts != nil && len(opts.Table) > 0 {
		tableName = opts.GetTable()
	}
	g.P(`return "`, tableName, `"`)
	g.P(`}`)
}

func (b *ORMBuilder) generateOrmable(g *protogen.GeneratedFile, message *protogen.Message) {
	ormable := b.getOrmable(message.GoIdent.GoName)
	g.P(`type `, ormable.Name, ` struct {`)

	for name, field := range ormable.Fields { // TODO: sorting, if it's required
		g.P(name, ` `, field.Type, b.renderGormTag(field))
	}

	g.P(`}`)
	g.P()
}

func (b *ORMBuilder) parseAssociations(msg *protogen.Message) {
	typeName := string(msg.Desc.Name()) // TODO: camelSnakeCase
	ormable := b.getOrmable(typeName)

	for _, field := range msg.Fields {
		options := field.Desc.Options().(*descriptorpb.FieldOptions)
		fieldOpts := getFieldOptions(options)
		if fieldOpts.GetDrop() {
			continue
		}

		fieldName := camelCase(string(field.Desc.Name()))
		fieldType := field.Desc.Kind().String() // was GoType
		fieldType = strings.Trim(fieldType, "[]*")
		parts := strings.Split(fieldType, ".")
		fieldTypeShort := parts[len(parts)-1]

		if b.isOrmable(fieldType) {
			if fieldOpts == nil {
				fieldOpts = &gorm.GormFieldOptions{}
			}
			assocOrmable := b.getOrmable(fieldType)

			if field.Desc.Cardinality() == protoreflect.Repeated {
				if fieldOpts.GetManyToMany() != nil {
					b.parseManyToMany(msg, ormable, fieldName, fieldTypeShort, assocOrmable, fieldOpts)
				} else {
					b.parseHasMany(msg, ormable, fieldName, fieldTypeShort, assocOrmable, fieldOpts)
				}
				fieldType = fmt.Sprintf("[]*%sORM", fieldType)
			} else {
				if fieldOpts.GetBelongsTo() != nil {
					b.parseBelongsTo(msg, ormable, fieldName, fieldTypeShort, assocOrmable, fieldOpts)
				} else {
					b.parseHasOne(msg, ormable, fieldName, fieldTypeShort, assocOrmable, fieldOpts)
				}
				fieldType = fmt.Sprintf("*%sORM", fieldType)
			}

			// Register type used, in case it's an imported type from another package
			b.GetFileImports().typesToRegister = append(b.GetFileImports().typesToRegister, fieldType) // maybe we need other fields type
			ormable.Fields[fieldName] = &Field{Type: fieldType, GormFieldOptions: fieldOpts}
		}
	}
}

func (b *ORMBuilder) hasPrimaryKey(ormable *OrmableType) bool {
	for _, field := range ormable.Fields {
		if field.GetTag().GetPrimaryKey() {
			return true
		}
	}
	for fieldName := range ormable.Fields {
		if strings.ToLower(fieldName) == "id" {
			return true
		}
	}
	return false
}

func (b *ORMBuilder) isOrmable(typeName string) bool {
	fmt.Fprintf(os.Stderr, "typeName: %s\n", typeName)
	_, ok := b.ormableTypes[typeName]
	return ok
}

func (b *ORMBuilder) findPrimaryKey(ormable *OrmableType) (string, *Field) {
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

	panic("no primary_key")
}

func (b *ORMBuilder) getOrmable(typeName string) *OrmableType {
	// TODO: implement me
	r, ok := b.ormableTypes[typeName]
	if !ok {
		panic("panic?")
	}
	return r
}

func (b *ORMBuilder) setFile(file string, pkg string) {
	b.currentFile = file
	b.currentPackage = pkg
	// b.Generator.SetFile(file) // TODO: do we need know current file?
}

func (p *ORMBuilder) parseManyToMany(msg *protogen.Message, ormable *OrmableType, fieldName string, fieldType string, assoc *OrmableType, opts *gorm.GormFieldOptions) {
	// TODO: implement me
}

func (p *ORMBuilder) parseHasOne(msg *protogen.Message, parent *OrmableType, fieldName string, fieldType string, child *OrmableType, opts *gorm.GormFieldOptions) {
	// TODO: implement me
}

func (p *ORMBuilder) parseHasMany(msg *protogen.Message, parent *OrmableType, fieldName string, fieldType string, child *OrmableType, opts *gorm.GormFieldOptions) {
	// TODO: implement me
}

func (p *ORMBuilder) parseBelongsTo(msg *protogen.Message, child *OrmableType, fieldName string, fieldType string, parent *OrmableType, opts *gorm.GormFieldOptions) {
	// TODO: implement me
}

func (b *ORMBuilder) parseBasicFields(msg *protogen.Message, g *protogen.GeneratedFile) {
	typeName := string(msg.Desc.Name())

	ormable, ok := b.ormableTypes[typeName]
	if !ok {
		panic("typeName should be found")
	}
	ormable.Name = fmt.Sprintf("%sORM", typeName) // TODO: there are no reason to do it here

	for _, field := range msg.Fields {
		fd := field.Desc
		options := fd.Options().(*descriptorpb.FieldOptions)
		gormOptions := getFieldOptions(options)
		if gormOptions == nil {
			gormOptions = &gorm.GormFieldOptions{}
		}
		if gormOptions.GetDrop() {
			continue
		}

		tag := gormOptions.Tag
		fieldName := camelCase(string(fd.Name())) // TODO: move to camelCase
		fieldType := fd.Kind().String()           // TODO: figure out GoType analog

		var typePackage string

		if b.dbEngine == ENGINE_POSTGRES && b.IsAbleToMakePQArray(fieldType) {
			switch fieldType {
			case "[]bool":
				fieldType = generateImport("BoolArray", pqImport, g)
				gormOptions.Tag = tagWithType(tag, "bool[]")
			case "[]float64":
				fieldType = generateImport("Float64Array", pqImport, g)
				gormOptions.Tag = tagWithType(tag, "float[]")
			case "[]int64":
				fieldType = generateImport("Int64Array", pqImport, g)
				gormOptions.Tag = tagWithType(tag, "integer[]")
			case "[]string":
				fieldType = generateImport("StringArray", pqImport, g)
				gormOptions.Tag = tagWithType(tag, "text[]")
			default:
				continue
			}
		} else if field.Enum != nil {
			fmt.Fprintf(os.Stderr, "field: %s is a enum\n", field.GoName)
			fieldType = "int32"
			if b.stringEnums {
				fieldType = "string"
			}
		} else if field.Message != nil {
			fmt.Fprintf(os.Stderr, "field: %s is a message\n", field.GoName)
		}

		if tName := gormOptions.GetReferenceOf(); tName != "" {
			if _, ok := b.messages[tName]; !ok {
				panic("unknow")
			}
		}

		f := &Field{
			GormFieldOptions: gormOptions,
			ParentGoType:     "",
			Type:             fieldType,
			Package:          typePackage,
			ParentOrigName:   typeName,
		}

		ormable.Fields[fieldName] = f
	}

	gormMsgOptions := getMessageOptions(msg)
	if gormMsgOptions.GetMultiAccount() {
		if accID, ok := ormable.Fields["AccountID"]; !ok {
			ormable.Fields["AccountID"] = &Field{Type: "string"}
		} else if accID.Type != "string" {
			panic("cannot include AccountID field")
		}
	}

	// TODO: GetInclude
	for _, field := range gormMsgOptions.GetInclude() {
		fieldName := field.GetName() // TODO: camel case
		if _, ok := ormable.Fields[fieldName]; !ok {
			b.addIncludedField(ormable, field, g)
		} else {
			panic("cound not include")
		}
	}
}

func (b *ORMBuilder) addIncludedField(ormable *OrmableType, field *gorm.ExtraField, g *protogen.GeneratedFile) {
	fieldName := field.GetName() // TODO: CamelCase
	isPtr := strings.HasPrefix(field.GetType(), "*")
	rawType := strings.TrimPrefix(field.GetType(), "*")
	// cut off any package subpaths
	rawType = rawType[strings.LastIndex(rawType, ".")+1:]
	var typePackage string
	// Handle types with a package defined
	if field.GetPackage() != "" {
		rawType = generateImport(field.GetPackage(), rawType, g)
		typePackage = field.GetPackage()
	} else {
		// Handle types without a package defined
		if _, ok := builtinTypes[rawType]; ok {
			// basic type, 100% okay, no imports or changes needed
		} else if rawType == "Time" {
			// b.UsingGoImports(stdTimeImport) // TODO: missing UsingGoImports
			rawType = generateImport("Time", stdTimeImport, g)
		} else if rawType == "UUID" {
			rawType = generateImport("UUID", uuidImport, g)
		} else if field.GetType() == "Jsonb" && b.dbEngine == ENGINE_POSTGRES {
			rawType = generateImport("Jsonb", gormpqImport, g)
		} else if rawType == "Inet" {
			rawType = generateImport("Inet", gtypesImport, g)
		} else {
			fmt.Fprintf(os.Stderr, "TODO: Warning")
			// p.warning(`included field %q of type %q is not a recognized special type, and no package specified. This type is assumed to be in the same package as the generated code`,
			// 	field.GetName(), field.GetType())
		}
	}
	if isPtr {
		rawType = fmt.Sprintf("*%s", rawType)
	}
	ormable.Fields[fieldName] = &Field{Type: rawType, Package: typePackage, GormFieldOptions: &gorm.GormFieldOptions{Tag: field.GetTag()}}
}

func getFieldOptions(options *descriptorpb.FieldOptions) *gorm.GormFieldOptions {
	if options == nil {
		return nil
	}

	v := proto.GetExtension(options, gorm.E_Field)
	if v == nil {
		return nil
	}

	opts, ok := v.(*gorm.GormFieldOptions)
	if !ok {
		return nil
	}

	return opts
}

// retrieves the GormMessageOptions from a message
func getMessageOptions(message *protogen.Message) *gorm.GormMessageOptions {
	options := message.Desc.Options()
	if options == nil {
		return nil
	}
	v := proto.GetExtension(options, gorm.E_Opts)
	if v != nil {
		return nil
	}

	opts, ok := v.(*gorm.GormMessageOptions)
	if !ok {
		return nil
	}

	return opts
}

func isOrmable(message *protogen.Message) bool {
	desc := message.Desc
	options := desc.Options()

	m, ok := proto.GetExtension(options, gorm.E_Opts).(*gorm.GormMessageOptions)
	if !ok || m == nil {
		return false
	}

	return m.Ormable
}

func (b *ORMBuilder) IsAbleToMakePQArray(fieldType string) bool {
	switch fieldType {
	case "[]bool":
		return true
	case "[]float64":
		return true
	case "[]int64":
		return true
	case "[]string":
		return true
	default:
		return false
	}
}

func (b *ORMBuilder) GetFileImports() *fileImports {
	return b.fileImports[b.currentFile]
}

func tagWithType(tag *gorm.GormTag, typename string) *gorm.GormTag {
	if tag == nil {
		tag = &gorm.GormTag{}
	}

	tag.Type = typename
	return tag
}

func camelCase(s string) string {
	if s == "" {
		return ""
	}
	t := make([]byte, 0, 32)
	i := 0
	if s[0] == '_' {
		// Need a capital letter; drop the '_'.
		t = append(t, 'X')
		i++
	}
	// Invariant: if the next letter is lower case, it must be converted
	// to upper case.
	// That is, we process a word at a time, where words are marked by _ or
	// upper case letter. Digits are treated as words.
	for ; i < len(s); i++ {
		c := s[i]
		if c == '_' && i+1 < len(s) && isASCIILower(s[i+1]) {
			continue // Skip the underscore in s.
		}
		if isASCIIDigit(c) {
			t = append(t, c)
			continue
		}
		// Assume we have a letter now - if not, it's a bogus identifier.
		// The next word is a sequence of characters that must start upper case.
		if isASCIILower(c) {
			c ^= ' ' // Make it a capital letter.
		}
		t = append(t, c) // Guaranteed not lower case.
		// Accept lower case sequence that follows.
		for i+1 < len(s) && isASCIILower(s[i+1]) {
			i++
			t = append(t, s[i])
		}
	}
	return string(t)
}

func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

func isASCIIDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func (p *ORMBuilder) renderGormTag(field *Field) string {
	var gormRes, atlasRes string
	tag := field.GetTag()
	if tag == nil {
		tag = &gorm.GormTag{}
	}

	if len(tag.Column) > 0 {
		gormRes += fmt.Sprintf("column:%s;", tag.GetColumn())
	}
	if len(tag.Type) > 0 {
		gormRes += fmt.Sprintf("type:%s;", tag.GetType())
	}
	if tag.GetSize() > 0 {
		gormRes += fmt.Sprintf("size:%d;", tag.GetSize())
	}
	if tag.Precision > 0 {
		gormRes += fmt.Sprintf("precision:%d;", tag.GetPrecision())
	}
	if tag.GetPrimaryKey() {
		gormRes += "primary_key;"
	}
	if tag.GetUnique() {
		gormRes += "unique;"
	}
	if len(tag.Default) > 0 {
		gormRes += fmt.Sprintf("default:%s;", tag.GetDefault())
	}
	if tag.GetNotNull() {
		gormRes += "not null;"
	}
	if tag.GetAutoIncrement() {
		gormRes += "auto_increment;"
	}
	if len(tag.Index) > 0 {
		if tag.GetIndex() == "" {
			gormRes += "index;"
		} else {
			gormRes += fmt.Sprintf("index:%s;", tag.GetIndex())
		}
	}
	if len(tag.UniqueIndex) > 0 {
		if tag.GetUniqueIndex() == "" {
			gormRes += "unique_index;"
		} else {
			gormRes += fmt.Sprintf("unique_index:%s;", tag.GetUniqueIndex())
		}
	}
	if tag.GetEmbedded() {
		gormRes += "embedded;"
	}
	if len(tag.EmbeddedPrefix) > 0 {
		gormRes += fmt.Sprintf("embedded_prefix:%s;", tag.GetEmbeddedPrefix())
	}
	if tag.GetIgnore() {
		gormRes += "-;"
	}

	var foreignKey, associationForeignKey, joinTable, joinTableForeignKey, associationJoinTableForeignKey string
	var associationAutoupdate, associationAutocreate, associationSaveReference, preload, replace, append, clear bool
	if hasOne := field.GetHasOne(); hasOne != nil {
		foreignKey = hasOne.Foreignkey
		associationForeignKey = hasOne.AssociationForeignkey
		associationAutoupdate = hasOne.AssociationAutoupdate
		associationAutocreate = hasOne.AssociationAutocreate
		associationSaveReference = hasOne.AssociationSaveReference
		preload = hasOne.Preload
		clear = hasOne.Clear
		replace = hasOne.Replace
		append = hasOne.Append
	} else if belongsTo := field.GetBelongsTo(); belongsTo != nil {
		foreignKey = belongsTo.Foreignkey
		associationForeignKey = belongsTo.AssociationForeignkey
		associationAutoupdate = belongsTo.AssociationAutoupdate
		associationAutocreate = belongsTo.AssociationAutocreate
		associationSaveReference = belongsTo.AssociationSaveReference
		preload = belongsTo.Preload
	} else if hasMany := field.GetHasMany(); hasMany != nil {
		foreignKey = hasMany.Foreignkey
		associationForeignKey = hasMany.AssociationForeignkey
		associationAutoupdate = hasMany.AssociationAutoupdate
		associationAutocreate = hasMany.AssociationAutocreate
		associationSaveReference = hasMany.AssociationSaveReference
		clear = hasMany.Clear
		preload = hasMany.Preload
		replace = hasMany.Replace
		append = hasMany.Append
		if len(hasMany.PositionField) > 0 {
			atlasRes += fmt.Sprintf("position:%s;", hasMany.GetPositionField())
		}
	} else if mtm := field.GetManyToMany(); mtm != nil {
		foreignKey = mtm.Foreignkey
		associationForeignKey = mtm.AssociationForeignkey
		joinTable = mtm.Jointable
		joinTableForeignKey = mtm.JointableForeignkey
		associationJoinTableForeignKey = mtm.AssociationJointableForeignkey
		associationAutoupdate = mtm.AssociationAutoupdate
		associationAutocreate = mtm.AssociationAutocreate
		associationSaveReference = mtm.AssociationSaveReference
		preload = mtm.Preload
		clear = mtm.Clear
		replace = mtm.Replace
		append = mtm.Append
	} else {
		foreignKey = tag.Foreignkey
		associationForeignKey = tag.AssociationForeignkey
		joinTable = tag.ManyToMany
		joinTableForeignKey = tag.JointableForeignkey
		associationJoinTableForeignKey = tag.AssociationJointableForeignkey
		associationAutoupdate = tag.AssociationAutoupdate
		associationAutocreate = tag.AssociationAutocreate
		associationSaveReference = tag.AssociationSaveReference
		preload = tag.Preload
	}

	if len(foreignKey) > 0 {
		gormRes += fmt.Sprintf("foreignkey:%s;", foreignKey)
	}

	if len(associationForeignKey) > 0 {
		gormRes += fmt.Sprintf("association_foreignkey:%s;", associationForeignKey)
	}

	if len(joinTable) > 0 {
		gormRes += fmt.Sprintf("many2many:%s;", joinTable)
	}
	if len(joinTableForeignKey) > 0 {
		gormRes += fmt.Sprintf("jointable_foreignkey:%s;", joinTableForeignKey)
	}
	if len(associationJoinTableForeignKey) > 0 {
		gormRes += fmt.Sprintf("association_jointable_foreignkey:%s;", associationJoinTableForeignKey)
	}

	if associationAutoupdate {
		gormRes += fmt.Sprintf("association_autoupdate:%s;", strconv.FormatBool(associationAutoupdate))
	}

	if associationAutocreate {
		gormRes += fmt.Sprintf("association_autocreate:%s;", strconv.FormatBool(associationAutocreate))
	}

	if associationSaveReference {
		gormRes += fmt.Sprintf("association_save_reference:%s;", strconv.FormatBool(associationSaveReference))
	}

	if preload {
		gormRes += fmt.Sprintf("preload:%s;", strconv.FormatBool(preload))
	}

	if clear {
		gormRes += fmt.Sprintf("clear:%s;", strconv.FormatBool(clear))
	} else if replace {
		gormRes += fmt.Sprintf("replace:%s;", strconv.FormatBool(replace))
	} else if append {
		gormRes += fmt.Sprintf("append:%s;", strconv.FormatBool(append))
	}

	var gormTag, atlasTag string
	if gormRes != "" {
		gormTag = fmt.Sprintf("gorm:\"%s\"", strings.TrimRight(gormRes, ";"))
	}
	if atlasRes != "" {
		atlasTag = fmt.Sprintf("atlas:\"%s\"", strings.TrimRight(atlasRes, ";"))
	}
	finalTag := strings.TrimSpace(strings.Join([]string{gormTag, atlasTag}, " "))
	if finalTag == "" {
		return ""
	} else {
		return fmt.Sprintf("`%s`", finalTag)
	}
}

func camelCaseSlice(elem []string) string { return camelCase(strings.Join(elem, "_")) }

func (p *ORMBuilder) setupOrderedHasMany(message *protogen.Message, g *protogen.GeneratedFile) {
	typeName := string(message.Desc.Name())
	ormable := p.getOrmable(typeName)
	for fieldName := range ormable.Fields { // TODO: do we need to sort?
		p.setupOrderedHasManyByName(message, fieldName, g)
	}
}

func (p *ORMBuilder) setupOrderedHasManyByName(message *protogen.Message, fieldName string, g *protogen.GeneratedFile) {
	typeName := string(message.Desc.Name())
	ormable := p.getOrmable(typeName)
	field := ormable.Fields[fieldName]

	if field == nil {
		return
	}

	if field.GetHasMany().GetPositionField() != "" {
		positionField := field.GetHasMany().GetPositionField()
		positionFieldType := p.getOrmable(field.Type).Fields[positionField].Type
		g.P(`for i, e := range `, `to.`, fieldName, `{`)
		g.P(`e.`, positionField, ` = `, positionFieldType, `(i)`)
		g.P(`}`)
	}
}

// Output code that will convert a field to/from orm.
func (b *ORMBuilder) generateFieldConversion(message *protogen.Message, field *protogen.Field,
	toORM bool, ofield *Field, g *protogen.GeneratedFile) error {

	// fieldName := generator.CamelCase(field.GetName())
	// fieldType, _ := p.GoType(message, field)
	fieldName := camelCase(string(field.Desc.Name()))
	fieldType := field.Desc.Kind().String() // was GoType
	if field.Desc.Cardinality() == protoreflect.Repeated {
		// Some repeated fields can be handled by github.com/lib/pq
		if b.dbEngine == ENGINE_POSTGRES && b.IsAbleToMakePQArray(fieldType) {
			g.P(`if m.`, fieldName, ` != nil {`)
			switch fieldType {
			case "[]bool":
				g.P(`to.`, fieldName, ` = make(`, generateImport("BoolArray", pqImport, g), `, len(m.`, fieldName, `))`)
			case "[]float64":
				g.P(`to.`, fieldName, ` = make(`, generateImport("Float64Array", pqImport, g), `, len(m.`, fieldName, `))`)
			case "[]int64":
				g.P(`to.`, fieldName, ` = make(`, generateImport("Int64Array", pqImport, g), `, len(m.`, fieldName, `))`)
			case "[]string":
				g.P(`to.`, fieldName, ` = make(`, generateImport("StringArray", pqImport, g), `, len(m.`, fieldName, `))`)
			}
			g.P(`copy(to.`, fieldName, `, m.`, fieldName, `)`)
			g.P(`}`)
		} else if b.isOrmable(fieldType) { // Repeated ORMable type
			//fieldType = strings.Trim(fieldType, "[]*")

			g.P(`for _, v := range m.`, fieldName, ` {`)
			g.P(`if v != nil {`)
			if toORM {
				g.P(`if temp`, fieldName, `, cErr := v.ToORM(ctx); cErr == nil {`)
			} else {
				g.P(`if temp`, fieldName, `, cErr := v.ToPB(ctx); cErr == nil {`)
			}
			g.P(`to.`, fieldName, ` = append(to.`, fieldName, `, &temp`, fieldName, `)`)
			g.P(`} else {`)
			g.P(`return to, cErr`)
			g.P(`}`)
			g.P(`} else {`)
			g.P(`to.`, fieldName, ` = append(to.`, fieldName, `, nil)`)
			g.P(`}`)
			g.P(`}`) // end repeated for
		} else {
			g.P(`// Repeated type `, fieldType, ` is not an ORMable message type`)
		}
	} else if field.Enum != nil { // Singular Enum, which is an int32 ---
		if toORM {
			if b.stringEnums {
				g.P(`to.`, fieldName, ` = `, fieldType, `_name[int32(m.`, fieldName, `)]`)
			} else {
				g.P(`to.`, fieldName, ` = int32(m.`, fieldName, `)`)
			}
		} else {
			if b.stringEnums {
				g.P(`to.`, fieldName, ` = `, fieldType, `(`, fieldType, `_value[m.`, fieldName, `])`)
			} else {
				g.P(`to.`, fieldName, ` = `, fieldType, `(m.`, fieldName, `)`)
			}
		}
	} else if field.Message != nil { // Singular Object -------------
		//Check for WKTs
		parts := strings.Split(fieldType, ".")
		coreType := parts[len(parts)-1]
		// Type is a WKT, convert to/from as ptr to base type
		if _, exists := wellKnownTypes[coreType]; exists { // Singular WKT -----
			if toORM {
				g.P(`if m.`, fieldName, ` != nil {`)
				g.P(`v := m.`, fieldName, `.Value`)
				g.P(`to.`, fieldName, ` = &v`)
				g.P(`}`)
			} else {
				g.P(`if m.`, fieldName, ` != nil {`)
				g.P(`to.`, fieldName, ` = &`, b.GetFileImports().wktPkgName, ".", coreType,
					`{Value: *m.`, fieldName, `}`)
				g.P(`}`)
			}
		} else if coreType == protoTypeUUIDValue { // Singular UUIDValue type ----
			if toORM {
				g.P(`if m.`, fieldName, ` != nil {`)
				g.P(`tempUUID, uErr := `, generateImport("FromString", uuidImport, g), `(m.`, fieldName, `.Value)`)
				g.P(`if uErr != nil {`)
				g.P(`return to, uErr`)
				g.P(`}`)
				g.P(`to.`, fieldName, ` = &tempUUID`)
				g.P(`}`)
			} else {
				g.P(`if m.`, fieldName, ` != nil {`)
				g.P(`to.`, fieldName, ` = &`, generateImport("UUIDValue", gtypesImport, g), `{Value: m.`, fieldName, `.String()}`)
				g.P(`}`)
			}
		} else if coreType == protoTypeUUID { // Singular UUID type --------------
			if toORM {
				g.P(`if m.`, fieldName, ` != nil {`)
				g.P(`to.`, fieldName, `, err = `, generateImport("FromString", uuidImport, g), `(m.`, fieldName, `.Value)`)
				g.P(`if err != nil {`)
				g.P(`return to, err`)
				g.P(`}`)
				g.P(`} else {`)
				g.P(`to.`, fieldName, ` = `, generateImport("Nil", uuidImport, g))
				g.P(`}`)
			} else {
				g.P(`to.`, fieldName, ` = &`, generateImport("UUID", gtypesImport, g), `{Value: m.`, fieldName, `.String()}`)
			}
		} else if coreType == protoTypeTimestamp { // Singular WKT Timestamp ---
			if toORM {
				g.P(`if m.`, fieldName, ` != nil {`)
				g.P(`var t time.Time`)
				g.P(`if t, err = `, generateImport("Timestamp", ptypesImport, g), `(m.`, fieldName, `); err != nil {`)
				g.P(`return to, err`)
				g.P(`}`)
				g.P(`to.`, fieldName, ` = &t`)
				g.P(`}`)
			} else {
				g.P(`if m.`, fieldName, ` != nil {`)
				g.P(`if to.`, fieldName, `, err = `, generateImport("TimestampProto", ptypesImport, g), `(*m.`, fieldName, `); err != nil {`)
				g.P(`return to, err`)
				g.P(`}`)
				g.P(`}`)
			}
		} else if coreType == protoTypeJSON {
			if b.dbEngine == ENGINE_POSTGRES {
				if toORM {
					g.P(`if m.`, fieldName, ` != nil {`)
					g.P(`to.`, fieldName, ` = &`, generateImport("Jsonb", gormpqImport, g), `{[]byte(m.`, fieldName, `.Value)}`)
					g.P(`}`)
				} else {
					g.P(`if m.`, fieldName, ` != nil {`)
					g.P(`to.`, fieldName, ` = &`, generateImport("JSONValue", gtypesImport, g), `{Value: string(m.`, fieldName, `.RawMessage)}`)
					g.P(`}`)
				}
			} // Potential TODO other DB engine handling if desired
		} else if coreType == protoTypeResource {
			resource := "nil" // assuming we do not know the PB type, nil means call codec for any resource
			if ofield != nil && ofield.ParentOrigName != "" {
				resource = "&" + ofield.ParentOrigName + "{}"
			}
			btype := strings.TrimPrefix(ofield.Type, "*")
			nillable := strings.HasPrefix(ofield.Type, "*")
			iface := ofield.Type == "interface{}"

			if toORM {
				if nillable {
					g.P(`if m.`, fieldName, ` != nil {`)
				}
				switch btype {
				case "int64":
					g.P(`if v, err :=`, generateImport("DecodeInt64", resourceImport, g), `(`, resource, `, m.`, fieldName, `); err != nil {`)
					g.P(`	return to, err`)
					g.P(`} else {`)
					if nillable {
						g.P(`to.`, fieldName, ` = &v`)
					} else {
						g.P(`to.`, fieldName, ` = v`)
					}
					g.P(`}`)
				case "[]byte":
					g.P(`if v, err :=`, generateImport("DecodeBytes", resourceImport, g), `(`, resource, `, m.`, fieldName, `); err != nil {`)
					g.P(`	return to, err`)
					g.P(`} else {`)
					g.P(`	to.`, fieldName, ` = v`)
					g.P(`}`)
				default:
					g.P(`if v, err :=`, generateImport("Decode", resourceImport, g), `(`, resource, `, m.`, fieldName, `); err != nil {`)
					g.P(`return to, err`)
					g.P(`} else if v != nil {`)
					if nillable {
						g.P(`vv := v.(`, btype, `)`)
						g.P(`to.`, fieldName, ` = &vv`)
					} else if iface {
						g.P(`to.`, fieldName, `= v`)
					} else {
						g.P(`to.`, fieldName, ` = v.(`, btype, `)`)
					}
					g.P(`}`)
				}
				if nillable {
					g.P(`}`)
				}
			}

			if !toORM {
				if nillable {
					g.P(`if m.`, fieldName, `!= nil {`)
					g.P(`	if v, err := `, generateImport("Encode", resourceImport, g), `(`, resource, `, *m.`, fieldName, `); err != nil {`)
					g.P(`		return to, err`)
					g.P(`	} else {`)
					g.P(`		to.`, fieldName, ` = v`)
					g.P(`	}`)
					g.P(`}`)

				} else {
					g.P(`if v, err := `, generateImport("Encode", resourceImport, g), `(`, resource, `, m.`, fieldName, `); err != nil {`)
					g.P(`return to, err`)
					g.P(`} else {`)
					g.P(`to.`, fieldName, ` = v`)
					g.P(`}`)
				}
			}
		} else if coreType == protoTypeInet { // Inet type for Postgres only, currently
			if toORM {
				g.P(`if m.`, fieldName, ` != nil {`)
				g.P(`if to.`, fieldName, `, err = `, generateImport("ParseInet", gtypesImport, g), `(m.`, fieldName, `.Value); err != nil {`)
				g.P(`return to, err`)
				g.P(`}`)
				g.P(`}`)
			} else {
				g.P(`if m.`, fieldName, ` != nil && m.`, fieldName, `.IPNet != nil {`)
				g.P(`to.`, fieldName, ` = &`, generateImport("InetValue", gtypesImport, g), `.InetValue{Value: m.`, fieldName, `.String()}`)
				g.P(`}`)
			}
		} else if coreType == protoTimeOnly { // Time only to support time via string
			if toORM {
				g.P(`if m.`, fieldName, ` != nil {`)
				g.P(`if to.`, fieldName, `, err = `, generateImport("ParseTime", gtypesImport, g), `(m.`, fieldName, `.Value); err != nil {`)
				g.P(`return to, err`)
				g.P(`}`)
				g.P(`}`)
			} else {
				g.P(`if m.`, fieldName, ` != "" {`)
				g.P(`if to.`, fieldName, `, err = `, generateImport("TimeOnlyByString", gtypesImport, g), `( m.`, fieldName, `); err != nil {`)
				g.P(`return to, err`)
				g.P(`}`)
				g.P(`}`)
			}
		} else if b.isOrmable(fieldType) {
			// Not a WKT, but a type we're building converters for
			g.P(`if m.`, fieldName, ` != nil {`)
			if toORM {
				g.P(`temp`, fieldName, `, err := m.`, fieldName, `.ToORM (ctx)`)
			} else {
				g.P(`temp`, fieldName, `, err := m.`, fieldName, `.ToPB (ctx)`)
			}
			g.P(`if err != nil {`)
			g.P(`return to, err`)
			g.P(`}`)
			g.P(`to.`, fieldName, ` = &temp`, fieldName)
			g.P(`}`)
		}
	} else { // Singular raw ----------------------------------------------------
		g.P(`to.`, fieldName, ` = m.`, fieldName)
	}
	return nil
}

func (b *ORMBuilder) generateDefaultHandlers(file *protogen.File, g *protogen.GeneratedFile) {
	for _, message := range file.Messages {
		if isOrmable(message) {
			b.generateCreateHandler(message, g)
			typeName := string(message.Desc.Name())
			ormable := b.getOrmable(typeName)

			if b.hasPrimaryKey(ormable) {
				b.generateReadHandler(message, g)
				b.generateDeleteHandler(message, g)
				b.generateDeleteSetHandler(message, g)
				b.generateStrictUpdateHandler(message, g)
				b.generatePatchHandler(message, g)
				b.generatePatchSetHandler(message, g)
			}

			b.generateApplyFieldMask(message, g)
			b.generateListHandler(message, g)
		}

	}
}

func (b *ORMBuilder) generateCreateHandler(message *protogen.Message, g *protogen.GeneratedFile) {
	typeName := string(message.Desc.Name())
	orm := b.getOrmable(typeName)
	g.P(`// DefaultCreate`, typeName, ` executes a basic gorm create call`)
	g.P(`func DefaultCreate`, typeName, `(ctx context.Context, in *`,
		typeName, `, db *`, generateImport("DB", gormImport, g), `) (*`, typeName, `, error) {`)
	g.P(`if in == nil {`)
	g.P(`return nil, `, generateImport("NilArgumentError", gerrorsImport, g))
	g.P(`}`)
	g.P(`ormObj, err := in.ToORM(ctx)`)
	g.P(`if err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)
	create := "Create_"
	b.generateBeforeHookCall(orm, create, g)
	g.P(`if err = db.Create(&ormObj).Error; err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)
	b.generateAfterHookCall(orm, create, g)
	g.P(`pbResponse, err := ormObj.ToPB(ctx)`)
	g.P(`return &pbResponse, err`)
	g.P(`}`)
	b.generateBeforeHookDef(orm, create, g)
	b.generateAfterHookDef(orm, create, g)
}

func (b *ORMBuilder) generateBeforeHookCall(orm *OrmableType, method string, g *protogen.GeneratedFile) {
	g.P(`if hook, ok := interface{}(&ormObj).(`, orm.Name, `WithBefore`, method, `); ok {`)
	g.P(`if db, err = hook.Before`, method, `(ctx, db); err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)
	g.P(`}`)
}

func (b *ORMBuilder) generateAfterHookCall(orm *OrmableType, method string, g *protogen.GeneratedFile) {
	g.P(`if hook, ok := interface{}(&ormObj).(`, orm.Name, `WithAfter`, method, `); ok {`)
	g.P(`if err = hook.After`, method, `(ctx, db); err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)
	g.P(`}`)
}

func (b *ORMBuilder) generateHookInterfaces(g *protogen.GeneratedFile, message *protogen.Message) {
	typeName := string(message.Desc.Name())
	g.P(`// The following are interfaces you can implement for special behavior during ORM/PB conversions`)
	g.P(`// of type `, typeName, ` the arg will be the target, the caller the one being converted from`)
	g.P()
	for _, desc := range [][]string{
		{"BeforeToORM", typeName + "ORM", " called before default ToORM code"},
		{"AfterToORM", typeName + "ORM", " called after default ToORM code"},
		{"BeforeToPB", typeName, " called before default ToPB code"},
		{"AfterToPB", typeName, " called after default ToPB code"},
	} {
		g.P(`// `, typeName, desc[0], desc[2])
		g.P(`type `, typeName, `With`, desc[0], ` interface {`)
		g.P(desc[0], `(context.Context, *`, desc[1], `) error`)
		g.P(`}`)
		g.P()
	}
}

func (b *ORMBuilder) generateReadHandler(message *protogen.Message, g *protogen.GeneratedFile) {
	typeName := string(message.Desc.Name())
	ormable := b.getOrmable(typeName)

	if b.readHasFieldSelection(ormable) {
		// TODO: not implemented return false
		// p.P(`func DefaultRead`, typeName, `(ctx context.Context, in *`,
		// 	typeName, `, db *`, p.Import(gormImport), `.DB, fs *`, p.Import(queryImport), `.FieldSelection) (*`, typeName, `, error) {`)
	} else {
		g.P(`func DefaultRead`, typeName, `(ctx context.Context, in *`,
			typeName, `, db *`, "gorm", `.DB) (*`, typeName, `, error) {`)
	}
	g.P(`if in == nil {`)
	g.P(`return nil, `, "errors", `.NilArgumentError`)
	g.P(`}`)

	g.P(`ormObj, err := in.ToORM(ctx)`)
	g.P(`if err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)

	k, f := b.findPrimaryKey(ormable)
	if strings.Contains(f.Type, "*") {
		g.P(`if ormObj.`, k, ` == nil || *ormObj.`, k, ` == `, b.guessZeroValue(f.Type), ` {`)
	} else {
		g.P(`if ormObj.`, k, ` == `, b.guessZeroValue(f.Type), ` {`)
	}
	g.P(`return nil, `, "errors", `.EmptyIdError`)
	g.P(`}`)

	var fs string
	if b.readHasFieldSelection(ormable) {
		fs = "fs"
	} else {
		fs = "nil"
	}

	b.generateBeforeReadHookCall(ormable, "ApplyQuery", g)
	g.P(`if db, err = `, generateImport("ApplyFieldSelection", tkgormImport, g), `(ctx, db, `, fs, `, &`, ormable.Name, `{}); err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)

	b.generateBeforeReadHookCall(ormable, "Find", g)
	g.P(`ormResponse := `, ormable.Name, `{}`)
	g.P(`if err = db.Where(&ormObj).First(&ormResponse).Error; err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)

	b.generateAfterReadHookCall(ormable, g)
	g.P(`pbResponse, err := ormResponse.ToPB(ctx)`)
	g.P(`return &pbResponse, err`)
	g.P(`}`)

	b.generateBeforeReadHookDef(ormable, "ApplyQuery", g)
	b.generateBeforeReadHookDef(ormable, "Find", g)
	b.generateAfterReadHookDef(ormable, g)

}

func (b *ORMBuilder) readHasFieldSelection(ormable *OrmableType) bool {
	return false
}

// guessZeroValue of the input type, so that we can check if a (key) value is set or not
func (b *ORMBuilder) guessZeroValue(typeName string) string {
	typeName = strings.ToLower(typeName)
	if strings.Contains(typeName, "string") {
		return `""`
	}
	if strings.Contains(typeName, "int") {
		return `0`
	}
	// TODO: import uuid
	// if strings.Contains(typeName, "uuid") {
	// 	return fmt.Sprintf(`%s.Nil`, p.Import(uuidImport))
	// }
	if strings.Contains(typeName, "[]byte") {
		return `nil`
	}
	if strings.Contains(typeName, "bool") {
		return `false`
	}
	return ``
}

func (b *ORMBuilder) generateBeforeReadHookCall(orm *OrmableType, suffix string, g *protogen.GeneratedFile) {
	g.P(`if hook, ok := interface{}(&ormObj).(`, orm.Name, `WithBeforeRead`, suffix, `); ok {`)
	hookCall := fmt.Sprint(`if db, err = hook.BeforeRead`, suffix, `(ctx, db`)
	if b.readHasFieldSelection(orm) {
		hookCall += `, fs`
	}
	hookCall += `); err != nil{`
	g.P(hookCall)
	g.P(`return nil, err`)
	g.P(`}`)
	g.P(`}`)
}

func (b *ORMBuilder) generateAfterReadHookCall(orm *OrmableType, g *protogen.GeneratedFile) {
	g.P(`if hook, ok := interface{}(&ormResponse).(`, orm.Name, `WithAfterReadFind`, `); ok {`)
	hookCall := `if err = hook.AfterReadFind(ctx, db`
	if b.readHasFieldSelection(orm) {
		hookCall += `, fs`
	}
	hookCall += `); err != nil {`
	g.P(hookCall)
	g.P(`return nil, err`)
	g.P(`}`)
	g.P(`}`)
}

func (b *ORMBuilder) generateBeforeReadHookDef(orm *OrmableType, suffix string, g *protogen.GeneratedFile) {
	gormDB := generateImport("DB", gormImport, g)
	g.P(`type `, orm.Name, `WithBeforeRead`, suffix, ` interface {`)
	hookSign := fmt.Sprint(`BeforeRead`, suffix, `(context.Context, *`, gormDB)
	if b.readHasFieldSelection(orm) {
		hookSign += fmt.Sprint(`, *`, generateImport("FieldSelection", queryImport, g))
	}

	hookSign += fmt.Sprint(`) (*`, gormDB, `, error)`)
	g.P(hookSign)
	g.P(`}`)
}

func (b *ORMBuilder) generateAfterReadHookDef(orm *OrmableType, g *protogen.GeneratedFile) {
	g.P(`type `, orm.Name, `WithAfterReadFind interface {`)
	hookSign := fmt.Sprint(`AfterReadFind`, `(context.Context, *`, generateImport("DB", gormImport, g))
	if b.readHasFieldSelection(orm) {
		hookSign += fmt.Sprint(`, *`, generateImport("FieldSelection", queryImport, g))
	}
	hookSign += `) error`
	g.P(hookSign)
	g.P(`}`)
}

func (b *ORMBuilder) generateDeleteHandler(message *protogen.Message, g *protogen.GeneratedFile) {
	typeName := string(message.Desc.Name())

	g.P(`func DefaultDelete`, typeName, `(ctx context.Context, in *`,
		typeName, `, db *`, generateImport("DB", gormImport, g), `) error {`)
	g.P(`if in == nil {`)
	g.P(`return `, generateImport("NilArgumentError", gerrorsImport, g))
	g.P(`}`)
	g.P(`ormObj, err := in.ToORM(ctx)`)
	g.P(`if err != nil {`)
	g.P(`return err`)
	g.P(`}`)

	ormable := b.getOrmable(typeName)
	pkName, pk := b.findPrimaryKey(ormable)
	if strings.Contains(pk.Type, "*") {
		g.P(`if ormObj.`, pkName, ` == nil || *ormObj.`, pkName, ` == `, b.guessZeroValue(pk.Type), ` {`)
	} else {
		g.P(`if ormObj.`, pkName, ` == `, b.guessZeroValue(pk.Type), `{`)
	}
	g.P(`return `, generateImport("EmptyIdError", gerrorsImport, g))
	g.P(`}`)

	b.generateBeforeDeleteHookCall(ormable, g)
	g.P(`err = db.Where(&ormObj).Delete(&`, ormable.Name, `{}).Error`)
	g.P(`if err != nil {`)
	g.P(`return err`)
	g.P(`}`)

	b.generateAfterDeleteHookCall(ormable, g)
	g.P(`return err`)
	g.P(`}`)
	delete := "Delete_"
	b.generateBeforeHookDef(ormable, delete, g)
	b.generateAfterHookDef(ormable, delete, g)
}

func (b *ORMBuilder) generateBeforeDeleteHookCall(orm *OrmableType, g *protogen.GeneratedFile) {
	g.P(`if hook, ok := interface{}(&ormObj).(`, orm.Name, `WithBeforeDelete_); ok {`)
	g.P(`if db, err = hook.BeforeDelete_(ctx, db); err != nil {`)
	g.P(`return err`)
	g.P(`}`)
	g.P(`}`)
}

func (b *ORMBuilder) generateAfterDeleteHookCall(orm *OrmableType, g *protogen.GeneratedFile) {
	g.P(`if hook, ok := interface{}(&ormObj).(`, orm.Name, `WithAfterDelete_); ok {`)
	g.P(`err = hook.AfterDelete_(ctx, db)`)
	g.P(`}`)
}

func (b *ORMBuilder) generateBeforeHookDef(orm *OrmableType, method string, g *protogen.GeneratedFile) {
	gormDB := generateImport("DB", gormImport, g)
	g.P(`type `, orm.Name, `WithBefore`, method, ` interface {`)
	g.P(`Before`, method, `(context.Context, *`, gormDB, `) (*`, gormDB, `, error)`)
	g.P(`}`)
}

func (b *ORMBuilder) generateAfterHookDef(orm *OrmableType, method string, g *protogen.GeneratedFile) {
	g.P(`type `, orm.Name, `WithAfter`, method, ` interface {`)
	g.P(`After`, method, `(context.Context, *`, generateImport("DB", gormImport, g), `) error`)
	g.P(`}`)
}

func (b *ORMBuilder) generateDeleteSetHandler(message *protogen.Message, g *protogen.GeneratedFile) {
	typeName := string(message.Desc.Name())
	gormDB := generateImport("DB", gormImport, g)

	g.P(`func DefaultDelete`, typeName, `Set(ctx context.Context, in []*`,
		typeName, `, db *`, gormDB, `) error {`)
	g.P(`if in == nil {`)
	g.P(`return `, generateImport("NilArgumentError", gerrorsImport, g))
	g.P(`}`)
	g.P(`var err error`)
	ormable := b.getOrmable(typeName)
	pkName, pk := b.findPrimaryKey(ormable)
	g.P(`keys := []`, pk.Type, `{}`)
	g.P(`for _, obj := range in {`)
	g.P(`ormObj, err := obj.ToORM(ctx)`)
	g.P(`if err != nil {`)
	g.P(`return err`)
	g.P(`}`)
	if strings.Contains(pk.Type, "*") {
		g.P(`if ormObj.`, pkName, ` == nil || *ormObj.`, pkName, ` == `, b.guessZeroValue(pk.Type), ` {`)
	} else {
		g.P(`if ormObj.`, pkName, ` == `, b.guessZeroValue(pk.Type), `{`)
	}
	g.P(`return `, generateImport("EmptyIdError", gerrorsImport, g))
	g.P(`}`)
	g.P(`keys = append(keys, ormObj.`, pkName, `)`)
	g.P(`}`)
	b.generateBeforeDeleteSetHookCall(ormable, g)
	if getMessageOptions(message).GetMultiAccount() {
		g.P(`acctId, err := `, generateImport("GetAccountID", authImport, g), `(ctx, nil)`)
		g.P(`if err != nil {`)
		g.P(`return err`)
		g.P(`}`)
		g.P(`err = db.Where("account_id = ? AND `, jgorm.ToDBName(pkName), ` in (?)", acctId, keys).Delete(&`, ormable.Name, `{}).Error`)
	} else {
		g.P(`err = db.Where("`, jgorm.ToDBName(pkName), ` in (?)", keys).Delete(&`, ormable.Name, `{}).Error`)
	}
	g.P(`if err != nil {`)
	g.P(`return err`)
	g.P(`}`)
	b.generateAfterDeleteSetHookCall(ormable, g)
	g.P(`return err`)
	g.P(`}`)
	g.P(`type `, ormable.Name, `WithBeforeDeleteSet interface {`)
	g.P(`BeforeDeleteSet(context.Context, []*`, ormable.OriginName, `, *`, gormDB, `) (*`, gormDB, `, error)`)
	g.P(`}`)
	g.P(`type `, ormable.Name, `WithAfterDeleteSet interface {`)
	g.P(`AfterDeleteSet(context.Context, []*`, ormable.OriginName, `, *`, gormDB, `) error`)
	g.P(`}`)
}

func (b *ORMBuilder) generateBeforeDeleteSetHookCall(orm *OrmableType, g *protogen.GeneratedFile) {
	g.P(`if hook, ok := (interface{}(&`, orm.Name, `{})).(`, orm.Name, `WithBeforeDeleteSet); ok {`)
	g.P(`if db, err = hook.BeforeDeleteSet(ctx, in, db); err != nil {`)
	g.P(`return err`)
	g.P(`}`)
	g.P(`}`)
}

func (b *ORMBuilder) generateAfterDeleteSetHookCall(orm *OrmableType, g *protogen.GeneratedFile) {
	g.P(`if hook, ok := (interface{}(&`, orm.Name, `{})).(`, orm.Name, `WithAfterDeleteSet); ok {`)
	g.P(`err = hook.AfterDeleteSet(ctx, in, db)`)
	g.P(`}`)
}

func (b *ORMBuilder) generateStrictUpdateHandler(message *protogen.Message, g *protogen.GeneratedFile) {
	_ = generateImport("", "fmt", g)
	typeName := string(message.Desc.Name())

	g.P(`// DefaultStrictUpdate`, typeName, ` clears / replaces / appends first level 1:many children and then executes a gorm update call`)
	g.P(`func DefaultStrictUpdate`, typeName, `(ctx context.Context, in *`,
		typeName, `, db *`, generateImport("DB", gormImport, g), `) (*`, typeName, `, error) {`)
	g.P(`if in == nil {`)
	g.P(`return nil, fmt.Errorf("Nil argument to DefaultStrictUpdate`, typeName, `")`)
	g.P(`}`)
	g.P(`ormObj, err := in.ToORM(ctx)`)
	g.P(`if err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)

	if getMessageOptions(message).GetMultiAccount() {
		b.generateAccountIdWhereClause(g)
	}

	ormable := b.getOrmable(typeName)
	if b.gateway {
		g.P(`var count int64`)
	}

	if b.hasPrimaryKey(ormable) {
		pkName, pk := b.findPrimaryKey(ormable)
		column := pk.GetTag().GetColumn()
		if len(column) == 0 {
			column = jgorm.ToDBName(pkName)
		}
		g.P(`lockedRow := &`, typeName, `ORM{}`)
		var count string
		var rowsAffected string
		if b.gateway {
			count = `count = `
			rowsAffected = `.RowsAffected`
		}
		g.P(count+`db.Model(&ormObj).Set("gorm:query_option", "FOR UPDATE").Where("`, column, `=?", ormObj.`, pkName, `).First(lockedRow)`+rowsAffected)
	}
	b.generateBeforeHookCall(ormable, "StrictUpdateCleanup", g)
	b.handleChildAssociations(message, g)
	b.generateBeforeHookCall(ormable, "StrictUpdateSave", g)
	g.P(`if err = db.Save(&ormObj).Error; err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)
	b.generateAfterHookCall(ormable, "StrictUpdateSave", g)
	g.P(`pbResponse, err := ormObj.ToPB(ctx)`)
	g.P(`if err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)

	if b.gateway {
		g.P(`if count == 0 {`)
		g.P(`err = `, generateImport("SetCreated", gatewayImport, g), `(ctx, "")`)
		g.P(`}`)
	}

	g.P(`return &pbResponse, err`)
	g.P(`}`)

	b.generateBeforeHookDef(ormable, "StrictUpdateCleanup", g)
	b.generateBeforeHookDef(ormable, "StrictUpdateSave", g)
	b.generateAfterHookDef(ormable, "StrictUpdateSave", g)
}

func (b *ORMBuilder) generateAccountIdWhereClause(g *protogen.GeneratedFile) {
	g.P(`accountID, err := `, generateImport("GetAccountID", authImport, g), `(ctx, nil)`)
	g.P(`if err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)
	g.P(`db = db.Where(map[string]interface{}{"account_id": accountID})`)
}

func (b *ORMBuilder) handleChildAssociations(message *protogen.Message, g *protogen.GeneratedFile) {
	ormable := b.getOrmable(string(message.Desc.Name()))

	for fieldName := range ormable.Fields { // TODO: skipped sorting
		b.handleChildAssociationsByName(message, fieldName, g)
	}
}

func (b *ORMBuilder) handleChildAssociationsByName(message *protogen.Message, fieldName string, g *protogen.GeneratedFile) {
	typeName := string(message.Desc.Name())
	ormable := b.getOrmable(typeName)
	field := ormable.Fields[fieldName]

	if field == nil {
		return
	}

	if field.GetHasMany() != nil || field.GetHasOne() != nil || field.GetManyToMany() != nil {
		var assocHandler string
		switch {
		case field.GetHasMany() != nil:
			switch {
			case field.GetHasMany().GetClear():
				assocHandler = "Clear"
			case field.GetHasMany().GetAppend():
				assocHandler = "Append"
			case field.GetHasMany().GetReplace():
				assocHandler = "Replace"
			default:
				assocHandler = "Remove"
			}
		case field.GetHasOne() != nil:
			switch {
			case field.GetHasOne().GetClear():
				assocHandler = "Clear"
			case field.GetHasOne().GetAppend():
				assocHandler = "Append"
			case field.GetHasOne().GetReplace():
				assocHandler = "Replace"
			default:
				assocHandler = "Remove"
			}
		case field.GetManyToMany() != nil:
			switch {
			case field.GetManyToMany().GetClear():
				assocHandler = "Clear"
			case field.GetManyToMany().GetAppend():
				assocHandler = "Append"
			case field.GetManyToMany().GetReplace():
				assocHandler = "Replace"
			default:
				assocHandler = "Replace"
			}
		}

		if assocHandler == "Remove" {
			b.removeChildAssociationsByName(message, fieldName, g)
			return
		}

		action := fmt.Sprintf("%s(ormObj.%s)", assocHandler, fieldName)
		if assocHandler == "Clear" {
			action = fmt.Sprintf("%s()", assocHandler)
		}

		g.P(`if err = db.Model(&ormObj).Association("`, fieldName, `").`, action, `.Error; err != nil {`)
		g.P(`return nil, err`)
		g.P(`}`)
		g.P(`ormObj.`, fieldName, ` = nil`)
	}
}

func (b *ORMBuilder) removeChildAssociationsByName(message *protogen.Message, fieldName string, g *protogen.GeneratedFile) {
	typeName := string(message.Desc.Name())
	ormable := b.getOrmable(typeName)
	field := ormable.Fields[fieldName]

	if field == nil {
		return
	}

	if field.GetHasMany() != nil || field.GetHasOne() != nil {
		var assocKeyName, foreignKeyName string
		switch {
		case field.GetHasMany() != nil:
			assocKeyName = field.GetHasMany().GetAssociationForeignkey()
			foreignKeyName = field.GetHasMany().GetForeignkey()
		case field.GetHasOne() != nil:
			assocKeyName = field.GetHasOne().GetAssociationForeignkey()
			foreignKeyName = field.GetHasOne().GetForeignkey()
		}
		assocKeyType := ormable.Fields[assocKeyName].Type
		assocOrmable := b.getOrmable(field.Type)
		foreignKeyType := assocOrmable.Fields[foreignKeyName].Type
		g.P(`filter`, fieldName, ` := `, strings.Trim(field.Type, "[]*"), `{}`)
		zeroValue := b.guessZeroValue(assocKeyType)
		if strings.Contains(assocKeyType, "*") {
			g.P(`if ormObj.`, assocKeyName, ` == nil || *ormObj.`, assocKeyName, ` == `, zeroValue, `{`)
		} else {
			g.P(`if ormObj.`, assocKeyName, ` == `, zeroValue, `{`)
		}
		g.P(`return nil, `, generateImport("EmptyIdError", gerrorsImport, g), `.EmptyIdError`)
		g.P(`}`)
		filterDesc := "filter" + fieldName + "." + foreignKeyName
		ormDesc := "ormObj." + assocKeyName
		if strings.HasPrefix(foreignKeyType, "*") {
			g.P(filterDesc, ` = new(`, strings.TrimPrefix(foreignKeyType, "*"), `)`)
			filterDesc = "*" + filterDesc
		}
		if strings.HasPrefix(assocKeyType, "*") {
			ormDesc = "*" + ormDesc
		}
		g.P(filterDesc, " = ", ormDesc)
		g.P(`if err = db.Where(filter`, fieldName, `).Delete(`, strings.Trim(field.Type, "[]*"), `{}).Error; err != nil {`)
		g.P(`return nil, err`)
		g.P(`}`)
	}
}

func (b *ORMBuilder) generatePatchHandler(message *protogen.Message, g *protogen.GeneratedFile) {
	var isMultiAccount bool

	typeName := string(message.Desc.Name())
	ormable := b.getOrmable(typeName)

	if getMessageOptions(message).GetMultiAccount() {
		isMultiAccount = true
	}

	if isMultiAccount && !b.hasIDField(message) {
		g.P(fmt.Sprintf("// Cannot autogen DefaultPatch%s: this is a multi-account table without an \"id\" field in the message.\n", typeName))
		return
	}

	g.P(`// DefaultPatch`, typeName, ` executes a basic gorm update call with patch behavior`)
	g.P(`func DefaultPatch`, typeName, `(ctx context.Context, in *`,
		typeName, `, updateMask *`, generateImport("FieldMask", fmImport, g), `, db *`, generateImport("DB", gormImport, g), `) (*`, typeName, `, error) {`)

	g.P(`if in == nil {`)
	g.P(`return nil, `, generateImport("NilArgumentError", gerrorsImport, g))
	g.P(`}`)
	g.P(`var pbObj `, typeName)
	g.P(`var err error`)
	b.generateBeforePatchHookCall(ormable, "Read", g)
	if b.readHasFieldSelection(ormable) {
		g.P(`pbReadRes, err := DefaultRead`, typeName, `(ctx, &`, typeName, `{Id: in.GetId()}, db, nil)`)
	} else {
		g.P(`pbReadRes, err := DefaultRead`, typeName, `(ctx, &`, typeName, `{Id: in.GetId()}, db)`)
	}

	g.P(`if err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)

	g.P(`pbObj = *pbReadRes`)

	b.generateBeforePatchHookCall(ormable, "ApplyFieldMask", g)
	g.P(`if _, err := DefaultApplyFieldMask`, typeName, `(ctx, &pbObj, in, updateMask, "", db); err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)

	b.generateBeforePatchHookCall(ormable, "Save", g)
	g.P(`pbResponse, err := DefaultStrictUpdate`, typeName, `(ctx, &pbObj, db)`)
	g.P(`if err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)
	b.generateAfterPatchHookCall(ormable, "Save", g)

	g.P(`return pbResponse, nil`)
	g.P(`}`)

	b.generateBeforePatchHookDef(ormable, "Read", g)
	b.generateBeforePatchHookDef(ormable, "ApplyFieldMask", g)
	b.generateBeforePatchHookDef(ormable, "Save", g)
	b.generateAfterPatchHookDef(ormable, "Save", g)

}

func (b *ORMBuilder) hasIDField(message *protogen.Message) bool {
	for _, field := range message.Fields {
		if strings.ToLower(field.GoName) == "id" { // TODO: not sure
			return true
		}
	}

	return false
}

func (b *ORMBuilder) generateBeforePatchHookCall(orm *OrmableType, suffix string, g *protogen.GeneratedFile) {
	g.P(`if hook, ok := interface{}(&pbObj).(`, orm.OriginName, `WithBeforePatch`, suffix, `); ok {`)
	g.P(`if db, err = hook.BeforePatch`, suffix, `(ctx, in, updateMask, db); err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)
	g.P(`}`)
}

func (b *ORMBuilder) generateAfterPatchHookCall(orm *OrmableType, suffix string, g *protogen.GeneratedFile) {
	g.P(`if hook, ok := interface{}(pbResponse).(`, orm.OriginName, `WithAfterPatch`, suffix, `); ok {`)
	g.P(`if err = hook.AfterPatch`, suffix, `(ctx, in, updateMask, db); err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)
	g.P(`}`)
}

func (b *ORMBuilder) generateBeforePatchHookDef(orm *OrmableType, suffix string, g *protogen.GeneratedFile) {
	g.P(`type `, orm.OriginName, `WithBeforePatch`, suffix, ` interface {`)
	g.P(`BeforePatch`, suffix, `(context.Context, *`, orm.OriginName, `, *`, generateImport("FieldMask", fmImport, g), `, *`, generateImport("DB", gormImport, g),
		`) (*`, generateImport("DB", gormImport, g), `, error)`)
	g.P(`}`)
}

func (b *ORMBuilder) generateAfterPatchHookDef(orm *OrmableType, suffix string, g *protogen.GeneratedFile) {
	g.P(`type `, orm.OriginName, `WithAfterPatch`, suffix, ` interface {`)
	g.P(`AfterPatch`, suffix, `(context.Context, *`, orm.OriginName, `, *`, generateImport("FieldMask", fmImport, g), `, *`, generateImport("DB", gormImport, g),
		`) error`)
	g.P(`}`)
}

func (b *ORMBuilder) generatePatchSetHandler(message *protogen.Message, g *protogen.GeneratedFile) {
	var isMultiAccount bool

	typeName := string(message.Desc.Name())
	if getMessageOptions(message).GetMultiAccount() {
		isMultiAccount = true
	}

	if isMultiAccount && !b.hasIDField(message) {
		g.P(fmt.Sprintf("// Cannot autogen DefaultPatchSet%s: this is a multi-account table without an \"id\" field in the message.\n", typeName))
		return
	}

	_ = generateImport("", "fmt", g)
	g.P(`// DefaultPatchSet`, typeName, ` executes a bulk gorm update call with patch behavior`)
	g.P(`func DefaultPatchSet`, typeName, `(ctx context.Context, objects []*`,
		typeName, `, updateMasks []*`, generateImport("FieldMask", fmImport, g), `, db *`, generateImport("DB", gormImport, g), `) ([]*`, typeName, `, error) {`)
	g.P(`if len(objects) != len(updateMasks) {`)
	g.P(`return nil, fmt.Errorf(`, generateImport("BadRepeatedFieldMaskTpl", gerrorsImport, g), `, len(updateMasks), len(objects))`)
	g.P(`}`)
	g.P(``)
	g.P(`results := make([]*`, typeName, `, 0, len(objects))`)
	g.P(`for i, patcher := range objects {`)
	g.P(`pbResponse, err := DefaultPatch`, typeName, `(ctx, patcher, updateMasks[i], db)`)
	g.P(`if err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)
	g.P(``)
	g.P(`results = append(results, pbResponse)`)
	g.P(`}`)
	g.P(``)
	g.P(`return results, nil`)
	g.P(`}`)

}

func (b *ORMBuilder) generateApplyFieldMask(message *protogen.Message, g *protogen.GeneratedFile) {
	typeName := string(message.Desc.Name())
	g.P(`// DefaultApplyFieldMask`, typeName, ` patches an pbObject with patcher according to a field mask.`)
	g.P(`func DefaultApplyFieldMask`, typeName, `(ctx context.Context, patchee *`,
		typeName, `, patcher *`, typeName, `, updateMask *`, generateImport("FieldMask", fmImport, g),
		`, prefix string, db *`, generateImport("DB", gormImport, g), `) (*`, typeName, `, error) {`)

	g.P(`if patcher == nil {`)
	g.P(`return nil, nil`)
	g.P(`} else if patchee == nil {`)
	g.P(`return nil, `, generateImport("NilArgumentError", gerrorsImport, g))
	g.P(`}`)
	g.P(`var err error`)

	hasNested := false
	for _, field := range message.Fields {
		fieldType := field.Desc.Kind().String()

		if field.Message != nil && !isSpecialType(fieldType) && field.Desc.Cardinality() != protoreflect.Repeated {
			g.P(`var updated`, camelCase(field.GoName), ` bool`)
			hasNested = true
		} else if strings.HasSuffix(fieldType, protoTypeJSON) {
			g.P(`var updated`, camelCase(field.GoName), ` bool`)
		}
	}

	// Patch pbObj with input according to a field mask.
	if hasNested {
		g.P(`for i, f := range updateMask.Paths {`)
	} else {
		g.P(`for _, f := range updateMask.Paths {`)
	}
	for _, field := range message.Fields {
		ccName := camelCase(field.GoName)
		fieldType := field.Desc.Kind().String()
		//  for ormable message, do recursive patching

		if field.Message != nil && b.isOrmable(fieldType) && field.Desc.Cardinality() != protoreflect.Repeated {
			_ = generateImport("", "stdStringsImport", g)

			g.P(`if !updated`, ccName, ` && strings.HasPrefix(f, prefix+"`, ccName, `.") {`)
			g.P(`updated`, ccName, ` = true`)
			g.P(`if patcher.`, ccName, ` == nil {`)
			g.P(`patchee.`, ccName, ` = nil`)
			g.P(`continue`)
			g.P(`}`)
			g.P(`if patchee.`, ccName, ` == nil {`)
			g.P(`patchee.`, ccName, ` = &`, strings.TrimPrefix(fieldType, "*"), `{}`)
			g.P(`}`)
			if s := strings.Split(fieldType, "."); len(s) == 2 {
				g.P(`if o, err := `, strings.TrimLeft(s[0], "*"), `.DefaultApplyFieldMask`, s[1], `(ctx, patchee.`, ccName,
					`, patcher.`, ccName, `, &`, generateImport("FieldMask", fmImport, g),
					`{Paths:updateMask.Paths[i:]}, prefix+"`, ccName, `.", db); err != nil {`)
			} else {
				g.P(`if o, err := DefaultApplyFieldMask`, strings.TrimPrefix(fieldType, "*"), `(ctx, patchee.`, ccName,
					`, patcher.`, ccName, `, &`, generateImport("FieldMask", fmImport, g),
					`{Paths:updateMask.Paths[i:]}, prefix+"`, ccName, `.", db); err != nil {`)
			}
			g.P(`return nil, err`)
			g.P(`} else {`)
			g.P(`patchee.`, ccName, ` = o`)
			g.P(`}`)
			g.P(`continue`)
			g.P(`}`)
			g.P(`if f == prefix+"`, ccName, `" {`)
			g.P(`updated`, ccName, ` = true`)
			g.P(`patchee.`, ccName, ` = patcher.`, ccName)
			g.P(`continue`)
			g.P(`}`)
		} else if field.Message != nil && !isSpecialType(fieldType) && field.Desc.Cardinality() != protoreflect.Repeated {
			_ = generateImport("", "stdStringsImport", g)
			g.P(`if !updated`, ccName, ` && strings.HasPrefix(f, prefix+"`, ccName, `.") {`)
			g.P(`if patcher.`, ccName, ` == nil {`)
			g.P(`patchee.`, ccName, ` = nil`)
			g.P(`continue`)
			g.P(`}`)
			g.P(`if patchee.`, ccName, ` == nil {`)
			g.P(`patchee.`, ccName, ` = &`, strings.TrimPrefix(fieldType, "*"), `{}`)
			g.P(`}`)
			g.P(`childMask := &`, generateImport("FieldMask", fmImport, g), `{}`)
			g.P(`for j := i; j < len(updateMask.Paths); j++ {`)
			g.P(`if trimPath := strings.TrimPrefix(updateMask.Paths[j], prefix+"`, ccName, `."); trimPath != updateMask.Paths[j] {`)
			g.P(`childMask.Paths = append(childMask.Paths, trimPath)`)
			g.P(`}`)
			g.P(`}`)
			g.P(`if err := `, generateImport("MergeWithMask", tkgormImport, g), `.(patcher.`, ccName, `, patchee.`, ccName, `, childMask); err != nil {`)
			g.P(`return nil, nil`)
			g.P(`}`)
			g.P(`}`)
			g.P(`if f == prefix+"`, ccName, `" {`)
			g.P(`updated`, ccName, ` = true`)
			g.P(`patchee.`, ccName, ` = patcher.`, ccName)
			g.P(`continue`)
			g.P(`}`)
		} else if strings.HasSuffix(fieldType, protoTypeJSON) && field.Desc.Cardinality() != protoreflect.Repeated {
			_ = generateImport("", "stdStringsImport", g)
			g.P(`if !updated`, ccName, ` && strings.HasPrefix(f, prefix+"`, ccName, `") {`)
			g.P(`patchee.`, ccName, ` = patcher.`, ccName)
			g.P(`updated`, ccName, ` = true`)
			g.P(`continue`)
			g.P(`}`)
		} else {
			g.P(`if f == prefix+"`, ccName, `" {`)
			g.P(`patchee.`, ccName, ` = patcher.`, ccName)
			g.P(`continue`)
			g.P(`}`)
		}
	}
	g.P(`}`)
	g.P(`if err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)
	g.P(`return patchee, nil`)
	g.P(`}`)
	g.P()
}

func isSpecialType(typeName string) bool {
	parts := strings.Split(typeName, ".")
	if len(parts) > 2 { // what kinda format is this????
		panic(fmt.Sprintf(""))
	}
	if len(parts) == 1 { // native to this package = not special
		return false
	}
	// anything that looks like a google_protobufX should be considered special
	if strings.HasPrefix(strings.TrimLeft(typeName, "[]*"), "google_protobuf") {
		return true
	}
	switch parts[len(parts)-1] {
	case protoTypeJSON,
		protoTypeUUID,
		protoTypeUUIDValue,
		protoTypeResource,
		protoTypeInet,
		protoTimeOnly:
		return true
	}
	return false
}

func (b *ORMBuilder) generateListHandler(message *protogen.Message, g *protogen.GeneratedFile) {
	typeName := string(message.Desc.Name())
	ormable := b.getOrmable(typeName)

	g.P(`// DefaultList`, typeName, ` executes a gorm list call`)
	listSign := fmt.Sprint(`func DefaultList`, typeName, `(ctx context.Context, db *`, generateImport("DB", gormImport, g))
	var f, s, pg, fs string
	if b.listHasFiltering(ormable) {
		listSign += fmt.Sprint(`, f `, `*`, generateImport("Filtering", queryImport, g))
		f = "f"
	} else {
		f = "nil"
	}
	if b.listHasSorting(ormable) {
		listSign += fmt.Sprint(`, s `, `*`, generateImport("Sorting", queryImport, g))
		s = "s"
	} else {
		s = "nil"
	}
	if b.listHasPagination(ormable) {
		listSign += fmt.Sprint(`, p `, `*`, generateImport("Pagination", queryImport, g))
		pg = "p"
	} else {
		pg = "nil"
	}
	if b.listHasFieldSelection(ormable) {
		listSign += fmt.Sprint(`, fs `, `*`, generateImport("FieldSelection", queryImport, g))
		fs = "fs"
	} else {
		fs = "nil"
	}
	listSign += fmt.Sprint(`) ([]*`, typeName, `, error) {`)
	g.P(listSign)
	g.P(`in := `, typeName, `{}`)
	g.P(`ormObj, err := in.ToORM(ctx)`)
	g.P(`if err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)
	b.generateBeforeListHookCall(ormable, "ApplyQuery", g)
	g.P(`db, err = `, generateImport("ApplyCollectionOperators", tkgormImport, g), `(ctx, db, &`, ormable.Name, `{}, &`, typeName, `{}, `, f, `,`, s, `,`, pg, `,`, fs, `)`)
	g.P(`if err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)
	b.generateBeforeListHookCall(ormable, "Find", g)
	g.P(`db = db.Where(&ormObj)`)

	// add default ordering by primary key
	if b.hasPrimaryKey(ormable) {
		pkName, pk := b.findPrimaryKey(ormable)
		column := pk.GetTag().GetColumn()
		if len(column) == 0 {
			column = jgorm.ToDBName(pkName)
		}
		g.P(`db = db.Order("`, column, `")`)
	}

	g.P(`ormResponse := []`, ormable.Name, `{}`)
	g.P(`if err := db.Find(&ormResponse).Error; err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)
	b.generateAfterListHookCall(ormable, g)
	g.P(`pbResponse := []*`, typeName, `{}`)
	g.P(`for _, responseEntry := range ormResponse {`)
	g.P(`temp, err := responseEntry.ToPB(ctx)`)
	g.P(`if err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)
	g.P(`pbResponse = append(pbResponse, &temp)`)
	g.P(`}`)
	g.P(`return pbResponse, nil`)
	g.P(`}`)
	b.generateBeforeListHookDef(ormable, "ApplyQuery", g)
	b.generateBeforeListHookDef(ormable, "Find", g)
	b.generateAfterListHookDef(ormable, g)
}

func (b *ORMBuilder) generateBeforeListHookCall(orm *OrmableType, suffix string, g *protogen.GeneratedFile) {
	g.P(`if hook, ok := interface{}(&ormObj).(`, orm.Name, `WithBeforeList`, suffix, `); ok {`)
	hookCall := fmt.Sprint(`if db, err = hook.BeforeList`, suffix, `(ctx, db`)
	if b.listHasFiltering(orm) {
		hookCall += `,f`
	}
	if b.listHasSorting(orm) {
		hookCall += `,s`
	}
	if b.listHasPagination(orm) {
		hookCall += `,p`
	}
	if b.listHasFieldSelection(orm) {
		hookCall += `,fs`
	}
	hookCall += `); err != nil {`
	g.P(hookCall)
	g.P(`return nil, err`)
	g.P(`}`)
	g.P(`}`)
}

func (b *ORMBuilder) generateAfterListHookCall(orm *OrmableType, g *protogen.GeneratedFile) {
	g.P(`if hook, ok := interface{}(&ormObj).(`, orm.Name, `WithAfterListFind); ok {`)
	hookCall := fmt.Sprint(`if err = hook.AfterListFind(ctx, db, &ormResponse`)
	if b.listHasFiltering(orm) {
		hookCall += `,f`
	}
	if b.listHasSorting(orm) {
		hookCall += `,s`
	}
	if b.listHasPagination(orm) {
		hookCall += `,p`
	}
	if b.listHasFieldSelection(orm) {
		hookCall += `,fs`
	}
	hookCall += `); err != nil {`
	g.P(hookCall)
	g.P(`return nil, err`)
	g.P(`}`)
	g.P(`}`)
}

func (p *ORMBuilder) generateBeforeListHookDef(orm *OrmableType, suffix string, g *protogen.GeneratedFile) {
	g.P(`type `, orm.Name, `WithBeforeList`, suffix, ` interface {`)
	hookSign := fmt.Sprint(`BeforeList`, suffix, `(context.Context, *`, generateImport("DB", gormImport, g))
	if p.listHasFiltering(orm) {
		hookSign += fmt.Sprint(`, *`, generateImport("Filtering", queryImport, g))
	}
	if p.listHasSorting(orm) {
		hookSign += fmt.Sprint(`, *`, generateImport("Sorting", queryImport, g))
	}
	if p.listHasPagination(orm) {
		hookSign += fmt.Sprint(`, *`, generateImport("Pagination", queryImport, g))
	}
	if p.listHasFieldSelection(orm) {
		hookSign += fmt.Sprint(`, *`, generateImport("FieldSelection", queryImport, g))
	}
	hookSign += fmt.Sprint(`) (*`, generateImport("DB", gormImport, g), `, error)`)
	g.P(hookSign)
	g.P(`}`)
}

func (b *ORMBuilder) generateAfterListHookDef(orm *OrmableType, g *protogen.GeneratedFile) {
	g.P(`type `, orm.Name, `WithAfterListFind interface {`)
	hookSign := fmt.Sprint(`AfterListFind(context.Context, *`, generateImport("DB", gormImport, g), `, *[]`, orm.Name)
	if b.listHasFiltering(orm) {
		hookSign += fmt.Sprint(`, *`, generateImport("Filtering", queryImport, g))
	}
	if b.listHasSorting(orm) {
		hookSign += fmt.Sprint(`, *`, generateImport("Sorting", queryImport, g))
	}
	if b.listHasPagination(orm) {
		hookSign += fmt.Sprint(`, *`, generateImport("Pagination", queryImport, g))
	}
	if b.listHasFieldSelection(orm) {
		hookSign += fmt.Sprint(`, *`, generateImport("FieldSelection", queryImport, g))
	}
	hookSign += fmt.Sprint(`) error`)
	g.P(hookSign)
	g.P(`}`)
}

func generateImport(name string, importPath string, g *protogen.GeneratedFile) string {
	return g.QualifiedGoIdent(protogen.GoIdent{
		GoName:       name,
		GoImportPath: protogen.GoImportPath(importPath),
	})
}

func (b *ORMBuilder) listHasFiltering(ormable *OrmableType) bool {
	// TODO: parse filtering
	//if read, ok := ormable.Methods[listService]; ok {
	//	if s := p.getFiltering(read.inType); s != "" {
	//		return true
	//	}
	//}
	return false
}

func (b *ORMBuilder) listHasSorting(ormable *OrmableType) bool {
	// TODO: parse sorting
	//if read, ok := ormable.Methods[listService]; ok {
	//	if s := p.getSorting(read.inType); s != "" {
	//		return true
	//	}
	//}
	return false
}

func (b *ORMBuilder) listHasPagination(ormable *OrmableType) bool {
	// TODO: parse pagination
	//if read, ok := ormable.Methods[listService]; ok {
	//	if s := p.getPagination(read.inType); s != "" {
	//		return true
	//	}
	//}
	return false
}

func (b *ORMBuilder) listHasFieldSelection(ormable *OrmableType) bool {
	// TODO: parse selections
	//if read, ok := ormable.Methods[listService]; ok {
	//	if s := p.getFieldSelection(read.inType); s != "" {
	//		return true
	//	}
	//}
	return false
}

func (b *ORMBuilder) parseServices(file *protogen.File) {
	//defaultSuppressWarn := b.suppressWarn
	for _, service := range file.Services {
		genSvc := autogenService{
			Service: service,
			ccName:                 camelCase(string(service.Desc.Name())),
			file:                   file,
		}

		if opts := getServiceOptions(service); opts != nil {
			genSvc.autogen = opts.GetAutogen()
			genSvc.usesTxnMiddleware = opts.GetTxnMiddleware()
			//fmt.Fprintf(os.Stderr, "options: %+v\n", opts)
		}

		if !genSvc.autogen {
			b.suppressWarn = true
		}

		for _, method := range service.Methods {
			input := method.Input
			output := method.Output
			methodName := string(method.Desc.Name())
			var verb, fmName, baseType string
			var follows bool
			//fmt.Fprintf(os.Stderr, "method: %+v\n", method)

			if strings.HasPrefix(methodName, createService) {
				verb = createService
				follows, baseType = b.followsCreateConventions(input, output, createService)
			} else if strings.HasPrefix(methodName, readService) {
				verb = readService
				follows, baseType = b.followsReadConventions(input, output, readService)
			} else if strings.HasPrefix(methodName, updateSetService) {
				verb = updateSetService
				follows, baseType, fmName = b.followsUpdateSetConventions(input, output, updateSetService)
			} else if strings.HasPrefix(methodName, updateService) {
				verb = updateService
				follows, baseType, fmName = b.followsUpdateConventions(input, output, updateService)
			} else if strings.HasPrefix(methodName, deleteSetService) {
				verb = deleteSetService

			} else if strings.HasPrefix(methodName, deleteService) {
				verb = deleteService

			} else if strings.HasPrefix(methodName, listService) {
				verb = listService

			}

			genMethod := autogenMethod{
				Method: method,
				ccName: methodName,
				inType: input,
				outType: output,
				fieldMaskName: fmName,
				verb: verb,
				followsConvention: follows,
				baseType: baseType,
			}

			fmt.Fprintf(os.Stderr, "genMethod: %+v\n", genMethod)
		}

		//for _, method := range service.GetMethod() {
		//	inType, outType, methodName := p.getMethodProps(method)
		//	var verb, fmName, baseType string
		//	var follows bool
		//	if strings.HasPrefix(methodName, createService) {
		//		verb = createService
		//		follows, baseType = p.followsCreateConventions(inType, outType, createService)
		//	} else if strings.HasPrefix(methodName, readService) {
		//		verb = readService
		//		follows, baseType = p.followsReadConventions(inType, outType, readService)
		//	} else if strings.HasPrefix(methodName, updateSetService) {
		//		verb = updateSetService
		//		follows, baseType, fmName = p.followsUpdateSetConventions(inType, outType, updateSetService)
		//	} else if strings.HasPrefix(methodName, updateService) {
		//		verb = updateService
		//		follows, baseType, fmName = p.followsUpdateConventions(inType, outType, updateService)
		//	} else if strings.HasPrefix(methodName, deleteSetService) {
		//		verb = deleteSetService
		//		follows, baseType = p.followsDeleteSetConventions(inType, outType, method)
		//	} else if strings.HasPrefix(methodName, deleteService) {
		//		verb = deleteService
		//		follows, baseType = p.followsDeleteConventions(inType, outType, method)
		//	} else if strings.HasPrefix(methodName, listService) {
		//		verb = listService
		//		follows, baseType = p.followsListConventions(inType, outType, listService)
		//	}
		//	genMethod := autogenMethod{
		//		MethodDescriptorProto: method,
		//		ccName:                methodName,
		//		inType:                inType,
		//		outType:               outType,
		//		baseType:              baseType,
		//		fieldMaskName:         fmName,
		//		followsConvention:     follows,
		//		verb:                  verb,
		//	}
		//	genSvc.methods = append(genSvc.methods, genMethod)
		//
		//	if genMethod.verb != "" && p.isOrmable(genMethod.baseType) {
		//		p.getOrmable(genMethod.baseType).Methods[genMethod.verb] = &genMethod
		//	}
		//}
		//p.ormableServices = append(p.ormableServices, genSvc)
		//p.suppressWarn = defaultSuppressWarn
	}

}

func (b *ORMBuilder) followsCreateConventions(inType *protogen.Message, outType *protogen.Message, methodName string) (bool, string) {
	var inTypeName string
	var typeOrmable bool
	for _, field := range inType.Fields {
		if field.GoName == "payload" {
			gType := field.Desc.Kind().String()
			inTypeName = strings.TrimPrefix(gType, "*")
			if b.isOrmable(inTypeName) {
				typeOrmable = true
			}
		}
	}
	if !typeOrmable {
		// TODO: how to produce waringing
		//p.warning(`stub will be generated for %s since %s incoming message doesn't have "payload" field of ormable type`, methodName, p.TypeName(inType))
		return false, ""
	}
	var outTypeName string
	for _, field := range outType.Fields {
		if field.GoName == "result" {
			gType := field.Desc.Kind().String()
			outTypeName = strings.TrimPrefix(gType, "*")
		}
	}
	fmt.Fprintf(os.Stderr, "conv: %s -- %s\n", inTypeName, outTypeName)
	if inTypeName != outTypeName {
		//p.warning(`stub will be generated for %s since "payload" field type of %s incoming message type doesn't match "result" field type of %s outcoming message`, methodName, p.TypeName(inType), p.TypeName(outType))
		return false, ""
	}
	return true, inTypeName
}

func (b *ORMBuilder) followsReadConventions(inType *protogen.Message, outType *protogen.Message, methodName string) (bool, string) {
	var hasID bool
	for _, field := range inType.Fields {
		if string(field.Desc.Name()) == "id" {
			hasID = true
		}
	}
	if !hasID {
		//p.warning(`stub will be generated for %s since %s incoming message doesn't have "id" field`, methodName, p.TypeName(inType))
		return false, ""
	}
	var outTypeName string
	var typeOrmable bool
	for _, field := range outType.Fields {
		if string(field.Desc.Name()) == "result" {
			gType := field.Desc.Kind().String()
			outTypeName = strings.TrimPrefix(gType, "*")
			if b.isOrmable(outTypeName) {
				typeOrmable = true
			}
		}
	}
	if !typeOrmable {
		//p.warning(`stub will be generated for %s since %s outcoming message doesn't have "result" field of ormable type`, methodName, p.TypeName(outType))
		return false, ""
	}
	if !b.hasPrimaryKey(b.getOrmable(outTypeName)) {
		//p.warning(`stub will be generated for %s since %s ormable type doesn't have a primary key`, methodName, outTypeName)
		return false, ""
	}

	return true, outTypeName
}

func (b *ORMBuilder) followsUpdateSetConventions(inType *protogen.Message, outType *protogen.Message, methodName string) (bool, string, string) {
	var (
		inEntity    *protogen.Field
		inFieldMask *protogen.Field
	)

	for _, f := range inType.Fields {
		if string(f.Desc.Name()) == "objects" {
			inEntity = f
		}

		if f.Desc.Kind().String() == ".google.protobuf.FieldMask" {
			if inFieldMask != nil {
				//p.warning("message must not contains double field mask, prev on field name %s, after on field %s", inFieldMask.GetName(), f.GetName())
				return false, "", ""
			}

			inFieldMask = f
		}
	}

	var outEntity *protogen.Field
	for _, f := range outType.Fields {
		if string(f.Desc.Name()) == "results" {
			outEntity = f
		}
	}

	if inFieldMask == nil || inFieldMask.Desc.Cardinality() != protoreflect.Repeated {
		//p.warning("repeated field mask should exist in request for method %q", methodName)
		return false, "", ""
	}

	if inEntity == nil || outEntity == nil {
		//p.warning(`method: %q, request should has repeated field 'objects' in request and repeated field 'results' in response`, methodName)
		return false, "", ""
	}

	if inEntity.Desc.Cardinality() != protoreflect.Repeated || outEntity.Desc.Cardinality() != protoreflect.Repeated {
		//p.warning(`method: %q, field 'objects' in request and field 'results' in response should be repeated`, methodName)
		return false, "", ""
	}

	inGoType := inEntity.Desc.Kind().String() // TODO: not sure
	outGoType := outEntity.Desc.Kind().String()
	inTypeName, outTypeName := strings.TrimPrefix(inGoType, "*"), strings.TrimPrefix(outGoType, "*")
	if !b.isOrmable(inTypeName) {
		//p.warning("method: %q, type %q must be ormable", methodName, inTypeName)
		return false, "", ""
	}

	if inTypeName != outTypeName {
		//p.warning("method: %q, field 'objects' in request has type: %q but field 'results' in response has: %q", methodName, inTypeName, outTypeName)
		return false, "", ""
	}

	return true, inTypeName, camelCase(inFieldMask.GoName)
}

func (b *ORMBuilder) followsUpdateConventions(inType *protogen.Message, outType *protogen.Message, methodName string) (bool, string, string) {
	var inTypeName string
	var typeOrmable bool
	var updateMask string
	for _, field := range inType.Fields {
		if string(field.Desc.Name()) == "payload" {
			gType := field.Desc.Kind().String()
			inTypeName = strings.TrimPrefix(gType, "*")
			if b.isOrmable(inTypeName) {
				typeOrmable = true
			}
		}

		// Check that type of field is a FieldMask
		if field.Desc.Kind().String() == ".google.protobuf.FieldMask" {
			// More than one mask in request is not allowed.
			if updateMask != "" {
				return false, "", ""
			}
			updateMask = string(field.Desc.Name())
		}

	}
	if !typeOrmable {
		//p.warning(`stub will be generated for %s since %s incoming message doesn't have "payload" field of ormable type`, methodName, p.TypeName(inType))
		return false, "", ""
	}

	var outTypeName string
	for _, field := range outType.Fields {
		if string(field.Desc.Name()) == "result" {
			gType := field.Desc.Kind().String()
			outTypeName = strings.TrimPrefix(gType, "*")
		}
	}
	if inTypeName != outTypeName {
		//p.warning(`stub will be generated for %s since "payload" field type of %s incoming message doesn't match "result" field type of %s outcoming message`, methodName, p.TypeName(inType), p.TypeName(outType))
		return false, "", ""
	}
	if !b.hasPrimaryKey(b.getOrmable(inTypeName)) {
		//p.warning(`stub will be generated for %s since %s ormable type doesn't have a primary key`, methodName, outTypeName)
		return false, "", ""
	}

	return true, inTypeName, generator.CamelCase(updateMask)
}

func getServiceOptions(service *protogen.Service) *gorm.AutoServerOptions {
	options := service.Desc.Options().(*descriptorpb.ServiceOptions)
	if options == nil {
		return nil
	}

	v := proto.GetExtension(options, gorm.E_Server)
	if v == nil {
		return nil
	}

	opts, ok := v.(*gorm.AutoServerOptions)
	if !ok {
		return nil
	}

	return opts
}

func (b *ORMBuilder) generateDefaultServer(file *protogen.File, g *protogen.GeneratedFile) {
	//for _, service := range b.ormableServices {
	//	if service.file != file || !service.autogen {
	//		continue
	//	}
	//	p.P(`type `, service.ccName, `DefaultServer struct {`)
	//	if !service.usesTxnMiddleware {
	//		p.P(`DB *`, p.Import(gormImport), `.DB`)
	//	}
	//	p.P(`}`)
	//	withSpan := getServiceOptions(service.ServiceDescriptorProto).WithTracing
	//	if withSpan {
	//		p.generateSpanInstantiationMethod(service)
	//		p.generateSpanErrorMethod(service)
	//		p.generateSpanResultMethod(service)
	//	}
	//	for _, method := range service.methods {
	//		//Import context there because it have used in functions parameters
	//		p.UsingGoImports(stdCtxImport)
	//		switch method.verb {
	//		case createService:
	//			p.generateCreateServerMethod(service, method)
	//		case readService:
	//			p.generateReadServerMethod(service, method)
	//		case updateService:
	//			p.generateUpdateServerMethod(service, method)
	//		case updateSetService:
	//			p.generateUpdateSetServerMethod(service, method)
	//		case deleteService:
	//			p.generateDeleteServerMethod(service, method)
	//		case deleteSetService:
	//			p.generateDeleteSetServerMethod(service, method)
	//		case listService:
	//			p.generateListServerMethod(service, method)
	//		default:
	//			p.generateMethodStub(service, method)
	//		}
	//	}
	//}
}
