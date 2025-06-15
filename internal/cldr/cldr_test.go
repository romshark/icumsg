package cldr_test

import (
	"testing"

	"github.com/romshark/icumsg/internal/cldr"
	"golang.org/x/text/language"
)

func requireEqual[T comparable](tb testing.TB, expect, actual T) {
	tb.Helper()
	if expect != actual {
		tb.Fatalf("\nexpected: %#v;\nreceived: %#v", expect, actual)
	}
}

func TestPluralRulesByTag(t *testing.T) {
	p := cldr.PluralRulesByTag[language.English]
	requireEqual(t, cldr.Rules{Other: true, One: true}, p.Cardinal)
	requireEqual(t, cldr.Rules{Other: true, One: true, Two: true, Few: true}, p.Ordinal)

	p = cldr.PluralRulesByTag[language.German]
	requireEqual(t, cldr.Rules{Other: true, One: true}, p.Cardinal)
	requireEqual(t, cldr.Rules{Other: true}, p.Ordinal)

	p = cldr.PluralRulesByTag[language.Ukrainian]
	requireEqual(t, cldr.Rules{Other: true, One: true, Few: true, Many: true}, p.Cardinal)
	requireEqual(t, cldr.Rules{Other: true, Few: true}, p.Ordinal)

	p = cldr.PluralRulesByTag[language.French]
	requireEqual(t, cldr.Rules{Other: true, One: true, Many: true}, p.Cardinal)
	requireEqual(t, cldr.Rules{Other: true, One: true}, p.Ordinal)
}
