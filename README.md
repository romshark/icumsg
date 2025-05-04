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
	// The English language only supports the 'one' and 'other' CLDR plural forms.
	msg := `{numMsgs,plural, one{# message} other{# messages} few{this is wrong}}`

	var tokenizer icumsg.Tokenizer
	_, err := tokenizer.Tokenize(language.English, nil, msg)
	if err != nil {
		fmt.Printf("Error at index %d: %v\n", tokenizer.Pos(), err)
	}

	// output:
	// Error at index 50: plural form unsupported for locale
}
```

## ICU Message Completeness

ICU messages can be valid yet icomplete when missing some
`select`, `plural` or `selectordinal` options.
`icumsg.Completeness` allows you to inspect a message in detail,
find unsupported select options and missing plural options.

https://go.dev/play/p/U9t0a0XH9U_h

```go
package main

import (
	"fmt"

	"github.com/romshark/icumsg"
	"golang.org/x/text/language"
)

func main() {
	// varNum is missing option "one"
	// varGender is missing option "male"
	// varGender lists unsupported option "unknown"
	msg := `This message is valid but has incomplete plural and unknown select options:
	{varNum, plural,
		other{
			{varGender, select,
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
	}`

	var tokenizer icumsg.Tokenizer
	tokens, err := tokenizer.Tokenize(language.English, nil, msg)
	if err != nil {
		fmt.Printf("ERR: at index %d: %v\n", tokenizer.Pos(), err)
		os.Exit(1)
	}

	// Option "other" doesn't need to be included because it's always required.
	optionsForVarGender := []string{"male", "female"}

	var incomplete, rejected []string
	totalChoices := icumsg.Completeness(msg, tokens, language.English,
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
		}, func(index int) {
			// This is called when an incomplete choice is encountered.
			tArg, tName := tokens[index], tokens[index+1]
			incomplete = append(incomplete,
				tArg.Type.String()+": "+tName.String(msg, tokens))
		}, func(index int) {
			// This is called when a rejected option is encountered.
			tArg, tName := tokens[index], tokens[index+1]
			rejected = append(rejected,
				tArg.Type.String()+": "+tName.String(msg, tokens))
		})

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
	// totalChoices: 2
	// incomplete (2):
	//  select argument: varGender
	//  plural argument: varNum
	// rejected (1):
	//  option: unknown
	// completeness: 0.00%
}
```
