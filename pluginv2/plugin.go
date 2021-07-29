package pluginv2

import (
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
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
			if message.Desc.Name() == "BlogPost" {
				fmt.Fprintf(os.Stderr, "%s -> %t\n", message.Desc.Name(), isOrmable(message))
			}
		}
	}

	return b.plugin.Response(), nil
}

func isOrmable(message *protogen.Message) bool {
	//fmt.Fprintf(os.Stderr, "ext: %+v\n", message.
	option := message.Desc.Options()
	if option == nil {
		return false
	}

	fmt.Fprintf(os.Stderr, "option: %+v\n", option)
	for _, extension := range message.Extensions {
		fmt.Fprintf(os.Stderr, "ext: %+v\n", extension)
	}
	return false
}
