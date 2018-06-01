package query

import (
	"fmt"
	"strconv"
	"unicode"
)

// FilteringLexer is impemented by lexical analyzers that are used by filtering expression parsers.
type FilteringLexer interface {
	NextToken() (Token, error)
}

// NewFilteringLexer returns a default FilteringLexer implementation.
// text is a filtering expression to analyze.
func NewFilteringLexer(text string) FilteringLexer {
	var runes []rune
	for _, r := range text {
		runes = append(runes, r)
	}
	if len(runes) > 0 {
		return &filteringLexer{runes, 0, runes[0], false}
	}
	return &filteringLexer{runes, 0, 0, true}
}

// UnexpectedSymbolError describes symbol S in position Pos that was not appropriate according to REST API Syntax Specification.
type UnexpectedSymbolError struct {
	S   rune
	Pos int
}

func (e *UnexpectedSymbolError) Error() string {
	return fmt.Sprintf("Unexpected symbol %c in %d position", e.S, e.Pos)
}

// Token is impelemented by all supported tokens in a filtering expression.
type Token interface {
	Token()
}

// TokenBase is used as a base type for all types which are tokens.
type TokenBase struct{}

// Token distinguishes tokens from other types.
func (t TokenBase) Token() {}

// LparenToken represents left parenthesis.
type LparenToken struct {
	TokenBase
}

func (t LparenToken) String() string {
	return "("
}

// RparenToken represents right parenthesis.
type RparenToken struct {
	TokenBase
}

func (t RparenToken) String() string {
	return ")"
}

// NumberToken represents a number literal.
// Value is a value of the literal.
type NumberToken struct {
	TokenBase
	Value float64
}

func (t NumberToken) String() string {
	return fmt.Sprint(t.Value)
}

// StringToken represents a string literal.
// Value is a value of the literal.
type StringToken struct {
	TokenBase
	Value string
}

func (t StringToken) String() string {
	return fmt.Sprint(t.Value)
}

// FieldToken represents a reference to a value of a resource.
// Value is a value of the reference.
type FieldToken struct {
	TokenBase
	Value string
}

func (t FieldToken) String() string {
	return fmt.Sprint(t.Value)
}

// AndToken represents logical and.
type AndToken struct {
	TokenBase
}

func (t AndToken) String() string {
	return "and"
}

// OrToken represents logical or.
type OrToken struct {
	TokenBase
}

func (t OrToken) String() string {
	return "or"
}

// NotToken represents logical not.
type NotToken struct {
	TokenBase
}

func (t NotToken) String() string {
	return "not"
}

// EqToken represents equals operator.
type EqToken struct {
	TokenBase
}

func (t EqToken) String() string {
	return "=="
}

// NeToken represents not equals operator.
type NeToken struct {
	TokenBase
}

func (t NeToken) String() string {
	return "!="
}

// MatchToken represents regular expression match.
type MatchToken struct {
	TokenBase
}

func (t MatchToken) String() string {
	return "~"
}

// NmatchToken represents negation of regular expression match.
type NmatchToken struct {
	TokenBase
}

func (t NmatchToken) String() string {
	return "!~"
}

// GtToken represents greater than operator.
type GtToken struct {
	TokenBase
}

func (t GtToken) String() string {
	return ">"
}

// GeToken represents greater than or equals operator.
type GeToken struct {
	TokenBase
}

func (t GeToken) String() string {
	return ">="
}

// LtToken represents less than operator.
type LtToken struct {
	TokenBase
}

func (t LtToken) String() string {
	return "<"
}

// LeToken represents less than or equals operator.
type LeToken struct {
	TokenBase
}

func (t LeToken) String() string {
	return "<="
}

// NullToken represents null literal.
type NullToken struct {
	TokenBase
}

func (t NullToken) String() string {
	return "null"
}

// EOFToken represents end of an expression.
type EOFToken struct {
	TokenBase
}

func (t EOFToken) String() string {
	return "EOF"
}

type filteringLexer struct {
	text    []rune
	pos     int
	curChar rune
	eof     bool
}

func (lexer *filteringLexer) advance() {
	lexer.pos++
	if lexer.pos < len(lexer.text) {
		lexer.curChar = lexer.text[lexer.pos]
	} else {
		lexer.eof = true
		lexer.curChar = 0
	}
}

func (lexer *filteringLexer) number() (Token, error) {
	number := string(lexer.curChar)
	metDot := false
	lexer.advance()
	for !lexer.eof {
		if unicode.IsDigit(lexer.curChar) {
			number += string(lexer.curChar)
		} else if !metDot && lexer.curChar == '.' {
			number += string(lexer.curChar)
			metDot = true
		} else {
			break
		}
		lexer.advance()
	}
	parsed, err := strconv.ParseFloat(number, 64)
	if err != nil {
		return nil, err
	}
	return NumberToken{Value: parsed}, nil
}

func (lexer *filteringLexer) string() (Token, error) {
	// Add quote escaping support
	term := lexer.curChar
	s := ""
	lexer.advance()
	for lexer.curChar != term {
		if lexer.eof {
			return nil, &UnexpectedSymbolError{lexer.curChar, lexer.pos}
		}
		s += string(lexer.curChar)
		lexer.advance()
	}
	lexer.advance()
	return StringToken{Value: s}, nil
}

func (lexer *filteringLexer) fieldOrReserved() (Token, error) {
	s := string(lexer.curChar)
	lexer.advance()
	for !lexer.eof {
		if unicode.IsDigit(lexer.curChar) ||
			unicode.IsLetter(lexer.curChar) ||
			lexer.curChar == '.' ||
			lexer.curChar == '-' ||
			lexer.curChar == '_' {
			s += string(lexer.curChar)
		} else {
			break
		}
		lexer.advance()
	}
	switch s {
	case "and":
		return AndToken{}, nil
	case "or":
		return OrToken{}, nil
	case "not":
		return NotToken{}, nil
	case "null":
		return NullToken{}, nil
	case "eq":
		return EqToken{}, nil
	case "ne":
		return NeToken{}, nil
	case "gt":
		return GtToken{}, nil
	case "ge":
		return GeToken{}, nil
	case "lt":
		return LtToken{}, nil
	case "le":
		return LeToken{}, nil
	case "match":
		return MatchToken{}, nil
	case "nomatch":
		return NmatchToken{}, nil
	default:
		return FieldToken{Value: s}, nil
	}
}

// NextToken returns the next token from the expression.
func (lexer *filteringLexer) NextToken() (Token, error) {
	for !lexer.eof {
		switch {
		case unicode.IsSpace(lexer.curChar):
			lexer.advance()
		case lexer.curChar == '(':
			lexer.advance()
			return LparenToken{}, nil
		case lexer.curChar == ')':
			lexer.advance()
			return RparenToken{}, nil
		case lexer.curChar == '~':
			lexer.advance()
			return MatchToken{}, nil
		case lexer.curChar == '=':
			lexer.advance()
			if lexer.curChar == '=' {
				lexer.advance()
				return EqToken{}, nil
			}
			return nil, &UnexpectedSymbolError{lexer.curChar, lexer.pos}
		case lexer.curChar == '!':
			lexer.advance()
			if lexer.curChar == '=' {
				lexer.advance()
				return NeToken{}, nil
			} else if lexer.curChar == '~' {
				lexer.advance()
				return NmatchToken{}, nil
			} else {
				return nil, &UnexpectedSymbolError{lexer.curChar, lexer.pos}
			}
		case lexer.curChar == '>':
			lexer.advance()
			if lexer.curChar == '=' {
				lexer.advance()
				return GeToken{}, nil
			}
			return GtToken{}, nil
		case lexer.curChar == '<':
			lexer.advance()
			if lexer.curChar == '=' {
				lexer.advance()
				return LeToken{}, nil
			}
			return LtToken{}, nil
		case lexer.curChar == '\'' || lexer.curChar == '"':
			return lexer.string()
		case unicode.IsDigit(lexer.curChar):
			return lexer.number()
		case unicode.IsLetter(lexer.curChar):
			return lexer.fieldOrReserved()
		default:
			return nil, &UnexpectedSymbolError{lexer.curChar, lexer.pos}
		}
	}
	return EOFToken{}, nil
}
