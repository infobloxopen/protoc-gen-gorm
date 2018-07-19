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
		if opts := getServiceOptions(service); opts != nil && opts.GetAutogen() {
			// All the default server has is a db connection
			p.P(`type `, svcName, `DefaultServer struct {`)
			if !opts.GetTxnMiddleware() {
				p.P(`DB *`, p.Import(gormImport), `.DB`)
			}
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

func (p *OrmPlugin) generateCreateServerMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	follows, typeName := p.followsCreateConventions(inType, outType)
	if follows {
		p.generateDBSetup(service, outType)
		p.generatePreserviceCall(svcName, typeName, "Create")
		p.P(`res, err := DefaultCreate`, typeName, `(ctx, in.GetPayload(), db)`)
		p.P(`if err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
		p.P(`return &`, p.TypeName(outType), `{Result: res}, nil`)
		p.P(`}`)
		p.generatePreserviceHook(svcName, typeName, p.TypeName(inType), "Create")
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
			if p.isOrmable(inTypeName) {
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
		return true, inTypeName
	}
	return false, ""
}

func (p *OrmPlugin) generateReadServerMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	follows, typeName := p.followsReadConventions(inType, outType)
	if follows {
		p.generateDBSetup(service, outType)
		p.generatePreserviceCall(svcName, typeName, "Read")
		p.P(`res, err := DefaultRead`, typeName, `(ctx, &`, typeName, `{Id: in.GetId()}, db)`)
		p.P(`if err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
		p.P(`return &`, p.TypeName(outType), `{Result: res}, nil`)
		p.P(`}`)
		p.generatePreserviceHook(svcName, typeName, p.TypeName(inType), "Read")
	} else {
		p.generateEmptyBody(outType)
	}
}

func (p *OrmPlugin) followsReadConventions(inType generator.Object, outType generator.Object) (bool, string) {
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
			if p.isOrmable(outTypeName) {
				typeOrmable = true
			}
		}
	}
	if hasID && typeOrmable && p.hasPrimaryKey(p.getOrmable(outTypeName)) {
		return true, outTypeName
	}
	return false, ""
}

func (p *OrmPlugin) generateUpdateServerMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	follows, typeName, updateMask := p.followsUpdateConventions(inType, outType)
	if follows {
		p.P(`var err error`)
		p.P(`var res *`, typeName)
		p.generateDBSetup(service, outType)
		p.generatePreserviceCall(svcName, typeName, "Update")
		if updateMask != "" {
			p.P(`if len(in.Get`, generator.CamelCase(updateMask),`().GetPaths()) == 0 {`)
			p.P(`res, err = DefaultStrictUpdate`, typeName, `(ctx, in.GetPayload(), db)`)
			p.P(`} else {`)
			p.P(`res, err = DefaultPatch`, typeName, `(ctx, in.GetPayload(), in.Get`, generator.CamelCase(updateMask), `(), db)`)
			p.P(`}`)
		} else {
			p.P(`res, err = DefaultStrictUpdate`, typeName, `(ctx, in.GetPayload(), db)`)
		}
		p.P(`if err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
		p.P(`return &`, p.TypeName(outType), `{Result: res}, nil`)
		p.P(`}`)
		p.generatePreserviceHook(svcName, typeName, p.TypeName(inType), "Update")
	} else {
		p.generateEmptyBody(outType)
	}
}

func (p *OrmPlugin) followsUpdateConventions(inType generator.Object, outType generator.Object) (bool, string, string) {
	inMsg := inType.(*generator.Descriptor)
	outMsg := outType.(*generator.Descriptor)
	var inTypeName string
	var typeOrmable bool
	var updateMask string
	for _, field := range inMsg.Field {
		if field.GetName() == "payload" {
			gType, _ := p.GoType(inMsg, field)
			inTypeName = strings.TrimPrefix(gType, "*")
			if p.isOrmable(inTypeName) {
				typeOrmable = true
			}
		}

		// Check that type of field is a FieldMask
		if field.GetTypeName() == ".google.protobuf.FieldMask" {
			// More than one mask in request is not allowed.
			if updateMask != "" {
				return false, "", ""
			}
			updateMask = field.GetName()
		}

	}
	var outTypeName string
	for _, field := range outMsg.Field {
		if field.GetName() == "result" {
			gType, _ := p.GoType(outMsg, field)
			outTypeName = strings.TrimPrefix(gType, "*")
		}
	}
	if inTypeName == outTypeName && typeOrmable && p.hasPrimaryKey(p.getOrmable(outTypeName)) {
		return true, inTypeName, updateMask
	}
	return false, "", ""
}

func (p *OrmPlugin) generateDeleteServerMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	follows, typeName := p.followsDeleteConventions(inType, outType, method)
	if follows {
		p.generateDBSetup(service, outType)
		p.generatePreserviceCall(svcName, typeName, "Delete")
		p.P(`return &`, p.TypeName(outType), `{}, `, `DefaultDelete`, typeName, `(ctx, &`, typeName, `{Id: in.GetId()}, db)`)
		p.P(`}`)
		p.generatePreserviceHook(svcName, typeName, p.TypeName(inType), "Delete")
	} else {
		p.generateEmptyBody(outType)
	}
}

func (p *OrmPlugin) followsDeleteConventions(inType generator.Object, outType generator.Object, method *descriptor.MethodDescriptorProto) (bool, string) {
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
			return false, ""
		}
		opts := v.(*gorm.MethodOptions)
		typeName = generator.CamelCase(opts.GetObjectType())
	}
	var typeOrmable bool
	if p.isOrmable(typeName) {
		typeOrmable = true
	}
	if hasID && typeOrmable && p.hasPrimaryKey(p.getOrmable(typeName)) {
		return true, typeName
	}
	return false, ""
}

func (p *OrmPlugin) generateListServerMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	inType, outType, methodName, svcName := p.getMethodProps(service, method)
	p.generateMethodSignature(inType, outType, methodName, svcName)
	follows, typeName := p.followsListConventions(inType, outType)
	if follows {
		p.generateDBSetup(service, outType)
		p.generatePreserviceCall(svcName, typeName, "List")
		p.P(`res, err := DefaultList`, typeName, `(ctx, db, in)`)
		p.P(`if err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
		p.P(`return &`, p.TypeName(outType), `{Results: res}, nil`)
		p.P(`}`)
		p.generatePreserviceHook(svcName, typeName, p.TypeName(inType), "List")
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
			if p.isOrmable(outTypeName) {
				typeOrmable = true
			}
		}
	}
	if typeOrmable {
		return true, outTypeName
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

func (p *OrmPlugin) generateDBSetup(service *descriptor.ServiceDescriptorProto, outType generator.Object) error {
	if opts := getServiceOptions(service); opts != nil && opts.GetTxnMiddleware() {
		p.P(`txn, ok := `, p.Import(tkgormImport), `.FromContext(ctx)`)
		p.P(`if !ok {`)
		p.P(`return nil, errors.New("Database Transaction For Request Missing")`)
		p.P(`}`)
		p.P(`db := txn.Begin()`)
		p.P(`if db.Error != nil {`)
		p.P(`return nil, db.Error`)
		p.P(`}`)
	} else {
		p.P(`db := m.DB`)
	}
	return nil
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

func (p *OrmPlugin) generatePreserviceCall(svc, typeName, mthd string) {
	p.P(`if custom, ok := interface{}(in).(`, svc, typeName, `WithBefore`, mthd, `); ok {`)
	p.P(`var err error`)
	p.P(`ctx, db, err = custom.Before`, mthd, `(ctx, in, db)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generatePreserviceHook(svc, typeName, inTypeName, mthd string) {
	p.P(`// `, svc, typeName, `WithBefore`, mthd, ` called before Default`, mthd, typeName, ` in the default `, mthd, ` handler`)
	p.P(`type `, svc, typeName, `WithBefore`, mthd, ` interface {`)
	p.P(`Before`, mthd, `(context.Context, *`, inTypeName, `, *`, p.Import(gormImport), `.DB) (context.Context, *`, p.Import(gormImport), `.DB, error)`)
	p.P(`}`)
}
