package op

import (
	"strings"
)

const (
	opCommonDelimiter      = ","
	opCommonInnerDelimiter = "."
)

//FieldSelectionMap is a convenience type that represents map[string]*Field
//used in FieldSelection and Field structs
type FieldSelectionMap map[string]*Field

func innerDelimiter(delimiter ...string) string {
	split := opCommonInnerDelimiter
	if delimiter != nil && len(delimiter) > 0 {
		split = delimiter[0]
	}
	return split
}

func toParts(input string, delimiter ...string) []string {
	split := innerDelimiter(delimiter...)
	return strings.Split(input, split)
}

//ParseFieldSelection transforms a string with comma-separated fields that comes
//from client to FieldSelection struct. For complex fields dot is used as a delimeter by
//default, but it is also possible to specify a different delimiter.
func ParseFieldSelection(input string, delimiter ...string) *FieldSelection {
	if len(input) == 0 {
		return &FieldSelection{Fields: nil}
	}

	fields := strings.Split(input, opCommonDelimiter)
	result := &FieldSelection{Fields: make(map[string]*Field, len(fields))}

	for _, field := range fields {
		result.Add(field, delimiter...)
	}

	return result
}

//GoString converts FieldSelection to a string representation
//It implements fmt.GoStringer interface and returns dot-notated fields separated by commas
func (f *FieldSelection) GoString() string {
	result := make([]string, 0, len(f.Fields))
	for _, field := range f.Fields {
		addChildFieldString(&result, "", field)
	}
	return strings.Join(result, opCommonDelimiter)
}

func addChildFieldString(result *[]string, parent string, field *Field) {
	if field.Subs == nil || len(field.Subs) == 0 {
		*result = append(*result, parent+field.Name)
	} else {
		parent = parent + field.Name + opCommonInnerDelimiter
		for _, f := range field.Subs {
			addChildFieldString(result, parent, f)
		}
	}
}

//Add allows to add new fields to FieldSelection
func (f *FieldSelection) Add(field string, delimiter ...string) {
	if len(field) == 0 {
		return
	}
	if f.Fields == nil {
		f.Fields = map[string]*Field{}
	}
	parts := toParts(field, delimiter...)
	name := parts[0]
	if _, ok := f.Fields[name]; !ok {
		f.Fields[name] = &Field{Name: name}
	}

	parent := f.Fields[name]
	for i := 1; i < len(parts); i++ {
		if parent.Subs == nil {
			parent.Subs = make(map[string]*Field)
		}
		name = parts[i]
		if _, ok := parent.Subs[name]; !ok {
			parent.Subs[name] = &Field{Name: name}
		}
		parent = parent.Subs[name]
	}
}

//Delete allows to remove fields from FieldSelection
func (f *FieldSelection) Delete(field string, delimiter ...string) bool {
	if len(field) == 0 || f.Fields == nil {
		return false
	}
	parts := toParts(field, delimiter...)
	name := parts[0]
	tmp := f.Fields

	for i := 0; i < len(parts); i++ {
		name = parts[i]
		if _, ok := tmp[name]; !ok {
			return false //such field do not exist if FieldSelection
		}
		if i < len(parts)-1 {
			tmp = tmp[name].Subs
			if tmp == nil {
				return false //such field do not exist if FieldSelection
			}
		}
	}
	delete(tmp, name)
	return true
}

//Get allows to get specified field from FieldSelection
func (f *FieldSelection) Get(field string, delimiter ...string) *Field {
	if len(field) == 0 || f.Fields == nil {
		return nil
	}
	parts := toParts(field, delimiter...)
	name := parts[0]
	tmp := f.Fields

	for i := 0; i < len(parts); i++ {
		name = parts[i]
		if _, ok := tmp[name]; !ok {
			return nil //such field do not exist if FieldSelection
		}
		if i < len(parts)-1 {
			tmp = tmp[name].Subs
			if tmp == nil {
				return nil //such field do not exist if FieldSelection
			}
		}
	}
	return tmp[name]
}
