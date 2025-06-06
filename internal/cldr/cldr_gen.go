// Generated by github.com/romshark/icumsg/internal/cmd/gencldr. DO NOT EDIT.
// CLDR Version: 47
// Unicode Version: 16.0.0

package cldr

import "golang.org/x/text/language"

// Forms defines supported CLDR plural forms.
type Forms struct{ Zero, One, Two, Few, Many, Other bool }

// PluralForms defines supported cardinal and ordinal CLDR plural forms.
type PluralForms struct {
	Cardinal Forms
	Ordinal  Forms
}

// PluralFormsByTag maps language tags to supported plural forms.
var PluralFormsByTag = make(map[language.Tag]PluralForms, 219)

// PluralFormsByBase maps base languages to supported plural forms.
var PluralFormsByBase = make(map[language.Base]PluralForms, 219)

func init() {
	{
		PluralFormsByTag[language.Und] = PluralForms{
			Cardinal: Forms{Other: true}, Ordinal: Forms{Other: true},
		}
		undBase, _ := language.Und.Base()
		PluralFormsByBase[undBase] = PluralForms{
			Cardinal: Forms{Other: true}, Ordinal: Forms{Other: true},
		}
	}
	register := func(s string, cardinal, ordinal Forms, isBase bool) {
		l, err := language.Parse(s)
		if err != nil {
			panic(err)
		}
		PluralFormsByTag[l] = PluralForms{cardinal, ordinal}
		if isBase {
			base, _ := l.Base()
			PluralFormsByBase[base] = PluralForms{cardinal, ordinal}
		}
	}
	register("af",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("ak",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("am",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("an",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("ar",
		Forms{Other: true, Zero: true, One: true, Two: true, Few: true, Many: true},
		Forms{Other: true}, true)
	register("ars",
		Forms{Other: true, Zero: true, One: true, Two: true, Few: true, Many: true},
		Forms{Other: true}, true)
	register("as",
		Forms{Other: true, One: true},
		Forms{Other: true, One: true, Two: true, Few: true, Many: true}, true)
	register("asa",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("ast",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("az",
		Forms{Other: true, One: true},
		Forms{Other: true, One: true, Few: true, Many: true}, true)
	register("bal",
		Forms{Other: true, One: true},
		Forms{Other: true, One: true}, true)
	register("be",
		Forms{Other: true, One: true, Few: true, Many: true},
		Forms{Other: true, Few: true}, true)
	register("bem",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("bez",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("bg",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("bho",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("blo",
		Forms{Other: true, Zero: true, One: true},
		Forms{Other: true, Zero: true, One: true, Few: true}, true)
	register("bm",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("bn",
		Forms{Other: true, One: true},
		Forms{Other: true, One: true, Two: true, Few: true, Many: true}, true)
	register("bo",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("br",
		Forms{Other: true, One: true, Two: true, Few: true, Many: true},
		Forms{Other: true}, true)
	register("brx",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("bs",
		Forms{Other: true, One: true, Few: true},
		Forms{Other: true}, true)
	register("ca",
		Forms{Other: true, One: true, Many: true},
		Forms{Other: true, One: true, Two: true, Few: true}, true)
	register("ce",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("ceb",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("cgg",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("chr",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("ckb",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("cs",
		Forms{Other: true, One: true, Few: true, Many: true},
		Forms{Other: true}, true)
	register("csw",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("cy",
		Forms{Other: true, Zero: true, One: true, Two: true, Few: true, Many: true},
		Forms{Other: true, Zero: true, One: true, Two: true, Few: true, Many: true}, true)
	register("da",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("de",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("doi",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("dsb",
		Forms{Other: true, One: true, Two: true, Few: true},
		Forms{Other: true}, true)
	register("dv",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("dz",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("ee",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("el",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("en",
		Forms{Other: true, One: true},
		Forms{Other: true, One: true, Two: true, Few: true}, true)
	register("eo",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("es",
		Forms{Other: true, One: true, Many: true},
		Forms{Other: true}, true)
	register("et",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("eu",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("fa",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("ff",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("fi",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("fil",
		Forms{Other: true, One: true},
		Forms{Other: true, One: true}, true)
	register("fo",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("fr",
		Forms{Other: true, One: true, Many: true},
		Forms{Other: true, One: true}, true)
	register("fur",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("fy",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("ga",
		Forms{Other: true, One: true, Two: true, Few: true, Many: true},
		Forms{Other: true, One: true}, true)
	register("gd",
		Forms{Other: true, One: true, Two: true, Few: true},
		Forms{Other: true, One: true, Two: true, Few: true}, true)
	register("gl",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("gsw",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("gu",
		Forms{Other: true, One: true},
		Forms{Other: true, One: true, Two: true, Few: true, Many: true}, true)
	register("guw",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("gv",
		Forms{Other: true, One: true, Two: true, Few: true, Many: true},
		Forms{Other: true}, true)
	register("ha",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("haw",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("he",
		Forms{Other: true, One: true, Two: true},
		Forms{Other: true}, true)
	register("hi",
		Forms{Other: true, One: true},
		Forms{Other: true, One: true, Two: true, Few: true, Many: true}, true)
	register("hnj",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("hr",
		Forms{Other: true, One: true, Few: true},
		Forms{Other: true}, true)
	register("hsb",
		Forms{Other: true, One: true, Two: true, Few: true},
		Forms{Other: true}, true)
	register("hu",
		Forms{Other: true, One: true},
		Forms{Other: true, One: true}, true)
	register("hy",
		Forms{Other: true, One: true},
		Forms{Other: true, One: true}, true)
	register("ia",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("id",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("ig",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("ii",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("io",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("is",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("it",
		Forms{Other: true, One: true, Many: true},
		Forms{Other: true, Many: true}, true)
	register("iu",
		Forms{Other: true, One: true, Two: true},
		Forms{Other: true}, true)
	register("ja",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("jbo",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("jgo",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("jmc",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("jv",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("jw",
		Forms{Other: true},
		Forms{Other: true}, false)
	register("ka",
		Forms{Other: true, One: true},
		Forms{Other: true, One: true, Many: true}, true)
	register("kab",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("kaj",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("kcg",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("kde",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("kea",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("kk",
		Forms{Other: true, One: true},
		Forms{Other: true, Many: true}, true)
	register("kkj",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("kl",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("km",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("kn",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("ko",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("ks",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("ksb",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("ksh",
		Forms{Other: true, Zero: true, One: true},
		Forms{Other: true}, true)
	register("ku",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("kw",
		Forms{Other: true, Zero: true, One: true, Two: true, Few: true, Many: true},
		Forms{Other: true, One: true, Many: true}, true)
	register("ky",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("lag",
		Forms{Other: true, Zero: true, One: true},
		Forms{Other: true}, true)
	register("lb",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("lg",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("lij",
		Forms{Other: true, One: true},
		Forms{Other: true, Many: true}, true)
	register("lkt",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("lld",
		Forms{Other: true, One: true, Many: true},
		Forms{Other: true, Many: true}, true)
	register("ln",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("lo",
		Forms{Other: true},
		Forms{Other: true, One: true}, true)
	register("lt",
		Forms{Other: true, One: true, Few: true, Many: true},
		Forms{Other: true}, true)
	register("lv",
		Forms{Other: true, Zero: true, One: true},
		Forms{Other: true}, true)
	register("mas",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("mg",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("mgo",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("mk",
		Forms{Other: true, One: true},
		Forms{Other: true, One: true, Two: true, Many: true}, true)
	register("ml",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("mn",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("mo",
		Forms{Other: true, One: true, Few: true},
		Forms{Other: true, One: true}, false)
	register("mr",
		Forms{Other: true, One: true},
		Forms{Other: true, One: true, Two: true, Few: true}, true)
	register("ms",
		Forms{Other: true},
		Forms{Other: true, One: true}, true)
	register("mt",
		Forms{Other: true, One: true, Two: true, Few: true, Many: true},
		Forms{Other: true}, true)
	register("my",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("nah",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("naq",
		Forms{Other: true, One: true, Two: true},
		Forms{Other: true}, true)
	register("nb",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("nd",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("ne",
		Forms{Other: true, One: true},
		Forms{Other: true, One: true}, true)
	register("nl",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("nn",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("nnh",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("no",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("nqo",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("nr",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("nso",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("ny",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("nyn",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("om",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("or",
		Forms{Other: true, One: true},
		Forms{Other: true, One: true, Two: true, Few: true, Many: true}, true)
	register("os",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("osa",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("pa",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("pap",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("pcm",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("pl",
		Forms{Other: true, One: true, Few: true, Many: true},
		Forms{Other: true}, true)
	register("prg",
		Forms{Other: true, Zero: true, One: true},
		Forms{Other: true}, true)
	register("ps",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("pt",
		Forms{Other: true, One: true, Many: true},
		Forms{Other: true}, true)
	register("pt-PT",
		Forms{Other: true, One: true, Many: true},
		Forms{Other: true}, false)
	register("rm",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("ro",
		Forms{Other: true, One: true, Few: true},
		Forms{Other: true, One: true}, true)
	register("rof",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("ru",
		Forms{Other: true, One: true, Few: true, Many: true},
		Forms{Other: true}, true)
	register("rwk",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("sah",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("saq",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("sat",
		Forms{Other: true, One: true, Two: true},
		Forms{Other: true}, true)
	register("sc",
		Forms{Other: true, One: true},
		Forms{Other: true, Many: true}, true)
	register("scn",
		Forms{Other: true, One: true, Many: true},
		Forms{Other: true, Many: true}, true)
	register("sd",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("sdh",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("se",
		Forms{Other: true, One: true, Two: true},
		Forms{Other: true}, true)
	register("seh",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("ses",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("sg",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("sh",
		Forms{Other: true, One: true, Few: true},
		Forms{Other: true}, false)
	register("shi",
		Forms{Other: true, One: true, Few: true},
		Forms{Other: true}, true)
	register("si",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("sk",
		Forms{Other: true, One: true, Few: true, Many: true},
		Forms{Other: true}, true)
	register("sl",
		Forms{Other: true, One: true, Two: true, Few: true},
		Forms{Other: true}, true)
	register("sma",
		Forms{Other: true, One: true, Two: true},
		Forms{Other: true}, true)
	register("smi",
		Forms{Other: true, One: true, Two: true},
		Forms{Other: true}, true)
	register("smj",
		Forms{Other: true, One: true, Two: true},
		Forms{Other: true}, true)
	register("smn",
		Forms{Other: true, One: true, Two: true},
		Forms{Other: true}, true)
	register("sms",
		Forms{Other: true, One: true, Two: true},
		Forms{Other: true}, true)
	register("sn",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("so",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("sq",
		Forms{Other: true, One: true},
		Forms{Other: true, One: true, Many: true}, true)
	register("sr",
		Forms{Other: true, One: true, Few: true},
		Forms{Other: true}, true)
	register("ss",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("ssy",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("st",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("su",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("sv",
		Forms{Other: true, One: true},
		Forms{Other: true, One: true}, true)
	register("sw",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("syr",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("ta",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("te",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("teo",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("th",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("ti",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("tig",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("tk",
		Forms{Other: true, One: true},
		Forms{Other: true, Few: true}, true)
	register("tl",
		Forms{Other: true, One: true},
		Forms{Other: true, One: true}, false)
	register("tn",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("to",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("tpi",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("tr",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("ts",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("tzm",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("ug",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("uk",
		Forms{Other: true, One: true, Few: true, Many: true},
		Forms{Other: true, Few: true}, true)
	register("ur",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("uz",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("ve",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("vec",
		Forms{Other: true, One: true, Many: true},
		Forms{Other: true, Many: true}, true)
	register("vi",
		Forms{Other: true},
		Forms{Other: true, One: true}, true)
	register("vo",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("vun",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("wa",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("wae",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("wo",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("xh",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("xog",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("yi",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
	register("yo",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("yue",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("zh",
		Forms{Other: true},
		Forms{Other: true}, true)
	register("zu",
		Forms{Other: true, One: true},
		Forms{Other: true}, true)
}
