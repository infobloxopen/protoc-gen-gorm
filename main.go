package main

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

func main() {
	gen := generator.New()
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		gen.Error(err, "reading input")
	}
	if err = proto.Unmarshal(data, gen.Request); err != nil {
		gen.Error(err, "unmarshalling proto")
	}
	for _, x := range gen.Request.ProtoFile {
		for i := 0; i < len(x.GetDependency()); i++ {
			// Our files don't actually require this, so it's cleaner to drop it
			if x.GetDependency()[i] == "github.com/infobloxopen/protoc-gen-gorm/options/orm.proto" {
				x.Dependency = append(x.Dependency[:i], x.Dependency[i+1:]...)
				i--
			}
		}
	}
	gen.CommandLineParameters(gen.Request.GetParameter())

	gen.WrapTypes()
	gen.SetPackageNames()
	gen.BuildTypeNameMap()
	plug := &ormPlugin{}
	gen.GeneratePlugin(plug)

	for i := 0; i < len(gen.Response.File); i++ {
		// Rename file type
		gen.Response.File[i].Name = proto.String(strings.Replace(*gen.Response.File[i].Name, ".pb.go", ".pb.gorm.go", -1))
	}
	data, err = proto.Marshal(gen.Response)
	if err != nil {
		gen.Error(err, "failed to marshal output proto")
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		gen.Error(err, "failed to write output proto")
	}
}
