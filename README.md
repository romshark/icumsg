<a href="https://pkg.go.dev/github.com/romshark/icumsg">
    <img src="https://godoc.org/github.com/romshark/icumsg?status.svg" alt="GoDoc">
</a>
<a href="https://goreportcard.com/report/github.com/romshark/icumsg">
    <img src="https://goreportcard.com/badge/github.com/romshark/icumsg" alt="GoReportCard">
</a>
<a href='https://coveralls.io/github/romshark/icumsg?branch=main'>
    <img src='https://coveralls.io/repos/github/romshark/icumsg/badge.svg?branch=main&service=github' alt='Coverage Status' />
</a>

# icumsg

This Go module provides an efficient
[ICU Message Format](https://unicode-org.github.io/icu/userguide/format_parse/messages/)
tokenizer.

https://go.dev/play/p/y7OA1YK2Wn4

```go
package main

import (
	"fmt"
	"os"

	"github.com/romshark/icumsg"
	"golang.org/x/text/language"
)

func main() {
	msg := `Hello {arg} ({rank, ordinal})!`

	var tokenizer icumsg.Tokenizer
	tokens, err := tokenizer.Tokenize(language.English, nil, msg)
	if err != nil {
		fmt.Printf("ERR: at index %d: %v\n", tokenizer.Pos(), err)
		os.Exit(1)
	}

	fmt.Printf("token (%d):\n", len(tokens))
	for i, token := range tokens {
		fmt.Printf(" %d (%s): %q\n", i,
			token.Type.String(), token.String(msg, tokens))
	}

	// output:
	// token (8):
	//  0 (literal): "Hello "
	//  1 (simple argument): "{arg}"
	//  2 (argument name): "arg"
	//  3 (literal): " ("
	//  4 (simple argument): "{rank, ordinal}"
	//  5 (argument name): "rank"
	//  6 (argument type ordinal): "ordinal"
	//  7 (literal): ")!"
}
```

## Specification

This module implements **ICU MessageFormat 1.0** for ICU 4.8 and later, the message
syntax of
[`MessageFormat`](https://unicode-org.github.io/icu-docs/apidoc/released/icu4j/com/ibm/icu/text/MessageFormat.html)
in ICU4J and ICU4C, as documented in the
[ICU User Guide](https://unicode-org.github.io/icu/userguide/format_parse/messages/).
Its successor MessageFormat 2.0
([UTS #35 Part 9](https://unicode.org/reports/tr35/tr35-messageFormat.html))
defines a different syntax and is **not** supported.

- Apostrophes follow ICU's default
  [`ApostropheMode.DOUBLE_OPTIONAL`](https://unicode-org.github.io/icu-docs/apidoc/released/icu4j/com/ibm/icu/text/MessagePattern.ApostropheMode.html)
  ("quote only where needed"). A single apostrophe starts quoted text only when it
  immediately precedes a syntax character, which is `{`, `}` or `#` inside a `plural`
  or `selectordinal` option. Everywhere else it's literal text. `He's there` needs no
  escaping, unlike in JDK `MessageFormat`. A pair of apostrophes is one literal
  apostrophe.
- Plural rules are generated from [CLDR](https://cldr.unicode.org/) version 48.
  See [internal/cldr](internal/cldr/cldr_gen.go).

Deviations from ICU:

- `choice` arguments are not supported. ICU deprecated them in favor of `plural` and
  `selectordinal`.
- Unclosed quoted text is rejected with `ErrUnclosedQuote`. ICU instead auto-quotes it
  to the end of the message.

## Error handling

https://go.dev/play/p/NI6gXkcJJcH

```go
package main

import (
	"fmt"

	"github.com/romshark/icumsg"
	"golang.org/x/text/language"
)

func main() {
	// The English language only supports the 'one' and 'other' CLDR plural rules.
	msg := `{numMsgs,plural, one{# message} other{# messages} few{this is wrong}}`

	var tokenizer icumsg.Tokenizer
	_, err := tokenizer.Tokenize(language.English, nil, msg)
	if err != nil {
		fmt.Printf("Error at index %d: %v\n", tokenizer.Pos(), err)
	}

	// output:
	// Error at index 50: plural rule unsupported for locale
}
```

## Semantic Analysis

ICU messages can be syntactically valid yet incomplete when missing `select`, `plural` or
`selectordinal` options required by the locale as well as semantically invalid when
featuring unsupported `select` options.
`icumsg.Analyze` allows you to inspect a message in detail and discover semantic issues.

https://go.dev/play/p/U9t0a0XH9U_h

```go
package main

import (
	"fmt"
	"os"

	"github.com/romshark/icumsg"
	"golang.org/x/text/language"
)

func main() {
	locale := language.English

	// varGender lists unsupported option "unknown"
	msg := `This message is valid but has incomplete plural and unknown select options:
	missing one: {varNum, plural,
		other{
			missing male: {varGender, select,
				unknown{
					varNum[other],varGender[unknown]
				}
				female{
					varNum[other],varGender[female]
				}
				other{
					varNum[other],varGender[other]
				}
			}
		}
	}
	complete: {varNum, plural,
		one{-}
		other{-}
	}`

	var tokenizer icumsg.Tokenizer
	tokens, err := tokenizer.Tokenize(locale, nil, msg)
	if err != nil {
		fmt.Printf("ERR: at index %d: %v\n", tokenizer.Pos(), err)
		os.Exit(1)
	}

	// Option "other" doesn't need to be included because it's always required.
	optionsForVarGender := []string{"male", "female"}

	var incomplete, rejected []string
	totalChoices, err := icumsg.Analyze(locale, msg, tokens,
		func(argName string) (
			options []string,
			policyPresence icumsg.OptionsPresencePolicy,
			policyUnknown icumsg.OptionUnknownPolicy,
		) {
			if argName == "varGender" {
				// Apply these policies and options only for argument "varGender"
				policyPresence = icumsg.OptionsPresencePolicyRequired
				policyUnknown = icumsg.OptionUnknownPolicyReject
				return optionsForVarGender, policyPresence, policyUnknown
			}
			return nil, 0, 0
		}, func(index int) error {
			// This is called when an incomplete choice is encountered.
			tArg, tName := tokens[index], tokens[index+1]
			incomplete = append(incomplete,
				tArg.Type.String()+": "+tName.String(msg, tokens))
			return nil
		}, func(indexArgument, indexOption int) error {
			// This is called when a rejected option is encountered.
			tArg, tName := tokens[indexArgument+1], tokens[indexOption+1]
			rejected = append(rejected, fmt.Sprintf("%q: option %q",
				tArg.String(msg, tokens), tName.String(msg, tokens)))
			return nil
		})
	if err != nil {
		panic(err)
	}

	fmt.Printf("totalChoices: %d\n", totalChoices)
	fmt.Printf("incomplete (%d):\n", len(incomplete))
	for _, s := range incomplete {
		fmt.Printf(" %s\n", s)
	}
	fmt.Printf("rejected (%d):\n", len(rejected))
	for _, s := range rejected {
		fmt.Printf(" %s\n", s)
	}

	{
		total := float64(totalChoices)
		incomplete := float64(len(incomplete))
		complete := total - incomplete
		percent := complete / total
		fmt.Printf("completeness: %.2f%%\n", percent*100)
	}

	// output:
	// totalChoices: 3
	// incomplete (2):
	//  select argument: varGender
	//  plural argument: varNum
	// rejected (1):
	//  "varGender": option "unknown"
	// completeness: 33.33%
}
```
