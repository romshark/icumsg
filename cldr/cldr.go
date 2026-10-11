// Package cldr reports a locale's [CLDR] plural categories.
//
// [CLDR]: https://cldr.unicode.org/
package cldr

import (
	"github.com/romshark/icumsg/internal/cldr"
	"golang.org/x/text/language"
)

// CLDR plural category names.
const (
	RuleZero  = "zero"
	RuleOne   = "one"
	RuleTwo   = "two"
	RuleFew   = "few"
	RuleMany  = "many"
	RuleOther = "other"
)

// PluralRules marks the plural categories a locale uses.
type PluralRules struct{ Zero, One, Two, Few, Many, Other bool }

// LocalePluralRules returns locale's cardinal and ordinal rules. Missing rules
// fall back to its explicit language, then to [language.Und], which has only "other".
// Tags without a language use the root. CLDR finds a [parent locale]
// by removing trailing subtags; it does not infer a language.
//
// [parent locale]: https://www.unicode.org/reports/tr35/#Parent_Locales
func LocalePluralRules(locale language.Tag) (cardinal, ordinal PluralRules) {
	r := cldr.Lookup(locale)
	return PluralRules(r.Cardinal), PluralRules(r.Ordinal)
}
