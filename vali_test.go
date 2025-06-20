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

type testCase struct { //nolint:govet // OK
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

	rgbChecker := func(v reflect.Value) error {
		if !slices.Contains([]string{"red", "green", "blue"}, fmt.Sprint(v.Interface())) {
			return errors.New("must be red, green or blue")
		}

		return nil
	}

	RegisterChecker("rgb", rgbChecker)

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
	x := struct { //nolint:varnamelen // OK
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

//nolint:maintidx,lll // OK
func TestValidate(t *testing.T) { //nolint:funlen // ok
	t.Parallel()

	vNoErr := New()
	vNoErr.ErrorOnPrivate = false
	testCases := []testCase{
		{struct{}{}, &Validator{}, "", "", nil},
		{struct{}{}, New(""), "", "", nil},
		{struct{}{}, nil, "", "", nil},
		{struct{ S *string }{}, nil, "", "", nil},
		{struct {
			S *string `validate:"min:3,max:5"`
		}{S: nil}, nil, "", "", nil},
		{struct {
			S *string `validate:"min:3,max:5"`
		}{S: p("hi")}, nil, "", "S: min check failed: len 2 is less than 3", ErrCheckFailed},
		{struct {
			S *string `validate:"min:3,max:5"`
		}{S: p("hello")}, nil, "", "", nil},
		{struct {
			S *string `validate:"min:3,max:5"`
		}{S: p("helloo")}, nil, "", "S: max check failed: len 6 is more than 5", ErrCheckFailed},
		{struct {
			S **string `validate:""`
		}{}, nil, "", "", nil},
		{struct {
			S ***string `validate:"uuid"`
		}{}, nil, "", "", nil},

		{struct {
			S *string `validate:"required"`
		}{}, nil, "", "S: required check failed: value missing", ErrCheckFailed},
		{struct {
			S **string `validate:"required"`
		}{}, nil, "", "S: required check failed: value missing", ErrCheckFailed},
		{struct {
			S ***string `validate:"required"`
		}{}, nil, "", "S: required check failed: value missing", ErrCheckFailed},
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
		{t2{}, vNoErr, "", "", nil},
		{t1{}, nil, "required", "required check failed: value missing", ErrCheckFailed},
		{t2{}, nil, "required", "required check failed: value missing", ErrCheckFailed},
		{t1{Foo: "foobar"}, nil, "required", "", nil},
		{&t2{t1: t1{Foo: "foobar"}}, vNoErr, "", "", nil},

		{0, nil, "eq:0", "", nil},
		{0, nil, "ne:0", "ne check failed: 0 is equal to 0", ErrCheckFailed},
		{float32(0), nil, "ne:0", "ne check failed: 0 is equal to 0", ErrCheckFailed},
		{float64(0), nil, "ne:0", "ne check failed: 0 is equal to 0", ErrCheckFailed},
		{"", nil, "ne:0", "ne check failed: len 0 is equal to 0", ErrCheckFailed},
		{[]string{}, nil, "ne:0", "ne check failed: len 0 is equal to 0", ErrCheckFailed},
		{map[int]string{}, nil, "ne:0", "ne check failed: len 0 is equal to 0", ErrCheckFailed},

		{1, nil, "eq:0", "eq check failed: 1 is not equal to 0", ErrCheckFailed},
		{float32(1.1), nil, "eq:0", "eq check failed: 1 is not equal to 0", ErrCheckFailed},
		{float64(1.1), nil, "eq:0", "eq check failed: 1 is not equal to 0", ErrCheckFailed},
		{"foo", nil, "eq:0", "eq check failed: len 3 is not equal to 0", ErrCheckFailed},
		{[]string{""}, nil, "eq:0", "eq check failed: len 1 is not equal to 0", ErrCheckFailed},
		{map[int]string{0: ""}, nil, "eq:0", "eq check failed: len 1 is not equal to 0", ErrCheckFailed},

		{1, nil, "eq:1", "", nil},
		{float32(1.1), nil, "eq:1.1", "", nil},
		{float64(1.1), nil, "eq:1.1", "", nil},
		{"foo", nil, "eq:3", "", nil},
		{[]string{""}, nil, "eq:1", "", nil},
		{map[int]string{0: ""}, nil, "eq:1", "", nil},

		{0, nil, "min:foo", `min check failed: strconv.ParseInt: parsing "foo": invalid syntax`, ErrCheckFailed},
		{0, nil, "min:5", "min check failed: 0 is less than 5", ErrCheckFailed},
		{0, nil, "required,min:5", "required check failed: value missing", ErrCheckFailed},
		{uint16(1), nil, "required,min:5", "min check failed: 1 is less than 5", ErrCheckFailed},
		{float32(1), nil, "required,min:5", "min check failed: 1 is less than 5", ErrCheckFailed},
		{float64(1), nil, "required,min:5", "min check failed: 1 is less than 5", ErrCheckFailed},
		{4, nil, "min:5", "min check failed: 4 is less than 5", ErrCheckFailed},
		{uint64(5), nil, "min:5", "", nil},
		{5_000_000_000_000_000, nil, "min:5", "", nil},

		{"", nil, "min:5", "min check failed: len 0 is less than 5", ErrCheckFailed},
		{"", nil, "required,min:5", "required check failed: value missing", ErrCheckFailed},
		{"a", nil, "required,min:5", "min check failed: len 1 is less than 5", ErrCheckFailed},
		{"abcd", nil, "min:5", "min check failed: len 4 is less than 5", ErrCheckFailed},
		{"abcde", nil, "min:5", "", nil},
		{strings.Repeat("abcde", 1_000), nil, "min:5", "", nil},

		{0, nil, "max:foo", `max check failed: strconv.ParseInt: parsing "foo": invalid syntax`, ErrCheckFailed},
		{0, nil, "max:5", "", nil},
		{int32(1000), nil, "max:5", "max check failed: 1000 is more than 5", ErrCheckFailed},
		{uint64(6), nil, "max:5", "max check failed: 6 is more than 5", ErrCheckFailed},
		{float32(6), nil, "max:5", "max check failed: 6 is more than 5", ErrCheckFailed},
		{float64(6), nil, "max:5", "max check failed: 6 is more than 5", ErrCheckFailed},
		{5, nil, "max:5", "", nil},
		{4, nil, "max:5", "", nil},

		{"", nil, "max:5", "", nil},
		{"", nil, "required,max:5", "required check failed: value missing", ErrCheckFailed},
		{"abcdef", nil, "required,max:5", "max check failed: len 6 is more than 5", ErrCheckFailed},
		{"abcde", nil, "max:5", "", nil},
		{"abcd", nil, "max:5", "", nil},
		{"abc", nil, "max:5", "", nil},
		{"ab", nil, "max:5", "", nil},
		{"a", nil, "max:5", "", nil},
		{strings.Repeat("abcde", 1_000), nil, "min:5", "", nil},

		{[]int{}, nil, "min:3", "min check failed: len 0 is less than 3", ErrCheckFailed},
		{[]int{1}, nil, "min:3", "min check failed: len 1 is less than 3", ErrCheckFailed},
		{[]float32{1, 2}, nil, "min:3", "min check failed: len 2 is less than 3", ErrCheckFailed},
		{[]int{1, 2, 3}, nil, "min:3", "", nil},
		{[]int{1, 2, 3, 4, 5}, nil, "min:3", "", nil},

		{[...]int{}, nil, "max:3", "", nil},
		{[...]int{1}, nil, "max:3", "", nil},
		{[...]float32{1, 2}, nil, "max:3", "", nil},
		{[...]int{1, 2, 3}, nil, "max:3", "", nil},
		{[...]float64{1, 2, 3, 4, 5}, nil, "max:3", "max check failed: len 5 is more than 3", ErrCheckFailed},

		{func() {}, nil, "min:2", "min check failed: len check failed: unsupported kind func", ErrCheckFailed},
		{int(1), nil, "eq:foo", `eq check failed: strconv.ParseInt: parsing "foo": invalid syntax`, ErrCheckFailed},
		{uint(1), nil, "ne:foo", `ne check failed: strconv.ParseUint: parsing "foo": invalid syntax`, ErrCheckFailed},
		{float32(1), nil, "min:foo", `min check failed: strconv.ParseFloat: parsing "foo": invalid syntax`, ErrCheckFailed},
		{float64(1), nil, "max:foo", `max check failed: strconv.ParseFloat: parsing "foo": invalid syntax`, ErrCheckFailed},
		{"", nil, "ne:foo", `ne check failed: strconv.Atoi: parsing "foo": invalid syntax`, ErrCheckFailed},

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
		}{foo: "foo"}, vNoErr, "", "", nil},
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
		}{}, nil, "", "Foo: invalid checker bogus", ErrInvalidChecker},
		{struct {
			Foo string `json:",omitempty" validate:"    required,   bogus          "`
			Bar int
		}{Foo: "foo"}, nil, "", "Foo: invalid checker bogus", ErrInvalidChecker},
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
		}{Foo: _uuid}, nil, "", "Foo: invalid checker :", ErrInvalidChecker},
		{struct {
			Foo string `json:",omitempty" validate:"required,foo:"`
			Bar int
		}{Foo: _uuid}, nil, "", "Foo: invalid checker foo:", ErrInvalidChecker},
		{struct {
			Foo string `json:",omitempty" validate:"required,:foo"`
			Bar int
		}{Foo: _uuid}, nil, "", "Foo: invalid checker :foo", ErrInvalidChecker},
		{struct {
			Foo string `json:",omitempty" validate:"required,foo:bar"`
			Bar int
		}{Foo: _uuid}, nil, "", "Foo: invalid checker foo:bar", ErrInvalidChecker},
		{struct {
			Foo string `json:",omitempty" validate:"required,regex:[A-"`
			Bar int
		}{Foo: _uuid}, nil, "", "Foo: invalid checker regex:[A-: error parsing regexp: missing closing ]: `[A-`", ErrInvalidChecker},
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

			var err error

			if tc.s != nil {
				err = tc.s.Validate(tc.v, tc.tag)
			} else {
				err = Validate(tc.v, tc.tag)
			}

			if !errors.Is(err, tc.expErr) {
				t.Fatalf("Expected %v got %v for %v (tag: %q)", tc.expErr, err, tc.v, tc.tag)
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
