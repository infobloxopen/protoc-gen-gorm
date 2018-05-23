package plugin

import (
	"strings"

	plugin "github.com/gogo/protobuf/protoc-gen-gogo/plugin"
)

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
	}
}
