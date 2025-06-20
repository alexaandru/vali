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
	"slices"
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

		// Checks in this list WILL be checked against the zero value.
		// By default, checks are not run against the zero value, unless they
		// are part of this list.
		DontSkipZeroChecks []string

		// ErrorOnPrivate indicates whether to return an error if a private field
		// has a validation tag. Private fields validation is NOT supported, so
		// setting a tag for it is most likel a mistake.
		ErrorOnPrivate bool

		sync.RWMutex
	}
)

// DefaultValidatorTagName holds the default struct tag name.
const DefaultValidatorTagName = "validate"

// DefaultValidator allows using the library directly, without creating
// a validator, similar to how flags and net/http packages work.
var DefaultValidator = New()

// DefaultDontSkipZero holds the default list of checks that do NOT skip
// the zero value. By default, checks are skipping it, unless they are
// in this list.
//
// This allows checks to be used for optional fields as well, i.e.:
// `validate:"uuid"` will allow an empty string and only validate it
// as uuid if not empty. To BOTH require it to be present and be an uuid,
// you would combine `validate:"required,uuid"`.
//
// In short, checks should be kept small, focused and composable and
// avoid overlapping their responsibilities.
var DefaultDontSkipZero = []string{"required", "eq", "ne", "min", "max"}

// New creates a new [Validator], initialized with the default checkers
// and ready to be used. You can optionally pass a struct tag name or
// use the [DefaultValidatorTagName].
//
// By default, it errors out if it encounters validation tags on private
// fields, but you can change that by setting the [Validator.ErrorOnPrivate]
// to false. The error will be of type [ErrPrivateField].
func New(opts ...string) (v *Validator) {
	tag := DefaultValidatorTagName
	if len(opts) > 0 {
		tag = opts[0]
	}

	v = &Validator{
		CheckSep: ",", CheckArgSep: ":",
		tag:                tag,
		checkers:           map[string]Checker{},
		checkerMakers:      map[string]CheckerMaker{},
		DontSkipZeroChecks: DefaultDontSkipZero,
		ErrorOnPrivate:     true,
	}

	v.RegisterChecker("required", required)
	v.RegisterChecker("uuid", uuid)
	v.RegisterChecker("email", email)
	v.RegisterChecker("url", urL)
	v.RegisterChecker("ipv4", ipv4)
	v.RegisterChecker("ipv6", ipv6)
	v.RegisterChecker("ip", ip)
	v.RegisterChecker("mac", mac)
	v.RegisterChecker("domain", domain)
	v.RegisterChecker("isbn", isbn)
	v.RegisterChecker("alpha", alpha)
	v.RegisterChecker("alphanum", alphaNum)
	v.RegisterChecker("numeric", numeric)
	v.RegisterChecker("boolean", boolean)
	v.RegisterChecker("creditcard", creditCard)
	v.RegisterChecker("mongoid", mongoID)
	v.RegisterChecker("hexadecimal", hexadecimal)
	v.RegisterChecker("base64", base64)
	v.RegisterChecker("json", jsoN)
	v.RegisterChecker("ascii", ascii)
	v.RegisterChecker("lowercase", lowercase)
	v.RegisterChecker("uppercase", uppercase)
	v.RegisterChecker("rgb", rgb)
	v.RegisterChecker("rgba", rgba)
	v.RegisterChecker("luhn", luhn)
	v.RegisterChecker("ssn", ssn)
	v.RegisterChecker("npi", npi)

	v.RegisterCheckerMaker("regex", Regex)
	v.RegisterCheckerMaker("eq", Eq)
	v.RegisterCheckerMaker("ne", Ne)
	v.RegisterCheckerMaker("min", Min)
	v.RegisterCheckerMaker("max", Max)
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
	ref := reflect.ValueOf(val)

	return v.validate(ref, tag)
}

func (v *Validator) validate(val reflect.Value, tag string, scope ...string) (err error) {
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if tag != "" {
		if err = v.validateScalar(val, tag, scope...); err != nil {
			return
		}
	}

	if val.Kind() != reflect.Struct {
		return
	}

	for i := range val.NumField() {
		iType := val.Type().Field(i)
		reflTag := iType.Tag
		tag = strings.TrimSpace(reflTag.Get(v.tag))

		if !iType.IsExported() {
			if v.ErrorOnPrivate && tag != "" {
				return fmt.Errorf("%s: %w, will not validate", strings.Join(append(scope, iType.Name), "."), ErrPrivateField)
			}

			continue
		}

		iVal := val.Field(i)
		for iVal.Kind() == reflect.Ptr {
			iVal = iVal.Elem()
		}

		if tag == "" && iVal.Kind() != reflect.Struct {
			continue
		}

		iName := val.Type().Field(i).Name
		localScope := append(scope, iName) //nolint:gocritic // ok

		err = v.validate(iVal, tag, localScope...)
		if err != nil {
			return
		}
	}

	return
}

func (v *Validator) validateScalar(val reflect.Value, tag string, scope ...string) (err error) {
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
		name := chkNames[i]
		if strings.Contains(name, v.CheckArgSep) {
			nx := strings.Split(name, v.CheckArgSep)
			name = nx[0]
		}

		if isZero(val) && !slices.Contains(v.DontSkipZeroChecks, name) {
			continue
		}

		if err = ck(val); err != nil {
			return fmt.Errorf("%s %w: %w", name, ErrCheckFailed, err)
		}
	}

	return
}

func (v *Validator) parse(tag string) (cx []Checker, cxNames []string, err error) {
	for tag := range strings.SplitSeq(tag, v.CheckSep) {
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
				return nil, nil, fmt.Errorf("%w %s", ErrInvalidChecker, tag)
			}

			v.RLock()
			cm := v.checkerMakers[tagz[0]]
			v.RUnlock()

			if cm == nil {
				return nil, nil, fmt.Errorf("%w %s", ErrInvalidChecker, tag)
			}

			c, err2 := cm(tagz[1])
			if err2 != nil {
				return nil, nil, fmt.Errorf("%w %s: %w", ErrInvalidChecker, tag, err2)
			}

			v.RegisterChecker(tag, c)
			cx = append(cx, c)
			cxNames = append(cxNames, tagz[0])
		default:
			return nil, nil, fmt.Errorf("%w %s", ErrInvalidChecker, tag)
		}
	}

	return
}
