package gorm

import (
	"fmt"
	"strings"

	"github.com/infobloxopen/atlas-app-toolkit/query"
)

// FilterStringToGorm is a shortcut to parse a filter string using default FilteringParser implementation
// and call FilteringToGorm on the returned filtering expression.
func FilterStringToGorm(filter string) (string, []interface{}, error) {
	f, err := query.ParseFiltering(filter)
	if err != nil {
		return "", nil, err
	}
	return FilteringToGorm(f)
}

// FilteringToGorm returns GORM Plain SQL representation of the filtering expression.
func FilteringToGorm(m *query.Filtering) (string, []interface{}, error) {
	if m == nil {
		return "", nil, nil
	}

	switch r := m.Root.(type) {
	case *query.Filtering_Operator:
		return LogicalOperatorToGorm(r.Operator)
	case *query.Filtering_StringCondition:
		return StringConditionToGorm(r.StringCondition)
	case *query.Filtering_NumberCondition:
		return NumberConditionToGorm(r.NumberCondition)
	case *query.Filtering_NullCondition:
		return NullConditionToGorm(r.NullCondition)
	default:
		return "", nil, fmt.Errorf("%T type is not supported in Filtering", r)
	}
}

// LogicalOperatorToGorm returns GORM Plain SQL representation of the logical operator.
func LogicalOperatorToGorm(lop *query.LogicalOperator) (string, []interface{}, error) {
	var lres string
	var largs []interface{}
	var err error
	switch l := lop.Left.(type) {
	case *query.LogicalOperator_LeftOperator:
		lres, largs, err = LogicalOperatorToGorm(l.LeftOperator)
	case *query.LogicalOperator_LeftStringCondition:
		lres, largs, err = StringConditionToGorm(l.LeftStringCondition)
	case *query.LogicalOperator_LeftNumberCondition:
		lres, largs, err = NumberConditionToGorm(l.LeftNumberCondition)
	case *query.LogicalOperator_LeftNullCondition:
		lres, largs, err = NullConditionToGorm(l.LeftNullCondition)
	default:
		return "", nil, fmt.Errorf("%T type is not supported in Filtering", l)
	}
	if err != nil {
		return "", nil, err
	}

	var rres string
	var rargs []interface{}
	switch r := lop.Right.(type) {
	case *query.LogicalOperator_RightOperator:
		rres, rargs, err = LogicalOperatorToGorm(r.RightOperator)
	case *query.LogicalOperator_RightStringCondition:
		rres, rargs, err = StringConditionToGorm(r.RightStringCondition)
	case *query.LogicalOperator_RightNumberCondition:
		rres, rargs, err = NumberConditionToGorm(r.RightNumberCondition)
	case *query.LogicalOperator_RightNullCondition:
		rres, rargs, err = NullConditionToGorm(r.RightNullCondition)
	default:
		return "", nil, fmt.Errorf("%T type is not supported in Filtering", r)
	}
	if err != nil {
		return "", nil, err
	}

	var o string
	switch lop.Type {
	case query.LogicalOperator_AND:
		o = "AND"
	case query.LogicalOperator_OR:
		o = "OR"
	}
	var neg string
	if lop.IsNegative {
		neg = "NOT"
	}
	return fmt.Sprintf("%s(%s %s %s)", neg, lres, o, rres), append(largs, rargs...), nil
}

// StringConditionToGorm returns GORM Plain SQL representation of the string condition.
func StringConditionToGorm(c *query.StringCondition) (string, []interface{}, error) {
	var o string
	switch c.Type {
	case query.StringCondition_EQ:
		o = "="
	case query.StringCondition_MATCH:
		o = "~"
	}
	var neg string
	if c.IsNegative {
		neg = "NOT"
	}
	return fmt.Sprintf("%s(%s %s ?)", neg, strings.Join(c.FieldPath, "."), o), []interface{}{c.Value}, nil
}

// NumberConditionToGorm returns GORM Plain SQL representation of the number condition.
func NumberConditionToGorm(c *query.NumberCondition) (string, []interface{}, error) {
	var o string
	switch c.Type {
	case query.NumberCondition_EQ:
		o = "="
	case query.NumberCondition_GT:
		o = ">"
	case query.NumberCondition_GE:
		o = ">="
	case query.NumberCondition_LT:
		o = "<"
	case query.NumberCondition_LE:
		o = "<="
	}
	var neg string
	if c.IsNegative {
		neg = "NOT"
	}
	return fmt.Sprintf("%s(%s %s ?)", neg, strings.Join(c.FieldPath, "."), o), []interface{}{c.Value}, nil
}

// NullConditionToGorm returns GORM Plain SQL representation of the null condition.
func NullConditionToGorm(c *query.NullCondition) (string, []interface{}, error) {
	o := "IS NULL"
	var neg string
	if c.IsNegative {
		neg = "NOT"
	}
	return fmt.Sprintf("%s(%s %s)", neg, strings.Join(c.FieldPath, "."), o), nil, nil
}
