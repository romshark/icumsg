package icumsg_test

import (
	"fmt"
	"os"

	"github.com/romshark/icumsg"
	"golang.org/x/text/language"
)

func ExampleTokenizer() {
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

func ExampleTokenizer_error() {
	msg := `{numMsgs,plural, one{# message} other{# messages} few{this is wrong}}`

	var tokenizer icumsg.Tokenizer
	_, err := tokenizer.Tokenize(language.English, nil, msg)
	if err != nil {
		fmt.Printf("Error at index %d: %v\n", tokenizer.Pos(), err)
	}

	// output:
	// Error at index 50: plural rule unsupported for locale
}

func ExampleErrors() {
	// varGender lists unsupported option "unknown"
	locale := language.Ukrainian
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
	}`

	var tokenizer icumsg.Tokenizer
	tokens, err := tokenizer.Tokenize(locale, nil, msg)
	if err != nil {
		fmt.Printf("ERR: at index %d: %v\n", tokenizer.Pos(), err)
		os.Exit(1)
	}

	// Option "other" doesn't need to be included because it's always required.
	optionsForVarGender := []string{"male", "female"}

	errSeq := icumsg.Errors(locale, msg, tokens,
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
		})

	for err := range errSeq {
		switch e := err.(type) {
		case icumsg.ErrorSelectMissingOption:
			varName := tokens[e.TokenIndex+1].String(msg, tokens)
			fmt.Printf("ERR at %s: %v\n", varName, err.Error())
		case icumsg.ErrorPluralMissingOption:
			varName := tokens[e.TokenIndex+1].String(msg, tokens)
			fmt.Printf("ERR at %s: %v\n", varName, err.Error())
		case icumsg.ErrorSelectInvalidOption:
			varName := tokens[e.TokenIndexArgument+1].String(msg, tokens)
			optName := tokens[e.TokenIndexOption+1].String(msg, tokens)
			fmt.Printf("ERR at %s (option %q): %v\n", varName, optName, err.Error())
		}
	}

	// output:
	// ERR at varGender (option "unknown"): invalid select option
	// ERR at varGender: missing select options [male]
	// ERR at varNum: missing cardinal plural options [one,few,many]
}

func ExampleAnalyze() {
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
