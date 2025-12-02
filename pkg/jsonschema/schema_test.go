package jsonschema

import (
	"bytes"
	"encoding/json"
	"testing"

	jsonschemaLib "github.com/santhosh-tekuri/jsonschema/v5"
)

// compileSchema is a helper to compile schema string
func compileSchema(t *testing.T, schemaStr string) *jsonschemaLib.Schema {
	compiler := jsonschemaLib.NewCompiler()
	compiler.ExtractAnnotations = true
	if err := compiler.AddResource("schema.json", bytes.NewReader([]byte(schemaStr))); err != nil {
		t.Fatalf("Failed to add schema resource: %v", err)
	}
	schema, err := compiler.Compile("schema.json")
	if err != nil {
		t.Fatalf("Failed to compile schema: %v", err)
	}
	return schema
}

// parseJSON is a helper to parse JSON string
func parseJSON(t *testing.T, jsonStr string) interface{} {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	return data
}

// assertEqual checks if two values are equal
func assertEqual(t *testing.T, name string, got, want interface{}) {
	if got != want {
		t.Errorf("%s: got %v, want %v", name, got, want)
	}
}

func TestApplyDefaults(t *testing.T) {
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
				}
			},
			"users": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"name": {
							"type": "string",
							"default": "Anonymous"
						},
						"age": {
							"type": "integer",
							"default": 0
						}
					}
				}
			}
		}
	}`

	schema := compileSchema(t, schemaStr)

	t.Run("empty object applies all defaults", func(t *testing.T) {
		data := parseJSON(t, `{}`)
		result := ApplyDefaults(data, schema)

		m, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("result should be a map")
		}

		assertEqual(t, "name", m["name"], "Unknown")
		assertEqual(t, "email", m["email"], "no-email@example.com")
		assertEqual(t, "active", m["active"], true)

		// Check numeric default (json.Number)
		if age, ok := m["age"].(json.Number); !ok || age != "0" {
			t.Errorf("age: got %v, want json.Number(\"0\")", m["age"])
		}
	})

	t.Run("partial data applies missing defaults", func(t *testing.T) {
		data := parseJSON(t, `{"name": "John"}`)
		result := ApplyDefaults(data, schema)

		m, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("result should be a map")
		}

		assertEqual(t, "name", m["name"], "John")
		assertEqual(t, "email", m["email"], "no-email@example.com")
		assertEqual(t, "active", m["active"], true)
	})

	t.Run("nested object defaults", func(t *testing.T) {
		data := parseJSON(t, `{"metadata": {}}`)
		result := ApplyDefaults(data, schema)

		m, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("result should be a map")
		}

		metadata, ok := m["metadata"].(map[string]interface{})
		if !ok {
			t.Fatal("metadata should be a map")
		}

		if version, ok := metadata["version"].(json.Number); !ok || version != "1" {
			t.Errorf("metadata.version: got %v, want json.Number(\"1\")", metadata["version"])
		}
		assertEqual(t, "metadata.created_at", metadata["created_at"], "2024-01-01T00:00:00Z")
	})

	t.Run("array item defaults", func(t *testing.T) {
		data := parseJSON(t, `{"users": [{"name": "John"}, {}]}`)
		result := ApplyDefaults(data, schema)

		m, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("result should be a map")
		}

		users, ok := m["users"].([]interface{})
		if !ok || len(users) != 2 {
			t.Fatalf("users should be array with 2 items, got %v", users)
		}

		user1, ok := users[0].(map[string]interface{})
		if !ok {
			t.Fatal("user1 should be a map")
		}
		assertEqual(t, "user1.name", user1["name"], "John")
		if age, ok := user1["age"].(json.Number); !ok || age != "0" {
			t.Errorf("user1.age: got %v, want json.Number(\"0\")", user1["age"])
		}

		user2, ok := users[1].(map[string]interface{})
		if !ok {
			t.Fatal("user2 should be a map")
		}
		assertEqual(t, "user2.name", user2["name"], "Anonymous")
		if age, ok := user2["age"].(json.Number); !ok || age != "0" {
			t.Errorf("user2.age: got %v, want json.Number(\"0\")", user2["age"])
		}
	})

	// Test with required properties
	schemaStrWithRequired := `{
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
			}
		},
		"required": ["name", "age"]
	}`

	schemaWithRequired := compileSchema(t, schemaStrWithRequired)

	t.Run("required properties are not applied with defaults", func(t *testing.T) {
		data := parseJSON(t, `{}`)
		result := ApplyDefaults(data, schemaWithRequired)

		m, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("result should be a map")
		}

		// Required properties should not have defaults applied
		if _, exists := m["name"]; exists {
			t.Error("name is required, should not have default applied")
		}
		if _, exists := m["age"]; exists {
			t.Error("age is required, should not have default applied")
		}

		// Non-required property should have default applied
		assertEqual(t, "email", m["email"], "no-email@example.com")
	})

	t.Run("non-required properties get defaults even when required ones are missing", func(t *testing.T) {
		data := parseJSON(t, `{"name": "John"}`)
		result := ApplyDefaults(data, schemaWithRequired)

		m, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("result should be a map")
		}

		// Required property provided by user
		assertEqual(t, "name", m["name"], "John")

		// Required property not provided - should not have default
		if _, exists := m["age"]; exists {
			t.Error("age is required but missing, should not have default applied")
		}

		// Non-required property should have default
		assertEqual(t, "email", m["email"], "no-email@example.com")
	})
}
