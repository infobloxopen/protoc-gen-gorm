package plugin

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	jgorm "github.com/jinzhu/gorm"
)

func (p *OrmPlugin) generateDefaultHandlers(file *generator.FileDescriptor) {
	for _, message := range file.Messages() {
		if getMessageOptions(message).GetOrmable() {
			p.UsingGoImports("context", "errors")

			p.generateCreateHandler(message)
			// FIXME: Temporary fix for Ormable objects that have no ID field but
			// have pk.
			if p.hasPrimaryKey(p.getOrmable(p.TypeName(message))) && p.hasIDField(message) {
				p.generateReadHandler(message)
				p.generateDeleteHandler(message)
				p.generateStrictUpdateHandler(message)
				p.generatePatchHandler(message)
			}
			p.generateApplyFieldMask(message)
			p.generateListHandler(message)
		}
	}
}

func (p *OrmPlugin) generateBeforeHookDef(orm *OrmableType, method string) {
	p.P(`type `, orm.Name, `WithBefore`, method, ` interface {`)
	p.P(`Before`, method, `(context.Context, *`, p.Import(gormImport), `.DB) (*`, p.Import(gormImport), `.DB, error)`)
	p.P(`}`)
}

func (p *OrmPlugin) generateAfterHookDef(orm *OrmableType, method string) {
	p.P(`type `, orm.Name, `WithAfter`, method, ` interface {`)
	p.P(`After`, method, `(context.Context, *`, p.Import(gormImport), `.DB) error`)
	p.P(`}`)
}

func (p *OrmPlugin) generateBeforeHookCall(orm *OrmableType, method string) {
	p.P(`if hook, ok := interface{}(&ormObj).(`, orm.Name, `WithBefore`, method, `); ok {`)
	p.P(`if db, err = hook.Before`, method, `(ctx, db); err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generateAfterHookCall(orm *OrmableType, method string) {
	p.P(`if hook, ok := interface{}(&ormObj).(`, orm.Name, `WithAfter`, method, `); ok {`)
	p.P(`if err = hook.After`, method, `(ctx, db); err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generateCreateHandler(message *generator.Descriptor) {
	typeName := p.TypeName(message)
	orm := p.getOrmable(typeName)
	p.P(`// DefaultCreate`, typeName, ` executes a basic gorm create call`)
	p.P(`func DefaultCreate`, typeName, `(ctx context.Context, in *`,
		typeName, `, db *`, p.Import(gormImport), `.DB) (*`, typeName, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, errors.New("Nil argument to DefaultCreate`, typeName, `")`)
	p.P(`}`)
	p.P(`ormObj, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.generateBeforeHookCall(orm, "Create")
	p.P(`if err = db.Create(&ormObj).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.generateAfterHookCall(orm, "Create")
	p.P(`pbResponse, err := ormObj.ToPB(ctx)`)
	p.P(`return &pbResponse, err`)
	p.P(`}`)
	p.generateBeforeHookDef(orm, "Create")
	p.generateAfterHookDef(orm, "Create")
}

func (p *OrmPlugin) generateReadHandler(message *generator.Descriptor) {
	typeName := p.TypeName(message)
	ormable := p.getOrmable(typeName)
	p.P(`// DefaultRead`, typeName, ` executes a basic gorm read call`)
	// Different behavior if there is a
	if p.readHasFieldSelection(ormable) {
		p.P(`func DefaultRead`, typeName, `(ctx context.Context, in *`,
			typeName, `, db *`, p.Import(gormImport), `.DB, fs *`, p.Import(queryImport), `.FieldSelection) (*`, typeName, `, error) {`)
	} else {
		p.P(`func DefaultRead`, typeName, `(ctx context.Context, in *`,
			typeName, `, db *`, p.Import(gormImport), `.DB) (*`, typeName, `, error) {`)
	}
	p.P(`if in == nil {`)
	p.P(`return nil, errors.New("Nil argument to DefaultRead`, typeName, `")`)
	p.P(`}`)

	p.P(`ormObj, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)

	k, f := p.findPrimaryKey(ormable)
	if strings.Contains(f.Type, "*") {
		p.P(`if ormObj.`, k, ` == nil || *ormObj.`, k, ` == `, p.guessZeroValue(f.Type), ` {`)
	} else {
		p.P(`if ormObj.`, k, ` == `, p.guessZeroValue(f.Type), ` {`)
	}
	p.P(`return nil, errors.New("DefaultRead`, typeName, ` requires a non-zero primary key")`)
	p.P(`}`)

	var fs string
	if p.readHasFieldSelection(ormable) {
		fs = "fs"
	} else {
		fs = "nil"
	}

	p.generateBeforeReadHookCall(ormable, "ApplyQuery")
	p.P(`if db, err = `, p.Import(tkgormImport), `.ApplyFieldSelection(ctx, db, `, fs, `, &`, ormable.Name, `{}); err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)

	p.generateBeforeReadHookCall(ormable, "Find")
	p.P(`ormResponse := `, ormable.Name, `{}`)
	p.P(`if err = db.Where(&ormObj).First(&ormResponse).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.generateAfterReadHookCall(ormable)
	p.P(`pbResponse, err := ormResponse.ToPB(ctx)`)
	p.P(`return &pbResponse, err`)
	p.P(`}`)
	p.generateBeforeReadHookDef(ormable, "ApplyQuery")
	p.generateBeforeReadHookDef(ormable, "Find")
	p.generateAfterReadHookDef(ormable)
}

func (p *OrmPlugin) generateBeforeReadHookDef(orm *OrmableType, suffix string) {
	p.P(`type `, orm.Name, `WithBeforeRead`, suffix, ` interface {`)
	hookSign := fmt.Sprint(`BeforeRead`, suffix, `(context.Context, *`, p.Import(gormImport), `.DB`)
	if p.readHasFieldSelection(orm) {
		hookSign += fmt.Sprint(`, *`, p.Import(queryImport), `.FieldSelection`)
	}
	hookSign += fmt.Sprint(`) (*`, p.Import(gormImport), `.DB, error)`)
	p.P(hookSign)
	p.P(`}`)
}

func (p *OrmPlugin) generateAfterReadHookDef(orm *OrmableType) {
	p.P(`type `, orm.Name, `WithAfterReadFind interface {`)
	hookSign := fmt.Sprint(`AfterReadFind`, `(context.Context, *`, p.Import(gormImport), `.DB`)
	if p.readHasFieldSelection(orm) {
		hookSign += fmt.Sprint(`, *`, p.Import(queryImport), `.FieldSelection`)
	}
	hookSign += `) error`
	p.P(hookSign)
	p.P(`}`)
}

func (p *OrmPlugin) generateBeforeReadHookCall(orm *OrmableType, suffix string) {
	p.P(`if hook, ok := interface{}(&ormObj).(`, orm.Name, `WithBeforeRead`, suffix, `); ok {`)
	hookCall := fmt.Sprint(`if db, err = hook.BeforeRead`, suffix, `(ctx, db`)
	if p.readHasFieldSelection(orm) {
		hookCall += `, fs`
	}
	hookCall += `); err != nil{`
	p.P(hookCall)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generateAfterReadHookCall(orm *OrmableType) {
	p.P(`if hook, ok := interface{}(&ormResponse).(`, orm.Name, `WithAfterReadFind`, `); ok {`)
	hookCall := fmt.Sprint(`if err = hook.AfterReadFind(ctx, db`)
	if p.readHasFieldSelection(orm) {
		hookCall += `, fs`
	}
	hookCall += `); err != nil {`
	p.P(hookCall)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generateApplyFieldMask(message *generator.Descriptor) {
	typeName := p.TypeName(message)
	p.P(`// DefaultApplyFieldMask`, typeName, ` patches an pbObject with patcher according to a field mask.`)
	p.P(`func DefaultApplyFieldMask`, typeName, `(ctx context.Context, patchee *`,
		typeName, `, patcher *`, typeName, `, updateMask *`, p.Import(fmImport),
		`.FieldMask, prefix string, db *`, p.Import(gormImport), `.DB) (*`, typeName, `, error) {`)

	p.P(`if patcher == nil {`)
	p.P(`return nil, nil`)
	p.P(`} else if patchee == nil {`)
	p.P(`return nil, errors.New("Patchee inputs to DefaultApplyFieldMask`,
		typeName, ` must be non-nil")`)
	p.P(`}`)
	p.P(`var err error`)
	hasNested := false
	for _, field := range message.GetField() {
		fieldType, _ := p.GoType(message, field)
		if field.IsMessage() && !isSpecialType(fieldType) && !field.IsRepeated() {
			p.P(`var updated`, generator.CamelCase(field.GetName()), ` bool`)
			hasNested = true
		}
	}
	// Patch pbObj with input according to a field mask.
	if hasNested {
		p.UsingGoImports("strings")
		p.P(`for i, f := range updateMask.Paths {`)
	} else {
		p.P(`for _, f := range updateMask.Paths {`)
	}
	for _, field := range message.GetField() {
		ccName := generator.CamelCase(field.GetName())
		fieldType, _ := p.GoType(message, field)
		//  for ormable message, do recursive patching
		if field.IsMessage() && p.isOrmable(fieldType) && !field.IsRepeated() {
			p.P(`if strings.HasPrefix(f, prefix+"`, ccName, `.") && !updated`, ccName, ` {`)
			p.P(`updated`, ccName, ` = true`)
			p.P(`if patcher.`, ccName, ` == nil {`)
			p.P(`patchee.`, ccName, ` = nil`)
			p.P(`continue`)
			p.P(`}`)
			p.P(`if patchee.`, ccName, ` == nil {`)
			p.P(`patchee.`, ccName, ` = &`, strings.TrimPrefix(fieldType, "*"), `{}`)
			p.P(`}`)
			if s := strings.Split(fieldType, "."); len(s) == 2 {
				p.P(`if o, err := `, strings.TrimLeft(s[0], "*"), `.DefaultApplyFieldMask`, s[1], `(ctx, patchee.`, ccName,
					`, patcher.`, ccName, `, &`, p.Import(fmImport),
					`.FieldMask{Paths:updateMask.Paths[i:]}, prefix+"`, ccName, `.", db); err != nil {`)
			} else {
				p.P(`if o, err := DefaultApplyFieldMask`, strings.TrimPrefix(fieldType, "*"), `(ctx, patchee.`, ccName,
					`, patcher.`, ccName, `, &`, p.Import(fmImport),
					`.FieldMask{Paths:updateMask.Paths[i:]}, prefix+"`, ccName, `.", db); err != nil {`)
			}
			p.P(`return nil, err`)
			p.P(`} else {`)
			p.P(`patchee.`, ccName, ` = o`)
			p.P(`}`)
			p.P(`continue`)
			p.P(`}`)
			p.P(`if f == prefix+"`, ccName, `" {`)
			p.P(`updated`, ccName, ` = true`)
			p.P(`patchee.`, ccName, ` = patcher.`, ccName)
			p.P(`continue`)
			p.P(`}`)
		} else if field.IsMessage() && !isSpecialType(fieldType) && !field.IsRepeated() {
			p.P(`if strings.HasPrefix(f, prefix+"`, ccName, `.") && !updated`, ccName, ` {`)
			p.P(`if patcher.`, ccName, ` == nil {`)
			p.P(`patchee.`, ccName, ` = nil`)
			p.P(`continue`)
			p.P(`}`)
			p.P(`if patchee.`, ccName, ` == nil {`)
			p.P(`patchee.`, ccName, ` = &`, strings.TrimPrefix(fieldType, "*"), `{}`)
			p.P(`}`)
			p.P(`childMask := &`, p.Import(fmImport), `.FieldMask{}`)
			p.P(`for j := i; j < len(updateMask.Paths); j++ {`)
			p.P(`if trimPath := strings.TrimPrefix(updateMask.Paths[j], prefix+"`, ccName, `."); trimPath != updateMask.Paths[j] {`)
			p.P(`childMask.Paths = append(childMask.Paths, trimPath)`)
			p.P(`}`)
			p.P(`}`)
			p.P(`if err := `, p.Import(tkgormImport), `.MergeWithMask(patcher.`, ccName, `, patchee.`, ccName, `, childMask); err != nil {`)
			p.P(`return nil, nil`)
			p.P(`}`)
			p.P(`}`)
			p.P(`if f == prefix+"`, ccName, `" {`)
			p.P(`updated`, ccName, ` = true`)
			p.P(`patchee.`, ccName, ` = patcher.`, ccName)
			p.P(`continue`)
			p.P(`}`)
		} else {
			p.P(`if f == prefix+"`, ccName, `" {`)
			p.P(`patchee.`, ccName, ` = patcher.`, ccName)
			p.P(`continue`)
			p.P(`}`)
		}
	}
	p.P(`}`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`return patchee, nil`)
	p.P(`}`)
	p.P()
}

func (p *OrmPlugin) hasIDField(message *generator.Descriptor) bool {
	for _, field := range message.GetField() {
		if strings.ToLower(field.GetName()) == "id" {
			return true
		}
	}

	return false
}

func (p *OrmPlugin) generatePatchHandler(message *generator.Descriptor) {
	var isMultiAccount bool

	typeName := p.TypeName(message)
	ormable := p.getOrmable(typeName)

	if opts := getMessageOptions(message); opts != nil && opts.GetMultiAccount() {
		isMultiAccount = true
	}

	if isMultiAccount && !p.hasIDField(message) {
		p.P(fmt.Sprintf("// Cannot autogen DefaultPatch%s: this is a multi-account table without an \"id\" field in the message.\n", typeName))
		return
	}

	p.P(`// DefaultPatch`, typeName, ` executes a basic gorm update call with patch behavior`)
	p.P(`func DefaultPatch`, typeName, `(ctx context.Context, in *`,
		typeName, `, updateMask *`, p.Import(fmImport), `.FieldMask, db *`, p.Import(gormImport), `.DB) (*`, typeName, `, error) {`)

	p.P(`if in == nil {`)
	p.P(`return nil, errors.New("Nil argument to DefaultPatch`, typeName, `")`)
	p.P(`}`)
	p.P(`var pbObj `, typeName)
	p.P(`var err error`)
	p.generateBeforePatchHookCall(ormable, "Read")
	if p.readHasFieldSelection(ormable) {
		p.P(`pbReadRes, err := DefaultRead`, typeName, `(ctx, &`, typeName, `{Id: in.GetId()}, db, nil)`)
	} else {
		p.P(`pbReadRes, err := DefaultRead`, typeName, `(ctx, &`, typeName, `{Id: in.GetId()}, db)`)
	}

	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)

	p.P(`pbObj = *pbReadRes`)

	p.generateBeforePatchHookCall(ormable, "ApplyFieldMask")
	p.P(`if _, err := DefaultApplyFieldMask`, typeName, `(ctx, &pbObj, in, updateMask, "", db); err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)

	p.generateBeforePatchHookCall(ormable, "Save")
	p.P(`pbResponse, err := DefaultStrictUpdate`, typeName, `(ctx, &pbObj, db)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.generateAfterPatchHookCall(ormable, "Save")

	p.P(`return pbResponse, nil`)
	p.P(`}`)

	p.generateBeforePatchHookDef(ormable, "Read")
	p.generateBeforePatchHookDef(ormable, "ApplyFieldMask")
	p.generateBeforePatchHookDef(ormable, "Save")
	p.generateAfterPatchHookDef(ormable, "Save")
}

func (p *OrmPlugin) generateBeforePatchHookDef(orm *OrmableType, suffix string) {
	p.P(`type `, orm.OriginName, `WithBeforePatch`, suffix, ` interface {`)
	p.P(`BeforePatch`, suffix, `(context.Context, *`, orm.OriginName, `, *`, p.Import(fmImport), `.FieldMask, *`, p.Import(gormImport),
		`.DB) (*`, p.Import(gormImport), `.DB, error)`)
	p.P(`}`)
}

func (p *OrmPlugin) generateBeforePatchHookCall(orm *OrmableType, suffix string) {
	p.P(`if hook, ok := interface{}(&pbObj).(`, orm.OriginName, `WithBeforePatch`, suffix, `); ok {`)
	p.P(`if db, err = hook.BeforePatch`, suffix, `(ctx, in, updateMask, db); err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generateAfterPatchHookDef(orm *OrmableType, suffix string) {
	p.P(`type `, orm.OriginName, `WithAfterPatch`, suffix, ` interface {`)
	p.P(`AfterPatch`, suffix, `(context.Context, *`, orm.OriginName, `, *`, p.Import(fmImport), `.FieldMask, *`, p.Import(gormImport),
		`.DB) error`)
	p.P(`}`)
}

func (p *OrmPlugin) generateAfterPatchHookCall(orm *OrmableType, suffix string) {
	p.P(`if hook, ok := interface{}(pbResponse).(`, orm.OriginName, `WithAfterPatch`, suffix, `); ok {`)
	p.P(`if err = hook.AfterPatch`, suffix, `(ctx, in, updateMask, db); err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generateDeleteHandler(message *generator.Descriptor) {
	typeName := p.TypeName(message)
	p.P(`func DefaultDelete`, typeName, `(ctx context.Context, in *`,
		typeName, `, db *`, p.Import(gormImport), `.DB) error {`)
	p.P(`if in == nil {`)
	p.P(`return errors.New("Nil argument to DefaultDelete`, typeName, `")`)
	p.P(`}`)
	p.P(`ormObj, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return err`)
	p.P(`}`)
	ormable := p.getOrmable(typeName)
	pkName, pk := p.findPrimaryKey(ormable)
	if strings.Contains(pk.Type, "*") {
		p.P(`if ormObj.`, pkName, ` == nil || *ormObj.`, pkName, ` == `, p.guessZeroValue(pk.Type), ` {`)
	} else {
		p.P(`if ormObj.`, pkName, ` == `, p.guessZeroValue(pk.Type), `{`)
	}
	p.P(`return errors.New("A non-zero ID value is required for a delete call")`)
	p.P(`}`)
	p.generateBeforeDeleteHookCall(ormable)
	p.P(`err = db.Where(&ormObj).Delete(&`, ormable.Name, `{}).Error`)
	p.P(`if err != nil {`)
	p.P(`return err`)
	p.P(`}`)
	p.generateAfterDeleteHookCall(ormable)
	p.P(`return err`)
	p.P(`}`)
	p.generateBeforeHookDef(ormable, "Delete")
	p.generateAfterHookDef(ormable, "Delete")
}

func (p *OrmPlugin) generateBeforeDeleteHookCall(orm *OrmableType) {
	p.P(`if hook, ok := interface{}(&ormObj).(`, orm.Name, `WithBeforeDelete); ok {`)
	p.P(`if db, err = hook.BeforeDelete(ctx, db); err != nil {`)
	p.P(`return err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generateAfterDeleteHookCall(orm *OrmableType) {
	p.P(`if hook, ok := interface{}(&ormObj).(`, orm.Name, `WithAfterDelete); ok {`)
	p.P(`err = hook.AfterDelete(ctx, db)`)
	p.P(`}`)
}

func (p *OrmPlugin) generateListHandler(message *generator.Descriptor) {
	typeName := p.TypeName(message)
	ormable := p.getOrmable(typeName)

	p.P(`// DefaultList`, typeName, ` executes a gorm list call`)
	listSign := fmt.Sprint(`func DefaultList`, typeName, `(ctx context.Context, db *`, p.Import(gormImport), `.DB`)
	var f, s, pg, fs string
	if p.listHasFiltering(ormable) {
		listSign += fmt.Sprint(`, f `, `*`, p.Import(queryImport), `.Filtering`)
		f = "f"
	} else {
		f = "nil"
	}
	if p.listHasSorting(ormable) {
		listSign += fmt.Sprint(`, s `, `*`, p.Import(queryImport), `.Sorting`)
		s = "s"
	} else {
		s = "nil"
	}
	if p.listHasPagination(ormable) {
		listSign += fmt.Sprint(`, p `, `*`, p.Import(queryImport), `.Pagination`)
		pg = "p"
	} else {
		pg = "nil"
	}
	if p.listHasFieldSelection(ormable) {
		listSign += fmt.Sprint(`, fs `, `*`, p.Import(queryImport), `.FieldSelection`)
		fs = "fs"
	} else {
		fs = "nil"
	}
	listSign += fmt.Sprint(`) ([]*`, typeName, `, error) {`)
	p.P(listSign)
	p.P(`in := `, typeName, `{}`)
	p.P(`ormObj, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.generateBeforeListHookCall(ormable, "ApplyQuery")
	p.P(`db, err = `, p.Import(tkgormImport), `.ApplyCollectionOperators(ctx, db, &`, ormable.Name, `{}, &`, typeName, `{}, `, f, `,`, s, `,`, pg, `,`, fs, `)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.generateBeforeListHookCall(ormable, "Find")
	p.P(`db = db.Where(&ormObj)`)

	// add default ordering by primary key
	if p.hasPrimaryKey(ormable) {
		pkName, pk := p.findPrimaryKey(ormable)
		column := pk.GetTag().GetColumn()
		if len(column) == 0 {
			column = jgorm.ToDBName(pkName)
		}
		p.P(`db = db.Order("`, column, `")`)
	}

	p.P(`ormResponse := []`, ormable.Name, `{}`)
	p.P(`if err := db.Find(&ormResponse).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.generateAfterListHookCall(ormable)
	p.P(`pbResponse := []*`, typeName, `{}`)
	p.P(`for _, responseEntry := range ormResponse {`)
	p.P(`temp, err := responseEntry.ToPB(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse = append(pbResponse, &temp)`)
	p.P(`}`)
	p.P(`return pbResponse, nil`)
	p.P(`}`)
	p.generateBeforeListHookDef(ormable, "ApplyQuery")
	p.generateBeforeListHookDef(ormable, "Find")
	p.generateAfterListHookDef(ormable)
}

func (p *OrmPlugin) generateBeforeListHookDef(orm *OrmableType, suffix string) {
	p.P(`type `, orm.Name, `WithBeforeList`, suffix, ` interface {`)
	hookSign := fmt.Sprint(`BeforeList`, suffix, `(context.Context, *`, p.Import(gormImport), `.DB`)
	if p.listHasFiltering(orm) {
		hookSign += fmt.Sprint(`, *`, p.Import(queryImport), `.Filtering`)
	}
	if p.listHasSorting(orm) {
		hookSign += fmt.Sprint(`, *`, p.Import(queryImport), `.Sorting`)
	}
	if p.listHasPagination(orm) {
		hookSign += fmt.Sprint(`, *`, p.Import(queryImport), `.Pagination`)
	}
	if p.listHasFieldSelection(orm) {
		hookSign += fmt.Sprint(`, *`, p.Import(queryImport), `.FieldSelection`)
	}
	hookSign += fmt.Sprint(`) (*`, p.Import(gormImport), `.DB, error)`)
	p.P(hookSign)
	p.P(`}`)
}

func (p *OrmPlugin) generateAfterListHookDef(orm *OrmableType) {
	p.P(`type `, orm.Name, `WithAfterListFind interface {`)
	hookSign := fmt.Sprint(`AfterListFind(context.Context, *`, p.Import(gormImport), `.DB, *[]`, orm.Name)
	if p.listHasFiltering(orm) {
		hookSign += fmt.Sprint(`, *`, p.Import(queryImport), `.Filtering`)
	}
	if p.listHasSorting(orm) {
		hookSign += fmt.Sprint(`, *`, p.Import(queryImport), `.Sorting`)
	}
	if p.listHasPagination(orm) {
		hookSign += fmt.Sprint(`, *`, p.Import(queryImport), `.Pagination`)
	}
	if p.listHasFieldSelection(orm) {
		hookSign += fmt.Sprint(`, *`, p.Import(queryImport), `.FieldSelection`)
	}
	hookSign += fmt.Sprint(`) error`)
	p.P(hookSign)
	p.P(`}`)
}

func (p *OrmPlugin) generateBeforeListHookCall(orm *OrmableType, suffix string) {
	p.P(`if hook, ok := interface{}(&ormObj).(`, orm.Name, `WithBeforeList`, suffix, `); ok {`)
	hookCall := fmt.Sprint(`if db, err = hook.BeforeList`, suffix, `(ctx, db`)
	if p.listHasFiltering(orm) {
		hookCall += `,f`
	}
	if p.listHasSorting(orm) {
		hookCall += `,s`
	}
	if p.listHasPagination(orm) {
		hookCall += `,p`
	}
	if p.listHasFieldSelection(orm) {
		hookCall += `,fs`
	}
	hookCall += `); err != nil {`
	p.P(hookCall)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generateAfterListHookCall(orm *OrmableType) {
	p.P(`if hook, ok := interface{}(&ormObj).(`, orm.Name, `WithAfterListFind); ok {`)
	hookCall := fmt.Sprint(`if err = hook.AfterListFind(ctx, db, &ormResponse`)
	if p.listHasFiltering(orm) {
		hookCall += `,f`
	}
	if p.listHasSorting(orm) {
		hookCall += `,s`
	}
	if p.listHasPagination(orm) {
		hookCall += `,p`
	}
	if p.listHasFieldSelection(orm) {
		hookCall += `,fs`
	}
	hookCall += `); err != nil {`
	p.P(hookCall)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`}`)
}

func (p *OrmPlugin) generateStrictUpdateHandler(message *generator.Descriptor) {
	typeName := p.TypeName(message)
	p.P(`// DefaultStrictUpdate`, typeName, ` clears first level 1:many children and then executes a gorm update call`)
	p.P(`func DefaultStrictUpdate`, typeName, `(ctx context.Context, in *`,
		typeName, `, db *`, p.Import(gormImport), `.DB) (*`, typeName, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, fmt.Errorf("Nil argument to DefaultStrictUpdate`, typeName, `")`)
	p.P(`}`)
	p.P(`ormObj, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)

	p.P(`count := 1`)
	// add default ordering by primary key
	ormable := p.getOrmable(typeName)
	if p.hasPrimaryKey(ormable) {
		pkName, pk := p.findPrimaryKey(ormable)
		column := pk.GetTag().GetColumn()
		if len(column) == 0 {
			column = jgorm.ToDBName(pkName)
		}
		p.P(`err = db.Model(&ormObj).Where("`, column, `=?", ormObj.`, pkName, `).Count(&count).Error`)
		p.P(`if err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
	}
	p.generateBeforeHookCall(ormable, "StrictUpdateCleanup")
	p.removeChildAssociations(message)
	p.generateBeforeHookCall(ormable, "StrictUpdateSave")
	if getMessageOptions(message).GetMultiAccount() {
		p.P(`db = db.Where(&`, typeName, `ORM{AccountID: ormObj.AccountID})`)
	}
	p.P(`if err = db.Save(&ormObj).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.generateAfterHookCall(ormable, "StrictUpdateSave")
	p.P(`pbResponse, err := ormObj.ToPB(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)

	p.P(`if count == 0 {`)
	p.P(`err = `, p.Import(gatewayImport), `.SetCreated(ctx, "")`)
	p.P(`}`)

	p.P(`return &pbResponse, err`)
	p.P(`}`)
	p.generateBeforeHookDef(ormable, "StrictUpdateCleanup")
	p.generateBeforeHookDef(ormable, "StrictUpdateSave")
	p.generateAfterHookDef(ormable, "StrictUpdateSave")
}

func (p *OrmPlugin) isFieldOrmable(message *generator.Descriptor, fieldName string) bool {
	_, ok := p.getOrmable(p.TypeName(message)).Fields[fieldName]
	return ok
}

func (p *OrmPlugin) removeChildAssociations(message *generator.Descriptor) {
	ormable := p.getOrmable(p.TypeName(message))
	for _, fieldName := range p.getSortedFieldNames(ormable.Fields) {
		p.removeChildAssociationsByName(message, fieldName)
	}
}

func (p *OrmPlugin) removeChildAssociationsByName(message *generator.Descriptor, fieldName string) {
	ormable := p.getOrmable(p.TypeName(message))
	field := ormable.Fields[fieldName]

	if field == nil {
		return
	}

	if field.GetHasMany() != nil || field.GetHasOne() != nil {
		var assocKeyName, foreignKeyName string
		switch {
		case field.GetHasMany() != nil:
			assocKeyName = field.GetHasMany().GetAssociationForeignkey()
			foreignKeyName = field.GetHasMany().GetForeignkey()
		case field.GetHasOne() != nil:
			assocKeyName = field.GetHasOne().GetAssociationForeignkey()
			foreignKeyName = field.GetHasOne().GetForeignkey()
		}
		assocKeyType := ormable.Fields[assocKeyName].Type
		assocOrmable := p.getOrmable(field.Type)
		foreignKeyType := assocOrmable.Fields[foreignKeyName].Type
		p.P(`filter`, fieldName, ` := `, strings.Trim(field.Type, "[]*"), `{}`)
		zeroValue := p.guessZeroValue(assocKeyType)
		if strings.Contains(assocKeyType, "*") {
			p.P(`if ormObj.`, assocKeyName, ` == nil || *ormObj.`, assocKeyName, ` == `, zeroValue, `{`)
		} else {
			p.P(`if ormObj.`, assocKeyName, ` == `, zeroValue, `{`)
		}
		p.P(`return nil, errors.New("Can't do overwriting update with no `, assocKeyName, ` value for `, ormable.Name, `")`)
		p.P(`}`)
		filterDesc := "filter" + fieldName + "." + foreignKeyName
		ormDesc := "ormObj." + assocKeyName
		if strings.HasPrefix(foreignKeyType, "*") {
			p.P(filterDesc, ` = new(`, strings.TrimPrefix(foreignKeyType, "*"), `)`)
			filterDesc = "*" + filterDesc
		}
		if strings.HasPrefix(assocKeyType, "*") {
			ormDesc = "*" + ormDesc
		}
		p.P(filterDesc, " = ", ormDesc)
		if _, ok := assocOrmable.Fields["AccountID"]; ok {
			p.P(`filter`, fieldName, `.AccountID = ormObj.AccountID`)
		}
		p.P(`if err = db.Where(filter`, fieldName, `).Delete(`, strings.Trim(field.Type, "[]*"), `{}).Error; err != nil {`)
		p.P(`return nil, err`)
		p.P(`}`)
	}
}

// guessZeroValue of the input type, so that we can check if a (key) value is set or not
func (p *OrmPlugin) guessZeroValue(typeName string) string {
	typeName = strings.ToLower(typeName)
	if strings.Contains(typeName, "string") {
		return `""`
	}
	if strings.Contains(typeName, "int") {
		return `0`
	}
	if strings.Contains(typeName, "uuid") {
		return fmt.Sprintf(`%s.Nil`, p.Import(uuidImport))
	}
	if strings.Contains(typeName, "[]byte") {
		return `nil`
	}
	if strings.Contains(typeName, "bool") {
		return `false`
	}
	return ``
}

func (p *OrmPlugin) readHasFieldSelection(ormable *OrmableType) bool {
	if read, ok := ormable.Methods[readService]; ok {
		if s := p.getFieldSelection(read.inType); s != "" {
			return true
		}
	}
	return false
}

func (p *OrmPlugin) listHasFiltering(ormable *OrmableType) bool {
	if read, ok := ormable.Methods[listService]; ok {
		if s := p.getFiltering(read.inType); s != "" {
			return true
		}
	}
	return false
}

func (p *OrmPlugin) listHasSorting(ormable *OrmableType) bool {
	if read, ok := ormable.Methods[listService]; ok {
		if s := p.getSorting(read.inType); s != "" {
			return true
		}
	}
	return false
}

func (p *OrmPlugin) listHasPagination(ormable *OrmableType) bool {
	if read, ok := ormable.Methods[listService]; ok {
		if s := p.getPagination(read.inType); s != "" {
			return true
		}
	}
	return false
}

func (p *OrmPlugin) listHasFieldSelection(ormable *OrmableType) bool {
	if read, ok := ormable.Methods[listService]; ok {
		if s := p.getFieldSelection(read.inType); s != "" {
			return true
		}
	}
	return false
}
