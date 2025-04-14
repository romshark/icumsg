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

func TestPluralFormsByTag(t *testing.T) {
	p := cldr.PluralFormsByTag[language.English]
	requireEqual(t, cldr.Forms{Other: true, One: true}, p.Cardinal)
	requireEqual(t, cldr.Forms{Other: true, One: true, Two: true, Few: true}, p.Ordinal)

	p = cldr.PluralFormsByTag[language.German]
	requireEqual(t, cldr.Forms{Other: true, One: true}, p.Cardinal)
	requireEqual(t, cldr.Forms{Other: true}, p.Ordinal)

	p = cldr.PluralFormsByTag[language.Ukrainian]
	requireEqual(t, cldr.Forms{Other: true, One: true, Few: true, Many: true}, p.Cardinal)
	requireEqual(t, cldr.Forms{Other: true, Few: true}, p.Ordinal)

	p = cldr.PluralFormsByTag[language.French]
	requireEqual(t, cldr.Forms{Other: true, One: true, Many: true}, p.Cardinal)
	requireEqual(t, cldr.Forms{Other: true, One: true}, p.Ordinal)
}
