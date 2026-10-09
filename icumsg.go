// Package icumsg provides a tokenizer for ICU MessageFormat 1.0 (ICU 4.8 and later).
// Its successor MessageFormat 2.0 (UTS #35 Part 9) is not supported.
// See https://unicode-org.github.io/icu/userguide/format_parse/messages/
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

	// Literal. [Token.IndexStart] and [Token.IndexEnd] are
	// byte offsets in the input string.

	TokenTypeLiteral      // Any literal
	TokenTypeSimpleArg    // { arg }
	TokenTypePluralOffset // offset:1
	TokenTypeArgName      // The name of any argument

	// The following token types always follow [TokenTypeArgName].
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
	TokenTypeArgStyleSkeleton

	// TokenTypeOptionName is the option name. Always follows [TokenTypeOption].
	TokenTypeOptionName

	// Complex. [Token.IndexEnd] is the index of the terminator
	// among the tokens of the message (see [Tokenizer.Tokenize]).

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

	// Terminator. [Token.IndexStart] is the index of the terminated token
	// among the tokens of the message (see [Tokenizer.Tokenize]).

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
	case TokenTypeArgStyleSkeleton:
		return "argument style skeleton"
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
	// [Token.IndexStart] and [Token.IndexEnd] have different meaning
	// depending on [Token.Type]. See the token type groups.
	IndexStart, IndexEnd int
	Type                 TokenType
}

type Tokenizer struct {
	loc    language.Tag
	plural cldr.PluralRules
	s      string
	pos    int
}

// Pos returns the last position (byte offset in the input string) the tokenizer was at.
// That's where Tokenize found an error, or the end of the input if it succeeded.
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
	ErrUnsupportedPluralRule = errors.New("plural rule unsupported for locale")
)

// String returns a slice of the input string token t represents.
// buffer must hold the tokens of s (see Tokenize).
func (t Token) String(s string, buffer []Token) string {
	if t.Type < TokenTypePlural {
		return s[t.IndexStart:t.IndexEnd] // Literals
	} else if t.Type > TokenTypeOptionNumber {
		return s[buffer[t.IndexStart].IndexStart:t.IndexEnd] // Terminators
	}
	// t.Type >= [TokenTypePlural] && t.Type <= [TokenTypeOptionNumber]
	return s[t.IndexStart:buffer[t.IndexEnd].IndexEnd] // Complex
}

// Options returns an iterator iterating over all options of a select,
// plural or selectordinal token at buffer[tokenIndex].
// The iterator provides the indexes of option tokens.
// Returns a no-op iterator if buffer[tokenIndex] is neither of:
//
//   - [TokenTypeSelect]
//   - [TokenTypePlural]
//   - [TokenTypeSelectOrdinal]
func Options(buffer []Token, tokenIndex int) iter.Seq[int] {
	var endIndex int
	switch buffer[tokenIndex].Type {
	case TokenTypeSelect, TokenTypePlural, TokenTypeSelectOrdinal:
		endIndex = buffer[tokenIndex].IndexEnd
	default:
		// Only select, plural and selectordinal can have options.
		return func(yield func(int) bool) {}
	}
	return func(yield func(int) bool) {
		// +1 To skip the argument name.
		for ti := tokenIndex + 2; ti < endIndex; {
			switch buffer[ti].Type {
			case TokenTypeOption,
				TokenTypeOptionZero,
				TokenTypeOptionOne,
				TokenTypeOptionTwo,
				TokenTypeOptionFew,
				TokenTypeOptionMany,
				TokenTypeOptionOther,
				TokenTypeOptionNumber:
				if !yield(ti) {
					return
				}
				// Skip contents, but never back, so that the tokens of more
				// than one message (see Tokenize) can't make this loop forever.
				ti = max(ti+1, buffer[ti].IndexEnd)
			default:
				ti++
			}
		}
	}
}

// OptionsPresencePolicy defines treatment of known options.
type OptionsPresencePolicy int8

const (
	// OptionsPresencePolicyOptional does not require all select options to be present
	// for the ICU message to be considered complete.
	OptionsPresencePolicyOptional OptionsPresencePolicy = iota

	// OptionsPresencePolicyRequired requires all select options to be present
	// for the ICU message to be considered complete.
	OptionsPresencePolicyRequired
)

// OptionUnknownPolicy defines treatment of unknown select options.
type OptionUnknownPolicy int8

const (
	// OptionUnknownPolicyIgnore ignores unknown select options.
	OptionUnknownPolicyIgnore OptionUnknownPolicy = iota
	// OptionUnknownPolicyReject rejects unknown select options.
	OptionUnknownPolicyReject
)

type SelectOptions func(argName string) (
	[]string, OptionsPresencePolicy, OptionUnknownPolicy,
)

// Tokenize resets the tokenizer and appends the tokens of s to buffer.
// Tokens link to each other by their index among the tokens of s,
// so [Token.String], [Options], [Analyze] and [Errors] take the tokens of s
// on their own, even if buffer wasn't empty:
//
//	n := len(buffer)
//	buffer, err = tokenizer.Tokenize(locale, buffer, s)
//	tokens := buffer[n:] // The tokens of s.
func (t *Tokenizer) Tokenize(
	locale language.Tag, buffer []Token, s string,
) ([]Token, error) {
	t.loc, t.s, t.pos = locale, s, 0 // Reset tokenizer.
	t.plural = cldr.Lookup(t.loc)

	if s == "" {
		return buffer, nil
	}
	if strings.IndexByte(s, '{') == -1 && strings.IndexByte(s, '}') == -1 {
		// Fast path for simple inputs.
		// Without braces there's no quoted text because every syntax character
		// an apostrophe can quote is a brace or appears only inside braces.
		// See startsQuote.
		t.pos = len(s)
		return append(buffer, Token{
			IndexStart: 0,
			IndexEnd:   len(s),
			Type:       TokenTypeLiteral,
		}), nil
	}

	// Tokenize into an empty slice so that the indexes linking tokens
	// count from the first token of s.
	tokens, err := t.consumeExpr(buffer[len(buffer):], false)
	if len(buffer) == 0 {
		buffer = tokens
	} else {
		buffer = append(buffer, tokens...)
	}
	if err != nil {
		return buffer, err
	}
	if t.pos != len(s) {
		return buffer, ErrUnexpectedToken
	}
	return buffer, nil
}

// consumeExpr consumes a message or sub-message.
// pluralStyle must be true inside a plural or selectordinal option.
// See startsQuote.
func (t *Tokenizer) consumeExpr(buffer []Token, pluralStyle bool) ([]Token, error) {
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
			buffer, err = t.consumeLiteral(buffer, pluralStyle)
			if err != nil {
				return buffer, err
			}
		}
	}
	return buffer, nil
}

// indexOfArgNameEnd returns the index of the first rune in s[i:]
// that is invalid in an ICU argName.
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
				var err error
				tokenArgStyle, err = t.consumeArgStyle()
				if err != nil {
					return buffer, err
				}
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

// consumeArgStyle consumes the argStyle of a simple argument.
// As in [ICU MessageFormat], the style ends before the first '}' that's
// neither quoted nor closes a '{' within the style, and every apostrophe
// in it starts or ends quoted text. A custom style is thus a string
// pattern that may contain any syntax, such as "#,##0.00" or "yyyy-MM-dd".
// Trailing whitespace isn't part of the style.
//
// [ICU MessageFormat]: https://unicode-org.github.io/icu-docs/apidoc/released/icu4j/com/ibm/icu/text/MessageFormat.html
func (t *Tokenizer) consumeArgStyle() (token Token, err error) {
	type TypeValPair struct {
		Value string
		Type  TokenType
	}
	start := t.pos

	nestedBraces := 0
LOOP:
	for ; t.pos < len(t.s); t.pos++ {
		switch t.s[t.pos] {
		case '\'':
			n := strings.IndexByte(t.s[t.pos+1:], '\'')
			if n == -1 {
				return Token{}, ErrUnclosedQuote
			}
			t.pos += 1 + n // Skip to the closing apostrophe.
		case '{':
			nestedBraces++
		case '}':
			if nestedBraces == 0 {
				break LOOP
			}
			nestedBraces--
		}
	}
	end := t.pos
	for end > start && isWhitespace(t.s[end-1]) {
		end--
	}
	t.pos = end
	style := t.s[start:end]

	if strings.HasPrefix(style, "::") {
		if style == "::" {
			t.pos = start + 2 // Rollback to after the "::".
			if t.isEOF() {
				return Token{}, ErrUnexpectedEOF
			}
			return Token{}, ErrUnexpectedToken
		}
		return Token{
			IndexStart: start,
			IndexEnd:   end,
			Type:       TokenTypeArgStyleSkeleton,
		}, nil
	}

	for _, argStyle := range [...]TypeValPair{
		{"short", TokenTypeArgStyleShort},
		{"medium", TokenTypeArgStyleMedium},
		{"long", TokenTypeArgStyleLong},
		{"full", TokenTypeArgStyleFull},
		{"integer", TokenTypeArgStyleInteger},
		{"currency", TokenTypeArgStyleCurrency},
		{"percent", TokenTypeArgStylePercent},
	} {
		if style == argStyle.Value {
			return Token{
				IndexStart: start,
				IndexEnd:   end,
				Type:       argStyle.Type,
			}, nil
		}
	}

	if style == "" {
		return Token{}, nil
	}
	return Token{
		IndexStart: start,
		IndexEnd:   end,
		Type:       TokenTypeArgStyleCustom,
	}, nil
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

	// +2 to skip [select,argName]
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
	// '#' is not a syntax character in a select option.
	buffer, err = t.consumeExpr(buffer, false)
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

func (t *Tokenizer) consumeOptionPlural(buffer []Token, f cldr.Rules) ([]Token, error) {
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
				return buffer, ErrUnsupportedPluralRule
			}
			tp = TokenTypeOptionZero
		case "one":
			if !f.One {
				t.pos = start // Rollback.
				return buffer, ErrUnsupportedPluralRule
			}
			tp = TokenTypeOptionOne
		case "two":
			if !f.Two {
				t.pos = start // Rollback.
				return buffer, ErrUnsupportedPluralRule
			}
			tp = TokenTypeOptionTwo
		case "few":
			if !f.Few {
				t.pos = start // Rollback.
				return buffer, ErrUnsupportedPluralRule
			}
			tp = TokenTypeOptionFew
		case "many":
			if !f.Many {
				t.pos = start // Rollback.
				return buffer, ErrUnsupportedPluralRule
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
	// '#' is a syntax character in plural and selectordinal options.
	buffer, err = t.consumeExpr(buffer, true)
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

// consumeSelectOrdinalArg consumes the part of the selectordinal argument
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

	// +2 to skip [selectordinal,argName]
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
			// Check each other sibling option.
			// Their contents are skipped because nested arguments may reuse names.
			for j := bufIndex; j < len(buffer); j++ {
				inner := buffer[j]
				switch inner.Type {
				case TokenTypeOption,
					TokenTypeOptionZero,
					TokenTypeOptionOne,
					TokenTypeOptionTwo,
					TokenTypeOptionFew,
					TokenTypeOptionMany,
					TokenTypeOptionOther,
					TokenTypeOptionNumber:
					if j != i && inner.Type == outer.Type {
						innerName := buffer[j+1]
						if name == t.s[innerName.IndexStart:innerName.IndexEnd] {
							t.pos = innerName.IndexStart
							return ErrDuplicateOption
						}
					}
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

// startsQuote reports whether an apostrophe immediately followed by b starts
// quoted literal text. pluralStyle must be true inside a plural or
// selectordinal option where '#' is a syntax character.
//
// ICU also quotes '|' inside choice arguments which this tokenizer doesn't support.
func startsQuote(b byte, pluralStyle bool) bool {
	return b == '{' || b == '}' || (pluralStyle && b == '#')
}

// consumeLiteral consumes literal text until the next unquoted '{' or '}'.
// pluralStyle is passed through to startsQuote.
//
// Apostrophes follow ICU's default ApostropheMode.DOUBLE_OPTIONAL:
//
//   - A pair of apostrophes is one literal apostrophe, both inside and outside
//     of quoted text.
//   - A single apostrophe starts quoted text only where startsQuote allows it.
//     Everywhere else it's literal text and needs no escaping, as in "aujourd'hui".
//
// Unclosed quoted text is rejected with ErrUnclosedQuote.
// ICU instead auto-quotes it to the end of the message.
//
// See https://unicode-org.github.io/icu/userguide/format_parse/messages/#quotingescaping
// and https://unicode-org.github.io/icu-docs/apidoc/released/icu4j/com/ibm/icu/text/MessagePattern.ApostropheMode.html
func (t *Tokenizer) consumeLiteral(buffer []Token, pluralStyle bool) ([]Token, error) {
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
			if inQuote {
				inQuote = false
				t.pos++
				continue
			}
			if t.pos+1 >= len(t.s) || !startsQuote(t.s[t.pos+1], pluralStyle) {
				t.pos++
				continue
			}
			inQuote, quoteStart = true, t.pos
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
