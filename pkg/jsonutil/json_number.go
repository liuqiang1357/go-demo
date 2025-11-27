package jsonutil

// This package provides helpers for JSON decoding with better number handling.
// In particular, it allows decoding JSON numbers into int64 when they are
// "safe" integers (no fractional part and within the int64 range), instead of
// always using float64 for numbers decoded into interface{}.

import (
	"bytes"
	"encoding/json"
	"strings"
)

// UnmarshalWithInt decodes JSON into an interface{} value, but represents
// numeric values as int64 when they are safe integers (no decimal point or
// exponent and within the int64 range). Other numeric values are decoded as
// float64.
//
// This is useful when you want to avoid losing integer precision that would
// otherwise be represented as float64 when using the standard json.Unmarshal
// into interface{}.
func UnmarshalWithInt(data []byte) (interface{}, error) {
	var v interface{}

	dec := json.NewDecoder(bytes.NewReader(data))
	// UseNumber ensures that numbers are initially decoded as json.Number
	// so that we can decide whether to treat them as int64 or float64.
	dec.UseNumber()

	if err := dec.Decode(&v); err != nil {
		return nil, err
	}

	return convertNumbers(v), nil
}

// convertNumbers walks the decoded JSON value and converts any json.Number
// instances into either int64 (for safe integers) or float64.
func convertNumbers(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(val))
		for k, vv := range val {
			out[k] = convertNumbers(vv)
		}
		return out
	case []interface{}:
		for i, vv := range val {
			val[i] = convertNumbers(vv)
		}
		return val
	case json.Number:
		return convertNumber(val)
	default:
		return v
	}
}

// convertNumber converts a json.Number into either int64 or float64.
// It treats a number as a "safe" integer if:
//   - it does not contain a decimal point or exponent, and
//   - it can be parsed into an int64 without overflow.
//
// In that case it returns int64; otherwise it returns float64.
func convertNumber(n json.Number) interface{} {
	s := n.String()

	// If the string representation contains a decimal point or exponent,
	// treat it as a floating-point number.
	if strings.ContainsAny(s, ".eE") {
		if f, err := n.Float64(); err == nil {
			return f
		}
		// Fallback: return the original string if parsing fails.
		return s
	}

	// Try to parse as int64 for integer-like strings.
	if i, err := n.Int64(); err == nil {
		return i
	}

	// If it cannot be parsed as int64 (e.g. too large), fall back to float64.
	if f, err := n.Float64(); err == nil {
		return f
	}

	// As a last resort, return the raw string.
	return s
}
