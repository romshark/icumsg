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

https://go.dev/play/p/uNHO3Gt128Z

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

Error handling:

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