package plugin

import (
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	gorm "github.com/infobloxopen/protoc-gen-gorm/options"
)

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
