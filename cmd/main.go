package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/infobloxopen/protoc-gen-gorm/internal/version"
	gorm "github.com/infobloxopen/protoc-gen-gorm/internal_gorm"
	"google.golang.org/protobuf/compiler/protogen"
)

const grpcDocURL = "https://pkg.go.dev/github.com/infobloxopen/protoc-gen-gorm"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Fprintf(os.Stdout, "%v %v\n", filepath.Base(os.Args[0]), version.String())
		os.Exit(0)
	}
	if len(os.Args) == 2 && os.Args[1] == "--help" {
		fmt.Fprintf(os.Stdout, "See "+grpcDocURL+" for usage information.\n")
		os.Exit(0)
	}

	// TODO: add options to plugins with this
	// var flags flag.FlagSet
	// font = flags.String("font", "doom", "font list available in github.com/common-nighthawk/go-figure")

	protogen.Options{
		// ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			gorm.GenerateFile(gen, f)
		}
		return nil
	})
}
