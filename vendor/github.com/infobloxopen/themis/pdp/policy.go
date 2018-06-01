package pdp

import (
	"fmt"
	"io"
)

// RuleCombiningAlg represent abstract rule combining algorithm. The algorithm
// defines how to evaluate policy rules and how to get paticular result.
type RuleCombiningAlg interface {
	execute(rules []*Rule, ctx *Context) Response
}

// RuleCombiningAlgMaker creates instance of rule combining algorithm.
// The function accepts set of policy rules and parameters of algorithm.
type RuleCombiningAlgMaker func(rules []*Rule, params interface{}) RuleCombiningAlg

var (
	firstApplicableEffectRCAInstance = firstApplicableEffectRCA{}
	denyOverridesRCAInstance         = denyOverridesRCA{}

	// RuleCombiningAlgs defines map of algorithm id to particular maker of
	// the algorithm. Contains only algorithms which don't require any
	// parameters.
	RuleCombiningAlgs = map[string]RuleCombiningAlgMaker{
		"firstapplicableeffect": makeFirstApplicableEffectRCA,
		"denyoverrides":         makeDenyOverridesRCA}

	// RuleCombiningParamAlgs defines map of algorithm id to particular maker
	// of the algorithm. Contains only algorithms which require parameters.
	RuleCombiningParamAlgs = map[string]RuleCombiningAlgMaker{
		"mapper": makeMapperRCA}

	ruleArrPrefix = []byte(",\"rules\":[")
)

// Policy represent PDP policy (minimal evaluable entity).
type Policy struct {
	ord         int
	id          string
	hidden      bool
	target      Target
	rules       []*Rule
	obligations []AttributeAssignmentExpression
	algorithm   RuleCombiningAlg
}

// NewPolicy creates new instance of policy with given id (or hidden), target,
// set of rules, algorithm and obligations. To make instance of algorithm it
// uses one of makers from RuleCombiningAlgs or RuleCombiningParamAlgs and its
// parameters if it requires any.
func NewPolicy(ID string, hidden bool, target Target, rules []*Rule, makeRCA RuleCombiningAlgMaker, params interface{}, obligations []AttributeAssignmentExpression) *Policy {
	for i, r := range rules {
		r.ord = i
	}

	return &Policy{
		id:          ID,
		hidden:      hidden,
		target:      target,
		rules:       rules,
		obligations: obligations,
		algorithm:   makeRCA(rules, params)}
}

func (p *Policy) describe() string {
	if pid, ok := p.GetID(); ok {
		return fmt.Sprintf("policy %q", pid)
	}

	return "hidden policy"
}

// GetID implements Evaluable interface and returns policy id if policy
// isn't hidden.
func (p *Policy) GetID() (string, bool) {
	return p.id, !p.hidden
}

// Calculate implements Evaluable interface and evaluates policy for given
// request contest.
func (p *Policy) Calculate(ctx *Context) Response {
	match, err := p.target.calculate(ctx)
	if err != nil {
		r := combineEffectAndStatus(err, p.algorithm.execute(p.rules, ctx))
		if r.status != nil {
			r.status = bindError(r.status, p.describe())
		}
		return r
	}

	if !match {
		return Response{EffectNotApplicable, nil, nil}
	}

	r := p.algorithm.execute(p.rules, ctx)
	if r.Effect == EffectDeny || r.Effect == EffectPermit {
		r.obligations = append(r.obligations, p.obligations...)
	}

	if r.status != nil {
		r.status = bindError(r.status, p.describe())
	}

	return r
}

// Append implements Evaluable interface and puts new rule to the policy.
// Argument path should be empty and v should contain a pointer to rule.
// Append can't put hidden rule to policy or any rule to hidden policy.
func (p *Policy) Append(path []string, v interface{}) (Evaluable, error) {
	if p.hidden {
		return p, newHiddenPolicyModificationError()
	}

	if len(path) > 0 {
		return p, bindError(newTooLongPathPolicyModificationError(path), p.id)
	}

	child, ok := v.(*Rule)
	if !ok {
		return p, bindError(newInvalidPolicyItemTypeError(v), p.id)
	}

	_, ok = child.GetID()
	if !ok {
		return p, bindError(newHiddenRuleAppendError(), p.id)
	}

	return p.putChild(child), nil
}

// Delete implements Evaluable interface and removes rule from the policy.
// Argument path should contain exactly one string which is id of rule
// to remove. Delete can't remove a rule from hidden policy.
func (p *Policy) Delete(path []string) (Evaluable, error) {
	if p.hidden {
		return p, newHiddenPolicyModificationError()
	}

	if len(path) <= 0 {
		return p, bindError(newTooShortPathPolicyModificationError(), p.id)
	}

	ID := path[0]

	if len(path) > 1 {
		return p, bindError(newTooLongPathPolicyModificationError(path[1:]), p.id)
	}

	r, err := p.delChild(ID)
	if err != nil {
		return p, bindError(err, p.id)
	}

	return r, nil
}

func (p *Policy) getOrder() int {
	return p.ord
}

func (p *Policy) setOrder(ord int) {
	p.ord = ord
}

func (p *Policy) updatedCopy(rules []*Rule, algorithm RuleCombiningAlg) *Policy {
	return &Policy{
		ord:         p.ord,
		id:          p.id,
		target:      p.target,
		rules:       rules,
		obligations: p.obligations,
		algorithm:   algorithm,
	}
}

func (p *Policy) getChild(ID string) (int, *Rule, error) {
	for i, r := range p.rules {
		if rID, ok := r.GetID(); ok && rID == ID {
			return i, r, nil
		}
	}

	return -1, nil, newMissingPolicyChildError(ID)
}

func (p *Policy) putChild(child *Rule) *Policy {
	ID, _ := child.GetID()

	var rules []*Rule

	i, old, err := p.getChild(ID)
	if err == nil {
		child.ord = old.ord

		rules = make([]*Rule, len(p.rules))
		if i > 0 {
			copy(rules, p.rules[:i])
		}

		rules[i] = child

		if i+1 < len(p.rules) {
			copy(rules[i+1:], p.rules[i+1:])
		}
	} else if len(p.rules) > 0 {
		child.ord = p.rules[len(p.rules)-1].ord + 1
		rules = make([]*Rule, len(p.rules)+1)
		copy(rules, p.rules)
		rules[len(p.rules)] = child
	} else {
		child.ord = 0
		rules = []*Rule{child}
	}

	algorithm := p.algorithm
	switch m := algorithm.(type) {
	case mapperRCA:
		algorithm = m.add(ID, child, old)

	case flagsMapperRCA:
		algorithm = m.add(ID, child, old)
	}

	return p.updatedCopy(rules, algorithm)
}

func (p *Policy) delChild(ID string) (*Policy, error) {
	i, old, err := p.getChild(ID)
	if err != nil {
		return nil, err
	}

	var rules []*Rule
	if len(p.rules) > 1 {
		rules = make([]*Rule, len(p.rules)-1)
		if i > 0 {
			copy(rules, p.rules[:i])
		}

		if i+1 < len(p.rules) {
			copy(rules[i:], p.rules[i+1:])
		}
	}

	algorithm := p.algorithm
	switch m := algorithm.(type) {
	case mapperRCA:
		algorithm = m.del(ID, old)

	case flagsMapperRCA:
		algorithm = m.del(ID, old)
	}

	return p.updatedCopy(rules, algorithm), nil
}

// MarshalWithDepth implements StorageMarshal
func (p Policy) MarshalWithDepth(out io.Writer, depth int) error {
	if depth < 0 {
		return newMarshalInvalidDepthError(depth)
	}
	err := marshalHeader(storageNodeFmt{
		Ord: p.ord,
		ID:  p.id,
	}, out)
	if err != nil {
		return bindErrorf(err, "pid=\"%s\"", p.id)
	}
	_, err = out.Write(ruleArrPrefix)
	if err != nil {
		return bindErrorf(err, "pid=\"%s\"", p.id)
	}
	if depth > 0 {
		var firstRule int
		for i, r := range p.rules {
			if _, ok := r.GetID(); !ok {
				continue
			}
			if err = r.MarshalWithDepth(out, depth-1); err != nil {
				return bindErrorf(err, "pid=\"%s\",i=%d", p.id, i)
			}
			firstRule = i
			break
		}
		for i, r := range p.rules[firstRule+1:] {
			if _, ok := r.GetID(); !ok {
				continue
			}
			if _, err := out.Write([]byte{','}); err != nil {
				return bindErrorf(err, "pid=\"%s\",i=%d", p.id, i)
			}
			if err = r.MarshalWithDepth(out, depth-1); err != nil {
				return bindErrorf(err, "pid=\"%s\",i=%d", p.id, i)
			}
		}
	}
	_, err = out.Write([]byte("]}"))
	if err != nil {
		return bindErrorf(err, "pid=\"%s\"", p.id)
	}
	return nil
}

type firstApplicableEffectRCA struct {
}

func makeFirstApplicableEffectRCA(rules []*Rule, params interface{}) RuleCombiningAlg {
	return firstApplicableEffectRCAInstance
}

func (a firstApplicableEffectRCA) execute(rules []*Rule, ctx *Context) Response {
	for _, rule := range rules {
		r := rule.calculate(ctx)
		if r.Effect != EffectNotApplicable {
			return r
		}
	}

	return Response{EffectNotApplicable, nil, nil}
}

type denyOverridesRCA struct {
}

func makeDenyOverridesRCA(rules []*Rule, params interface{}) RuleCombiningAlg {
	return denyOverridesRCAInstance
}

func (a denyOverridesRCA) describe() string {
	return "deny overrides"
}

func (a denyOverridesRCA) execute(rules []*Rule, ctx *Context) Response {
	errs := []error{}
	obligations := make([]AttributeAssignmentExpression, 0)

	indetD := 0
	indetP := 0
	indetDP := 0

	permits := 0

	for _, rule := range rules {
		r := rule.calculate(ctx)
		if r.Effect == EffectDeny {
			obligations = append(obligations, r.obligations...)
			return r
		}

		if r.Effect == EffectPermit {
			permits++
			obligations = append(obligations, r.obligations...)
			continue
		}

		if r.Effect == EffectNotApplicable {
			continue
		}

		if r.Effect == EffectIndeterminateD {
			indetD++
		} else {
			if r.Effect == EffectIndeterminateP {
				indetP++
			} else {
				indetDP++
			}

		}

		errs = append(errs, r.status)
	}

	var err boundError
	if len(errs) > 1 {
		err = bindError(newMultiError(errs), a.describe())
	} else if len(errs) > 0 {
		err = bindError(errs[0], a.describe())
	}

	if indetDP > 0 || (indetD > 0 && (indetP > 0 || permits > 0)) {
		return Response{EffectIndeterminateDP, err, nil}
	}

	if indetD > 0 {
		return Response{EffectIndeterminateD, err, nil}
	}

	if permits > 0 {
		return Response{EffectPermit, nil, obligations}
	}

	if indetP > 0 {
		return Response{EffectIndeterminateP, err, nil}
	}

	return Response{EffectNotApplicable, nil, nil}
}
