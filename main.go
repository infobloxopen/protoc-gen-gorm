package main

import (
	"github.com/gogo/protobuf/vanity/command"
	"github.com/infobloxopen/protoc-gen-gorm/plugin"

	_ "github.com/golang/protobuf/protoc-gen-go/descriptor"
)

func main() {
	plugin := &plugin.OrmPlugin{}
	response := command.GeneratePlugin(command.Read(), plugin, ".pb.gorm.go")
	plugin.CleanFiles(response)
	command.Write(response)
}
