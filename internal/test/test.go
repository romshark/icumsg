package test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func RequireEqual[T comparable](tb testing.TB, expect, actual T, msg ...any) {
	tb.Helper()
	if expect != actual {
		m := ""
		if msg != nil {
			m = "\n" + fmt.Sprintf(msg[0].(string), msg[1:]...)
		}
		tb.Fatalf("\nexpected: %#v;\nreceived: %#v%s", expect, actual, m)
	}
}

func RequireDeepEqual[T any](tb testing.TB, expect, actual T, msg ...any) {
	tb.Helper()
	if !reflect.DeepEqual(expect, actual) {
		m := ""
		if msg != nil {
			m = "\n" + fmt.Sprintf(msg[0].(string), msg[1:]...)
		}
		tb.Fatalf("\nexpected: %#v;\nreceived: %#v%s", expect, actual, m)
	}
}

func RequireErrIs(t *testing.T, expect, actual error, msg ...any) {
	t.Helper()
	if !errors.Is(actual, expect) {
		m := ""
		if msg != nil {
			m = fmt.Sprintf(msg[0].(string), msg[1:]...)
		}
		t.Fatalf("\nexpected: %#v;\nreceived: %#v%s", expect, actual, m)
	}
}

func RequireNoErr(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("\nexpected: no error;\nreceived: %#v", err)
	}
}
