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
		svcName := lintName(generator.CamelCase(service.GetName()))
		if service.Options != nil {
			v, err := proto.GetExtension(service.GetOptions(), gorm.E_Server)
			opts := v.(*gorm.AutoServerOptions)
			if err == nil && opts != nil && *opts.Autogen {
				p.usingGRPC = true
				// All the default server has is a db connection
				p.P(`type `, svcName, `DefaultServer struct {`)
				p.P(`DB *`, p.gormPkgName, `.DB`)
				p.P(`}`)
				for _, method := range service.GetMethod() {
					methodName := lintName(generator.CamelCase(method.GetName()))
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
	p.P(`var out `, p.TypeName(outType))
	p.P(`res, err := DefaultCreate`, strings.TrimPrefix(methodName, "Create"), `(ctx, in.GetPayload(), m.DB)`)
	p.P(`out.Result = res`)
	p.P(`return &out, err`)
	p.P(`}`)
}

func (p *OrmPlugin) generateReadServerMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	p.P(`var out `, p.TypeName(outType))
	p.P(`res, err := DefaultRead`, strings.TrimPrefix(methodName, "Read"), `(ctx, in.GetPayload(), m.DB)`)
	p.P(`out.Result = res`)
	p.P(`return &out, err`)
	p.P(`}`)
}

func (p *OrmPlugin) generateUpdateServerMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	p.P(`var out `, p.TypeName(outType))
	p.P(`res, err := DefaultUpdate`, strings.TrimPrefix(methodName, "Update"), `(ctx, in.GetPayload(), m.DB)`)
	p.P(`out.Result = res`)
	p.P(`return &out, err`)
	p.P(`}`)
}

func (p *OrmPlugin) generateDeleteServerMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	p.P(`return nil, DefaultDelete`, strings.TrimPrefix(methodName, "Delete"), `(ctx, in.GetPayload(), m.DB)`)
	p.P(`}`)
}

func (p *OrmPlugin) generateListServerMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	p.P(`var out `, p.TypeName(outType))
	p.P(`res, err := DefaultList`, strings.TrimPrefix(methodName, "List"), `(ctx, m.DB)`)
	p.P(`out.Results = res`)
	p.P(`return &out, err`)
	p.P(`}`)
}

func (p *OrmPlugin) generateMethodStub(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	p.P(`return nil, nil`)
	p.P(`}`)
}

func (p *OrmPlugin) generateMethodSignature(inType, outType generator.Object, methodName, svcName string) {
	p.P(`// `, methodName, ` ...`)
	p.P(`func (m *`, svcName, `DefaultServer) `, methodName, ` (ctx context.Context, in *`,
		p.TypeName(inType), `, opts ...grpc.CallOption) (*`, p.TypeName(outType), `, error) {`)
}

func (p *OrmPlugin) getMethodProps(service *descriptor.ServiceDescriptorProto,
	method *descriptor.MethodDescriptorProto) (generator.Object, generator.Object, string, string) {
	inType := p.ObjectNamed(method.GetInputType())
	p.RecordTypeUse(method.GetInputType())
	outType := p.ObjectNamed(method.GetOutputType())
	p.RecordTypeUse(method.GetOutputType())
	methodName := lintName(generator.CamelCase(method.GetName()))
	svcName := lintName(generator.CamelCase(service.GetName()))
	return inType, outType, methodName, svcName
}
