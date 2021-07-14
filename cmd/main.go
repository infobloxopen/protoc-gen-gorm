package main

import (
	"flag"
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

	var flags flag.FlagSet
	engine := flags.String("engine", "", "sql engine, only postgres supported")
	enums := flags.Bool("enums", true, "treat enums as strings instead of ints")
	gateway := flags.Bool("gateway", false, "")
	quiet := flags.Bool("quiet", false, "suppress warnings")

	flag.Parse()
	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			gorm.GenerateFile(gen, f, gorm.Params{
				Engine:  *engine,
				Enums:   *enums,
				Gateway: *gateway,
				Quiet:   *quiet,
			})
		}
		return nil
	})
}
