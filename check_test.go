package vali

import (
	"reflect"
	"testing"
)

func TestEmail(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"Valid email", "test@example.com", false},
		{"Valid email with subdomain", "test@sub.example.com", false},
		{"Valid email with plus", "test+tag@example.com", false},
		{"Valid email with dots", "first.last@example.com", false},
		{"Missing @", "testexample.com", true},
		{"Missing domain", "test@", true},
		{"Missing local part", "@example.com", true},
		{"Invalid format", "test@test@example.com", true}, // Changed to a definitely invalid format.
		{"Numeric", 12345, true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := email(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("email() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestURL(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"Valid HTTP URL", "http://example.com", false},
		{"Valid HTTPS URL", "https://example.com", false},
		{"Valid URL with path", "https://example.com/path", false},
		{"Valid URL with query", "https://example.com/path?query=value", false},
		{"Valid URL with port", "https://example.com:8080", false},
		{"Missing scheme", "example.com", true},
		{"Invalid URL", "htt:/example.com", true},
		{"Invalid URL", "\x12", true},
		{"Missing host", "http://", true},
		{"Numeric", 12345, true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := urL(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("url_() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIPv4(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"Valid IPv4", "192.168.1.1", false},
		{"Valid IPv4 zeros", "0.0.0.0", false},
		{"Valid IPv4 broadcast", "255.255.255.255", false},
		{"Invalid IPv4 format", "192.168.1", true},
		{"Invalid IPv4 values", "256.256.256.256", true},
		{"IPv6 address", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"Not an IP", "not-an-ip", true},
		{"Numeric", 12345, true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ipv4(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ipv4() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIPv6(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"Valid IPv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
		{"Valid IPv6 shortened", "2001:db8:85a3::8a2e:370:7334", false},
		{"Valid IPv6 loopback", "::1", false},
		{"Valid IPv6 unspecified", "::", false},
		{"IPv4 address", "192.168.1.1", true},
		{"Invalid IPv6 format", "2001:0db8:85a3", true},
		{"Not an IP", "not-an-ip", true},
		{"Numeric", 12345, true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ipv6(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ipv6() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIP(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"Valid IPv4", "192.168.1.1", false},
		{"Valid IPv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
		{"Valid IPv6 shortened", "2001:db8:85a3::8a2e:370:7334", false},
		{"Invalid IP format", "192.168.1", true},
		{"Not an IP", "not-an-ip", true},
		{"Numeric", 12345, true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ip(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ip() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMAC(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"Valid MAC", "01:23:45:67:89:ab", false},
		{"Valid MAC uppercase", "01:23:45:67:89:AB", false},
		{"Valid MAC with dashes", "01-23-45-67-89-ab", false},
		{"Valid MAC dot notation", "0123.4567.89ab", false},
		{"Invalid MAC too short", "01:23:45:67:89", true},
		{"Invalid MAC too long", "01:23:45:67:89:ab:cd", true},
		{"Invalid MAC format", "01:23:45:67:89:zz", true},
		{"Not a MAC", "not-a-mac", true},
		{"Numeric", 12345, true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := mac(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("mac() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDomain(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"Valid domain", "example.com", false},
		{"Valid subdomain", "sub.example.com", false},
		{"Valid multilevel", "a.b.c.example.com", false},
		{"Valid with numbers", "example123.com", false},
		{"Valid with hyphens", "ex-am-ple.com", false},
		{"Too short", "a.co", false},
		{"Invalid TLD too short", "example.c", true},
		{"Invalid chars", "ex@mple.com", true},
		{"Invalid format", ".example.com", true},
		{"Invalid format leading hyphen", "-example.com", true},
		{"Invalid format trailing hyphen", "example-.com", true},
		{"Not a domain", "not_a_domain", true},
		{"Numeric", 12345, true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := domain(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("domain() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestISBN(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"Valid ISBN-10", "0-306-40615-2", false},
		{"Valid ISBN-10 no dashes", "0306406152", false},
		{"Valid ISBN-10 with X", "0-8044-2957-X", false},
		{"Valid ISBN-13", "978-3-16-148410-0", false},
		{"Valid ISBN-13 no dashes", "9783161484100", false},
		{"Valid ISBN-13 no dashes, numeric", 9783161484100, false},
		{"Invalid ISBN-10 checksum", "0-306-40615-3", true},
		{"Invalid ISBN-10 last char", "0-306-40615-Y", true},
		{"Invalid ISBN-13 checksum", "978-3-16-148410-1", true},
		{"Invalid length", "978-3-16-148410", true},
		{"Invalid chars", "978-3-16-14841A-0", true},
		{"Not an ISBN", "not-an-isbn", true},
		{"Numeric invalid", 12345, true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := isbn(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("isbn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAlpha(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"Lowercase letters", "abcdef", false},
		{"Uppercase letters", "ABCDEF", false},
		{"Mixed case", "AbCdEf", false},
		{"With numbers", "abc123", true},
		{"With symbols", "abc!", true},
		{"With spaces", "abc def", true},
		{"Numeric", 12345, true},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := alpha(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("alpha() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAlphaNum(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"Lowercase letters", "abcdef", false},
		{"Uppercase letters", "ABCDEF", false},
		{"Numbers only", "123456", false},
		{"Mixed alphanumeric", "abc123DEF", false},
		{"With symbols", "abc123!", true},
		{"With spaces", "abc 123", true},
		{"Numeric", 12345, false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := alphaNum(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("alphaNum() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNumeric(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"Digits only", "123456", false},
		{"With letters", "123abc", true},
		{"With symbols", "123!", true},
		{"With spaces", "123 456", true},
		{"Numeric", 12345, false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := numeric(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("numeric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBoolean(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"true", "true", false},
		{"True uppercase", "True", false},
		{"TRUE all caps", "TRUE", false},
		{"t", "t", false},
		{"1", "1", false},
		{"yes", "yes", false},
		{"y", "y", false},
		{"on", "on", false},
		{"false", "false", false},
		{"False uppercase", "False", false},
		{"FALSE all caps", "FALSE", false},
		{"f", "f", false},
		{"0", "0", false},
		{"no", "no", false},
		{"n", "n", false},
		{"off", "off", false},
		{"Invalid bool", "not-a-bool", true},
		{"Invalid bool 2", "2", true},
		{"Numeric invalid", 12345, true},
		{"Numeric true", 1, false},
		{"Numeric false", 0, false},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := boolean(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("boolean() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreditCard(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"Valid Visa", "4111 1111 1111 1111", false},
		{"Valid Mastercard", "5500 0000 0000 0004", false},
		{"Valid Amex", "3400 0000 0000 009", false},
		{"Valid with dashes", "4111-1111-1111-1111", false},
		{"Valid no spaces", "4111111111111111", false},
		{"Valid no spaces, numeric", 4111111111111111, false},
		{"Invalid checksum", "4111 1111 1111 1112", true},
		{"Invalid length too short", "4111 1111 1111", true},
		{"Invalid length too long", "4111 1111 1111 1111 1111", true},
		{"Invalid chars", "4111 1111 1111 111a", true},
		{"Not a credit card", "not-a-credit-card", true},
		{"Numeric invalid", 12345, true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := creditCard(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("creditCard() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJSON(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"Valid object", `{"key": "value"}`, false},
		{"Valid array", `[1, 2, 3]`, false},
		{"Valid string", `"hello"`, false},
		{"Valid number", `42`, false},
		{"Valid boolean", `true`, false},
		{"Valid null", `null`, false},
		{"Invalid syntax missing quote", `{"key: "value"}`, true},
		{"Invalid syntax missing bracket", `{"key": "value"`, true},
		{"Invalid syntax trailing comma", `[1, 2, 3,]`, true},
		{"Not JSON", "not-json", true},
		{"Numeric", 12345, false},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := jsoN(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("json_() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestASCII(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"ASCII letters", "abcdefABCDEF", false},
		{"ASCII digits", "0123456789", false},
		{"ASCII symbols", "!@#$%^&*()", false},
		{"ASCII mixed", "Hello, World! 123", false},
		{"Non-ASCII", "Héllö", true},
		{"Emoji", "👋", true},
		{"Numeric", 12345, false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ascii(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ascii() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLowercase(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"Lowercase letters", "abcdef", false},
		{"Lowercase with digits", "abcdef123", false},
		{"Lowercase with symbols", "abcdef!@#", false},
		{"With uppercase", "abcDef", true},
		{"All uppercase", "ABCDEF", true},
		{"Numeric", 12345, false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := lowercase(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("lowercase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUppercase(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"Uppercase letters", "ABCDEF", false},
		{"Uppercase with digits", "ABCDEF123", false},
		{"Uppercase with symbols", "ABCDEF!@#", false},
		{"With lowercase", "ABCdEF", true},
		{"All lowercase", "abcdef", true},
		{"Numeric", 12345, false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := uppercase(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("uppercase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateISBN10(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid ISBN-10", "0306406152", false},
		{"Valid ISBN-10 with X", "080442957X", false},
		{"Invalid checksum", "0306406153", true},
		{"Invalid character", "03064061A2", true},
		{"Last character invalid", "03064061AX", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateISBN10(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateISBN10() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateISBN13(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid ISBN-13", "9783161484100", false},
		{"Invalid checksum", "9783161484101", true},
		{"Invalid character", "978316148410A", true},
		{"Last character invalid", "978316148410A", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateISBN13(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateISBN13() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRGB(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"Valid RGB", "rgb(255,255,255)", false},
		{"Valid RGB", "rgb(155,255,255)", false},
		{"Valid RGB", "rgb(255,55,255)", false},
		{"Valid RGB", "rgb(255,255,5)", false},
		{"Valid RGB with zeros", "rgb(0,0,0)", false},
		{"Invalid RGB out of range", "rgb(256,255,255)", true},
		{"Invalid RGB format", "rgb(255,255)", true},
		{"Invalid RGB with spaces", "rgb(255, 255, 255)", true},
		{"Invalid RGB mixed spaces", "rgb(255,0, 128)", true},
		{"Invalid RGB no parentheses", "rgb255,255,255", true},
		{"Invalid RGB wrong scheme", "rgba(255,255,255)", true},
		{"Invalid RGB too many values", "rgb(255,255,255,0)", true},
		{"Invalid RGB negative value", "rgb(-1,255,255)", true},
		{"Invalid RGB non-numeric", "rgb(a,b,c)", true},
		{"Numeric", 12345, true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := rgb(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("rgb() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRGBA(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"Valid RGBA", "rgba(255,255,255,0.5)", false},
		{"Valid RGBA with zeros", "rgba(0,0,0,1)", false},
		{"Invalid RGBA out of range", "rgba(256,255,255,0.5)", true},
		{"Invalid RGBA opacity", "rgba(255,255,255,1.5)", true},
		{"Invalid RGBA format", "rgba(255,255,255)", true},
		{"Invalid RGBA with spaces", "rgba(255, 255, 255, 0.5)", true},
		{"Invalid RGBA mixed spaces", "rgba(255,0, 128, 0.75)", true},
		{"Valid RGBA min values", "rgba(0,0,0,0)", false},
		{"Valid RGBA max values", "rgba(255,255,255,1)", false},
		{"Invalid RGBA no parentheses", "rgba255,255,255,0.5", true},
		{"Invalid RGBA wrong scheme", "rgb(255,255,255,0.5)", true},
		{"Invalid RGBA too few values", "rgba(255,255,255)", true},
		{"Invalid RGBA negative value", "rgba(-1,255,255,0.5)", true},
		{"Invalid RGBA non-numeric", "rgba(a,b,c,d)", true},
		{"Invalid RGBA opacity format", "rgba(255,255,255,50%)", true},
		{"Numeric", 12345, true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := rgba(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("rgba() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLuhn(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"Valid Luhn string", "4111111111111111", false},
		{"Valid Luhn number", 4111111111111111, false},
		{"Invalid Luhn string", "4111111111111112", true},
		{"Invalid Luhn number", 4111111111111112, true},
		{"Valid Luhn with spaces", "4111 1111 1111 1111", false},
		{"Valid Luhn with dashes", "4111-1111-1111-1111", false},
		{"Invalid characters", "4111-1111-1111-111a", true},
		{"Empty string", "", true},
		{"Valid Luhn int", 4111111111111111, false},
		{"Invalid Luhn int", 4111111111111112, true},
		{"Valid Luhn int64", int64(4111111111111111), false},
		{"Valid Luhn uint", uint(4242424242424242), false},
		{"Valid Luhn float64", float64(4111111111111111), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := luhn(reflect.ValueOf(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("luhn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSSN(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"Valid SSN", "123-45-6789", false},
		{"Invalid SSN format", "123456789", true},
		{"Invalid SSN with letters", "123-45-67a9", true},
		{"Numeric", 12345, true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ssn(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ssn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNPI(t *testing.T) {
	t.Parallel()

	tests := []struct { //nolint:govet // ok
		name    string
		input   any
		wantErr bool
	}{
		{"Valid NPI", "1234567893", false},
		{"Valid NPI numeric", 1234567893, false},
		{"Invalid NPI checksum", "1234567890", true},
		{"Invalid NPI checksum, numeric", 1234567890, true},
		{"Invalid NPI length", "123456789", true},
		{"Invalid NPI with letters", "12345678a3", true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := npi(val(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("npi() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func val[T any](s T) reflect.Value {
	return reflect.ValueOf(s)
}
