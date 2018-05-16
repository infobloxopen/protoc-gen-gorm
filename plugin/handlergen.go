package plugin

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	jgorm "github.com/jinzhu/gorm"
)

func (p *OrmPlugin) generateDefaultHandlers(file *generator.FileDescriptor) {
	for _, message := range file.Messages() {
		if message.Options != nil {
			if opts := getMessageOptions(message); opts == nil || !*opts.Ormable {
				continue
			} else if opts.GetMultiAccount() {
				p.usingAuth = true
			}
		} else {
			continue
		}
		p.gormPkgName = "gorm"
		p.lftPkgName = "ops"

		p.generateCreateHandler(message)
		p.generateReadHandler(message)
		p.generateUpdateHandler(message)
		p.generateDeleteHandler(message)
		p.generateListHandler(message)
		p.generateStrictUpdateHandler(message)
	}
}

func (p *OrmPlugin) generateCreateHandler(message *generator.Descriptor) {
	typeName := p.TypeName(message)
	p.P(`// DefaultCreate`, typeName, ` executes a basic gorm create call`)
	p.P(`func DefaultCreate`, typeName, `(ctx context.Context, in *`,
		typeName, `, db *`, p.gormPkgName, `.DB) (*`, typeName, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, errors.New("Nil argument to DefaultCreate`, typeName, `")`)
	p.P(`}`)
	p.P(`ormObj, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	if opts := getMessageOptions(message); opts != nil && opts.GetMultiAccount() {
		p.P("accountID, err := auth.GetAccountID(ctx, nil)")
		p.P("if err != nil {")
		p.P("return nil, err")
		p.P("}")
		p.P("ormObj.AccountID = accountID")
	}
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
	p.P(`// DefaultRead`, typeName, ` executes a basic gorm read call`)
	p.P(`func DefaultRead`, typeName, `(ctx context.Context, in *`,
		typeName, `, db *`, p.gormPkgName, `.DB) (*`, typeName, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, errors.New("Nil argument to DefaultRead`, typeName, `")`)
	p.P(`}`)
	p.P(`ormParams, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	if opts := getMessageOptions(message); opts != nil && opts.GetMultiAccount() {
		p.P("accountID, err := auth.GetAccountID(ctx, nil)")
		p.P("if err != nil {")
		p.P("return nil, err")
		p.P("}")
		p.P("ormParams.AccountID = accountID")
	}
	p.P(`ormResponse := `, typeName, `ORM{}`)
	p.sortOrderedHasMany(message)
	p.P(`if err = db.Set("gorm:auto_preload", true).Where(&ormParams).First(&ormResponse).Error; err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`pbResponse, err := ormResponse.ToPB(ctx)`)
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
		typeName, `, db *`, p.gormPkgName, `.DB) (*`, typeName, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, errors.New("Nil argument to DefaultUpdate`, typeName, `")`)
	p.P(`}`)
	if isMultiAccount {
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
		typeName, `, db *`, p.gormPkgName, `.DB) error {`)
	p.P(`if in == nil {`)
	p.P(`return errors.New("Nil argument to DefaultDelete`, typeName, `")`)
	p.P(`}`)
	p.P(`ormObj, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return err`)
	p.P(`}`)
	if opts := getMessageOptions(message); opts != nil && opts.GetMultiAccount() {
		p.P("accountID, err := auth.GetAccountID(ctx, nil)")
		p.P("if err != nil {")
		p.P("return err")
		p.P("}")
		p.P("ormObj.AccountID = accountID")
	}
	p.P(`err = db.Where(&ormObj).Delete(&`, typeName, `ORM{}).Error`)
	p.P(`return err`)
	p.P(`}`)
	p.P()
}

func (p *OrmPlugin) generateListHandler(message *generator.Descriptor) {
	typeName := p.TypeName(message)

	p.P(`// DefaultList`, typeName, ` executes a gorm list call`)
	p.P(`func DefaultList`, typeName, `(ctx context.Context, db *`, p.gormPkgName,
		`.DB) ([]*`, typeName, `, error) {`)
	p.P(`ormResponse := []`, typeName, `ORM{}`)
	p.P(`db, err := `, p.lftPkgName, `.ApplyCollectionOperators(db, ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	if opts := getMessageOptions(message); opts != nil && opts.GetMultiAccount() {
		p.P("accountID, err := auth.GetAccountID(ctx, nil)")
		p.P("if err != nil {")
		p.P("return nil, err")
		p.P("}")
		p.P(`db = db.Where(&`, typeName, `ORM{AccountID: accountID})`)
	}
	p.sortOrderedHasMany(message)
	p.P(`if err := db.Set("gorm:auto_preload", true).Find(&ormResponse).Error; err != nil {`)
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
	multiAccount := getMessageOptions(message).GetMultiAccount()
	p.P(`// DefaultStrictUpdate`, typeName, ` clears first level 1:many children and then executes a gorm update call`)
	p.P(`func DefaultStrictUpdate`, typeName, `(ctx context.Context, in *`,
		typeName, `, db *`, p.gormPkgName, `.DB) (*`, typeName, `, error) {`)
	p.P(`if in == nil {`)
	p.P(`return nil, fmt.Errorf("Nil argument to DefaultCascadedUpdate`, typeName, `")`)
	p.P(`}`)
	p.P(`ormObj, err := in.ToORM(ctx)`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	if multiAccount {
		p.P("accountID, err := auth.GetAccountID(ctx, nil)")
		p.P("if err != nil {")
		p.P("return nil, err")
		p.P("}")
	}
	p.removeChildAssociations(message)
	if multiAccount {
		p.P(`db = db.Where(&`, typeName, `ORM{AccountID: accountID})`)
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

func (p *OrmPlugin) setupOrderedHasMany(message *generator.Descriptor) {
	ormable := p.getOrmable(p.TypeName(message))
	for fieldName, field := range ormable.Fields {
		if field.GetHasMany().GetPositionField() != "" {
			positionField := field.GetHasMany().GetPositionField()
			p.P(`for i, e := range `, `ormObj.`, fieldName, `{`)
			p.P(`e.`, positionField, ` = i`)
			p.P(`}`)
		}
	}
}

func (p *OrmPlugin) sortOrderedHasMany(message *generator.Descriptor) {
	ormable := p.getOrmable(p.TypeName(message))
	for fieldName, field := range ormable.Fields {
		if field.GetHasMany().GetPositionField() != "" {
			positionField := field.GetHasMany().GetPositionField()
			p.P(`db = db.Preload("`, fieldName, `", func(db *gorm.DB) *gorm.DB {`)
			p.P(`return db.Order("`, jgorm.ToDBName(positionField), `")`)
			p.P(`})`)
		}
	}
}

func (p *OrmPlugin) removeChildAssociations(message *generator.Descriptor) {
	ormable := p.getOrmable(p.TypeName(message))
	for fieldName, field := range ormable.Fields {
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
			p.P(`filter`, fieldName, ` := `, strings.Trim(field.Type, "[]*"), `{}`)
			zeroValue := p.guessZeroValue(ormable.Fields[assocKeyName].Type)
			p.P(`if ormObj.`, assocKeyName, ` == `, zeroValue, `{`)
			p.P(`return nil, errors.New("Can't do overwriting update with no `, assocKeyName, ` value for `, ormable.Name, `")`)
			p.P(`}`)
			p.P(`filter`, fieldName, `.`, foreignKeyName, ` = `, `ormObj.`, assocKeyName)
			if getMessageOptions(message).GetMultiAccount() {
				p.P(`filter`, fieldName, `.`, `AccountID`, ` = accountID`)
			}
			p.P(`if err = db.Where(filter`, fieldName, `).Delete(`, strings.Trim(field.Type, "[]*"), `{}).Error; err != nil {`)
			p.P(`return nil, err`)
			p.P(`}`)
		}
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
		return `uuid.Nil`
	}
	return ``
}
