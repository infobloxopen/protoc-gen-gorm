package plugin

import (
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	gorm "github.com/infobloxopen/protoc-gen-gorm/options"
)

func (p *OrmPlugin) getMsgName(o generator.Object) string {
	fqTypeName := p.TypeName(o)
	a := strings.Split(fqTypeName, ".")
	return a[len(a)-1]
}

// retrieves the GormMessageOptions from a message
func getMessageOptions(message *generator.Descriptor) *gorm.GormMessageOptions {
	if message.Options == nil {
		return nil
	}
	v, err := proto.GetExtension(message.Options, gorm.E_Opts)
	if err != nil {
		return nil
	}
	opts, ok := v.(*gorm.GormMessageOptions)
	if !ok {
		return nil
	}
	return opts
}

func getFieldOptions(field *descriptor.FieldDescriptorProto) *gorm.GormFieldOptions {
	if field.Options == nil {
		return nil
	}
	v, err := proto.GetExtension(field.Options, gorm.E_Field)
	if err != nil {
		return nil
	}
	opts, ok := v.(*gorm.GormFieldOptions)
	if !ok {
		return nil
	}
	return opts
}

func getServiceOptions(service *descriptor.ServiceDescriptorProto) *gorm.AutoServerOptions {
	if service.Options == nil {
		return nil
	}
	v, err := proto.GetExtension(service.Options, gorm.E_Server)
	if err != nil {
		return nil
	}
	opts, ok := v.(*gorm.AutoServerOptions)
	if !ok {
		return nil
	}
	return opts
}
