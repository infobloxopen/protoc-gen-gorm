package pdp

import "sort"

type flagsMapperPCA struct {
	argument  Expression
	policies  []Evaluable
	def       Evaluable
	err       Evaluable
	order     int
	algorithm PolicyCombiningAlg
}

func collectFlaggedSubPolicies(n uint64, policies []Evaluable) []Evaluable {
	out := []Evaluable{}
	for i, p := range policies {
		if p != nil && n&(1<<uint(i)) != 0 {
			out = append(out, p)
		}
	}

	return out
}

func (a flagsMapperPCA) execute(policies []Evaluable, ctx *Context) Response {
	v, err := a.argument.Calculate(ctx)
	if err != nil {
		switch err.(type) {
		case *MissingValueError:
			if a.def != nil {
				return a.def.Calculate(ctx)
			}
		}

		return a.calculateErrorPolicy(ctx, err)
	}

	t := v.GetResultType()
	if t, ok := t.(*FlagsType); ok {
		n, err := v.flags()
		if err != nil {
			return a.calculateErrorPolicy(ctx, err)
		}

		sub := collectFlaggedSubPolicies(n, a.getPoliciesMap(policies, t))
		if len(sub) > 1 && a.order == MapperPCAInternalOrder {
			sort.Sort(byPolicyOrder(sub))
		}

		r := a.algorithm.execute(sub, ctx)
		if r.Effect == EffectNotApplicable && a.def != nil {
			return a.def.Calculate(ctx)
		}

		return r
	}

	return a.calculateErrorPolicy(ctx, newFlagsMapperRCAArgumentTypeError(t))
}

func (a flagsMapperPCA) add(ID string, child, old Evaluable) PolicyCombiningAlg {
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
		return flagsMapperPCA{
			argument:  a.argument,
			policies:  a.policies,
			def:       def,
			err:       err,
			order:     a.order,
			algorithm: a.algorithm,
		}
	}

	m := make([]Evaluable, len(a.policies))
	copy(m, a.policies)
	m[i] = child

	return flagsMapperPCA{
		argument:  a.argument,
		policies:  m,
		def:       def,
		err:       err,
		order:     a.order,
		algorithm: a.algorithm,
	}
}

func (a flagsMapperPCA) del(ID string, old Evaluable) PolicyCombiningAlg {
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
		return flagsMapperPCA{
			argument:  a.argument,
			policies:  a.policies,
			def:       def,
			err:       err,
			order:     a.order,
			algorithm: a.algorithm,
		}
	}

	m := make([]Evaluable, len(a.policies))
	copy(m, a.policies)
	m[i] = nil

	return flagsMapperPCA{
		argument:  a.argument,
		policies:  m,
		def:       def,
		err:       err,
		order:     a.order,
		algorithm: a.algorithm,
	}
}

func (a flagsMapperPCA) describe() string {
	return "mapper"
}

func (a flagsMapperPCA) calculateErrorPolicy(ctx *Context, err error) Response {
	if a.err != nil {
		return a.err.Calculate(ctx)
	}

	return Response{EffectIndeterminate, bindError(err, a.describe()), nil}
}

func (a flagsMapperPCA) getPoliciesMap(policies []Evaluable, t *FlagsType) []Evaluable {
	if a.policies != nil {
		return a.policies
	}

	out := make([]Evaluable, len(t.b))
	for _, p := range policies {
		if id, ok := p.GetID(); ok {
			i := t.GetFlagBit(id)
			if i >= 0 {
				out[i] = p
			}
		}
	}

	return out
}
