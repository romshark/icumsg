package icumsg

import (
	"errors"
	"iter"
	"slices"
	"strings"

	"github.com/romshark/icumsg/cldr"
	"golang.org/x/text/language"
)

type ErrorPluralMissingOption struct {
	Need, Has  cldr.PluralRules
	TokenIndex int

	// When Ordinal == false the plural options are cardinal.
	Ordinal bool
}

func (e ErrorPluralMissingOption) MissingOptions() iter.Seq[string] {
	return func(yield func(string) bool) {
		if e.Need.Zero && !e.Has.Zero {
			if !yield(cldr.RuleZero) {
				return
			}
		}
		if e.Need.One && !e.Has.One {
			if !yield(cldr.RuleOne) {
				return
			}
		}
		if e.Need.Two && !e.Has.Two {
			if !yield(cldr.RuleTwo) {
				return
			}
		}
		if e.Need.Few && !e.Has.Few {
			if !yield(cldr.RuleFew) {
				return
			}
		}
		if e.Need.Many && !e.Has.Many {
			if !yield(cldr.RuleMany) {
				return
			}
		}
		// Other is always needed.
		// An ICU message will fail to parse if other is missing.
	}
}

func (e ErrorPluralMissingOption) Error() string {
	var b strings.Builder
	if e.Ordinal {
		b.WriteString("missing ordinal plural options [")
	} else {
		b.WriteString("missing cardinal plural options [")
	}
	sep := ""
	for o := range e.MissingOptions() {
		b.WriteString(sep)
		b.WriteString(o)
		sep = ","
	}
	b.WriteByte(']')
	return b.String()
}

type ErrorSelectMissingOption struct {
	Need, Has  []string
	TokenIndex int
}

func (e ErrorSelectMissingOption) MissingOptions() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, n := range e.Need {
			if slices.Index(e.Has, n) == -1 {
				if !yield(n) {
					break
				}
			}
		}
	}
}

func (e ErrorSelectMissingOption) Error() string {
	var b strings.Builder
	b.WriteString("missing select options [")
	sep := ""
	for o := range e.MissingOptions() {
		b.WriteString(sep)
		b.WriteString(o)
		sep = ","
	}
	b.WriteByte(']')
	return b.String()
}

type ErrorSelectInvalidOption struct {
	TokenIndexArgument int
	TokenIndexOption   int
}

func (e ErrorSelectInvalidOption) Error() string {
	return "invalid select option"
}

var errStopAnalyze = errors.New("stop")

// Errors returns an iterator over all semantic (non-syntax) errors
// for the given ICU message.
func Errors(
	locale language.Tag, raw string, tokens []Token, selectOptions SelectOptions,
) iter.Seq[error] {
	return func(yield func(error) bool) {
		// No need to check the returned error because it can only ever
		// return errStopAnalyze and nothing else.
		_, _ = Analyze(
			locale, raw, tokens,
			selectOptions,
			func(index int) error {
				// On incomplete.
				tok := tokens[index]
				argNameTok := tokens[index+1]
				argName := argNameTok.String(raw, tokens)
				var err error
				switch tok.Type {
				case TokenTypePlural:
					need, _ := cldr.LocalePluralRules(locale)
					has := findAllPluralOptions(tokens, index)
					err = ErrorPluralMissingOption{
						Need: need, Has: has, TokenIndex: index,
					}

				case TokenTypeSelectOrdinal:
					_, need := cldr.LocalePluralRules(locale)
					has := findAllPluralOptions(tokens, index)
					err = ErrorPluralMissingOption{
						Need: need, Has: has, Ordinal: true, TokenIndex: index,
					}

				case TokenTypeSelect:
					need, _, _ := selectOptions(argName)
					has := findAllSelectOptions(raw, tokens, index)
					err = ErrorSelectMissingOption{
						Need: need, Has: has, TokenIndex: index,
					}
				}
				if err != nil {
					if !yield(err) {
						return errStopAnalyze
					}
				}
				return nil
			},
			func(indexArgument, indexOption int) error {
				if !yield(ErrorSelectInvalidOption{
					TokenIndexOption:   indexOption,
					TokenIndexArgument: indexArgument,
				}) {
					return errStopAnalyze
				}
				return nil
			},
		)
	}
}

func findAllPluralOptions(tokens []Token, index int) (has cldr.PluralRules) {
	for t := range Options(tokens, index) {
		switch tokens[t].Type {
		case TokenTypeOptionZero:
			has.Zero = true
		case TokenTypeOptionOne:
			has.One = true
		case TokenTypeOptionTwo:
			has.Two = true
		case TokenTypeOptionFew:
			has.Few = true
		case TokenTypeOptionMany:
			has.Many = true
		case TokenTypeOptionOther:
			has.Other = true
		}
	}
	return has
}

func findAllSelectOptions(msg string, tokens []Token, index int) (has []string) {
	for t := range Options(tokens, index) {
		switch tokens[t].Type {
		case TokenTypeOptionOther:
			has = append(has, "other")
		case TokenTypeOption:
			has = append(has, tokens[t+1].String(msg, tokens))
		}
	}
	return has
}

// Analyze returns the total number of choices in src.
// onIncomplete is invoked when an incomplete select, plural or selectordinal
// is encountered.
// onRejected is invoked when an unknown select option was encountered.
// selectOptions is invoked when a select is encountered and if it returns
// a non-nil slice then those will be the expected options the presence of which
// will define whether the select is complete (depending on the policies returned).
// An empty non-nil slice expects no options besides "other".
// A nil slice applies no policies.
// selectOptions is not invoked for plural and selectordinal, instead locale is used
// to determine what options are required.
// If onIncomplete or onRejected returns an error it's returned immediately.
func Analyze(
	locale language.Tag,
	src string,
	buffer []Token,
	selectOptions func(argName string) (
		[]string, OptionsPresencePolicy, OptionUnknownPolicy,
	),
	onIncomplete func(index int) error,
	onRejected func(indexArgument, indexOption int) error,
) (totalChoices int, err error) {
	return analyze(0, len(buffer), src, buffer, locale,
		selectOptions, onIncomplete, onRejected)
}

func analyze(
	startIndex, endIndex int,
	src string,
	buffer []Token,
	locale language.Tag,
	selectOptions SelectOptions,
	onIncomplete func(index int) error,
	onRejected func(indexArgument, indexOption int) error,
) (total int, err error) {
	cardinal, ordinal := cldr.LocalePluralRules(locale)

	for i := startIndex; i < endIndex; i++ {
		t := buffer[i]
		switch t.Type {
		case TokenTypeSelect:
			total++
			tn := buffer[i+1]
			opts, presencePolicy, unknownPolicy := selectOptions(tn.String(src, buffer))
			reqCount := len(opts)
			for j := range Options(buffer, i) {
				// A nil opts applies no policies.
				if opts != nil && buffer[j].Type != TokenTypeOptionOther {
					name := buffer[j+1].String(src, buffer)
					if inOpts := slices.Contains(opts, name); inOpts {
						reqCount--
					} else if unknownPolicy == OptionUnknownPolicyReject {
						if err := onRejected(i, j); err != nil {
							return total, err
						}
					}
				}
				n, err := analyze(
					j, buffer[j].IndexEnd,
					src, buffer, locale,
					selectOptions, onIncomplete, onRejected,
				)
				if err != nil {
					return total, err
				}
				total += n
			}
			if presencePolicy == OptionsPresencePolicyRequired && reqCount != 0 {
				if err := onIncomplete(i); err != nil {
					return total, err
				}
			}
			i = t.IndexEnd + 1
		case TokenTypePlural:
			total++
			var rules cldr.PluralRules
			for j := range Options(buffer, i) {
				switch buffer[j].Type {
				case TokenTypeOptionZero:
					rules.Zero = true
				case TokenTypeOptionOne:
					rules.One = true
				case TokenTypeOptionTwo:
					rules.Two = true
				case TokenTypeOptionFew:
					rules.Few = true
				case TokenTypeOptionMany:
					rules.Many = true
				case TokenTypeOptionOther:
					rules.Other = true
				}
				n, err := analyze(
					j, buffer[j].IndexEnd,
					src, buffer, locale,
					selectOptions, onIncomplete, onRejected,
				)
				if err != nil {
					return total, err
				}
				total += n
			}
			if rules != cardinal {
				if err := onIncomplete(i); err != nil {
					return total, err
				}
			}
			i = t.IndexEnd + 1
		case TokenTypeSelectOrdinal:
			total++
			var rules cldr.PluralRules
			for j := range Options(buffer, i) {
				switch buffer[j].Type {
				case TokenTypeOptionZero:
					rules.Zero = true
				case TokenTypeOptionOne:
					rules.One = true
				case TokenTypeOptionTwo:
					rules.Two = true
				case TokenTypeOptionFew:
					rules.Few = true
				case TokenTypeOptionMany:
					rules.Many = true
				case TokenTypeOptionOther:
					rules.Other = true
				}
				n, err := analyze(
					j, buffer[j].IndexEnd,
					src, buffer, locale,
					selectOptions, onIncomplete, onRejected,
				)
				if err != nil {
					return total, err
				}
				total += n
			}
			if rules != ordinal {
				if err := onIncomplete(i); err != nil {
					return total, err
				}
			}
			i = t.IndexEnd + 1
		}
	}
	return total, nil
}
