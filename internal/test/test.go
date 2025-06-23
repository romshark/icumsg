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
		tb.Fatalf("\nexpected: %#v;\nreceived: %#v\n%s", expect, actual, msgStr(msg...))
	}
}

func RequireDeepEqual[T any](tb testing.TB, expect, actual T, msg ...any) {
	tb.Helper()
	if !reflect.DeepEqual(expect, actual) {
		tb.Fatalf("\nexpected: %#v;\nreceived: %#v\n%s", expect, actual, msgStr(msg...))
	}
}

func RequireErrIs(tb testing.TB, expect, actual error, msg ...any) {
	tb.Helper()
	if !errors.Is(actual, expect) {
		tb.Fatalf("\nexpected: %#v;\nreceived: %#v\n%s", expect, actual, msgStr(msg...))
	}
}

func RequireErrType[T any](tb testing.TB, actual error, msg ...any) {
	tb.Helper()
	var zero T
	if !errors.As(actual, &zero) {
		tb.Fatalf("\nexpected: %T;\nreceived: %#v\n%s", zero, actual, msgStr(msg...))
	}
}

func RequireNoErr(tb testing.TB, err error, msg ...any) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("\nexpected: nil;\nreceived: %#v\n%s", err, msgStr(msg...))
	}
}

func msgStr(msg ...any) string {
	if msg == nil {
		return ""
	}
	return fmt.Sprintf(msg[0].(string), msg[1:]...)
}
