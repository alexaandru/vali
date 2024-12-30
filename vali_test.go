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
	s      *ValidationSet
	exp    string
	expErr error
}

type foo []byte

var (
	_uuid = "550e8400-e29b-41d4-a716-446655440000"

	v0 = struct {
		Foo struct {
			Bar struct {
				Baz struct {
					Foobar string `json:",omitempty" xoxo:"required,uuid"`
				}
			}
		}
		Foobar int
	}{}
	v1, v2 = v0, v0
	_      = func() int {
		v1.Foo.Bar.Baz.Foobar = "foo"
		v2.Foo.Bar.Baz.Foobar = _uuid

		return 0
	}()
)

func (f foo) String() string {
	return string(f)
}

func TestValidate(t *testing.T) {
	t.Parallel()

	testCases := slices.Concat(slices.Clone(testCases()), []testCase{
		{v0, NewValidator("xoxo"), "Foo.Bar.Baz.Foobar: required value missing", ErrRequired},
		{v1, NewValidator("xoxo"), `Foo.Bar.Baz.Foobar: invalid regex: "foo" does not match (?i)^[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}$`, ErrInvalid},
		{v2, NewValidator("xoxo"), "", nil},

		{&v0, NewValidator("xoxo"), "Foo.Bar.Baz.Foobar: required value missing", ErrRequired},
		{&v1, NewValidator("xoxo"), `Foo.Bar.Baz.Foobar: invalid regex: "foo" does not match (?i)^[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}$`, ErrInvalid},
		{&v2, NewValidator("xoxo"), "", nil},

		{p(&v0), NewValidator("xoxo"), "Foo.Bar.Baz.Foobar: required value missing", ErrRequired},
		{p(&v1), NewValidator("xoxo"), `Foo.Bar.Baz.Foobar: invalid regex: "foo" does not match (?i)^[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}$`, ErrInvalid},
		{p(&v2), NewValidator("xoxo"), "", nil},
	})

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			v := tc.s
			if v == nil {
				v = NewValidator("validate")
			}

			err := v.Validate(tc.v)
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

func TestTopLevelValidate(t *testing.T) {
	t.Parallel()

	for _, tc := range testCases() {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			err := Validate(tc.v)
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

func TestValidationSetRegisterChecker(t *testing.T) {
	x := struct {
		Foo foo `validate:"rgb"`
	}{
		Foo: []byte("pink"),
	}

	rgb := func(v reflect.Value) error {
		if !slices.Contains([]string{"red", "green", "blue"}, fmt.Sprint(v.Interface())) {
			return fmt.Errorf("%w rgb color: must be red, green or blue", ErrInvalid)
		}

		return nil
	}

	RegisterChecker("rgb", rgb)

	err := Validate(x)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Expected %v got %v", ErrInvalid, err)
	}

	exp := "Foo: invalid rgb color: must be red, green or blue"
	if act := err.Error(); act != exp {
		t.Fatalf("Expected %q got %q", exp, act)
	}
}

func TestValidationSetRegisterCheckerMaker(t *testing.T) {
	x := struct {
		Foo foo `validate:"one_of3:foo|bar|baz"`
	}{
		Foo: []byte("foobar"),
	}

	_oneOf := func(args string) (c Checker, err error) {
		vals := strings.Split(args, "|")
		if len(vals) == 0 {
			return nil, fmt.Errorf("%w one_of: must pass at least one value", ErrInvalid)
		}

		c = func(v reflect.Value) (err error) {
			act := fmt.Sprint(v.Interface())
			if slices.Contains(vals, act) {
				return
			}

			return fmt.Errorf("%w one_of: %q is not one of %v", ErrInvalid, act, vals)
		}

		return
	}

	RegisterCheckerMaker("one_of3", _oneOf)

	err := Validate(x)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Expected %v got %v", ErrInvalid, err)
	}

	exp := `Foo: invalid one_of: "foobar" is not one of [foo bar baz]`
	if act := err.Error(); act != exp {
		t.Fatalf("Expected %q got %q", exp, act)
	}
}

func TestValidationSetConfigurableSeparators(t *testing.T) {
	x := struct {
		Foo string `val:"required    one_of=foo|bar"`
	}{Foo: "bar"}
	v := NewValidator("val")

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

func testCases() []testCase { //nolint:funlen // ok
	return []testCase{
		{struct{}{}, &ValidationSet{}, "", nil},
		{struct{}{}, NewValidator(""), "", nil},
		{struct{}{}, nil, "", nil},
		{"foo", &ValidationSet{}, "", ErrNotAStruct},
		{struct {
			Foo string
			Bar int
		}{}, nil, "", nil},
		{struct {
			Foo string `validate:""`
			Bar int
		}{}, nil, "", nil},
		{struct {
			foo string `validate:"required,uuid"`
			bar int
		}{foo: "foo"}, nil, "", nil},
		{struct {
			Foo string `validate:"                                "`
			Bar int
		}{}, nil, "", nil},
		{struct {
			Foo string `json:",omitempty" validate:"required"`
			Bar int
		}{}, nil, "Foo: required value missing", ErrRequired},
		{struct {
			Foo string `json:",omitempty" validate:"  required   "`
			Bar int
		}{}, nil, "Foo: required value missing", ErrRequired},
		{struct {
			Foo string `json:",omitempty" validate:"    bogus   ,   required   "`
			Bar int
		}{}, nil, "Foo: invalid checker: bogus", ErrInvalidChecker},
		{struct {
			Foo string `json:",omitempty" validate:"    required,   bogus          "`
			Bar int
		}{Foo: "foo"}, nil, "Foo: invalid checker: bogus", ErrInvalidChecker},
		{struct {
			Foo *string `json:",omitempty" validate:"required"`
			Bar int
		}{Foo: p("")}, nil, "Foo: required value missing", ErrRequired},
		{struct {
			Foo ***string `json:",omitempty" validate:"required"`
			Bar int
		}{Foo: p(p(p("")))}, nil, "Foo: required value missing", ErrRequired},
		{struct {
			Foo *string `json:",omitempty" validate:"required"`
			Bar int
		}{}, nil, "Foo: required value missing", ErrRequired},
		{struct {
			Foo *string `json:",omitempty" validate:"uuid"`
			Bar int
		}{}, nil, "", nil},
		{struct {
			Foo *string `json:",omitempty" validate:"uuid,required"`
			Bar int
		}{}, nil, "Foo: required value missing", ErrRequired},
		{struct {
			Foo *string `json:",omitempty" validate:"required,uuid"`
			Bar int
		}{}, nil, "Foo: required value missing", ErrRequired},
		{struct {
			Foo *string `json:",omitempty" validate:"required,uuid"`
			Bar int
		}{Foo: p("foo")}, nil, `Foo: invalid regex: "foo" does not match (?i)^[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}$`, ErrInvalid},
		{struct {
			Foo    string `json:",omitempty" validate:"required"`
			Bar    string `json:",omitempty" validate:"required"`
			Baz    string `json:",omitempty" validate:"required"`
			Foobar int
		}{Foo: "foo", Bar: "bar"}, nil, "Baz: required value missing", ErrRequired},
		{struct {
			Foo string `json:",omitempty" validate:"required,,,,uuid"`
			Bar int
		}{Foo: _uuid}, nil, "", nil},
		{struct {
			Foo string `json:",omitempty" validate:"required,:"`
			Bar int
		}{Foo: _uuid}, nil, "Foo: invalid checker: :", ErrInvalidChecker},
		{struct {
			Foo string `json:",omitempty" validate:"required,foo:"`
			Bar int
		}{Foo: _uuid}, nil, "Foo: invalid checker: foo:", ErrInvalidChecker},
		{struct {
			Foo string `json:",omitempty" validate:"required,:foo"`
			Bar int
		}{Foo: _uuid}, nil, "Foo: invalid checker: :foo", ErrInvalidChecker},
		{struct {
			Foo string `json:",omitempty" validate:"required,foo:bar"`
			Bar int
		}{Foo: _uuid}, nil, "Foo: invalid checker: foo:bar", ErrInvalidChecker},
		{struct {
			Foo string `json:",omitempty" validate:"required,regex:[A-"`
			Bar int
		}{Foo: _uuid}, nil, "Foo: invalid checker: regex:[A-: error parsing regexp: missing closing ]: `[A-`", ErrInvalidChecker},
		{
			struct {
				Foo    string `json:",omitempty" validate:"required"`
				Bar    string `json:",omitempty" validate:"required"`
				Baz    string `json:",omitempty" validate:"required,regex:^(foo|bar|baz)$"`
				Foobar int
			}{Foo: "foo", Bar: "bar", Baz: "other"},
			nil,
			`Baz: invalid regex: "other" does not match ^(foo|bar|baz)$`, ErrInvalid,
		},
		{
			struct {
				Foo    string `json:",omitempty" validate:"regex:^(foo|bar|baz)$"`
				Bar    string `json:",omitempty" validate:"regex:^(foo|bar|baz)$"`
				Baz    string `json:",omitempty" validate:"regex:^(foo|bar|baz)$"`
				Foobar int
			}{Foo: "foo", Bar: "", Baz: "baz"},
			nil, "", nil,
		},
		{
			struct {
				Foo    string `json:",omitempty" validate:"one_of:foo|bar|baz"`
				Bar    string `json:",omitempty" validate:"one_of:foo|bar|baz"`
				Baz    string `json:",omitempty" validate:"one_of:foo|bar|baz"`
				Foobar int
			}{Foo: "foo", Bar: "Bar", Baz: "baz"},
			nil, `Bar: invalid regex: "Bar" does not match ^(foo|bar|baz)$`, ErrInvalid,
		},
	}
}

func p[T any](v T) *T {
	return &v
}
