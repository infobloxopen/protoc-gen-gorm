package main

import (
	"github.com/gogo/protobuf/vanity/command"
	"github.com/YReshetko/protoc-gen-gorm/plugin"
)

func main() {
	plugin := &plugin.OrmPlugin{}
	response := command.GeneratePlugin(command.Read(), plugin, ".pb.gorm.go")
	plugin.CleanFiles(response)
	command.Write(response)
}
