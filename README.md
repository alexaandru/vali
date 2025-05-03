# Vali, Yet Another **Vali**dator

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Build and Test](https://github.com/alexaandru/vali/actions/workflows/ci.yml/badge.svg)](https://github.com/alexaandru/vali/actions/workflows/ci.yml)
[![Coverage Status](https://coveralls.io/repos/github/alexaandru/vali/badge.svg)](https://coveralls.io/github/alexaandru/vali)
[![Go Report Card](https://goreportcard.com/badge/github.com/alexaandru/vali)](https://goreportcard.com/report/github.com/alexaandru/vali)
[![Go Reference](https://pkg.go.dev/badge/github.com/alexaandru/vali.svg)](https://pkg.go.dev/github.com/alexaandru/vali)
[![Socket.dev](https://socket.dev/api/badge/go/package/github.com/alexaandru/vali)](https://socket.dev/go/package/github.com/alexaandru/vali)

**Vali**, a purposefully tiny validator. 🔋🔋🔋 are not included,
but there are recipes on how to make them ☺️.

## Description

**Vali** aims for the most things you could do with the least amount
of checks, therefore it has a minimal [set of checks](#available-checks)
and an easy way to add your own checks, see the [example](example_test.go)
and [vali_test.go](vali_test.go) files.

You can also change the struct tag name being used (by creating
a new `Validator`) and a few other bits, see `Validator` type
definition.

It is pointer-insensitive, will always validate the value
behind the pointer, that is, given:

```Go
Foo *string `validate:"required"`
```

passes if `*Foo != ""` NOT if `Foo != nil`.

Finally, it only validates public/exported fields. Adding validation
tags to private fields will be ignored.

Non-goals:

- `slice`/`map` dive;
- cross field checks;
- anything that needs a 3rd party dep.

## Available Checks

| Check          | Description                    | Domain                                                                                                                                                                                                        |
| -------------- | ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| required       | must NOT be `IsZero()`         | `any`                                                                                                                                                                                                         |
| regex:`<rx>`   | must match `<rx>`              | `string`, `Stringer`                                                                                                                                                                                          |
| eq:`<number>`  | must == `number`               | [CanInt](https://pkg.go.dev/reflect#Value.CanInt), [CanUint](https://pkg.go.dev/reflect#Value.CanUint), [CanFloat](https://pkg.go.dev/reflect#Value.CanFloat), Can[Len](https://pkg.go.dev/reflect#Value.Len) |
| ne:`<number>`  | must != `number`               | same as `eq`                                                                                                                                                                                                  |
| min:`<number>` | must be >= `number`            | same as `eq`                                                                                                                                                                                                  |
| max:`<number>` | must be <= `number`            | same as `eq`                                                                                                                                                                                                  |
| one_of:a\|b\|c | must be one of {a,b,c}         | same as `regex`                                                                                                                                                                                               |
| uuid           | 32 (dash separated) hexdigits  | same as `regex`                                                                                                                                                                                               |
| email          | valid email address            | `string`, `Stringer`                                                                                                                                                                                          |
| url            | valid URL with scheme and host | `string`, `Stringer`                                                                                                                                                                                          |
| ipv4           | valid IPv4 address             | `string`, `Stringer`                                                                                                                                                                                          |
| ipv6           | valid IPv6 address             | `string`, `Stringer`                                                                                                                                                                                          |
| ip             | valid IP address (v4 or v6)    | `string`, `Stringer`                                                                                                                                                                                          |
| mac            | valid MAC address              | `string`, `Stringer`                                                                                                                                                                                          |
| domain         | valid domain name              | `string`, `Stringer`                                                                                                                                                                                          |
| isbn           | valid ISBN-10 or ISBN-13       | `string`, `Stringer`                                                                                                                                                                                          |
| alpha          | letters only                   | `string`, `Stringer`                                                                                                                                                                                          |
| alphanum       | letters and numbers only       | `string`, `Stringer`                                                                                                                                                                                          |
| numeric        | numbers only                   | `string`, `Stringer`                                                                                                                                                                                          |
| boolean        | valid boolean representation   | `string`, `Stringer`                                                                                                                                                                                          |
| creditcard     | valid credit card number       | `string`, `Stringer`                                                                                                                                                                                          |
| json           | valid JSON format              | `string`, `Stringer`                                                                                                                                                                                          |
| ascii          | ASCII characters only          | `string`, `Stringer`                                                                                                                                                                                          |
| lowercase      | lowercase characters only      | `string`, `Stringer`                                                                                                                                                                                          |
| uppercase      | uppercase characters only      | `string`, `Stringer`                                                                                                                                                                                          |
| hexadecimal    | valid hexadecimal string       | same as `regex`                                                                                                                                                                                               |
| base64         | valid base64 string            | same as `regex`                                                                                                                                                                                               |
| mongoid        | valid MongoDB ObjectID         | same as `regex`                                                                                                                                                                                               |
| rgb            | valid RGB color                | `string`, `Stringer`                                                                                                                                                                                          |
| rgba           | valid RGBA color               | `string`, `Stringer`                                                                                                                                                                                          |
| luhn           | valid Luhn algorithm check     | `string`, `Stringer`                                                                                                                                                                                          |
| ssn            | valid Social Security Number   | `string`, `Stringer`                                                                                                                                                                                          |
| npi            | valid NPI number               | `string`, `Stringer`                                                                                                                                                                                          |
| `<your_own>`   | you can easily add your own... | ...                                                                                                                                                                                                           |

Multiple checks must be combined with a comma (,) extra space
is forgiven, and empty checks are ignored i.e.:
`validate:"required,,,,  uuid   , one_of:foo|bar|baz"` is fine, albeit unclean.

Both separators (between checks and between a check and its arguments)
are configurable, whereas the separator between a check's arguments (the
pipe symbol in the `a|b|c` example above) are up the each individual checker,
the library doesn't care, it will just pass all the arguments as a string
to the `Checker` func.

## Sample Usage

```Go
s := struct {
	Foo struct {
		Bar string `validate:"required, one_of:foo|bar|baz"`
	}
}{}

if err := vali.Validate(s); err != nil {
    fmt.Println("oh noes!...")
}
```

## Documentation

- this README;
- the [example](example_test.go) file;
- the [code documentation](https://pkg.go.dev/github.com/alexaandru/vali) and
- the [tests](vali_test.go).
