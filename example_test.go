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
	// Error at index 50: plural form unsupported for locale
}
