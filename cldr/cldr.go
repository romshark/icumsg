package cldr

import (
	"github.com/romshark/icumsg/internal/cldr"
	"golang.org/x/text/language"
)

type PluralRules struct{ Zero, One, Two, Few, Many, Other bool }

// LocalePluralRules returns cardinal and ordinal plural rules for locale.
func LocalePluralRules(locale language.Tag) (cardinal, ordinal PluralRules) {
	r, ok := cldr.PluralRulesByTag[locale]
	if !ok {
		base, _ := locale.Base()
		r = cldr.PluralRulesByBase[base]
	}
	return PluralRules(r.Cardinal), PluralRules(r.Ordinal)
}
