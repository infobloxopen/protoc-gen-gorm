package pluginv2

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	gorm "github.com/infobloxopen/protoc-gen-gorm/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
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

func NewOrmableType(orignalName string, pkg string, file *protogen.File) *OrmableType {
	return &OrmableType{
		Name:    orignalName,
		Package: pkg,
		File:    file,
		Fields:  make(map[string]*Field),
		Methods: make(map[string]*autogenMethod),
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
}

type fileImports struct {
	wktPkgName      string
	packages        map[string]*pkgImport
	typesToRegister []string
	stdImports      []string
}

func newFileImports() *fileImports {
	return &fileImports{packages: make(map[string]*pkgImport)}
}

type pkgImport struct {
	packagePath string
	alias       string
}

func (b *ORMBuilder) Generate() (*pluginpb.CodeGeneratorResponse, error) {
	for _, protoFile := range b.plugin.Files {
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
				b.parseBasicFields(message)
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

	// TODO: parse services
	// for _, protoFile := range b.plugin.Files {
	// 	fmt.Fprintf(os.Stderr, "TODO: generate services: %+v\n", protoFile)
	// }

	for _, protoFile := range b.plugin.Files {
		// generate actual code
		fileName := protoFile.GeneratedFilenamePrefix + ".gorm.go"
		genFile := b.plugin.NewGeneratedFile(fileName, ".")
		genFile.P("package ", protoFile.GoPackageName)

		for _, message := range protoFile.Messages {
			if isOrmable(message) {
				b.generateOrmable(genFile, message)
			}
		}
	}

	return b.plugin.Response(), nil
}

func (b *ORMBuilder) generateOrmable(genFile *protogen.GeneratedFile, message *protogen.Message) {
	ormable := b.getOrmable(message.GoIdent.GoName)
	genFile.P(`type `, ormable.Name, ` struct {`)

	for name, field := range ormable.Fields { // TODO: sorting, if it's required
		genFile.P(name, ` `, field.Type, b.renderGormTag(field))
	}

	genFile.P(`}`)
	genFile.P()
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
	// TODO: implement me
	return false
}

func (b *ORMBuilder) isOrmable(fieldType string) bool {
	// TODO: implement me
	return false
}

func (b *ORMBuilder) findPrimaryKey(ormable *OrmableType) (string, *Field) {
	// TODO: implement me
	return "", &Field{}
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

func (b *ORMBuilder) parseBasicFields(msg *protogen.Message) {
	typeName := string(msg.Desc.Name())
	fmt.Fprintf(os.Stderr, "parseBasicFields message Name: %s\n", typeName)

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
			fmt.Fprintf(os.Stderr, "droping field: %s, %+v -> %t\n",
				field.Desc.TextName(), gormOptions, gormOptions.GetDrop())
			continue
		}

		tag := gormOptions.Tag
		fieldName := camelCase(string(fd.Name())) // TODO: move to camelCase
		fieldType := fd.Kind().String()           // TODO: figure out GoType analog

		fmt.Fprintf(os.Stderr, "field name: %s, type: %s, tag: %+v\n",
			fieldName, fieldType, tag)

		var typePackage string

		if b.dbEngine == ENGINE_POSTGRES && b.IsAbleToMakePQArray(fieldType) {
			switch fieldType {
			case "[]bool":
				fieldType = fmt.Sprintf("%s.BoolArray", b.Import(pqImport))
				gormOptions.Tag = tagWithType(tag, "bool[]")
			case "[]float64":
				fieldType = fmt.Sprintf("%s.Float64Array", b.Import(pqImport))
				gormOptions.Tag = tagWithType(tag, "float[]")
			case "[]int64":
				fieldType = fmt.Sprintf("%s.Int64Array", b.Import(pqImport))
				gormOptions.Tag = tagWithType(tag, "integer[]")
			case "[]string":
				fieldType = fmt.Sprintf("%s.StringArray", b.Import(pqImport))
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

		fmt.Fprintf(os.Stderr, "detected field type is -> %s\n", fieldType)

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
			b.addIncludedField(ormable, field)
		} else {
			panic("cound not include")
		}
	}

	fmt.Fprintf(os.Stderr, "parseBasicFields end, ormable: %+v\n", ormable)
}

func (b *ORMBuilder) addIncludedField(ormable *OrmableType, field *gorm.ExtraField) {
	fieldName := field.GetName() // TODO: CamelCase
	isPtr := strings.HasPrefix(field.GetType(), "*")
	rawType := strings.TrimPrefix(field.GetType(), "*")
	// cut off any package subpaths
	rawType = rawType[strings.LastIndex(rawType, ".")+1:]
	var typePackage string
	// Handle types with a package defined
	if field.GetPackage() != "" {
		alias := b.Import(field.GetPackage())
		rawType = fmt.Sprintf("%s.%s", alias, rawType)
		typePackage = field.GetPackage()
	} else {
		// Handle types without a package defined
		if _, ok := builtinTypes[rawType]; ok {
			// basic type, 100% okay, no imports or changes needed
		} else if rawType == "Time" {
			// b.UsingGoImports(stdTimeImport) // TODO: missing UsingGoImports
			typePackage = stdTimeImport
			rawType = fmt.Sprintf("%s.Time", typePackage)
		} else if rawType == "UUID" {
			rawType = fmt.Sprintf("%s.UUID", b.Import(uuidImport))
			typePackage = uuidImport
		} else if field.GetType() == "Jsonb" && b.dbEngine == ENGINE_POSTGRES {
			rawType = fmt.Sprintf("%s.Jsonb", b.Import(gormpqImport))
			typePackage = gormpqImport
		} else if rawType == "Inet" {
			rawType = fmt.Sprintf("%s.Inet", b.Import(gtypesImport))
			typePackage = gtypesImport
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

func (b *ORMBuilder) Import(packagePath string) string {
	subpath := packagePath[strings.LastIndex(packagePath, "/")+1:]
	// package will always be suffixed with an integer to prevent any collisions
	// with standard package imports
	for i := 1; ; i++ {
		newAlias := fmt.Sprintf("%s%d", strings.Replace(subpath, ".", "_", -1), i)
		if pkg, ok := b.GetFileImports().packages[newAlias]; ok {
			if packagePath == pkg.packagePath {
				return pkg.alias
			}
		} else {
			b.GetFileImports().packages[newAlias] = &pkgImport{packagePath: packagePath, alias: newAlias}
			return newAlias
		}
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
