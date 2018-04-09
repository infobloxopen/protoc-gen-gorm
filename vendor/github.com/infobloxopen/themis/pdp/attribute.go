package pdp

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
)

// Type* constants represent all data types PDP can work with.
const (
	// TypeUndefined stands for type of undefined value. The value usually
	// means that evaluation can't be done.
	TypeUndefined = iota
	// TypeBoolean is boolean data type.
	TypeBoolean
	// TypeString is string data type.
	TypeString
	// TypeInteger is integer data type.
	TypeInteger
	// TypeFloat is float data type.
	TypeFloat
	// TypeAddress is IPv4 or IPv6 address data type.
	TypeAddress
	// TypeNetwork is IPv4 or IPv6 network data type.
	TypeNetwork
	// TypeDomain is domain name data type.
	TypeDomain
	// TypeSetOfStrings is set of strings data type (internally stores order
	// in which it was created).
	TypeSetOfStrings
	// TypeSetOfNetworks is set of networks data type (unordered).
	TypeSetOfNetworks
	// TypeSetOfDomains is set of domains data type (unordered).
	TypeSetOfDomains
	// TypeListOfStrings is list of strings data type.
	TypeListOfStrings

	typesTotal
)

// Type* collections bind type names and IDs.
var (
	// TypeNames is list of humanreadable type names. The order must be kept
	// in sync with Type* constants order.
	TypeNames = []string{
		"Undefined",
		"Boolean",
		"String",
		"Integer",
		"Float",
		"Address",
		"Network",
		"Domain",
		"Set of Strings",
		"Set of Networks",
		"Set of Domains",
		"List of Strings"}

	// TypeKeys maps Type* constants to type IDs. Type ID is all lower case
	// type name. The slice is filled by init function.
	TypeKeys []string
	// TypeIDs maps type IDs to Type* constants. The map is filled by init
	// function.
	TypeIDs = map[string]int{}
)

var undefinedValue = AttributeValue{}

func init() {
	TypeKeys = make([]string, typesTotal)
	for t := 0; t < typesTotal; t++ {
		key := strings.ToLower(TypeNames[t])
		TypeKeys[t] = key
		TypeIDs[key] = t
	}
}

// Attribute represents attribute definition which binds attribute name
// and type.
type Attribute struct {
	id string
	t  int
}

// MakeAttribute creates new attribute instance. It requires attribute name
// as "ID" argument and type as "t" argument. Value of "t" should be one of
// Type* constants.
func MakeAttribute(ID string, t int) Attribute {
	return Attribute{id: ID, t: t}
}

// GetType returns attribute type.
func (a Attribute) GetType() int {
	return a.t
}

func (a Attribute) describe() string {
	return fmt.Sprintf("attr(%s.%s)", a.id, TypeNames[a.t])
}

// AttributeValue represents attribute value which binds data type and data.
// Value with undefined type indicates that evaluation can't get particular
// value.
type AttributeValue struct {
	t int
	v interface{}
}

// MakeBooleanValue creates instance of boolean attribute value.
func MakeBooleanValue(v bool) AttributeValue {
	return AttributeValue{
		t: TypeBoolean,
		v: v}
}

// MakeStringValue creates instance of string attribute value.
func MakeStringValue(v string) AttributeValue {
	return AttributeValue{
		t: TypeString,
		v: v}
}

// MakeIntegerValue creates instance of integer attribute value.
func MakeIntegerValue(v int64) AttributeValue {
	return AttributeValue{
		t: TypeInteger,
		v: v}
}

// MakeFloatValue creates instance of float attribute value.
func MakeFloatValue(v float64) AttributeValue {
	return AttributeValue{
		t: TypeFloat,
		v: v}
}

// MakeAddressValue creates instance of IP address attribute value.
func MakeAddressValue(v net.IP) AttributeValue {
	return AttributeValue{
		t: TypeAddress,
		v: v}
}

// MakeNetworkValue creates instance of IP network address attribute value.
// Argument should not be nil. Caller is responsible for the validation.
func MakeNetworkValue(v *net.IPNet) AttributeValue {
	return AttributeValue{
		t: TypeNetwork,
		v: v}
}

// MakeDomainValue creates instance of domain name attribute value. Argument
// should be valid domain name. Caller is responsible for the validation.
func MakeDomainValue(v domaintree.WireDomainNameLower) AttributeValue {
	return AttributeValue{
		t: TypeDomain,
		v: v}
}

// MakeSetOfStringsValue creates instance of set of strings attribute value.
func MakeSetOfStringsValue(v *strtree.Tree) AttributeValue {
	return AttributeValue{
		t: TypeSetOfStrings,
		v: v}
}

// MakeSetOfNetworksValue creates instance of set of networks attribute value.
func MakeSetOfNetworksValue(v *iptree.Tree) AttributeValue {
	return AttributeValue{
		t: TypeSetOfNetworks,
		v: v}
}

// MakeSetOfDomainsValue creates instance of set of domains attribute value.
func MakeSetOfDomainsValue(v *domaintree.Node) AttributeValue {
	return AttributeValue{
		t: TypeSetOfDomains,
		v: v}
}

// MakeListOfStringsValue creates instance of list of strings attribute value.
func MakeListOfStringsValue(v []string) AttributeValue {
	return AttributeValue{
		t: TypeListOfStrings,
		v: v}
}

// MakeValueFromString creates instance of attribute value by given type and
// string representation. The function performs necessary validation.
// No covertion defined for undefined type and collection types.
func MakeValueFromString(t int, s string) (AttributeValue, error) {
	switch t {
	case TypeUndefined:
		return undefinedValue, newInvalidTypeStringCastError(t)

	case TypeSetOfStrings, TypeSetOfNetworks, TypeSetOfDomains, TypeListOfStrings:
		return undefinedValue, newNotImplementedStringCastError(t)

	case TypeBoolean:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return undefinedValue, newInvalidBooleanStringCastError(s, err)
		}

		return MakeBooleanValue(b), nil

	case TypeString:
		return MakeStringValue(s), nil

	case TypeInteger:
		n, err := strconv.ParseInt(s, 0, 64)
		if err != nil {
			return undefinedValue, newInvalidIntegerStringCastError(s, err)
		}

		return MakeIntegerValue(n), nil

	case TypeFloat:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return undefinedValue, newInvalidFloatStringCastError(s, err)
		}

		return MakeFloatValue(f), nil

	case TypeAddress:
		a := net.ParseIP(s)
		if a == nil {
			return undefinedValue, newInvalidAddressStringCastError(s)
		}

		return MakeAddressValue(a), nil

	case TypeNetwork:
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			return undefinedValue, newInvalidNetworkStringCastError(s, err)
		}

		return MakeNetworkValue(n), nil

	case TypeDomain:
		d, err := domaintree.MakeWireDomainNameLower(s)
		if err != nil {
			return undefinedValue, newInvalidDomainNameStringCastError(s, err)
		}

		return MakeDomainValue(d), nil
	}

	return undefinedValue, newUnknownTypeStringCastError(t)
}

// GetResultType returns type of attribute value (implements Expression
// interface).
func (v AttributeValue) GetResultType() int {
	return v.t
}

func (v AttributeValue) describe() string {
	switch v.t {
	case TypeUndefined:
		return "val(undefined)"

	case TypeBoolean:
		return fmt.Sprintf("%v", v.v.(bool))

	case TypeString:
		return fmt.Sprintf("%q", v.v.(string))

	case TypeInteger:
		return strconv.FormatInt(v.v.(int64), 10)

	case TypeFloat:
		return strconv.FormatFloat(v.v.(float64), 'G', -1, 64)

	case TypeAddress:
		return v.v.(net.IP).String()

	case TypeNetwork:
		return v.v.(*net.IPNet).String()

	case TypeDomain:
		return fmt.Sprintf("domain(%s)", v.v.(domaintree.WireDomainNameLower).String())

	case TypeSetOfStrings:
		var s []string
		for p := range v.v.(*strtree.Tree).Enumerate() {
			s = append(s, fmt.Sprintf("%q", p.Key))
			if len(s) > 2 {
				s[2] = "..."
				break
			}
		}

		return fmt.Sprintf("set(%s)", strings.Join(s, ", "))

	case TypeSetOfNetworks:
		var s []string
		for p := range v.v.(*iptree.Tree).Enumerate() {
			s = append(s, p.Key.String())
			if len(s) > 2 {
				s[2] = "..."
				break
			}
		}

		return fmt.Sprintf("set(%s)", strings.Join(s, ", "))

	case TypeSetOfDomains:
		var s []string
		for p := range v.v.(*domaintree.Node).Enumerate() {
			s = append(s, fmt.Sprintf("%q", p.Key))
			if len(s) > 2 {
				s[2] = "..."
				break
			}
		}

		return fmt.Sprintf("domains(%s)", strings.Join(s, ", "))

	case TypeListOfStrings:
		var s []string
		for _, item := range v.v.([]string) {
			s = append(s, fmt.Sprintf("%q", item))
			if len(s) > 2 {
				s[2] = "..."
				break
			}
		}

		return fmt.Sprintf("[%s]", strings.Join(s, ", "))
	}

	return "val(unknown type)"
}

func (v AttributeValue) typeCheck(t int) error {
	if v.t != t {
		return bindError(newAttributeValueTypeError(t, v.t), v.describe())
	}

	return nil
}

func (v AttributeValue) boolean() (bool, error) {
	err := v.typeCheck(TypeBoolean)
	if err != nil {
		return false, err
	}

	return v.v.(bool), nil
}

func (v AttributeValue) str() (string, error) {
	err := v.typeCheck(TypeString)
	if err != nil {
		return "", err
	}

	return v.v.(string), nil
}

func (v AttributeValue) integer() (int64, error) {
	err := v.typeCheck(TypeInteger)
	if err != nil {
		return 0, err
	}

	return v.v.(int64), nil
}

func (v AttributeValue) float() (float64, error) {
	err := v.typeCheck(TypeFloat)
	if err != nil {
		return 0, err
	}

	return v.v.(float64), nil
}

func (v AttributeValue) address() (net.IP, error) {
	err := v.typeCheck(TypeAddress)
	if err != nil {
		return nil, err
	}

	return v.v.(net.IP), nil
}

func (v AttributeValue) network() (*net.IPNet, error) {
	err := v.typeCheck(TypeNetwork)
	if err != nil {
		return nil, err
	}

	return v.v.(*net.IPNet), nil
}

func (v AttributeValue) domain() (domaintree.WireDomainNameLower, error) {
	err := v.typeCheck(TypeDomain)
	if err != nil {
		return nil, err
	}

	return v.v.(domaintree.WireDomainNameLower), nil
}

func (v AttributeValue) setOfStrings() (*strtree.Tree, error) {
	err := v.typeCheck(TypeSetOfStrings)
	if err != nil {
		return nil, err
	}

	return v.v.(*strtree.Tree), nil
}

func (v AttributeValue) setOfNetworks() (*iptree.Tree, error) {
	err := v.typeCheck(TypeSetOfNetworks)
	if err != nil {
		return nil, err
	}

	return v.v.(*iptree.Tree), nil
}

func (v AttributeValue) setOfDomains() (*domaintree.Node, error) {
	err := v.typeCheck(TypeSetOfDomains)
	if err != nil {
		return nil, err
	}

	return v.v.(*domaintree.Node), nil
}

func (v AttributeValue) listOfStrings() ([]string, error) {
	err := v.typeCheck(TypeListOfStrings)
	if err != nil {
		return nil, err
	}

	return v.v.([]string), nil
}

// Calculate implements Expression interface and returns calculated value
func (v AttributeValue) Calculate(ctx *Context) (AttributeValue, error) {
	return v, nil
}

// Serialize converts attribute value to its string representation.
// No conversion defined for undefined value.
func (v AttributeValue) Serialize() (string, error) {
	switch v.t {
	case TypeUndefined:
		return "", newInvalidTypeSerializationError(v.t)

	case TypeBoolean:
		return strconv.FormatBool(v.v.(bool)), nil

	case TypeString:
		return v.v.(string), nil

	case TypeInteger:
		return strconv.FormatInt(v.v.(int64), 10), nil

	case TypeFloat:
		return strconv.FormatFloat(v.v.(float64), 'G', -1, 64), nil

	case TypeAddress:
		return v.v.(net.IP).String(), nil

	case TypeNetwork:
		return v.v.(*net.IPNet).String(), nil

	case TypeDomain:
		return v.v.(domaintree.WireDomainNameLower).String(), nil

	case TypeSetOfStrings:
		s := sortSetOfStrings(v.v.(*strtree.Tree))
		for i, item := range s {
			s[i] = strconv.Quote(item)
		}

		return strings.Join(s, ","), nil

	case TypeSetOfNetworks:
		var s []string
		for p := range v.v.(*iptree.Tree).Enumerate() {
			s = append(s, strconv.Quote(p.Key.String()))
		}

		return strings.Join(s, ","), nil

	case TypeSetOfDomains:
		var s []string
		for p := range v.v.(*domaintree.Node).Enumerate() {
			s = append(s, strconv.Quote(p.Key))
		}

		return strings.Join(s, ","), nil

	case TypeListOfStrings:
		var s []string
		for _, item := range v.v.([]string) {
			s = append(s, strconv.Quote(item))
		}

		return strings.Join(s, ","), nil
	}

	return "", newUnknownTypeSerializationError(v.t)
}

// AttributeAssignmentExpression represents assignment of arbitrary expression
// result to an attribute.
type AttributeAssignmentExpression struct {
	a Attribute
	e Expression
}

// MakeAttributeAssignmentExpression creates attribute assignment expression.
func MakeAttributeAssignmentExpression(a Attribute, e Expression) AttributeAssignmentExpression {
	return AttributeAssignmentExpression{
		a: a,
		e: e}
}

// Serialize evaluates assignment expression and returns string representation
// of resulting attribute name, type and value or error if the evaluaction
// can't be done.
func (a AttributeAssignmentExpression) Serialize(ctx *Context) (string, string, string, error) {
	ID := a.a.id
	typeName := TypeKeys[a.a.t]

	v, err := a.e.Calculate(ctx)
	if err != nil {
		return ID, typeName, "", bindErrorf(err, "assignment to %q", ID)
	}

	t := v.GetResultType()
	if a.a.t != t {
		return ID, typeName, "", bindErrorf(newAssignmentTypeMismatch(a.a, t), "assignment to %q", ID)
	}

	s, err := v.Serialize()
	if err != nil {
		return ID, typeName, "", bindErrorf(err, "assignment to %q", ID)
	}

	return ID, typeName, s, nil
}

// AttributeDesignator represents an expression which result is corresponding
// attribute value from request context.
type AttributeDesignator struct {
	a Attribute
}

// MakeAttributeDesignator creates designator expression instance for given
// attribute.
func MakeAttributeDesignator(a Attribute) AttributeDesignator {
	return AttributeDesignator{a}
}

// GetID returns ID of wrapped attribute.
func (d AttributeDesignator) GetID() string {
	return d.a.id
}

// GetResultType returns type of wrapped attribute (implements Expression
// interface).
func (d AttributeDesignator) GetResultType() int {
	return d.a.t
}

// Calculate implements Expression interface and returns calculated value
func (d AttributeDesignator) Calculate(ctx *Context) (AttributeValue, error) {
	return ctx.getAttribute(d.a)
}
