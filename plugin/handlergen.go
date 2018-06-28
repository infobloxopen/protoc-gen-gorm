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
			if p.hasPrimaryKey(p.getOrmable(p.TypeName(message))) {
				p.generateReadHandler(message)
				p.generateUpdateHandler(message)
				p.generateDeleteHandler(message)
				p.generateStrictUpdateHandler(message)
				p.generatePatchHandler(message)
			}
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
		typeName, `, db *`, p.Import(gormImport), `.DB) (*`, typeName, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, errors.New("Nil argument to DefaultRead`, typeName, `")`)
	p.P(`}`)
	p.P(`ormParams, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.sortOrderedHasMany(message)
	p.generatePreloading(ormable)
	p.P(`ormResponse := `, ormable.Name, `{}`)
	p.P(`if err = db.Where(&ormParams).First(&ormResponse).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse, err := ormResponse.ToPB(ctx)`)
	p.P(`return &pbResponse, err`)
	p.P(`}`)
	p.P()
}

func (p *OrmPlugin) generatePatchHandler(message *generator.Descriptor) {
	typeName := p.TypeName(message)

	hasIDField := false
	for _, field := range message.GetField() {
		if strings.ToLower(field.GetName()) == "id" {
			hasIDField = true
			break
		}
	}
	isMultiAccount := false
	if opts := getMessageOptions(message); opts != nil && opts.GetMultiAccount() {
		isMultiAccount = true
	}

	if isMultiAccount && !hasIDField {
		p.P(fmt.Sprintf("// Cannot autogen DefaultPatch%s: this is a multi-account table without an \"id\" field in the message.\n", typeName))
		return
	}

	p.P(`// DefaultPatch`, typeName, ` executes a basic gorm update call with patch behavior`)
	p.P(`func DefaultPatch`, typeName, `(ctx context.Context, in *`,
		typeName, `, fieldMask []string, db *`, p.Import(gormImport), `.DB) (*`, typeName, `, error) {`)

	p.P(`if in == nil {`)
	p.P(`return nil, errors.New("Nil argument to DefaultUpdate`, typeName, `")`)
	p.P(`}`)
	if isMultiAccount {
		p.P("accountID, err := ", p.Import(authImport), ".GetAccountID(ctx, nil)")
		p.P("if err != nil {")
		p.P("return nil, err")
		p.P("}")
	}

	p.P(`var ormObj `, typeName, `ORM`)
	p.P(fmt.Sprintf("if ormObj, err := DefaultRead%s(ctx, &%s{Id: in.GetId()}, db); err != nil {",
		typeName, typeName))
	p.P("return nil, err")
	p.P("} else if ormObj == nil {")
	p.P(fmt.Sprintf("return nil, errors.New(\"%s not found\")", typeName))
	p.P("}")

	p.P(`patcher, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)

	p.P(`for _, f := range fieldMask {`)
	for _, field := range message.GetField() {

		ccName := generator.CamelCase(field.GetName())

		if field.GetName() == "id" || !p.isFieldOrmable(message, ccName) {
			continue
		}

		p.P(`if f == "`, field.GetName(), `" {`)
		p.removeChildAssociationsByName(message, ccName)
		p.setupOrderedHasManyByName(message, ccName)
		p.P(`ormObj.`, ccName, ` = patcher.`, ccName)
		p.P(`}`)
	}
	p.P(`}`)

	if getMessageOptions(message).GetMultiAccount() {
		p.P(`db = db.Where(&`, typeName, `ORM{AccountID: accountID})`)
	}
	p.P(`if err = db.Save(&ormObj).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse, err := ormObj.ToPB(ctx)`)
	p.P(`return &pbResponse, err`)
	p.P(`}`)
	p.P()
}

func (p *OrmPlugin) generateUpdateHandler(message *generator.Descriptor) {
	typeName := p.TypeName(message)

	hasIDField := false
	for _, field := range message.GetField() {
		if strings.ToLower(field.GetName()) == "id" {
			hasIDField = true
			break
		}
	}
	isMultiAccount := false
	if opts := getMessageOptions(message); opts != nil && opts.GetMultiAccount() {
		isMultiAccount = true
	}

	if isMultiAccount && !hasIDField {
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
		p.P(fmt.Sprintf("if exists, err := DefaultRead%s(ctx, &%s{Id: in.GetId()}, db); err != nil {",
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
		`.DB) ([]*`, typeName, `, error) {`)
	p.P(`ormResponse := []`, ormable.Name, `{}`)
	p.P(`db, err := `, p.Import(tkgormImport), `.ApplyCollectionOperators(db, ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`in := `, typeName, `{}`)
	p.P(`ormParams, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`db = db.Where(&ormParams)`)
	p.sortOrderedHasMany(message)
	p.generatePreloading(ormable)

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
	p.P(`return nil, fmt.Errorf("Nil argument to DefaultCascadedUpdate`, typeName, `")`)
	p.P(`}`)
	p.P(`ormObj, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
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
	p.P(`return &pbResponse, nil`)
	p.P(`}`)
	p.P()
}

func (p *OrmPlugin) generatePreloading(ormable *OrmableType) {
	var assocList []string
	for _, fieldName := range p.getSortedFieldNames(ormable.Fields) {
		field := ormable.Fields[fieldName]
		if field.GetAssociation() != nil && field.GetHasMany().GetPositionField() == "" {
			assocList = append(assocList, fieldName)
		}
	}
	if len(assocList) != 0 {
		preload := ""
		for _, assoc := range assocList {
			preload += fmt.Sprintf(`.Preload("%s")`, assoc)
		}
		p.P(`db = db`, preload)
	}
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
