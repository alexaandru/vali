package vali

import (
	"cmp"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
)

type testCase struct {
	v      any
	s      *Validator
	tag    string
	exp    string
	expErr error
}

type (
	t1 struct {
		Foo string
		Bar string
	}

	t2 struct {
		t1 `validate:"required"`
	}
)

type foo []byte

var _uuid = "550e8400-e29b-41d4-a716-446655440000"

func (f foo) String() string {
	return string(f)
}

func TestNew(t *testing.T) {
	t.Skip("tested implicitly")
}

func TestRegisterChecker(t *testing.T) {
	t.Parallel()

	x := struct {
		Foo foo `validate:"rgb"`
	}{
		Foo: []byte("pink"),
	}

	rgb := func(v reflect.Value) error {
		if !slices.Contains([]string{"red", "green", "blue"}, fmt.Sprint(v.Interface())) {
			return errors.New("must be red, green or blue")
		}

		return nil
	}

	RegisterChecker("rgb", rgb)

	err := Validate(x)
	if !errors.Is(err, ErrCheckFailed) {
		t.Fatalf("Expected %v got %v", ErrCheckFailed, err)
	}

	exp := "Foo: rgb check failed: must be red, green or blue"
	if act := err.Error(); act != exp {
		t.Fatalf("Expected %q got %q", exp, act)
	}
}

func TestValidatorRegisterChecker(t *testing.T) {
	t.Skip("tested implicitly")
}

func TestRegisterCheckerMaker(t *testing.T) {
	x := struct {
		Foo foo `validate:"one_of3:foo|bar|baz"`
	}{
		Foo: []byte("foobar"),
	}

	_oneOf := func(args string) (c Checker, err error) {
		vals := strings.Split(args, "|")
		if len(vals) == 0 {
			return nil, errors.New("must pass at least one value")
		}

		c = func(v reflect.Value) (err error) {
			act := fmt.Sprint(v.Interface())
			if slices.Contains(vals, act) {
				return
			}

			return fmt.Errorf("%q is not one of %v", act, vals)
		}

		return
	}

	RegisterCheckerMaker("one_of3", _oneOf)

	err := Validate(x)
	if !errors.Is(err, ErrCheckFailed) {
		t.Fatalf("Expected %v got %v", ErrCheckFailed, err)
	}

	exp := `Foo: one_of3 check failed: "foobar" is not one of [foo bar baz]`
	if act := err.Error(); act != exp {
		t.Fatalf("Expected %q got %q", exp, act)
	}
}

func TestValidatorRegisterCheckerMaker(t *testing.T) {
	t.Skip("tested implicitly")
}

func TestValidate(t *testing.T) { //nolint:funlen // ok
	t.Parallel()

	testCases := []testCase{
		{struct{}{}, &Validator{}, "", "", nil},
		{struct{}{}, New(""), "", "", nil},
		{struct{}{}, nil, "", "", nil},

		{false, nil, "required", "required check failed: value missing", ErrCheckFailed},
		{"", nil, "required", "required check failed: value missing", ErrCheckFailed},
		{0, nil, "required", "required check failed: value missing", ErrCheckFailed},
		{(func())(nil), nil, "required", "required check failed: value missing", ErrCheckFailed},
		{"123", nil, "required,uuid", `uuid check failed: "123" does not match (?i)^[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}$`, ErrCheckFailed},

		{true, nil, "required", "", nil},
		{"foo", nil, "required", "", nil},
		{1, nil, "required", "", nil},
		{func() {}, nil, "required", "", nil},
		{_uuid, nil, "required,uuid", "", nil},

		{t1{}, nil, "", "", nil},
		{t2{}, nil, "", "", nil},
		{t1{}, nil, "required", "required check failed: value missing", ErrCheckFailed},
		{t2{}, nil, "required", "required check failed: value missing", ErrCheckFailed},
		{t1{Foo: "foobar"}, nil, "required", "", nil},
		{&t2{t1: t1{Foo: "foobar"}}, nil, "", "", nil},

		{struct {
			Foo string
			Bar int
		}{}, nil, "", "", nil},
		{struct {
			Foo string `validate:""`
			Bar int
		}{}, nil, "", "", nil},
		{struct {
			foo string `validate:"required,uuid"`
			bar int
		}{foo: "foo"}, nil, "", "", nil},
		{struct {
			Foo string `validate:"                                "`
			Bar int
		}{}, nil, "", "", nil},
		{struct {
			Foo string `json:",omitempty" validate:"required"`
			Bar int
		}{}, nil, "", "Foo: required check failed: value missing", ErrRequired},
		{struct {
			Foo string `json:",omitempty" validate:"  required   "`
			Bar int
		}{}, nil, "", "Foo: required check failed: value missing", ErrRequired},
		{struct {
			Foo string `json:",omitempty" validate:"    bogus   ,   required   "`
			Bar int
		}{}, nil, "", "Foo: invalid checker: bogus", ErrInvalidChecker},
		{struct {
			Foo string `json:",omitempty" validate:"    required,   bogus          "`
			Bar int
		}{Foo: "foo"}, nil, "", "Foo: invalid checker: bogus", ErrInvalidChecker},
		{struct {
			Foo *string `json:",omitempty" validate:"required"`
			Bar int
		}{Foo: p("")}, nil, "", "Foo: required check failed: value missing", ErrRequired},
		{struct {
			Foo ***string `json:",omitempty" validate:"required"`
			Bar int
		}{Foo: p(p(p("")))}, nil, "", "Foo: required check failed: value missing", ErrRequired},
		{struct {
			Foo *string `json:",omitempty" validate:"required"`
			Bar int
		}{}, nil, "", "Foo: required check failed: value missing", ErrRequired},
		{struct {
			Foo *string `json:",omitempty" validate:"uuid"`
			Bar int
		}{}, nil, "", "", nil},
		{struct {
			Foo *string `json:",omitempty" validate:"uuid,required"`
			Bar int
		}{}, nil, "", "Foo: required check failed: value missing", ErrRequired},
		{struct {
			Foo *string `json:",omitempty" validate:"required,uuid"`
			Bar int
		}{}, nil, "", "Foo: required check failed: value missing", ErrRequired},
		{struct {
			Foo *string `json:",omitempty" validate:"required,uuid"`
			Bar int
		}{Foo: p("foo")}, nil, "", `Foo: uuid check failed: "foo" does not match (?i)^[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}$`, ErrCheckFailed},
		{struct {
			Foo    string `json:",omitempty" validate:"required"`
			Bar    string `json:",omitempty" validate:"required"`
			Baz    string `json:",omitempty" validate:"required"`
			Foobar int
		}{Foo: "foo", Bar: "bar"}, nil, "", "Baz: required check failed: value missing", ErrRequired},
		{struct {
			Foo string `json:",omitempty" validate:"required,,,,uuid"`
			Bar int
		}{Foo: _uuid}, nil, "", "", nil},
		{struct {
			Foo string `json:",omitempty" validate:"required,:"`
			Bar int
		}{Foo: _uuid}, nil, "", "Foo: invalid checker: :", ErrInvalidChecker},
		{struct {
			Foo string `json:",omitempty" validate:"required,foo:"`
			Bar int
		}{Foo: _uuid}, nil, "", "Foo: invalid checker: foo:", ErrInvalidChecker},
		{struct {
			Foo string `json:",omitempty" validate:"required,:foo"`
			Bar int
		}{Foo: _uuid}, nil, "", "Foo: invalid checker: :foo", ErrInvalidChecker},
		{struct {
			Foo string `json:",omitempty" validate:"required,foo:bar"`
			Bar int
		}{Foo: _uuid}, nil, "", "Foo: invalid checker: foo:bar", ErrInvalidChecker},
		{struct {
			Foo string `json:",omitempty" validate:"required,regex:[A-"`
			Bar int
		}{Foo: _uuid}, nil, "", "Foo: invalid checker: regex:[A-: error parsing regexp: missing closing ]: `[A-`", ErrInvalidChecker},
		{
			struct {
				Foo    string `json:",omitempty" validate:"required"`
				Bar    string `json:",omitempty" validate:"required"`
				Baz    string `json:",omitempty" validate:"required,regex:^(foo|bar|baz)$"`
				Foobar int
			}{Foo: "foo", Bar: "bar", Baz: "other"},
			nil, "",
			`Baz: regex check failed: "other" does not match ^(foo|bar|baz)$`, ErrCheckFailed,
		},
		{
			struct {
				Foo    string `json:",omitempty" validate:"regex:^(foo|bar|baz)$"`
				Bar    string `json:",omitempty" validate:"regex:^(foo|bar|baz)$"`
				Baz    string `json:",omitempty" validate:"regex:^(foo|bar|baz)$"`
				Foobar int
			}{Foo: "foo", Bar: "", Baz: "baz"},
			nil, "", "", nil,
		},
		{
			struct {
				Foo    string `json:",omitempty" validate:"one_of:foo|bar|baz"`
				Bar    string `json:",omitempty" validate:"one_of:foo|bar|baz"`
				Baz    string `json:",omitempty" validate:"one_of:foo|bar|baz"`
				Foobar int
			}{Foo: "foo", Bar: "Bar", Baz: "baz"},
			nil, "", `Bar: one_of check failed: "Bar" does not match ^(foo|bar|baz)$`, ErrCheckFailed,
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			err := Validate(tc.v, tc.tag)
			if !errors.Is(err, tc.expErr) {
				t.Fatalf("Expected %v got %v", tc.expErr, err)
			}

			if err == nil {
				return
			}

			exp := cmp.Or(tc.exp, tc.expErr.Error())
			if act := err.Error(); err.Error() != exp {
				t.Fatalf("Expected %q got %q", exp, act)
			}
		})
	}
}

func TestValidatorValidate(t *testing.T) {
	t.Skip("tested implicitly")
}

func TestValidatorConfigurableSeparators(t *testing.T) {
	x := struct {
		Foo string `val:"required    one_of=foo|bar"`
	}{Foo: "bar"}
	v := New("val")

	err := v.Validate(x)
	if err == nil {
		t.Fatalf("Expected error")
	}

	v.CheckSep = " "
	v.CheckArgSep = "="

	err = v.Validate(x)
	if err != nil {
		t.Fatalf("Expected no error")
	}
}

func p[T any](v T) *T {
	return &v
}
