package plugin

import (
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	gorm "github.com/infobloxopen/protoc-gen-gorm/options"
)

func (p *OrmPlugin) generateDefaultServer(file *generator.FileDescriptor) {
	for _, service := range file.GetService() {
		svcName := generator.CamelCase(service.GetName())
		if service.GetOptions() != nil {
			v, err := proto.GetExtension(service.GetOptions(), gorm.E_Server)
			if err != nil {
				continue
			}
			opts := v.(*gorm.AutoServerOptions)
			if opts.GetAutogen() {
				// All the default server has is a db connection
				p.P(`type `, svcName, `DefaultServer struct {`)
				p.P(`DB *`, p.gormPkgName, `.DB`)
				p.P(`}`)
				for _, method := range service.GetMethod() {
					methodName := generator.CamelCase(method.GetName())
					if strings.HasPrefix(methodName, "Create") {
						p.generateCreateServerMethod(service, method)
					} else if strings.HasPrefix(methodName, "Read") {
						p.generateReadServerMethod(service, method)
					} else if strings.HasPrefix(methodName, "Update") {
						p.generateUpdateServerMethod(service, method)
					} else if strings.HasPrefix(methodName, "Delete") {
						p.generateDeleteServerMethod(service, method)
					} else if strings.HasPrefix(methodName, "List") {
						p.generateListServerMethod(service, method)
					} else {
						p.generateMethodStub(service, method)
					}
				}
			}
		}
	}
}

func (p *OrmPlugin) generateCreateServerMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	follows, typeName := p.followsCreateConventions(inType, outType)
	if follows {
		p.P(`res, err := DefaultCreate`, typeName, `(ctx, in.GetPayload(), m.DB)`)
		p.P(`if err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
		p.P(`return &`, p.TypeName(outType), `{Result: res}, nil`)
		p.P(`}`)
	} else {
		p.generateEmptyBody(outType)
	}
}

func (p *OrmPlugin) followsCreateConventions(inType generator.Object, outType generator.Object) (bool, string) {
	inMsg := inType.(*generator.Descriptor)
	outMsg := outType.(*generator.Descriptor)
	var inTypeName string
	var typeOrmable bool
	for _, field := range inMsg.Field {
		if field.GetName() == "payload" {
			gType, _ := p.GoType(inMsg, field)
			inTypeName = strings.TrimPrefix(gType, "*")
			if _, exists := convertibleTypes[inTypeName]; exists {
				typeOrmable = true
			}
		}
	}
	var outTypeName string
	for _, field := range outMsg.Field {
		if field.GetName() == "result" {
			gType, _ := p.GoType(outMsg, field)
			outTypeName = strings.TrimPrefix(gType, "*")
		}
	}
	if inTypeName == outTypeName && typeOrmable {
		return true, lintName(inTypeName)
	}
	return false, ""
}

func (p *OrmPlugin) generateReadServerMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	follows, pbTypeName, ormTypeName := p.followsReadConventions(inType, outType)
	if follows {
		p.P(`res, err := DefaultRead`, ormTypeName, `(ctx, &`, pbTypeName, `{Id: in.GetId()}, m.DB)`)
		p.P(`if err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
		p.P(`return &`, p.TypeName(outType), `{Result: res}, nil`)
		p.P(`}`)
	} else {
		p.generateEmptyBody(outType)
	}
}

func (p *OrmPlugin) followsReadConventions(inType generator.Object, outType generator.Object) (bool, string, string) {
	inMsg := inType.(*generator.Descriptor)
	outMsg := outType.(*generator.Descriptor)
	var hasID bool
	for _, field := range inMsg.Field {
		if field.GetName() == "id" {
			hasID = true
		}
	}
	var outTypeName string
	var typeOrmable bool
	for _, field := range outMsg.Field {
		if field.GetName() == "result" {
			gType, _ := p.GoType(inMsg, field)
			outTypeName = strings.TrimPrefix(gType, "*")
			if _, exists := convertibleTypes[outTypeName]; exists {
				typeOrmable = true
			}
		}
	}
	if hasID && typeOrmable {
		return true, outTypeName, lintName(outTypeName)
	}
	return false, "", ""
}

func (p *OrmPlugin) generateUpdateServerMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	follows, typeName := p.followsUpdateConventions(inType, outType)
	if follows {
		p.P(`res, err := DefaultUpdate`, typeName, `(ctx, in.GetPayload(), m.DB)`)
		p.P(`if err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
		p.P(`return &`, p.TypeName(outType), `{Result: res}, nil`)
		p.P(`}`)
	} else {
		p.generateEmptyBody(outType)
	}
}

func (p *OrmPlugin) followsUpdateConventions(inType generator.Object, outType generator.Object) (bool, string) {
	inMsg := inType.(*generator.Descriptor)
	outMsg := outType.(*generator.Descriptor)
	var inTypeName string
	var typeOrmable bool
	for _, field := range inMsg.Field {
		if field.GetName() == "payload" {
			gType, _ := p.GoType(inMsg, field)
			inTypeName = strings.TrimPrefix(gType, "*")
			if _, exists := convertibleTypes[inTypeName]; exists {
				typeOrmable = true
			}
		}
	}
	var outTypeName string
	for _, field := range outMsg.Field {
		if field.GetName() == "result" {
			gType, _ := p.GoType(outMsg, field)
			outTypeName = strings.TrimPrefix(gType, "*")
		}
	}
	if inTypeName == outTypeName && typeOrmable {
		return true, lintName(inTypeName)
	}
	return false, ""
}

func (p *OrmPlugin) generateDeleteServerMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	follows, pbTypeName, ormTypeName := p.followsDeleteConventions(inType, outType, method)
	if follows {
		p.P(`return &`, p.TypeName(outType), `{}, `, `DefaultDelete`, ormTypeName, `(ctx, &`, pbTypeName, `{Id: in.GetId()}, m.DB)`)
		p.P(`}`)
	} else {
		p.generateEmptyBody(outType)
	}
}

func (p *OrmPlugin) followsDeleteConventions(inType generator.Object, outType generator.Object, method *descriptor.MethodDescriptorProto) (bool, string, string) {
	inMsg := inType.(*generator.Descriptor)
	var hasID bool
	for _, field := range inMsg.Field {
		if field.GetName() == "id" {
			hasID = true
		}
	}
	var typeName string
	if method.GetOptions() != nil {
		v, err := proto.GetExtension(method.GetOptions(), gorm.E_Method)
		if err != nil {
			return false, "", ""
		}
		opts := v.(*gorm.MethodOptions)
		typeName = generator.CamelCase(opts.GetObjectType())
	}
	var typeOrmable bool
	if _, exists := convertibleTypes[typeName]; exists {
		typeOrmable = true
	}
	if hasID && typeOrmable {
		return true, typeName, lintName(typeName)
	}
	return false, "", ""
}

func (p *OrmPlugin) generateListServerMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	follows, typeName := p.followsListConventions(inType, outType)
	if follows {
		p.P(`res, err := DefaultList`, typeName, `(ctx, m.DB)`)
		p.P(`if err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
		p.P(`return &`, p.TypeName(outType), `{Results: res}, nil`)
		p.P(`}`)
	} else {
		p.generateEmptyBody(outType)
	}
}

func (p *OrmPlugin) followsListConventions(inType generator.Object, outType generator.Object) (bool, string) {
	outMsg := outType.(*generator.Descriptor)
	var outTypeName string
	var typeOrmable bool
	for _, field := range outMsg.Field {
		if field.GetName() == "results" {
			gType, _ := p.GoType(outMsg, field)
			outTypeName = strings.TrimPrefix(gType, "[]*")
			if _, exists := convertibleTypes[outTypeName]; exists {
				typeOrmable = true
			}
		}
	}
	if typeOrmable {
		return true, lintName(outTypeName)
	}
	return false, ""
}

func (p *OrmPlugin) generateMethodStub(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	p.generateEmptyBody(outType)
}

func (p *OrmPlugin) generateMethodSignature(inType, outType generator.Object, methodName, svcName string) {
	p.P(`// `, methodName, ` ...`)
	p.P(`func (m *`, svcName, `DefaultServer) `, methodName, ` (ctx context.Context, in *`,
		p.TypeName(inType), `) (*`, p.TypeName(outType), `, error) {`)
}

func (p OrmPlugin) generateEmptyBody(outType generator.Object) {
	p.P(`return &`, p.TypeName(outType), `{}, nil`)
	p.P(`}`)
}

func (p *OrmPlugin) getMethodProps(service *descriptor.ServiceDescriptorProto,
	method *descriptor.MethodDescriptorProto) (generator.Object, generator.Object, string, string) {
	inType := p.ObjectNamed(method.GetInputType())
	p.RecordTypeUse(method.GetInputType())
	outType := p.ObjectNamed(method.GetOutputType())
	p.RecordTypeUse(method.GetOutputType())
	methodName := generator.CamelCase(method.GetName())
	svcName := generator.CamelCase(service.GetName())
	return inType, outType, methodName, svcName
}
