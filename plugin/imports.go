package plugin

import (
	"sort"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

/* --------- Response file import cleaning -------- */

// Imports that are added by default but unneeded in GORM code
var unneededImports = []string{
	"import proto \"github.com/gogo/protobuf/proto\"\n",
	"import _ \"github.com/infobloxopen/protoc-gen-gorm/options\"\n",
	// if needed will be imported with an alias
	"import _ \"github.com/infobloxopen/protoc-gen-gorm/types\"\n",
	"var _ = proto.Marshal\n",
}

// CleanImports removes extraneous imports and lines from a proto response
// file Content
func CleanImports(pFileText *string) *string {
	if pFileText == nil {
		return pFileText
	}
	fileText := *pFileText
	for _, dep := range unneededImports {
		fileText = strings.Replace(fileText, dep, "", -1)
	}
	return &fileText
}

/* --------- Plugin level import handling --------- */

type fileImports struct {
	wktPkgName      string
	usingGORM       bool
	usingUUID       bool
	usingTime       bool
	usingAuth       bool
	usingJSON       bool
	usingInet       bool
	typesToRegister []string
}

// GenerateImports writes out required imports for the generated files
func (p *OrmPlugin) GenerateImports(file *generator.FileDescriptor) {
	imports := p.fileImports[file]
	stdImports := []string{}
	githubImports := map[string]string{}
	if imports.usingGORM {
		stdImports = append(stdImports, "context", "errors")
		githubImports["gorm"] = "github.com/jinzhu/gorm"
		githubImports["tkgorm"] = "github.com/infobloxopen/atlas-app-toolkit/gorm"
	}
	if imports.usingUUID {
		githubImports["uuid"] = "github.com/satori/go.uuid"
		githubImports["gtypes"] = "github.com/infobloxopen/protoc-gen-gorm/types"
	}
	if imports.usingTime {
		stdImports = append(stdImports, "time")
		githubImports["ptypes"] = "github.com/golang/protobuf/ptypes"
	}
	if imports.usingAuth {
		githubImports["auth"] = "github.com/infobloxopen/atlas-app-toolkit/auth"
	}
	if imports.usingJSON {
		if p.dbEngine == ENGINE_POSTGRES {
			githubImports["gormpq"] = "github.com/jinzhu/gorm/dialects/postgres"
			githubImports["gtypes"] = "github.com/infobloxopen/protoc-gen-gorm/types"
		}
	}
	if imports.usingInet {
		githubImports["gtypes"] = "github.com/infobloxopen/protoc-gen-gorm/types"
	}
	for _, typeName := range imports.typesToRegister {
		p.RecordTypeUse(typeName)
	}
	sort.Strings(stdImports)
	for _, dep := range stdImports {
		p.PrintImport(dep, dep)
	}
	p.P()
	keys := []string{}
	for k := range githubImports {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		p.PrintImport(key, githubImports[key])
	}
	p.P()
}
