package icumsg

import (
	"errors"
	"iter"
	"slices"
	"strings"

	"github.com/romshark/icumsg/cldr"
	"golang.org/x/text/language"
)

// ErrorPluralMissingOption reports missing locale plural categories.
// Need lists the locale's categories, Has the argument's, and TokenIndex the
// argument's token index. Ordinal distinguishes selectordinal from plural.
type ErrorPluralMissingOption struct {
	Need, Has  cldr.PluralRules
	TokenIndex int

	// False means cardinal.
	Ordinal bool
}

// MissingOptions yields the required categories absent from the argument.
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
		// Tokenize already requires other.
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

// ErrorSelectMissingOption reports options required by [SelectOptions] but
// absent from a select. Need lists required options, Has present options,
// and TokenIndex the argument's token index.
type ErrorSelectMissingOption struct {
	Need, Has  []string
	TokenIndex int
}

// MissingOptions yields each absent required option once.
func (e ErrorSelectMissingOption) MissingOptions() iter.Seq[string] {
	return func(yield func(string) bool) {
		for i, n := range e.Need {
			// Ignore duplicates in Need.
			if slices.Index(e.Has, n) == -1 && slices.Index(e.Need[:i], n) == -1 {
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

// ErrorSelectInvalidOption reports an unlisted option rejected by [SelectOptions].
// Its fields are token indexes.
type ErrorSelectInvalidOption struct {
	TokenIndexArgument int
	TokenIndexOption   int
}

func (e ErrorSelectInvalidOption) Error() string {
	return "invalid select option"
}

var errStopAnalyze = errors.New("stop")

// Errors yields the message's semantic errors: [ErrorPluralMissingOption],
// [ErrorSelectMissingOption], and [ErrorSelectInvalidOption].
func Errors(
	locale language.Tag, raw string, tokens []Token, selectOptions SelectOptions,
) iter.Seq[error] {
	return func(yield func(error) bool) {
		// Analyze can return only errStopAnalyze here.
		_, _ = Analyze(
			locale, raw, tokens,
			selectOptions,
			func(index int) error {
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
		has = append(has, selectOptionName(msg, tokens, t))
	}
	return has
}

// hasSelectOptions reports whether the select contains every name.
func hasSelectOptions(msg string, tokens []Token, index int, names []string) bool {
NAMES:
	for _, n := range names {
		for t := range Options(tokens, index) {
			if selectOptionName(msg, tokens, t) == n {
				continue NAMES
			}
		}
		return false
	}
	return true
}

// selectOptionName returns the name of the select option at tokens[index].
func selectOptionName(msg string, tokens []Token, index int) string {
	if tokens[index].Type == TokenTypeOptionOther {
		return "other"
	}
	return tokens[index+1].String(msg, tokens)
}

// Analyze returns the number of select, plural, and selectordinal arguments in src,
// whose tokens are in buffer. locale defines plural requirements and selectOptions
// defines select policies. onIncomplete receives each incomplete argument index;
// onRejected receives the select and rejected option indexes.
// Callback errors stop analysis.
// selectOptions, onIncomplete and onRejected must not be nil.
func Analyze(
	locale language.Tag,
	src string,
	buffer []Token,
	selectOptions SelectOptions,
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

	// Never skip backward across separately tokenized messages.
	for i := startIndex; i < endIndex; i++ {
		t := buffer[i]
		switch t.Type {
		case TokenTypeSelect:
			total++
			tn := buffer[i+1]
			opts, presencePolicy, unknownPolicy := selectOptions(tn.String(src, buffer))
			for j := range Options(buffer, i) {
				// Nil opts disable policies.
				if opts != nil && unknownPolicy == OptionUnknownPolicyReject &&
					buffer[j].Type != TokenTypeOptionOther &&
					!slices.Contains(opts, buffer[j+1].String(src, buffer)) {
					if err := onRejected(i, j); err != nil {
						return total, err
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
			if presencePolicy == OptionsPresencePolicyRequired &&
				!hasSelectOptions(src, buffer, i, opts) {
				if err := onIncomplete(i); err != nil {
					return total, err
				}
			}
			i = max(i, t.IndexEnd) // Skip contents.
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
			i = max(i, t.IndexEnd) // Skip contents.
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
			i = max(i, t.IndexEnd) // Skip contents.
		}
	}
	return total, nil
}
