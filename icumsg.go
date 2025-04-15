// Package icumsg provides an ICU Message Format
// (See https://unicode-org.github.io/icu/userguide/format_parse/messages/)
package icumsg

import (
	"errors"
	"iter"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/romshark/icumsg/internal/cldr"
	"golang.org/x/text/language"
)

type TokenType uint8

const (
	_ TokenType = iota

	// Literal. IndexStart and IndexEnd are byte offsets in the input string.
	TokenTypeLiteral      // Any literal
	TokenTypeSimpleArg    // { arg }
	TokenTypePluralOffset // offset:1
	TokenTypeArgName      // The name of any argument

	// The following token types always follow TokenTypeArgName.
	TokenTypeArgTypeNumber   // "You have {count, number} new messages."
	TokenTypeArgTypeDate     // "Your appointment is on {appointmentDate, date}."
	TokenTypeArgTypeTime     // "The train departs at {departureTime, time}."
	TokenTypeArgTypeSpellout // "You have {count, spellout} new notifications."
	TokenTypeArgTypeOrdinal  // "You came in {place, ordinal} place!"
	TokenTypeArgTypeDuration // "Estimated time: {seconds, duration}."

	// The following token types always follow any argument type.
	TokenTypeArgStyleShort
	TokenTypeArgStyleMedium
	TokenTypeArgStyleLong
	TokenTypeArgStyleFull
	TokenTypeArgStyleInteger
	TokenTypeArgStyleCurrency
	TokenTypeArgStylePercent
	TokenTypeArgStyleCustom

	// TokenTypeOptionName is the select option. Always follows TokenTypeOption.
	TokenTypeOptionName

	// Complex. IndexEnd is an index of the token buffer.
	TokenTypePlural        // {arg, plural, ...}
	TokenTypeSelect        // {arg, select, ...}
	TokenTypeSelectOrdinal // {arg, selectordinal, ...}
	TokenTypeOption        // The { ... } that follows an option name.
	TokenTypeOptionZero    // zero { ... }
	TokenTypeOptionOne     // one { ... }
	TokenTypeOptionTwo     // two { ... }
	TokenTypeOptionFew     // few { ... }
	TokenTypeOptionMany    // many { ... }
	TokenTypeOptionOther   // other { ... }
	TokenTypeOptionNumber  // =2 { ... }

	// Terminator. IndexStart is an index of the token buffer.
	TokenTypeOptionTerm     // } Terminator of an option
	TokenTypeComplexArgTerm // } Terminator of a complex argument
)

func (t TokenType) String() string {
	switch t {
	case TokenTypeLiteral:
		return "literal"
	case TokenTypeSimpleArg:
		return "simple argument"
	case TokenTypePlural:
		return "plural argument"
	case TokenTypePluralOffset:
		return "plural argument offset"
	case TokenTypeSelect:
		return "select argument"
	case TokenTypeSelectOrdinal:
		return "select ordinal argument"
	case TokenTypeArgName:
		return "argument name"
	case TokenTypeArgTypeNumber:
		return "argument type number"
	case TokenTypeArgTypeDate:
		return "argument type date"
	case TokenTypeArgTypeTime:
		return "argument type time"
	case TokenTypeArgTypeSpellout:
		return "argument type spellout"
	case TokenTypeArgTypeOrdinal:
		return "argument type ordinal"
	case TokenTypeArgTypeDuration:
		return "argument type duration"
	case TokenTypeArgStyleShort:
		return "argument style short"
	case TokenTypeArgStyleMedium:
		return "argument style medium"
	case TokenTypeArgStyleLong:
		return "argument style long"
	case TokenTypeArgStyleFull:
		return "argument style full"
	case TokenTypeArgStyleInteger:
		return "argument style integer"
	case TokenTypeArgStyleCurrency:
		return "argument style currency"
	case TokenTypeArgStylePercent:
		return "argument style percent"
	case TokenTypeArgStyleCustom:
		return "argument style custom"
	case TokenTypeOptionName:
		return "option name"
	case TokenTypeOption:
		return "option"
	case TokenTypeOptionZero:
		return "option zero"
	case TokenTypeOptionOne:
		return "option one"
	case TokenTypeOptionTwo:
		return "option two"
	case TokenTypeOptionFew:
		return "option few"
	case TokenTypeOptionMany:
		return "option many"
	case TokenTypeOptionOther:
		return "option other"
	case TokenTypeOptionNumber:
		return "option =n"
	case TokenTypeOptionTerm:
		return "option terminator"
	case TokenTypeComplexArgTerm:
		return "complex argument terminator"
	}
	return "unknown"
}

type Token struct {
	// IndexStart and IndexEnd have different meaning depending on Type.
	// See the token type groups.
	IndexStart, IndexEnd int
	Type                 TokenType
}

type Tokenizer struct {
	loc    language.Tag
	plural cldr.PluralForms
	s      string
	pos    int
}

// Pos returns the last position (byte offset in the input string) the tokenizer was at.
func (t *Tokenizer) Pos() int { return t.pos }

var (
	ErrUnclosedQuote         = errors.New("unclosed quote")
	ErrUnexpectedToken       = errors.New("unexpected token")
	ErrUnexpectedEOF         = errors.New("unexpected EOF")
	ErrExpectedComma         = errors.New("expected comma")
	ErrExpectedColon         = errors.New("expected colon")
	ErrExpectBracketOpen     = errors.New("expect opening bracket")
	ErrExpectBracketClose    = errors.New("expect closing bracket")
	ErrMissingOptionOther    = errors.New("missing the mandatory 'other' option")
	ErrEmptyOption           = errors.New("empty option")
	ErrDuplicateOption       = errors.New("duplicate option")
	ErrInvalidOffset         = errors.New("invalid offset")
	ErrUnsupportedPluralForm = errors.New("plural form unsupported for locale")
)

// String returns a slice of the input string token t represents.
func (t Token) String(s string, buffer []Token) string {
	if t.Type < TokenTypePlural {
		return s[t.IndexStart:t.IndexEnd] // Literals
	} else if t.Type > TokenTypeOptionNumber {
		return s[buffer[t.IndexStart].IndexStart:t.IndexEnd] // Terminators
	}
	// t.Type >= TokenTypePlural && t.Type <= TokenTypeOptionNumber
	return s[t.IndexStart:buffer[t.IndexEnd].IndexEnd] // Complex
}

// Options returns an iterator iterating over all options of a select,
// plural or selectordinal token at buffer[tokenIndex].
// The iterator provides the indexes of option tokens.
// Returns a no-op iterator if buffer[tokenIndex] is neither of:
//
//   - TokenTypeSelect
//   - TokenTypePlural
//   - TokenTypeSelectOrdinal
func Options(buffer []Token, tokenIndex int) iter.Seq[int] {
	switch buffer[tokenIndex].Type {
	case TokenTypeSelect, TokenTypePlural, TokenTypeSelectOrdinal:
	default:
		// Only select, plural and selectordinal can have options.
		return func(yield func(int) bool) {}
	}
	return func(yield func(int) bool) {
		// +1 To skip the argument name.
		for ti := tokenIndex + 2; ti < len(buffer); {
			switch buffer[ti].Type {
			case TokenTypeOption,
				TokenTypeOptionZero,
				TokenTypeOptionOne,
				TokenTypeOptionTwo,
				TokenTypeOptionFew,
				TokenTypeOptionMany,
				TokenTypeOptionOther:
				if !yield(ti) {
					return
				}
				ti = buffer[ti].IndexEnd // Skip contents.
			default:
				ti++
			}
		}
	}
}

// Tokenize resets the tokenizer and appends any tokens encountered to buffer.
func (t *Tokenizer) Tokenize(
	locale language.Tag, buffer []Token, s string,
) ([]Token, error) {
	t.loc, t.s, t.pos = locale, s, 0 // Reset tokenizer.

	{ // Select plural forms
		var ok bool
		if t.plural, ok = cldr.PluralFormsByTag[t.loc]; !ok {
			base, _ := t.loc.Base()
			t.plural = cldr.PluralFormsByBase[base]
		}
	}

	if s == "" {
		return buffer, nil
	}
	if strings.IndexByte(s, '\'') == -1 &&
		strings.IndexByte(s, '{') == -1 &&
		strings.IndexByte(s, '}') == -1 {
		// Fast path for simple inputs.
		return append(buffer, Token{
			IndexStart: 0,
			IndexEnd:   len(s),
			Type:       TokenTypeLiteral,
		}), nil
	}

	var err error
	buffer, err = t.consumeExpr(buffer)
	if err != nil {
		return buffer, err
	}
	if t.pos != len(s) {
		return buffer, ErrUnexpectedToken
	}
	return buffer, nil
}

func (t *Tokenizer) consumeExpr(buffer []Token) ([]Token, error) {
	var err error
	for t.pos < len(t.s) {
		if t.s[t.pos] == '}' {
			break
		}
		if t.s[t.pos] == '{' {
			buffer, err = t.consumeArgument(buffer)
			if err != nil {
				return buffer, err
			}
		} else {
			buffer, err = t.consumeLiteral(buffer)
			if err != nil {
				return buffer, err
			}
		}
	}
	return buffer, nil
}

// indexOfArgNameEnd returns the index of the first rune in s[i:] that is invalid in an ICU argName.
func indexOfArgNameEnd(s string, i int) int {
	for j := i; j < len(s); {
		r, size := rune(s[j]), 1
		if r >= utf8.RuneSelf {
			r, size = utf8.DecodeRuneInString(s[j:])
		}
		if unicode.Is(unicode.Pattern_Syntax, r) ||
			unicode.Is(unicode.Pattern_White_Space, r) {
			return j
		}
		j += size
	}
	return len(s)
}

func (t *Tokenizer) consumeArgument(buffer []Token) ([]Token, error) {
	start := t.pos
	t.pos++ // Consume the '{'.

	t.skipWhitespaces()
	startName := t.pos

	endName := indexOfArgNameEnd(t.s, t.pos)
	t.pos = endName
	t.skipWhitespaces()

	if t.isEOF() {
		return buffer, ErrUnexpectedEOF
	}
	beforeSign := t.pos
	switch t.s[t.pos] {
	case '}':
		// Simple argument.

		if startName == endName {
			return buffer, ErrUnexpectedToken
		}

		t.pos++ // Consume the '}'.
		buffer = append(buffer, Token{
			IndexStart: start,
			IndexEnd:   t.pos,
			Type:       TokenTypeSimpleArg,
		}, Token{
			IndexStart: startName,
			IndexEnd:   endName,
			Type:       TokenTypeArgName,
		})
		return buffer, nil
	case ',':
		// Simple argument with formatting or complex argument.
		t.pos++ // Consume the comma.
		t.skipWhitespaces()
		if t.isEOF() {
			return buffer, ErrUnexpectedEOF
		}

		tokenArgType := t.consumeArgType()
		if tokenArgType.Type != 0 {
			t.skipWhitespaces()
			if t.isEOF() {
				return buffer, ErrUnexpectedEOF
			}
			var tokenArgStyle Token
			if t.s[t.pos] == ',' {
				t.pos++ // Consume the comma.
				t.skipWhitespaces()
				tokenArgStyle = t.consumeArgStyle()
				t.skipWhitespaces()
			}

			if t.isEOF() {
				return buffer, ErrUnexpectedEOF
			}
			if t.s[t.pos] != '}' {
				return buffer, ErrExpectBracketClose
			}
			t.pos++ // Consume the closing bracket.

			buffer = append(buffer, Token{
				IndexStart: start,
				IndexEnd:   t.pos,
				Type:       TokenTypeSimpleArg,
			}, Token{
				IndexStart: startName,
				IndexEnd:   endName,
				Type:       TokenTypeArgName,
			}, tokenArgType)
			if tokenArgStyle.Type != 0 {
				buffer = append(buffer, tokenArgStyle)
			}

			return buffer, nil
		}

		switch {
		case strings.HasPrefix(t.s[t.pos:], "plural"):
			t.pos += len("plural") // Consume "plural".
			t.skipWhitespaces()
			return t.consumePluralArg(buffer, start, startName, endName)
		case strings.HasPrefix(t.s[t.pos:], "selectordinal"):
			t.pos += len("selectordinal") // Consume "selectordinal".
			t.skipWhitespaces()
			return t.consumeSelectOrdinalArg(buffer, start, startName, endName)
		case strings.HasPrefix(t.s[t.pos:], "select"):
			t.pos += len("select") // Consume "select".
			t.skipWhitespaces()
			return t.consumeSelectArg(buffer, start, startName, endName)
		}
		return buffer, ErrUnexpectedToken
	default:
		t.pos = beforeSign // Rollback.
		return buffer, ErrUnexpectedToken
	}
}

func (t *Tokenizer) consumeArgType() (token Token) {
	type TypeValPair struct {
		Value string
		Type  TokenType
	}
	for _, argType := range [...]TypeValPair{
		{"number", TokenTypeArgTypeNumber},
		{"date", TokenTypeArgTypeDate},
		{"time", TokenTypeArgTypeTime},
		{"spellout", TokenTypeArgTypeSpellout},
		{"ordinal", TokenTypeArgTypeOrdinal},
		{"duration", TokenTypeArgTypeDuration},
	} {
		if strings.HasPrefix(t.s[t.pos:], argType.Value) {
			start := t.pos
			t.pos += len(argType.Value) // Consume argType.
			return Token{
				IndexStart: start,
				IndexEnd:   t.pos,
				Type:       argType.Type,
			}
		}
	}
	return Token{}
}

func (t *Tokenizer) consumeArgStyle() (token Token) {
	type TypeValPair struct {
		Value string
		Type  TokenType
	}
	start := t.pos
	for _, argType := range [...]TypeValPair{
		{"short", TokenTypeArgStyleShort},
		{"medium", TokenTypeArgStyleMedium},
		{"long", TokenTypeArgStyleLong},
		{"full", TokenTypeArgStyleFull},
		{"integer", TokenTypeArgStyleInteger},
		{"currency", TokenTypeArgStyleCurrency},
		{"percent", TokenTypeArgStylePercent},
	} {
		if strings.HasPrefix(t.s[t.pos:], argType.Value) {
			t.pos += len(argType.Value) // Consume argType.
			return Token{
				IndexStart: start,
				IndexEnd:   t.pos,
				Type:       argType.Type,
			}
		}
	}

	// Try to parse custom
	end := indexOfArgNameEnd(t.s, t.pos)
	if end != t.pos {
		t.pos = end // Consume the custom arg style.
		return Token{
			IndexStart: start,
			IndexEnd:   end,
			Type:       TokenTypeArgStyleCustom,
		}
	}

	return Token{}
}

func (t *Tokenizer) skipWhitespaces() {
	for ; t.pos < len(t.s); t.pos++ {
		if !isWhitespace(t.s[t.pos]) {
			break
		}
	}
}

// consumeSelectArg consumes the part of the select argument after `{name, select,`
func (t *Tokenizer) consumeSelectArg(
	buffer []Token, start, startName, endName int,
) ([]Token, error) {
	if t.isEOF() {
		return buffer, ErrUnexpectedEOF
	}
	if t.s[t.pos] != ',' {
		return buffer, ErrExpectedComma
	}
	t.pos++ // Consume the comma.
	t.skipWhitespaces()

	initiatorBufIndex := len(buffer)

	buffer = append(buffer, Token{
		IndexStart: start,
		IndexEnd:   0, // This is determined later.
		Type:       TokenTypeSelect,
	}, Token{
		IndexStart: startName,
		IndexEnd:   endName,
		Type:       TokenTypeArgName,
	})
	for {
		t.skipWhitespaces()
		if t.isEOF() {
			return buffer, ErrUnexpectedEOF
		}
		if t.s[t.pos] == '}' {
			t.pos++ // Consume the closing bracket.
			break
		}

		var err error
		buffer, err = t.consumeOption(buffer)
		if err != nil {
			return buffer, err
		}
	}

	// +2 to skip [plural,argName]
	if err := t.validateOptions(buffer, initiatorBufIndex+2, start); err != nil {
		return buffer, err
	}

	// Link the argument initiator to the argument terminator.
	buffer[initiatorBufIndex].IndexEnd = len(buffer)
	buffer = append(buffer, Token{
		IndexStart: initiatorBufIndex,
		IndexEnd:   t.pos,
		Type:       TokenTypeComplexArgTerm,
	})
	return buffer, nil
}

var ErrInvalidOption = errors.New("invalid plural option")

func (t *Tokenizer) consumeOption(buffer []Token) ([]Token, error) {
	start := t.pos
	var initiatorBufIndex int
	tp := TokenTypeOption

	end := indexOfArgNameEnd(t.s, t.pos)
	if start == end {
		return buffer, ErrInvalidOption
	}
	if t.s[start:end] == "other" {
		tp = TokenTypeOptionOther
	}
	t.pos = end

	initiatorBufIndex = len(buffer)
	buffer = append(buffer, Token{
		IndexStart: start,
		IndexEnd:   0, // Set later to terminator buffer index.
		Type:       tp,
	})
	if tp == TokenTypeOption {
		buffer = append(buffer, Token{
			IndexStart: start,
			IndexEnd:   t.pos,
			Type:       TokenTypeOptionName,
		})
	}

	t.skipWhitespaces()
	if t.isEOF() {
		return buffer, ErrUnexpectedEOF
	}
	bracketOpen := t.pos
	if t.s[t.pos] != '{' {
		return buffer, ErrExpectBracketOpen
	}
	t.pos++ // Consume the opening bracket.
	t.skipWhitespaces()

	{
		afterOpeningBracket := t.pos
		t.skipWhitespaces()
		if t.isEOF() {
			return buffer, ErrUnexpectedEOF
		}
		if t.s[t.pos] == '}' {
			t.pos = bracketOpen // Rollback to begin of block.
			return buffer, ErrEmptyOption
		}
		t.pos = afterOpeningBracket // Revert to before the lookahead.
	}

	var err error
	buffer, err = t.consumeExpr(buffer)
	if err != nil {
		return buffer, err
	}

	if t.isEOF() {
		return buffer, ErrUnexpectedEOF
	}
	if t.s[t.pos] != '}' {
		return buffer, ErrExpectBracketClose
	}
	t.pos++ // Consume closing bracket.

	// Link the argument initiator to the argument terminator.
	buffer[initiatorBufIndex].IndexEnd = len(buffer)
	buffer = append(buffer, Token{
		IndexStart: initiatorBufIndex,
		IndexEnd:   t.pos,
		Type:       TokenTypeOptionTerm,
	})

	return buffer, nil
}

func (t *Tokenizer) consumeOptionPlural(buffer []Token, f cldr.Forms) ([]Token, error) {
	start := t.pos
	tp := TokenTypeOptionNumber
	var initiatorBufIndex int
	numStart := t.pos
	if t.s[t.pos] == '=' {
		t.pos++ // Consume the equal sign.
		digitsStart := t.pos
		for {
			if t.isEOF() {
				return buffer, ErrUnexpectedEOF
			}
			if t.s[t.pos] < '0' || t.s[t.pos] > '9' {
				break
			}
			t.pos++
		}
		if digitsStart == t.pos {
			return buffer, ErrInvalidOption // '=' not followed by digits.
		}
		option := t.s[start:t.pos]
		if len(option) > 2 && option[1] == '0' {
			t.pos = start
			return buffer, ErrInvalidOption // Leading zero is illegal.
		}
	} else {
	LOOP:
		for ; t.pos < len(t.s); t.pos++ {
			b := t.s[t.pos]
			switch b {
			case '{', '}', ',', ' ', '\t', '\n', '\r':
				break LOOP
			}
		}

		option := t.s[start:t.pos]
		switch option {
		case "zero":
			if !f.Zero {
				t.pos = start // Rollback.
				return buffer, ErrUnsupportedPluralForm
			}
			tp = TokenTypeOptionZero
		case "one":
			if !f.One {
				t.pos = start // Rollback.
				return buffer, ErrUnsupportedPluralForm
			}
			tp = TokenTypeOptionOne
		case "two":
			if !f.Two {
				t.pos = start // Rollback.
				return buffer, ErrUnsupportedPluralForm
			}
			tp = TokenTypeOptionTwo
		case "few":
			if !f.Few {
				t.pos = start // Rollback.
				return buffer, ErrUnsupportedPluralForm
			}
			tp = TokenTypeOptionFew
		case "many":
			if !f.Many {
				t.pos = start // Rollback.
				return buffer, ErrUnsupportedPluralForm
			}
			tp = TokenTypeOptionMany
		case "other":
			tp = TokenTypeOptionOther
		default:
			t.pos = start // Roll back to start.
			return buffer, ErrInvalidOption
		}
	}

	initiatorBufIndex = len(buffer)
	buffer = append(buffer, Token{
		IndexStart: start,
		IndexEnd:   0, // Set later to terminator buffer index.
		Type:       tp,
	})
	if tp == TokenTypeOptionNumber {
		buffer = append(buffer, Token{
			IndexStart: numStart,
			IndexEnd:   t.pos,
			Type:       TokenTypeOptionName,
		})
	}

	t.skipWhitespaces()
	if t.isEOF() {
		return buffer, ErrUnexpectedEOF
	}
	bracketOpen := t.pos
	if t.s[t.pos] != '{' {
		return buffer, ErrExpectBracketOpen
	}
	t.pos++ // Consume the opening bracket.

	{
		afterOpeningBracket := t.pos
		t.skipWhitespaces()
		if t.isEOF() {
			return buffer, ErrUnexpectedEOF
		}
		if t.s[t.pos] == '}' {
			t.pos = bracketOpen // Rollback to begin of block.
			return buffer, ErrEmptyOption
		}
		t.pos = afterOpeningBracket // Revert to before the lookahead.
	}

	var err error
	buffer, err = t.consumeExpr(buffer)
	if err != nil {
		return buffer, err
	}

	if t.isEOF() {
		return buffer, ErrUnexpectedEOF
	}
	if t.s[t.pos] != '}' {
		return buffer, ErrExpectBracketClose
	}

	// Link the argument initiator to the argument terminator.
	buffer[initiatorBufIndex].IndexEnd = len(buffer)
	t.pos++ // Consume closing bracket.
	buffer = append(buffer, Token{
		IndexStart: initiatorBufIndex,
		IndexEnd:   t.pos,
		Type:       TokenTypeOptionTerm,
	})

	return buffer, nil
}

// consumeSelectOrdinalArg consumes the part of the select argument
// after `{name, selectordinal,`
func (t *Tokenizer) consumeSelectOrdinalArg(
	buffer []Token, start, startName, endName int,
) ([]Token, error) {
	if t.isEOF() {
		return buffer, ErrUnexpectedEOF
	}
	if t.s[t.pos] != ',' {
		return buffer, ErrExpectedComma
	}
	t.pos++ // Consume the comma.
	t.skipWhitespaces()

	initiatorBufIndex := len(buffer)

	buffer = append(buffer, Token{
		IndexStart: start,
		IndexEnd:   0, // This is determined later.
		Type:       TokenTypeSelectOrdinal,
	}, Token{
		IndexStart: startName,
		IndexEnd:   endName,
		Type:       TokenTypeArgName,
	})
	for {
		t.skipWhitespaces()
		if t.isEOF() {
			return buffer, ErrUnexpectedEOF
		}
		if t.s[t.pos] == '}' {
			t.pos++ // Consume the closing bracket.
			break
		}

		var err error
		buffer, err = t.consumeOptionPlural(buffer, t.plural.Ordinal)
		if err != nil {
			return buffer, err
		}
	}

	// +2 to skip [plural,argName]
	if err := t.validateOptions(buffer, initiatorBufIndex+2, start); err != nil {
		return buffer, err
	}

	// TODO: check illegal options relative to the selected base lang.

	// Link the argument initiator to the argument terminator.
	buffer[initiatorBufIndex].IndexEnd = len(buffer)
	buffer = append(buffer, Token{
		IndexStart: initiatorBufIndex,
		IndexEnd:   t.pos,
		Type:       TokenTypeComplexArgTerm,
	})
	return buffer, nil
}

func (t *Tokenizer) consumePluralArg(
	buffer []Token, start, startName, endName int,
) ([]Token, error) {
	if t.isEOF() {
		return buffer, ErrUnexpectedEOF
	}
	if t.s[t.pos] != ',' {
		return buffer, ErrExpectedComma
	}
	t.pos++ // Consume the comma.
	t.skipWhitespaces()

	initiatorBufIndex := len(buffer)

	buffer = append(buffer, Token{
		IndexStart: start,
		IndexEnd:   0, // This is determined later.
		Type:       TokenTypePlural,
	}, Token{
		IndexStart: startName,
		IndexEnd:   endName,
		Type:       TokenTypeArgName,
	})

	// Check for optional "offset" parameter.
	if t.isEOF() {
		return buffer, ErrUnexpectedEOF
	}
	if strings.HasPrefix(t.s[t.pos:], "offset") {
		t.pos += len("offset") // Consume "offset".
		t.skipWhitespaces()
		if t.isEOF() {
			return buffer, ErrUnexpectedEOF
		}
		if t.s[t.pos] != ':' {
			return buffer, ErrExpectedColon
		}
		t.pos++ // Consume the colon.
		t.skipWhitespaces()

		var err error
		buffer, err = t.consumePluralOffsetNum(buffer)
		if err != nil {
			return buffer, err
		}
		t.skipWhitespaces()

		if t.isEOF() {
			return buffer, ErrUnexpectedEOF
		}
		if t.s[t.pos] == ',' {
			t.pos++ // Consume optional comma.
			t.skipWhitespaces()
		}
	}
	for {
		t.skipWhitespaces()
		if t.isEOF() {
			return buffer, ErrUnexpectedEOF
		}
		if t.s[t.pos] == '}' {
			t.pos++ // Consume the closing bracket.
			break
		}

		var err error
		buffer, err = t.consumeOptionPlural(buffer, t.plural.Cardinal)
		if err != nil {
			return buffer, err
		}
	}

	// +2 to skip [plural,argName]
	if err := t.validateOptions(buffer, initiatorBufIndex+2, start); err != nil {
		return buffer, err
	}

	// Link the argument initiator to the argument terminator.
	buffer[initiatorBufIndex].IndexEnd = len(buffer)
	buffer = append(buffer, Token{
		IndexStart: initiatorBufIndex,
		IndexEnd:   t.pos,
		Type:       TokenTypeComplexArgTerm,
	})
	return buffer, nil
}

func (t *Tokenizer) validateOptions(buffer []Token, bufIndex, startArg int) error {
	var zero, one, two, few, many, other bool
	for i := bufIndex; i < len(buffer); i++ {
		outer := buffer[i]
		switch outer.Type {
		case TokenTypeOptionZero:
			if zero {
				t.pos = outer.IndexStart
				return ErrDuplicateOption
			}
			zero = true
			i = outer.IndexEnd // Skip contents.
		case TokenTypeOptionOne:
			if one {
				t.pos = outer.IndexStart
				return ErrDuplicateOption
			}
			one = true
			i = outer.IndexEnd // Skip contents.
		case TokenTypeOptionTwo:
			if two {
				t.pos = outer.IndexStart
				return ErrDuplicateOption
			}
			two = true
			i = outer.IndexEnd // Skip contents.
		case TokenTypeOptionFew:
			if few {
				t.pos = outer.IndexStart
				return ErrDuplicateOption
			}
			few = true
			i = outer.IndexEnd // Skip contents.
		case TokenTypeOptionMany:
			if many {
				t.pos = outer.IndexStart
				return ErrDuplicateOption
			}
			many = true
			i = outer.IndexEnd // Skip contents.
		case TokenTypeOptionOther:
			if other {
				t.pos = outer.IndexStart
				return ErrDuplicateOption
			}
			other = true
			i = outer.IndexEnd // Skip contents.
		case TokenTypeOptionNumber, TokenTypeOption:
			nameToken := buffer[i+1]
			name := t.s[nameToken.IndexStart:nameToken.IndexEnd]
			// Check each other option.
			for j := bufIndex; j < len(buffer); j++ {
				inner := buffer[j]
				if j == i {
					continue
				}
				switch inner.Type {
				case outer.Type:
					j++ // Skip the option and go straight to name.
					inner = buffer[j]
					inr := t.s[inner.IndexStart:inner.IndexEnd]
					if name == inr {
						t.pos = inner.IndexStart
						return ErrDuplicateOption
					}
				case TokenTypeOptionZero,
					TokenTypeOptionOne,
					TokenTypeOptionTwo,
					TokenTypeOptionFew,
					TokenTypeOptionMany,
					TokenTypeOptionOther:
					j = inner.IndexEnd // Skip contents.
				}
			}
			i = outer.IndexEnd // Skip contents.
		}
	}
	if !other {
		t.pos = startArg // Rollback.
		return ErrMissingOptionOther
	}
	return nil
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func (t *Tokenizer) isEOF() bool { return t.pos >= len(t.s) }

func (t *Tokenizer) consumePluralOffsetNum(buffer []Token) ([]Token, error) {
	if t.isEOF() {
		return buffer, ErrUnexpectedEOF
	}
	start := t.pos
	if t.s[t.pos] == '0' {
		t.pos++ // Consume the zero as the number since leading zeros are not allowed.
		buffer = append(buffer, Token{
			IndexStart: start,
			IndexEnd:   t.pos,
			Type:       TokenTypePluralOffset,
		})
		return buffer, nil
	}
	for ; t.pos < len(t.s); t.pos++ {
		if t.s[t.pos] < '0' || t.s[t.pos] > '9' {
			// End of number.
			if start == t.pos {
				return buffer, ErrInvalidOffset
			}

			buffer = append(buffer, Token{
				IndexStart: start,
				IndexEnd:   t.pos,
				Type:       TokenTypePluralOffset,
			})
			break
		}
	}
	return buffer, nil
}

var endOfLiteral = [256]bool{'\'': true, '{': true, '}': true}

func (t *Tokenizer) consumeLiteral(buffer []Token) ([]Token, error) {
	start := t.pos
	inQuote := false
	quoteStart := start

	for t.pos < len(t.s) {
		if t.pos+8 < len(t.s) {
			if endOfLiteral[t.s[t.pos]] {
				goto CHECK
			}
			if endOfLiteral[t.s[t.pos+1]] {
				t.pos++
				goto CHECK
			}
			if endOfLiteral[t.s[t.pos+2]] {
				t.pos += 2
				goto CHECK
			}
			if endOfLiteral[t.s[t.pos+3]] {
				t.pos += 3
				goto CHECK
			}
			if endOfLiteral[t.s[t.pos+4]] {
				t.pos += 4
				goto CHECK
			}
			if endOfLiteral[t.s[t.pos+5]] {
				t.pos += 5
				goto CHECK
			}
			if endOfLiteral[t.s[t.pos+6]] {
				t.pos += 6
				goto CHECK
			}
			if endOfLiteral[t.s[t.pos+7]] {
				t.pos += 7
				goto CHECK
			}
			t.pos += 8
			continue
		}

	CHECK:
		b := t.s[t.pos]
		if b == '\'' {
			// Lookahead for escaped quote
			if t.pos+1 < len(t.s) && t.s[t.pos+1] == '\'' {
				t.pos += 2 // skip both
				continue
			}
			inQuote = !inQuote
			if inQuote {
				quoteStart = t.pos
			}
			t.pos++
			continue
		}

		if !inQuote && (b == '{' || b == '}') {
			break // End of literal.
		}

		t.pos++
	}

	if inQuote {
		t.pos = quoteStart // Rollback.
		return buffer, ErrUnclosedQuote
	}
	if t.pos > start {
		buffer = append(buffer, Token{
			IndexStart: start,
			IndexEnd:   t.pos,
			Type:       TokenTypeLiteral,
		})
	}
	return buffer, nil
}
