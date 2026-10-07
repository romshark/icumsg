package icumsg_test

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/text/language"

	"github.com/romshark/icumsg"
	"github.com/romshark/icumsg/internal/test"
)

func ReadFile[S string | []byte](tb testing.TB, fileName string) S {
	tb.Helper()
	fc, err := os.ReadFile(fileName)
	test.RequireNoErr(tb, err)
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
		test.RequireEqual(t, expect, tp.String())
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
	f(t, "argument style skeleton", icumsg.TokenTypeArgStyleSkeleton)
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
		test.RequireNoErr(t, err)
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
	f(t, language.English, "a'}'b", []Token{
		{Str: "a'}'b", Type: icumsg.TokenTypeLiteral},
	}...)

	// An apostrophe not immediately preceding a syntax character is literal text.
	f(t, language.English, "This isn't obvious", []Token{
		{Str: "This isn't obvious", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, language.English, "That's {n} messages", []Token{
		{Str: "That's ", Type: icumsg.TokenTypeLiteral},
		{Str: "{n}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "n", Type: icumsg.TokenTypeArgName},
		{Str: " messages", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, language.English, "aujourd'hui", []Token{
		{Str: "aujourd'hui", Type: icumsg.TokenTypeLiteral},
	}...)
	// A doubled apostrophe escapes one literal apostrophe.
	f(t, language.English, "He''s here {n}", []Token{
		{Str: "He''s here ", Type: icumsg.TokenTypeLiteral},
		{Str: "{n}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "n", Type: icumsg.TokenTypeArgName},
	}...)
	// Ordinary and quoting apostrophes in the same message.
	f(t, language.English, "it's not '{arg}'", []Token{
		{Str: "it's not '{arg}'", Type: icumsg.TokenTypeLiteral},
	}...)
	// A trailing apostrophe never opens a quote.
	f(t, language.English, "{n} isn't", []Token{
		{Str: "{n}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "n", Type: icumsg.TokenTypeArgName},
		{Str: " isn't", Type: icumsg.TokenTypeLiteral},
	}...)

	// '#' is a syntax character in plural and selectordinal options.
	// An apostrophe preceding it starts quoted text.
	f(t, language.English, "{n,plural,other{'#{x}' }}", []Token{
		{Str: "{n,plural,other{'#{x}' }}", Type: icumsg.TokenTypePlural},
		{Str: "n", Type: icumsg.TokenTypeArgName},
		{Str: "other{'#{x}' }", Type: icumsg.TokenTypeOptionOther},
		{Str: "'#{x}' ", Type: icumsg.TokenTypeLiteral},
		{Str: "other{'#{x}' }", Type: icumsg.TokenTypeOptionTerm},
		{Str: "{n,plural,other{'#{x}' }}", Type: icumsg.TokenTypeComplexArgTerm},
	}...)
	f(t, language.English, "{n,selectordinal,other{'#{x}' }}", []Token{
		{Str: "{n,selectordinal,other{'#{x}' }}", Type: icumsg.TokenTypeSelectOrdinal},
		{Str: "n", Type: icumsg.TokenTypeArgName},
		{Str: "other{'#{x}' }", Type: icumsg.TokenTypeOptionOther},
		{Str: "'#{x}' ", Type: icumsg.TokenTypeLiteral},
		{Str: "other{'#{x}' }", Type: icumsg.TokenTypeOptionTerm},
		{Str: "{n,selectordinal,other{'#{x}' }}", Type: icumsg.TokenTypeComplexArgTerm},
	}...)
	// The same sub-message in a select where '#' is not a syntax character.
	// The apostrophes stay literal and {x} remains an argument.
	f(t, language.English, "{g,select,other{'#{x}' }}", []Token{
		{Str: "{g,select,other{'#{x}' }}", Type: icumsg.TokenTypeSelect},
		{Str: "g", Type: icumsg.TokenTypeArgName},
		{Str: "other{'#{x}' }", Type: icumsg.TokenTypeOptionOther},
		{Str: "'#", Type: icumsg.TokenTypeLiteral},
		{Str: "{x}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "x", Type: icumsg.TokenTypeArgName},
		{Str: "' ", Type: icumsg.TokenTypeLiteral},
		{Str: "other{'#{x}' }", Type: icumsg.TokenTypeOptionTerm},
		{Str: "{g,select,other{'#{x}' }}", Type: icumsg.TokenTypeComplexArgTerm},
	}...)
	// Only the immediately enclosing argument decides.
	// '#' is not quotable in a select nested inside a plural.
	f(t, language.English, "{n,plural,other{{g,select,other{'#{x}' }}}}", []Token{
		{Str: "{n,plural,other{{g,select,other{'#{x}' }}}}", Type: icumsg.TokenTypePlural},
		{Str: "n", Type: icumsg.TokenTypeArgName},
		{Str: "other{{g,select,other{'#{x}' }}}", Type: icumsg.TokenTypeOptionOther},
		{Str: "{g,select,other{'#{x}' }}", Type: icumsg.TokenTypeSelect},
		{Str: "g", Type: icumsg.TokenTypeArgName},
		{Str: "other{'#{x}' }", Type: icumsg.TokenTypeOptionOther},
		{Str: "'#", Type: icumsg.TokenTypeLiteral},
		{Str: "{x}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "x", Type: icumsg.TokenTypeArgName},
		{Str: "' ", Type: icumsg.TokenTypeLiteral},
		{Str: "other{'#{x}' }", Type: icumsg.TokenTypeOptionTerm},
		{Str: "{g,select,other{'#{x}' }}", Type: icumsg.TokenTypeComplexArgTerm},
		{Str: "other{{g,select,other{'#{x}' }}}", Type: icumsg.TokenTypeOptionTerm},
		{Str: "{n,plural,other{{g,select,other{'#{x}' }}}}", Type: icumsg.TokenTypeComplexArgTerm},
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

	// Custom styles are string patterns that end at the first '}'
	// that's neither quoted nor closing a '{' within the style.
	f(t, language.English, "{n, number, #,##0.00}", []Token{
		{Str: "{n, number, #,##0.00}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "n", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "#,##0.00", Type: icumsg.TokenTypeArgStyleCustom},
	}...)
	f(t, language.English, "{d, date, yyyy-MM-dd} after", []Token{
		{Str: "{d, date, yyyy-MM-dd}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "d", Type: icumsg.TokenTypeArgName},
		{Str: "date", Type: icumsg.TokenTypeArgTypeDate},
		{Str: "yyyy-MM-dd", Type: icumsg.TokenTypeArgStyleCustom},
		{Str: " after", Type: icumsg.TokenTypeLiteral},
	}...)
	f(t, language.English, "{d, date, EEE, MMM d }", []Token{
		{Str: "{d, date, EEE, MMM d }", Type: icumsg.TokenTypeSimpleArg},
		{Str: "d", Type: icumsg.TokenTypeArgName},
		{Str: "date", Type: icumsg.TokenTypeArgTypeDate},
		{Str: "EEE, MMM d", Type: icumsg.TokenTypeArgStyleCustom},
	}...)
	f(t, language.English, "{n, spellout, %spellout-ordinal}", []Token{
		{Str: "{n, spellout, %spellout-ordinal}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "n", Type: icumsg.TokenTypeArgName},
		{Str: "spellout", Type: icumsg.TokenTypeArgTypeSpellout},
		{Str: "%spellout-ordinal", Type: icumsg.TokenTypeArgStyleCustom},
	}...)
	// Every apostrophe in a style starts or ends quoted text.
	f(t, language.English, "{t, time, h 'o''clock' a}", []Token{
		{Str: "{t, time, h 'o''clock' a}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "t", Type: icumsg.TokenTypeArgName},
		{Str: "time", Type: icumsg.TokenTypeArgTypeTime},
		{Str: "h 'o''clock' a", Type: icumsg.TokenTypeArgStyleCustom},
	}...)
	f(t, language.English, "{n, number, #'}'}", []Token{
		{Str: "{n, number, #'}'}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "n", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "#'}'", Type: icumsg.TokenTypeArgStyleCustom},
	}...)
	f(t, language.English, "{n, number, {#}}", []Token{
		{Str: "{n, number, {#}}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "n", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "{#}", Type: icumsg.TokenTypeArgStyleCustom},
	}...)
	// A keyword is only recognized as the whole style.
	f(t, language.English, "{n, number, integer, foobar}", []Token{
		{Str: "{n, number, integer, foobar}", Type: icumsg.TokenTypeSimpleArg},
		{Str: "n", Type: icumsg.TokenTypeArgName},
		{Str: "number", Type: icumsg.TokenTypeArgTypeNumber},
		{Str: "integer, foobar", Type: icumsg.TokenTypeArgStyleCustom},
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

	// Whitespace at the start of an option is message text.
	f(t, language.English, "{g,select,male{ he} other{  they}}", []Token{
		{Str: "{g,select,male{ he} other{  they}}", Type: icumsg.TokenTypeSelect},
		{Str: "g", Type: icumsg.TokenTypeArgName},
		{Str: "male{ he}", Type: icumsg.TokenTypeOption},
		{Str: "male", Type: icumsg.TokenTypeOptionName},
		{Str: " he", Type: icumsg.TokenTypeLiteral},
		{Str: "male{ he}", Type: icumsg.TokenTypeOptionTerm},
		{Str: "other{  they}", Type: icumsg.TokenTypeOptionOther},
		{Str: "  they", Type: icumsg.TokenTypeLiteral},
		{Str: "other{  they}", Type: icumsg.TokenTypeOptionTerm},
		{Str: "{g,select,male{ he} other{  they}}", Type: icumsg.TokenTypeComplexArgTerm},
	}...)

	{ // Indentation in a multi-line message.
		full := "{g,select,\n  male {\n    {n,plural,other{ #}}\n  }\n  other {x}\n}"
		male := "male {\n    {n,plural,other{ #}}\n  }"
		plural := "{n,plural,other{ #}}"
		f(t, language.English, full, []Token{
			{Str: full, Type: icumsg.TokenTypeSelect},
			{Str: "g", Type: icumsg.TokenTypeArgName},
			{Str: male, Type: icumsg.TokenTypeOption},
			{Str: "male", Type: icumsg.TokenTypeOptionName},
			{Str: "\n    ", Type: icumsg.TokenTypeLiteral},
			{Str: plural, Type: icumsg.TokenTypePlural},
			{Str: "n", Type: icumsg.TokenTypeArgName},
			{Str: "other{ #}", Type: icumsg.TokenTypeOptionOther},
			{Str: " #", Type: icumsg.TokenTypeLiteral},
			{Str: "other{ #}", Type: icumsg.TokenTypeOptionTerm},
			{Str: plural, Type: icumsg.TokenTypeComplexArgTerm},
			{Str: "\n  ", Type: icumsg.TokenTypeLiteral},
			{Str: male, Type: icumsg.TokenTypeOptionTerm},
			{Str: "other {x}", Type: icumsg.TokenTypeOptionOther},
			{Str: "x", Type: icumsg.TokenTypeLiteral},
			{Str: "other {x}", Type: icumsg.TokenTypeOptionTerm},
			{Str: full, Type: icumsg.TokenTypeComplexArgTerm},
		}...)
	}

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

	// Nested arguments may reuse the option names of the enclosing argument.
	f(t, language.English, "{a,select,x{{b,select,x{ok} other{ok}}} other{ok}}",
		[]Token{
			{Str: "{a,select,x{{b,select,x{ok} other{ok}}} other{ok}}", Type: icumsg.TokenTypeSelect},
			{Str: "a", Type: icumsg.TokenTypeArgName},
			{Str: "x{{b,select,x{ok} other{ok}}}", Type: icumsg.TokenTypeOption},
			{Str: "x", Type: icumsg.TokenTypeOptionName},
			{Str: "{b,select,x{ok} other{ok}}", Type: icumsg.TokenTypeSelect},
			{Str: "b", Type: icumsg.TokenTypeArgName},
			{Str: "x{ok}", Type: icumsg.TokenTypeOption},
			{Str: "x", Type: icumsg.TokenTypeOptionName},
			{Str: "ok", Type: icumsg.TokenTypeLiteral},
			{Str: "x{ok}", Type: icumsg.TokenTypeOptionTerm},
			{Str: "other{ok}", Type: icumsg.TokenTypeOptionOther},
			{Str: "ok", Type: icumsg.TokenTypeLiteral},
			{Str: "other{ok}", Type: icumsg.TokenTypeOptionTerm},
			{Str: "{b,select,x{ok} other{ok}}", Type: icumsg.TokenTypeComplexArgTerm},
			{Str: "x{{b,select,x{ok} other{ok}}}", Type: icumsg.TokenTypeOptionTerm},
			{Str: "other{ok}", Type: icumsg.TokenTypeOptionOther},
			{Str: "ok", Type: icumsg.TokenTypeLiteral},
			{Str: "other{ok}", Type: icumsg.TokenTypeOptionTerm},
			{Str: "{a,select,x{{b,select,x{ok} other{ok}}} other{ok}}", Type: icumsg.TokenTypeComplexArgTerm},
		}...)
	f(t, language.English, "{a,plural,=0{{b,plural,=0{ok} other{ok}}} other{ok}}",
		[]Token{
			{Str: "{a,plural,=0{{b,plural,=0{ok} other{ok}}} other{ok}}", Type: icumsg.TokenTypePlural},
			{Str: "a", Type: icumsg.TokenTypeArgName},
			{Str: "=0{{b,plural,=0{ok} other{ok}}}", Type: icumsg.TokenTypeOptionNumber},
			{Str: "=0", Type: icumsg.TokenTypeOptionName},
			{Str: "{b,plural,=0{ok} other{ok}}", Type: icumsg.TokenTypePlural},
			{Str: "b", Type: icumsg.TokenTypeArgName},
			{Str: "=0{ok}", Type: icumsg.TokenTypeOptionNumber},
			{Str: "=0", Type: icumsg.TokenTypeOptionName},
			{Str: "ok", Type: icumsg.TokenTypeLiteral},
			{Str: "=0{ok}", Type: icumsg.TokenTypeOptionTerm},
			{Str: "other{ok}", Type: icumsg.TokenTypeOptionOther},
			{Str: "ok", Type: icumsg.TokenTypeLiteral},
			{Str: "other{ok}", Type: icumsg.TokenTypeOptionTerm},
			{Str: "{b,plural,=0{ok} other{ok}}", Type: icumsg.TokenTypeComplexArgTerm},
			{Str: "=0{{b,plural,=0{ok} other{ok}}}", Type: icumsg.TokenTypeOptionTerm},
			{Str: "other{ok}", Type: icumsg.TokenTypeOptionOther},
			{Str: "ok", Type: icumsg.TokenTypeLiteral},
			{Str: "other{ok}", Type: icumsg.TokenTypeOptionTerm},
			{Str: "{a,plural,=0{{b,plural,=0{ok} other{ok}}} other{ok}}", Type: icumsg.TokenTypeComplexArgTerm},
		}...)
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
		22, icumsg.ErrUnsupportedPluralRule,
	},
	{
		"{x,plural, other{yes} few{no}}", language.AmericanEnglish,
		22, icumsg.ErrUnsupportedPluralRule,
	},
	{
		"{x,plural, other{yes} zero{no}}", language.Ukrainian,
		22, icumsg.ErrUnsupportedPluralRule,
	},
	{
		"{x,plural, one{yes} two{no} other{yes}}", language.German,
		20, icumsg.ErrUnsupportedPluralRule,
	},
	{
		"{x,selectordinal, other{yes} one{no}}", language.German,
		29, icumsg.ErrUnsupportedPluralRule,
	},
	{
		"{x,selectordinal, other{yes} zero{no}}", language.German,
		29, icumsg.ErrUnsupportedPluralRule,
	},
	{
		"{x,selectordinal, other{yes} two{no}}", language.German,
		29, icumsg.ErrUnsupportedPluralRule,
	},
	{
		"{x,selectordinal, other{yes} many{no}}", language.German,
		29, icumsg.ErrUnsupportedPluralRule,
	},
	{
		"{x,selectordinal, other{yes} few{no}}", language.German,
		29, icumsg.ErrUnsupportedPluralRule,
	},
	{
		"{x,selectordinal, other{yes} zero{no}}", language.Ukrainian,
		29, icumsg.ErrUnsupportedPluralRule,
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
			test.RequireErrIs(t, tt.ExpectErr, err)
			test.RequireEqual(t, tt.ExpectErrIndex, tokenizer.Pos())
		})
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
	{"{x, number, ::", 14, icumsg.ErrUnexpectedEOF},
	{"{x, number, {#}", 15, icumsg.ErrUnexpectedEOF}, // Unbalanced '{' in style.
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
	// Unclosed quote.
	// Every apostrophe below immediately precedes a syntax character.
	{"prefix '{unclosed quote", 7, icumsg.ErrUnclosedQuote},
	{"prefix '}unclosed quote", 7, icumsg.ErrUnclosedQuote},
	{"prefix '' '{unclosed quote", 10, icumsg.ErrUnclosedQuote},
	{"prefix '{}' '{unclosed quote", 12, icumsg.ErrUnclosedQuote},
	{"{x,plural, other { '{}' '{ }}", 24, icumsg.ErrUnclosedQuote},
	{"'{", 0, icumsg.ErrUnclosedQuote},
	{"'}", 0, icumsg.ErrUnclosedQuote},
	// '#' is a syntax character in a plural option.
	{"{x,plural, other {'# }}", 18, icumsg.ErrUnclosedQuote},
	{"{x,selectordinal, other {'# }}", 25, icumsg.ErrUnclosedQuote},
	// The trailing apostrophe precedes '}' and opens a quote that swallows
	// the closing brackets.
	{"{g,select,other{'#{x}'}}", 21, icumsg.ErrUnclosedQuote},
	// Every apostrophe in an argument style starts or ends quoted text.
	{"{x, number, '#}", 12, icumsg.ErrUnclosedQuote},
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
	// Duplicate option next to a nested argument reusing the option name.
	{"{a,select,x{{b,select,x{ok} x{ok} other{ok}}} other{ok}}", 28,
		icumsg.ErrDuplicateOption},
	{"{a,select,x{{b,select,x{ok} other{ok}}} x{ok} other{ok}}", 40,
		icumsg.ErrDuplicateOption},
	{"{a,plural,=0{{b,plural,=0{ok} =0{ok} other{ok}}} other{ok}}", 30,
		icumsg.ErrDuplicateOption},
	{"{a,plural,=0{{b,plural,=0{ok} other{ok}}} =0{ok} other{ok}}", 42,
		icumsg.ErrDuplicateOption},
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
		test.RequireNoErr(t, err)
		t.Run("", func(t *testing.T) {
			buffer = buffer[:0]
			_, err := tokenizer.Tokenize(l, buffer, tt.Input)
			t.Logf("input: %q", tt.Input)
			test.RequireErrIs(t, tt.ExpectErr, err)
			test.RequireEqual(t, tt.ExpectErrIndex, tokenizer.Pos())
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
		test.RequireNoErr(t, err)
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
	fn(t, "Prefix {x,plural,=0{zero} one{one} other{other}}", 1,
		Token{Str: "=0{zero}", Type: icumsg.TokenTypeOptionNumber},
		Token{Str: "one{one}", Type: icumsg.TokenTypeOptionOne},
		Token{Str: "other{other}", Type: icumsg.TokenTypeOptionOther})

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

	{
		// Options nested in an =n option belong to the nested argument.
		nested := "Prefix {x,plural,=0{{g,select,male{m} other{o}}} one{a} other{b}}"
		fn(t, nested, 1, // plural
			Token{Str: "=0{{g,select,male{m} other{o}}}", Type: icumsg.TokenTypeOptionNumber},
			Token{Str: "one{a}", Type: icumsg.TokenTypeOptionOne},
			Token{Str: "other{b}", Type: icumsg.TokenTypeOptionOther})
		fn(t, nested, 5, // x=0
			Token{Str: "male{m}", Type: icumsg.TokenTypeOption},
			Token{Str: "other{o}", Type: icumsg.TokenTypeOptionOther})
	}
}

func TestOptionsBreak(t *testing.T) {
	var tokenizer icumsg.Tokenizer

	input := "Prefix {x,selectordinal,other{o}one{a}few{b}two{c}}"
	buffer, err := tokenizer.Tokenize(language.English, nil, input)
	test.RequireNoErr(t, err)

	itr := 0
	for range icumsg.Options(buffer, 1) {
		itr++
		break
	}
	test.RequireEqual(t, 1, itr)
}

func TestAnalyze(t *testing.T) {
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
		test.RequireNoErr(t, err)
		var incomplete, rejected []string
		actualTotal, err := icumsg.Analyze(locale, input, buffer,
			func(argName string) (
				[]string, icumsg.OptionsPresencePolicy, icumsg.OptionUnknownPolicy,
			) {
				return options[argName], presencePolicy, unknownPolicy
			},
			func(index int) error {
				incomplete = append(incomplete, buffer[index].String(input, buffer))
				return nil
			},
			func(indexArgument, indexOption int) error {
				rejected = append(rejected, buffer[indexOption].String(input, buffer))
				return nil
			})
		test.RequireNoErr(t, err)
		test.RequireEqual(t, expectTotal, actualTotal, "total")
		test.RequireDeepEqual(t, expectIncomplete, incomplete, "incomplete")
		test.RequireDeepEqual(t, expectRejected, rejected, "rejected")
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

	// An empty non-nil slice of options allows only other.
	fn(t, language.English, "unknown male: {_0, select, male{x} other{y}}",
		map[string][]string{"_0": {}},
		icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyReject,
		1, nil, []string{"male{x}"})

	// A nil slice of options applies no policies.
	fn(t, language.English, "{_0, select, male{x} other{y}}",
		nil, icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyReject,
		1, nil, nil)

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

	// Nested plural in a select without options.
	fn(t, language.English,
		"missing one: {_0, select, male{{_1, plural, other{x}}} other{x}}",
		nil, 0, 0, 2, []string{"{_1, plural, other{x}}"}, nil)

	// Nested select in an =n option.
	fn(t, language.English,
		`missing b in =0: {_0, plural,
			=0{    {_1, select, a{a} other{o}} }
			one{   {_1, select, a{a} b{b} other{o}} }
			other{ {_1, select, a{a} b{b} other{o}} }
		}`,
		map[string][]string{"_1": {"a", "b"}},
		icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyIgnore,
		4, []string{"{_1, select, a{a} other{o}}"}, nil)

	// Options of a plural nested in an =n option don't count toward the outer plural.
	fn(t, language.English,
		"missing one: {_0, plural, =0{{_1, plural, one{a} other{b}}} other{c}}",
		nil, 0, 0, 2,
		[]string{"{_0, plural, =0{{_1, plural, one{a} other{b}}} other{c}}"}, nil)

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

	// Adjacent choices.
	fn(t, language.English,
		"{_0, select, other{a}}{_1, plural, other{b}}"+
			"{_2, selectordinal, other{c}}{_3, plural, other{d}}",
		nil, 0, 0, 4, []string{
			"{_1, plural, other{b}}",
			"{_2, selectordinal, other{c}}",
			"{_3, plural, other{d}}",
		}, nil)

	// Adjacent choices nested in an option.
	fn(t, language.English,
		"{_0, plural, one{x} other{{_1, plural, other{a}}{_2, plural, other{b}}}}",
		nil, 0, 0, 3, []string{
			"{_1, plural, other{a}}",
			"{_2, plural, other{b}}",
		}, nil)
}

func TestCompletenessErrors(t *testing.T) {
	var tokenizer icumsg.Tokenizer
	var buffer []icumsg.Token

	optionsNone := func(argName string) (
		[]string, icumsg.OptionsPresencePolicy, icumsg.OptionUnknownPolicy,
	) {
		return nil, 0, 0
	}
	optionsGender := func(argName string) (
		[]string, icumsg.OptionsPresencePolicy, icumsg.OptionUnknownPolicy,
	) {
		if strings.HasSuffix(argName, "_gender") {
			return []string{"male", "female"},
				icumsg.OptionsPresencePolicyRequired, icumsg.OptionUnknownPolicyReject
		}
		return nil, 0, 0
	}

	fn := func(
		t *testing.T,
		locale language.Tag,
		input string,
		options icumsg.SelectOptions,
		checkErr func(t *testing.T, errs []error),
	) {
		t.Helper()
		buffer = buffer[:0]
		var err error
		buffer, err = tokenizer.Tokenize(locale, buffer, input)
		test.RequireNoErr(t, err, "tokenizer error")
		errSeq := icumsg.Errors(locale, input, buffer, options)

		var actual []error
		for err := range errSeq {
			actual = append(actual, err)
		}

		checkErr(t, actual)
	}

	checkMissingPlural := func(tp, missing string) func(t *testing.T, errs []error) {
		return func(t *testing.T, errs []error) {
			test.RequireEqual(t, 1, len(errs))
			test.RequireErrType[icumsg.ErrorPluralMissingOption](t, errs[0])
			expect := fmt.Sprintf("missing %s plural options [%s]", tp, missing)
			test.RequireEqual(t, expect, errs[0].Error())
		}
	}

	checkNoErrs := func(t *testing.T, errs []error) { test.RequireEqual(t, 0, len(errs)) }

	// Expect full completeness.
	fn(t, language.English, "Simple literal text message", optionsNone, checkNoErrs)

	fn(t, language.English,
		"{var0, plural, one{1} other{2}}",
		optionsNone, checkNoErrs)

	fn(t, language.English,
		"{var0_gender, select, other{2}}",
		optionsNone, checkNoErrs)

	fn(t, language.English,
		"{var0_gender, select, female{1} male{2} other{3}}",
		optionsGender, checkNoErrs)

	fn(t, language.English,
		"{var0, selectordinal, one{1} two{2} few{3} other{4}}",
		optionsNone, checkNoErrs)

	fn(t, language.Arabic,
		"{var0, plural, zero{0} one{1} two{2} few{3} many{4} other{5}}",
		optionsNone, checkNoErrs)

	fn(t, language.MustParse("cy"),
		"{var0, selectordinal, zero{0} one{1} two{2} few{3} many{4} other{5}}",
		optionsNone, checkNoErrs)

	// Expect invalid select option.
	fn(t, language.English,
		"unknown is wrong: {var0_gender, select, unknown{-} female{-} male{-} other{-}}",
		optionsGender, func(t *testing.T, errs []error) {
			test.RequireEqual(t, 1, len(errs))
			test.RequireErrType[icumsg.ErrorSelectInvalidOption](t, errs[0])
			e := errs[0].(icumsg.ErrorSelectInvalidOption)
			test.RequireEqual(t, 1, e.TokenIndexArgument)
			test.RequireEqual(t, 3, e.TokenIndexOption)
			test.RequireEqual(t, "invalid select option", errs[0].Error())
		})

	// Expect incomplete.
	fn(t, language.English,
		"missing cardinal: {var0, plural, other{o}}",
		optionsNone, checkMissingPlural("cardinal", "one"))

	fn(t, language.Ukrainian,
		"missing cardinal: {var0, plural, other{o}}",
		optionsNone, checkMissingPlural("cardinal", "one,few,many"))

	fn(t, language.Arabic,
		"missing cardinal: {var0, plural, other{o}}",
		optionsNone, checkMissingPlural("cardinal", "zero,one,two,few,many"))

	fn(t, language.Ukrainian,
		"missing cardinal: {var0, plural, one{1} other{2}}",
		optionsNone, checkMissingPlural("cardinal", "few,many"))

	fn(t, language.Ukrainian,
		"missing cardinal: {var0, plural, one{1} many{3} other{2}}",
		optionsNone, checkMissingPlural("cardinal", "few"))

	fn(t, language.Arabic,
		"{var0, plural, one{1} two{2} few{3} many{4} other{5}}",
		optionsNone, checkMissingPlural("cardinal", "zero"))

	fn(t, language.Arabic,
		"{var0, plural, zero{0} one{1} two{2} many{4} other{5}}",
		optionsNone, checkMissingPlural("cardinal", "few"))

	fn(t, language.English,
		"missing ordinal: {var0, selectordinal, other{1}}",
		optionsNone, checkMissingPlural("ordinal", "one,two,few"))

	fn(t, language.English,
		"missing ordinal: {var0, selectordinal, other{1} two{2}}",
		optionsNone, checkMissingPlural("ordinal", "one,few"))

	fn(t, language.English,
		"{var0_gender, select, other{1}}",
		optionsGender,
		func(t *testing.T, errs []error) {
			test.RequireEqual(t, 1, len(errs))
			test.RequireErrType[icumsg.ErrorSelectMissingOption](t, errs[0])
			test.RequireEqual(t, "missing select options [male,female]",
				errs[0].Error())
		})

	fn(t, language.English,
		"{var0_gender, select, male{2} other{3}}",
		optionsGender,
		func(t *testing.T, errs []error) {
			test.RequireEqual(t, 1, len(errs))
			test.RequireErrType[icumsg.ErrorSelectMissingOption](t, errs[0])
			test.RequireEqual(t, "missing select options [female]",
				errs[0].Error())
		})

	// Multiple.
	fn(t, language.Ukrainian,
		`
			missing male: {var0_gender, select, female{0} other{3}}
			missing female: {var1_gender, select, male{0} other{3}}
			missing male and female: {var2_gender, select, other{3}}
			missing few: {var3, selectordinal, other{3}}
			missing few: {var4, plural, other{3}}
		`,
		optionsGender,
		func(t *testing.T, errs []error) {
			test.RequireEqual(t, 5, len(errs))
			test.RequireErrType[icumsg.ErrorSelectMissingOption](t, errs[0])
			test.RequireEqual(t, "missing select options [male]",
				errs[0].Error())

			test.RequireErrType[icumsg.ErrorSelectMissingOption](t, errs[1])
			test.RequireEqual(t, "missing select options [female]",
				errs[1].Error())

			test.RequireErrType[icumsg.ErrorSelectMissingOption](t, errs[2])
			test.RequireEqual(t, "missing select options [male,female]",
				errs[2].Error())

			test.RequireErrType[icumsg.ErrorPluralMissingOption](t, errs[3])
			test.RequireEqual(t, "missing ordinal plural options [few]",
				errs[3].Error())

			test.RequireErrType[icumsg.ErrorPluralMissingOption](t, errs[4])
			test.RequireEqual(t, "missing cardinal plural options [one,few,many]",
				errs[4].Error())
		})

	// Nested.
	fn(t, language.Ukrainian,
		`{var0_gender, select, female{0} male{
			{var1, plural, other{x}}
		} other{3}}`,
		optionsGender,
		func(t *testing.T, errs []error) {
			test.RequireEqual(t, 1, len(errs))
			test.RequireErrType[icumsg.ErrorPluralMissingOption](t, errs[0])
			test.RequireEqual(t, "missing cardinal plural options [one,few,many]",
				errs[0].Error())
		})

	fn(t, language.Ukrainian,
		`{var0, plural, one{-} few{-} many{
			{var1, plural, one{-} other{-}}
		} other{-}}`,
		optionsGender,
		func(t *testing.T, errs []error) {
			test.RequireEqual(t, 1, len(errs))
			test.RequireErrType[icumsg.ErrorPluralMissingOption](t, errs[0])
			test.RequireEqual(t, "missing cardinal plural options [few,many]",
				errs[0].Error())
		})

	fn(t, language.English,
		`{var0, selectordinal, one{-} two{-} few{-} other{
			{var1, selectordinal, one{-} other{-}}
		}}`,
		optionsGender,
		func(t *testing.T, errs []error) {
			test.RequireEqual(t, 1, len(errs))
			test.RequireErrType[icumsg.ErrorPluralMissingOption](t, errs[0])
			test.RequireEqual(t, "missing ordinal plural options [two,few]",
				errs[0].Error())
		})

	// The nested "one" doesn't count toward the outer plural.
	fn(t, language.English,
		"{var0, plural, =0{{var1, plural, one{-} other{-}}} other{-}}",
		optionsNone, checkMissingPlural("cardinal", "one"))

	// Nested in a select without options.
	fn(t, language.English,
		"{var0, select, male{{var1, plural, other{-}}} other{-}}",
		optionsNone, checkMissingPlural("cardinal", "one"))

	// Multiple nested.
	fn(t, language.English,
		`missing female: {var0_gender, select, male{-} other{
			missing two: {var1, selectordinal, one{-} few{-} other{
				complete: {var2, plural, one{-} other{
					missing one: {var3, plural, other{-}}
				}}
				missing and invalid: {var4_gender, select, none{-} male{-} other{-}}
			}}
		}}`,
		optionsGender,
		func(t *testing.T, errs []error) {
			test.RequireEqual(t, 5, len(errs))

			test.RequireErrType[icumsg.ErrorPluralMissingOption](t, errs[0])
			test.RequireEqual(t, "missing cardinal plural options [one]",
				errs[0].Error())

			test.RequireErrType[icumsg.ErrorSelectInvalidOption](t, errs[1])
			test.RequireEqual(t, "invalid select option",
				errs[1].Error())

			test.RequireErrType[icumsg.ErrorSelectMissingOption](t, errs[2])
			test.RequireEqual(t, "missing select options [female]",
				errs[2].Error())

			test.RequireErrType[icumsg.ErrorPluralMissingOption](t, errs[3])
			test.RequireEqual(t, "missing ordinal plural options [two]",
				errs[3].Error())

			test.RequireErrType[icumsg.ErrorSelectMissingOption](t, errs[4])
			test.RequireEqual(t, "missing select options [female]",
				errs[4].Error())

		})
}

func TestCompletenessErrorsBreak(t *testing.T) {
	optionsNone := func(argName string) (
		[]string, icumsg.OptionsPresencePolicy, icumsg.OptionUnknownPolicy,
	) {
		return nil, 0, 0
	}

	locale := language.English
	const input = `
		first: {var0, plural, other{-}}
		second: {var1, plural, other{-}}
	`

	var tokenizer icumsg.Tokenizer
	buffer, err := tokenizer.Tokenize(locale, nil, input)
	test.RequireNoErr(t, err, "tokenizer error")

	errSeq := icumsg.Errors(locale, input, buffer, optionsNone)

	var actual []error
	for err := range errSeq {
		actual = append(actual, err)
		break
	}
	test.RequireEqual(t, 1, len(actual))
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
