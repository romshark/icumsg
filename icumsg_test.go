package icumsg_test

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/romshark/icumsg"
)

func requireEqual[T comparable](tb testing.TB, expect, actual T) {
	tb.Helper()
	if expect != actual {
		tb.Fatalf("\nexpected: %#v;\nreceived: %#v", expect, actual)
	}
}

func requireErrIs(t *testing.T, expect, actual error) {
	t.Helper()
	if !errors.Is(actual, expect) {
		t.Fatalf("\nexpected: %#v;\nreceived: %#v", expect, actual)
	}
}

func requireNoErr(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("\nexpected: no error;\nreceived: %#v", err)
	}
}

func ReadFile[S string | []byte](tb testing.TB, fileName string) S {
	tb.Helper()
	fc, err := os.ReadFile(fileName)
	requireNoErr(tb, err)
	return S(fc)
}

type Token struct {
	Str  string
	Type icumsg.TokenType
}

// ToTestTokens creates a slice of simplified tokens with their string values.
func ToTestTokens(input string, toks []icumsg.Token) []Token {
	if len(toks) == 0 {
		return nil
	}
	tokens := make([]Token, len(toks))
	for i, tok := range toks {
		tokens[i] = Token{
			Str:  tok.String(input, toks),
			Type: tok.Type,
		}
	}
	return tokens
}

func compareTokens(t *testing.T, expect, actual []Token) {
	t.Helper()
	if !reflect.DeepEqual(expect, actual) {
		for i := range actual {
			if i >= len(expect) || !reflect.DeepEqual(expect[i], actual[i]) {
				t.Logf("diff at index %d\n\n", i)
				break
			}
		}

		t.Logf("expected (%d):", len(expect))
		for i, e := range expect {
			t.Logf(" %d (%s): %q:", i, e.Type.String(), e.Str)
		}
		t.Logf("received (%d):", len(actual))
		for i, e := range actual {
			t.Logf(" %d (%s): %q:", i, e.Type.String(), e.Str)
		}
		t.Error("unexpected tokens")
	}
}

func TestTokenTypeString(t *testing.T) {
	f := func(t *testing.T, expect string, tp icumsg.TokenType) {
		t.Helper()
		requireEqual(t, expect, tp.String())
	}

	f(t, "unknown", 0)
	f(t, "literal", icumsg.TokenTypeLiteral)
	f(t, "simple argument", icumsg.TokenTypeSimpleArg)
	f(t, "plural argument offset", icumsg.TokenTypePluralOffset)
	f(t, "argument name", icumsg.TokenTypeArgName)
	f(t, "argument type number", icumsg.TokenTypeArgTypeNumber)
	f(t, "argument type date", icumsg.TokenTypeArgTypeDate)
	f(t, "argument type time", icumsg.TokenTypeArgTypeTime)
	f(t, "argument type spellout", icumsg.TokenTypeArgTypeSpellout)
	f(t, "argument type ordinal", icumsg.TokenTypeArgTypeOrdinal)
	f(t, "argument type duration", icumsg.TokenTypeArgTypeDuration)
	f(t, "argument style short", icumsg.TokenTypeArgStyleShort)
	f(t, "argument style medium", icumsg.TokenTypeArgStyleMedium)
	f(t, "argument style long", icumsg.TokenTypeArgStyleLong)
	f(t, "argument style full", icumsg.TokenTypeArgStyleFull)
	f(t, "argument style integer", icumsg.TokenTypeArgStyleInteger)
	f(t, "argument style currency", icumsg.TokenTypeArgStyleCurrency)
	f(t, "argument style percent", icumsg.TokenTypeArgStylePercent)
	f(t, "argument style custom", icumsg.TokenTypeArgStyleCustom)
	f(t, "option name", icumsg.TokenTypeOptionName)
	f(t, "plural argument", icumsg.TokenTypePlural)
	f(t, "select argument", icumsg.TokenTypeSelect)
	f(t, "select ordinal argument", icumsg.TokenTypeSelectOrdinal)
	f(t, "option", icumsg.TokenTypeOption)
	f(t, "option zero", icumsg.TokenTypeOptionZero)
	f(t, "option one", icumsg.TokenTypeOptionOne)
	f(t, "option two", icumsg.TokenTypeOptionTwo)
	f(t, "option few", icumsg.TokenTypeOptionFew)
	f(t, "option many", icumsg.TokenTypeOptionMany)
	f(t, "option other", icumsg.TokenTypeOptionOther)
	f(t, "option =n", icumsg.TokenTypeOptionNumber)
	f(t, "option terminator", icumsg.TokenTypeOptionTerm)
	f(t, "complex argument terminator", icumsg.TokenTypeComplexArgTerm)
}

func TestTokenize(t *testing.T) {
	t.Parallel()

	var tokenizer icumsg.Tokenizer
	var buffer []icumsg.Token
	f := func(t *testing.T, input string, expect ...Token) {
		t.Helper()
		buffer = buffer[:0]
		buffer, err := tokenizer.Tokenize(buffer, input)
		requireNoErr(t, err)
		actual := ToTestTokens(input, buffer)
		compareTokens(t, expect, actual)
	}

	f(t, "")
	f(t, "foo", []Token{
		{Str: "foo", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, "foo bar\n\tbazz", []Token{
		{Str: "foo bar\n\tbazz", Type: icumsg.TokenTypeLiteral},
	}...)

	// Escaping
	f(t, "''", []Token{
		{Str: "''", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, "'{}' '{}'", []Token{
		{Str: "'{}' '{}'", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, "before '{x '' y}' after", []Token{
		{Str: "before '{x '' y}' after", Type: icumsg.TokenTypeLiteral},
	}...)

	// Argument
	f(t, "{_}", []Token{
		{Str: "{_}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "_", Type: icumsg.TokenTypeArgName},
	}...)
	f(t, "{1}", []Token{
		{Str: "{1}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "1", Type: icumsg.TokenTypeArgName},
	}...)
	f(t, "{arg}", []Token{
		{Str: "{arg}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
	}...)
	f(t, "{аргумент}", []Token{
		{Str: "{аргумент}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "аргумент", Type: icumsg.TokenTypeArgName},
	}...)
	f(t, "{ arg }", []Token{
		{Str: "{ arg }", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
	}...)
	f(t, "{\n arg \n}", []Token{
		{Str: "{\n arg \n}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
	}...)

	// Argument type
	f(t, "Before {arg, number} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, number}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, "Before {arg, date} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, date}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "date", Type: icumsg.TokenTypeArgTypeDate},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, "Before {arg, time} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, time}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "time", Type: icumsg.TokenTypeArgTypeTime},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, "Before {arg, spellout} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, spellout}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "spellout", Type: icumsg.TokenTypeArgTypeSpellout},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, "Before {arg, ordinal} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, ordinal}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "ordinal", Type: icumsg.TokenTypeArgTypeOrdinal},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, "Before {arg, duration} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, duration}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "duration", Type: icumsg.TokenTypeArgTypeDuration},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)

	// Argument style
	f(t, "Before {arg, number, short} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, number, short}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "short", Type: icumsg.TokenTypeArgStyleShort},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, "Before {arg, number, medium} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, number, medium}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "medium", Type: icumsg.TokenTypeArgStyleMedium},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, "Before {arg, number, long} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, number, long}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "long", Type: icumsg.TokenTypeArgStyleLong},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, "Before {arg, number, full} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, number, full}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "full", Type: icumsg.TokenTypeArgStyleFull},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, "Before {arg, number, integer} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, number, integer}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "integer", Type: icumsg.TokenTypeArgStyleInteger},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, "Before {arg, number, percent} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, number, percent}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "percent", Type: icumsg.TokenTypeArgStylePercent},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, "Before {arg, number, customAnything} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, number, customAnything}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "customAnything", Type: icumsg.TokenTypeArgStyleCustom},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)

	// Plural
	f(t, "{var,plural,other{#messages}one{#message}}", []Token{
		{
			Str:  "{var,plural,other{#messages}one{#message}}",
			Type: icumsg.TokenTypePlural,
		},
		{Str: "var", Type: icumsg.TokenTypeArgName},
		{Str: "other{#messages}", Type: icumsg.TokenTypeOptionOther},
		{Str: "#messages", Type: icumsg.TokenTypeLiteral},
		{Str: "other{#messages}", Type: icumsg.TokenTypeOptionTerm},
		{Str: "one{#message}", Type: icumsg.TokenTypeOptionOne},
		{Str: "#message", Type: icumsg.TokenTypeLiteral},
		{Str: "one{#message}", Type: icumsg.TokenTypeOptionTerm},
		{
			Str:  "{var,plural,other{#messages}one{#message}}",
			Type: icumsg.TokenTypeComplexArgTerm,
		},
	}...)

	// Select ordinal
	f(t, "{_n,selectordinal,one{#st}two{#nd}few{#rd}other{#th}}", []Token{
		{
			Str:  "{_n,selectordinal,one{#st}two{#nd}few{#rd}other{#th}}",
			Type: icumsg.TokenTypeSelectOrdinal,
		},
		{Str: "_n", Type: icumsg.TokenTypeArgName},

		{Str: "one{#st}", Type: icumsg.TokenTypeOptionOne},
		{Str: "#st", Type: icumsg.TokenTypeLiteral},
		{Str: "one{#st}", Type: icumsg.TokenTypeOptionTerm},

		{Str: "two{#nd}", Type: icumsg.TokenTypeOptionTwo},
		{Str: "#nd", Type: icumsg.TokenTypeLiteral},
		{Str: "two{#nd}", Type: icumsg.TokenTypeOptionTerm},

		{Str: "few{#rd}", Type: icumsg.TokenTypeOptionFew},
		{Str: "#rd", Type: icumsg.TokenTypeLiteral},
		{Str: "few{#rd}", Type: icumsg.TokenTypeOptionTerm},

		{Str: "other{#th}", Type: icumsg.TokenTypeOptionOther},
		{Str: "#th", Type: icumsg.TokenTypeLiteral},
		{Str: "other{#th}", Type: icumsg.TokenTypeOptionTerm},

		{
			Str:  "{_n,selectordinal,one{#st}two{#nd}few{#rd}other{#th}}",
			Type: icumsg.TokenTypeComplexArgTerm,
		},
	}...)

	{ // Select offset
		full := `{x,plural,offset:3,other{o}}`
		f(t, full, []Token{
			{Str: full, Type: icumsg.TokenTypePlural},
			{Str: "x", Type: icumsg.TokenTypeArgName},
			{Str: "3", Type: icumsg.TokenTypePluralOffset},
			{Str: "other{o}", Type: icumsg.TokenTypeOptionOther},
			{Str: "o", Type: icumsg.TokenTypeLiteral},
			{Str: "other{o}", Type: icumsg.TokenTypeOptionTerm},
			{Str: full, Type: icumsg.TokenTypeComplexArgTerm},
		}...)
	}

	{ // Select offset no comma
		full := `{ x , plural , offset : 3 other{o}}`
		f(t, full, []Token{
			{Str: full, Type: icumsg.TokenTypePlural},
			{Str: "x", Type: icumsg.TokenTypeArgName},
			{Str: "3", Type: icumsg.TokenTypePluralOffset},
			{Str: "other{o}", Type: icumsg.TokenTypeOptionOther},
			{Str: "o", Type: icumsg.TokenTypeLiteral},
			{Str: "other{o}", Type: icumsg.TokenTypeOptionTerm},
			{Str: full, Type: icumsg.TokenTypeComplexArgTerm},
		}...)
	}

	// Select
	f(t, "{x,select,foo{Foo}bar{Bar}other{Other}}", []Token{
		{
			Str:  "{x,select,foo{Foo}bar{Bar}other{Other}}",
			Type: icumsg.TokenTypeSelect,
		},
		{Str: "x", Type: icumsg.TokenTypeArgName},

		{Str: "foo{Foo}", Type: icumsg.TokenTypeOption},
		{Str: "foo", Type: icumsg.TokenTypeOptionName},
		{Str: "Foo", Type: icumsg.TokenTypeLiteral},
		{Str: "foo{Foo}", Type: icumsg.TokenTypeOptionTerm},

		{Str: "bar{Bar}", Type: icumsg.TokenTypeOption},
		{Str: "bar", Type: icumsg.TokenTypeOptionName},
		{Str: "Bar", Type: icumsg.TokenTypeLiteral},
		{Str: "bar{Bar}", Type: icumsg.TokenTypeOptionTerm},

		{Str: "other{Other}", Type: icumsg.TokenTypeOptionOther},
		{Str: "Other", Type: icumsg.TokenTypeLiteral},
		{Str: "other{Other}", Type: icumsg.TokenTypeOptionTerm},

		{
			Str:  "{x,select,foo{Foo}bar{Bar}other{Other}}",
			Type: icumsg.TokenTypeComplexArgTerm,
		},
	}...)

	{ // Nested choices.
		// Nested choices.
		// Male
		maleEq0 := `=0 {У нього немає повідомлень.}`
		maleOne := `one {У нього одне повідомлення.}`
		maleOther := `other {У нього # повідомлень.}`
		maleMessages := fmt.Sprintf(`{ numMessages , plural , %s %s %s}`,
			maleEq0, maleOne, maleOther)
		optionMale := `male {` + maleMessages + `}`

		// Female
		femaleEq0 := `=0 {У неї немає повідомлень.}`
		femaleOne := `one {У неї одне повідомлення.}`
		femaleOther := `other {У неї # повідомлень.}`
		femaleMessages := fmt.Sprintf(`{ numMessages , plural , %s %s %s}`,
			femaleEq0, femaleOne, femaleOther)
		optionFemale := `female {` + femaleMessages + `}`

		// Other
		otherEq0 := `=0 {У них немає повідомлень.}`
		otherOne := `one {У них одне повідомлення.}`
		otherOther := `other {У них # повідомлень.}`
		otherMessages := fmt.Sprintf(`{ numMessages , plural , %s %s %s}`,
			otherEq0, otherOne, otherOther)
		optionOther := `other {` + otherMessages + `}`

		full := fmt.Sprintf(`{ gender , select , %s %s %s}`,
			optionMale, optionFemale, optionOther)

		f(t, full, []Token{
			{Str: full, Type: icumsg.TokenTypeSelect},
			{Str: "gender", Type: icumsg.TokenTypeArgName},
			// { gender=male
			{Str: optionMale, Type: icumsg.TokenTypeOption},
			{Str: "male", Type: icumsg.TokenTypeOptionName},
			{Str: maleMessages, Type: icumsg.TokenTypePlural},
			{Str: "numMessages", Type: icumsg.TokenTypeArgName},
			// gender=male; numMessages=0
			{Str: maleEq0, Type: icumsg.TokenTypeOptionNumber},
			{Str: "=0", Type: icumsg.TokenTypeOptionName},
			{Str: "У нього немає повідомлень.", Type: icumsg.TokenTypeLiteral},
			{Str: maleEq0, Type: icumsg.TokenTypeOptionTerm},
			// gender=male; numMessages=one
			{Str: maleOne, Type: icumsg.TokenTypeOptionOne},
			{Str: "У нього одне повідомлення.", Type: icumsg.TokenTypeLiteral},
			{Str: maleOne, Type: icumsg.TokenTypeOptionTerm},
			// gender=male; numMessages=other
			{Str: maleOther, Type: icumsg.TokenTypeOptionOther},
			{Str: "У нього # повідомлень.", Type: icumsg.TokenTypeLiteral},
			{Str: maleOther, Type: icumsg.TokenTypeOptionTerm},
			{Str: maleMessages, Type: icumsg.TokenTypeComplexArgTerm},
			// }
			{Str: optionMale, Type: icumsg.TokenTypeOptionTerm},
			// { gender=female
			{Str: optionFemale, Type: icumsg.TokenTypeOption},
			{Str: "female", Type: icumsg.TokenTypeOptionName},
			{Str: femaleMessages, Type: icumsg.TokenTypePlural},
			{Str: "numMessages", Type: icumsg.TokenTypeArgName},
			// gender=female; numMessages=0
			{Str: femaleEq0, Type: icumsg.TokenTypeOptionNumber},
			{Str: "=0", Type: icumsg.TokenTypeOptionName},
			{Str: "У неї немає повідомлень.", Type: icumsg.TokenTypeLiteral},
			{Str: femaleEq0, Type: icumsg.TokenTypeOptionTerm},
			// gender=female; numMessages=one
			{Str: femaleOne, Type: icumsg.TokenTypeOptionOne},
			{Str: "У неї одне повідомлення.", Type: icumsg.TokenTypeLiteral},
			{Str: femaleOne, Type: icumsg.TokenTypeOptionTerm},
			// gender=female; numMessages=other
			{Str: femaleOther, Type: icumsg.TokenTypeOptionOther},
			{Str: "У неї # повідомлень.", Type: icumsg.TokenTypeLiteral},
			{Str: femaleOther, Type: icumsg.TokenTypeOptionTerm},
			{Str: femaleMessages, Type: icumsg.TokenTypeComplexArgTerm},
			// }
			{Str: optionFemale, Type: icumsg.TokenTypeOptionTerm},
			// { gender=other
			{Str: optionOther, Type: icumsg.TokenTypeOptionOther},
			{Str: otherMessages, Type: icumsg.TokenTypePlural},
			{Str: "numMessages", Type: icumsg.TokenTypeArgName},
			// gender=other; numMessages=0
			{Str: otherEq0, Type: icumsg.TokenTypeOptionNumber},
			{Str: "=0", Type: icumsg.TokenTypeOptionName},
			{Str: "У них немає повідомлень.", Type: icumsg.TokenTypeLiteral},
			{Str: otherEq0, Type: icumsg.TokenTypeOptionTerm},
			// gender=other; numMessages=one
			{Str: otherOne, Type: icumsg.TokenTypeOptionOne},
			{Str: "У них одне повідомлення.", Type: icumsg.TokenTypeLiteral},
			{Str: otherOne, Type: icumsg.TokenTypeOptionTerm},
			// gender=other; numMessages=other
			{Str: otherOther, Type: icumsg.TokenTypeOptionOther},
			{Str: "У них # повідомлень.", Type: icumsg.TokenTypeLiteral},
			{Str: otherOther, Type: icumsg.TokenTypeOptionTerm},
			{Str: otherMessages, Type: icumsg.TokenTypeComplexArgTerm},
			// }
			{Str: optionOther, Type: icumsg.TokenTypeOptionTerm},

			{Str: full, Type: icumsg.TokenTypeComplexArgTerm},
		}...)
	}
}

type TestError struct {
	Input          string
	ExpectErrIndex int
	ExpectErr      error
}

var TestsErrors = []TestError{
	{"{", 1, icumsg.ErrUnexpectedEOF},
	{"{x", 2, icumsg.ErrUnexpectedEOF},
	{"{x ", 3, icumsg.ErrUnexpectedEOF},
	{"{x,", 3, icumsg.ErrUnexpectedEOF},
	{"{x, ", 4, icumsg.ErrUnexpectedEOF},
	{"{x, number", 10, icumsg.ErrUnexpectedEOF},
	{"{x, number ", 11, icumsg.ErrUnexpectedEOF},
	{"{x, number ,", 12, icumsg.ErrUnexpectedEOF},
	{"{x, number , ", 13, icumsg.ErrUnexpectedEOF},
	{"{x, number , integer", 20, icumsg.ErrUnexpectedEOF},
	{"{x, number , integer ", 21, icumsg.ErrUnexpectedEOF},
	{"{x,select, other", 16, icumsg.ErrUnexpectedEOF},
	{"{x,select, other ", 17, icumsg.ErrUnexpectedEOF},
	{"{x,select, other {", 18, icumsg.ErrUnexpectedEOF},
	{"{x,select, other { ", 19, icumsg.ErrUnexpectedEOF},
	{"{x,select, other { asd", 22, icumsg.ErrUnexpectedEOF},
	{"{x,select, other { asd ", 23, icumsg.ErrUnexpectedEOF},
	{"{x,select, other { asd }", 24, icumsg.ErrUnexpectedEOF},
	{"{x,select, other { asd } ", 25, icumsg.ErrUnexpectedEOF},
	{"{x,select, other { asd } =", 26, icumsg.ErrUnexpectedEOF},
	{"{x,selectordinal, other", 23, icumsg.ErrUnexpectedEOF},
	{"{x,selectordinal, other ", 24, icumsg.ErrUnexpectedEOF},
	{"{x,selectordinal, other {", 25, icumsg.ErrUnexpectedEOF},
	{"{x,selectordinal, other { ", 26, icumsg.ErrUnexpectedEOF},
	{"{x,selectordinal, other { asd", 29, icumsg.ErrUnexpectedEOF},
	{"{x,selectordinal, other { asd ", 30, icumsg.ErrUnexpectedEOF},
	{"{x,selectordinal, other { asd }", 31, icumsg.ErrUnexpectedEOF},
	{"{x,selectordinal, other { asd } ", 32, icumsg.ErrUnexpectedEOF},
	{"{x,selectordinal, other { asd } =", 33, icumsg.ErrUnexpectedEOF},
	{"{x,plural,offset", 16, icumsg.ErrUnexpectedEOF},
	{"{x,plural,offset ", 17, icumsg.ErrUnexpectedEOF},
	{"{x,plural,offset:", 17, icumsg.ErrUnexpectedEOF},
	{"{x,plural,offset: ", 18, icumsg.ErrUnexpectedEOF},
	{"{x,plural,offset:1", 18, icumsg.ErrUnexpectedEOF},
	{"{x,plural,offset:1,", 19, icumsg.ErrUnexpectedEOF},
	{"{x,plural,offset:1, ", 20, icumsg.ErrUnexpectedEOF},
	{"{x,plural", 9, icumsg.ErrUnexpectedEOF},
	{"{x,plural ", 10, icumsg.ErrUnexpectedEOF},
	{"{x,plural,", 10, icumsg.ErrUnexpectedEOF},
	{"{x,plural, ", 11, icumsg.ErrUnexpectedEOF},
	{"{x,plural, other", 16, icumsg.ErrUnexpectedEOF},
	{"{x,plural, other ", 17, icumsg.ErrUnexpectedEOF},
	{"{x,plural, other {", 18, icumsg.ErrUnexpectedEOF},
	{"{x,plural, other { ", 19, icumsg.ErrUnexpectedEOF},
	{"{x,plural, other { asd", 22, icumsg.ErrUnexpectedEOF},
	{"{x,plural, other { asd ", 23, icumsg.ErrUnexpectedEOF},
	{"{x,plural, other { asd }", 24, icumsg.ErrUnexpectedEOF},
	{"{x,plural, other { asd } ", 25, icumsg.ErrUnexpectedEOF},
	{"{x,plural, other { asd } =", 26, icumsg.ErrUnexpectedEOF},
	// Invalid option
	{"{x,plural, other { asd } =01 {x} }", 25, icumsg.ErrInvalidOption},
	{"{x,plural, other { asd } =a {x} }", 26, icumsg.ErrInvalidOption},
	{"{x,plural, other { asd } unknown {x} }", 25, icumsg.ErrInvalidOption},
	// Unexpected token
	{"{}", 1, icumsg.ErrUnexpectedToken},
	{"{n x}", 3, icumsg.ErrUnexpectedToken},
	{"{n {}}", 3, icumsg.ErrUnexpectedToken},
	{"{x, unknown}", 4, icumsg.ErrUnexpectedToken},
	{"{x: plural, other{x}}", 2, icumsg.ErrUnexpectedToken},
	{"{x| plural, other{x}}", 2, icumsg.ErrUnexpectedToken},
	{"{x? plural, other{x}}", 2, icumsg.ErrUnexpectedToken},
	{"{x__? plural, other{x}}", 4, icumsg.ErrUnexpectedToken},
	{"{x_, unknown, other{x}}", 5, icumsg.ErrUnexpectedToken},
	{"{x,plural,other{{}}}", 17, icumsg.ErrUnexpectedToken},
	// Expected colon.
	{"{x,plural,offset,", 16, icumsg.ErrExpectedColon},
	{"{x,plural,offset ,", 17, icumsg.ErrExpectedColon},
	// Expected comma.
	{"{x_, plural: other{x}}", 11, icumsg.ErrExpectedComma},
	{"{x_, plural | other{x}}", 12, icumsg.ErrExpectedComma},
	{"{x, select: other{x}}", 10, icumsg.ErrExpectedComma},
	{"{x, selectordinal: other{x}}", 17, icumsg.ErrExpectedComma},
	// Expected opening bracket.
	{"{x_, plural, other, one{x} }", 18, icumsg.ErrExpectBracketOpen},
	{"{x_, plural, other , one{x} }", 19, icumsg.ErrExpectBracketOpen},
	{"{x_, plural, other , one{x} }", 19, icumsg.ErrExpectBracketOpen},
	{"{x,plural, other { asd } =1a {x} }", 27, icumsg.ErrExpectBracketOpen},
	{"{x,plural, other { asd } =1? {x} }", 27, icumsg.ErrExpectBracketOpen},
	{"{x_, selectordinal, other, one{x} }", 25, icumsg.ErrExpectBracketOpen},
	{"{x_, select, other, one{x} }", 18, icumsg.ErrExpectBracketOpen},
	// Expected closing bracket.
	{"{n, number, integer, foobar}", 19, icumsg.ErrExpectBracketClose},
	{"{n, number foobar}", 11, icumsg.ErrExpectBracketClose},
	// Empty option
	{"{x,plural, other { } }", 17, icumsg.ErrEmptyOption},
	{"{x,plural, one {x} other {} }", 25, icumsg.ErrEmptyOption},
	{"{x,selectordinal, one {x} other {} }", 32, icumsg.ErrEmptyOption},
	{"{x,select, one {x} other {} }", 25, icumsg.ErrEmptyOption},
	{"{x,select, one {x} other {{y,select,other{}} } }", 41, icumsg.ErrEmptyOption},
	{"{x,plural, one {x} other {{y,select,other{}} } }", 41, icumsg.ErrEmptyOption},
	{
		"{x,selectordinal, one {x} other {{y,select,other{}} } }",
		48, icumsg.ErrEmptyOption,
	},
	// Duplicate option in plural
	{"{n, plural, other{a} other{c}}", 21, icumsg.ErrDuplicateOption},
	{"{n, plural, other{a} one{b} other{c}}", 28, icumsg.ErrDuplicateOption},
	{"{n, plural, other{a} zero{b} zero{c}}", 29, icumsg.ErrDuplicateOption},
	{"{n, plural, other{a} one{b} one{c}}", 28, icumsg.ErrDuplicateOption},
	{"{n, plural, other{a} two{b} two{c}}", 28, icumsg.ErrDuplicateOption},
	{"{n, plural, other{a} few{b} few{c}}", 28, icumsg.ErrDuplicateOption},
	{"{n, plural, other{a} many{b} many{c}}", 29, icumsg.ErrDuplicateOption},
	{"{n, plural, other{a} =0{b} =0{c}}", 27, icumsg.ErrDuplicateOption},
	{"{n, plural, other{a} =0{b} =1{c} =0{d}}", 33, icumsg.ErrDuplicateOption},
	// Duplicate option in selectordinal
	{"{n, selectordinal, other{a} other{c}}", 28, icumsg.ErrDuplicateOption},
	{"{n, selectordinal, other{a} one{b} other{c}}", 35, icumsg.ErrDuplicateOption},
	{"{n, selectordinal, other{a} zero{b} zero{c}}", 36, icumsg.ErrDuplicateOption},
	{"{n, selectordinal, other{a} one{b} one{c}}", 35, icumsg.ErrDuplicateOption},
	{"{n, selectordinal, other{a} two{b} two{c}}", 35, icumsg.ErrDuplicateOption},
	{"{n, selectordinal, other{a} few{b} few{c}}", 35, icumsg.ErrDuplicateOption},
	{"{n, selectordinal, other{a} many{b} many{c}}", 36, icumsg.ErrDuplicateOption},
	{"{n, selectordinal, other{a} =0{b} =0{c}}", 34, icumsg.ErrDuplicateOption},
	{"{n, selectordinal, other{a} =0{b} =1{c} =0{d}}", 40, icumsg.ErrDuplicateOption},
	// Duplicate option in select
	{"{n, select, other{a} other{c}}", 21, icumsg.ErrDuplicateOption},
	{"{n, select, other{a} one{b} other{c}}", 28, icumsg.ErrDuplicateOption},
	{"{n, select, other{a} zero{b} zero{c}}", 29, icumsg.ErrDuplicateOption},
	{"{n, select, other{a} one{b} one{c}}", 28, icumsg.ErrDuplicateOption},
	{"{n, select, other{a} two{b} two{c}}", 28, icumsg.ErrDuplicateOption},
	{"{n, select, other{a} few{b} few{c}}", 28, icumsg.ErrDuplicateOption},
	{"{n, select, other{a} many{b} many{c}}", 29, icumsg.ErrDuplicateOption},
	// Missing option 'other'
	{"before {x, select, one{a}}", 7, icumsg.ErrMissingOptionOther},
	{"before {x, select, one{a} two{b}}", 7, icumsg.ErrMissingOptionOther},
	{"before {x, select, x{a} y{b}}", 7, icumsg.ErrMissingOptionOther},
	{"before {n, plural, one{a}}", 7, icumsg.ErrMissingOptionOther},
	{"before {n, plural, one{a} two{b}}", 7, icumsg.ErrMissingOptionOther},
	{"before {n, selectordinal, one{a}}", 7, icumsg.ErrMissingOptionOther},
	{"before {n, selectordinal, one{a} two{b}}", 7, icumsg.ErrMissingOptionOther},
}

func TestTokenizeErr(t *testing.T) {
	t.Parallel()

	var tokenizer icumsg.Tokenizer
	var buffer []icumsg.Token

	for _, tt := range TestsErrors {
		t.Run("", func(t *testing.T) {
			buffer = buffer[:0]
			_, err := tokenizer.Tokenize(buffer, tt.Input)
			t.Logf("input: %q", tt.Input)
			requireErrIs(t, tt.ExpectErr, err)
			requireEqual(t, tt.ExpectErrIndex, tokenizer.Pos())
		})
	}
}

func Fuzz(f *testing.F) {
	for _, tt := range TestsErrors {
		f.Add(tt.Input)
	}
	f.Add("Very small")
	f.Add("Good morning {userName}, how are you?")
	f.Add(ReadFile[string](f, "testdata/lorem_ipsum.txt"))
	f.Add(ReadFile[string](f, "testdata/lorem_ipsum_args.icu.txt"))
	f.Add(ReadFile[string](f, "testdata/nested.icu.txt"))

	var tokenizer icumsg.Tokenizer
	buffer := make([]icumsg.Token, 0, 64)

	f.Fuzz(func(t *testing.T, input string) {
		buffer = buffer[:0]
		_, _ = tokenizer.Tokenize(buffer, input)
	})
}

func BenchmarkTokenize(b *testing.B) {
	var tokenizer icumsg.Tokenizer
	buffer := make([]icumsg.Token, 0, 64)

	for _, input := range [...]string{
		"Very small",
		"Good morning {userName}, how are you?",
		ReadFile[string](b, "testdata/lorem_ipsum.txt"),
		ReadFile[string](b, "testdata/lorem_ipsum_args.icu.txt"),
		ReadFile[string](b, "testdata/nested.icu.txt"),
	} {
		b.Run("", func(b *testing.B) {
			for b.Loop() {
				var err error
				buffer = buffer[:0] // Reset buffer.
				buffer, err = tokenizer.Tokenize(buffer, input)
				if err != nil {
					panic(err)
				}
			}
		})
	}
}
