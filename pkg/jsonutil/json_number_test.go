package jsonutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/flosch/pongo2/v6"
)

func TestUnmarshalWithInt_MixedNumbers(t *testing.T) {
	jsonStr := `{
		"int_small": 1,
		"int_large": 9223372036854775807,
		"too_large": 9223372036854775808,
		"float_simple": 1.5,
		"float_exp": 1e3,
		"nested": {
			"int": 42,
			"float": 3.14
		},
		"array": [0, 2.5, 100]
	}`

	v, err := UnmarshalWithInt([]byte(jsonStr))
	if err != nil {
		t.Fatalf("UnmarshalWithInt failed: %v", err)
	}

	root, ok := v.(map[string]interface{})
	if !ok {
		t.Fatalf("expected root to be map[string]interface{}, got %T", v)
	}

	// Safe integers should be int64
	if got := root["int_small"]; reflect.TypeOf(got).Kind() != reflect.Int64 {
		t.Fatalf("int_small should be int64, got %T (%v)", got, got)
	}
	if got := root["int_large"]; reflect.TypeOf(got).Kind() != reflect.Int64 {
		t.Fatalf("int_large should be int64, got %T (%v)", got, got)
	}

	// Too large integer should fall back to float64
	if got := root["too_large"]; reflect.TypeOf(got).Kind() != reflect.Float64 {
		t.Fatalf("too_large should be float64, got %T (%v)", got, got)
	}

	// Floats should remain float64
	if got := root["float_simple"]; reflect.TypeOf(got).Kind() != reflect.Float64 {
		t.Fatalf("float_simple should be float64, got %T (%v)", got, got)
	}
	if got := root["float_exp"]; reflect.TypeOf(got).Kind() != reflect.Float64 {
		t.Fatalf("float_exp should be float64, got %T (%v)", got, got)
	}

	// Nested values
	nested, ok := root["nested"].(map[string]interface{})
	if !ok {
		t.Fatalf("nested should be map[string]interface{}, got %T", root["nested"])
	}
	if got := nested["int"]; reflect.TypeOf(got).Kind() != reflect.Int64 {
		t.Fatalf("nested.int should be int64, got %T (%v)", got, got)
	}
	if got := nested["float"]; reflect.TypeOf(got).Kind() != reflect.Float64 {
		t.Fatalf("nested.float should be float64, got %T (%v)", got, got)
	}

	// Array values
	arr, ok := root["array"].([]interface{})
	if !ok {
		t.Fatalf("array should be []interface{}, got %T", root["array"])
	}
	if got := arr[0]; reflect.TypeOf(got).Kind() != reflect.Int64 {
		t.Fatalf("array[0] should be int64, got %T (%v)", got, got)
	}
	if got := arr[1]; reflect.TypeOf(got).Kind() != reflect.Float64 {
		t.Fatalf("array[1] should be float64, got %T (%v)", got, got)
	}
	if got := arr[2]; reflect.TypeOf(got).Kind() != reflect.Int64 {
		t.Fatalf("array[2] should be int64, got %T (%v)", got, got)
	}
}

func TestUnmarshalWithInt_CompatibilityWithStdlib(t *testing.T) {
	jsonStr := `{"a": 1, "b": 2.5}`

	// Standard library behavior for comparison
	var std interface{}
	if err := json.Unmarshal([]byte(jsonStr), &std); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Our custom behavior
	custom, err := UnmarshalWithInt([]byte(jsonStr))
	if err != nil {
		t.Fatalf("UnmarshalWithInt failed: %v", err)
	}

	stdMap := std.(map[string]interface{})
	customMap := custom.(map[string]interface{})

	// Standard library uses float64 for both
	if _, ok := stdMap["a"].(float64); !ok {
		t.Fatalf("stdlib: expected a to be float64, got %T", stdMap["a"])
	}
	if _, ok := stdMap["b"].(float64); !ok {
		t.Fatalf("stdlib: expected b to be float64, got %T", stdMap["b"])
	}

	// Custom decoder should use int64 for integer and float64 for float
	if _, ok := customMap["a"].(int64); !ok {
		t.Fatalf("custom: expected a to be int64, got %T", customMap["a"])
	}
	if _, ok := customMap["b"].(float64); !ok {
		t.Fatalf("custom: expected b to be float64, got %T", customMap["b"])
	}
}

// TestJSONNumberBehavior tests how json.Number behaves in different scenarios:
// 1. Converting to string
// 2. JSON marshaling (serialization)
// 3. Template engine rendering (pongo2)
func TestJSONNumberBehavior(t *testing.T) {
	jsonStr := `{
		"int_val": 42,
		"float_val": 3.14,
		"large_int": 9223372036854775807,
		"exp_val": 1e3
	}`

	// Decode with UseNumber to get json.Number instances
	var vWithNumber map[string]interface{}
	dec := json.NewDecoder(bytes.NewReader([]byte(jsonStr)))
	dec.UseNumber()
	if err := dec.Decode(&vWithNumber); err != nil {
		t.Fatalf("Failed to decode with UseNumber: %v", err)
	}

	// Decode with our custom function to get int64/float64
	vConverted, err := UnmarshalWithInt([]byte(jsonStr))
	if err != nil {
		t.Fatalf("UnmarshalWithInt failed: %v", err)
	}
	vConvertedMap := vConverted.(map[string]interface{})

	t.Run("String Conversion", func(t *testing.T) {
		// json.Number.String() returns the original string representation
		intNum := vWithNumber["int_val"].(json.Number)
		floatNum := vWithNumber["float_val"].(json.Number)
		expNum := vWithNumber["exp_val"].(json.Number)

		t.Logf("json.Number.String() for 42: %q", intNum.String())
		t.Logf("json.Number.String() for 3.14: %q", floatNum.String())
		t.Logf("json.Number.String() for 1e3: %q", expNum.String())

		if intNum.String() != "42" {
			t.Errorf("Expected '42', got %q", intNum.String())
		}
		if floatNum.String() != "3.14" {
			t.Errorf("Expected '3.14', got %q", floatNum.String())
		}
		if expNum.String() != "1e3" {
			t.Errorf("Expected '1e3', got %q", expNum.String())
		}

		// Converted types: int64 and float64 use fmt.Sprintf
		intVal := vConvertedMap["int_val"].(int64)
		floatVal := vConvertedMap["float_val"].(float64)
		t.Logf("int64(42) as string: %q", fmt.Sprintf("%d", intVal))
		t.Logf("float64(3.14) as string: %q", fmt.Sprintf("%g", floatVal))
	})

	t.Run("JSON Marshaling", func(t *testing.T) {
		// json.Number marshals to its original string representation
		withNumberJSON, err := json.Marshal(vWithNumber)
		if err != nil {
			t.Fatalf("Failed to marshal with json.Number: %v", err)
		}

		// Converted types: int64 marshals as integer, float64 as float
		convertedJSON, err := json.Marshal(vConvertedMap)
		if err != nil {
			t.Fatalf("Failed to marshal converted values: %v", err)
		}

		t.Logf("Marshaled with json.Number:\n%s", string(withNumberJSON))
		t.Logf("Marshaled with int64/float64:\n%s", string(convertedJSON))

		// Both should produce valid JSON, but formatting may differ
		var check1, check2 interface{}
		if err := json.Unmarshal(withNumberJSON, &check1); err != nil {
			t.Errorf("json.Number marshaled JSON is invalid: %v", err)
		}
		if err := json.Unmarshal(convertedJSON, &check2); err != nil {
			t.Errorf("Converted values marshaled JSON is invalid: %v", err)
		}

		// json.Number preserves original format (e.g., "1e3" stays as "1e3")
		// int64/float64 may format differently (e.g., 1e3 becomes 1000 or 1e+03)
		if !strings.Contains(string(withNumberJSON), `"int_val":42`) {
			t.Error("json.Number should marshal integer as number")
		}
		if !strings.Contains(string(convertedJSON), `"int_val":42`) {
			t.Error("int64 should marshal as integer number")
		}
	})

	t.Run("Template Engine Rendering", func(t *testing.T) {
		// Test with pongo2 template engine
		templateStr := `Integer: {{ int_val }}, Float: {{ float_val }}, Large: {{ large_int }}, Exp: {{ exp_val }}`

		tpl, err := pongo2.FromString(templateStr)
		if err != nil {
			t.Fatalf("Failed to parse template: %v", err)
		}

		// Test with json.Number
		ctxWithNumber := pongo2.Context{
			"int_val":   vWithNumber["int_val"],
			"float_val": vWithNumber["float_val"],
			"large_int": vWithNumber["large_int"],
			"exp_val":   vWithNumber["exp_val"],
		}

		outputWithNumber, err := tpl.Execute(ctxWithNumber)
		if err != nil {
			t.Fatalf("Failed to execute template with json.Number: %v", err)
		}

		// Test with converted int64/float64
		ctxConverted := pongo2.Context{
			"int_val":   vConvertedMap["int_val"],
			"float_val": vConvertedMap["float_val"],
			"large_int": vConvertedMap["large_int"],
			"exp_val":   vConvertedMap["exp_val"],
		}

		outputConverted, err := tpl.Execute(ctxConverted)
		if err != nil {
			t.Fatalf("Failed to execute template with converted types: %v", err)
		}

		t.Logf("Template output with json.Number:\n%s", outputWithNumber)
		t.Logf("Template output with int64/float64:\n%s", outputConverted)

		// Both should render the numbers, but format might differ
		if !strings.Contains(outputWithNumber, "42") {
			t.Error("Template should render json.Number integer")
		}
		if !strings.Contains(outputConverted, "42") {
			t.Error("Template should render int64 integer")
		}

		// json.Number might preserve original format (e.g., "1e3")
		// while float64 might render as "1000" or "1e+03"
		if !strings.Contains(outputWithNumber, "1e3") && !strings.Contains(outputWithNumber, "1000") {
			t.Error("Template should render exponential number")
		}
	})

	t.Run("Comparison Summary", func(t *testing.T) {
		t.Log("\n=== Behavior Summary ===")
		t.Log("json.Number:")
		t.Log("  - String(): Returns original JSON string representation")
		t.Log("  - Marshal: Preserves original format (e.g., '1e3' stays '1e3')")
		t.Log("  - Template: Renders as original string representation")
		t.Log("\nint64/float64 (converted):")
		t.Log("  - String(): Uses Go's default formatting")
		t.Log("  - Marshal: Uses standard JSON number formatting")
		t.Log("  - Template: Renders using Go's string conversion")
		t.Log("\nKey difference: json.Number preserves original JSON format,")
		t.Log("while int64/float64 use Go's native formatting.")
	})
}
