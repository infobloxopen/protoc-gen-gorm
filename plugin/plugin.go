package plugin

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	gormopts "github.com/infobloxopen/protoc-gen-gorm/options"
	"github.com/jinzhu/inflection"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
	gschema "gorm.io/gorm/schema"
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
	ErrNotOrmable = errors.New("type is not ormable")
)

var (
	gormImport         = "gorm.io/gorm"
	tkgormImport       = "github.com/infobloxopen/atlas-app-toolkit/gorm"
	uuidImport         = "github.com/satori/go.uuid"
	authImport         = "github.com/infobloxopen/protoc-gen-gorm/auth"
	gtypesImport       = "github.com/infobloxopen/protoc-gen-gorm/types"
	resourceImport     = "github.com/infobloxopen/atlas-app-toolkit/gorm/resource"
	queryImport        = "github.com/infobloxopen/atlas-app-toolkit/query"
	ocTraceImport      = "go.opencensus.io/trace"
	gatewayImport      = "github.com/infobloxopen/atlas-app-toolkit/gateway"
	pqImport           = "github.com/lib/pq"
	gerrorsImport      = "github.com/infobloxopen/protoc-gen-gorm/errors"
	timestampImport    = "google.golang.org/protobuf/types/known/timestamppb"
	durationImport     = "google.golang.org/protobuf/types/known/durationpb"
	wktImport          = "google.golang.org/protobuf/types/known/wrapperspb"
	fmImport           = "google.golang.org/genproto/protobuf/field_mask"
	stdFmtImport       = "fmt"
	stdCtxImport       = "context"
	stdStringsImport   = "strings"
	stdTimeImport      = "time"
	encodingJsonImport = "encoding/json"
	bigintImport       = "math/big"
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

var optionalTypes = map[string]string{
	"string":  "*string",
	"float64": "*float64",
	"float32": "*float32",
	"int32":   "*int32",
	"int64":   "*int64",
	"uint32":  "*uint32",
	"uint64":  "*uint64",
	"bool":    "*bool",
}

type pkFieldObjs struct {
	name  string
	field *Field
}

const (
	protoTypeTimestamp = "Timestamp" // last segment, first will be *google_protobufX
	protoTypeDuration  = "Duration"
	protoTypeJSON      = "JSONValue"
	protoTypeUUID      = "UUID"
	protoTypeUUIDValue = "UUIDValue"
	protoTypeResource  = "Identifier"
	protoTypeInet      = "InetValue"
	protoTimeOnly      = "TimeOnly"
	protoTypeBigInt    = "BigInt"
)

// DB Engine Enum
const (
	ENGINE_UNSET = iota
	ENGINE_POSTGRES
)

type ORMBuilder struct {
	plugin          *protogen.Plugin
	ormableTypes    map[string]*OrmableType
	messages        map[string]struct{}
	currentFile     string
	currentPackage  string
	ormableServices []autogenService
	dbEngine        int
	stringEnums     bool
	gateway         bool
	suppressWarn    bool
}

func New(opts protogen.Options, request *pluginpb.CodeGeneratorRequest) (*ORMBuilder, error) {
	plugin, err := opts.New(request)
	if err != nil {
		return nil, err
	}
	SetSupportedFeaturesOnPluginGen(plugin)

	builder := &ORMBuilder{
		plugin:       plugin,
		ormableTypes: make(map[string]*OrmableType),
		messages:     make(map[string]struct{}),
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
	Methods    []*autogenMethod
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
		Methods:    []*autogenMethod{},
	}
}

type Field struct {
	*gormopts.GormFieldOptions
	ParentGoType         string
	TypeName             string
	Type                 *OrmableType
	Package              string
	ParentOrigName       string
	FieldAssociationInfo fieldAssociationInfo
}

type autogenMethod struct {
	*protogen.Method
	outType           *protogen.Message
	inType            *protogen.Message
	verb              string
	baseType          string
	fieldMaskName     string
	ccName            string
	followsConvention bool
}

type fileImports struct {
	wktPkgName      string
	packages        map[string]*pkgImport
	typesToRegister []string
	stdImports      []string
}

type autogenService struct {
	*protogen.Service
	file              *protogen.File
	ccName            string
	methods           []autogenMethod
	usesTxnMiddleware bool
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

		b.currentPackage = protoFile.GoImportPath.String()

		// first traverse: preload the messages
		for _, message := range protoFile.Messages {
			if message.Desc.IsMapEntry() {
				continue
			}

			typeName := string(message.Desc.Name())
			b.messages[typeName] = struct{}{}

			if isOrmable(message) {
				ormable := NewOrmableType(typeName, string(protoFile.GoPackageName), protoFile)
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
		for _, message := range protoFile.Messages {
			typeName := string(message.Desc.Name())
			if isOrmable(message) {
				b.parseAssociations(message, g)
				o := b.getOrmable(typeName)
				if b.hasPrimaryKey(o) {
					_, fd := b.findPrimaryKey(o)
					fd.ParentOrigName = o.OriginName
				}
			}
		}

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

		if !protoFile.Generate {
			g.Skip()
			continue
		}

		skip := true

		for _, message := range protoFile.Messages {
			if isOrmable(message) {
				skip = false
				break
			}
		}

		for _, service := range protoFile.Services {
			if getServiceOptions(service) != nil {
				skip = false
				break
			}
		}

		if skip {
			g.Skip()
			continue
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

	SetSupportedFeaturesOnCodeGeneratorResponse(b.plugin.Response())
	return b.plugin.Response(), nil
}

func (b *ORMBuilder) generateConvertFunctions(g *protogen.GeneratedFile, message *protogen.Message) {
	typeName := string(message.Desc.Name())
	ormable := b.getOrmable(camelCase(typeName))

	// /// To Orm
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

	if getMessageOptions(message).GetMultiCompartment() {
		g.P("compartmentID, err := ", generateImport("GetCompartmentID", authImport, g), "(ctx, nil)")
		g.P("if err != nil {")
		g.P("return to, err")
		g.P("}")
		g.P("to.CompartmentID = compartmentID")
	}

	b.setupOrderedHasMany(message, g)
	g.P(`if posthook, ok := interface{}(m).(`, typeName, `WithAfterToORM); ok {`)
	g.P(`err = posthook.AfterToORM(ctx, &to)`)
	g.P(`}`)
	g.P(`return to, err`)
	g.P(`}`)

	g.P()
	// /// To Pb
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
		ofield := ormable.Fields[camelCase(field.GoName)]
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

	tableName := gschema.NamingStrategy{}.TableName(msgName)
	if opts := getMessageOptions(message); opts != nil && len(opts.Table) > 0 {
		tableName = opts.GetTable()
	}
	g.P(`return "`, tableName, `"`)
	g.P(`}`)
}

func (b *ORMBuilder) generateOrmable(g *protogen.GeneratedFile, message *protogen.Message) {
	ormable := b.getOrmable(message.GoIdent.GoName)
	g.P(`type `, ormable.Name, ` struct {`)

	var names []string
	for name := range ormable.Fields {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		field := ormable.Fields[name]
		sp := strings.Split(field.TypeName, ".")

		if len(sp) == 2 && sp[1] == "UUID" {
			s := generateImport("UUID", uuidImport, g)
			if field.TypeName[0] == '*' {
				field.TypeName = "*" + s
			} else {
				field.TypeName = s
			}
		}

		if len(sp) == 2 && sp[1] == "BigInt" {
			s := generateImport("BigInt", bigintImport, g)
			if field.TypeName[0] == '*' {
				field.TypeName = "*" + s
			} else {
				field.TypeName = s
			}
		}

		g.P(name, ` `, field.TypeName, b.renderGormTag(field))
	}

	g.P(`}`)
	g.P()
}

func (b *ORMBuilder) parseAssociations(msg *protogen.Message, g *protogen.GeneratedFile) {
	typeName := camelCase(string(msg.Desc.Name())) // TODO: camelSnakeCase
	ormable := b.getOrmable(typeName)

	for _, field := range msg.Fields {
		options := field.Desc.Options().(*descriptorpb.FieldOptions)
		fieldOpts := getFieldOptions(options)
		if fieldOpts.GetDrop() {
			continue
		}

		fieldName := camelCase(string(field.Desc.Name()))
		var fieldType string

		if field.Desc.Message() == nil {
			fieldType = field.Desc.Kind().String() // was GoType
		} else {
			fieldType = string(field.Desc.Message().Name())
		}
		fieldType = strings.Trim(fieldType, "[]*")
		parts := strings.Split(fieldType, ".")
		fieldTypeShort := parts[len(parts)-1]

		if b.isOrmable(fieldType) {
			if fieldOpts == nil {
				fieldOpts = &gormopts.GormFieldOptions{}
			}
			assocOrmable := b.getOrmable(fieldType)

			if field.Message != nil {
				fieldType = b.typeName(field.Message.GoIdent, g)
			}

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

			fInfo := parseGormAssosiationTags(fieldOpts)
			// Register type used, in case it's an imported type from another package
			// b.GetFileImports().typesToRegister = append(b.GetFileImports().typesToRegister, fieldType) // maybe we need other fields type
			ormable.Fields[fieldName] = &Field{TypeName: fieldType, GormFieldOptions: fieldOpts, Type: assocOrmable, FieldAssociationInfo: fInfo}
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

func (b *ORMBuilder) hasCompositePrimaryKey(ormable *OrmableType) bool {
	keys := 0
	for _, field := range ormable.Fields {
		if field.GetTag().GetPrimaryKey() && keys <= 1 {
			keys++
		}
	}
	if keys > 1 {
		return true
	}
	return false
}

func (b *ORMBuilder) isOrmable(typeName string) bool {
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

// getPrimaryKeys returns a sorted list of primary key field objects
func (b *ORMBuilder) getPrimaryKeys(ormable *OrmableType) []pkFieldObjs {
	var (
		fieldobjs []pkFieldObjs
	)

	for fieldName, field := range ormable.Fields {
		if field.GetTag().GetPrimaryKey() {
			fieldobjs = append(fieldobjs, pkFieldObjs{
				fieldName,
				field,
			})
		}
	}
	sort.Slice(fieldobjs, func(i, j int) bool {
		return fieldobjs[i].name < fieldobjs[j].name
	})
	// if no primary key is found, use the field named "id"
	if len(fieldobjs) == 0 {
		for fieldName, field := range ormable.Fields {
			if strings.ToLower(fieldName) == "id" {
				fieldobjs = append(fieldobjs, pkFieldObjs{
					fieldName,
					field,
				})
			}
		}
	}
	return fieldobjs
}

func (b *ORMBuilder) getOrmable(typeName string) *OrmableType {
	r, err := GetOrmable(b.ormableTypes, typeName)
	if err != nil {
		panic(err)
	}

	return r
}

func (b *ORMBuilder) parseManyToMany(msg *protogen.Message, ormable *OrmableType, fieldName string, fieldType string, assoc *OrmableType, opts *gormopts.GormFieldOptions) {
	typeName := camelCase(string(msg.Desc.Name()))
	mtm := opts.GetManyToMany()
	if mtm == nil {
		mtm = &gormopts.ManyToManyOptions{}
		opts.Association = &gormopts.GormFieldOptions_ManyToMany{mtm}
	}

	var foreignKeyName string
	if foreignKeyName = camelCase(mtm.GetForeignkey()); foreignKeyName == "" {
		foreignKeyName, _ = b.findPrimaryKey(ormable)
	} else {
		var ok bool
		_, ok = ormable.Fields[foreignKeyName]
		if !ok {
			panic(fmt.Sprintf("Missing %s field in %s", foreignKeyName, ormable.Name))
		}
	}
	mtm.Foreignkey = foreignKeyName
	var assocKeyName string
	if assocKeyName = camelCase(mtm.GetAssociationForeignkey()); assocKeyName == "" {
		assocKeyName, _ = b.findPrimaryKey(assoc)
	} else {
		var ok bool
		_, ok = assoc.Fields[assocKeyName]
		if !ok {
			panic(fmt.Sprintf("Missing %s field in %s", assocKeyName, assoc.Name))
		}
	}
	mtm.AssociationForeignkey = assocKeyName
	ns := gschema.NamingStrategy{SingularTable: true}
	var jt string
	if jt = ns.TableName(mtm.GetJointable()); jt == "" {
		if b.countManyToManyAssociationDimension(msg, fieldType) == 1 && typeName != fieldType {
			jt = ns.TableName(typeName + inflection.Plural(fieldType))
		} else {
			jt = ns.TableName(typeName + inflection.Plural(fieldName))
		}
	}
	mtm.Jointable = jt
	var jtForeignKey string
	if jtForeignKey = camelCase(mtm.GetJointableForeignkey()); jtForeignKey == "" {
		jtForeignKey = camelCase(ns.TableName(typeName + foreignKeyName))
	}
	mtm.JointableForeignkey = jtForeignKey
	var jtAssocForeignKey string
	if jtAssocForeignKey = camelCase(mtm.GetAssociationJointableForeignkey()); jtAssocForeignKey == "" {
		if typeName == fieldType {
			jtAssocForeignKey = ns.TableName(inflection.Singular(fieldName) + assocKeyName)
		} else {
			jtAssocForeignKey = ns.TableName(fieldType + assocKeyName)
		}
	}
	mtm.AssociationJointableForeignkey = camelCase(jtAssocForeignKey)
}

func (b *ORMBuilder) parseHasOne(msg *protogen.Message, parent *OrmableType, fieldName string, fieldType string, child *OrmableType, opts *gormopts.GormFieldOptions) {
	typeName := camelCase(string(msg.Desc.Name()))
	hasOne := opts.GetHasOne()
	if hasOne == nil {
		hasOne = &gormopts.HasOneOptions{}
		opts.Association = &gormopts.GormFieldOptions_HasOne{hasOne}
	}

	var assocKey *Field
	var assocKeyName string

	if assocKeyName = camelCase(hasOne.GetAssociationForeignkey()); assocKeyName == "" {
		assocKeyName, assocKey = b.findPrimaryKey(parent)
	} else {
		var ok bool
		assocKey, ok = parent.Fields[assocKeyName]
		if !ok {
			panic(fmt.Sprintf("Missing %s field in %s.", assocKeyName, parent.Name))
		}
	}

	hasOne.AssociationForeignkey = assocKeyName
	var foreignKeyType string
	if hasOne.GetForeignkeyTag().GetNotNull() {
		foreignKeyType = strings.TrimPrefix(assocKey.TypeName, "*")
	} else if strings.HasPrefix(assocKey.TypeName, "*") {
		foreignKeyType = assocKey.TypeName
	} else if strings.Contains(assocKey.TypeName, "[]byte") {
		foreignKeyType = assocKey.TypeName
	} else {
		foreignKeyType = "*" + assocKey.TypeName
	}

	foreignKey := &Field{TypeName: foreignKeyType, Package: assocKey.Package, GormFieldOptions: &gormopts.GormFieldOptions{Tag: hasOne.GetForeignkeyTag()}}
	var foreignKeyName string
	if foreignKeyName = camelCase(hasOne.GetForeignkey()); foreignKeyName == "" {
		if b.countHasAssociationDimension(msg, fieldType) == 1 {
			foreignKeyName = fmt.Sprintf(typeName + assocKeyName)
		} else {
			foreignKeyName = fmt.Sprintf(fieldName + typeName + assocKeyName)
		}
	}

	hasOne.Foreignkey = foreignKeyName
	if _, ok := child.Fields[foreignKeyName]; child.Package != parent.Package && !ok {
		panic(fmt.Sprintf("Object %s from package %s cannot be user for has-one in %s since it does not have FK field %s defined. Manually define the key, or switch to belongs-to.",
			child.Name, child.Package, parent.Name, foreignKeyName))
	}
	if exField, ok := child.Fields[foreignKeyName]; !ok {
		child.Fields[foreignKeyName] = foreignKey
	} else {
		if exField.TypeName == "interface{}" {
			exField.TypeName = foreignKey.TypeName
		} else if !b.sameType(exField, foreignKey) {
			panic(fmt.Sprintf("Cannot include %s field into %s as it already exists there with a different type: %s, %s",
				foreignKeyName, child.Name, exField.TypeName, foreignKey.TypeName))
		}
	}

	child.Fields[foreignKeyName].ParentOrigName = parent.OriginName
}

func (b *ORMBuilder) parseHasMany(msg *protogen.Message, parent *OrmableType, fieldName string, fieldType string, child *OrmableType, opts *gormopts.GormFieldOptions) {
	typeName := camelCase(string(msg.Desc.Name()))
	hasMany := opts.GetHasMany()
	if hasMany == nil {
		hasMany = &gormopts.HasManyOptions{}
		opts.Association = &gormopts.GormFieldOptions_HasMany{hasMany}
	}
	var assocKey *Field
	var assocKeyName string
	if assocKeyName = camelCase(hasMany.GetAssociationForeignkey()); assocKeyName == "" {
		assocKeyName, assocKey = b.findPrimaryKey(parent)
	} else {
		var ok bool
		assocKey, ok = parent.Fields[assocKeyName]
		if !ok {
			panic(fmt.Sprintf("Missing %s field in %s", assocKeyName, parent.Name))
		}
	}

	hasMany.AssociationForeignkey = assocKeyName
	var foreignKeyType string
	if hasMany.GetForeignkeyTag().GetNotNull() {
		foreignKeyType = strings.TrimPrefix(assocKey.TypeName, "*")
	} else if strings.HasPrefix(assocKey.TypeName, "*") {
		foreignKeyType = assocKey.TypeName
	} else if strings.Contains(assocKey.TypeName, "[]byte") {
		foreignKeyType = assocKey.TypeName
	} else {
		foreignKeyType = "*" + assocKey.TypeName
	}
	foreignKey := &Field{TypeName: foreignKeyType, Package: assocKey.Package, GormFieldOptions: &gormopts.GormFieldOptions{Tag: hasMany.GetForeignkeyTag()}}
	var foreignKeyName string
	if foreignKeyName = hasMany.GetForeignkey(); foreignKeyName == "" {
		if b.countHasAssociationDimension(msg, fieldType) == 1 {
			foreignKeyName = fmt.Sprintf(typeName + assocKeyName)
		} else {
			foreignKeyName = fmt.Sprintf(fieldName + typeName + assocKeyName)
		}
	}
	hasMany.Foreignkey = foreignKeyName
	if _, ok := child.Fields[foreignKeyName]; child.Package != parent.Package && !ok {
		panic(fmt.Sprintf("Object %s from package %s cannot be user for has-many in %s since it does not have FK field %s defined. Manually define the key, or switch to many-to-many.",
			child.Name, child.Package, parent.Name, foreignKeyName))

	}
	if exField, ok := child.Fields[foreignKeyName]; !ok {
		child.Fields[foreignKeyName] = foreignKey
	} else {
		if exField.TypeName == "interface{}" {
			exField.TypeName = foreignKey.TypeName
		} else if !b.sameType(exField, foreignKey) {
			panic(fmt.Sprintf("Cannot include %s field into %s as it already exists there with a different type: %s, %s",
				foreignKeyName, child.Name, exField.TypeName, foreignKey.TypeName))
		}
	}
	child.Fields[foreignKeyName].ParentOrigName = parent.OriginName

	var posField string
	if posField = camelCase(hasMany.GetPositionField()); posField != "" {
		if exField, ok := child.Fields[posField]; !ok {
			child.Fields[posField] = &Field{TypeName: "int", GormFieldOptions: &gormopts.GormFieldOptions{Tag: hasMany.GetPositionFieldTag()}}
		} else {
			if !strings.Contains(exField.TypeName, "int") {
				panic(fmt.Sprintf("Cannot include %s field into %s as it already exists there with a different type.",
					posField, child.Name))
			}
		}
		hasMany.PositionField = posField
	}
}

func (b *ORMBuilder) parseBelongsTo(msg *protogen.Message, child *OrmableType, fieldName string, fieldType string, parent *OrmableType, opts *gormopts.GormFieldOptions) {
	belongsTo := opts.GetBelongsTo()
	if belongsTo == nil {
		belongsTo = &gormopts.BelongsToOptions{}
		opts.Association = &gormopts.GormFieldOptions_BelongsTo{belongsTo}
	}
	var assocKey *Field
	var assocKeyName string
	if assocKeyName = camelCase(belongsTo.GetAssociationForeignkey()); assocKeyName == "" {
		assocKeyName, assocKey = b.findPrimaryKey(parent)
	} else {
		var ok bool
		assocKey, ok = parent.Fields[assocKeyName]
		if !ok {
			panic(fmt.Sprintf("Missing %s field in %s", assocKeyName, parent.Name))
		}
	}
	belongsTo.AssociationForeignkey = assocKeyName
	var foreignKeyType string
	if belongsTo.GetForeignkeyTag().GetNotNull() {
		foreignKeyType = strings.TrimPrefix(assocKey.TypeName, "*")
	} else if strings.HasPrefix(assocKey.TypeName, "*") {
		foreignKeyType = assocKey.TypeName
	} else if strings.Contains(assocKey.TypeName, "[]byte") {
		foreignKeyType = assocKey.TypeName
	} else {
		foreignKeyType = "*" + assocKey.TypeName
	}
	foreignKey := &Field{TypeName: foreignKeyType, Package: assocKey.Package, GormFieldOptions: &gormopts.GormFieldOptions{Tag: belongsTo.GetForeignkeyTag()}}
	var foreignKeyName string
	if foreignKeyName = camelCase(belongsTo.GetForeignkey()); foreignKeyName == "" {
		if b.countBelongsToAssociationDimension(msg, fieldType) == 1 {
			foreignKeyName = fmt.Sprintf(fieldType + assocKeyName)
		} else {
			foreignKeyName = fmt.Sprintf(fieldName + assocKeyName)
		}
	}
	belongsTo.Foreignkey = foreignKeyName
	if exField, ok := child.Fields[foreignKeyName]; !ok {
		child.Fields[foreignKeyName] = foreignKey
	} else {
		if exField.TypeName == "interface{}" {
			exField.TypeName = foreignKeyType
		} else if !b.sameType(exField, foreignKey) {
			panic(fmt.Sprintf("Cannot include %s field into %s as it already exists there with a different type: %s, %s",
				foreignKeyName, child.Name, exField.TypeName, foreignKey.TypeName))
		}
	}
	child.Fields[foreignKeyName].ParentOrigName = parent.OriginName
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
			gormOptions = &gormopts.GormFieldOptions{}
		}
		if gormOptions.GetDrop() {
			continue
		}

		tag := gormOptions.Tag
		fieldName := camelCase(string(fd.Name()))
		fieldType := fd.Kind().String()
		var typePackage string

		if b.dbEngine == ENGINE_POSTGRES && b.IsAbleToMakePQArray(fieldType) && field.Desc.IsList() {
			switch fieldType {
			case "bool":
				fieldType = generateImport("BoolArray", pqImport, g)
				gormOptions.Tag = tagWithType(tag, "bool[]")
			case "double":
				fieldType = generateImport("Float64Array", pqImport, g)
				gormOptions.Tag = tagWithType(tag, "float[]")
			case "int64":
				fieldType = generateImport("Int64Array", pqImport, g)
				gormOptions.Tag = tagWithType(tag, "integer[]")
			case "string":
				fieldType = generateImport("StringArray", pqImport, g)
				gormOptions.Tag = tagWithType(tag, "text[]")
			default:
				continue
			}
		} else if (field.Message == nil || !b.isOrmable(fieldType)) && field.Desc.IsList() {
			// not implemented
			continue
		} else if field.Enum != nil {
			fieldType = "int32"
			if b.stringEnums {
				fieldType = "string"
			}
		} else if field.Message != nil {
			xs := strings.Split(string(field.Message.Desc.FullName()), ".")
			rawType := xs[len(xs)-1]

			if v, ok := wellKnownTypes[rawType]; ok {
				fieldType = v
			} else if rawType == protoTypeBigInt {
				typePackage = bigintImport
				fieldType = "*" + generateImport("Int", bigintImport, g)
				if b.dbEngine == ENGINE_POSTGRES {
					gormOptions.Tag = tagWithType(tag, "numeric")
				}
			} else if rawType == protoTypeUUID {
				typePackage = uuidImport
				fieldType = generateImport("UUID", uuidImport, g)
				if b.dbEngine == ENGINE_POSTGRES {
					gormOptions.Tag = tagWithType(tag, "uuid")
				}
			} else if rawType == protoTypeUUIDValue {
				typePackage = uuidImport
				fieldType = "*" + generateImport("UUID", uuidImport, g)
				if b.dbEngine == ENGINE_POSTGRES {
					gormOptions.Tag = tagWithType(tag, "uuid")
				}
			} else if rawType == protoTypeTimestamp {
				typePackage = stdTimeImport
				fieldType = "*" + generateImport("Time", stdTimeImport, g)
			} else if rawType == protoTypeDuration {
				typePackage = stdTimeImport
				fieldType = "*" + generateImport("Duration", stdTimeImport, g)
			} else if rawType == protoTypeJSON {
				if b.dbEngine == ENGINE_POSTGRES {
					typePackage = gtypesImport
					fieldType = "*" + generateImport("Jsonb", gtypesImport, g)
					gormOptions.Tag = tagWithType(tag, "jsonb")
				} else {
					// Potential TODO: add types we want to use in other/default DB engine
					continue
				}
			} else if rawType == protoTypeResource {
				ttype := strings.ToLower(tag.GetType())
				if strings.Contains(ttype, "char") {
					ttype = "char"
				}
				if field.Desc.IsList() {
					ttype = "array"
				}
				switch ttype {
				case "uuid", "text", "char", "array", "cidr", "inet", "macaddr":
					fieldType = "*string"
				case "smallint", "integer", "bigint", "numeric", "smallserial", "serial", "bigserial":
					fieldType = "*int64"
				case "jsonb", "bytea":
					fieldType = "[]byte"
				case "":
					fieldType = "interface{}" // we do not know the type yet (if it association we will fix the type later)
				default:
					panic("unknown tag type of atlas.rpc.Identifier")
				}
				if tag.GetNotNull() || tag.GetPrimaryKey() {
					fieldType = strings.TrimPrefix(fieldType, "*")
				}
			} else if rawType == protoTypeInet {
				typePackage = gtypesImport
				fieldType = "*" + generateImport("Inet", gtypesImport, g)

				if b.dbEngine == ENGINE_POSTGRES {
					gormOptions.Tag = tagWithType(tag, "inet")
				} else {
					gormOptions.Tag = tagWithType(tag, "varchar(48)")
				}
			} else if rawType == protoTimeOnly {
				fieldType = "string"
				gormOptions.Tag = tagWithType(tag, "time")
			} else {
				continue
			}
		}

		switch fieldType {
		case "bytes":
			fieldType = "[]byte"
			gormOptions.Tag = tagWithType(tag, "bytea")
		case "float":
			fieldType = "float32"
		case "double":
			fieldType = "float64"
		}

		// handle optional fields
		if fd.HasOptionalKeyword() {
			if v, ok := optionalTypes[fieldType]; ok {
				fieldType = v
			}
		}

		f := &Field{
			GormFieldOptions: gormOptions,
			ParentGoType:     "",
			TypeName:         fieldType,
			Package:          typePackage,
		}

		if tName := gormOptions.GetReferenceOf(); tName != "" {
			if _, ok := b.messages[tName]; !ok {
				panic("unknow")
			}
			f.ParentOrigName = tName
		}

		ormable.Fields[fieldName] = f
	}

	gormMsgOptions := getMessageOptions(msg)
	if gormMsgOptions.GetMultiAccount() {
		if accID, ok := ormable.Fields["AccountID"]; !ok {
			ormable.Fields["AccountID"] = &Field{TypeName: "string"}
		} else if accID.TypeName != "string" {
			panic("cannot include AccountID field")
		}
	}

	if gormMsgOptions.GetMultiCompartment() {
		if comID, ok := ormable.Fields["CompartmentID"]; !ok {
			ormable.Fields["CompartmentID"] = &Field{TypeName: "string"}
		} else if comID.TypeName != "string" {
			panic("cannot include CompartmentID field")
		}
	}

	// TODO: GetInclude
	for _, field := range gormMsgOptions.GetInclude() {
		fieldName := camelCase(field.GetName())
		if _, ok := ormable.Fields[fieldName]; !ok {
			b.addIncludedField(ormable, field, g)
		} else {
			panic("cound not include")
		}
	}
}

func (b *ORMBuilder) addIncludedField(ormable *OrmableType, field *gormopts.ExtraField, g *protogen.GeneratedFile) {
	fieldName := camelCase(field.GetName())
	isPtr := strings.HasPrefix(field.GetType(), "*")
	rawType := strings.TrimPrefix(field.GetType(), "*")
	// cut off any package subpaths
	rawType = rawType[strings.LastIndex(rawType, ".")+1:]
	var typePackage string
	// Handle types with a package defined
	if field.GetPackage() != "" {
		rawType = generateImport(rawType, field.GetPackage(), g)
		typePackage = field.GetPackage()
	} else {
		// Handle types without a package defined
		if _, ok := builtinTypes[rawType]; ok {
			// basic type, 100% okay, no imports or changes needed
		} else if rawType == "Time" {
			// b.UsingGoImports(stdTimeImport) // TODO: missing UsingGoImports
			rawType = generateImport("Time", stdTimeImport, g)
		} else if rawType == "BigInt" {
			rawType = generateImport("Int", bigintImport, g)
		} else if rawType == "UUID" {
			rawType = generateImport("UUID", uuidImport, g)
		} else if field.GetType() == "Jsonb" {
			rawType = generateImport("Jsonb", gtypesImport, g)
		} else if rawType == "Inet" {
			rawType = generateImport("Inet", gtypesImport, g)
		} else {
			fmt.Fprintf(os.Stderr, "included field %q of type %q is not a recognized special type, and no package specified. This type is assumed to be in the same package as the generated code",
				field.GetName(), field.GetType())
		}
	}
	if isPtr {
		rawType = fmt.Sprintf("*%s", rawType)
	}
	ormable.Fields[fieldName] = &Field{TypeName: rawType, Package: typePackage, GormFieldOptions: &gormopts.GormFieldOptions{Tag: field.GetTag()}}
}

func getFieldOptions(options *descriptorpb.FieldOptions) *gormopts.GormFieldOptions {
	if options == nil {
		return nil
	}

	v := proto.GetExtension(options, gormopts.E_Field)
	if v == nil {
		return nil
	}

	opts, ok := v.(*gormopts.GormFieldOptions)
	if !ok {
		return nil
	}

	return opts
}

// retrieves the GormMessageOptions from a message
func getMessageOptions(message *protogen.Message) *gormopts.GormMessageOptions {
	options := message.Desc.Options()
	if options == nil {
		return nil
	}
	v := proto.GetExtension(options, gormopts.E_Opts)
	if v == nil {
		return nil
	}

	opts, ok := v.(*gormopts.GormMessageOptions)
	if !ok {
		return nil
	}

	return opts
}

func isOrmable(message *protogen.Message) bool {
	desc := message.Desc
	options := desc.Options()

	m, ok := proto.GetExtension(options, gormopts.E_Opts).(*gormopts.GormMessageOptions)
	if !ok || m == nil {
		return false
	}

	return m.Ormable
}

func (b *ORMBuilder) IsAbleToMakePQArray(fieldType string) bool {
	switch fieldType {
	case "bool", "double", "int64", "string":
		return true
	default:
		return false
	}
}

func tagWithType(tag *gormopts.GormTag, typename string) *gormopts.GormTag {
	if tag == nil {
		tag = &gormopts.GormTag{}
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

type fieldAssociationInfo interface {
	GetDisableAssociationAutocreate() bool
	GetDisableAssociationAutoupdate() bool
	GetPreload() bool
}

func parseGormAssosiationTags(field *gormopts.GormFieldOptions) fieldAssociationInfo {
	switch {
	case field.GetHasOne() != nil:
		return field.GetHasOne()
	case field.GetBelongsTo() != nil:
		return field.GetBelongsTo()
	case field.GetHasMany() != nil:
		return field.GetHasMany()
	case field.GetManyToMany() != nil:
		return field.GetManyToMany()
	default:
		return field.GetTag()
	}
}

func (b *ORMBuilder) renderGormTag(field *Field) string {
	var gormRes, atlasRes string
	tag := field.GetTag()
	if tag == nil {
		tag = &gormopts.GormTag{}
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
		gormRes += "primaryKey;"
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
		gormRes += "autoIncrement;"
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
			gormRes += "uniqueIndex;"
		} else {
			gormRes += fmt.Sprintf("uniqueIndex:%s;", tag.GetUniqueIndex())
		}
	}
	if tag.GetEmbedded() {
		gormRes += "embedded;"
	}
	if len(tag.EmbeddedPrefix) > 0 {
		gormRes += fmt.Sprintf("embeddedPrefix:%s;", tag.GetEmbeddedPrefix())
	}
	if tag.GetIgnore() {
		gormRes += "-;"
	}

	if len(tag.GetSerializer()) > 0 {
		gormRes += fmt.Sprintf("serializer:%s;", tag.GetSerializer())
	}

	var foreignKey, associationForeignKey, joinTable, joinTableForeignKey, associationJoinTableForeignKey string
	var replace, append, clear bool
	if hasOne := field.GetHasOne(); hasOne != nil {
		foreignKey = hasOne.Foreignkey
		associationForeignKey = hasOne.AssociationForeignkey
		clear = hasOne.Clear
		replace = hasOne.Replace
		append = hasOne.Append
	} else if belongsTo := field.GetBelongsTo(); belongsTo != nil {
		foreignKey = belongsTo.Foreignkey
		associationForeignKey = belongsTo.AssociationForeignkey
	} else if hasMany := field.GetHasMany(); hasMany != nil {
		foreignKey = hasMany.Foreignkey
		associationForeignKey = hasMany.AssociationForeignkey
		clear = hasMany.Clear
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
		clear = mtm.Clear
		replace = mtm.Replace
		append = mtm.Append
	} else {
		foreignKey = tag.Foreignkey
		associationForeignKey = tag.AssociationForeignkey
		joinTable = tag.ManyToMany
		joinTableForeignKey = tag.JointableForeignkey
		associationJoinTableForeignKey = tag.AssociationJointableForeignkey
	}

	if len(foreignKey) > 0 {
		gormRes += fmt.Sprintf("foreignKey:%s;", foreignKey)
	}

	if len(associationForeignKey) > 0 {
		gormRes += fmt.Sprintf("references:%s;", associationForeignKey)
	}

	if len(joinTable) > 0 {
		gormRes += fmt.Sprintf("many2many:%s;", joinTable)
	}
	if len(joinTableForeignKey) > 0 {
		gormRes += fmt.Sprintf("joinForeignKey:%s;", joinTableForeignKey)
	}
	if len(associationJoinTableForeignKey) > 0 {
		gormRes += fmt.Sprintf("joinReferences:%s;", associationJoinTableForeignKey)
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

func (b *ORMBuilder) setupOrderedHasMany(message *protogen.Message, g *protogen.GeneratedFile) {
	typeName := string(message.Desc.Name())
	ormable := b.getOrmable(typeName)
	var fieldNames []string
	for name := range ormable.Fields {
		fieldNames = append(fieldNames, name)
	}
	sort.Strings(fieldNames)

	for _, fieldName := range fieldNames {
		b.setupOrderedHasManyByName(message, fieldName, g)
	}
}

func (b *ORMBuilder) setupOrderedHasManyByName(message *protogen.Message, fieldName string, g *protogen.GeneratedFile) {
	typeName := string(message.Desc.Name())
	ormable := b.getOrmable(typeName)
	field := ormable.Fields[fieldName]

	if field == nil {
		return
	}

	if field.GetHasMany().GetPositionField() != "" {
		positionField := field.GetHasMany().GetPositionField()
		positionFieldType := b.getOrmable(field.TypeName).Fields[positionField].TypeName
		g.P(`for i, e := range `, `to.`, fieldName, `{`)
		g.P(`e.`, positionField, ` = `, positionFieldType, `(i)`)
		g.P(`}`)
	}
}

// Output code that will convert a field to/from orm.
func (b *ORMBuilder) generateFieldConversion(message *protogen.Message, field *protogen.Field,
	toORM bool, ofield *Field, g *protogen.GeneratedFile) error {

	fieldName := camelCase(string(field.Desc.Name()))
	fieldType := field.Desc.Kind().String() // was GoType
	if field.Desc.Message() != nil {
		parts := strings.Split(string(field.Desc.Message().FullName()), ".")
		fieldType = parts[len(parts)-1]
	}
	if field.Desc.Cardinality() == protoreflect.Repeated {
		// Some repeated fields can be handled by github.com/lib/pq
		if b.dbEngine == ENGINE_POSTGRES && b.IsAbleToMakePQArray(fieldType) && field.Desc.IsList() {
			g.P(`if m.`, fieldName, ` != nil {`)
			switch fieldType {
			case "bool":
				g.P(`to.`, fieldName, ` = make(`, generateImport("BoolArray", pqImport, g), `, len(m.`, fieldName, `))`)
			case "double":
				g.P(`to.`, fieldName, ` = make(`, generateImport("Float64Array", pqImport, g), `, len(m.`, fieldName, `))`)
			case "int64":
				g.P(`to.`, fieldName, ` = make(`, generateImport("Int64Array", pqImport, g), `, len(m.`, fieldName, `))`)
			case "string":
				g.P(`to.`, fieldName, ` = make(`, generateImport("StringArray", pqImport, g), `, len(m.`, fieldName, `))`)
			}
			g.P(`copy(to.`, fieldName, `, m.`, fieldName, `)`)
			g.P(`}`)
		} else if b.isOrmable(fieldType) { // Repeated ORMable type
			// fieldType = strings.Trim(fieldType, "[]*")

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
		fieldType = b.typeName(field.Enum.GoIdent, g)
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
		// Check for WKTs
		// Type is a WKT, convert to/from as ptr to base type
		if _, exists := wellKnownTypes[fieldType]; exists { // Singular WKT -----
			if toORM {
				g.P(`if m.`, fieldName, ` != nil {`)
				g.P(`v := m.`, fieldName, `.Value`)
				g.P(`to.`, fieldName, ` = &v`)
				g.P(`}`)
			} else {
				g.P(`if m.`, fieldName, ` != nil {`)
				// g.P(`to.`, fieldName, ` = &`, b.GetFileImports().wktPkgName, ".", fieldType,
				// 	`{Value: *m.`, fieldName, `}`)
				g.P(`to.`, fieldName, ` = &`, generateImport(fieldType, wktImport, g),
					`{Value: *m.`, fieldName, `}`)
				g.P(`}`)
			}
		} else if fieldType == protoTypeBigInt { // Singular BigInt type ----
			if toORM {
				g.P(`if m.`, fieldName, ` != nil {`)
				g.P(`var ok bool`)
				g.P(`to.`, fieldName, ` = new(big.Int)`)
				g.P(`to.`, fieldName, `, ok = to.`, fieldName, `.SetString(m.`, fieldName, `.Value, 0)`)
				g.P(`if !ok {`)
				g.P(`return to, fmt.Errorf("unable convert `, fieldName, ` to big.Int")`)
				g.P(`}`)
				g.P(`}`)
			} else {
				g.P(`to.`, fieldName, ` = &`, generateImport("BigInt", gtypesImport, g), `{Value: m.`, fieldName, `.String()}`)
			}
		} else if fieldType == protoTypeUUIDValue { // Singular UUIDValue type ----
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
		} else if fieldType == protoTypeUUID { // Singular UUID type --------------
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
		} else if fieldType == protoTypeTimestamp { // Singular WKT Timestamp ---
			if toORM {
				g.P(`if m.`, fieldName, ` != nil {`)
				g.P(`t := m.`, fieldName, `.AsTime()`)
				g.P(`to.`, fieldName, ` = &t`)
				g.P(`}`)
			} else {
				g.P(`if m.`, fieldName, ` != nil {`)
				g.P(`to.`, fieldName, ` = `, generateImport("New", timestampImport, g), `(*m.`, fieldName, `)`)
				g.P(`}`)
			}
		} else if fieldType == protoTypeDuration {
			if toORM {
				g.P(`if m.`, fieldName, ` != nil {`)
				g.P(`t := m.`, fieldName, `.AsDuration()`)
				g.P(`to.`, fieldName, ` = &t`)
				g.P(`}`)
			} else {
				g.P(`if m.`, fieldName, ` != nil {`)
				g.P(`to.`, fieldName, ` = `, generateImport("New", durationImport, g), `(*m.`, fieldName, `)`)
				g.P(`}`)
			}
		} else if fieldType == protoTypeJSON {
			if toORM {
				g.P(`if m.`, fieldName, ` != nil {`)
				g.P(`to.`, fieldName, ` = &`, generateImport("Jsonb", gtypesImport, g), `{[]byte(m.`, fieldName, `.Value)}`)
				g.P(`}`)
			} else {
				g.P(`if m.`, fieldName, ` != nil {`)
				g.P(`to.`, fieldName, ` = &`, generateImport("JSONValue", gtypesImport, g), `{Value: string(m.`, fieldName, `.RawMessage)}`)
				g.P(`}`)
			}
			// Potential TODO other DB engine handling if desired
		} else if fieldType == protoTypeResource {
			resource := "nil" // assuming we do not know the PB type, nil means call codec for any resource
			if ofield != nil && ofield.ParentOrigName != "" {
				resource = "&" + ofield.ParentOrigName + "{}"
			}
			btype := strings.TrimPrefix(ofield.TypeName, "*")
			nillable := strings.HasPrefix(ofield.TypeName, "*")
			iface := ofield.TypeName == "interface{}"

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
		} else if fieldType == protoTypeInet { // Inet type for Postgres only, currently
			if toORM {
				g.P(`if m.`, fieldName, ` != nil {`)
				g.P(`if to.`, fieldName, `, err = `, generateImport("ParseInet", gtypesImport, g), `(m.`, fieldName, `.Value); err != nil {`)
				g.P(`return to, err`)
				g.P(`}`)
				g.P(`}`)
			} else {
				g.P(`if m.`, fieldName, ` != nil && m.`, fieldName, `.IPNet != nil {`)
				g.P(`to.`, fieldName, ` = &`, generateImport("InetValue", gtypesImport, g), `{Value: m.`, fieldName, `.String()}`)
				g.P(`}`)
			}
		} else if fieldType == protoTimeOnly { // Time only to support time via string
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

			if b.hasCompositePrimaryKey(ormable) {
				// TODO support for Update,Patch,DeleteSet when composite primary keys exist
				b.generateReadHandler(message, g)
				b.generateDeleteHandler(message, g)
			} else if b.hasPrimaryKey(ormable) {
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

// fieldPath defines the path to the nested field
//
// NOTE: fieldPath stores path in reverse order internally
type fieldPath struct {
	// path is reversed path to the field
	path []string
}

func NewFieldPath(part string) *fieldPath {
	return &fieldPath{path: []string{part}}
}

func (p *fieldPath) String() string {
	if len(p.path) == 0 {
		return ""
	}

	bldr := strings.Builder{}
	bldr.WriteString(p.path[len(p.path)-1])
	for i := len(p.path) - 2; i >= 0; i-- {
		bldr.WriteString(p.path[i])
		bldr.WriteString(".")
	}

	return bldr.String()
}

func (p *fieldPath) Quoted() string {
	return fmt.Sprintf(`"%s"`, p)
}

func (p *fieldPath) Add(part string) *fieldPath {
	p.path = append(p.path)
	return p
}

func parseRecursiveFields(model *OrmableType, take func(*Field) bool) []*fieldPath {
	// handled is used to check for recursive models
	// if the model was already handled it should not be handled one more time
	handled := make(map[string]struct{})

	var rec func(*OrmableType) []*fieldPath
	rec = func(m *OrmableType) (paths []*fieldPath) {
		if m == nil {
			return
		}

		if _, ok := handled[m.Name]; ok {
			return
		}
		// mark model as handled already
		handled[m.Name] = struct{}{}

		for name, field := range m.Fields {
			if take(field) {
				paths = append(paths, NewFieldPath(name))
				continue
			}

			subpaths := rec(field.Type)
			for _, path := range subpaths {
				paths = append(paths, path.Add(name))
			}
		}

		return
	}

	return rec(model)
}

// fieldPathsToQuoted parses each path and qoutes it in ["{path.String()}",] separated by ','
func fieldPathsToQuoted(paths []*fieldPath) string {
	strs := make([]string, 0, len(paths))
	for _, path := range paths {
		strs = append(strs, path.Quoted())
	}

	return strings.Join(strs, ", ")
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
	omitPaths := parseRecursiveFields(orm,
		func(f *Field) bool {
			// check only fields with association info (e.g. other Gorm types)
			if f.FieldAssociationInfo == nil {
				return false
			}
			// omit field if autocreate disabled
			return f.FieldAssociationInfo.GetDisableAssociationAutocreate()
		},
	)

	preloadPaths := parseRecursiveFields(orm,
		func(f *Field) bool {
			// check only fields with association info (e.g. other Gorm types)
			if f.FieldAssociationInfo == nil {
				return false
			}
			return f.FieldAssociationInfo.GetPreload()
		},
	)

	preloadBldr := strings.Builder{}
	for _, path := range preloadPaths {
		preloadBldr.WriteString(`Preload(`)
		preloadBldr.WriteString(path.Quoted())
		preloadBldr.WriteString(`).`)
	}

	g.P(`if err = db.Omit(`, fieldPathsToQuoted(omitPaths), `).`,
		preloadBldr.String(),
		`Create(&ormObj).Error; err != nil {`)
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
		g.P(`func DefaultRead`, typeName, `(ctx context.Context, in *`,
			typeName, `, db *`, generateImport("DB", gormImport, g), `, fs *`, generateImport("FieldSelection", queryImport, g), `) (*`, typeName, `, error) {`)
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

	pkfieldMapping := b.getPrimaryKeys(ormable)
	for _, pkfieldObj := range pkfieldMapping {
		if strings.Contains(pkfieldObj.field.TypeName, "*") {
			g.P(`if ormObj.`, pkfieldObj.name, ` == nil || *ormObj.`, pkfieldObj.name, ` == `, b.guessZeroValue(pkfieldObj.field.TypeName, g), ` {`)
		} else {
			g.P(`if ormObj.`, pkfieldObj.name, ` == `, b.guessZeroValue(pkfieldObj.field.TypeName, g), ` {`)
		}
		g.P(`return nil, `, "errors", `.EmptyIdError`)
		g.P(`}`)
	}

	var fs string
	if b.readHasFieldSelection(ormable) {
		fs = "fs"
	} else {
		fs = "nil"
	}

	b.generateBeforeReadHookCall(ormable, "ApplyQuery", g)
	if fs != "nil" {
		g.P(`if db, err = `, generateImport("ApplyFieldSelection", tkgormImport, g), `(ctx, db, `, fs, `, &`, ormable.Name, `{}); err != nil {`)
		g.P(`return nil, err`)
		g.P(`}`)
	}

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
	for _, method := range ormable.Methods {
		if method.verb == readService {
			if s := b.getFieldSelection(method.inType); s != "" {
				return true
			}
		}
	}
	return false
}

// guessZeroValue of the input type, so that we can check if a (key) value is set or not
func (b *ORMBuilder) guessZeroValue(typeName string, g *protogen.GeneratedFile) string {
	typeName = strings.ToLower(typeName)
	if strings.Contains(typeName, "string") {
		return `""`
	}
	if strings.Contains(typeName, "int") {
		return `0`
	}
	if strings.Contains(typeName, "uuid") {
		return generateImport("Nil", uuidImport, g)
	}
	if strings.Contains(typeName, "bigint") {
		return generateImport("Nil", bigintImport, g)
	}
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
	pkFieldMapping := b.getPrimaryKeys(ormable)
	for _, pkFieldObj := range pkFieldMapping {
		if strings.Contains(pkFieldObj.field.TypeName, "*") {
			g.P(`if ormObj.`, pkFieldObj.name, ` == nil || *ormObj.`, pkFieldObj.name, ` == `, b.guessZeroValue(pkFieldObj.field.TypeName, g), ` {`)
		} else {
			g.P(`if ormObj.`, pkFieldObj.name, ` == `, b.guessZeroValue(pkFieldObj.field.TypeName, g), `{`)
		}
		g.P(`return `, generateImport("EmptyIdError", gerrorsImport, g))
		g.P(`}`)
	}

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
	ns := gschema.NamingStrategy{SingularTable: true}
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
	g.P(`keys := []`, pk.TypeName, `{}`)
	g.P(`for _, obj := range in {`)
	g.P(`ormObj, err := obj.ToORM(ctx)`)
	g.P(`if err != nil {`)
	g.P(`return err`)
	g.P(`}`)
	if strings.Contains(pk.TypeName, "*") {
		g.P(`if ormObj.`, pkName, ` == nil || *ormObj.`, pkName, ` == `, b.guessZeroValue(pk.TypeName, g), ` {`)
	} else {
		g.P(`if ormObj.`, pkName, ` == `, b.guessZeroValue(pk.TypeName, g), `{`)
	}
	g.P(`return `, generateImport("EmptyIdError", gerrorsImport, g))
	g.P(`}`)
	g.P(`keys = append(keys, ormObj.`, pkName, `)`)
	g.P(`}`)
	b.generateBeforeDeleteSetHookCall(ormable, g)

	if getMessageOptions(message).GetMultiAccount() {
		g.P(`accountId, err := `, generateImport("GetAccountID", authImport, g), `(ctx, nil)`)
		g.P(`if err != nil {`)
		g.P(`return err`)
		g.P(`}`)

		if getMessageOptions(message).GetMultiCompartment() {
			g.P(`compartmentId, err := `, generateImport("GetCompartmentID", authImport, g), `(ctx, nil)`)
			g.P(`if err != nil {`)
			g.P(`return err`)
			g.P(`}`)
			g.P(`if compartmentId != "" {`)
			g.P(`err = db.Where("account_id = ? AND compartment_id like ?% AND `, ns.TableName(pkName), ` in (?)", accountId, compartmentId, keys).Delete(&`, ormable.Name, `{}).Error`)
			g.P(`if err != nil {`)
			g.P(`return err`)
			g.P(`}`)
			g.P(`} else {`)
			g.P(`err = db.Where("account_id = ? AND `, ns.TableName(pkName), ` in (?)", accountId, keys).Delete(&`, ormable.Name, `{}).Error`)
			g.P(`if err != nil {`)
			g.P(`return err`)
			g.P(`}`)
			g.P(`}`)
		} else {
			g.P(`err = db.Where("account_id = ? AND `, ns.TableName(pkName), ` in (?)", accountId, keys).Delete(&`, ormable.Name, `{}).Error`)
			g.P(`if err != nil {`)
			g.P(`return err`)
			g.P(`}`)
		}
	} else {
		g.P(`err = db.Where("`, ns.TableName(pkName), ` in (?)", keys).Delete(&`, ormable.Name, `{}).Error`)
		g.P(`if err != nil {`)
		g.P(`return err`)
		g.P(`}`)
	}

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
		if getMessageOptions(message).GetMultiCompartment() {
			b.generateAccountIdAndCompartmentIdWhereClause(g)
		} else {
			b.generateAccountIdWhereClause(g)
		}
	}

	ormable := b.getOrmable(typeName)
	if b.gateway {
		g.P(`var count int64`)
	}

	if b.hasPrimaryKey(ormable) {
		pkName, pk := b.findPrimaryKey(ormable)
		column := pk.GetTag().GetColumn()
		if len(column) == 0 {
			column = gschema.NamingStrategy{SingularTable: true}.TableName(pkName)
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
	omitPaths := parseRecursiveFields(ormable,
		func(f *Field) bool {
			// check only fields with association info (e.g. other Gorm types)
			if f.FieldAssociationInfo == nil {
				return false
			}
			// omit field if autoupdate disabled
			return f.FieldAssociationInfo.GetDisableAssociationAutoupdate()
		},
	)

	preloadPaths := parseRecursiveFields(ormable,
		func(f *Field) bool {
			// check only fields with association info (e.g. other Gorm types)
			if f.FieldAssociationInfo == nil {
				return false
			}
			return f.FieldAssociationInfo.GetPreload()
		},
	)

	preloadBldr := strings.Builder{}
	for _, path := range preloadPaths {
		preloadBldr.WriteString(`Preload(`)
		preloadBldr.WriteString(path.Quoted())
		preloadBldr.WriteString(`).`)
	}

	g.P(`if err = db.Omit(`, fieldPathsToQuoted(omitPaths), `).`,
		preloadBldr.String(),
		`Save(&ormObj).Error; err != nil {`)
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

func (b *ORMBuilder) generateAccountIdAndCompartmentIdWhereClause(g *protogen.GeneratedFile) {
	g.P(`accountID, err := `, generateImport("GetAccountID", authImport, g), `(ctx, nil)`)
	g.P(`if err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)
	g.P(`compartmentID, err := `, generateImport("GetCompartmentID", authImport, g), `(ctx, nil)`)
	g.P(`if err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)
	g.P(`if compartmentID != "" {`)
	g.P(`db = db.Where(map[string]interface{}{"account_id": accountID, "compartment_id": compartmentID})`)
	g.P(`} else {`)
	g.P(`db = db.Where(map[string]interface{}{"account_id": accountID})`)
	g.P(`}`)
}

func (b *ORMBuilder) handleChildAssociations(message *protogen.Message, g *protogen.GeneratedFile) {
	ormable := b.getOrmable(string(message.Desc.Name()))

	var fieldNames []string
	for name := range ormable.Fields {
		fieldNames = append(fieldNames, name)
	}
	sort.Strings(fieldNames)

	for _, fieldName := range fieldNames {
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

		g.P(`if err = db.Model(&ormObj).Association("`, fieldName, `").`, action, `; err != nil {`)
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
		assocKeyType := ormable.Fields[assocKeyName].TypeName
		assocOrmable := b.getOrmable(field.TypeName)
		foreignKeyType := assocOrmable.Fields[foreignKeyName].TypeName
		g.P(`filter`, fieldName, ` := `, strings.Trim(field.TypeName, "[]*"), `{}`)
		zeroValue := b.guessZeroValue(assocKeyType, g)
		if strings.Contains(assocKeyType, "*") {
			g.P(`if ormObj.`, assocKeyName, ` == nil || *ormObj.`, assocKeyName, ` == `, zeroValue, `{`)
		} else {
			g.P(`if ormObj.`, assocKeyName, ` == `, zeroValue, `{`)
		}
		g.P(`return nil, `, generateImport("EmptyIdError", gerrorsImport, g))
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
		g.P(`if err = db.Where(filter`, fieldName, `).Delete(`, strings.Trim(field.TypeName, "[]*"), `{}).Error; err != nil {`)
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

	// TODO: not in original code, but it doesn't make a lot of sense to generate code with id if message doesn't have it
	if b.hasIDField(message) {
		getIDFormatter := "{Id: in.GetId()},"
		if b.IsIDFieldOptional(message) {
			// This is necessary because the GetID returns a non pointer object
			getIDFormatter = "{Id: in.Id},"
		}
		if b.readHasFieldSelection(ormable) {
			g.P(`pbReadRes, err := DefaultRead`, typeName, `(ctx, &`, typeName, getIDFormatter, ` db, nil)`)
		} else {
			g.P(`pbReadRes, err := DefaultRead`, typeName, `(ctx, &`, typeName, getIDFormatter, ` db)`)
		}

		g.P(`if err != nil {`)
		g.P(`return nil, err`)
		g.P(`}`)
		g.P(`pbObj = *pbReadRes`)
	}

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

// IsIDFieldOptional if the ID is an optional field
func (b *ORMBuilder) IsIDFieldOptional(message *protogen.Message) bool {
	for _, field := range message.Fields {
		if strings.ToLower(field.GoName) == "id" {
			if field.Desc.HasOptionalKeyword() {
				return true
			}
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
		fieldType := getFieldType(field)

		if field.Message != nil && !isSpecialType(fieldType) && field.Desc.Cardinality() != protoreflect.Repeated {
			g.P(`var updated`, camelCase(field.GoName), ` bool`)
			hasNested = true
		} else if strings.HasSuffix(fieldType, protoTypeJSON) && field.Desc.Cardinality() != protoreflect.Repeated {
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

		fieldType := getFieldType(field)
		//  for ormable message, do recursive patching
		if field.Message != nil && b.isOrmable(fieldType) && field.Desc.Cardinality() != protoreflect.Repeated {
			if field.Message != nil {
				// a hack work around imported types
				fieldType = b.typeName(field.Message.GoIdent, g)
			}
			_ = generateImport("", stdStringsImport, g)
			g.P(`if !updated`, ccName, ` && strings.HasPrefix(f, prefix+"`, ccName, `.") {`)
			g.P(`updated`, ccName, ` = true`)
			g.P(`if patcher.`, ccName, ` == nil {`)
			g.P(`patchee.`, ccName, ` = nil`)
			g.P(`continue`)
			g.P(`}`)
			g.P(`if patchee.`, ccName, ` == nil {`)
			g.P(`patchee.`, ccName, ` = &`, strings.TrimPrefix(b.typeName(getFieldIdent(field), g), "*"), `{}`)
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
			_ = generateImport("", stdStringsImport, g)
			g.P(`if !updated`, ccName, ` && strings.HasPrefix(f, prefix+"`, ccName, `.") {`)
			g.P(`if patcher.`, ccName, ` == nil {`)
			g.P(`patchee.`, ccName, ` = nil`)
			g.P(`continue`)
			g.P(`}`)
			g.P(`if patchee.`, ccName, ` == nil {`)
			g.P(`patchee.`, ccName, ` = &`, strings.TrimPrefix(b.typeName(getFieldIdent(field), g), "*"), `{}`)
			g.P(`}`)
			g.P(`childMask := &`, generateImport("FieldMask", fmImport, g), `{}`)
			g.P(`for j := i; j < len(updateMask.Paths); j++ {`)
			g.P(`if trimPath := strings.TrimPrefix(updateMask.Paths[j], prefix+"`, ccName, `."); trimPath != updateMask.Paths[j] {`)
			g.P(`childMask.Paths = append(childMask.Paths, trimPath)`)
			g.P(`}`)
			g.P(`}`)
			g.P(`if err := `, generateImport("MergeWithMask", tkgormImport, g), `(patcher.`, ccName, `, patchee.`, ccName, `, childMask); err != nil {`)
			g.P(`return nil, nil`)
			g.P(`}`)
			g.P(`}`)
			g.P(`if f == prefix+"`, ccName, `" {`)
			g.P(`updated`, ccName, ` = true`)
			g.P(`patchee.`, ccName, ` = patcher.`, ccName)
			g.P(`continue`)
			g.P(`}`)
		} else if strings.HasSuffix(fieldType, protoTypeJSON) && field.Desc.Cardinality() != protoreflect.Repeated {
			_ = generateImport("", stdStringsImport, g)
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
	switch typeName {
	case protoTypeJSON, protoTypeBigInt, protoTypeUUID, protoTypeUUIDValue, protoTypeResource, protoTypeInet, protoTimeOnly:
		return true
	default:
		return false
	}
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
	if f != "nil" || s != "nil" || pg != "nil" || fs != "nil" {
		g.P(`db, err = `, generateImport("ApplyCollectionOperators", tkgormImport, g), `(ctx, db, &`, ormable.Name, `{}, &`, typeName, `{}, `, f, `,`, s, `,`, pg, `,`, fs, `)`)
		g.P(`if err != nil {`)
		g.P(`return nil, err`)
		g.P(`}`)
	}
	b.generateBeforeListHookCall(ormable, "Find", g)
	g.P(`db = db.Where(&ormObj)`)

	// TODO handle composite primary keys order considering priority tag
	if b.hasCompositePrimaryKey(ormable) {
		pkFieldMapping := b.getPrimaryKeys(ormable)
		var columns []string
		for _, pkFieldObj := range pkFieldMapping {
			column := pkFieldObj.field.GetTag().GetColumn()
			if len(column) == 0 {
				column = gschema.NamingStrategy{SingularTable: true}.TableName(pkFieldObj.name)
			}
			columns = append(columns, column)
		}
		g.P(`db = db.Order("`, strings.Join(columns, ", "), `")`)
	} else if b.hasPrimaryKey(ormable) {
		pkName, pk := b.findPrimaryKey(ormable)
		column := pk.GetTag().GetColumn()
		if len(column) == 0 {
			column = gschema.NamingStrategy{SingularTable: true}.TableName(pkName)
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
	hookCall := `if err = hook.AfterListFind(ctx, db, &ormResponse`
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

func (b *ORMBuilder) generateBeforeListHookDef(orm *OrmableType, suffix string, g *protogen.GeneratedFile) {
	g.P(`type `, orm.Name, `WithBeforeList`, suffix, ` interface {`)
	hookSign := fmt.Sprint(`BeforeList`, suffix, `(context.Context, *`, generateImport("DB", gormImport, g))
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
	hookSign += `) error`
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
	for _, method := range ormable.Methods {
		if s := b.getFiltering(method.inType); s != "" {
			return true
		}
	}
	return false
}

func (b *ORMBuilder) listHasSorting(ormable *OrmableType) bool {
	for _, method := range ormable.Methods {
		if s := b.getSorting(method.inType); s != "" {
			return true
		}
	}

	return false
}

func (b *ORMBuilder) listHasPagination(ormable *OrmableType) bool {
	for _, method := range ormable.Methods {
		if s := b.getPagination(method.inType); s != "" {
			return true
		}
	}

	return false
}

func (b *ORMBuilder) listHasFieldSelection(ormable *OrmableType) bool {
	for _, method := range ormable.Methods {
		if s := b.getFieldSelection(method.inType); s != "" {
			return true
		}
	}

	return false
}

func (b *ORMBuilder) parseServices(file *protogen.File) {
	for _, service := range file.Services {
		genSvc := autogenService{
			Service: service,
			ccName:  camelCase(string(service.Desc.Name())),
			file:    file,
		}

		if opts := getServiceOptions(service); opts != nil {
			genSvc.autogen = opts.GetAutogen()
			genSvc.usesTxnMiddleware = opts.GetTxnMiddleware()
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
				follows, baseType = b.followsDeleteSetConventions(input, output, method)
			} else if strings.HasPrefix(methodName, deleteService) {
				verb = deleteService
				follows, baseType = b.followsDeleteConventions(input, output, method)
			} else if strings.HasPrefix(methodName, listService) {
				verb = listService
				follows, baseType = b.followsListConventions(input, output, listService)
			}

			genMethod := autogenMethod{
				Method:            method,
				ccName:            methodName,
				inType:            input,
				outType:           output,
				fieldMaskName:     fmName,
				verb:              verb,
				followsConvention: follows,
				baseType:          baseType,
			}

			genSvc.methods = append(genSvc.methods, genMethod)

			if genMethod.verb != "" && b.isOrmable(genMethod.baseType) {
				b.getOrmable(genMethod.baseType).Methods = append(b.getOrmable(genMethod.baseType).Methods, &genMethod)
			}
		}

		b.ormableServices = append(b.ormableServices, genSvc)
	}
}

func (b *ORMBuilder) followsCreateConventions(inType *protogen.Message, outType *protogen.Message, methodName string) (bool, string) {
	var inTypeName string
	var typeOrmable bool
	for _, field := range inType.Fields {
		if string(field.Desc.Name()) == "payload" && field.Desc.Message() != nil {
			gType := string(field.Desc.Message().Name())
			inTypeName = strings.TrimPrefix(gType, "*")
			if b.isOrmable(inTypeName) {
				typeOrmable = true
			}
		}
	}
	if !typeOrmable {
		fmt.Fprintf(os.Stderr, "stub will be generated for %s since %s incoming message doesn't have \"payload\" field of ormable type.\n", methodName, inType.Desc.Name())
		return false, ""
	}
	var outTypeName string
	for _, field := range outType.Fields {
		if string(field.Desc.Name()) == "result" {
			gType := string(field.Desc.Message().Name())
			outTypeName = strings.TrimPrefix(gType, "*")
		}
	}
	if inTypeName != outTypeName {
		fmt.Fprintf(os.Stderr, "stub will be generated for %s since \"payload\" field type of %s incoming message type doesn't match \"result\" field type of %s outcoming message.\n", methodName,
			inType.Desc.Name(), outType.Desc.Name())
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
		fmt.Fprintf(os.Stderr, "stub will be generated for %s since %s incoming message doesn't have \"id\" field", methodName, inType.Desc.Name())
		return false, ""
	}

	var outTypeName string
	var typeOrmable bool
	for _, field := range outType.Fields {
		if string(field.Desc.Name()) == "result" {
			gType := string(field.Desc.Message().Name())
			outTypeName = strings.TrimPrefix(gType, "*")
			if b.isOrmable(outTypeName) {
				typeOrmable = true
			}
		}
	}
	if !typeOrmable {
		fmt.Fprintf(os.Stderr, "stub will be generated for %s since %s outcoming message doesn't have \"result\" field of ormable type.\n", methodName, outTypeName)
		return false, ""
	}
	if !b.hasPrimaryKey(b.getOrmable(outTypeName)) {
		fmt.Fprintf(os.Stderr, "stub will be generated for %s since %s ormable type doesn't have a primary key.\n", methodName, outTypeName)
		return false, ""
	}

	return true, outTypeName
}

func (b *ORMBuilder) followsUpdateSetConventions(inType *protogen.Message, outType *protogen.Message, methodName string) (bool, string, string) {
	var (
		inEntity    *protogen.Field
		inFieldMask *protogen.Field
	)

	for _, field := range inType.Fields {
		if string(field.Desc.Name()) == "objects" {
			inEntity = field
		}

		if string(field.Desc.Message().FullName()) == "google.protobuf.FieldMask" {
			if inFieldMask != nil {
				fmt.Fprintf(os.Stderr, "message must not contains double field mask, prev on field name %s, after on field %s.\n", inFieldMask.GoName, field.GoName)
				return false, "", ""
			}

			inFieldMask = field
		}
	}

	var outEntity *protogen.Field
	for _, field := range outType.Fields {
		if string(field.Desc.Name()) == "results" {
			outEntity = field
		}
	}

	if inFieldMask == nil || inFieldMask.Desc.Cardinality() != protoreflect.Repeated {
		fmt.Fprintf(os.Stderr, "repeated field mask should exist in request for method %q.\n", methodName)
		return false, "", ""
	}

	if inEntity == nil || outEntity == nil {
		fmt.Fprintf(os.Stderr, `method: %q, request should has repeated field 'objects' in request and repeated field 'results' in response.\n`, methodName)
		return false, "", ""
	}

	if inEntity.Desc.Cardinality() != protoreflect.Repeated || outEntity.Desc.Cardinality() != protoreflect.Repeated {
		fmt.Fprintf(os.Stderr, `method: %q, field 'objects' in request and field 'results' in response should be repeated.\n`, methodName)
		return false, "", ""
	}

	inGoType := string(inEntity.Message.Desc.Name())
	outGoType := string(outEntity.Message.Desc.Name())
	inTypeName, outTypeName := strings.TrimPrefix(inGoType, "*"), strings.TrimPrefix(outGoType, "*")
	if !b.isOrmable(inTypeName) {
		fmt.Fprintf(os.Stderr, "method: %q, type %q must be ormable.\n", methodName, inTypeName)
		return false, "", ""
	}

	if inTypeName != outTypeName {
		fmt.Fprintf(os.Stderr, "method: %q, field 'objects' in request has type: %q but field 'results' in response has: %q.\n", methodName, inTypeName, outTypeName)
		return false, "", ""
	}

	return true, inTypeName, camelCase(inFieldMask.GoName)
}

func (b *ORMBuilder) followsUpdateConventions(inType *protogen.Message, outType *protogen.Message, methodName string) (bool, string, string) {
	var inTypeName string
	var typeOrmable bool
	var updateMask string
	for _, field := range inType.Fields {
		if string(field.Desc.Name()) == "payload" && field.Desc.Message() != nil {
			gType := string(field.Desc.Message().Name())
			inTypeName = strings.TrimPrefix(gType, "*")
			if b.isOrmable(inTypeName) {
				typeOrmable = true
			}
		}

		// Check that type of field is a FieldMask
		if string(field.Desc.Message().FullName()) == "google.protobuf.FieldMask" {
			// More than one mask in request is not allowed.
			if updateMask != "" {
				return false, "", ""
			}
			updateMask = string(field.Desc.Name())
		}

	}
	if !typeOrmable {
		fmt.Fprintf(os.Stderr, `stub will be generated for %s since %s incoming message doesn't have "payload" field of ormable type.\n`, methodName, inType.Desc.Name())
		return false, "", ""
	}

	var outTypeName string
	for _, field := range outType.Fields {
		if string(field.Desc.Name()) == "result" {
			gType := string(field.Desc.Message().Name())
			outTypeName = strings.TrimPrefix(gType, "*")
		}
	}
	if inTypeName != outTypeName {
		fmt.Fprintf(os.Stderr, `stub will be generated for %s since "payload" field type of %s incoming message doesn't match "result" field type of %s outcoming message.\n`,
			methodName, inType.Desc.Name(), outType.Desc.Name())
		return false, "", ""
	}
	if !b.hasPrimaryKey(b.getOrmable(inTypeName)) {
		fmt.Fprintf(os.Stderr, `stub will be generated for %s since %s ormable type doesn't have a primary key.\n`, methodName, outTypeName)
		return false, "", ""
	}

	return true, inTypeName, camelCase(updateMask)
}

func (b *ORMBuilder) followsDeleteSetConventions(inType *protogen.Message, outType *protogen.Message, method *protogen.Method) (bool, string) {
	var hasIDs bool

	for _, field := range inType.Fields {
		if string(field.Desc.Name()) == "ids" && field.Desc.Cardinality() == protoreflect.Repeated {
			hasIDs = true
		}
	}

	methodName := string(method.Desc.Name())
	if !hasIDs {
		fmt.Fprintf(os.Stderr, `stub will be generated for %s since %s incoming message doesn't have "ids" field.\n`, methodName, inType.Desc.Name())
		return false, ""
	}
	typeName := camelCase(getMethodOptions(method).GetObjectType())

	if typeName == "" {
		fmt.Fprintf(os.Stderr, `stub will be generated for %s since (gorm.method).object_type option is not specified.\n`, methodName)
		return false, ""
	}

	if !b.isOrmable(typeName) {
		fmt.Fprintf(os.Stderr, `stub will be generated for %s since %s is not an ormable type.\n`, methodName, typeName)
		return false, ""
	}

	if !b.hasPrimaryKey(b.getOrmable(typeName)) {
		fmt.Fprintf(os.Stderr, `stub will be generated for %s since %s ormable type doesn't have a primary key.\n`, methodName, typeName)
		return false, ""
	}

	return true, typeName
}

func (b *ORMBuilder) followsDeleteConventions(inType *protogen.Message, outType *protogen.Message, method *protogen.Method) (bool, string) {
	var hasID bool
	for _, field := range inType.Fields {
		if string(field.Desc.Name()) == "id" {
			hasID = true
		}
	}

	methodName := string(method.Desc.Name())
	if !hasID {
		fmt.Fprintf(os.Stderr, `stub will be generated for %s since %s incoming message doesn't have "id" field.\n`, methodName, inType.Desc.Name())
		return false, ""
	}

	typeName := camelCase(getMethodOptions(method).GetObjectType())
	if typeName == "" {
		fmt.Fprintf(os.Stderr, `stub will be generated for %s since (gorm.method).object_type option is not specified.\n`, methodName)
		return false, ""
	}

	if !b.isOrmable(typeName) {
		fmt.Fprintf(os.Stderr, `stub will be generated for %s since %s is not an ormable type.\n`, methodName, typeName)
		return false, ""
	}

	if !b.hasPrimaryKey(b.getOrmable(typeName)) {
		fmt.Fprintf(os.Stderr, `stub will be generated for %s since %s ormable type doesn't have a primary key`, methodName, typeName)
		return false, ""
	}

	return true, typeName
}

func (b *ORMBuilder) followsListConventions(inType *protogen.Message, outType *protogen.Message, methodName string) (bool, string) {
	var outTypeName string
	var typeOrmable bool
	for _, field := range outType.Fields {
		if string(field.Desc.Name()) == "results" {
			gType := string(field.Desc.Message().Name())
			outTypeName = strings.TrimPrefix(gType, "[]*")
			if b.isOrmable(outTypeName) {
				typeOrmable = true
			}
		}
	}
	if !typeOrmable {
		fmt.Fprintf(os.Stderr, `stub will be generated for %s since %s incoming message doesn't have "results" field of ormable type`, methodName, outType.Desc.Name())
		return false, ""
	}

	return true, outTypeName
}

func getServiceOptions(service *protogen.Service) *gormopts.AutoServerOptions {
	options := service.Desc.Options().(*descriptorpb.ServiceOptions)
	if options == nil {
		return nil
	}

	v := proto.GetExtension(options, gormopts.E_Server)
	if v == nil {
		return nil
	}

	opts, ok := v.(*gormopts.AutoServerOptions)
	if !ok {
		return nil
	}

	return opts
}

func getMethodOptions(method *protogen.Method) *gormopts.MethodOptions {
	options := method.Desc.Options().(*descriptorpb.MethodOptions)
	if options == nil {
		return nil
	}

	v := proto.GetExtension(options, gormopts.E_Method)
	if v == nil {
		return nil
	}

	opts, ok := v.(*gormopts.MethodOptions)
	if !ok {
		return nil
	}

	return opts
}

func (b *ORMBuilder) generateDefaultServer(file *protogen.File, g *protogen.GeneratedFile) {
	for _, service := range b.ormableServices {
		if service.file != file || !service.autogen {
			continue
		}

		g.P(`type `, service.ccName, `DefaultServer struct {`)
		if !service.usesTxnMiddleware {
			g.P(`DB *`, generateImport("DB", gormImport, g))
		}
		g.P(`}`)

		withSpan := getServiceOptions(service.Service).WithTracing

		if withSpan {
			b.generateSpanInstantiationMethod(service, g)
			b.generateSpanErrorMethod(service, g)
			b.generateSpanResultMethod(service, g)
		}

		for _, method := range service.methods {
			_ = generateImport("", "context", g)
			switch method.verb {
			case createService:
				b.generateCreateServerMethod(service, method, g)
			case readService:
				b.generateReadServerMethod(service, method, g)
			case updateService:
				b.generateUpdateServerMethod(service, method, g)
			case updateSetService:
				b.generateUpdateSetServerMethod(service, method, g)
			case deleteService:
				b.generateDeleteServerMethod(service, method, g)
			case deleteSetService:
				b.generateDeleteSetServerMethod(service, method, g)
			case listService:
				b.generateListServerMethod(service, method, g)
			default:
				b.generateMethodStub(service, method, g)
			}
		}
	}
}

func (b *ORMBuilder) generateSpanInstantiationMethod(service autogenService, g *protogen.GeneratedFile) {
	serviceName := service.GoName
	_ = generateImport("", "fmt", g)
	g.P(`func (m *`, serviceName, `DefaultServer) spanCreate(ctx context.Context, in interface{}, methodName string) (*`, generateImport("Span", ocTraceImport, g), `, error) {`)
	g.P(`_, span := `, generateImport("StartSpan", ocTraceImport, g), `(ctx, fmt.Sprint("`, serviceName, `DefaultServer.", methodName))`)
	g.P(`raw, err := `, generateImport("Marshal", encodingJsonImport, g), `(in)`)
	g.P(`if err != nil {`)
	g.P(`return nil, err`)
	g.P(`}`)
	g.P(`span.Annotate([]`, generateImport("Attribute", ocTraceImport, g), `{`, generateImport("StringAttribute", ocTraceImport, g), `("in", string(raw))}, "in parameter")`)
	g.P(`return span, nil`)
	g.P(`}`)
}

func (b *ORMBuilder) generateSpanErrorMethod(service autogenService, g *protogen.GeneratedFile) {
	g.P(`// spanError ...`)
	g.P(`func (m *`, service.GoName, `DefaultServer) spanError(span *`, generateImport("Span", ocTraceImport, g), `, err error) error {`)
	g.P(`span.SetStatus(`, generateImport("Status", ocTraceImport, g), `{`)
	g.P(`Code: `, generateImport("StatusCodeUnknown", ocTraceImport, g), `,`)
	g.P(`Message: err.Error(),`)
	g.P(`})`)
	g.P(`return err`)
	g.P(`}`)
}

func (b *ORMBuilder) generateSpanResultMethod(service autogenService, g *protogen.GeneratedFile) {
	g.P(`// spanResult ...`)
	g.P(`func (m *`, service.GoName, `DefaultServer) spanResult(span *`, generateImport("Span", ocTraceImport, g), `, out interface{}) error {`)
	g.P(`raw, err := `, generateImport("Marshal", encodingJsonImport, g), `(out)`)
	g.P(`if err != nil {`)
	g.P(`return err`)
	g.P(`}`)
	g.P(`span.Annotate([]`, generateImport("Attribute", ocTraceImport, g), `{`, generateImport("StringAttribute", ocTraceImport, g), `("out", string(raw))}, "out parameter")`)
	g.P(`return nil`)
	g.P(`}`)
}

func (b *ORMBuilder) generateCreateServerMethod(service autogenService, method autogenMethod, g *protogen.GeneratedFile) {
	b.generateMethodSignature(service, method, g)
	if method.followsConvention {
		b.generateDBSetup(service, g)
		b.generatePreserviceCall(service, method.baseType, method.ccName, g)
		g.P(`res, err := DefaultCreate`, method.baseType, `(ctx, in.GetPayload(), db)`)
		g.P(`if err != nil {`)
		g.P(`return nil, `, b.wrapSpanError(service, "err"))
		g.P(`}`)
		g.P(`out := &`, b.typeName(method.outType.GoIdent, g), `{Result: res}`)
		if b.gateway {
			g.P(`err = `, generateImport("SetCreated", gatewayImport, g), `(ctx, "")`)
			g.P(`if err != nil {`)
			g.P(`return nil, `, b.wrapSpanError(service, "err"))
			g.P(`}`)
		}

		b.generatePostserviceCall(service, method.baseType, method.ccName, g)
		b.spanResultHandling(service, g)
		g.P(`return out, nil`)
		g.P(`}`)
		b.generatePreserviceHook(service.ccName, method.baseType, method.ccName, g)
		b.generatePostserviceHook(service.ccName, method.baseType, b.typeName(method.outType.GoIdent, g), method.ccName, g)
	} else {
		b.generateEmptyBody(service, method.outType, g)
	}
}

func (b *ORMBuilder) generateMethodSignature(service autogenService, method autogenMethod, g *protogen.GeneratedFile) {
	in := b.typeName(method.inType.GoIdent, g)
	out := b.typeName(method.outType.GoIdent, g)

	g.P(`// `, method.ccName, ` ...`)
	g.P(`func (m *`, service.GoName, `DefaultServer) `, method.ccName, ` (ctx context.Context, in *`,
		in, `) (*`, out, `, error) {`)
	withSpan := getServiceOptions(service.Service).WithTracing
	if withSpan {
		g.P(`span, errSpanCreate := m.spanCreate(ctx, in, "`, method.ccName, `")`)
		g.P(`if errSpanCreate != nil {`)
		g.P(`return nil, errSpanCreate`)
		g.P(`}`)
		g.P(`defer span.End()`)
	}
}

func (b ORMBuilder) generateEmptyBody(service autogenService, outType *protogen.Message, g *protogen.GeneratedFile) {
	g.P(`out:= &`, b.typeName(outType.GoIdent, g), `{}`)
	b.spanResultHandling(service, g)
	g.P(`return out, nil`)
	g.P(`}`)
}

func (b *ORMBuilder) spanResultHandling(service autogenService, g *protogen.GeneratedFile) {
	withSpan := getServiceOptions(service.Service).WithTracing
	if withSpan {
		g.P(`errSpanResult := m.spanResult(span, out)`)
		g.P(`if errSpanResult != nil {`)
		g.P(`return nil, `, b.wrapSpanError(service, "errSpanResult"))
		g.P(`}`)
	}
}

func (b *ORMBuilder) wrapSpanError(service autogenService, errVarName string) string {
	withSpan := getServiceOptions(service.Service).WithTracing
	if withSpan {
		return fmt.Sprint(`m.spanError(span, `, errVarName, `)`)
	}
	return errVarName
}

func (b *ORMBuilder) generateDBSetup(service autogenService, g *protogen.GeneratedFile) error {
	if service.usesTxnMiddleware {
		g.P(`txn, ok := `, generateImport("FromContext", tkgormImport, g), `(ctx)`)
		g.P(`if !ok {`)
		g.P(`return nil, `, generateImport("NoTransactionError", gerrorsImport, g))
		g.P(`}`)
		g.P(`db := txn.Begin()`)
		g.P(`if db.Error != nil {`)
		g.P(`return nil, db.Error`)
		g.P(`}`)
	} else {
		g.P(`db := m.DB`)
	}
	return nil
}

func (b *ORMBuilder) generatePreserviceCall(service autogenService, typeName, method string, g *protogen.GeneratedFile) {
	g.P(`if custom, ok := interface{}(in).(`, service.ccName, typeName, `WithBefore`, method, `); ok {`)
	g.P(`var err error`)
	g.P(`if db, err = custom.Before`, method, `(ctx, db); err != nil {`)
	g.P(`return nil, `, b.wrapSpanError(service, "err"))
	g.P(`}`)
	g.P(`}`)
}

func (b *ORMBuilder) generatePostserviceCall(service autogenService, typeName, method string, g *protogen.GeneratedFile) {
	g.P(`if custom, ok := interface{}(in).(`, service.ccName, typeName, `WithAfter`, method, `); ok {`)
	g.P(`var err error`)
	g.P(`if err = custom.After`, method, `(ctx, out, db); err != nil {`)
	g.P(`return nil, `, b.wrapSpanError(service, "err"))
	g.P(`}`)
	g.P(`}`)
}

func (b *ORMBuilder) generatePreserviceHook(svc, typeName, method string, g *protogen.GeneratedFile) {
	g.P(`// `, svc, typeName, `WithBefore`, method, ` called before Default`, method, typeName, ` in the default `, method, ` handler`)
	g.P(`type `, svc, typeName, `WithBefore`, method, ` interface {`)
	g.P(`Before`, method, `(context.Context, *`, generateImport("DB", gormImport, g), `) (*`, generateImport("DB", gormImport, g), `, error)`)
	g.P(`}`)
}

func (b *ORMBuilder) generatePostserviceHook(svc, typeName, outTypeName, method string, g *protogen.GeneratedFile) {
	g.P(`// `, svc, typeName, `WithAfter`, method, ` called before Default`, method, typeName, ` in the default `, method, ` handler`)
	g.P(`type `, svc, typeName, `WithAfter`, method, ` interface {`)
	g.P(`After`, method, `(context.Context, *`, outTypeName, `, *`, generateImport("DB", gormImport, g), `) error`)
	g.P(`}`)
}

func (b *ORMBuilder) generateReadServerMethod(service autogenService, method autogenMethod, g *protogen.GeneratedFile) {
	b.generateMethodSignature(service, method, g)
	if method.followsConvention {
		b.generateDBSetup(service, g)
		b.generatePreserviceCall(service, method.baseType, method.ccName, g)
		typeName := method.baseType
		getIDFormatter := "{Id: in.GetId()},"
		if b.IsIDFieldOptional(method.inType) {
			getIDFormatter = "{Id: in.Id},"
		}
		if fields := b.getFieldSelection(method.inType); fields != "" {
			g.P(`res, err := DefaultRead`, typeName, `(ctx, &`, typeName, getIDFormatter, `db, in.`, fields, `)`)
		} else {
			g.P(`res, err := DefaultRead`, typeName, `(ctx, &`, typeName, getIDFormatter, ` db)`)
		}
		g.P(`if err != nil {`)
		g.P(`return nil, `, b.wrapSpanError(service, "err"))
		g.P(`}`)
		g.P(`out := &`, b.typeName(method.outType.GoIdent, g), `{Result: res}`)
		b.generatePostserviceCall(service, method.baseType, method.ccName, g)
		b.spanResultHandling(service, g)
		g.P(`return out, nil`)
		g.P(`}`)
		b.generatePreserviceHook(service.ccName, method.baseType, method.ccName, g)
		b.generatePostserviceHook(service.ccName, method.baseType, b.typeName(method.outType.GoIdent, g), method.ccName, g)
	} else {
		b.generateEmptyBody(service, method.outType, g)
	}
}

func (b *ORMBuilder) getFieldSelection(message *protogen.Message) string {
	return b.getFieldOfType(message, "FieldSelection")
}

func (b *ORMBuilder) getFiltering(message *protogen.Message) string {
	return b.getFieldOfType(message, "Filtering")
}

func (b *ORMBuilder) getSorting(message *protogen.Message) string {
	return b.getFieldOfType(message, "Sorting")
}

func (b *ORMBuilder) getPagination(message *protogen.Message) string {
	return b.getFieldOfType(message, "Pagination")
}

func (b *ORMBuilder) getPageInfo(message *protogen.Message) string {
	return b.getFieldOfType(message, "PageInfo")
}

func (b *ORMBuilder) getFieldOfType(message *protogen.Message, fieldType string) string {
	for _, field := range message.Fields {
		if field.Desc.Message() != nil {
			goFieldName := camelCase(field.GoName)
			goFieldType := string(field.Desc.Message().FullName())
			// FullName is here because split really on it, but i don't see any
			// evidence that this is necessary
			parts := strings.Split(goFieldType, ".")
			if parts[len(parts)-1] == fieldType {
				return goFieldName
			}
		}
	}
	return ""
}

func (b *ORMBuilder) generateUpdateServerMethod(service autogenService, method autogenMethod, g *protogen.GeneratedFile) {
	b.generateMethodSignature(service, method, g)
	if method.followsConvention {
		g.P(`var err error`)
		typeName := method.baseType
		g.P(`var res *`, typeName)
		b.generateDBSetup(service, g)
		b.generatePreserviceCall(service, method.baseType, method.ccName, g)
		if method.fieldMaskName != "" {
			g.P(`if in.Get`, method.fieldMaskName, `() == nil {`)
			g.P(`res, err = DefaultStrictUpdate`, typeName, `(ctx, in.GetPayload(), db)`)
			g.P(`} else {`)
			g.P(`res, err = DefaultPatch`, typeName, `(ctx, in.GetPayload(), in.Get`, method.fieldMaskName, `(), db)`)
			g.P(`}`)
		} else {
			g.P(`res, err = DefaultStrictUpdate`, typeName, `(ctx, in.GetPayload(), db)`)
		}
		g.P(`if err != nil {`)
		g.P(`return nil, `, b.wrapSpanError(service, "err"))
		g.P(`}`)
		g.P(`out := &`, b.typeName(method.outType.GoIdent, g), `{Result: res}`)
		b.generatePostserviceCall(service, method.baseType, method.ccName, g)
		b.spanResultHandling(service, g)
		g.P(`return out, nil`)
		g.P(`}`)
		b.generatePreserviceHook(service.ccName, method.baseType, method.ccName, g)
		b.generatePostserviceHook(service.ccName, method.baseType, b.typeName(method.outType.GoIdent, g), method.ccName, g)
	} else {
		b.generateEmptyBody(service, method.outType, g)
	}
}

func (b *ORMBuilder) generateUpdateSetServerMethod(service autogenService, method autogenMethod, g *protogen.GeneratedFile) {
	b.generateMethodSignature(service, method, g)
	if method.followsConvention {
		typeName := method.baseType
		typeName = strings.TrimPrefix(typeName, "[]*")
		g.P(`if in == nil {`)
		g.P(`return nil,`, generateImport("NilArgumentError", gerrorsImport, g))
		g.P(`}`)
		g.P(``)
		b.generateDBSetup(service, g)
		g.P(``)
		b.generatePreserviceCall(service, typeName, method.ccName, g)

		g.P(``)
		g.P(`res, err := DefaultPatchSet`, typeName, `(ctx, in.GetObjects(), in.Get`, method.fieldMaskName, `(), db)`)
		g.P(`if err != nil {`)
		g.P(`return nil, `, b.wrapSpanError(service, "err"))
		g.P(`}`)
		g.P(``)
		g.P(`out := &`, b.typeName(method.outType.GoIdent, g), `{Results: res}`)

		g.P(``)
		b.generatePostserviceCall(service, typeName, method.ccName, g)
		g.P(``)
		withSpan := getServiceOptions(service.Service).WithTracing
		if withSpan {
			g.P(`err = m.spanResult(span, out)`)
			g.P(`if err != nil {`)
			g.P(`return nil,`, b.wrapSpanError(service, "err"))
			g.P(`}`)
		}
		g.P(`return out, nil`)
		g.P(`}`)

		b.generatePreserviceHook(service.ccName, typeName, method.ccName, g)
		b.generatePostserviceHook(service.ccName, typeName, b.typeName(method.outType.GoIdent, g), method.ccName, g)
	} else {
		b.generateEmptyBody(service, method.outType, g)
	}
}

func (b *ORMBuilder) generateDeleteServerMethod(service autogenService, method autogenMethod, g *protogen.GeneratedFile) {
	b.generateMethodSignature(service, method, g)
	if method.followsConvention {
		typeName := method.baseType
		b.generateDBSetup(service, g)
		b.generatePreserviceCall(service, method.baseType, method.ccName, g)
		getIDFormatter := "{Id: in.GetId()},"
		if b.IsIDFieldOptional(method.inType) {
			getIDFormatter = "{Id: in.Id},"
		}
		g.P(`err := DefaultDelete`, typeName, `(ctx, &`, typeName, getIDFormatter, ` db)`)
		g.P(`if err != nil {`)
		g.P(`return nil, `, b.wrapSpanError(service, "err"))
		g.P(`}`)
		g.P(`out := &`, b.typeName(method.outType.GoIdent, g), `{}`)
		b.generatePostserviceCall(service, method.baseType, method.ccName, g)
		b.spanResultHandling(service, g)
		g.P(`return out, nil`)
		g.P(`}`)
		b.generatePreserviceHook(service.ccName, method.baseType, method.ccName, g)
		b.generatePostserviceHook(service.ccName, method.baseType, b.typeName(method.outType.GoIdent, g), method.ccName, g)
	} else {
		b.generateEmptyBody(service, method.outType, g)
	}
}

func (b *ORMBuilder) generateDeleteSetServerMethod(service autogenService, method autogenMethod, g *protogen.GeneratedFile) {
	b.generateMethodSignature(service, method, g)
	if method.followsConvention {
		typeName := method.baseType
		b.generateDBSetup(service, g)
		g.P(`objs := []*`, typeName, `{}`)
		g.P(`for _, id := range in.Ids {`)
		g.P(`objs = append(objs, &`, typeName, `{Id: id})`)
		g.P(`}`)
		b.generatePreserviceCall(service, method.baseType, method.ccName, g)
		g.P(`err := DefaultDelete`, typeName, `Set(ctx, objs, db)`)
		g.P(`if err != nil {`)
		g.P(`return nil, `, b.wrapSpanError(service, "err"))
		g.P(`}`)
		g.P(`out := &`, b.typeName(method.outType.GoIdent, g), `{}`)
		b.generatePostserviceCall(service, method.baseType, method.ccName, g)
		b.spanResultHandling(service, g)
		g.P(`return out, nil`)
		g.P(`}`)
		b.generatePreserviceHook(service.ccName, method.baseType, method.ccName, g)
		b.generatePostserviceHook(service.ccName, method.baseType, b.typeName(method.outType.GoIdent, g), method.ccName, g)
	} else {
		b.generateEmptyBody(service, method.outType, g)
	}
}

func (b *ORMBuilder) generateListServerMethod(service autogenService, method autogenMethod, g *protogen.GeneratedFile) {
	b.generateMethodSignature(service, method, g)
	if method.followsConvention {
		b.generateDBSetup(service, g)
		ormable := b.getOrmable(method.baseType)
		b.generatePreserviceCall(service, method.baseType, method.ccName, g)
		pg := b.getPagination(method.inType)
		pi := b.getPageInfo(method.outType)
		if pg != "" && pi != "" {
			b.generatePagedRequestSetup(pg, g)
		}
		handlerCall := fmt.Sprint(`res, err := DefaultList`, method.baseType, `(ctx, db`)
		if f := b.getFiltering(method.inType); f != "" {
			handlerCall += fmt.Sprint(",in.", f)
		} else if b.listHasFiltering(ormable) {
			handlerCall += fmt.Sprint(", nil")
		}
		if s := b.getSorting(method.inType); s != "" {
			handlerCall += fmt.Sprint(",in.", s)
		} else if b.listHasSorting(ormable) {
			handlerCall += fmt.Sprint(", nil")
		}
		if pg != "" {
			handlerCall += fmt.Sprint(",in.", pg)
		} else if b.listHasPagination(ormable) {
			handlerCall += fmt.Sprint(", nil")
		}
		if fs := b.getFieldSelection(method.inType); fs != "" {
			handlerCall += fmt.Sprint(",in.", fs)
		} else if b.listHasFieldSelection(ormable) {
			handlerCall += fmt.Sprint(", nil")
		}
		handlerCall += ")"
		g.P(handlerCall)
		g.P(`if err != nil {`)
		g.P(`return nil, `, b.wrapSpanError(service, "err"))
		g.P(`}`)
		var pageInfoIfExist string
		if pg != "" && pi != "" {
			b.generatePagedRequestHandling(pg, g)
			pageInfoIfExist = ", " + pi + ": resPaging"
		}
		g.P(`out := &`, b.typeName(method.outType.GoIdent, g), `{Results: res`, pageInfoIfExist, ` }`)
		b.generatePostserviceCall(service, method.baseType, method.ccName, g)
		b.spanResultHandling(service, g)
		g.P(`return out, nil`)
		g.P(`}`)
		b.generatePreserviceHook(service.ccName, method.baseType, method.ccName, g)
		b.generatePostserviceHook(service.ccName, method.baseType, b.typeName(method.outType.GoIdent, g), method.ccName, g)
	} else {
		b.generateEmptyBody(service, method.outType, g)
	}
}

func (b *ORMBuilder) generatePagedRequestSetup(pg string, g *protogen.GeneratedFile) {
	g.P(`pagedRequest := false`)
	g.P(fmt.Sprintf(`if in.Get%s().GetLimit()>=1 {`, pg))
	g.P(fmt.Sprintf(`in.%s.Limit ++`, pg))
	g.P(`pagedRequest=true`)
	g.P(`}`)
}

func (b *ORMBuilder) generatePagedRequestHandling(pg string, g *protogen.GeneratedFile) {
	g.P(fmt.Sprintf(`var resPaging *%s`, generateImport("PageInfo", queryImport, g)))
	g.P(`if pagedRequest {`)
	g.P(`var offset int32`)
	g.P(`var size int32 = int32(len(res))`)
	g.P(fmt.Sprintf(`if size == in.Get%s().GetLimit(){`, pg))
	g.P(`size--`)
	g.P(`res=res[:size]`)
	g.P(fmt.Sprintf(`offset=in.Get%s().GetOffset()+size`, pg))
	g.P(`}`)
	g.P(fmt.Sprintf(`resPaging = &%s{Offset: offset}`, generateImport("PageInfo", queryImport, g)))
	g.P(`}`)
}

func (b *ORMBuilder) generateMethodStub(service autogenService, method autogenMethod, g *protogen.GeneratedFile) {
	b.generateMethodSignature(service, method, g)
	b.generateEmptyBody(service, method.outType, g)
}

func (b *ORMBuilder) typeName(ident protogen.GoIdent, g *protogen.GeneratedFile) string {
	// drop package prefix, no need to import
	if b.currentPackage == ident.GoImportPath.String() {
		return ident.GoName
	}

	return generateImport(ident.GoName, string(ident.GoImportPath), g)
}

func GetOrmable(ormableTypes map[string]*OrmableType, typeName string) (*OrmableType, error) {
	parts := strings.Split(typeName, ".")
	ormable, ok := ormableTypes[strings.TrimSuffix(strings.Trim(parts[len(parts)-1], "[]*"), "ORM")]
	var err error
	if !ok {
		err = ErrNotOrmable
	}
	return ormable, err
}

func (b *ORMBuilder) countHasAssociationDimension(message *protogen.Message, typeName string) int {
	dim := 0
	for _, field := range message.Fields {
		options := field.Desc.Options().(*descriptorpb.FieldOptions)
		fieldOpts := getFieldOptions(options)
		if fieldOpts.GetDrop() {
			continue
		}

		var fieldType string
		if field.Desc.Message() == nil {
			fieldType = field.Desc.Kind().String() // was GoType
		} else {
			fieldType = string(field.Desc.Message().Name())
		}

		if fieldOpts.GetManyToMany() == nil && fieldOpts.GetBelongsTo() == nil {
			if strings.Trim(typeName, "[]*") == strings.Trim(fieldType, "[]*") {
				dim++
			}
		}
	}

	return dim
}

func (b *ORMBuilder) countBelongsToAssociationDimension(message *protogen.Message, typeName string) int {
	dim := 0
	for _, field := range message.Fields {
		options := field.Desc.Options().(*descriptorpb.FieldOptions)
		fieldOpts := getFieldOptions(options)
		if fieldOpts.GetDrop() {
			continue
		}

		var fieldType string
		if field.Desc.Message() == nil {
			fieldType = field.Desc.Kind().String() // was GoType
		} else {
			fieldType = string(field.Desc.Message().Name())
		}

		if fieldOpts.GetBelongsTo() != nil {
			if strings.Trim(typeName, "[]*") == strings.Trim(fieldType, "[]*") {
				dim++
			}
		}
	}

	return dim
}

func (b *ORMBuilder) countManyToManyAssociationDimension(message *protogen.Message, typeName string) int {
	dim := 0

	for _, field := range message.Fields {
		options := field.Desc.Options().(*descriptorpb.FieldOptions)
		fieldOpts := getFieldOptions(options)

		if fieldOpts.GetDrop() {
			continue
		}
		var fieldType string

		if field.Desc.Message() == nil {
			fieldType = field.Desc.Kind().String() // was GoType
		} else {
			fieldType = string(field.Desc.Message().Name())
		}

		if fieldOpts.GetManyToMany() != nil {
			if strings.Trim(typeName, "[]*") == strings.Trim(fieldType, "[]*") {
				dim++
			}
		}
	}

	return dim
}

func (b *ORMBuilder) sameType(field1 *Field, field2 *Field) bool {
	isPointer1 := strings.HasPrefix(field1.TypeName, "*")
	typeParts1 := strings.Split(field1.TypeName, ".")

	if len(typeParts1) == 2 {
		isPointer2 := strings.HasPrefix(field2.TypeName, "*")
		typeParts2 := strings.Split(field2.TypeName, ".")

		if len(typeParts2) == 2 && isPointer1 == isPointer2 && typeParts1[1] == typeParts2[1] && field1.Package == field2.Package {
			return true
		}

		return false
	}

	return field1.TypeName == field2.TypeName
}

func getFieldType(field *protogen.Field) string {
	if field.Desc.Message() == nil {
		return field.Desc.Kind().String()
	}

	return string(field.Desc.Message().Name())
}

func getFieldIdent(field *protogen.Field) protogen.GoIdent {
	if field.Desc.Message() != nil {
		return field.Message.GoIdent
	}

	return field.GoIdent
}
