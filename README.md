# Vali, Yet Another **Vali**dator

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Build and Test](https://github.com/alexaandru/vali/actions/workflows/ci.yml/badge.svg)](https://github.com/alexaandru/vali/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/alexaandru/vali)](https://goreportcard.com/report/github.com/alexaandru/vali)
[![Go Reference](https://pkg.go.dev/badge/github.com/alexaandru/vali.svg)](https://pkg.go.dev/github.com/alexaandru/vali)

**Vali**, a purposefully tiny validator. 🔋🔋🔋 are not included,
but there are recipes on how to make them ☺️.

Rather than a large collection of checks (which ultimately
imply many deps) I wanted to try and go for the most things
I could do with the least amount of checks.

I didn't do a good job at that, I'll reckon, I think this
library already has too many checks 🙃. I may delete some
and it's highly unlikely I'll add any new ones.

## Description

It has a minimal [set of checks](#available-checks) and
an easy way to add your own checks, see the [example](example_test.go) and
[vali_test.go](vali_test.go) files.

You can also change the struct tag being used (by creating
a new `ValidationSet`) and a few other bits, see `ValidationSet` type
definition.

It is pointer-insensitive, will always validate the value
behind the pointer, that is, given:

```Go
Foo *string `validate:"required"`
```

passes if `*Foo != ""` NOT if `Foo != nil`.

Finally, it only validates public/exported fields. Adding validation
tags to private fields will be ignored.

### Available checks:

| Check          | Description                             |
| -------------- | --------------------------------------- |
| required       | must NOT be `IsZero()`                  |
| uuid           | 32 (dash separated) hexdigits           |
| regex:`<rx>`   | string representation must match `<rx>` |
| one_of:a\|b\|c | must be one of {a,b,c}                  |
| `<your_own>`   | you can easily add your own...          |

Multiple checks must be combined with a comma (,) extra space
is forgiven, and empty checks are ignored i.e.:
`validate:"required,,,,  uuid   , one_of:foo|bar|baz"` is fine, albeit unclean.

Both separators (between checks and between a check and its arguments)
are configurable, whereas the separator between arguments themselves (the
pipe symbol in the `a|b|c` example above) are up the each individual check,
the library doesn't care it will just pass all the arguments as a string
to the `Checker` func.

## Sample Usage

```Go
s := struct {
	Foo struct {
		Bar string `validate:"required"`
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
