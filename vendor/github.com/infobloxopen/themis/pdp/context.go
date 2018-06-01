// Package pdp implements Policy Decision Point (PDP). It is responsible for
// making authorization decisions based on policies it has.
package pdp

import (
	"fmt"
	"net"
	"strings"

	"github.com/infobloxopen/go-trees/domain"
	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
)

// Effect* constants define possible consequences of decision evaluation.
const (
	// EffectDeny indicates that request is denied.
	EffectDeny = iota
	// EffectPermit indicates that request is permitted.
	EffectPermit

	// EffectNotApplicable indicates that policies don't contain any policy
	// and rule applicable to the request.
	EffectNotApplicable

	// EffectIndeterminate indicates that evaluation can't be done for
	// the request. For example required attribute is missing.
	EffectIndeterminate
	// EffectIndeterminateD indicates that evaluation can't be done for
	// the request but if it could effect would be EffectDeny.
	EffectIndeterminateD
	// EffectIndeterminateD indicates that evaluation can't be done for
	// the request but if it could effect would be EffectPermit.
	EffectIndeterminateP
	// EffectIndeterminateD indicates that evaluation can't be done for
	// the request but if it could effect would be only EffectDeny or
	// EffectPermit.
	EffectIndeterminateDP

	EffectOutOfRange
)

var (
	effectNames = []string{
		"Deny",
		"Permit",
		"NotApplicable",
		"Indeterminate",
		"Indeterminate{D}",
		"Indeterminate{P}",
		"Indeterminate{DP}"}

	// EffectIDs maps all possible values of rule's effect to its id.
	EffectIDs = map[string]int{
		"deny":   EffectDeny,
		"permit": EffectPermit}
)

// Context represents request context. The context contains all information
// needed to evaluate request.
type Context struct {
	a map[string]interface{}
	c *LocalContentStorage
}

// EffectNameFromEnum returns human readable name for Effect enum
func EffectNameFromEnum(effectEnum int) string {
	if effectEnum >= EffectOutOfRange {
		return "EffectOutOfRange"
	}
	return effectNames[effectEnum]
}

// NewContext creates new instance of context. It requires pointer to local
// content storage and request attributes. The storage can be nil only
// if there is no policies or rules require it (otherwise evaluation may
// crash reaching it). Context collects input attributes by calling "f"
// function. The function is called "count" times and on each call it gets
// incrementing number starting from 0. The function should return attribute
// name and value. If "f" function returns error NewContext stops iterations
// and returns the same error. All pairs of attribute name and type should be
// unique.
func NewContext(c *LocalContentStorage, count int, f func(i int) (string, AttributeValue, error)) (*Context, error) {
	ctx := &Context{a: make(map[string]interface{}, count), c: c}

	for i := 0; i < count; i++ {
		ID, av, err := f(i)
		if err != nil {
			return nil, err
		}

		t := av.GetResultType()
		if v, ok := ctx.a[ID]; ok {
			switch v := v.(type) {
			default:
				panic(fmt.Errorf("expected AttributeValue or map[int]AttributeValue but got: %T, %#v", v, v))

			case AttributeValue:
				if v.t == t {
					return nil, newDuplicateAttributeValueError(ID, t, av, v)
				}

				m := make(map[Type]AttributeValue, 2)
				m[v.t] = v
				m[t] = av

			case map[Type]AttributeValue:
				if old, ok := v[t]; ok {
					return nil, newDuplicateAttributeValueError(ID, t, av, old)
				}

				v[t] = av
			}
		} else {
			ctx.a[ID] = av
		}
	}

	return ctx, nil
}

// String implements Stringer interface.
func (c *Context) String() string {
	lines := []string{}
	if c.c != nil {
		if s := c.c.String(); len(s) > 0 {
			lines = append(lines, s)
		}
	}

	if len(c.a) > 0 {
		if len(lines) > 0 {
			lines = append(lines, "")
		}

		lines = append(lines, "attributes:")
		for name, attrs := range c.a {
			switch v := attrs.(type) {
			default:
				panic(fmt.Errorf("expected AttributeValue or map[int]AttributeValue but got: %T, %#v", v, v))

			case AttributeValue:
				lines = append(lines, fmt.Sprintf("- %s.(%s): %s", name, v.t, v.describe()))

			case map[Type]AttributeValue:
				for t, av := range v {
					lines = append(lines, fmt.Sprintf("- %s.(%s): %s", name, t, av.describe()))
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

func (c *Context) getAttribute(a Attribute) (AttributeValue, error) {
	v, ok := c.a[a.id]
	if !ok {
		return UndefinedValue, bindError(newMissingAttributeError(), a.describe())
	}

	switch v := v.(type) {
	case AttributeValue:
		if v.t != a.t {
			return UndefinedValue, bindError(newMissingAttributeError(), a.describe())
		}

		return v, nil

	case map[Type]AttributeValue:
		av, ok := v[a.t]
		if !ok {
			return UndefinedValue, bindError(newMissingAttributeError(), a.describe())
		}

		return av, nil
	}

	panic(fmt.Errorf("expected AttributeValue or map[int]AttributeValue but got: %T, %#v", v, v))
}

// GetContentItem returns content item value
func (c *Context) GetContentItem(cID, iID string) (*ContentItem, error) {
	return c.c.Get(cID, iID)
}

func (c *Context) calculateBooleanExpression(e Expression) (bool, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return false, err
	}

	return v.boolean()
}

func (c *Context) calculateStringExpression(e Expression) (string, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return "", err
	}

	return v.str()
}

func (c *Context) calculateIntegerExpression(e Expression) (int64, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return 0, err
	}

	return v.integer()
}

func (c *Context) calculateFloatExpression(e Expression) (float64, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return 0, err
	}

	return v.float()
}

func (c *Context) calculateFloatOrIntegerExpression(e Expression) (float64, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return 0, err
	}

	if v.GetResultType() == TypeInteger {
		intVal, err := v.integer()
		if err != nil {
			return 0, err
		}

		return float64(intVal), nil
	}
	return v.float()
}

func (c *Context) calculateAddressExpression(e Expression) (net.IP, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return nil, err
	}

	return v.address()
}

func (c *Context) calculateDomainExpression(e Expression) (domain.Name, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return domain.Name{}, err
	}

	return v.domain()
}

func (c *Context) calculateNetworkExpression(e Expression) (*net.IPNet, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return nil, err
	}

	return v.network()
}

func (c *Context) calculateSetOfStringsExpression(e Expression) (*strtree.Tree, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return nil, err
	}

	return v.setOfStrings()
}

func (c *Context) calculateSetOfNetworksExpression(e Expression) (*iptree.Tree, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return nil, err
	}

	return v.setOfNetworks()
}

func (c *Context) calculateSetOfDomainsExpression(e Expression) (*domaintree.Node, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return nil, err
	}

	return v.setOfDomains()
}

func (c *Context) calculateFlags8Expression(e Expression) (uint8, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return 0, err
	}

	return v.flags8()
}

func (c *Context) calculateFlags16Expression(e Expression) (uint16, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return 0, err
	}

	return v.flags16()
}

func (c *Context) calculateFlags32Expression(e Expression) (uint32, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return 0, err
	}

	return v.flags32()
}

func (c *Context) calculateFlags64Expression(e Expression) (uint64, error) {
	v, err := e.Calculate(c)
	if err != nil {
		return 0, err
	}

	return v.flags64()
}

// Response represent result of policies evaluation.
type Response struct {
	// Effect is resulting effect.
	Effect      int
	status      boundError
	obligations []AttributeAssignmentExpression
}

// Status returns response's effect, obligation (as list of assignment
// expression) and error if any occurs during evaluation.
func (r Response) Status() (int, []AttributeAssignmentExpression, error) {
	return r.Effect, r.obligations, r.status
}

// Evaluable interface defines abstract PDP's entity which can be evaluated
// for given context (policy set or policy).
type Evaluable interface {
	GetID() (string, bool)
	Calculate(ctx *Context) Response
	Append(path []string, v interface{}) (Evaluable, error)
	Delete(path []string) (Evaluable, error)

	getOrder() int
	setOrder(ord int)
	describe() string
}
