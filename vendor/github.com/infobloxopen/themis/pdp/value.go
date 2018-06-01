package pdp

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/infobloxopen/go-trees/domain"
	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
)

// AttributeValue represents attribute value which binds data type and data.
// Value with undefined type indicates that evaluation can't get particular
// value.
type AttributeValue struct {
	t Type
	v interface{}
}

// UndefinedValue is used to represent a failure to get particular value.
var UndefinedValue = AttributeValue{
	t: TypeUndefined,
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
func MakeDomainValue(v domain.Name) AttributeValue {
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

// MakeFlagsValue8 creates instance of given flags value which fits 8 bits integer.
func MakeFlagsValue8(v uint8, t Type) AttributeValue {
	if t, ok := t.(*FlagsType); ok {
		if t.c != 8 {
			panic(fmt.Errorf("expected %d bits value for %q but got 8", t.c, t))
		}

		return AttributeValue{
			t: t,
			v: v}
	}

	panic(fmt.Errorf("can't make flags value for type %q", t))
}

// MakeFlagsValue16 creates instance of given flags value which fits 16 bits integer.
func MakeFlagsValue16(v uint16, t Type) AttributeValue {
	if t, ok := t.(*FlagsType); ok {
		if t.c != 16 {
			panic(fmt.Errorf("expected %d bits value for %q but got 16", t.c, t))
		}

		return AttributeValue{
			t: t,
			v: v}
	}

	panic(fmt.Errorf("can't make flags value for type %q", t))
}

// MakeFlagsValue32 creates instance of given flags value which fits 32 bits integer.
func MakeFlagsValue32(v uint32, t Type) AttributeValue {
	if t, ok := t.(*FlagsType); ok {
		if t.c != 32 {
			panic(fmt.Errorf("expected %d bits value for %q but got 32", t.c, t))
		}

		return AttributeValue{
			t: t,
			v: v}
	}

	panic(fmt.Errorf("can't make flags value for type %q", t))
}

// MakeFlagsValue64 creates instance of given flags value which fits 64 bits integer.
func MakeFlagsValue64(v uint64, t Type) AttributeValue {
	if t, ok := t.(*FlagsType); ok {
		if t.c != 64 {
			panic(fmt.Errorf("expected %d bits value for %q but got 64", t.c, t))
		}

		return AttributeValue{
			t: t,
			v: v}
	}

	panic(fmt.Errorf("can't make flags value for type %q", t))
}

// MakeValueFromString creates instance of attribute value by given type and
// string representation. The function performs necessary validation.
// No covertion defined for undefined type and collection types.
func MakeValueFromString(t Type, s string) (AttributeValue, error) {
	if _, ok := t.(*FlagsType); ok {
		return UndefinedValue, newNotImplementedStringCastError(t)
	}

	switch t {
	case TypeUndefined:
		return UndefinedValue, newInvalidTypeStringCastError(t)

	case TypeSetOfStrings, TypeSetOfNetworks, TypeSetOfDomains, TypeListOfStrings:
		return UndefinedValue, newNotImplementedStringCastError(t)

	case TypeBoolean:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return UndefinedValue, newInvalidBooleanStringCastError(s, err)
		}

		return MakeBooleanValue(b), nil

	case TypeString:
		return MakeStringValue(s), nil

	case TypeInteger:
		n, err := strconv.ParseInt(s, 0, 64)
		if err != nil {
			return UndefinedValue, newInvalidIntegerStringCastError(s, err)
		}

		return MakeIntegerValue(n), nil

	case TypeFloat:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return UndefinedValue, newInvalidFloatStringCastError(s, err)
		}

		return MakeFloatValue(f), nil

	case TypeAddress:
		a := net.ParseIP(s)
		if a == nil {
			return UndefinedValue, newInvalidAddressStringCastError(s)
		}

		return MakeAddressValue(a), nil

	case TypeNetwork:
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			return UndefinedValue, newInvalidNetworkStringCastError(s, err)
		}

		return MakeNetworkValue(n), nil

	case TypeDomain:
		d, err := domain.MakeNameFromString(s)
		if err != nil {
			return UndefinedValue, newInvalidDomainNameStringCastError(s, err)
		}

		return MakeDomainValue(d), nil
	}

	return UndefinedValue, newUnknownTypeStringCastError(t)
}

// GetResultType returns type of attribute value (implements Expression
// interface).
func (v AttributeValue) GetResultType() Type {
	return v.t
}

func (v AttributeValue) describe() string {
	if t, ok := v.t.(*FlagsType); ok {
		var n uint64
		switch t.c {
		case 8:
			n = uint64(v.v.(uint8))

		case 16:
			n = uint64(v.v.(uint16))

		case 32:
			n = uint64(v.v.(uint32))

		case 64:
			n = v.v.(uint64)
		}

		var s []string
		for i := 0; i < len(t.b); i++ {
			if n&(1<<uint(i)) != 0 {
				s = append(s, strconv.Quote(t.b[i]))
				if len(s) > 2 {
					s[2] = "..."
					break
				}
			}
		}

		return fmt.Sprintf("flags<%q>(%s)", t, strings.Join(s, ", "))
	}

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
		return fmt.Sprintf("domain(%s)", v.v.(domain.Name).String())

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

func (v AttributeValue) typeCheck(t Type) error {
	if v.t != t {
		return bindError(newAttributeValueTypeError(t, v.t), v.describe())
	}

	return nil
}

func (v AttributeValue) flagsTypeCheck() (*FlagsType, error) {
	if t, ok := v.t.(*FlagsType); ok {
		return t, nil
	}

	return nil, bindError(newAttributeValueFlagsTypeError(v.t), v.describe())
}

func (v AttributeValue) flagsTypeCheckN(n int) error {
	t, err := v.flagsTypeCheck()
	if err != nil {
		return err
	}

	if t.c != n {
		return bindError(newAttributeValueFlagsBitsError(v.t, n, len(t.f)), v.describe())
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

func (v AttributeValue) domain() (domain.Name, error) {
	err := v.typeCheck(TypeDomain)
	if err != nil {
		return domain.Name{}, err
	}

	return v.v.(domain.Name), nil
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

func (v AttributeValue) flags8() (uint8, error) {
	err := v.flagsTypeCheckN(8)
	if err != nil {
		return 0, err
	}

	return v.v.(uint8), nil
}

func (v AttributeValue) flags16() (uint16, error) {
	err := v.flagsTypeCheckN(16)
	if err != nil {
		return 0, err
	}

	return v.v.(uint16), nil
}

func (v AttributeValue) flags32() (uint32, error) {
	err := v.flagsTypeCheckN(32)
	if err != nil {
		return 0, err
	}

	return v.v.(uint32), nil
}

func (v AttributeValue) flags64() (uint64, error) {
	err := v.flagsTypeCheckN(64)
	if err != nil {
		return 0, err
	}

	return v.v.(uint64), nil
}

func (v AttributeValue) flags() (uint64, error) {
	t, err := v.flagsTypeCheck()
	if err != nil {
		return 0, err
	}

	switch t.Capacity() {
	case 8:
		return uint64(v.v.(uint8)), nil

	case 16:
		return uint64(v.v.(uint16)), nil

	case 32:
		return uint64(v.v.(uint32)), nil
	}

	return v.v.(uint64), nil
}

// Calculate implements Expression interface and returns calculated value
func (v AttributeValue) Calculate(ctx *Context) (AttributeValue, error) {
	return v, nil
}

// Serialize converts attribute value to its string representation.
// No conversion defined for undefined value.
func (v AttributeValue) Serialize() (string, error) {
	if t, ok := v.t.(*FlagsType); ok {
		var n uint64
		switch t.c {
		case 8:
			n = uint64(v.v.(uint8))

		case 16:
			n = uint64(v.v.(uint16))

		case 32:
			n = uint64(v.v.(uint32))

		case 64:
			n = v.v.(uint64)
		}

		var s []string
		for i := 0; i < len(t.b); i++ {
			if n&(1<<uint(i)) != 0 {
				s = append(s, strconv.Quote(t.b[i]))
			}
		}

		return strings.Join(s, ","), nil
	}

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
		return v.v.(domain.Name).String(), nil

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

// Rebind produces copy of the value with given type if the type matches original value type.
func (v AttributeValue) Rebind(t Type) (AttributeValue, error) {
	if v.t == t {
		return v, nil
	}

	if !v.t.Match(t) {
		return v, newNotMatchingTypeRebindError(t, v.t)
	}

	switch t.(type) {
	case *FlagsType:
		return AttributeValue{
			t: t,
			v: v.v,
		}, nil
	}

	return v, newUnknownMetaType(t)
}
