package main

import (
	"github.com/gogo/protobuf/vanity/command"
)

func main() {
	req := command.Read()
	for _, x := range req.ProtoFile {
		for i := 0; i < len(x.GetDependency()); i++ {
			// Our files don't actually require this, so it's cleaner to drop it
			if x.GetDependency()[i] == "github.com/infobloxopen/protoc-gen-gorm/options/gorm.proto" {
				x.Dependency = append(x.Dependency[:i], x.Dependency[i+1:]...)
				i--
			}
		}
	}
	command.Write(command.GeneratePlugin(req, &ormPlugin{}, ".pb.gorm.go"))
}
