package cldr

import "golang.org/x/text/language"

// Lookup returns the plural rules for locale.
// If CLDR has no plural rules for locale, it falls back to those of the base
// language of locale and then to those of the root locale [language.Und],
// which has only "other".
func Lookup(locale language.Tag) PluralRules {
	if r, ok := PluralRulesByTag[locale]; ok {
		return r
	}
	base, _ := locale.Base()
	if r, ok := PluralRulesByBase[base]; ok {
		return r
	}
	return PluralRulesByTag[language.Und]
}
