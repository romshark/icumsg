package cldr_test

import (
	"testing"

	"github.com/romshark/icumsg/cldr"
	"github.com/romshark/icumsg/internal/test"
	"golang.org/x/text/language"
)

func TestLocalePluralRules(t *testing.T) {
	f := func(
		t *testing.T, locale language.Tag, expectCardinal, expectOrdinal cldr.PluralRules,
	) {
		t.Helper()
		cardinal, ordinal := cldr.LocalePluralRules(locale)
		test.RequireEqual(t, expectCardinal, cardinal)
		test.RequireEqual(t, expectOrdinal, ordinal)
	}

	f(t, language.English,
		cldr.PluralRules{One: true, Other: true},
		cldr.PluralRules{One: true, Two: true, Few: true, Other: true},
	)
	f(t, language.BritishEnglish,
		cldr.PluralRules{One: true, Other: true},
		cldr.PluralRules{One: true, Two: true, Few: true, Other: true},
	)
	f(t, language.AmericanEnglish,
		cldr.PluralRules{One: true, Other: true},
		cldr.PluralRules{One: true, Two: true, Few: true, Other: true},
	)
	f(t, language.Ukrainian,
		cldr.PluralRules{One: true, Few: true, Many: true, Other: true},
		cldr.PluralRules{Few: true, Other: true},
	)
	f(t, language.Russian,
		cldr.PluralRules{One: true, Few: true, Many: true, Other: true},
		cldr.PluralRules{Other: true},
	)
	f(t, language.German,
		cldr.PluralRules{One: true, Other: true},
		cldr.PluralRules{Other: true},
	)
	f(t, language.Portuguese,
		cldr.PluralRules{One: true, Many: true, Other: true},
		// Expect other even though CLDR doesn't define ordinal for Portuguese.
		cldr.PluralRules{Other: true})
	f(t, language.French,
		cldr.PluralRules{One: true, Many: true, Other: true},
		cldr.PluralRules{One: true, Other: true},
	)
	// Haitian French falls back to "fr".
	f(t, language.MustParse("fr-HT"),
		cldr.PluralRules{One: true, Many: true, Other: true},
		cldr.PluralRules{One: true, Other: true},
	)
	f(t, language.Arabic,
		cldr.PluralRules{
			Zero: true, One: true, Two: true, Few: true, Many: true, Other: true,
		},
		cldr.PluralRules{Other: true},
	)
	f(t, language.Japanese,
		cldr.PluralRules{Other: true},
		cldr.PluralRules{Other: true},
	)
	f(t, language.Chinese,
		cldr.PluralRules{Other: true},
		cldr.PluralRules{Other: true},
	)
	// Chinese simplified script (China) falls back to "zh".
	f(t, language.MustParse("zh-Hans-CN"),
		cldr.PluralRules{Other: true},
		cldr.PluralRules{Other: true},
	)
}
