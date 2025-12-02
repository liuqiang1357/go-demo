package jsonschema

import (
	"bytes"
	"encoding/json"
	"testing"

	jsonschemaLib "github.com/santhosh-tekuri/jsonschema/v5"
)

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

func parseJSON(t *testing.T, jsonStr string) interface{} {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	return data
}

func TestApplyDefaults_Basic(t *testing.T) {
	schemaStr := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"name": {"type": "string", "default": "Unknown"},
			"email": {"type": "string", "default": "no-email@example.com"},
			"metadata": {
				"type": "object",
				"properties": {
					"version": {"type": "integer", "default": 1}
				}
			},
			"users": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"name": {"type": "string", "default": "Anonymous"}
					}
				}
			}
		}
	}`
	schema := compileSchema(t, schemaStr)

	// Empty object gets defaults
	data := parseJSON(t, `{}`)
	result := ApplyDefaults(data, schema)
	m := result.(map[string]interface{})
	if m["name"] != "Unknown" || m["email"] != "no-email@example.com" {
		t.Errorf("Empty object should get defaults")
	}

	// Partial data gets missing defaults
	data = parseJSON(t, `{"name": "John"}`)
	result = ApplyDefaults(data, schema)
	m = result.(map[string]interface{})
	if m["name"] != "John" || m["email"] != "no-email@example.com" {
		t.Errorf("Partial data should get missing defaults")
	}

	// Nested objects get defaults
	data = parseJSON(t, `{"metadata": {}}`)
	result = ApplyDefaults(data, schema)
	m = result.(map[string]interface{})
	metadata := m["metadata"].(map[string]interface{})
	if version, ok := metadata["version"].(json.Number); !ok || version != "1" {
		t.Errorf("Nested object should get defaults")
	}

	// Array items get defaults
	data = parseJSON(t, `{"users": [{}]}`)
	result = ApplyDefaults(data, schema)
	m = result.(map[string]interface{})
	users := m["users"].([]interface{})
	user := users[0].(map[string]interface{})
	if user["name"] != "Anonymous" {
		t.Errorf("Array items should get defaults")
	}
}

func TestApplyDefaults_TupleItems(t *testing.T) {
	schemaStr := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"tuple": {
				"type": "array",
				"items": [
					{
						"type": "object",
						"properties": {
							"a": {"type": "string", "default": "A"}
						}
					},
					{
						"type": "object",
						"properties": {
							"b": {"type": "string", "default": "B"}
						}
					}
				]
			}
		}
	}`
	schema := compileSchema(t, schemaStr)

	// First element uses first item schema, second and later use last schema
	data := parseJSON(t, `{"tuple": [{}, {}, {}]}`)
	result := ApplyDefaults(data, schema)
	m := result.(map[string]interface{})
	tuple := m["tuple"].([]interface{})

	first := tuple[0].(map[string]interface{})
	if first["a"] != "A" {
		t.Errorf("First tuple element should get default a=A, got %#v", first)
	}
	second := tuple[1].(map[string]interface{})
	if second["b"] != "B" {
		t.Errorf("Second tuple element should get default b=B, got %#v", second)
	}
	third := tuple[2].(map[string]interface{})
	if third["b"] != "B" {
		t.Errorf("Third tuple element should reuse last item schema and get b=B, got %#v", third)
	}
}

func TestApplyDefaults_RequiredProperties(t *testing.T) {
	schemaStr := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"name": {"type": "string", "default": "Unknown"},
			"email": {"type": "string", "default": "no-email@example.com"}
		},
		"required": ["name"]
	}`
	schema := compileSchema(t, schemaStr)

	data := parseJSON(t, `{}`)
	result := ApplyDefaults(data, schema)
	m := result.(map[string]interface{})

	if _, exists := m["name"]; exists {
		t.Error("Required property should not get default")
	}
	if m["email"] != "no-email@example.com" {
		t.Error("Non-required property should get default")
	}
}

func TestApplyDefaults_Ref(t *testing.T) {
	schemaStr := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"definitions": {
			"address": {
				"type": "object",
				"properties": {
					"street": {"type": "string", "default": "Main St"},
					"city": {"type": "string", "default": "Unknown"}
				}
			}
		},
		"type": "object",
		"properties": {
			"address": {"$ref": "#/definitions/address"},
			"name": {"type": "string", "default": "John"}
		}
	}`
	schema := compileSchema(t, schemaStr)

	data := parseJSON(t, `{"address": {}}`)
	result := ApplyDefaults(data, schema)
	m := result.(map[string]interface{})

	address := m["address"].(map[string]interface{})
	if address["street"] != "Main St" || address["city"] != "Unknown" {
		t.Error("$ref should resolve and apply defaults")
	}
	if m["name"] != "John" {
		t.Error("Base schema properties should work with $ref")
	}
}

func TestApplyDefaults_AllOf(t *testing.T) {
	schemaStr := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"allOf": [
			{"properties": {"name": {"type": "string", "default": "Unknown"}}},
			{"properties": {"email": {"type": "string", "default": "no-email@example.com"}}}
		],
		"properties": {"active": {"type": "boolean", "default": true}}
	}`
	schema := compileSchema(t, schemaStr)

	data := parseJSON(t, `{}`)
	result := ApplyDefaults(data, schema)
	m := result.(map[string]interface{})

	if m["name"] != "Unknown" || m["email"] != "no-email@example.com" || m["active"] != true {
		t.Error("allOf should merge defaults from all subschemas")
	}
}

func TestApplyDefaults_AnyOf(t *testing.T) {
	schemaStr := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"anyOf": [
			{"properties": {"type": {"type": "string", "enum": ["person"], "default": "person"}}},
			{"properties": {"type": {"type": "string", "enum": ["company"], "default": "company"}}}
		]
	}`
	schema := compileSchema(t, schemaStr)

	// Data matching first schema
	data := parseJSON(t, `{"type": "person"}`)
	result := ApplyDefaults(data, schema)
	m := result.(map[string]interface{})
	if m["type"] != "person" {
		t.Error("anyOf should apply defaults from matching schema")
	}
}

func TestApplyDefaults_OneOf(t *testing.T) {
	schemaStr := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"oneOf": [
			{"properties": {"type": {"type": "string", "enum": ["student"], "default": "student"}, "studentId": {"type": "string", "default": "S001"}}},
			{"properties": {"type": {"type": "string", "enum": ["teacher"], "default": "teacher"}, "teacherId": {"type": "string", "default": "T001"}}}
		]
	}`
	schema := compileSchema(t, schemaStr)

	// Data matching first schema
	data := parseJSON(t, `{"type": "student"}`)
	result := ApplyDefaults(data, schema)
	m := result.(map[string]interface{})
	if m["type"] != "student" || m["studentId"] != "S001" {
		t.Error("oneOf should apply defaults from matching schema")
	}
	if _, exists := m["teacherId"]; exists {
		t.Error("oneOf should not apply defaults from non-matching schema")
	}
}

func TestApplyDefaults_Combined(t *testing.T) {
	schemaStr := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"definitions": {
			"base": {
				"type": "object",
				"properties": {"id": {"type": "integer", "default": 0}}
			}
		},
		"type": "object",
		"allOf": [
			{"$ref": "#/definitions/base"},
			{"properties": {"name": {"type": "string", "default": "Item"}}}
		],
		"properties": {"active": {"type": "boolean", "default": true}}
	}`
	schema := compileSchema(t, schemaStr)

	data := parseJSON(t, `{}`)
	result := ApplyDefaults(data, schema)
	m := result.(map[string]interface{})

	if id, ok := m["id"].(json.Number); !ok || id != "0" {
		t.Error("allOf with $ref should apply defaults from both")
	}
	if m["name"] != "Item" || m["active"] != true {
		t.Error("Combined schemas should merge defaults correctly")
	}
}
