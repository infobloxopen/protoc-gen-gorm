package pdp

import (
	"sort"
)

type flagsMapperRCA struct {
	argument  Expression
	rules     []*Rule
	def       *Rule
	err       *Rule
	order     int
	algorithm RuleCombiningAlg
}

func collectFlaggedSubRules(n uint64, rules []*Rule) []*Rule {
	out := []*Rule{}
	for i, r := range rules {
		if r != nil && n&(1<<uint(i)) != 0 {
			out = append(out, r)
		}
	}

	return out
}

func (a flagsMapperRCA) execute(rules []*Rule, ctx *Context) Response {
	v, err := a.argument.Calculate(ctx)
	if err != nil {
		switch err.(type) {
		case *MissingValueError:
			if a.def != nil {
				return a.def.calculate(ctx)
			}
		}

		return a.calculateErrorRule(ctx, err)
	}

	t := v.GetResultType()
	if t, ok := t.(*FlagsType); ok {
		n, err := v.flags()
		if err != nil {
			return a.calculateErrorRule(ctx, err)
		}

		sub := collectFlaggedSubRules(n, a.getRulesMap(rules, t))
		if len(sub) > 1 && a.order == MapperRCAInternalOrder {
			sort.Sort(byRuleOrder(sub))
		}

		r := a.algorithm.execute(sub, ctx)
		if r.Effect == EffectNotApplicable && a.def != nil {
			return a.def.calculate(ctx)
		}

		return r
	}

	return a.calculateErrorRule(ctx, newFlagsMapperRCAArgumentTypeError(t))
}

func (a flagsMapperRCA) add(ID string, child, old *Rule) RuleCombiningAlg {
	def := a.def
	if old != nil && old == def {
		def = child
	}

	err := a.err
	if old != nil && old == err {
		err = child
	}

	i := a.argument.GetResultType().(*FlagsType).GetFlagBit(ID)
	if i < 0 {
		return flagsMapperRCA{
			argument:  a.argument,
			rules:     a.rules,
			def:       def,
			err:       err,
			order:     a.order,
			algorithm: a.algorithm,
		}
	}

	m := make([]*Rule, len(a.rules))
	copy(m, a.rules)
	m[i] = child

	return flagsMapperRCA{
		argument:  a.argument,
		rules:     m,
		def:       def,
		err:       err,
		order:     a.order,
		algorithm: a.algorithm,
	}
}

func (a flagsMapperRCA) del(ID string, old *Rule) RuleCombiningAlg {
	def := a.def
	if old != nil && old == def {
		def = nil
	}

	err := a.err
	if old != nil && old == err {
		err = nil
	}

	i := a.argument.GetResultType().(*FlagsType).GetFlagBit(ID)
	if i < 0 {
		return flagsMapperRCA{
			argument:  a.argument,
			rules:     a.rules,
			def:       def,
			err:       err,
			order:     a.order,
			algorithm: a.algorithm,
		}
	}

	m := make([]*Rule, len(a.rules))
	copy(m, a.rules)
	m[i] = nil

	return flagsMapperRCA{
		argument:  a.argument,
		rules:     m,
		def:       def,
		err:       err,
		order:     a.order,
		algorithm: a.algorithm,
	}
}

func (a flagsMapperRCA) describe() string {
	return "mapper"
}

func (a flagsMapperRCA) calculateErrorRule(ctx *Context, err error) Response {
	if a.err != nil {
		return a.err.calculate(ctx)
	}

	return Response{EffectIndeterminate, bindError(err, a.describe()), nil}
}

func (a flagsMapperRCA) getRulesMap(rules []*Rule, t *FlagsType) []*Rule {
	if a.rules != nil {
		return a.rules
	}

	out := make([]*Rule, len(t.b))
	for _, r := range rules {
		if !r.hidden {
			i := t.GetFlagBit(r.id)
			if i >= 0 {
				out[i] = r
			}
		}
	}

	return out
}
