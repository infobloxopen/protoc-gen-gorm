package main

import (
	"github.com/gogo/protobuf/vanity/command"
)

func main() {
	command.Write(command.GeneratePlugin(command.Read(), &ormPlugin{}, ".pb.gorm.go"))
}
