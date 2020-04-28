package plugin

import (
	"strings"

	"fmt"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

const (
	createService    = "Create"
	readService      = "Read"
	updateService    = "Update"
	updateSetService = "UpdateSet"
	deleteService    = "Delete"
	deleteSetService = "DeleteSet"
	listService      = "List"
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
			if strings.HasPrefix(methodName, createService) {
				verb = createService
				follows, baseType = p.followsCreateConventions(inType, outType, createService)
			} else if strings.HasPrefix(methodName, readService) {
				verb = readService
				follows, baseType = p.followsReadConventions(inType, outType, readService)
			} else if strings.HasPrefix(methodName, updateSetService) {
				verb = updateSetService
				follows, baseType, fmName = p.followsUpdateSetConventions(inType, outType, updateSetService)
			} else if strings.HasPrefix(methodName, updateService) {
				verb = updateService
				follows, baseType, fmName = p.followsUpdateConventions(inType, outType, updateService)
			} else if strings.HasPrefix(methodName, deleteSetService) {
				verb = deleteSetService
				follows, baseType = p.followsDeleteSetConventions(inType, outType, method)
			} else if strings.HasPrefix(methodName, deleteService) {
				verb = deleteService
				follows, baseType = p.followsDeleteConventions(inType, outType, method)
			} else if strings.HasPrefix(methodName, listService) {
				verb = listService
				follows, baseType = p.followsListConventions(inType, outType, listService)
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
		withSpan := getServiceOptions(service.ServiceDescriptorProto).WithTracing
		if withSpan != nil && *withSpan {
			p.generateSpanInstantiationMethod(service)
			p.generateSpanErrorMethod(service)
			p.generateSpanResultMethod(service)
		}
		for _, method := range service.methods {
			//Import context there because it have used in functions parameters
			p.UsingGoImports(stdCtxImport)
			switch method.verb {
			case createService:
				p.generateCreateServerMethod(service, method)
			case readService:
				p.generateReadServerMethod(service, method)
			case updateService:
				p.generateUpdateServerMethod(service, method)
			case updateSetService:
				p.generateUpdateSetServerMethod(service, method)
			case deleteService:
				p.generateDeleteServerMethod(service, method)
			case deleteSetService:
				p.generateDeleteSetServerMethod(service, method)
			case listService:
				p.generateListServerMethod(service, method)
			default:
				p.generateMethodStub(service, method)
			}
		}
	}
}

func (p *OrmPlugin) generateSpanInstantiationMethod(service autogenService) {
	p.UsingGoImports(stdFmtImport)
	p.P(`func (m *`, service.GetName(), `DefaultServer) spanCreate(ctx context.Context, in interface{}, methodName string) (*`, p.Import(ocTraceImport), `.Span, error) {`)
	p.P(`_, span := `, p.Import(ocTraceImport), `.StartSpan(ctx, fmt.Sprint("`, service.GetName(), `DefaultServer.", methodName))`)
	p.P(`raw, err := `, p.Import(encodingJsonImport), `.Marshal(in)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`span.Annotate([]`, p.Import(ocTraceImport), `.Attribute{`, p.Import(ocTraceImport), `.StringAttribute("in", string(raw))}, "in parameter")`)
	p.P(`return span, nil`)
	p.P(`}`)
}

func (p *OrmPlugin) generateSpanErrorMethod(service autogenService) {
	p.P(`// spanError ...`)
	p.P(`func (m *`, service.GetName(), `DefaultServer) spanError(span *`, p.Import(ocTraceImport), `.Span, err error) error {`)
	p.P(`span.SetStatus(`, p.Import(ocTraceImport), `.Status{`)
	p.P(`Code: `, p.Import(ocTraceImport), `.StatusCodeUnknown,`)
	p.P(`Message: err.Error(),`)
	p.P(`})`)
	p.P(`return err`)
	p.P(`}`)
}

func (p *OrmPlugin) generateSpanResultMethod(service autogenService) {
	p.P(`// spanResult ...`)
	p.P(`func (m *`, service.GetName(), `DefaultServer) spanResult(span *`, p.Import(ocTraceImport), `.Span, out interface{}) error {`)
	p.P(`raw, err := `, p.Import(encodingJsonImport), `.Marshal(out)`)
	p.P(`if err != nil {`)
	p.P(`return err`)
	p.P(`}`)
	p.P(`span.Annotate([]`, p.Import(ocTraceImport), `.Attribute{`, p.Import(ocTraceImport), `.StringAttribute("out", string(raw))}, "out parameter")`)
	p.P(`return nil`)
	p.P(`}`)
}

func (p *OrmPlugin) wrapSpanError(service autogenService, errVarName string) string {
	withSpan := getServiceOptions(service.ServiceDescriptorProto).WithTracing
	if withSpan != nil && *withSpan {
		return fmt.Sprint(`m.spanError(span, `, errVarName, `)`)
	}
	return errVarName
}

func (p *OrmPlugin) generateCreateServerMethod(service autogenService, method autogenMethod) {
	p.generateMethodSignature(service, method)
	if method.followsConvention {
		p.generateDBSetup(service)
		p.generatePreserviceCall(service, method.baseType, method.ccName)
		p.P(`res, err := DefaultCreate`, method.baseType, `(ctx, in.GetPayload(), db)`)
		p.P(`if err != nil {`)
		p.P(`return nil, `, p.wrapSpanError(service, "err"))
		p.P(`}`)
		p.P(`out := &`, p.TypeName(method.outType), `{Result: res}`)
		if p.gateway {
			p.P(`err = `, p.Import(gatewayImport), `.SetCreated(ctx, "")`)
			p.P(`if err != nil {`)
			p.P(`return nil, `, p.wrapSpanError(service, "err"))
			p.P(`}`)
		}

		p.generatePostserviceCall(service, method.baseType, method.ccName)
		p.spanResultHandling(service)
		p.P(`return out, nil`)
		p.P(`}`)
		p.generatePreserviceHook(service.ccName, method.baseType, method.ccName)
		p.generatePostserviceHook(service.ccName, method.baseType, p.TypeName(method.outType), method.ccName)
	} else {
		p.generateEmptyBody(service, method.outType)
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
		p.generatePreserviceCall(service, method.baseType, method.ccName)
		typeName := method.baseType
		if fields := p.getFieldSelection(method.inType); fields != "" {
			p.P(`res, err := DefaultRead`, typeName, `(ctx, &`, typeName, `{Id: in.GetId()}, db, in.`, fields, `)`)
		} else {
			p.P(`res, err := DefaultRead`, typeName, `(ctx, &`, typeName, `{Id: in.GetId()}, db)`)
		}
		p.P(`if err != nil {`)
		p.P(`return nil, `, p.wrapSpanError(service, "err"))
		p.P(`}`)
		p.P(`out := &`, p.TypeName(method.outType), `{Result: res}`)
		p.generatePostserviceCall(service, method.baseType, method.ccName)
		p.spanResultHandling(service)
		p.P(`return out, nil`)
		p.P(`}`)
		p.generatePreserviceHook(service.ccName, method.baseType, method.ccName)
		p.generatePostserviceHook(service.ccName, method.baseType, p.TypeName(method.outType), method.ccName)
	} else {
		p.generateEmptyBody(service, method.outType)
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
		p.generatePreserviceCall(service, method.baseType, method.ccName)
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
		p.P(`return nil, `, p.wrapSpanError(service, "err"))
		p.P(`}`)
		p.P(`out := &`, p.TypeName(method.outType), `{Result: res}`)
		p.generatePostserviceCall(service, method.baseType, method.ccName)
		p.spanResultHandling(service)
		p.P(`return out, nil`)
		p.P(`}`)
		p.generatePreserviceHook(service.ccName, method.baseType, method.ccName)
		p.generatePostserviceHook(service.ccName, method.baseType, p.TypeName(method.outType), method.ccName)
	} else {
		p.generateEmptyBody(service, method.outType)
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

func (p *OrmPlugin) generateUpdateSetServerMethod(service autogenService, method autogenMethod) {
	p.generateMethodSignature(service, method)
	if method.followsConvention {
		typeName := method.baseType
		typeName = strings.TrimPrefix(typeName, "[]*")
		p.P(`if in == nil {`)
		p.P(`return nil,`, p.Import(gerrorsImport), `.NilArgumentError`)
		p.P(`}`)
		p.P(``)
		p.generateDBSetup(service)
		p.P(``)
		p.generatePreserviceCall(service, typeName, method.ccName)

		p.P(``)
		p.P(`res, err := DefaultPatchSet`, typeName, `(ctx, in.GetObjects(), in.Get`, method.fieldMaskName, `(), db)`)
		p.P(`if err != nil {`)
		p.P(`return nil, `, p.wrapSpanError(service, "err"))
		p.P(`}`)
		p.P(``)
		p.P(`out := &`, p.TypeName(method.outType), `{Results: res}`)

		p.P(``)
		p.generatePostserviceCall(service, typeName, method.ccName)
		p.P(``)
		withSpan := getServiceOptions(service.ServiceDescriptorProto).WithTracing
		if withSpan != nil && *withSpan {
			p.P(`err = m.spanResult(span, out)`)
			p.P(`if err != nil {`)
			p.P(`return nil,`, p.wrapSpanError(service, "err"))
			p.P(`}`)
		}
		p.P(`return out, nil`)
		p.P(`}`)

		p.generatePreserviceHook(service.ccName, typeName, method.ccName)
		p.generatePostserviceHook(service.ccName, typeName, p.TypeName(method.outType), method.ccName)
	} else {
		p.generateEmptyBody(service, method.outType)
	}
}

func (p *OrmPlugin) followsUpdateSetConventions(inType generator.Object, outType generator.Object, methodName string) (bool, string, string) {

	inMsg, outMsg := inType.(*generator.Descriptor), outType.(*generator.Descriptor)

	var (
		inEntity    *descriptor.FieldDescriptorProto
		inFieldMask *descriptor.FieldDescriptorProto
	)
	for _, f := range inMsg.Field {
		if f.GetName() == "objects" {
			inEntity = f
		}

		if f.GetTypeName() == ".google.protobuf.FieldMask" {
			if inFieldMask != nil {
				p.warning("message must not contains double field mask, prev on field name %s, after on field %s", inFieldMask.GetName(), f.GetName())
				return false, "", ""
			}

			inFieldMask = f
		}
	}

	var outEntity *descriptor.FieldDescriptorProto
	for _, f := range outMsg.Field {
		if f.GetName() == "results" {
			outEntity = f
		}
	}

	if inFieldMask == nil || !inFieldMask.IsRepeated() {
		p.warning("repeated field mask should exist in request for method %q", methodName)
		return false, "", ""
	}

	if inEntity == nil || outEntity == nil {
		p.warning(`method: %q, request should has repeated field 'objects' in request and repeated field 'results' in response`, methodName)
		return false, "", ""
	}

	if !inEntity.IsRepeated() || !outEntity.IsRepeated() {
		p.warning(`method: %q, field 'objects' in request and field 'results' in response should be repeated`, methodName)
		return false, "", ""
	}

	inGoType, _ := p.GoType(inMsg, inEntity)
	outGoType, _ := p.GoType(outMsg, outEntity)
	inTypeName, outTypeName := strings.TrimPrefix(inGoType, "*"), strings.TrimPrefix(outGoType, "*")
	if !p.isOrmable(inTypeName) {
		p.warning("method: %q, type %q must be ormable", methodName, inTypeName)
		return false, "", ""
	}

	if inTypeName != outTypeName {
		p.warning("method: %q, field 'objects' in request has type: %q but field 'results' in response has: %q", methodName, inTypeName, outTypeName)
		return false, "", ""
	}

	return true, inTypeName, generator.CamelCase(inFieldMask.GetName())
}

func (p *OrmPlugin) generateDeleteServerMethod(service autogenService, method autogenMethod) {
	p.generateMethodSignature(service, method)
	if method.followsConvention {
		typeName := method.baseType
		p.generateDBSetup(service)
		p.generatePreserviceCall(service, method.baseType, method.ccName)
		p.P(`err := DefaultDelete`, typeName, `(ctx, &`, typeName, `{Id: in.GetId()}, db)`)
		p.P(`if err != nil {`)
		p.P(`return nil, `, p.wrapSpanError(service, "err"))
		p.P(`}`)
		p.P(`out := &`, p.TypeName(method.outType), `{}`)
		p.generatePostserviceCall(service, method.baseType, method.ccName)
		p.spanResultHandling(service)
		p.P(`return out, nil`)
		p.P(`}`)
		p.generatePreserviceHook(service.ccName, method.baseType, method.ccName)
		p.generatePostserviceHook(service.ccName, method.baseType, p.TypeName(method.outType), method.ccName)
	} else {
		p.generateEmptyBody(service, method.outType)
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

func (p *OrmPlugin) generateDeleteSetServerMethod(service autogenService, method autogenMethod) {
	p.generateMethodSignature(service, method)
	if method.followsConvention {
		typeName := method.baseType
		p.generateDBSetup(service)
		p.P(`objs := []*`, typeName, `{}`)
		p.P(`for _, id := range in.Ids {`)
		p.P(`objs = append(objs, &`, typeName, `{Id: id})`)
		p.P(`}`)
		p.generatePreserviceCall(service, method.baseType, method.ccName)
		p.P(`err := DefaultDelete`, typeName, `Set(ctx, objs, db)`)
		p.P(`if err != nil {`)
		p.P(`return nil, `, p.wrapSpanError(service, "err"))
		p.P(`}`)
		p.P(`out := &`, p.TypeName(method.outType), `{}`)
		p.generatePostserviceCall(service, method.baseType, method.ccName)
		p.spanResultHandling(service)
		p.P(`return out, nil`)
		p.P(`}`)
		p.generatePreserviceHook(service.ccName, method.baseType, method.ccName)
		p.generatePostserviceHook(service.ccName, method.baseType, p.TypeName(method.outType), method.ccName)
	} else {
		p.generateEmptyBody(service, method.outType)
	}
}

func (p *OrmPlugin) followsDeleteSetConventions(inType generator.Object, outType generator.Object, method *descriptor.MethodDescriptorProto) (bool, string) {
	inMsg := inType.(*generator.Descriptor)
	methodName := generator.CamelCase(method.GetName())
	var hasIDs bool
	for _, field := range inMsg.Field {
		if field.GetName() == "ids" && field.IsRepeated() {
			hasIDs = true
		}
	}
	if !hasIDs {
		p.warning(`stub will be generated for %s since %s incoming message doesn't have "ids" field`, methodName, p.TypeName(inType))
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
		p.generatePreserviceCall(service, method.baseType, method.ccName)
		pg := p.getPagination(method.inType)
		pi := p.getPageInfo(method.outType)
		if pg != "" && pi != "" {
			p.generatePagedRequestSetup(pg)
		}
		handlerCall := fmt.Sprint(`res, err := DefaultList`, method.baseType, `(ctx, db`)
		if f := p.getFiltering(method.inType); f != "" {
			handlerCall += fmt.Sprint(",in.", f)
		}
		if s := p.getSorting(method.inType); s != "" {
			handlerCall += fmt.Sprint(",in.", s)
		}
		if pg != "" {
			handlerCall += fmt.Sprint(",in.", pg)
		}
		if fs := p.getFieldSelection(method.inType); fs != "" {
			handlerCall += fmt.Sprint(",in.", fs)
		}
		handlerCall += ")"
		p.P(handlerCall)
		p.P(`if err != nil {`)
		p.P(`return nil, `, p.wrapSpanError(service, "err"))
		p.P(`}`)
		var pageInfoIfExist string
		if pg != "" && pi != "" {
			p.generatePagedRequestHandling(pg)
			pageInfoIfExist = ", " + pi + ": resPaging"
		}
		p.P(`out := &`, p.TypeName(method.outType), `{Results: res`, pageInfoIfExist, ` }`)
		p.generatePostserviceCall(service, method.baseType, method.ccName)
		p.spanResultHandling(service)
		p.P(`return out, nil`)
		p.P(`}`)
		p.generatePreserviceHook(service.ccName, method.baseType, method.ccName)
		p.generatePostserviceHook(service.ccName, method.baseType, p.TypeName(method.outType), method.ccName)
	} else {
		p.generateEmptyBody(service, method.outType)
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
	p.generateEmptyBody(service, method.outType)
}

func (p *OrmPlugin) generateMethodSignature(service autogenService, method autogenMethod) {
	p.P(`// `, method.ccName, ` ...`)
	p.P(`func (m *`, service.GetName(), `DefaultServer) `, method.ccName, ` (ctx context.Context, in *`,
		p.TypeName(method.inType), `) (*`, p.TypeName(method.outType), `, error) {`)
	p.RecordTypeUse(method.GetInputType())
	p.RecordTypeUse(method.GetOutputType())
	withSpan := getServiceOptions(service.ServiceDescriptorProto).WithTracing
	if withSpan != nil && *withSpan {
		p.P(`span, errSpanCreate := m.spanCreate(ctx, in, "`, method.ccName, `")`)
		p.P(`if errSpanCreate != nil {`)
		p.P(`return nil, errSpanCreate`)
		p.P(`}`)
		p.P(`defer span.End()`)
	}
}

func (p *OrmPlugin) generateDBSetup(service autogenService) error {
	if service.usesTxnMiddleware {
		p.P(`txn, ok := `, p.Import(tkgormImport), `.FromContext(ctx)`)
		p.P(`if !ok {`)
		p.P(`return nil, `, p.Import(gerrorsImport), `.NoTransactionError`)
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

func (p *OrmPlugin) spanResultHandling(service autogenService) {
	withSpan := getServiceOptions(service.ServiceDescriptorProto).WithTracing
	if withSpan != nil && *withSpan {
		p.P(`errSpanResult := m.spanResult(span, out)`)
		p.P(`if errSpanResult != nil {`)
		p.P(`return nil, `, p.wrapSpanError(service, "errSpanResult"))
		p.P(`}`)
	}
}

func (p OrmPlugin) generateEmptyBody(service autogenService, outType generator.Object) {
	p.P(`out:= &`, p.TypeName(outType), `{}`)
	p.spanResultHandling(service)
	p.P(`return out, nil`)
	p.P(`}`)
}

func (p *OrmPlugin) getMethodProps(method *descriptor.MethodDescriptorProto) (generator.Object, generator.Object, string) {
	inType := p.ObjectNamed(method.GetInputType())
	outType := p.ObjectNamed(method.GetOutputType())
	methodName := generator.CamelCase(method.GetName())
	return inType, outType, methodName
}

func (p *OrmPlugin) generatePreserviceCall(service autogenService, typeName, mthd string) {
	p.P(`if custom, ok := interface{}(in).(`, service.ccName, typeName, `WithBefore`, mthd, `); ok {`)
	p.P(`var err error`)
	p.P(`if db, err = custom.Before`, mthd, `(ctx, db); err != nil {`)
	p.P(`return nil, `, p.wrapSpanError(service, "err"))
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generatePagedRequestSetup(pg string) {
	p.P(`pagedRequest := false`)
	p.P(fmt.Sprintf(`if in.Get%s().GetLimit()>=1 {`, pg))
	p.P(fmt.Sprintf(`in.%s.Limit ++`, pg))
	p.P(`pagedRequest=true`)
	p.P(`}`)
}

func (p *OrmPlugin) generatePagedRequestHandling(pg string) {
	p.P(fmt.Sprintf(`var resPaging *%s.PageInfo`, p.Import(queryImport)))
	p.P(`if pagedRequest {`)
	p.P(`var offset int32`)
	p.P(`var size int32 = int32(len(res))`)
	p.P(fmt.Sprintf(`if size == in.Get%s().GetLimit(){`, pg))
	p.P(`size--`)
	p.P(`res=res[:size]`)
	p.P(fmt.Sprintf(`offset=in.Get%s().GetOffset()+size`, pg))
	p.P(`}`)
	p.P(fmt.Sprintf(`resPaging = &%s.PageInfo{Offset: offset}`, p.Import(queryImport)))
	p.P(`}`)
}

func (p *OrmPlugin) generatePreserviceHook(svc, typeName, mthd string) {
	p.P(`// `, svc, typeName, `WithBefore`, mthd, ` called before Default`, mthd, typeName, ` in the default `, mthd, ` handler`)
	p.P(`type `, svc, typeName, `WithBefore`, mthd, ` interface {`)
	p.P(`Before`, mthd, `(context.Context, *`, p.Import(gormImport), `.DB) (*`, p.Import(gormImport), `.DB, error)`)
	p.P(`}`)
}

func (p *OrmPlugin) generatePostserviceCall(service autogenService, typeName, mthd string) {
	p.P(`if custom, ok := interface{}(in).(`, service.ccName, typeName, `WithAfter`, mthd, `); ok {`)
	p.P(`var err error`)
	p.P(`if err = custom.After`, mthd, `(ctx, out, db); err != nil {`)
	p.P(`return nil, `, p.wrapSpanError(service, "err"))
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

func (p *OrmPlugin) getPageInfo(object generator.Object) string {
	return p.getFieldOfType(object, "PageInfo")
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
