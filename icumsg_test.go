package icumsg_test

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/romshark/icumsg"
	"golang.org/x/text/language"
)

func requireEqual[T comparable](tb testing.TB, expect, actual T, msg ...any) {
	tb.Helper()
	if expect != actual {
		m := ""
		if msg != nil {
			m = "\n" + fmt.Sprintf(msg[0].(string), msg[1:]...)
		}
		tb.Fatalf("\nexpected: %#v;\nreceived: %#v%s", expect, actual, m)
	}
}

func requireDeepEqual[T any](tb testing.TB, expect, actual T, msg ...any) {
	tb.Helper()
	if !reflect.DeepEqual(expect, actual) {
		m := ""
		if msg != nil {
			m = "\n" + fmt.Sprintf(msg[0].(string), msg[1:]...)
		}
		tb.Fatalf("\nexpected: %#v;\nreceived: %#v%s", expect, actual, m)
	}
}

func requireErrIs(t *testing.T, expect, actual error, msg ...any) {
	t.Helper()
	if !errors.Is(actual, expect) {
		m := ""
		if msg != nil {
			m = fmt.Sprintf(msg[0].(string), msg[1:]...)
		}
		t.Fatalf("\nexpected: %#v;\nreceived: %#v%s", expect, actual, m)
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
func ToTestTokens(input string, buffer, toks []icumsg.Token) []Token {
	if len(toks) == 0 {
		return nil
	}
	tokens := make([]Token, len(toks))
	for i, tok := range toks {
		tokens[i] = Token{
			Str:  tok.String(input, buffer),
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
	f := func(t *testing.T, locale language.Tag, input string, expect ...Token) {
		t.Helper()
		buffer = buffer[:0]
		buffer, err := tokenizer.Tokenize(locale, buffer, input)
		requireNoErr(t, err)
		actual := ToTestTokens(input, buffer, buffer)
		compareTokens(t, expect, actual)
	}

	f(t, language.English, "")
	f(t, language.English, "foo", []Token{
		{Str: "foo", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, language.English, "foo bar\n\tbazz", []Token{
		{Str: "foo bar\n\tbazz", Type: icumsg.TokenTypeLiteral},
	}...)

	// Escaping
	f(t, language.English, "''", []Token{
		{Str: "''", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, language.English, "'{}' '{}'", []Token{
		{Str: "'{}' '{}'", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, language.English, "before '{x '' y}' after", []Token{
		{Str: "before '{x '' y}' after", Type: icumsg.TokenTypeLiteral},
	}...)

	// Argument
	f(t, language.English, "{_}", []Token{
		{Str: "{_}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "_", Type: icumsg.TokenTypeArgName},
	}...)
	f(t, language.English, "{1}", []Token{
		{Str: "{1}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "1", Type: icumsg.TokenTypeArgName},
	}...)
	f(t, language.English, "{arg}", []Token{
		{Str: "{arg}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
	}...)
	f(t, language.English, "{аргумент}", []Token{
		{Str: "{аргумент}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "аргумент", Type: icumsg.TokenTypeArgName},
	}...)
	f(t, language.English, "{ arg }", []Token{
		{Str: "{ arg }", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
	}...)
	f(t, language.English, "{\n arg \n}", []Token{
		{Str: "{\n arg \n}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
	}...)

	// Argument type
	f(t, language.English, "Before {arg, number} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, number}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, language.English, "Before {arg, date} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, date}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "date", Type: icumsg.TokenTypeArgTypeDate},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, language.English, "Before {arg, time} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, time}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "time", Type: icumsg.TokenTypeArgTypeTime},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, language.English, "Before {arg, spellout} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, spellout}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "spellout", Type: icumsg.TokenTypeArgTypeSpellout},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, language.English, "Before {arg, ordinal} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, ordinal}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "ordinal", Type: icumsg.TokenTypeArgTypeOrdinal},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, language.English, "Before {arg, duration} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, duration}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "duration", Type: icumsg.TokenTypeArgTypeDuration},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)

	// Argument style
	f(t, language.English, "Before {arg, number, short} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, number, short}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "short", Type: icumsg.TokenTypeArgStyleShort},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, language.English, "Before {arg, number, medium} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, number, medium}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "medium", Type: icumsg.TokenTypeArgStyleMedium},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, language.English, "Before {arg, number, long} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, number, long}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "long", Type: icumsg.TokenTypeArgStyleLong},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, language.English, "Before {arg, number, full} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, number, full}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "full", Type: icumsg.TokenTypeArgStyleFull},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, language.English, "Before {arg, number, integer} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, number, integer}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "integer", Type: icumsg.TokenTypeArgStyleInteger},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, language.English, "Before {arg, number, ::x} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, number, ::x}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "::x", Type: icumsg.TokenTypeArgStyleSkeleton},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, language.English, "Before {arg, number, ::currency/auto} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, number, ::currency/auto}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "::currency/auto", Type: icumsg.TokenTypeArgStyleSkeleton},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, language.English, "Before {arg, number, ::sign-always} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, number, ::sign-always}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "::sign-always", Type: icumsg.TokenTypeArgStyleSkeleton},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)

	f(t, language.English, "Before {arg, number, percent} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, number, percent}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "percent", Type: icumsg.TokenTypeArgStylePercent},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, language.English, "Before {arg, number, customAnything} after", []Token{
		{Str: "Before ", Type: icumsg.TokenTypeLiteral},
		{Str: "{arg, number, customAnything}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "arg", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "customAnything", Type: icumsg.TokenTypeArgStyleCustom},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)

	// Plural
	f(t, language.English, "{var,plural,other{#messages}one{#message}}", []Token{
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
	f(t, language.English, "{_n,selectordinal,one{#st}two{#nd}few{#rd}other{#th}}", []Token{
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
		f(t, language.English, full, []Token{
			{Str: full, Type: icumsg.TokenTypePlural},
			{Str: "x", Type: icumsg.TokenTypeArgName},
			{Str: "3", Type: icumsg.TokenTypePluralOffset},
			{Str: "other{o}", Type: icumsg.TokenTypeOptionOther},
			{Str: "o", Type: icumsg.TokenTypeLiteral},
			{Str: "other{o}", Type: icumsg.TokenTypeOptionTerm},
			{Str: full, Type: icumsg.TokenTypeComplexArgTerm},
		}...)
	}

	{ // Select offset zero
		full := `{x,plural,offset:0,other{o}}`
		f(t, language.English, full, []Token{
			{Str: full, Type: icumsg.TokenTypePlural},
			{Str: "x", Type: icumsg.TokenTypeArgName},
			{Str: "0", Type: icumsg.TokenTypePluralOffset},
			{Str: "other{o}", Type: icumsg.TokenTypeOptionOther},
			{Str: "o", Type: icumsg.TokenTypeLiteral},
			{Str: "other{o}", Type: icumsg.TokenTypeOptionTerm},
			{Str: full, Type: icumsg.TokenTypeComplexArgTerm},
		}...)
	}

	{ // Select offset no comma
		full := `{ x , plural , offset : 3 other{o}}`
		f(t, language.English, full, []Token{
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
	f(t, language.English, "{x,select,foo{Foo}bar{Bar}other{Other}}", []Token{
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

		f(t, language.English, full, []Token{
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

type TestErrorLocale struct {
	Input          string
	Locale         language.Tag
	ExpectErrIndex int
	ExpectErr      error
}

var TestsErrorsLocale = []TestErrorLocale{
	{
		"{x,plural, other{yes} few{no}}", language.English,
		22, icumsg.ErrUnsupportedPluralForm,
	},
	{
		"{x,plural, other{yes} few{no}}", language.AmericanEnglish,
		22, icumsg.ErrUnsupportedPluralForm,
	},
	{
		"{x,plural, other{yes} zero{no}}", language.Ukrainian,
		22, icumsg.ErrUnsupportedPluralForm,
	},
	{
		"{x,plural, one{yes} two{no} other{yes}}", language.German,
		20, icumsg.ErrUnsupportedPluralForm,
	},
	{
		"{x,selectordinal, other{yes} one{no}}", language.German,
		29, icumsg.ErrUnsupportedPluralForm,
	},
	{
		"{x,selectordinal, other{yes} zero{no}}", language.German,
		29, icumsg.ErrUnsupportedPluralForm,
	},
	{
		"{x,selectordinal, other{yes} two{no}}", language.German,
		29, icumsg.ErrUnsupportedPluralForm,
	},
	{
		"{x,selectordinal, other{yes} many{no}}", language.German,
		29, icumsg.ErrUnsupportedPluralForm,
	},
	{
		"{x,selectordinal, other{yes} few{no}}", language.German,
		29, icumsg.ErrUnsupportedPluralForm,
	},
	{
		"{x,selectordinal, other{yes} zero{no}}", language.Ukrainian,
		29, icumsg.ErrUnsupportedPluralForm,
	},
}

func TestTokenizeErrLocale(t *testing.T) {
	t.Parallel()

	var tokenizer icumsg.Tokenizer
	var buffer []icumsg.Token

	for _, tt := range TestsErrorsLocale {
		t.Run("", func(t *testing.T) {
			buffer = buffer[:0]
			_, err := tokenizer.Tokenize(tt.Locale, buffer, tt.Input)
			t.Logf("input: %q", tt.Input)
			requireErrIs(t, tt.ExpectErr, err)
			requireEqual(t, tt.ExpectErrIndex, tokenizer.Pos())
		})
	}
}

func TestFUCK(t *testing.T) {
	t.Parallel()

	var tokenizer icumsg.Tokenizer
	_, err := tokenizer.Tokenize(language.English, nil, "{x, number, ::")
	t.Log(err)
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
	{"{x, number, ::", 14, icumsg.ErrUnexpectedEOF},
	{"{x,select, other", 16, icumsg.ErrUnexpectedEOF},
	{"{x,select, other ", 17, icumsg.ErrUnexpectedEOF},
	{"{x,select, other {", 18, icumsg.ErrUnexpectedEOF},
	{"{x,select, other { ", 19, icumsg.ErrUnexpectedEOF},
	{"{x,select, other { asd", 22, icumsg.ErrUnexpectedEOF},
	{"{x,select, other { asd ", 23, icumsg.ErrUnexpectedEOF},
	{"{x,select, other { asd }", 24, icumsg.ErrUnexpectedEOF},
	{"{x,select, other { asd } ", 25, icumsg.ErrUnexpectedEOF},
	{"{x,select", 9, icumsg.ErrUnexpectedEOF},
	{"{x,plural", 9, icumsg.ErrUnexpectedEOF},
	{"{x,selectordinal", 16, icumsg.ErrUnexpectedEOF},
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
	{"{x,select, other { asd } {x} }", 25, icumsg.ErrInvalidOption},
	{"{x,plural, other { asd } =01 {x} }", 25, icumsg.ErrInvalidOption},
	{"{x,plural, other { asd } =a {x} }", 26, icumsg.ErrInvalidOption},
	{"{x,plural, other { asd } ?{x} }", 25, icumsg.ErrInvalidOption},
	{"{x,plural, other { asd } unknown {x} }", 25, icumsg.ErrInvalidOption},
	{"{x,plural, offset:0x1 other{foo}}", 19, icumsg.ErrInvalidOption},
	{"{x,select, other { asd } =1 {x} }", 25, icumsg.ErrInvalidOption},
	// Unclosed quote
	{"prefix 'unclosed quote", 7, icumsg.ErrUnclosedQuote},
	{"prefix '' 'unclosed quote", 10, icumsg.ErrUnclosedQuote},
	{"prefix '{}' 'unclosed quote", 12, icumsg.ErrUnclosedQuote},
	{"{x,plural, other { '{}' ' }}", 24, icumsg.ErrUnclosedQuote},
	{"'", 0, icumsg.ErrUnclosedQuote},
	// Unexpected token.
	{"}", 0, icumsg.ErrUnexpectedToken},
	{"prefix }", 7, icumsg.ErrUnexpectedToken},
	{"prefix } suffix", 7, icumsg.ErrUnexpectedToken},
	{"{}", 1, icumsg.ErrUnexpectedToken},
	{"{'}", 1, icumsg.ErrUnexpectedToken},
	{"{?}", 1, icumsg.ErrUnexpectedToken},
	{"{n x}", 3, icumsg.ErrUnexpectedToken},
	{"{n {}}", 3, icumsg.ErrUnexpectedToken},
	{"{x, unknown}", 4, icumsg.ErrUnexpectedToken},
	{"{x: plural, other{x}}", 2, icumsg.ErrUnexpectedToken},
	{"{x| plural, other{x}}", 2, icumsg.ErrUnexpectedToken},
	{"{x? plural, other{x}}", 2, icumsg.ErrUnexpectedToken},
	{"{x__? plural, other{x}}", 4, icumsg.ErrUnexpectedToken},
	{"{x_, unknown, other{x}}", 5, icumsg.ErrUnexpectedToken},
	{"{x,plural,other{{}}}", 17, icumsg.ErrUnexpectedToken},
	{"{n, plural, other{x} }}", 22, icumsg.ErrUnexpectedToken},
	{"{x, number, ::}", 14, icumsg.ErrUnexpectedToken},
	// Expected colon.
	{"{x,plural,offset,", 16, icumsg.ErrExpectedColon},
	{"{x,plural,offset ,", 17, icumsg.ErrExpectedColon},
	// Expected comma.
	{"{x_, plural: other{x}}", 11, icumsg.ErrExpectedComma},
	{"{x_, plural | other{x}}", 12, icumsg.ErrExpectedComma},
	{"{x, select: other{x}}", 10, icumsg.ErrExpectedComma},
	{"{x, selectordinal: other{x}}", 17, icumsg.ErrExpectedComma},
	// Invalid offset.
	{"{x,plural,offset:a", 17, icumsg.ErrInvalidOffset},
	{"{x,plural,offset:?, other{foo}}", 17, icumsg.ErrInvalidOffset},
	{"{x,plural,offset:-1, other{foo}}", 17, icumsg.ErrInvalidOffset},
	{"{x,plural,offset: , other{foo}}", 18, icumsg.ErrInvalidOffset},
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
	// Empty option.
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
	// Duplicate option in plural.
	{"{n, plural, other{a} other{c}}", 21, icumsg.ErrDuplicateOption},
	{"{n, plural, other{a} one{b} other{c}}", 28, icumsg.ErrDuplicateOption},
	{"{n, plural, other{a} zero{b} zero{c}}", 29, icumsg.ErrDuplicateOption},
	{"{n, plural, other{a} one{b} one{c}}", 28, icumsg.ErrDuplicateOption},
	{"{n, plural, other{a} two{b} two{c}}", 28, icumsg.ErrDuplicateOption},
	{"{n, plural, other{a} few{b} few{c}}", 28, icumsg.ErrDuplicateOption},
	{"{n, plural, other{a} many{b} many{c}}", 29, icumsg.ErrDuplicateOption},
	{"{n, plural, other{a} =0{b} =0{c}}", 27, icumsg.ErrDuplicateOption},
	{"{n, plural, other{a} =0{b} =1{c} =0{d}}", 33, icumsg.ErrDuplicateOption},
	// Duplicate option in selectordinal.
	{"{n, selectordinal, other{a} other{c}}", 28, icumsg.ErrDuplicateOption},
	{"{n, selectordinal, other{a} one{b} other{c}}", 35, icumsg.ErrDuplicateOption},
	{"{n, selectordinal, other{a} zero{b} zero{c}}", 36, icumsg.ErrDuplicateOption},
	{"{n, selectordinal, other{a} one{b} one{c}}", 35, icumsg.ErrDuplicateOption},
	{"{n, selectordinal, other{a} two{b} two{c}}", 35, icumsg.ErrDuplicateOption},
	{"{n, selectordinal, other{a} few{b} few{c}}", 35, icumsg.ErrDuplicateOption},
	{"{n, selectordinal, other{a} many{b} many{c}}", 36, icumsg.ErrDuplicateOption},
	{"{n, selectordinal, other{a} =0{b} =0{c}}", 34, icumsg.ErrDuplicateOption},
	{"{n, selectordinal, other{a} =0{b} =1{c} =0{d}}", 40, icumsg.ErrDuplicateOption},
	// Duplicate option in select.
	{"{n, select, other{a} other{c}}", 21, icumsg.ErrDuplicateOption},
	{"{n, select, other{a} one{b} other{c}}", 28, icumsg.ErrDuplicateOption},
	{"{n, select, other{a} zero{b} zero{c}}", 29, icumsg.ErrDuplicateOption},
	{"{n, select, other{a} one{b} one{c}}", 28, icumsg.ErrDuplicateOption},
	{"{n, select, other{a} two{b} two{c}}", 28, icumsg.ErrDuplicateOption},
	{"{n, select, other{a} few{b} few{c}}", 28, icumsg.ErrDuplicateOption},
	{"{n, select, other{a} many{b} many{c}}", 29, icumsg.ErrDuplicateOption},
	// Missing option 'other'.
	{"prefix {x,plural, }", 7, icumsg.ErrMissingOptionOther},
	{"prefix {x,select, }", 7, icumsg.ErrMissingOptionOther},
	{"prefix {x,selectordinal, }", 7, icumsg.ErrMissingOptionOther},
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
		l, err := language.Parse("cy")
		requireNoErr(t, err)
		t.Run("", func(t *testing.T) {
			buffer = buffer[:0]
			_, err := tokenizer.Tokenize(l, buffer, tt.Input)
			t.Logf("input: %q", tt.Input)
			requireErrIs(t, tt.ExpectErr, err)
			requireEqual(t, tt.ExpectErrIndex, tokenizer.Pos())
		})
	}
}

func TestOptions(t *testing.T) {
	var tokenizer icumsg.Tokenizer
	var buffer []icumsg.Token

	fn := func(t *testing.T, input string, index int, expect ...Token) {
		t.Helper()
		buffer = buffer[:0]
		var err error
		buffer, err = tokenizer.Tokenize(language.English, buffer, input)
		requireNoErr(t, err)
		var collected []icumsg.Token
		for i := range icumsg.Options(buffer, index) {
			collected = append(collected, buffer[i])
		}
		actual := ToTestTokens(input, buffer, collected)
		compareTokens(t, expect, actual)
	}

	fn(t, "Not a plural, select or selectordinal", 0)
	fn(t, "Prefix {x, plural, other {a} one {b}}", 1,
		Token{Str: "other {a}", Type: icumsg.TokenTypeOptionOther},
		Token{Str: "one {b}", Type: icumsg.TokenTypeOptionOne})
	fn(t, "Prefix {x,select, other{x}}", 1,
		Token{Str: "other{x}", Type: icumsg.TokenTypeOptionOther})
	fn(t, "Prefix {x,select,other{o}opt1{a}opt2{b}opt3{c}opt4{d}}", 1,
		Token{Str: "other{o}", Type: icumsg.TokenTypeOptionOther},
		Token{Str: "opt1{a}", Type: icumsg.TokenTypeOption},
		Token{Str: "opt2{b}", Type: icumsg.TokenTypeOption},
		Token{Str: "opt3{c}", Type: icumsg.TokenTypeOption},
		Token{Str: "opt4{d}", Type: icumsg.TokenTypeOption})
	fn(t, "Prefix {x,selectordinal,other{o}one{a}few{b}two{c}}", 1,
		Token{Str: "other{o}", Type: icumsg.TokenTypeOptionOther},
		Token{Str: "one{a}", Type: icumsg.TokenTypeOptionOne},
		Token{Str: "few{b}", Type: icumsg.TokenTypeOptionFew},
		Token{Str: "two{c}", Type: icumsg.TokenTypeOptionTwo})

	{
		nested := `Prefix {x,plural,
			other{byGender: {gender, select,
				other  {A_OTHER}
				female {A_FEMALE}
				male   {A_MALE}
			}}
			one{byGender: {gender, select,
				other  {B_OTHER}
				female {B_FEMALE}
				male   {B_MALE}
			}}
		}`
		fn(t, nested, 1, // plural
			Token{Str: `other{byGender: {gender, select,
				other  {A_OTHER}
				female {A_FEMALE}
				male   {A_MALE}
			}}`, Type: icumsg.TokenTypeOptionOther},
			Token{Str: `one{byGender: {gender, select,
				other  {B_OTHER}
				female {B_FEMALE}
				male   {B_MALE}
			}}`, Type: icumsg.TokenTypeOptionOne})
		fn(t, nested, 5, // x=other
			Token{Str: `other  {A_OTHER}`, Type: icumsg.TokenTypeOptionOther},
			Token{Str: `female {A_FEMALE}`, Type: icumsg.TokenTypeOption},
			Token{Str: `male   {A_MALE}`, Type: icumsg.TokenTypeOption})
		fn(t, nested, 22, // x=one
			Token{Str: `other  {B_OTHER}`, Type: icumsg.TokenTypeOptionOther},
			Token{Str: `female {B_FEMALE}`, Type: icumsg.TokenTypeOption},
			Token{Str: `male   {B_MALE}`, Type: icumsg.TokenTypeOption})
	}
}

func TestOptionsBreak(t *testing.T) {
	var tokenizer icumsg.Tokenizer

	input := "Prefix {x,selectordinal,other{o}one{a}few{b}two{c}}"
	buffer, err := tokenizer.Tokenize(language.English, nil, input)
	requireNoErr(t, err)

	itr := 0
	for range icumsg.Options(buffer, 1) {
		itr++
		break
	}
	requireEqual(t, 1, itr)
}

func TestCompleteness(t *testing.T) {
	var tokenizer icumsg.Tokenizer
	var buffer []icumsg.Token

	fn := func(
		t *testing.T,
		locale language.Tag,
		input string,
		options map[string][]string,
		presencePolicy icumsg.OptionsPresencePolicy,
		unknownPolicy icumsg.OptionUnknownPolicy,
		expectTotal int,
		expectIncomplete, expectRejected []string,
	) {
		t.Helper()
		buffer = buffer[:0]
		var err error
		buffer, err = tokenizer.Tokenize(locale, buffer, input)
		requireNoErr(t, err)
		var incomplete, rejected []string
		actualTotal := icumsg.Completeness(input, buffer, locale,
			func(argName string) (
				[]string, icumsg.OptionsPresencePolicy, icumsg.OptionUnknownPolicy,
			) {
				return options[argName], presencePolicy, unknownPolicy
			},
			func(index int) {
				incomplete = append(incomplete, buffer[index].String(input, buffer))
			},
			func(index int) {
				rejected = append(rejected, buffer[index].String(input, buffer))
			})
		requireEqual(t, expectTotal, actualTotal, "total")
		requireDeepEqual(t, expectIncomplete, incomplete, "incomplete")
		requireDeepEqual(t, expectRejected, rejected, "rejected")
	}

	// Expect full completeness.
	fn(t, language.English, "Simple literal text message",
		nil, icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyIgnore,
		0, nil, nil)
	fn(t, language.AmericanEnglish, "Simple literal text message",
		nil, icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyIgnore,
		0, nil, nil)
	fn(t, language.Ukrainian, "Simple literal text message",
		nil, icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyIgnore,
		0, nil, nil)
	fn(t, language.Arabic, "Simple literal text message",
		nil, icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyIgnore,
		0, nil, nil)

	// Expect full completeness for no options provided.
	fn(t, language.English, "{_0, select, other{a}}",
		nil, icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyIgnore,
		1, nil, nil)

	// Expect full completeness for ordinal and plural.
	fn(t, language.English, "{_0, plural, other{# a} one{# b}}",
		nil, icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyIgnore,
		1, nil, nil)
	fn(t, language.English, "{_0, plural, one{# b} other{# a}}",
		nil, icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyIgnore,
		1, nil, nil)
	fn(t, language.English, "{_0, plural, =0{a} =1{b} one{# c} other{# d}}",
		nil, icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyIgnore,
		1, nil, nil)
	fn(t, language.Arabic,
		"{_0, plural, one{# b} other{# a} few{# c} many{# d} zero{# e} two{# f}}",
		nil, icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyIgnore,
		1, nil, nil)
	fn(t, language.English,
		"{_0, selectordinal, other{# a} one{# b} two{# c} few{# d}}",
		nil, icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyIgnore,
		1, nil, nil)
	fn(t, language.Ukrainian,
		"{_0, selectordinal, other{# a} few{# b}}",
		nil, icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyIgnore,
		1, nil, nil)
	fn(t, language.Arabic,
		"{_0, selectordinal, other{# a}}",
		nil, icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyIgnore,
		1, nil, nil)
	fn(t, language.Kazakh,
		"{_0, selectordinal, other{# a} many{# b}}",
		nil, icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyIgnore,
		1, nil, nil)
	fn(t, language.MustParse("cy"),
		"{_0, selectordinal, other{#a} zero{#b} one{#c} two{#d} few{#e} many{#f}}",
		nil, icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyIgnore,
		1, nil, nil)

	// Expect full completeness for required options.
	fn(t, language.English,
		"{_0, select, foo{foo} bar{bar} other{other}}",
		map[string][]string{"_0": {"foo", "bar"}},
		icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyIgnore,
		1, nil, nil)
	fn(t, language.English,
		"{_0, select, foo{foo} bar{bar} other{other}}",
		map[string][]string{"_0": {"foo", "bar"}},
		icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyIgnore,
		1, nil, nil)

	// Expect full completeness for optional options.
	fn(t, language.English,
		"no bar: {_0, select, foo{foo} other{other}}",
		map[string][]string{"_0": {"foo", "bar"}},
		icumsg.OptionsPresencePolicyOptional, icumsg.OptionUnknownPolicyIgnore,
		1, nil, nil)
	fn(t, language.English,
		"no foo: {_0, select, bar{bar} other{other}}",
		map[string][]string{"_0": {"foo", "bar"}},
		icumsg.OptionsPresencePolicyOptional, icumsg.OptionUnknownPolicyIgnore,
		1, nil, nil)
	fn(t, language.English,
		"neither foo nor bar: {_0, select, other{other}}",
		map[string][]string{"_0": {"foo", "bar"}},
		icumsg.OptionsPresencePolicyOptional, icumsg.OptionUnknownPolicyIgnore,
		1, nil, nil)

	// Expect incomplete options.
	fn(t, language.English, "missing one: {_0, plural, =0{a} =1{b} other{# d}}",
		nil, 0, 0, 1, []string{"{_0, plural, =0{a} =1{b} other{# d}}"}, nil)
	fn(t, language.English, "missing one: {_0, plural, other{# a}}",
		nil, 0, 0, 1, []string{"{_0, plural, other{# a}}"}, nil)
	fn(t, language.Ukrainian, "missing: one,few,many: {_0, selectordinal, other{# a}}",
		nil, 0, 0, 1, []string{"{_0, selectordinal, other{# a}}"}, nil)
	fn(t, language.Arabic,
		"missing two: {_0, plural, one{# b} other{# a} few{# c} many{# d} zero{# e}}",
		nil, 0, 0, 1, []string{
			"{_0, plural, one{# b} other{# a} few{# c} many{# d} zero{# e}}",
		}, nil)
	fn(t, language.Kazakh,
		"missing: many {_0, selectordinal, other{# a}}",
		nil, 0, 0, 1, []string{"{_0, selectordinal, other{# a}}"}, nil)

	// Expect rejected options.
	fn(t, language.English, "unknown c,d: {_0, select, d{d} a{a} b{b} c{c} other{o}}",
		map[string][]string{"_0": {"a", "b"}},
		icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyReject,
		1, nil, []string{"d{d}", "c{c}"})
	fn(t, language.English,
		"missing a and unknown c,d: {_0, select, d{d} b{b} c{c} other{o}}",
		map[string][]string{"_0": {"a", "b"}},
		icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyReject,
		1, []string{"{_0, select, d{d} b{b} c{c} other{o}}"}, []string{"d{d}", "c{c}"})
	fn(t, language.English,
		"missing a and unknown c,d: {_0, select, d{d} b{b} c{c} other{o}}",
		map[string][]string{"_0": {"a", "b"}},
		icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyReject,
		1, []string{"{_0, select, d{d} b{b} c{c} other{o}}"}, []string{"d{d}", "c{c}"})

	// Ignore unknown options.
	fn(t, language.English,
		"missing a and ignored unknown c,d: {_0, select, d{d} b{b} c{c} other{o}}",
		map[string][]string{"_0": {"a", "b"}},
		icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyIgnore,
		1, []string{"{_0, select, d{d} b{b} c{c} other{o}}"}, nil)

	// Nested select in plural.
	fn(t, language.English,
		`missing a in one and b in other: {_0, plural,
			one{   {_1, select, b{b} other{o}} }
			other{ {_1, select, a{a} other{o}} }
		}`,
		map[string][]string{"_1": {"a", "b"}},
		icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyIgnore,
		3, []string{
			"{_1, select, b{b} other{o}}",
			"{_1, select, a{a} other{o}}",
		}, nil)

	// Select in ordinal.
	fn(t, language.English,
		`some missing and some rejected: {_0, selectordinal,
			one{   {_1, select, b{b} other{o}} }
			two{   {_1, select, a{a} other{o}} }
			few{   {_1, select, a{a} b{b} c{c} other{o}} }
			other{ {_1, select, a{a} b{b} d{d} other{o}} }
		}`,
		map[string][]string{"_1": {"a", "b"}},
		icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyReject,
		5, []string{
			"{_1, select, b{b} other{o}}",
			"{_1, select, a{a} other{o}}",
		}, []string{
			"c{c}",
			"d{d}",
		})
}

func Fuzz(f *testing.F) {
	for _, tt := range TestsErrors {
		f.Add(tt.Input)
	}
	for _, tt := range TestsErrorsLocale {
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
		_, _ = tokenizer.Tokenize(language.English, buffer, input)
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
				buffer, err = tokenizer.Tokenize(language.English, buffer, input)
				if err != nil {
					panic(err)
				}
			}
		})
	}
}
