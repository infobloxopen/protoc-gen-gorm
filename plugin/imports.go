package plugin

import (
	"fmt"
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

var (
	gormImport         = "github.com/jinzhu/gorm"
	tkgormImport       = "github.com/infobloxopen/atlas-app-toolkit/gorm"
	uuidImport         = "github.com/satori/go.uuid"
	authImport         = "github.com/infobloxopen/atlas-app-toolkit/auth"
	gormpqImport       = "github.com/jinzhu/gorm/dialects/postgres"
	gtypesImport       = "github.com/infobloxopen/protoc-gen-gorm/types"
	ptypesImport       = "github.com/golang/protobuf/ptypes"
	wktImport          = "github.com/golang/protobuf/ptypes/wrappers"
	resourceImport     = "github.com/infobloxopen/atlas-app-toolkit/gorm/resource"
	fmImport           = "google.golang.org/genproto/protobuf/field_mask"
	queryImport        = "github.com/infobloxopen/atlas-app-toolkit/query"
	ocTraceImport      = "go.opencensus.io/trace"
	gatewayImport      = "github.com/infobloxopen/atlas-app-toolkit/gateway"
	pqImport           = "github.com/lib/pq"
	gerrorsImport      = "github.com/infobloxopen/protoc-gen-gorm/errors"
	stdFmtImport       = "fmt"
	stdCtxImport       = "context"
	stdStringsImport   = "strings"
	stdTimeImport      = "time"
	encodingJsonImport = "encoding/json"
)

type pkgImport struct {
	packagePath string
	alias       string
}

// Import takes a package and adds it to the list of packages to import
// It will generate a unique new alias using the last portion of the import path
// unless the package is already imported for this file. Either way, it returns
// the package alias
func (p *OrmPlugin) Import(packagePath string) string {
	subpath := packagePath[strings.LastIndex(packagePath, "/")+1:]
	// package will always be suffixed with an integer to prevent any collisions
	// with standard package imports
	for i := 1; ; i++ {
		newAlias := fmt.Sprintf("%s%d", strings.Replace(subpath, ".", "_", -1), i)
		if pkg, ok := p.GetFileImports().packages[newAlias]; ok {
			if packagePath == pkg.packagePath {
				return pkg.alias
			}
		} else {
			p.GetFileImports().packages[newAlias] = &pkgImport{packagePath: packagePath, alias: newAlias}
			return newAlias
		}
	}
	// Should never reach here
}

// UsingGoImports should be used with basic packages like "time", or "context"
func (p *OrmPlugin) UsingGoImports(pkgNames ...string) {
	p.GetFileImports().stdImports = append(p.GetFileImports().stdImports, pkgNames...)
}

type fileImports struct {
	wktPkgName      string
	typesToRegister []string
	stdImports      []string
	packages        map[string]*pkgImport
}

func newFileImports() *fileImports {
	return &fileImports{packages: make(map[string]*pkgImport)}
}

func (p *OrmPlugin) GetFileImports() *fileImports {
	return p.fileImports[p.currentFile]
}

// GenerateImports writes out required imports for the generated files
func (p *OrmPlugin) GenerateImports(file *generator.FileDescriptor) {
	imports := p.fileImports[file]
	for _, typeName := range imports.typesToRegister {
		p.RecordTypeUse(typeName)
	}
	githubImports := imports.packages
	sort.Strings(imports.stdImports)
	for _, dep := range imports.stdImports {
		p.PrintImport(dep, dep)
	}
	p.P()
	aliases := []string{}
	for a := range githubImports {
		aliases = append(aliases, a)
	}
	sort.Strings(aliases)
	for _, a := range aliases {
		p.PrintImport(a, githubImports[a].packagePath)
	}
	p.P()
}
