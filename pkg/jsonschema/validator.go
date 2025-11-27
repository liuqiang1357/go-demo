package jsonschema

import (
	"bytes"
	"encoding/json"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// This package provides JSON schema validation and default value application utilities.
// See validator_test.go for usage examples and tests.

// applyDefaults recursively applies default values from schema to JSON data
func applyDefaults(data interface{}, schema *jsonschema.Schema, schemaMap map[string]interface{}) interface{} {
	if schemaMap == nil {
		return data
	}

	properties, ok := schemaMap["properties"].(map[string]interface{})
	if !ok {
		return data
	}

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return data
	}

	result := make(map[string]interface{})

	// Copy existing data
	for k, v := range dataMap {
		result[k] = v
	}

	// Apply defaults for missing properties
	for propName, propSchema := range properties {
		propMap, ok := propSchema.(map[string]interface{})
		if !ok {
			continue
		}

		if _, exists := result[propName]; !exists {
			// Property doesn't exist, apply default
			if defaultValue, hasDefault := propMap["default"]; hasDefault {
				result[propName] = defaultValue
			} else if propType, hasType := propMap["type"].(string); hasType {
				// Apply type-specific defaults
				switch propType {
				case "object":
					// For nested objects, recursively apply defaults
					if nestedSchemaMap, ok := propMap["properties"].(map[string]interface{}); ok {
						nestedData := applyDefaults(map[string]interface{}{}, nil, map[string]interface{}{
							"properties": nestedSchemaMap,
						})
						if len(nestedData.(map[string]interface{})) > 0 {
							result[propName] = nestedData
						}
					}
				case "array":
					result[propName] = []interface{}{}
				}
			}
		} else {
			// Property exists, check if it's a nested object that needs defaults applied
			if nestedObj, ok := result[propName].(map[string]interface{}); ok {
				// Check if this property schema defines it as an object with properties
				if nestedProps, hasNestedProps := propMap["properties"].(map[string]interface{}); hasNestedProps {
					// Recursively apply defaults to nested object
					nestedSchemaMap := map[string]interface{}{
						"properties": nestedProps,
					}
					enrichedNested := applyDefaults(nestedObj, nil, nestedSchemaMap)
					// Only update if we got enriched data back
					if enrichedMap, ok := enrichedNested.(map[string]interface{}); ok {
						result[propName] = enrichedMap
					}
				}
			}
		}
	}

	return result
}

// ApplyDefaultsFromSchema applies default values from a compiled schema to JSON data
func ApplyDefaultsFromSchema(data interface{}, schemaStr string) (interface{}, error) {
	var schemaMap map[string]interface{}
	if err := json.Unmarshal([]byte(schemaStr), &schemaMap); err != nil {
		return nil, err
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", bytes.NewReader([]byte(schemaStr))); err != nil {
		return nil, err
	}

	schema, err := compiler.Compile("schema.json")
	if err != nil {
		return nil, err
	}

	return applyDefaults(data, schema, schemaMap), nil
}
