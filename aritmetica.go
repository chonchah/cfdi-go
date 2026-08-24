package cfdi

import (
	"fmt"
	"strings"
)

// decScale is the internal fixed-point scale used to parse monetary and
// quantity strings without floating point drift.
const decScale = 6

// maxFracDec is the maximum number of fractional decimals allowed in
// monetary attributes (SAT amounts are centavo precise).
const maxFracDec = 2

// parseScaled parses a plain decimal string such as "123", "123.45" or
// "-0.5" into an int64 scaled by 10^decScale. ok is false for empty,
// malformed or too precise values.
func parseScaled(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	neg := false
	switch s[0] {
	case '+':
		s = s[1:]
	case '-':
		neg = true
		s = s[1:]
	}
	intPart, fracPart := s, ""
	if i := strings.IndexByte(s, '.'); i >= 0 {
		intPart, fracPart = s[:i], s[i+1:]
	}
	if intPart == "" && fracPart == "" {
		return 0, false
	}
	if len(fracPart) > decScale {
		return 0, false // more precision than we support
	}
	for _, r := range intPart + fracPart {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	for len(fracPart) < decScale {
		fracPart += "0"
	}
	digits := intPart + fracPart
	var v int64
	for i := 0; i < len(digits); i++ {
		d := int64(digits[i] - '0')
		if v > (int64(^uint(0)>>1)-d)/10 {
			return 0, false // overflow
		}
		v = v*10 + d
	}
	if neg {
		v = -v
	}
	return v, true
}

// hasMaxDec reports whether s has at most n fractional decimals (and is a
// plain non-negative-or-signed decimal).
func hasMaxDec(s string, n int) bool {
	s = strings.TrimSpace(s)
	i := strings.IndexByte(s, '.')
	return i < 0 || len(s)-i-1 <= n
}

// roundToCents converts a value scaled by 10^decScale into cents using
// ROUND_HALF_UP on the absolute value (the rounding style used by the SAT).
func roundToCents(scaled int64) int64 {
	const f = decScale - maxFracDec // digits dropped: 4
	const div = int64(10000)
	q, r := scaled/div, scaled%div
	if r < 0 {
		r = -r
	}
	if r*2 >= div {
		if scaled >= 0 {
			q++
		} else {
			q--
		}
	}
	return q
}

// mulRoundCents returns round_half_up(a × b) expressed in cents, where a
// and b are values scaled by 10^decScale (for example amount × rate).
func mulRoundCents(a, b int64) int64 {
	prod := a * b // scale = 2*decScale = 12
	const div = int64(10000000000) // 10^(12-2)
	q, r := prod/div, prod%div
	if r < 0 {
		r = -r
	}
	if r*2 >= div {
		if prod >= 0 {
			q++
		} else {
			q--
		}
	}
	return q
}

// amountCents parses an amount string into cents (ROUND_HALF_UP when it
// carries extra precision). Malformed input yields 0; Validate is the gate
// responsible for reporting malformed amounts.
func amountCents(s string) int64 {
	v, ok := parseScaled(s)
	if !ok {
		return 0
	}
	return roundToCents(v)
}

// formatCents renders cents as a plain decimal string with exactly two
// fraction digits, e.g. -3 -> "-0.03".
func formatCents(c int64) string {
	sign := ""
	if c < 0 {
		sign = "-"
		c = -c
	}
	return fmt.Sprintf("%s%d.%02d", sign, c/100, c%100)
}

// isZeroAmount reports whether s is a valid representation of zero
// ("", "0", "0.00", "-0.000"...).
func isZeroAmount(s string) bool {
	v, ok := parseScaled(s)
	return ok && v == 0
}
