# Vali, Yet Another **Vali**dator

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Build and Test](https://github.com/alexaandru/vali/actions/workflows/ci.yml/badge.svg)](https://github.com/alexaandru/vali/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/alexaandru/vali)](https://goreportcard.com/report/github.com/alexaandru/vali)
[![Go Reference](https://pkg.go.dev/badge/github.com/alexaandru/vali.svg)](https://pkg.go.dev/github.com/alexaandru/vali)

**Vali** is a, purposefully tiny, validator which started as an exercise
of _"what it would take to...?"_ and ended up being quite useful.

It is pointer-insensitive, will always validate the value
behind the pointer, that is, given:

```Go
Foo *string `validate:"required"`
```

passes if `*Foo != ""` NOT if `Foo != nil`.

It has only a few, very basic checks, but it can easily be extended
with custom checks, see the [example](example_test.go) and
[vali_test.go](vali_test.go) files. You can also easily change the
struct tag being used (by creating a new `ValidationSet`).

It will most likely NOT be extended with any other checks, if you need
more power use something like [go-playground/validator](https://github.com/go-playground/validator)
or fork it and have fun!

At **180**LOC (not counting blank lines/comments/tests) it's already
bigger than I intended. I want the code to remain easy to understand
and easy to extend, which is why I plan to keep it as simple as possible.

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
