package vali_test

import (
	"fmt"

	"github.com/alexaandru/vali"
)

func ExampleValidator_Validate() {
	s := struct {
		Foo struct {
			Bar string `validate:"required"`
		}
	}{}
	err := vali.Validate(s)
	fmt.Println(err)
	// Output: Foo.Bar: required check failed: value missing
}

func ExampleValidator_Validate_custom_checker() {
	var phone string

	s := struct {
		Foo struct {
			Bar *string `validate:"phone"`
		}
	}{}
	s.Foo.Bar = &phone

	p, err := vali.Regex(`^\d{3}-?\d{3}-?\d{4}$`)
	if err != nil {
		fmt.Println(err)
	}

	vali.RegisterChecker("phone", p)

	phone = "123"
	err = vali.Validate(s)
	fmt.Println(err) // this should err

	phone = "123-456-7890"
	err = vali.Validate(s)
	fmt.Println(err) // this should not

	// Output: Foo.Bar: phone check failed: "123" does not match ^\d{3}-?\d{3}-?\d{4}$
	// <nil>
}

func ExampleValidator_Validate_custom_min() {
	s := struct {
		Foo struct {
			Bar int8 `validate:"min10"`
		}
	}{}

	min10, err := vali.Min("10")
	if err != nil {
		fmt.Println(err)
	}

	vali.RegisterChecker("min10", min10)

	s.Foo.Bar = 9
	err = vali.Validate(s)
	fmt.Println(err) // this should err

	s.Foo.Bar = 10
	err = vali.Validate(s)
	fmt.Println(err) // this should not

	// Output: Foo.Bar: min10 check failed: 9 is less than 10
	// <nil>
}

func ExampleValidator_Validate_unexported() {
	s := struct {
		Foo struct {
			bar string `validate:"required,uuid"`
		}
	}{}

	s.Foo.bar = "123"

	err := vali.Validate(s)
	fmt.Println(err) // will not validate private fields.
	// Output: <nil>
}
