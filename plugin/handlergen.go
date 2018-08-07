package plugin

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	jgorm "github.com/jinzhu/gorm"
)

var isCollectionFetcherGenerated bool

func (p *OrmPlugin) generateCollectionOperatorsFetcher() {
	if isCollectionFetcherGenerated == true {
		return
	}
	isCollectionFetcherGenerated = true
	p.P(`// getCollectionOperators takes collection operator values from corresponding message fields`)
	p.P(`func getCollectionOperators(in interface{}) (*`, p.Import(queryImport), `.Filtering, *`, p.Import(queryImport), `.Sorting, *`, p.Import(queryImport), `.Pagination, *`, p.Import(queryImport), `.FieldSelection, error) {`)
	p.P(`f := &`, p.Import(queryImport), `.Filtering{}`)
	p.P(`err := `, p.Import(gatewayImport), `.GetCollectionOp(in, f)`)
	p.P(`if err != nil {`)
	p.P(`return nil, nil, nil, nil, err`)
	p.P(`}`)

	p.P(`s := &`, p.Import(queryImport), `.Sorting{}`)
	p.P(`err = `, p.Import(gatewayImport), `.GetCollectionOp(in, s)`)
	p.P(`if err != nil {`)
	p.P(`return nil, nil, nil, nil, err`)
	p.P(`}`)

	p.P(`p := &`, p.Import(queryImport), `.Pagination{}`)
	p.P(`err = `, p.Import(gatewayImport), `.GetCollectionOp(in, p)`)
	p.P(`if err != nil {`)
	p.P(`return nil, nil, nil, nil, err`)
	p.P(`}`)

	p.P(`fs := &`, p.Import(queryImport), `.FieldSelection{}`)
	p.P(`err = `, p.Import(gatewayImport), `.GetCollectionOp(in, fs)`)
	p.P(`if err != nil {`)
	p.P(`return nil, nil, nil, nil, err`)
	p.P(`}`)

	p.P(`return f, s, p, fs, nil`)
	p.P(`}`)
}

func (p *OrmPlugin) generateDefaultHandlers(file *generator.FileDescriptor) {
	for _, message := range file.Messages() {
		if getMessageOptions(message).GetOrmable() {
			p.UsingGoImports("context", "errors")

			p.generateCreateHandler(message)
			// FIXME: Temporary fix for Ormable objects that have no ID field but
			// have pk.
			if p.hasPrimaryKey(p.getOrmable(p.TypeName(message))) && p.hasIDField(message) {
				p.generateReadHandler(message)
				p.generateUpdateHandler(message)
				p.generateDeleteHandler(message)
				p.generateStrictUpdateHandler(message)
				p.generatePatchHandler(message)
				p.generateApplyFieldMask(message)
			}
			p.generateCollectionOperatorsFetcher()
			p.generateListHandler(message)
		}
	}
}

func (p *OrmPlugin) generateCreateHandler(message *generator.Descriptor) {
	typeName := p.TypeName(message)
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
	p.setupOrderedHasMany(message)
	p.P(`if err = db.Create(&ormObj).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse, err := ormObj.ToPB(ctx)`)
	p.P(`return &pbResponse, err`)
	p.P(`}`)
	p.P()
}

func (p *OrmPlugin) generateReadHandler(message *generator.Descriptor) {
	typeName := p.TypeName(message)
	ormable := p.getOrmable(typeName)
	p.P(`// DefaultRead`, typeName, ` executes a basic gorm read call`)
	p.P(`func DefaultRead`, typeName, `(ctx context.Context, in *`,
		typeName, `, db *`, p.Import(gormImport), `.DB, preload bool) (*`, typeName, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, errors.New("Nil argument to DefaultRead`, typeName, `")`)
	p.P(`}`)
	p.P(`if preload {`)
	p.generatePreloading()
	p.P(`}`)

	p.P(`ormParams, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)

	p.sortOrderedHasMany(message)
	p.P(`ormResponse := `, ormable.Name, `{}`)
	p.P(`if err = db.Where(&ormParams).First(&ormResponse).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse, err := ormResponse.ToPB(ctx)`)
	p.P(`return &pbResponse, err`)
	p.P(`}`)
	p.P()
}

func (p *OrmPlugin) generateApplyFieldMask(message *generator.Descriptor) {
	typeName := p.TypeName(message)

	p.P(`// DefaultApplyFieldMask`, typeName, ` patches an pbObject with patcher according to a field mask.`)
	p.P(`func DefaultApplyFieldMask`, typeName, `(ctx context.Context, patchee *`,
		typeName, `,ormObj *`, typeName, `ORM, patcher *`,
		typeName, `, updateMask *`, p.Import(fmImport),
		`.FieldMask, db *`, p.Import(gormImport), `.DB) (*`, typeName, `, error) {`)
	p.P(`var err error`)

	// Patch pbObj with input according to a field mask.
	p.P(`for _, f := range updateMask.GetPaths() {`)
	for _, field := range message.GetField() {
		ccName := generator.CamelCase(field.GetName())
		p.P(`if f == "`, ccName, `" {`)
		p.P(`patchee.`, ccName, ` = patcher.`, ccName)
		p.removeChildAssociationsByName(message, ccName)
		p.setupOrderedHasManyByName(message, ccName)
		p.P(`}`)
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
	//ormable := p.getOrmable(typeName)

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

	if isMultiAccount {
		p.P("accountID, err := ", p.Import(authImport), ".GetAccountID(ctx, nil)")
		p.P("if err != nil {")
		p.P("return nil, err")
		p.P("}")
	}

	p.P(`pbReadRes, err := DefaultRead`, typeName, `(ctx, &`, typeName, `{Id: in.GetId()}, db, true)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)

	p.P(`pbObj := *pbReadRes`)

	p.P(`ormObj, err := pbObj.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)

	p.P(`if _, err := DefaultApplyFieldMask`, typeName, `(ctx, &pbObj, &ormObj, in, updateMask, db); err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)

	p.P(`if hook, ok := interface{}(&pbObj).(`, typeName, `WithBeforePatchSave); ok {`)
	p.P(`if ctx, db, err = hook.BeforePatchSave(ctx, in, updateMask, db); err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`}`)

	// Convert pbObj back to ormObj to trigger any logic that was
	// written for BeforeToORM/AfterToORM and perform db.Save call.
	p.P(`ormObj, err = pbObj.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)

	if isMultiAccount {
		p.P(`db = db.Where(&`, typeName, `ORM{AccountID: accountID})`)
	}

	p.P(`if err = db.Save(&ormObj).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)

	// convert ormObj to pbObj again (sic!) to trigger any logic that was
	// written for AfterToPB/BeforeToPB.
	p.P(`pbObj, err = ormObj.ToPB(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)

	p.P(`return &pbObj, err`)
	p.P(`}`)
	p.P()

	p.P(`type `, typeName, `WithBeforePatchSave interface {`)
	p.P(`BeforePatchSave(context.Context, *`,
		typeName, `, *`, p.Import(fmImport), `.FieldMask, *`, p.Import(gormImport),
		`.DB) (context.Context, *`, p.Import(gormImport), `.DB, error)`)
	p.P(`}`)
}

func (p *OrmPlugin) generateUpdateHandler(message *generator.Descriptor) {
	typeName := p.TypeName(message)

	isMultiAccount := false
	if opts := getMessageOptions(message); opts != nil && opts.GetMultiAccount() {
		isMultiAccount = true
	}

	if isMultiAccount && !p.hasIDField(message) {
		p.P(fmt.Sprintf("// Cannot autogen DefaultUpdate%s: this is a multi-account table without an \"id\" field in the message.\n", typeName))
		return
	}

	p.P(`// DefaultUpdate`, typeName, ` executes a basic gorm update call`)
	p.P(`func DefaultUpdate`, typeName, `(ctx context.Context, in *`,
		typeName, `, db *`, p.Import(gormImport), `.DB) (*`, typeName, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, errors.New("Nil argument to DefaultUpdate`, typeName, `")`)
	p.P(`}`)
	if isMultiAccount {
		p.P("accountID, err := ", p.Import(authImport), ".GetAccountID(ctx, nil)")
		p.P("if err != nil {")
		p.P("return nil, err")
		p.P("}")
		p.P(fmt.Sprintf("if exists, err := DefaultRead%s(ctx, &%s{Id: in.GetId()}, db, true); err != nil {",
			typeName, typeName))
		p.P("return nil, err")
		p.P("} else if exists == nil {")
		p.P(fmt.Sprintf("return nil, errors.New(\"%s not found\")", typeName))
		p.P("}")
	}

	p.P(`ormObj, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	if isMultiAccount {
		p.P(`ormObj.AccountID = accountID`)
		p.P(`db = db.Where(&`, typeName, `ORM{AccountID: accountID})`)
	}
	p.setupOrderedHasMany(message)
	p.P(`if err = db.Save(&ormObj).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse, err := ormObj.ToPB(ctx)`)
	p.P(`return &pbResponse, err`)
	p.P(`}`)
	p.P()
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
	p.P(`err = db.Where(&ormObj).Delete(&`, ormable.Name, `{}).Error`)
	p.P(`return err`)
	p.P(`}`)
	p.P()
}

func (p *OrmPlugin) generateListHandler(message *generator.Descriptor) {
	typeName := p.TypeName(message)
	ormable := p.getOrmable(typeName)

	p.P(`// DefaultList`, typeName, ` executes a gorm list call`)
	p.P(`func DefaultList`, typeName, `(ctx context.Context, db *`, p.Import(gormImport), ``,
		`.DB, req interface{}) ([]*`, typeName, `, error) {`)
	p.P(`ormResponse := []`, ormable.Name, `{}`)

	p.P(`f, s, p, fs, err := getCollectionOperators(req)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)

	p.P(`db, err = `, p.Import(tkgormImport), `.ApplyCollectionOperators(db, &`, ormable.Name, `{}, f, s, p, fs)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`if fs.GetFields() == nil {`)
	p.generatePreloading()
	p.P(`}`)
	p.P(`in := `, typeName, `{}`)
	p.P(`ormParams, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`db = db.Where(&ormParams)`)
	p.sortOrderedHasMany(message)

	// add default ordering by primary key
	if p.hasPrimaryKey(ormable) {
		pkName, pk := p.findPrimaryKey(ormable)
		column := pk.GetTag().GetColumn()
		if len(column) == 0 {
			column = jgorm.ToDBName(pkName)
		}
		p.P(`db = db.Order("`, column, `")`)
	}

	p.P(`if err := db.Find(&ormResponse).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
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
	p.P()
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

	p.removeChildAssociations(message)
	if getMessageOptions(message).GetMultiAccount() {
		p.P(`db = db.Where(&`, typeName, `ORM{AccountID: ormObj.AccountID})`)
	}
	p.setupOrderedHasMany(message)
	p.P(`if err = db.Save(&ormObj).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse, err := ormObj.ToPB(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)

	p.P(`if count == 0 {`)
	p.P(`err = `, p.Import(gatewayImport), `.SetCreated(ctx, "")`)
	p.P(`}`)

	p.P(`return &pbResponse, err`)
	p.P(`}`)
	p.P()
}

func (p *OrmPlugin) generatePreloading() {
	p.P(`db = db.Set("gorm:auto_preload", true)`)
}

func (p *OrmPlugin) setupOrderedHasMany(message *generator.Descriptor) {
	ormable := p.getOrmable(p.TypeName(message))
	for _, fieldName := range p.getSortedFieldNames(ormable.Fields) {
		p.setupOrderedHasManyByName(message, fieldName)
	}
}

func (p *OrmPlugin) setupOrderedHasManyByName(message *generator.Descriptor, fieldName string) {
	ormable := p.getOrmable(p.TypeName(message))
	field := ormable.Fields[fieldName]

	if field == nil {
		return
	}

	if field.GetHasMany().GetPositionField() != "" {
		positionField := field.GetHasMany().GetPositionField()
		positionFieldType := p.getOrmable(field.Type).Fields[positionField].Type
		p.P(`for i, e := range `, `ormObj.`, fieldName, `{`)
		p.P(`e.`, positionField, ` = `, positionFieldType, `(i)`)
		p.P(`}`)
	}
}

func (p *OrmPlugin) sortOrderedHasMany(message *generator.Descriptor) {
	ormable := p.getOrmable(p.TypeName(message))
	for _, fieldName := range p.getSortedFieldNames(ormable.Fields) {
		field := ormable.Fields[fieldName]
		if field.GetHasMany().GetPositionField() != "" {
			positionField := field.GetHasMany().GetPositionField()
			p.P(`db = db.Preload("`, fieldName, `", func(db *`, p.Import(gormImport), `.DB) *`, p.Import(gormImport), `.DB {`)
			p.P(`return db.Order("`, jgorm.ToDBName(positionField), `")`)
			p.P(`})`)
		}
	}
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
