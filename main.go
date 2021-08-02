package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/infobloxopen/protoc-gen-gorm/pluginv2"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	var request pluginpb.CodeGeneratorRequest
	err = proto.Unmarshal(input, &request)
	if err != nil {
		panic(err)
	}

	opts := protogen.Options{}

	builder, err := pluginv2.New(opts, &request)
	if err != nil {
		panic(err)
	}

	response, err := builder.Generate()
	if err != nil {
		panic(err)
	}

	out, err := proto.Marshal(response)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(os.Stdout, string(out))
}
