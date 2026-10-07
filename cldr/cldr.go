package cldr

import (
	"github.com/romshark/icumsg/internal/cldr"
	"golang.org/x/text/language"
)

const (
	RuleZero  = "zero"
	RuleOne   = "one"
	RuleTwo   = "two"
	RuleFew   = "few"
	RuleMany  = "many"
	RuleOther = "other"
)

type PluralRules struct{ Zero, One, Two, Few, Many, Other bool }

// LocalePluralRules returns cardinal and ordinal plural rules for locale.
// If CLDR has no plural rules for locale, it falls back to those of the base
// language of locale and then to those of the root locale [language.Und],
// which has only "other".
func LocalePluralRules(locale language.Tag) (cardinal, ordinal PluralRules) {
	r := cldr.Lookup(locale)
	return PluralRules(r.Cardinal), PluralRules(r.Ordinal)
}
