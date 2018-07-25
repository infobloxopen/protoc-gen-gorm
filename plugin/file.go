package plugin

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	plugin "github.com/gogo/protobuf/protoc-gen-gogo/plugin"
)

var ProtocGenGormVersion string
var AtlasAppToolkitVersion string

// CleanFiles will prevent generation of any files where no real code is generated
func (p *OrmPlugin) CleanFiles(response *plugin.CodeGeneratorResponse) {

	for i := 0; i < len(response.File); i++ {
		file := response.File[i]
		for _, skip := range p.EmptyFiles {
			if strings.Contains(file.GetName(), strings.Trim(skip, ".proto")) {
				response.File = append(response.File[:i], response.File[i+1:]...)
				i--
				break
			}
		}
		file.Content = CleanImports(file.Content)

		sections := strings.SplitAfterN(file.GetContent(), "\n", 3)
		versionString := ""
		if ProtocGenGormVersion != "" {
			versionString = fmt.Sprintf("\n// Generated with protoc-gen-gorm version: %s\n", ProtocGenGormVersion)
		}
		if AtlasAppToolkitVersion != "" {
			versionString = fmt.Sprintf("%s// Anticipating compatibility with atlas-app-toolkit version: %s\n", versionString, AtlasAppToolkitVersion)
		}
		file.Content = proto.String(fmt.Sprintf("%s%s%s%s", sections[0], sections[1], versionString, sections[2]))
	}
}
