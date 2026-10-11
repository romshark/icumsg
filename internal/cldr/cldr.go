package cldr

import "golang.org/x/text/language"

// Lookup returns locale's plural rules. Missing rules fall back to its explicit
// language, then to [language.Und], which has only "other". Tags without a
// language use the root. CLDR finds a [parent locale] by removing trailing subtags;
// it doesn't infer a language.
//
// [parent locale]: https://www.unicode.org/reports/tr35/#Parent_Locales
func Lookup(locale language.Tag) PluralRules {
	if r, ok := PluralRulesByTag[locale]; ok {
		return r
	}
	// Base would infer a language from likely subtags.
	if base, conf := locale.Base(); conf == language.Exact {
		if r, ok := PluralRulesByBase[base]; ok {
			return r
		}
	}
	return PluralRulesByTag[language.Und]
}
