package op

import (
	"fmt"
	"strings"
)

// ParseFiltering is a shortcut to parse a filtering expression using default FilteringParser implementation
func ParseFiltering(text string) (*Filtering, error) {
	return (&filteringParser{}).Parse(text)
}

// FilteringParser is implemented by parsers of a filtering expression that conforms to REST API Syntax Specification.
type FilteringParser interface {
	Parse(string) (*Filtering, error)
}

// NewFilteringParser returns a default FilteringParser implementation.
func NewFilteringParser() FilteringParser {
	return &filteringParser{}
}

// UnexpectedTokenError describes a token that was not appropriate according to REST API Syntax Specification.
type UnexpectedTokenError struct {
	T Token
}

func (e *UnexpectedTokenError) Error() string {
	return fmt.Sprintf("Unexpected token %s", e.T)
}

// parser implements recursive descent parser of a filtering expression that conforms to REST API Syntax Specification.
// Some insights into recursive descent: https://en.wikipedia.org/wiki/Recursive_descent_parser .
type filteringParser struct {
	lexer    FilteringLexer
	curToken Token
}

// Parse builds an AST from an expression in text according to the following grammar:
// expr      : term (OR term)*
// term      : factor (AND factor)*
// factor    : ?NOT (LPAREN expr RPAREN | condition)
// condition : FIELD ((== | !=) (STRING | NUMBER | NULL) | (~ | !~) STRING | (> | >= | < | <=) NUMBER).
func (p *filteringParser) Parse(text string) (*Filtering, error) {
	p.lexer = NewFilteringLexer(text)
	token, err := p.lexer.NextToken()
	if err != nil {
		return nil, err
	}
	if _, ok := token.(EOFToken); ok {
		return nil, nil
	}
	p.curToken = token
	var expr FilteringExpression
	expr, err = p.expr()
	if err != nil {
		return nil, err
	}
	switch p.curToken.(type) {
	case EOFToken:
		f := &Filtering{}
		err := f.SetRoot(expr)
		if err != nil {
			return nil, err
		}
		return f, nil
	default:
		return nil, &UnexpectedTokenError{p.curToken}
	}
}

func (p *filteringParser) negateNode(node FilteringExpression) {
	switch v := node.(type) {
	case *LogicalOperator:
		v.IsNegative = !v.IsNegative
	case *StringCondition:
		v.IsNegative = !v.IsNegative
	case *NumberCondition:
		v.IsNegative = !v.IsNegative
	case *NullCondition:
		v.IsNegative = !v.IsNegative
	}
}

func (p *filteringParser) eatToken() error {
	token, err := p.lexer.NextToken()
	if err != nil {
		return err
	}
	p.curToken = token
	return nil
}

func (p *filteringParser) expr() (FilteringExpression, error) {
	node, err := p.term()
	if err != nil {
		return nil, err
	}
	_, isOr := p.curToken.(OrToken)
	for isOr {
		if err := p.eatToken(); err != nil {
			return nil, err
		}
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		newNode := &LogicalOperator{Type: LogicalOperator_OR}
		err = newNode.SetLeft(node)
		if err != nil {
			return nil, err
		}
		err = newNode.SetRight(right)
		if err != nil {
			return nil, err
		}
		node = newNode
		_, isOr = p.curToken.(OrToken)
	}
	return node, nil
}

func (p *filteringParser) term() (FilteringExpression, error) {
	node, err := p.factor()
	if err != nil {
		return nil, err
	}
	_, isAnd := p.curToken.(AndToken)
	for isAnd {
		if err := p.eatToken(); err != nil {
			return nil, err
		}
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		newNode := &LogicalOperator{Type: LogicalOperator_AND}
		err = newNode.SetLeft(node)
		if err != nil {
			return nil, err
		}
		err = newNode.SetRight(right)
		if err != nil {
			return nil, err
		}
		node = newNode
		_, isAnd = p.curToken.(AndToken)
	}
	return node, nil
}

func (p *filteringParser) factor() (FilteringExpression, error) {
	isNot := false
	switch p.curToken.(type) {
	case NotToken:
		isNot = true
		if err := p.eatToken(); err != nil {
			return nil, err
		}
	}
	switch p.curToken.(type) {
	case LparenToken:
		if err := p.eatToken(); err != nil {
			return nil, err
		}
		node, err := p.expr()
		if err != nil {
			return nil, err
		}
		switch p.curToken.(type) {
		case RparenToken:
			if err := p.eatToken(); err != nil {
				return nil, err
			}
			if isNot {
				p.negateNode(node)
			}
			return node, nil
		default:
			return nil, &UnexpectedTokenError{p.curToken}
		}
	default:
		node, err := p.condition()
		if err != nil {
			return nil, err
		}
		if isNot {
			p.negateNode(node)
		}
		return node, nil
	}
}

func (p *filteringParser) condition() (FilteringExpression, error) {
	field, ok := p.curToken.(FieldToken)
	if !ok {
		return nil, &UnexpectedTokenError{p.curToken}
	}
	if err := p.eatToken(); err != nil {
		return nil, err
	}
	switch p.curToken.(type) {
	case EqToken:
		if err := p.eatToken(); err != nil {
			return nil, err
		}
		switch token := p.curToken.(type) {
		case StringToken:
			if err := p.eatToken(); err != nil {
				return nil, err
			}
			return &StringCondition{
				FieldPath:  strings.Split(field.Value, "."),
				Value:      token.Value,
				Type:       StringCondition_EQ,
				IsNegative: false,
			}, nil
		case NumberToken:
			if err := p.eatToken(); err != nil {
				return nil, err
			}
			return &NumberCondition{
				FieldPath:  strings.Split(field.Value, "."),
				Value:      token.Value,
				Type:       NumberCondition_EQ,
				IsNegative: false,
			}, nil
		case NullToken:
			if err := p.eatToken(); err != nil {
				return nil, err
			}
			return &NullCondition{
				FieldPath:  strings.Split(field.Value, "."),
				IsNegative: false,
			}, nil
		default:
			return nil, &UnexpectedTokenError{p.curToken}
		}
	case NeToken:
		if err := p.eatToken(); err != nil {
			return nil, err
		}
		switch token := p.curToken.(type) {
		case StringToken:
			if err := p.eatToken(); err != nil {
				return nil, err
			}
			return &StringCondition{
				FieldPath:  strings.Split(field.Value, "."),
				Value:      token.Value,
				Type:       StringCondition_EQ,
				IsNegative: true,
			}, nil
		case NumberToken:
			if err := p.eatToken(); err != nil {
				return nil, err
			}
			return &NumberCondition{
				FieldPath:  strings.Split(field.Value, "."),
				Value:      token.Value,
				Type:       NumberCondition_EQ,
				IsNegative: true,
			}, nil
		case NullToken:
			if err := p.eatToken(); err != nil {
				return nil, err
			}
			return &NullCondition{
				FieldPath:  strings.Split(field.Value, "."),
				IsNegative: true,
			}, nil
		default:
			return nil, &UnexpectedTokenError{p.curToken}
		}
	case MatchToken:
		if err := p.eatToken(); err != nil {
			return nil, err
		}
		switch token := p.curToken.(type) {
		case StringToken:
			if err := p.eatToken(); err != nil {
				return nil, err
			}
			return &StringCondition{
				FieldPath:  strings.Split(field.Value, "."),
				Value:      token.Value,
				Type:       StringCondition_MATCH,
				IsNegative: false,
			}, nil
		default:
			return nil, &UnexpectedTokenError{p.curToken}
		}
	case NmatchToken:
		if err := p.eatToken(); err != nil {
			return nil, err
		}
		switch token := p.curToken.(type) {
		case StringToken:
			if err := p.eatToken(); err != nil {
				return nil, err
			}
			return &StringCondition{
				FieldPath:  strings.Split(field.Value, "."),
				Value:      token.Value,
				Type:       StringCondition_MATCH,
				IsNegative: true,
			}, nil
		default:
			return nil, &UnexpectedTokenError{p.curToken}
		}
	case GtToken:
		if err := p.eatToken(); err != nil {
			return nil, err
		}
		switch token := p.curToken.(type) {
		case NumberToken:
			if err := p.eatToken(); err != nil {
				return nil, err
			}
			return &NumberCondition{
				FieldPath:  strings.Split(field.Value, "."),
				Value:      token.Value,
				Type:       NumberCondition_GT,
				IsNegative: false,
			}, nil
		default:
			return nil, &UnexpectedTokenError{p.curToken}
		}
	case GeToken:
		if err := p.eatToken(); err != nil {
			return nil, err
		}
		switch token := p.curToken.(type) {
		case NumberToken:
			if err := p.eatToken(); err != nil {
				return nil, err
			}
			return &NumberCondition{
				FieldPath:  strings.Split(field.Value, "."),
				Value:      token.Value,
				Type:       NumberCondition_GE,
				IsNegative: false,
			}, nil
		default:
			return nil, &UnexpectedTokenError{p.curToken}
		}
	case LtToken:
		if err := p.eatToken(); err != nil {
			return nil, err
		}
		switch token := p.curToken.(type) {
		case NumberToken:
			if err := p.eatToken(); err != nil {
				return nil, err
			}
			return &NumberCondition{
				FieldPath:  strings.Split(field.Value, "."),
				Value:      token.Value,
				Type:       NumberCondition_LT,
				IsNegative: false,
			}, nil
		default:
			return nil, &UnexpectedTokenError{p.curToken}
		}
	case LeToken:
		if err := p.eatToken(); err != nil {
			return nil, err
		}
		switch token := p.curToken.(type) {
		case NumberToken:
			if err := p.eatToken(); err != nil {
				return nil, err
			}
			return &NumberCondition{
				FieldPath:  strings.Split(field.Value, "."),
				Value:      token.Value,
				Type:       NumberCondition_LE,
				IsNegative: false,
			}, nil
		default:
			return nil, &UnexpectedTokenError{p.curToken}
		}
	default:
		return nil, &UnexpectedTokenError{p.curToken}
	}
}
