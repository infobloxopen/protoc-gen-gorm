package query

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/golang/protobuf/proto"
)

// Filter is a shortcut to parse a filter string using default FilteringParser implementation
// and call Filter on the returned filtering expression.
func Filter(obj interface{}, filter string) (bool, error) {
	f, err := ParseFiltering(filter)
	if err != nil {
		return false, err
	}
	return f.Filter(obj)
}

// FilteringExpression is the interface implemented by types that represent nodes in a filtering expression AST.
type FilteringExpression interface {
	Filter(interface{}) (bool, error)
}

// Matcher is implemented by structs that require custom filtering logic.
type Matcher interface {
	Match(*Filtering) (bool, error)
}

// Filter evaluates underlying filtering expression against obj.
// If obj implements Matcher, call it's custom implementation.
func (m *Filtering) Filter(obj interface{}) (bool, error) {
	if m == nil {
		return true, nil
	}
	if matcher, ok := obj.(Matcher); ok {
		return matcher.Match(m)
	}
	r := m.Root
	if f, ok := r.(FilteringExpression); ok {
		return f.Filter(obj)
	} else {
		return false, fmt.Errorf("%T type does not implement FilteringExpression", r)
	}
}

// TypeMismatchError representes a type that is required for a value under FieldPath.
type TypeMismatchError struct {
	ReqType   string
	FieldPath []string
}

func (e *TypeMismatchError) Error() string {
	return fmt.Sprintf("%s is not a %s type", strings.Join(e.FieldPath, "."), e.ReqType)
}

// UnsupportedOperatorError represents an operator that is not supported by a particular field type.
type UnsupportedOperatorError struct {
	Type string
	Op   string
}

func (e *UnsupportedOperatorError) Error() string {
	return fmt.Sprintf("%s is not supported for %s type", e.Op, e.Type)
}

// Filter evaluates filtering expression against obj.
func (lop *LogicalOperator) Filter(obj interface{}) (bool, error) {
	var res bool
	var err error
	l := lop.Left
	if f, ok := l.(FilteringExpression); ok {
		res, err = f.Filter(obj)
		if err != nil {
			return false, err
		}
	} else {
		return false, fmt.Errorf("%T type does not implement FilteringExpression", l)
	}
	if lop.Type == LogicalOperator_AND && !res {
		return negateIfNeeded(lop.IsNegative, false), nil
	} else if lop.Type == LogicalOperator_OR && res {
		return negateIfNeeded(lop.IsNegative, true), nil
	}
	r := lop.Right
	if f, ok := r.(FilteringExpression); ok {
		res, err = f.Filter(obj)
		if err != nil {
			return false, err
		}
	} else {
		return false, fmt.Errorf("%T type does not implement FilteringExpression", r)
	}
	return negateIfNeeded(lop.IsNegative, res), nil
}

// Filter evaluates string condition against obj.
// If obj is a proto message, then 'protobuf' tag is used to map FieldPath to obj's struct fields,
// otherwise 'json' tag is used.
func (c *StringCondition) Filter(obj interface{}) (bool, error) {
	fv := fieldByFieldPath(obj, c.FieldPath)
	fv = dereferenceValue(fv)
	if fv.Kind() != reflect.String {
		return false, &TypeMismatchError{"string", c.FieldPath}
	}
	s := fv.String()
	switch c.Type {
	case StringCondition_EQ:
		return negateIfNeeded(s == c.Value, c.IsNegative), nil
	case StringCondition_MATCH:
		// add regex caching
		matched, err := regexp.MatchString(c.Value, s)
		if err != nil {
			return false, err
		}
		return negateIfNeeded(matched, c.IsNegative), nil
	default:
		return false, &UnsupportedOperatorError{"string", c.Type.String()}
	}
}

// Filter evaluates number condition against obj.
// If obj is a proto message, then 'protobuf' tag is used to map FieldPath to obj's struct fields,
// otherwise 'json' tag is used.
func (c *NumberCondition) Filter(obj interface{}) (bool, error) {
	fv := fieldByFieldPath(obj, c.FieldPath)
	fv = dereferenceValue(fv)
	var f float64
	switch fv.Kind() {
	case reflect.Float32, reflect.Float64:
		f = fv.Float()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		f = float64(fv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		f = float64(fv.Uint())
	default:
		return false, &TypeMismatchError{"number", c.FieldPath}
	}
	switch c.Type {
	case NumberCondition_EQ:
		return negateIfNeeded(f == c.Value, c.IsNegative), nil
	case NumberCondition_GT:
		return negateIfNeeded(f > c.Value, c.IsNegative), nil
	case NumberCondition_GE:
		return negateIfNeeded(f >= c.Value, c.IsNegative), nil
	case NumberCondition_LT:
		return negateIfNeeded(f < c.Value, c.IsNegative), nil
	case NumberCondition_LE:
		return negateIfNeeded(f <= c.Value, c.IsNegative), nil
	default:
		return false, &UnsupportedOperatorError{"number", c.Type.String()}
	}
}

// Filter evaluates null condition against obj.
// If obj is a proto message, then 'protobuf' tag is used to map FieldPath to obj's struct fields,
// otherwise 'json' tag is used.
func (c *NullCondition) Filter(obj interface{}) (bool, error) {
	fv := fieldByFieldPath(obj, c.FieldPath)
	if fv.Kind() != reflect.Ptr {
		return false, &TypeMismatchError{"nullable", c.FieldPath}
	}
	return negateIfNeeded(fv.IsNil(), c.IsNegative), nil
}

func fieldByFieldPath(obj interface{}, fieldPath []string) reflect.Value {
	switch obj.(type) {
	case proto.Message:
		return fieldByProtoPath(obj, fieldPath)
	default:
		return fieldByJSONPath(obj, fieldPath)
	}
}

func fieldByProtoPath(obj interface{}, protoPath []string) reflect.Value {
	v := dereferenceValue(reflect.ValueOf(obj))
	props := proto.GetProperties(v.Type())
	for _, p := range props.Prop {
		if p.OrigName == protoPath[0] {
			return v.FieldByName(p.Name)
		}
		if p.JSONName == protoPath[0] {
			return v.FieldByName(p.Name)
		}
	}
	return reflect.Value{}
}

func fieldByJSONPath(obj interface{}, jsonPath []string) reflect.Value {
	v := dereferenceValue(reflect.ValueOf(obj))
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if getJSONName(sf) == jsonPath[0] {
			return v.Field(i)
		}
	}
	return reflect.Value{}
}

func getJSONName(sf reflect.StructField) string {
	if jsonTag, ok := sf.Tag.Lookup("json"); ok {
		return strings.Split(jsonTag, ",")[0]
	}
	return sf.Name
}

func dereferenceValue(value reflect.Value) reflect.Value {
	kind := value.Kind()
	for kind == reflect.Ptr || kind == reflect.Interface {
		value = value.Elem()
		kind = value.Kind()
	}
	return value
}

func negateIfNeeded(neg bool, value bool) bool {
	if neg {
		return !value
	}
	return value
}

func (m *Filtering_Operator) Filter(obj interface{}) (bool, error) {
	return m.Operator.Filter(obj)
}

func (m *Filtering_StringCondition) Filter(obj interface{}) (bool, error) {
	return m.StringCondition.Filter(obj)
}

func (m *Filtering_NumberCondition) Filter(obj interface{}) (bool, error) {
	return m.NumberCondition.Filter(obj)
}

func (m *Filtering_NullCondition) Filter(obj interface{}) (bool, error) {
	return m.NullCondition.Filter(obj)
}

func (m *LogicalOperator_LeftOperator) Filter(obj interface{}) (bool, error) {
	return m.LeftOperator.Filter(obj)
}

func (m *LogicalOperator_LeftStringCondition) Filter(obj interface{}) (bool, error) {
	return m.LeftStringCondition.Filter(obj)
}

func (m *LogicalOperator_LeftNumberCondition) Filter(obj interface{}) (bool, error) {
	return m.LeftNumberCondition.Filter(obj)
}

func (m *LogicalOperator_LeftNullCondition) Filter(obj interface{}) (bool, error) {
	return m.LeftNullCondition.Filter(obj)
}

func (m *LogicalOperator_RightOperator) Filter(obj interface{}) (bool, error) {
	return m.RightOperator.Filter(obj)
}

func (m *LogicalOperator_RightStringCondition) Filter(obj interface{}) (bool, error) {
	return m.RightStringCondition.Filter(obj)
}

func (m *LogicalOperator_RightNumberCondition) Filter(obj interface{}) (bool, error) {
	return m.RightNumberCondition.Filter(obj)
}

func (m *LogicalOperator_RightNullCondition) Filter(obj interface{}) (bool, error) {
	return m.RightNullCondition.Filter(obj)
}

// SetRoot automatically wraps r into appropriate oneof structure and sets it to Root.
func (m *Filtering) SetRoot(r interface{}) error {
	switch x := r.(type) {
	case *LogicalOperator:
		m.Root = &Filtering_Operator{x}
	case *StringCondition:
		m.Root = &Filtering_StringCondition{x}
	case *NumberCondition:
		m.Root = &Filtering_NumberCondition{x}
	case *NullCondition:
		m.Root = &Filtering_NullCondition{x}
	case nil:
		m.Root = nil
	default:
		return fmt.Errorf("Filtering.Root cannot be assigned to type %T", x)
	}
	return nil
}

// SetLeft automatically wraps l into appropriate oneof structure and sets it to Root.
func (m *LogicalOperator) SetLeft(l interface{}) error {
	switch x := l.(type) {
	case *LogicalOperator:
		m.Left = &LogicalOperator_LeftOperator{x}
	case *StringCondition:
		m.Left = &LogicalOperator_LeftStringCondition{x}
	case *NumberCondition:
		m.Left = &LogicalOperator_LeftNumberCondition{x}
	case *NullCondition:
		m.Left = &LogicalOperator_LeftNullCondition{x}
	case nil:
		m.Left = nil
	default:
		return fmt.Errorf("Filtering.Left cannot be assigned to type %T", x)
	}
	return nil
}

// SetRight automatically wraps r into appropriate oneof structure and sets it to Root.
func (m *LogicalOperator) SetRight(r interface{}) error {
	switch x := r.(type) {
	case *LogicalOperator:
		m.Right = &LogicalOperator_RightOperator{x}
	case *StringCondition:
		m.Right = &LogicalOperator_RightStringCondition{x}
	case *NumberCondition:
		m.Right = &LogicalOperator_RightNumberCondition{x}
	case *NullCondition:
		m.Right = &LogicalOperator_RightNullCondition{x}
	case nil:
		m.Right = nil
	default:
		return fmt.Errorf("Filtering.Right cannot be assigned to type %T", x)
	}
	return nil
}
