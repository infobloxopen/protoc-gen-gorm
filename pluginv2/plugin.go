package pluginv2

import (
	"fmt"
	"os"

	gorm "github.com/infobloxopen/protoc-gen-gorm/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

type ORMBuilder struct {
	plugin   *protogen.Plugin
	dbEngine int
}

type OrmableType struct {
	Name       string
	OriginName string
	Package    string
	File       *protogen.File
	Fields     map[string]*Field
	Methods    map[string]*autogenMethod
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
	ParentGoType   string
	Type           string
	Package        string
	ParentOrigName string
	*gorm.GormFieldOptions
}

type autogenMethod struct {
}

func New(opts protogen.Options, request *pluginpb.CodeGeneratorRequest) (*ORMBuilder, error) {
	plugin, err := opts.New(request)
	if err != nil {
		return nil, err
	}

	return &ORMBuilder{
		plugin: plugin,
	}, nil
}

func (b *ORMBuilder) Generate() (*pluginpb.CodeGeneratorResponse, error) {
	for _, protoFile := range b.plugin.Files {
		//fmt.Fprintf(os.Stderr, "%s\n", file.GeneratedFilenamePrefix)

		// 1. Collect all types that are ORMable
		for _, message := range protoFile.Messages {
			//fmt.Fprintf(os.Stderr, "%s -> %t\n", message.Desc.Name(), isOrmable(message))
			if isOrmable(message) {
				b.parseBasicField(message)
			}
		}
	}

	return b.plugin.Response(), nil
}

func (b *ORMBuilder) parseBasicField(msg *protogen.Message) {
	typeName := msg.Desc.Name()
	fmt.Fprintf(os.Stderr, "typeName: %s\n", typeName)
	if isOrmable(msg) {
		ormableName := fmt.Sprintf("%sORM", typeName)
		fmt.Fprintf(os.Stderr, "ormName: %s\n", ormableName)
	}

	for _, field := range msg.Fields {
		fd := field.Desc
		options := fd.Options().(*descriptorpb.FieldOptions)
		fmt.Fprintf(os.Stderr, "field options: %+v\n", options)

		// 1. get Field Options if not exists create default one
		gormOptions := getFieldOptions(options)
		if gormOptions == nil {
			gormOptions = &gorm.GormFieldOptions{}
		}

		// 2. skip Field if option have drop flag
		if gormOptions.GetDrop() {
			fmt.Fprintf(os.Stderr, "field options: %+v -> %t\n",
				gormOptions, gormOptions.GetDrop())
			continue
		}

		// 3. get field Tag
		tag := gormOptions.Tag
		fmt.Fprintf(os.Stderr, "field tag: %+v\n", tag)

		// 4. get fieldName and fieldType
		fieldName := fd.Name() // not CamelCase yet
		fieldType := fd.Kind()
		fmt.Fprintf(os.Stderr, "field name: %+v, type: %s\n", fieldName, fieldType)

		// next we need to know what kind of database engine we using
	}
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

func isOrmable(message *protogen.Message) bool {
	desc := message.Desc
	options := desc.Options()

	m, ok := proto.GetExtension(options, gorm.E_Opts).(*gorm.GormMessageOptions)
	if !ok || m == nil {
		return false
	}

	return m.Ormable
}
