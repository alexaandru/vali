package vali_test

import (
	"fmt"

	"github.com/alexaandru/vali"
)

type foo struct{}

func (foo) Foo() string {
	return "hello world"
}

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
	fmt.Println(err) // This should err.

	phone = "123-456-7890"
	err = vali.Validate(s)
	fmt.Println(err) // This should not.

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
	fmt.Println(err) // This should err.

	s.Foo.Bar = 10
	err = vali.Validate(s)
	fmt.Println(err) // This should not.

	// Output: Foo.Bar: min10 check failed: 9 is less than 10
	// <nil>
}

func ExampleValidator_Validate_interface() {
	type fooer interface {
		Foo() string
	}

	s := struct {
		F fooer `validate:"required"`
	}{}

	err := vali.Validate(s)
	fmt.Println(err) // This should err.

	s.F = foo{}
	err = vali.Validate(s)
	fmt.Println(err) // This should not.

	// Output: F: required check failed: value missing
	// <nil>
}

func ExampleValidator_Validate_unexported() {
	s := struct {
		Foo struct {
			bar string `validate:"required,uuid"`
		}
	}{}

	s.Foo.bar = "123"

	v := vali.New()
	v.ErrorOnPrivate = false
	err := v.Validate(s)
	fmt.Println(err) // Will not validate private fields.

	v.ErrorOnPrivate = true // The default.
	err = v.Validate(s)
	fmt.Println(err) // This will error out.

	// Output: <nil>
	// Foo.bar: private field, will not validate
}

func ExampleValidator_Validate_luhn() {
	s := struct {
		CreditCard string `validate:"luhn"`
	}{}

	// Valid credit card number (passing Luhn algorithm).
	s.CreditCard = "4111 1111 1111 1111"
	err := vali.Validate(s)
	fmt.Println(err)

	// Invalid credit card number (failing Luhn algorithm).
	s.CreditCard = "4111 1111 1111 1112"
	if err = vali.Validate(s); err != nil {
		fmt.Println("Invalid")
	}

	// Output: <nil>
	// Invalid
}

func ExampleValidator_Validate_ssn_npi() {
	s := struct {
		SSN string `validate:"ssn"`
		NPI string `validate:"npi"`
	}{}

	// Valid SSN.
	s.SSN = "123-45-6789"
	err := vali.Validate(s)
	fmt.Println("Valid SSN:", err)

	// Invalid SSN format.
	s.SSN = "12345-6789"
	err = vali.Validate(s)
	fmt.Println("Invalid SSN:", err != nil)

	// Valid NPI (with 80840 prefix for Luhn check).
	s.SSN = ""
	s.NPI = "1234567893"
	err = vali.Validate(s)
	fmt.Println("Valid NPI:", err)

	// Invalid NPI.
	s.NPI = "12345"
	err = vali.Validate(s)
	fmt.Println("Invalid NPI:", err != nil)

	// Output: Valid SSN: <nil>
	// Invalid SSN: true
	// Valid NPI: <nil>
	// Invalid NPI: true
}
