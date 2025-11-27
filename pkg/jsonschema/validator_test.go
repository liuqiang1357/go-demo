package jsonschema

import (
	"bytes"
	"encoding/json"
	"testing"

	jsonschemaLib "github.com/santhosh-tekuri/jsonschema/v5"
)

func TestValidation(t *testing.T) {
	// Define a JSON schema
	schemaStr := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"minLength": 1,
				"maxLength": 100
			},
			"age": {
				"type": "integer",
				"minimum": 0,
				"maximum": 150
			},
			"email": {
				"type": "string",
				"format": "email"
			},
			"tags": {
				"type": "array",
				"items": {
					"type": "string"
				},
				"minItems": 1
			}
		},
		"required": ["name", "age"]
	}`

	// Compile the schema
	compiler := jsonschemaLib.NewCompiler()
	if err := compiler.AddResource("schema.json", bytes.NewReader([]byte(schemaStr))); err != nil {
		t.Fatalf("Failed to add schema resource: %v", err)
	}

	schema, err := compiler.Compile("schema.json")
	if err != nil {
		t.Fatalf("Failed to compile schema: %v", err)
	}

	t.Run("Valid JSON", func(t *testing.T) {
		validJSON := `{
			"name": "John Doe",
			"age": 30,
			"email": "john@example.com",
			"tags": ["golang", "testing"]
		}`

		var validData interface{}
		if err := json.Unmarshal([]byte(validJSON), &validData); err != nil {
			t.Fatalf("Failed to unmarshal valid JSON: %v", err)
		}

		if err := schema.Validate(validData); err != nil {
			t.Errorf("Validation should pass but failed: %v", err)
		}
	})

	t.Run("Missing Required Field", func(t *testing.T) {
		invalidJSON := `{
			"name": "Jane Smith",
			"email": "jane@example.com"
		}`

		var invalidData interface{}
		if err := json.Unmarshal([]byte(invalidJSON), &invalidData); err != nil {
			t.Fatalf("Failed to unmarshal invalid JSON: %v", err)
		}

		if err := schema.Validate(invalidData); err == nil {
			t.Error("Validation should fail for missing required field 'age'")
		}
	})

	t.Run("Wrong Type", func(t *testing.T) {
		invalidJSON := `{
			"name": "Bob",
			"age": "thirty",
			"tags": []
		}`

		var invalidData interface{}
		if err := json.Unmarshal([]byte(invalidJSON), &invalidData); err != nil {
			t.Fatalf("Failed to unmarshal invalid JSON: %v", err)
		}

		if err := schema.Validate(invalidData); err == nil {
			t.Error("Validation should fail for wrong type (age should be integer)")
		}
	})

	t.Run("Constraint Violation", func(t *testing.T) {
		invalidJSON := `{
			"name": "Alice",
			"age": 200,
			"tags": ["test"]
		}`

		var invalidData interface{}
		if err := json.Unmarshal([]byte(invalidJSON), &invalidData); err != nil {
			t.Fatalf("Failed to unmarshal invalid JSON: %v", err)
		}

		if err := schema.Validate(invalidData); err == nil {
			t.Error("Validation should fail for age exceeding maximum (150)")
		}
	})

	t.Run("Empty Tags Array", func(t *testing.T) {
		invalidJSON := `{
			"name": "Test",
			"age": 25,
			"tags": []
		}`

		var invalidData interface{}
		if err := json.Unmarshal([]byte(invalidJSON), &invalidData); err != nil {
			t.Fatalf("Failed to unmarshal invalid JSON: %v", err)
		}

		if err := schema.Validate(invalidData); err == nil {
			t.Error("Validation should fail for empty tags array (minItems: 1)")
		}
	})
}

func TestDefaultValues(t *testing.T) {
	// Define a JSON schema with default values
	schemaStr := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"default": "Unknown"
			},
			"age": {
				"type": "integer",
				"default": 0
			},
			"email": {
				"type": "string",
				"default": "no-email@example.com"
			},
			"active": {
				"type": "boolean",
				"default": true
			},
			"tags": {
				"type": "array",
				"default": []
			},
			"metadata": {
				"type": "object",
				"properties": {
					"version": {
						"type": "integer",
						"default": 1
					},
					"created_at": {
						"type": "string",
						"default": "2024-01-01T00:00:00Z"
					}
				},
				"default": {}
			}
		}
	}`

	// Compile the schema
	compiler := jsonschemaLib.NewCompiler()
	if err := compiler.AddResource("schema.json", bytes.NewReader([]byte(schemaStr))); err != nil {
		t.Fatalf("Failed to add schema resource: %v", err)
	}

	schema, err := compiler.Compile("schema.json")
	if err != nil {
		t.Fatalf("Failed to compile schema: %v", err)
	}

	t.Run("Partial JSON with Missing Fields", func(t *testing.T) {
		partialJSON := `{
			"name": "John Doe",
			"age": 30
		}`

		var partialData interface{}
		if err := json.Unmarshal([]byte(partialJSON), &partialData); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		// Parse schema map
		var schemaMap map[string]interface{}
		if err := json.Unmarshal([]byte(schemaStr), &schemaMap); err != nil {
			t.Fatalf("Failed to unmarshal schema: %v", err)
		}

		// Apply defaults
		enrichedData := applyDefaults(partialData, schema, schemaMap)

		enrichedMap, ok := enrichedData.(map[string]interface{})
		if !ok {
			t.Fatal("Enriched data should be a map")
		}

		// Verify defaults are applied
		if enrichedMap["email"] != "no-email@example.com" {
			t.Errorf("Expected email default 'no-email@example.com', got %v", enrichedMap["email"])
		}
		if enrichedMap["active"] != true {
			t.Errorf("Expected active default true, got %v", enrichedMap["active"])
		}
		if enrichedMap["tags"] == nil {
			t.Error("Expected tags default empty array")
		}

		enrichedJSON, _ := json.MarshalIndent(enrichedData, "", "  ")
		t.Logf("Enriched JSON:\n%s", string(enrichedJSON))
	})

	t.Run("Empty JSON Object", func(t *testing.T) {
		emptyJSON := `{}`

		var emptyData interface{}
		if err := json.Unmarshal([]byte(emptyJSON), &emptyData); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		// Parse schema map
		var schemaMap map[string]interface{}
		if err := json.Unmarshal([]byte(schemaStr), &schemaMap); err != nil {
			t.Fatalf("Failed to unmarshal schema: %v", err)
		}

		enrichedData := applyDefaults(emptyData, schema, schemaMap)

		enrichedMap, ok := enrichedData.(map[string]interface{})
		if !ok {
			t.Fatal("Enriched data should be a map")
		}

		// Verify all defaults are applied
		if enrichedMap["name"] != "Unknown" {
			t.Errorf("Expected name default 'Unknown', got %v", enrichedMap["name"])
		}
		if enrichedMap["age"] != float64(0) {
			t.Errorf("Expected age default 0, got %v", enrichedMap["age"])
		}
		if enrichedMap["email"] != "no-email@example.com" {
			t.Errorf("Expected email default 'no-email@example.com', got %v", enrichedMap["email"])
		}
		if enrichedMap["active"] != true {
			t.Errorf("Expected active default true, got %v", enrichedMap["active"])
		}

		enrichedJSON, _ := json.MarshalIndent(enrichedData, "", "  ")
		t.Logf("Enriched JSON:\n%s", string(enrichedJSON))
	})

	t.Run("Nested Object with Defaults", func(t *testing.T) {
		nestedJSON := `{
			"name": "Jane Smith",
			"metadata": {}
		}`

		var nestedData interface{}
		if err := json.Unmarshal([]byte(nestedJSON), &nestedData); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		// Parse schema map
		var schemaMap map[string]interface{}
		if err := json.Unmarshal([]byte(schemaStr), &schemaMap); err != nil {
			t.Fatalf("Failed to unmarshal schema: %v", err)
		}

		enrichedData := applyDefaults(nestedData, schema, schemaMap)

		enrichedMap, ok := enrichedData.(map[string]interface{})
		if !ok {
			t.Fatal("Enriched data should be a map")
		}

		// Verify nested defaults are applied
		metadata, ok := enrichedMap["metadata"].(map[string]interface{})
		if !ok {
			t.Fatal("Metadata should be a map")
		}

		if metadata["version"] != float64(1) {
			t.Errorf("Expected metadata.version default 1, got %v", metadata["version"])
		}
		if metadata["created_at"] != "2024-01-01T00:00:00Z" {
			t.Errorf("Expected metadata.created_at default, got %v", metadata["created_at"])
		}

		enrichedJSON, _ := json.MarshalIndent(enrichedData, "", "  ")
		t.Logf("Enriched JSON:\n%s", string(enrichedJSON))
	})
}

