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

	// ValidationSet holds the validation context.
	// You can create your own or use the default one provided by this library.
	ValidationSet struct {
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

// DefaultValidator allows using the library directly, without creating
// a validator, similar to how flags and net/http packages work.
var DefaultValidator *ValidationSet

// NewValidator creates a new [ValidationSet], initialized with
// the default checkers and ready to be used.
func NewValidator(tag string) (s *ValidationSet) {
	s = &ValidationSet{
		CheckSep: ",", CheckArgSep: ":",
		tag:           tag,
		checkers:      map[string]Checker{},
		checkerMakers: map[string]CheckerMaker{},
	}

	s.RegisterChecker("required", required)
	s.RegisterChecker("uuid", uuid)
	s.RegisterCheckerMaker("regex", Regex)
	s.RegisterCheckerMaker("one_of", oneOf)

	return
}

// RegisterChecker registers a new [Checker] to the [DefaultValidator].
func RegisterChecker(name string, fn Checker) {
	DefaultValidator.RegisterChecker(name, fn)
}

// RegisterChecker registers a new [Checker] to the [ValidationSet].
func (s *ValidationSet) RegisterChecker(name string, fn Checker) {
	s.Lock()
	defer s.Unlock()

	s.checkers[name] = fn
}

// RegisterCheckerMaker registers a new [CheckerMaker] to the [DefaultValidator].
func RegisterCheckerMaker(name string, fn CheckerMaker) {
	DefaultValidator.RegisterCheckerMaker(name, fn)
}

// RegisterCheckerMaker registers a new [CheckerMaker] to the [ValidationSet].
func (s *ValidationSet) RegisterCheckerMaker(name string, fn CheckerMaker) {
	s.Lock()
	defer s.Unlock()

	s.checkerMakers[name] = fn
}

// Validate validates v against [DefaultValidator].
// See [ValidationSet.Validate] for details.
func Validate(v any, scope ...string) error {
	return DefaultValidator.Validate(v, scope...)
}

// Validate validates a struct. The passed value v can be a value or
// a pointer (or pointer to a pointer, although there's no point to do that in Go).
// It will validate all the fields that have the `s.tag` present, recursively.
func (s *ValidationSet) Validate(v any, scope ...string) (err error) {
	x := reflect.ValueOf(v)
	for x.Kind() == reflect.Ptr {
		x = x.Elem()
	}

	if x.Kind() != reflect.Struct {
		return ErrNotAStruct
	}

	for i := range x.NumField() {
		xType := x.Type().Field(i)
		if !xType.IsExported() {
			continue
		}

		reflTag := xType.Tag
		tag := strings.TrimSpace(reflTag.Get(s.tag))

		y := x.Field(i)
		for y.Kind() == reflect.Ptr {
			y = y.Elem()
		}

		if tag == "" && y.Kind() != reflect.Struct {
			continue
		}

		yName := x.Type().Field(i).Name
		localScope := append(scope, yName) //nolint:gocritic // ok

		if y.Kind() == reflect.Struct {
			err = s.Validate(y.Interface(), localScope...)
		} else {
			err = s.validateScalar(y, tag, localScope...)
		}

		if err != nil {
			return
		}
	}

	return
}

func (s *ValidationSet) validateScalar(v reflect.Value, tag string, scope ...string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s: %w", strings.Join(scope, "."), err)
		}
	}()

	checks, chkNames, err := s.parse(tag)
	if err != nil {
		return
	}

	for i, ck := range checks {
		if err = ck(v); err != nil {
			name := chkNames[i]
			if strings.Contains(name, s.CheckArgSep) {
				nx := strings.Split(name, s.CheckArgSep)
				name = nx[0]
			}

			return fmt.Errorf("%s %w: %w", name, ErrCheckFailed, err)
		}
	}

	return
}

func (s *ValidationSet) parse(tag string) (cx []Checker, cxNames []string, err error) {
	for _, tag := range strings.Split(tag, s.CheckSep) {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}

		s.RLock()
		v := s.checkers[tag]
		s.RUnlock()

		switch {
		case v != nil:
			cx = append(cx, v)
			cxNames = append(cxNames, tag)
		case strings.Contains(tag, s.CheckArgSep):
			tagz := strings.Split(tag, s.CheckArgSep)
			if len(tagz) != 2 || tagz[0] == "" || tagz[1] == "" {
				return nil, nil, fmt.Errorf("%w: %s", ErrInvalidChecker, tag)
			}

			s.RLock()
			cm := s.checkerMakers[tagz[0]]
			s.RUnlock()

			if cm == nil {
				return nil, nil, fmt.Errorf("%w: %s", ErrInvalidChecker, tag)
			}

			c, err2 := cm(tagz[1])
			if err2 != nil {
				return nil, nil, fmt.Errorf("%w: %s: %w", ErrInvalidChecker, tag, err2)
			}

			s.RegisterChecker(tag, c)
			cx = append(cx, c)
			cxNames = append(cxNames, tagz[0])
		default:
			return nil, nil, fmt.Errorf("%w: %s", ErrInvalidChecker, tag)
		}
	}

	return
}
