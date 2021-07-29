package pluginv2

import (
	"fmt"
	gorm "github.com/infobloxopen/protoc-gen-gorm/options"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
	"os"
)

type ORMBuilder struct {
	plugin *protogen.Plugin
}

type OrmableType struct {
	Name string
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
			fmt.Fprintf(os.Stderr, "%s -> %t\n", message.Desc.Name(), isOrmable(message))
		}
	}

	return b.plugin.Response(), nil
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
