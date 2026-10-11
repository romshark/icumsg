package cldr

import "golang.org/x/text/language"

// Lookup returns the plural rules for locale.
// If CLDR has no plural rules for locale, it falls back to those of the
// language of locale, if locale specifies one, and then to those of the root
// locale [language.Und], which has only "other". A tag without a language,
// such as und-UA, uses the root rules. To find the [parent locale] of a tag,
// CLDR removes subtags from the end. It does not add a language.
//
// [parent locale]: https://www.unicode.org/reports/tr35/#Parent_Locales
func Lookup(locale language.Tag) PluralRules {
	if r, ok := PluralRulesByTag[locale]; ok {
		return r
	}
	// Base infers a language for a tag without one from likely subtags.
	if base, conf := locale.Base(); conf == language.Exact {
		if r, ok := PluralRulesByBase[base]; ok {
			return r
		}
	}
	return PluralRulesByTag[language.Und]
}
