package vali

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type expOutcome int

const (
	expLess expOutcome = iota - 1
	expEq
	expMore
	expNotEq
)

const rgbRange = `(?:2(?:5[0-5]|[0-4]\d)|1\d\d|[1-9]?\d)`

// Possible errors.
var (
	ErrCheckFailed    = errors.New("check failed")
	ErrRequired       = errors.New("value missing")
	ErrInvalidChecker = errors.New("invalid checker")
	ErrInvalidCmp     = errors.New("invalid comparison")
	ErrPrivateField   = errors.New("private field")
)

//nolint:errcheck,lll // well covered with tests
var (
	npiRx          = regexp.MustCompile(`^\d{10}$`)
	uuid, _        = Regex(`(?i)^[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}$`)
	mongoID, _     = Regex(`(?i)^[0-9a-f]{24}$`)
	hexadecimal, _ = Regex(`(?i)^[0-9a-f]+$`)
	base64, _      = Regex(`(?i)^(?:[a-z0-9+/]{4})*(?:[a-z0-9+/]{2}==|[a-z0-9+/]{3}=)?$`)
	domain, _      = Regex(`(?i)^([a-z0-9]([a-z0-9\-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$`)
	ssn, _         = Regex(`^(0(0[1-9]|[1-9]\d)|[1-5]\d\d|6([0-5]\d|6[0-5]|6[7-9]|[7-9]\d)|[7-8]\d\d)-(0[1-9]|[1-9]\d)-(000[1-9]|00[1-9]\d|0[1-9]\d\d|[1-9]\d\d\d)$`)
	alpha, _       = Regex(`(?i)^[a-z]*$`)
	alphaNum, _    = Regex(`(?i)^[a-z0-9]*$`)
	numeric, _     = Regex(`^\d*$`)
	rgb, _         = Regex(`^rgb\((` + rgbRange + `),(` + rgbRange + `),(` + rgbRange + `)\)$`)
	rgba, _        = Regex(`^rgba\((` + rgbRange + `),(` + rgbRange + `),(` + rgbRange + `),(0|1|0?\.\d+)\)$`)
)

var expLabel = map[expOutcome]string{
	expLess:  "more than",
	expMore:  "less than",
	expEq:    "not equal to",
	expNotEq: "equal to",
}

func email(v reflect.Value) (err error) {
	s := fmt.Sprint(v.Interface())
	if _, err = mail.ParseAddress(s); err != nil {
		return fmt.Errorf("%q is not a valid email address", s)
	}

	return
}

func urL(v reflect.Value) (err error) {
	s := fmt.Sprint(v.Interface())

	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("%q is not a valid URL: %w", s, err)
	}

	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("%q is not a valid URL (missing scheme or host)", s)
	}

	return
}

func ip(v reflect.Value) (err error) {
	if s := fmt.Sprint(v.Interface()); net.ParseIP(s) == nil {
		return fmt.Errorf("%q is not a valid IP address", s)
	}

	return
}

func ipv4(v reflect.Value) (err error) {
	s := fmt.Sprint(v.Interface())
	if ip := net.ParseIP(s); ip == nil || ip.To4() == nil {
		return fmt.Errorf("%q is not a valid IPv4 address", s)
	}

	return
}

func ipv6(v reflect.Value) (err error) {
	s := fmt.Sprint(v.Interface())
	if ip := net.ParseIP(s); ip == nil || ip.To4() != nil {
		return fmt.Errorf("%q is not a valid IPv6 address", s)
	}

	return
}

func mac(v reflect.Value) (err error) {
	s := fmt.Sprint(v.Interface())
	if _, err = net.ParseMAC(s); err != nil {
		return fmt.Errorf("%q is not a valid MAC address", s)
	}

	return
}

func isbn(v reflect.Value) (err error) {
	switch s := strings.ReplaceAll(fmt.Sprint(v.Interface()), "-", ""); len(s) {
	case 10:
		return validateISBN10(s)
	case 13:
		return validateISBN13(s)
	default:
		return fmt.Errorf("%q is not a valid ISBN (must be 10 or 13 digits)", s)
	}
}

func validateISBN10(s string) (err error) {
	var sum int

	// Sum the first 9 digits with weights 10 to 2 (position-based).
	for i := range 9 {
		if s[i] < '0' || s[i] > '9' {
			return fmt.Errorf("invalid character in ISBN-10: %c", s[i])
		}

		sum += int(s[i]-'0') * (10 - i)
	}

	switch v := s[9]; { // Check the last character.
	case v == 'X' || v == 'x': // (can be 'X' which represents 10).
		sum += 10
	case v >= '0' && v <= '9':
		sum += int(v - '0')
	default:
		return fmt.Errorf("invalid character in ISBN-10: %c", v)
	}

	if sum%11 != 0 {
		return fmt.Errorf("%q is not a valid ISBN-10 (checksum failed)", s)
	}

	return
}

func validateISBN13(s string) (err error) {
	var sum int

	// For the first 12 digits, odd positions have weight 1, even positions have weight 3.
	for i := range 12 {
		if s[i] < '0' || s[i] > '9' {
			return fmt.Errorf("invalid character in ISBN-13: %c", s[i])
		}

		if i%2 == 0 {
			sum += int(s[i] - '0') // Weight 1 for odd positions (0-indexed).
		} else {
			sum += 3 * int(s[i]-'0') // Weight 3 for even positions (0-indexed).
		}
	}

	// Validate the check digit.
	if s[12] < '0' || s[12] > '9' {
		return fmt.Errorf("invalid character in ISBN-13: %c", s[12])
	}

	// Calculate expected check digit: (10 - (sum % 10)) % 10.
	checkDigit := (10 - (sum % 10)) % 10
	if int(s[12]-'0') != checkDigit {
		return fmt.Errorf("%q is not a valid ISBN-13 (checksum failed)", s)
	}

	return
}

func boolean(v reflect.Value) (err error) {
	switch s := fmt.Sprint(v.Interface()); strings.ToLower(s) {
	case "1", "t", "true", "yes", "y", "on":
		return
	case "0", "f", "false", "no", "n", "off":
		return
	default:
		return fmt.Errorf("%q is not a valid boolean value", s)
	}
}

func creditCard(v reflect.Value) (err error) {
	s := fmt.Sprint(v.Interface())
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")

	if len(s) < 13 || len(s) > 19 {
		return fmt.Errorf("%q is not a valid credit card number (wrong length)", s)
	}

	for _, r := range s {
		if !unicode.IsDigit(r) {
			return fmt.Errorf("%q contains non-digit character: %q", s, r)
		}
	}

	return luhn(reflect.ValueOf(s))
}

func jsoN(v reflect.Value) (err error) {
	var (
		s  = fmt.Sprint(v.Interface())
		js any
	)

	if err = json.Unmarshal([]byte(s), &js); err != nil {
		return fmt.Errorf("%q is not valid JSON: %w", s, err)
	}

	return
}

func ascii(v reflect.Value) (err error) {
	s := fmt.Sprint(v.Interface())
	for i, r := range s {
		if r > unicode.MaxASCII {
			return fmt.Errorf("%q contains non-ASCII character %q at position %d", s, r, i)
		}
	}

	return
}

func lowercase(v reflect.Value) (err error) {
	s := fmt.Sprint(v.Interface())
	for i, r := range s {
		if unicode.IsUpper(r) {
			return fmt.Errorf("%q contains uppercase character %q at position %d", s, r, i)
		}
	}

	return
}

func uppercase(v reflect.Value) (err error) {
	s := fmt.Sprint(v.Interface())
	for i, r := range s {
		if unicode.IsLower(r) {
			return fmt.Errorf("%q contains lowercase character %q at position %d", s, r, i)
		}
	}

	return
}

// Luhn validates strings or numbers using the Luhn algorithm.
func luhn(v reflect.Value) (err error) {
	var s string //nolint:varnamelen // ok

	// Convert numeric types to string without scientific notation.
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		// Format float without scientific notation and remove decimal point.
		s = strings.ReplaceAll(fmt.Sprintf("%.0f", v.Float()), ".", "")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s = fmt.Sprintf("%d", v.Interface())
	default:
		s = fmt.Sprint(v.Interface())
	}

	s = strings.ReplaceAll(strings.ReplaceAll(s, " ", ""), "-", "")
	if s == "" {
		return fmt.Errorf("%q is not a valid input for Luhn validation", s)
	}

	sum, double := 0, false

	for i := len(s) - 1; i >= 0; i-- {
		if s[i] < '0' || s[i] > '9' {
			return fmt.Errorf("%q contains non-digit character: %q", s, s[i])
		}

		digit := int(s[i] - '0')
		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		double = !double
	}

	if sum%10 != 0 {
		return fmt.Errorf("%q is not valid according to the Luhn algorithm", s)
	}

	return
}

// NPI validates if a string is a valid National Provider Identifier.
func npi(v reflect.Value) (err error) {
	s := fmt.Sprint(v.Interface())
	if !npiRx.MatchString(s) {
		return fmt.Errorf("%q is not a valid NPI", s)
	}

	// NPI validation requires prefixing "80840" to the 10-digit number before applying the Luhn algorithm.
	return luhn(reflect.ValueOf("80840" + s))
}

func required(v reflect.Value) (err error) {
	if isZero(v) {
		return ErrRequired
	}

	return
}

// Regex allows you to easily create regex-based checkers.
func Regex(arg string) (c Checker, err error) {
	rx, err := regexp.Compile(arg)
	if err != nil {
		return
	}

	return func(v reflect.Value) (err error) {
		act := fmt.Sprint(v.Interface())
		if rx.MatchString(act) {
			return
		}

		return fmt.Errorf("%q does not match %s", act, arg)
	}, nil
}

// Eq checks numbers for being == `arg` and things with a `len()`
// (`array`, `chan`, `map`, `slice`, `string`) for having len == `arg`.
func Eq(arg string) (c Checker, err error) {
	return sizeCmp(arg, expEq)
}

// Ne checks numbers for being != `arg` and things with a `len()`
// (`array`, `chan`, `map`, `slice`, `string`) for having len != `arg`.
func Ne(arg string) (c Checker, err error) {
	return sizeCmp(arg, expNotEq)
}

// Min checks numbers for being at least `arg` and things with a `len()`
// (`array`, `chan`, `map`, `slice`, `string`) for having len at least `arg`.
func Min(arg string) (c Checker, err error) {
	return sizeCmp(arg, expMore)
}

// Max checks numbers for being at most `arg` and things with a `len()`
// (`array`, `chan`, `map`, `slice`, `string`) for having len at most `arg`.
func Max(arg string) (c Checker, err error) {
	return sizeCmp(arg, expLess)
}

//nolint:nakedret,gocognit,funlen,cyclop // ok
func sizeCmp(arg string, exp expOutcome) (c Checker, err error) {
	label := expLabel[exp]

	return func(v reflect.Value) (err error) {
		defer func() {
			if r := recover(); r != nil {
				if v, ok := r.(error); ok {
					err = v
				} else {
					err = errors.New(fmt.Sprint(r))
				}
			}
		}()

		switch {
		case v.CanInt():
			var x int64

			if x, err = strconv.ParseInt(arg, 10, 64); err != nil {
				return
			}

			if y := v.Int(); cmp2(y, x, exp) {
				return fmt.Errorf("%d is %s %d", y, label, x)
			}
		case v.CanUint():
			var x uint64

			if x, err = strconv.ParseUint(arg, 10, 64); err != nil {
				return
			}

			if y := v.Uint(); cmp2(y, x, exp) {
				return fmt.Errorf("%d is %s %d", y, label, x)
			}
		case v.CanFloat():
			var x float64

			switch vv := v.Interface().(type) {
			case float32:
				if x, err = strconv.ParseFloat(arg, 32); err != nil {
					return
				}

				if cmp2(vv, float32(x), exp) {
					return fmt.Errorf("%.0f is %s %.0f", vv, label, x)
				}
			case float64:
				if x, err = strconv.ParseFloat(arg, 64); err != nil {
					return
				}

				if cmp2(vv, x, exp) {
					return fmt.Errorf("%.0f is %s %.0f", vv, label, x)
				}
			}
		default:
			var x int //nolint:varnamelen // ok

			if x, err = strconv.Atoi(arg); err != nil {
				return
			}

			for v.Kind() == reflect.Ptr {
				if v.IsNil() {
					return
				}

				v = v.Elem()
			}

			if v.Kind() == reflect.Invalid {
				return nil
			}

			switch v.Kind() {
			case reflect.Array, reflect.String:
				if y := v.Len(); cmp2(y, x, exp) {
					return fmt.Errorf("len %d is %s %d", y, label, x)
				}
			case reflect.Map, reflect.Slice, reflect.Chan:
				if v.IsNil() {
					return
				}

				if y := v.Len(); cmp2(y, x, exp) {
					return fmt.Errorf("len %d is %s %d", y, label, x)
				}
			default:
				return fmt.Errorf("len check failed: unsupported kind %s", v.Kind())
			}
		}

		return
	}, nil
}

func cmp2[T cmp.Ordered](a, b T, exp expOutcome) bool {
	switch act := expOutcome(cmp.Compare(a, b)); exp {
	case expLess:
		return act != expLess && act != 0
	case expMore:
		return act != expMore && act != 0
	case expEq:
		return act != expEq
	default:
		return act == expEq
	}
}

func oneOf(args string) (Checker, error) {
	return Regex(fmt.Sprintf("^(%s)$", args))
}

// TODO: When this is closed, remove this:
// https://github.com/golang/go/issues/51649
//
//nolint:godox // OK
func isZero(v reflect.Value) (ok bool) {
	defer func() {
		if x := recover(); x != nil {
			ok = true
		}
	}()

	return v.IsZero()
}
