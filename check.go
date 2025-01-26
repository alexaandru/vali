package vali

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
)

// Possible errors.
var (
	ErrCheckFailed    = errors.New("check failed")
	ErrNotAStruct     = errors.New("not a struct")
	ErrRequired       = errors.New("value missing")
	ErrInvalidChecker = errors.New("invalid checker")
)

var uuid Checker

func required(v reflect.Value) (err error) {
	if isZero(v) {
		return ErrRequired
	}

	return
}

// Regex allows you to easily create regex-based checkers.
func Regex(args string) (c Checker, err error) {
	rx, err := regexp.Compile(args)
	if err != nil {
		return
	}

	c = func(v reflect.Value) (err error) {
		if isZero(v) {
			return
		}

		act := fmt.Sprint(v.Interface())
		if rx.MatchString(act) {
			return
		}

		return fmt.Errorf("%q does not match %s", act, args)
	}

	return
}

func oneOf(args string) (Checker, error) {
	return Regex(fmt.Sprintf("^(%s)$", args))
}

// TODO: When this is closed, remove this:
// https://github.com/golang/go/issues/51649
func isZero(v reflect.Value) (ok bool) {
	defer func() {
		if x := recover(); x != nil {
			ok = true
		}
	}()

	return v.IsZero()
}

func init() {
	// NOTE: It is well covered with tests, the regexp is fine.
	// If I put an `if` here, it can never get covered :-).
	uuid, _ = Regex(`(?i)^[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}$`) //nolint:errcheck // ok
	DefaultValidator = NewValidator("validate")
}
