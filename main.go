package main

import (
	"github.com/gogo/protobuf/vanity/command"
	"github.com/infobloxopen/protoc-gen-gorm/plugin"
)

func main() {
	response := command.GeneratePlugin(command.Read(), &plugin.OrmPlugin{}, ".pb.gorm.go")
	for _, file := range response.GetFile() {
		file.Content = plugin.CleanImports(file.Content)
	}
	command.Write(response)
}
