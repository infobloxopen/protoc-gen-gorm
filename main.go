package main

import (
	"strings"

	"github.com/gogo/protobuf/vanity/command"
)

func main() {
	response := command.GeneratePlugin(command.Read(), &ormPlugin{}, ".pb.gorm.go")
	for _, file := range response.GetFile() {
		file.Content = cleanImports(file.Content)
	}
	command.Write(response)
}

// Imports that are added by default but unneeded in GORM code
var unneededImports = []string{
	"import proto \"github.com/gogo/protobuf/proto\"\n",
	"import _ \"github.com/infobloxopen/protoc-gen-gorm/options\"\n",
	// if needed will be imported with an alias
	"import _ \"github.com/infobloxopen/protoc-gen-gorm/types\"",
	"var _ = proto.Marshal\n",
}

func cleanImports(pFileText *string) *string {
	if pFileText == nil {
		return pFileText
	}
	fileText := *pFileText
	for _, dep := range unneededImports {
		fileText = strings.Replace(fileText, dep, "", -1)
	}
	return &fileText
}
