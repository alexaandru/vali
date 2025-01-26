package vali

import (
	"cmp"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

// Possible errors.
var (
	ErrCheckFailed    = errors.New("check failed")
	ErrRequired       = errors.New("value missing")
	ErrInvalidChecker = errors.New("invalid checker")
	ErrInvalidCmp     = errors.New("invalid comparison")
)

// NOTE: It is well covered with tests, the regexp is fine.
var uuid, _ = Regex(`(?i)^[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}$`) //nolint:errcheck // ok

func required(v reflect.Value) (err error) {
	if isZero(v) {
		return ErrRequired
	}

	return
}

// Regex allows you to easily create regex-based checkers.
func Regex(arg string) (c Checker, err error) {
	rx, err := regexp.Compile(arg)
	if err != nil {
		return
	}

	return func(v reflect.Value) (err error) {
		act := fmt.Sprint(v.Interface())
		if rx.MatchString(act) {
			return
		}

		return fmt.Errorf("%q does not match %s", act, arg)
	}, nil
}

// Eq checks numbers for being == `arg` and things with a `len()`
// (`array`, `chan`, `map`, `slice`, `string`) for having len == `arg`.
func Eq(arg string) (c Checker, err error) {
	return sizeCmp(arg, 0)
}

// Ne checks numbers for being != `arg` and things with a `len()`
// (`array`, `chan`, `map`, `slice`, `string`) for having len != `arg`.
func Ne(arg string) (c Checker, err error) {
	return sizeCmp(arg, 2)
}

// Min checks numbers for being at least `arg` and things with a `len()`
// (`array`, `chan`, `map`, `slice`, `string`) for having len at least `arg`.
func Min(arg string) (c Checker, err error) {
	return sizeCmp(arg, 1)
}

// Max checks numbers for being at most `arg` and things with a `len()`
// (`array`, `chan`, `map`, `slice`, `string`) for having len at most `arg`.
func Max(arg string) (c Checker, err error) {
	return sizeCmp(arg, -1)
}

func sizeCmp(arg string, expCmp int) (c Checker, err error) { //nolint:gocognit,funlen // ok
	var label string

	switch expCmp {
	case -1:
		label = "more than"
	case 1:
		label = "less than"
	case 0:
		label = "not equal to"
	case 2:
		label = "equal to"
	}

	return func(v reflect.Value) (err error) {
		defer func() {
			if r := recover(); r != nil {
				if v, ok := r.(error); ok {
					err = v
				} else {
					err = errors.New(fmt.Sprint(r))
				}
			}
		}()

		switch {
		case v.CanInt():
			var x int

			if x, err = strconv.Atoi(arg); err != nil {
				return
			}

			y := int(v.Int())
			if cmp2(y, x, expCmp) {
				return fmt.Errorf("%d is %s %d", y, label, x)
			}
		case v.CanUint():
			var x int

			if x, err = strconv.Atoi(arg); err != nil {
				return
			}

			y := uint(v.Uint())
			if cmp2(y, uint(x), expCmp) { //nolint:gosec // ok
				return fmt.Errorf("%d is %s %d", y, label, x)
			}
		case v.CanFloat():
			var x float64

			switch vv := v.Interface().(type) {
			case float32:
				if x, err = strconv.ParseFloat(arg, 32); err != nil {
					return
				}

				if cmp2(vv, float32(x), expCmp) {
					return fmt.Errorf("%.0f is %s %.0f", vv, label, x)
				}
			case float64:
				if x, err = strconv.ParseFloat(arg, 64); err != nil {
					return
				}

				if cmp2(vv, x, expCmp) {
					return fmt.Errorf("%.0f is %s %.0f", vv, label, x)
				}
			}
		default:
			var x int

			if x, err = strconv.Atoi(arg); err != nil {
				return
			}

			// FIXME: Len can panic!
			y := v.Len()
			if cmp2(y, x, expCmp) {
				return fmt.Errorf("len %d is %s %d", y, label, x)
			}
		}

		return
	}, nil
}

func cmp2[T cmp.Ordered](a, b T, expCmp int) bool {
	switch actCmp := cmp.Compare(a, b); expCmp {
	case -1:
		return actCmp != -1 && actCmp != 0
	case 1:
		return actCmp != 1 && actCmp != 0
	case 0:
		return actCmp != 0
	case 2:
		return actCmp == 0
	default:
		return false
	}
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
