// Package vali is a tiny validation library.
//
// It is pointer-insensitive, will always validate the value
// behind the pointer (i.e. *string required passes if string != ""
// not if *string != nil).
//
// You can pass it a struct, a *struct, a *****struct, doesn't matter,
// it will always fast-forward to the value and ignore any pointers.
//
// It is very small, but extensible, you can easily add your own checkers
// or "checker makers" (basically, checkers that can take arguments).
package vali

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type (
	// Checker repesents a basic checker (one that takes no arguments, i.e. "required").
	Checker func(reflect.Value) error

	// CheckerMaker is a way to construct checkers with arguments (i.e. "regex:^[A-Z]$").
	CheckerMaker func(args string) (Checker, error)

	// Validator holds the validation context.
	// You can create your own or use the default one provided by this library.
	Validator struct {
		checkers      map[string]Checker
		checkerMakers map[string]CheckerMaker
		tag           string

		// Separator between checks (a), cheks and their arguments (b). The check between
		// arguments themselves is not configurable (c), as that is ultimately up to each
		// individual checker (how to parse the arguments). The only builtin check that uses
		// it is `one_of` and that one requires it to be the pipe symbol.
		//
		//     `validate:"required(a)uuid(a)one_of(b)foo|bar|baz"` which defaults to:
		//     `validate:"required,uuid,one_of:foo|bar|baz"`
		CheckSep,
		CheckArgSep string

		sync.RWMutex
	}
)

// DefaultValidatorTagName holds the default struct tag name.
const DefaultValidatorTagName = "validate"

// DefaultValidator allows using the library directly, without creating
// a validator, similar to how flags and net/http packages work.
var DefaultValidator *Validator

// New creates a new [Validator], initialized with the default checkers
// and ready to be used. You can optionally pass a struct tag name or
// use the [DefaultValidatorTagName].
func New(opts ...string) (v *Validator) {
	tag := DefaultValidatorTagName
	if len(opts) > 0 {
		tag = opts[0]
	}

	v = &Validator{
		CheckSep: ",", CheckArgSep: ":",
		tag:           tag,
		checkers:      map[string]Checker{},
		checkerMakers: map[string]CheckerMaker{},
	}

	v.RegisterChecker("required", required)
	v.RegisterChecker("uuid", uuid)
	v.RegisterCheckerMaker("regex", Regex)
	v.RegisterCheckerMaker("one_of", oneOf)

	return
}

// RegisterChecker registers a new [Checker] to the [DefaultValidator].
func RegisterChecker(name string, fn Checker) {
	DefaultValidator.RegisterChecker(name, fn)
}

// RegisterChecker registers a new [Checker] to the [Validator].
func (v *Validator) RegisterChecker(name string, fn Checker) {
	v.Lock()
	defer v.Unlock()

	v.checkers[name] = fn
}

// RegisterCheckerMaker registers a new [CheckerMaker] to the [DefaultValidator].
func RegisterCheckerMaker(name string, fn CheckerMaker) {
	DefaultValidator.RegisterCheckerMaker(name, fn)
}

// RegisterCheckerMaker registers a new [CheckerMaker] to the [Validator].
func (v *Validator) RegisterCheckerMaker(name string, fn CheckerMaker) {
	v.Lock()
	defer v.Unlock()

	v.checkerMakers[name] = fn
}

// Validate validates v against [DefaultValidator].
// See [Validator.Validate] for details.
func Validate(val any, tags ...string) error {
	return DefaultValidator.Validate(val, tags...)
}

// Validate validates a struct. The passed value v can be a value or
// a pointer (or pointer to a pointer, although there's no point to do that in Go).
// It will validate all the fields that have the `s.tag` present, recursively.
func (v *Validator) Validate(val any, tags ...string) (err error) {
	tag := strings.Join(tags, v.CheckSep)
	x := reflect.ValueOf(val)

	return v.validate(x, tag)
}

func (v *Validator) validate(x reflect.Value, tag string, scope ...string) (err error) {
	for x.Kind() == reflect.Ptr {
		x = x.Elem()
	}

	if tag != "" {
		if err = v.validateScalar(x, tag, scope...); err != nil {
			return
		}
	}

	if x.Kind() != reflect.Struct {
		return
	}

	for i := range x.NumField() {
		xType := x.Type().Field(i)
		if !xType.IsExported() {
			continue
		}

		reflTag := xType.Tag
		tag = strings.TrimSpace(reflTag.Get(v.tag))

		y := x.Field(i)
		for y.Kind() == reflect.Ptr {
			y = y.Elem()
		}

		if tag == "" && y.Kind() != reflect.Struct {
			continue
		}

		yName := x.Type().Field(i).Name
		localScope := append(scope, yName) //nolint:gocritic // ok

		err = v.validate(y, tag, localScope...)
		if err != nil {
			return
		}
	}

	return
}

func (v *Validator) validateScalar(x reflect.Value, tag string, scope ...string) (err error) {
	defer func() {
		if err != nil && len(scope) > 0 {
			err = fmt.Errorf("%s: %w", strings.Join(scope, "."), err)
		}
	}()

	checks, chkNames, err := v.parse(tag)
	if err != nil {
		return
	}

	for i, ck := range checks {
		if err = ck(x); err != nil {
			name := chkNames[i]
			if strings.Contains(name, v.CheckArgSep) {
				nx := strings.Split(name, v.CheckArgSep)
				name = nx[0]
			}

			return fmt.Errorf("%s %w: %w", name, ErrCheckFailed, err)
		}
	}

	return
}

func (v *Validator) parse(tag string) (cx []Checker, cxNames []string, err error) {
	for _, tag := range strings.Split(tag, v.CheckSep) {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}

		v.RLock()
		ck := v.checkers[tag]
		v.RUnlock()

		switch {
		case ck != nil:
			cx = append(cx, ck)
			cxNames = append(cxNames, tag)
		case strings.Contains(tag, v.CheckArgSep):
			tagz := strings.Split(tag, v.CheckArgSep)
			if len(tagz) != 2 || tagz[0] == "" || tagz[1] == "" {
				return nil, nil, fmt.Errorf("%w: %s", ErrInvalidChecker, tag)
			}

			v.RLock()
			cm := v.checkerMakers[tagz[0]]
			v.RUnlock()

			if cm == nil {
				return nil, nil, fmt.Errorf("%w: %s", ErrInvalidChecker, tag)
			}

			c, err2 := cm(tagz[1])
			if err2 != nil {
				return nil, nil, fmt.Errorf("%w: %s: %w", ErrInvalidChecker, tag, err2)
			}

			v.RegisterChecker(tag, c)
			cx = append(cx, c)
			cxNames = append(cxNames, tagz[0])
		default:
			return nil, nil, fmt.Errorf("%w: %s", ErrInvalidChecker, tag)
		}
	}

	return
}
