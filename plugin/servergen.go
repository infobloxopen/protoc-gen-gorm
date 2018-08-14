package plugin

import (
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

const (
	createService = "Create"
	readService   = "Read"
	updateService = "Update"
	deleteService = "Delete"
	listService   = "List"
)

var ormableServices []autogenService

type autogenService struct {
	*descriptor.ServiceDescriptorProto
	ccName            string
	file              *generator.FileDescriptor
	usesTxnMiddleware bool
	methods           []autogenMethod
}

type autogenMethod struct {
	*descriptor.MethodDescriptorProto
	ccName            string
	verb              string
	followsConvention bool
	baseType          string
	inType            generator.Object
	outType           generator.Object
	fieldMaskName     string
}

func (p *OrmPlugin) parseServices(file *generator.FileDescriptor) {
	for _, service := range file.GetService() {
		if opts := getServiceOptions(service); opts != nil && opts.GetAutogen() {
			genSvc := autogenService{
				ServiceDescriptorProto: service,
				ccName:                 generator.CamelCase(service.GetName()),
				file:                   file,
			}
			if opts := getServiceOptions(service); opts != nil && opts.GetTxnMiddleware() {
				genSvc.usesTxnMiddleware = true
			}
			for _, method := range service.GetMethod() {
				inType, outType, methodName := p.getMethodProps(method)
				var verb, fmName, baseType string
				var follows bool
				if strings.HasPrefix(methodName, createService) {
					verb = createService
					follows, baseType = p.followsCreateConventions(inType, outType, methodName)
				} else if strings.HasPrefix(methodName, readService) {
					verb = readService
					follows, baseType = p.followsReadConventions(inType, outType, methodName)
				} else if strings.HasPrefix(methodName, updateService) {
					verb = updateService
					follows, baseType, fmName = p.followsUpdateConventions(inType, outType, methodName)
				} else if strings.HasPrefix(methodName, deleteService) {
					verb = deleteService
					follows, baseType = p.followsDeleteConventions(inType, outType, method)
				} else if strings.HasPrefix(methodName, listService) {
					verb = listService
					follows, baseType = p.followsListConventions(inType, outType, methodName)
				} else {
				}
				genMethod := autogenMethod{
					MethodDescriptorProto: method,
					ccName:                methodName,
					inType:                inType,
					outType:               outType,
					baseType:              baseType,
					fieldMaskName:         fmName,
					followsConvention:     follows,
					verb:                  verb,
				}
				genSvc.methods = append(genSvc.methods, genMethod)

				if genMethod.verb != "" && p.isOrmable(genMethod.baseType) {
					p.getOrmable(genMethod.baseType).Methods[genMethod.verb] = &genMethod
				}
			}
			ormableServices = append(ormableServices, genSvc)
		}
	}
}

func (p *OrmPlugin) generateDefaultServer(file *generator.FileDescriptor) {
	for _, service := range ormableServices {
		if service.file != file {
			continue
		}
		p.P(`type `, service.ccName, `DefaultServer struct {`)
		if !service.usesTxnMiddleware {
			p.P(`DB *`, p.Import(gormImport), `.DB`)
		}
		p.P(`}`)
		for _, method := range service.methods {
			switch method.verb {
			case createService:
				p.generateCreateServerMethod(service, method)
			case readService:
				p.generateReadServerMethod(service, method)
			case updateService:
				p.generateUpdateServerMethod(service, method)
			case deleteService:
				p.generateDeleteServerMethod(service, method)
			case listService:
				p.generateListServerMethod(service, method)
			default:
				p.generateMethodStub(service, method)
			}
		}
	}
}

func (p *OrmPlugin) generateCreateServerMethod(service autogenService, method autogenMethod) {
	p.generateMethodSignature(service, method)
	if method.followsConvention {
		p.generateDBSetup(service)
		p.generatePreserviceCall(service.ccName, method.baseType, createService)
		p.P(`res, err := DefaultCreate`, method.baseType, `(ctx, in.GetPayload(), db)`)
		p.P(`if err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
		p.P(`return &`, p.TypeName(method.outType), `{Result: res}, nil`)
		p.P(`}`)
		p.generatePreserviceHook(service.ccName, method.baseType, p.TypeName(method.inType), createService)
	} else {
		p.generateEmptyBody(method.outType)
	}
}

func (p *OrmPlugin) followsCreateConventions(inType generator.Object, outType generator.Object, methodName string) (bool, string) {
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
	if !typeOrmable {
		p.warning(`stub will be generated for %s since %s incoming message doesn't have "payload" field of ormable type`, methodName, p.TypeName(inType))
		return false, ""
	}
	var outTypeName string
	for _, field := range outMsg.Field {
		if field.GetName() == "result" {
			gType, _ := p.GoType(outMsg, field)
			outTypeName = strings.TrimPrefix(gType, "*")
		}
	}
	if inTypeName != outTypeName {
		p.warning(`stub will be generated for %s since "payload" field type of %s incoming message type doesn't match "result" field type of %s outcoming message`, methodName, p.TypeName(inType), p.TypeName(outType))
		return false, ""
	}
	return true, inTypeName
}

func (p *OrmPlugin) generateReadServerMethod(service autogenService, method autogenMethod) {
	p.generateMethodSignature(service, method)
	if method.followsConvention {
		p.generateDBSetup(service)
		p.generatePreserviceCall(service.ccName, method.baseType, readService)
		typeName := method.baseType
		if fields := p.hasFieldsSelector(method.inType); fields != "" {
			p.P(`var err error`)
			p.P(`if in.`, fields, ` == nil {`)
			p.generatePreloading()
			p.P(`} else if db, err = `, p.Import(tkgormImport), `.ApplyFieldSelection(ctx, db, in.`, fields, `, &`, typeName, `{}); err != nil {`)
			p.P(`return nil, err`)
			p.P(`}`)
			p.P(`res, err := DefaultRead`, typeName, `(ctx, &`, typeName, `{Id: in.GetId()}, db, false)`)
		} else {
			p.P(`res, err := DefaultRead`, typeName, `(ctx, &`, typeName, `{Id: in.GetId()}, db, true)`)
		}
		p.P(`if err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
		p.P(`return &`, p.TypeName(method.outType), `{Result: res}, nil`)
		p.P(`}`)
		p.generatePreserviceHook(service.ccName, method.baseType, p.TypeName(method.inType), readService)
	} else {
		p.generateEmptyBody(method.outType)
	}
}

func (p *OrmPlugin) followsReadConventions(inType generator.Object, outType generator.Object, methodName string) (bool, string) {
	inMsg := inType.(*generator.Descriptor)
	outMsg := outType.(*generator.Descriptor)
	var hasID bool
	for _, field := range inMsg.Field {
		if field.GetName() == "id" {
			hasID = true
		}
	}
	if !hasID {
		p.warning(`stub will be generated for %s since %s incoming message doesn't have "id" field`, methodName, p.TypeName(inType))
		return false, ""
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
	if !typeOrmable {
		p.warning(`stub will be generated for %s since %s outcoming message doesn't have "result" field of ormable type`, methodName, p.TypeName(outType))
		return false, ""
	}
	if !p.hasPrimaryKey(p.getOrmable(outTypeName)) {
		p.warning(`stub will be generated for %s since %s ormable type doesn't have a primary key`, methodName, outTypeName)
		return false, ""
	}
	return true, outTypeName
}

func (p *OrmPlugin) generateUpdateServerMethod(service autogenService, method autogenMethod) {
	p.generateMethodSignature(service, method)
	if method.followsConvention {
		p.P(`var err error`)
		typeName := method.baseType
		p.P(`var res *`, typeName)
		p.generateDBSetup(service)
		p.generatePreserviceCall(service.ccName, method.baseType, updateService)
		if method.fieldMaskName != "" {
			p.P(`if in.Get`, method.fieldMaskName, `() == nil {`)
			p.P(`res, err = DefaultStrictUpdate`, typeName, `(ctx, in.GetPayload(), db)`)
			p.P(`} else {`)
			p.P(`res, err = DefaultPatch`, typeName, `(ctx, in.GetPayload(), in.Get`, method.fieldMaskName, `(), db)`)
			p.P(`}`)
		} else {
			p.P(`res, err = DefaultStrictUpdate`, typeName, `(ctx, in.GetPayload(), db)`)
		}
		p.P(`if err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
		p.P(`return &`, p.TypeName(method.outType), `{Result: res}, nil`)
		p.P(`}`)
		p.generatePreserviceHook(service.ccName, method.baseType, p.TypeName(method.inType), updateService)
	} else {
		p.generateEmptyBody(method.outType)
	}
}

func (p *OrmPlugin) followsUpdateConventions(inType generator.Object, outType generator.Object, methodName string) (bool, string, string) {
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
	if !typeOrmable {
		p.warning(`stub will be generated for %s since %s incoming message doesn't have "payload" field of ormable type`, methodName, p.TypeName(inType))
		return false, "", ""
	}
	var outTypeName string
	for _, field := range outMsg.Field {
		if field.GetName() == "result" {
			gType, _ := p.GoType(outMsg, field)
			outTypeName = strings.TrimPrefix(gType, "*")
		}
	}
	if inTypeName != outTypeName {
		p.warning(`stub will be generated for %s since "payload" field type of %s incoming message doesn't match "result" field type of %s outcoming message`, methodName, p.TypeName(inType), p.TypeName(outType))
		return false, "", ""
	}
	if !p.hasPrimaryKey(p.getOrmable(inTypeName)) {
		p.warning(`stub will be generated for %s since %s ormable type doesn't have a primary key`, methodName, outTypeName)
		return false, "", ""
	}
	return true, inTypeName, generator.CamelCase(updateMask)
}

func (p *OrmPlugin) generateDeleteServerMethod(service autogenService, method autogenMethod) {
	p.generateMethodSignature(service, method)
	if method.followsConvention {
		typeName := method.baseType
		p.generateDBSetup(service)
		p.generatePreserviceCall(service.ccName, method.baseType, deleteService)
		p.P(`return &`, p.TypeName(method.outType), `{}, `, `DefaultDelete`, typeName, `(ctx, &`, typeName, `{Id: in.GetId()}, db)`)
		p.P(`}`)
		p.generatePreserviceHook(service.ccName, method.baseType, p.TypeName(method.inType), deleteService)
	} else {
		p.generateEmptyBody(method.outType)
	}
}

func (p *OrmPlugin) followsDeleteConventions(inType generator.Object, outType generator.Object, method *descriptor.MethodDescriptorProto) (bool, string) {
	inMsg := inType.(*generator.Descriptor)
	methodName := generator.CamelCase(method.GetName())
	var hasID bool
	for _, field := range inMsg.Field {
		if field.GetName() == "id" {
			hasID = true
		}
	}
	if !hasID {
		p.warning(`stub will be generated for %s since %s incoming message doesn't have "id" field`, methodName, p.TypeName(inType))
		return false, ""
	}
	typeName := generator.CamelCase(getMethodOptions(method).GetObjectType())
	if typeName == "" {
		p.warning(`stub will be generated for %s since (gorm.method).object_type option is not specified`, methodName)
		return false, ""
	}
	if !p.isOrmable(typeName) {
		p.warning(`stub will be generated for %s since %s is not an ormable type`, methodName, typeName)
		return false, ""
	}
	if !p.hasPrimaryKey(p.getOrmable(typeName)) {
		p.warning(`stub will be generated for %s since %s ormable type doesn't have a primary key`, methodName, typeName)
		return false, ""
	}
	return true, typeName
}

func (p *OrmPlugin) generateListServerMethod(service autogenService, method autogenMethod) {
	p.generateMethodSignature(service, method)
	if method.followsConvention {
		p.generateDBSetup(service)
		p.generatePreserviceCall(service.ccName, method.baseType, listService)
		p.P(`res, err := DefaultList`, method.baseType, `(ctx, db, in)`)
		p.P(`if err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
		p.P(`return &`, p.TypeName(method.outType), `{Results: res}, nil`)
		p.P(`}`)
		p.generatePreserviceHook(service.ccName, method.baseType, p.TypeName(method.inType), listService)
	} else {
		p.generateEmptyBody(method.outType)
	}
}

func (p *OrmPlugin) followsListConventions(inType generator.Object, outType generator.Object, methodName string) (bool, string) {
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
	if !typeOrmable {
		p.warning(`stub will be generated for %s since %s incoming message doesn't have "results" field of ormable type`, methodName, p.TypeName(outType))
		return false, ""
	}
	return true, outTypeName
}

func (p *OrmPlugin) generateMethodStub(service autogenService, method autogenMethod) {
	p.generateMethodSignature(service, method)
	p.generateEmptyBody(method.outType)
}

func (p *OrmPlugin) generateMethodSignature(service autogenService, method autogenMethod) {
	p.P(`// `, method.ccName, ` ...`)
	p.P(`func (m *`, service.GetName(), `DefaultServer) `, method.ccName, ` (ctx context.Context, in *`,
		p.TypeName(method.inType), `) (*`, p.TypeName(method.outType), `, error) {`)
	p.RecordTypeUse(method.GetInputType())
	p.RecordTypeUse(method.GetOutputType())
}

func (p *OrmPlugin) generateDBSetup(service autogenService) error {
	if service.usesTxnMiddleware {
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

func (p *OrmPlugin) getMethodProps(method *descriptor.MethodDescriptorProto) (generator.Object, generator.Object, string) {
	inType := p.ObjectNamed(method.GetInputType())
	outType := p.ObjectNamed(method.GetOutputType())
	methodName := generator.CamelCase(method.GetName())
	return inType, outType, methodName
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

func (p *OrmPlugin) hasFieldsSelector(object generator.Object) string {
	msg := object.(*generator.Descriptor)
	for _, field := range msg.Field {
		fieldName := generator.CamelCase(field.GetName())
		fieldType, _ := p.GoType(msg, field)
		parts := strings.Split(fieldType, ".")
		if parts[len(parts)-1] == "FieldSelection" {
			return fieldName
		}
	}
	return ""
}
