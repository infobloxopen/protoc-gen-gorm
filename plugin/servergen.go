package plugin

import (
	"strings"

	"fmt"

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

type autogenService struct {
	*descriptor.ServiceDescriptorProto
	ccName            string
	file              *generator.FileDescriptor
	usesTxnMiddleware bool
	methods           []autogenMethod
	autogen           bool
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
	defaultSuppressWarn := p.suppressWarn
	for _, service := range file.GetService() {
		genSvc := autogenService{
			ServiceDescriptorProto: service,
			ccName:                 generator.CamelCase(service.GetName()),
			file:                   file,
		}
		if opts := getServiceOptions(service); opts != nil {
			genSvc.autogen = opts.GetAutogen()
			genSvc.usesTxnMiddleware = opts.GetTxnMiddleware()
		}
		if !genSvc.autogen {
			p.suppressWarn = true
		}
		for _, method := range service.GetMethod() {
			inType, outType, methodName := p.getMethodProps(method)
			var verb, fmName, baseType string
			var follows bool
			if methodName == createService {
				verb = createService
				follows, baseType = p.followsCreateConventions(inType, outType, methodName)
			} else if methodName == readService {
				verb = readService
				follows, baseType = p.followsReadConventions(inType, outType, methodName)
			} else if methodName == updateService {
				verb = updateService
				follows, baseType, fmName = p.followsUpdateConventions(inType, outType, methodName)
			} else if methodName == deleteService {
				verb = deleteService
				follows, baseType = p.followsDeleteConventions(inType, outType, method)
			} else if methodName == listService {
				verb = listService
				follows, baseType = p.followsListConventions(inType, outType, methodName)
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
		p.ormableServices = append(p.ormableServices, genSvc)
		p.suppressWarn = defaultSuppressWarn
	}
}

func (p *OrmPlugin) generateDefaultServer(file *generator.FileDescriptor) {
	for _, service := range p.ormableServices {
		if service.file != file || !service.autogen {
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
		p.P(`out := &`, p.TypeName(method.outType), `{Result: res}`)
		p.generatePostserviceCall(service.ccName, method.baseType, createService)
		p.P(`return out, nil`)
		p.P(`}`)
		p.generatePreserviceHook(service.ccName, method.baseType, createService)
		p.generatePostserviceHook(service.ccName, method.baseType, p.TypeName(method.outType), method.ccName)
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
		p.generatePreserviceCall(service.ccName, method.baseType, method.ccName)
		typeName := method.baseType
		if fields := p.getFieldSelection(method.inType); fields != "" {
			p.P(`res, err := DefaultRead`, typeName, `(ctx, &`, typeName, `{Id: in.GetId()}, db, in.`, fields, `)`)
		} else {
			p.P(`res, err := DefaultRead`, typeName, `(ctx, &`, typeName, `{Id: in.GetId()}, db)`)
		}
		p.P(`if err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
		p.P(`out := &`, p.TypeName(method.outType), `{Result: res}`)
		p.generatePostserviceCall(service.ccName, method.baseType, method.ccName)
		p.P(`return out, nil`)
		p.P(`}`)
		p.generatePreserviceHook(service.ccName, method.baseType, method.ccName)
		p.generatePostserviceHook(service.ccName, method.baseType, p.TypeName(method.outType), method.ccName)
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
		p.generatePreserviceCall(service.ccName, method.baseType, method.ccName)
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
		p.P(`out := &`, p.TypeName(method.outType), `{Result: res}`)
		p.generatePostserviceCall(service.ccName, method.baseType, method.ccName)
		p.P(`return out, nil`)
		p.P(`}`)
		p.generatePreserviceHook(service.ccName, method.baseType, method.ccName)
		p.generatePostserviceHook(service.ccName, method.baseType, p.TypeName(method.outType), method.ccName)
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
		p.generatePreserviceCall(service.ccName, method.baseType, method.ccName)
		p.P(`err := DefaultDelete`, typeName, `(ctx, &`, typeName, `{Id: in.GetId()}, db)`)
		p.P(`if err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
		p.P(`out := &`, p.TypeName(method.outType), `{}`)
		p.generatePostserviceCall(service.ccName, method.baseType, method.ccName)
		p.P(`return out, nil`)
		p.P(`}`)
		p.generatePreserviceHook(service.ccName, method.baseType, method.ccName)
		p.generatePostserviceHook(service.ccName, method.baseType, p.TypeName(method.outType), method.ccName)
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
		p.generatePreserviceCall(service.ccName, method.baseType, method.ccName)
		handlerCall := fmt.Sprint(`res, err := DefaultList`, method.baseType, `(ctx, db`)
		if f := p.getFiltering(method.inType); f != "" {
			handlerCall += fmt.Sprint(",in.", f)
		}
		if s := p.getSorting(method.inType); s != "" {
			handlerCall += fmt.Sprint(",in.", s)
		}
		if pg := p.getPagination(method.inType); pg != "" {
			handlerCall += fmt.Sprint(",in.", pg)
		}
		if fs := p.getFieldSelection(method.inType); fs != "" {
			handlerCall += fmt.Sprint(",in.", fs)
		}
		handlerCall += ")"
		p.P(handlerCall)
		p.P(`if err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
		p.P(`out := &`, p.TypeName(method.outType), `{Results: res}`)
		p.generatePostserviceCall(service.ccName, method.baseType, method.ccName)
		p.P(`return out, nil`)
		p.P(`}`)
		p.generatePreserviceHook(service.ccName, method.baseType, method.ccName)
		p.generatePostserviceHook(service.ccName, method.baseType, p.TypeName(method.outType), method.ccName)
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
	p.P(`if db, err = custom.Before`, mthd, `(ctx, db); err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generatePreserviceHook(svc, typeName, mthd string) {
	p.P(`// `, svc, typeName, `WithBefore`, mthd, ` called before Default`, mthd, typeName, ` in the default `, mthd, ` handler`)
	p.P(`type `, svc, typeName, `WithBefore`, mthd, ` interface {`)
	p.P(`Before`, mthd, `(context.Context, *`, p.Import(gormImport), `.DB) (*`, p.Import(gormImport), `.DB, error)`)
	p.P(`}`)
}

func (p *OrmPlugin) generatePostserviceCall(svc, typeName, mthd string) {
	p.P(`if custom, ok := interface{}(in).(`, svc, typeName, `WithAfter`, mthd, `); ok {`)
	p.P(`var err error`)
	p.P(`if err = custom.After`, mthd, `(ctx, out, db); err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generatePostserviceHook(svc, typeName, outTypeName, mthd string) {
	p.P(`// `, svc, typeName, `WithAfter`, mthd, ` called before Default`, mthd, typeName, ` in the default `, mthd, ` handler`)
	p.P(`type `, svc, typeName, `WithAfter`, mthd, ` interface {`)
	p.P(`After`, mthd, `(context.Context, *`, outTypeName, `, *`, p.Import(gormImport), `.DB) error`)
	p.P(`}`)
}

func (p *OrmPlugin) getFieldSelection(object generator.Object) string {
	return p.getFieldOfType(object, "FieldSelection")
}

func (p *OrmPlugin) getFiltering(object generator.Object) string {
	return p.getFieldOfType(object, "Filtering")
}

func (p *OrmPlugin) getSorting(object generator.Object) string {
	return p.getFieldOfType(object, "Sorting")
}

func (p *OrmPlugin) getPagination(object generator.Object) string {
	return p.getFieldOfType(object, "Pagination")
}

func (p *OrmPlugin) getFieldOfType(object generator.Object, fieldType string) string {
	msg := object.(*generator.Descriptor)
	for _, field := range msg.Field {
		goFieldName := generator.CamelCase(field.GetName())
		goFieldType, _ := p.GoType(msg, field)
		parts := strings.Split(goFieldType, ".")
		if parts[len(parts)-1] == fieldType {
			return goFieldName
		}
	}
	return ""
}
