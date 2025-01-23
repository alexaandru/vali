package vali_test

import (
	"fmt"

	"github.com/alexaandru/vali"
)

func ExampleValidationSet_Validate() {
	s := struct {
		Foo struct {
			Bar string `validate:"required"`
		}
	}{}
	err := vali.Validate(s)
	fmt.Println(err)
	// Output: Foo.Bar: required value missing
}

func ExampleValidationSet_Validate_custom_checker() {
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

	// Output: Foo.Bar: regex check failed: "123" does not match ^\d{3}-?\d{3}-?\d{4}$
	// <nil>
}

func ExampleValidationSet_Validate_unexported() {
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
