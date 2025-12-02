package jsonschema

import (
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// This package provides JSON schema validation and default value application utilities.
// See schema_test.go for usage examples and tests.

// ApplyDefaults applies default values from schema to JSON data
func ApplyDefaults(data interface{}, schema *jsonschema.Schema) interface{} {
	if schema == nil {
		return data
	}

	// Check if schema has object type
	if hasType(schema, "object") && schema.Properties != nil {
		return applyDefaultsToObject(data, schema)
	}

	// Check if schema has array type
	if hasType(schema, "array") {
		return applyDefaultsToArray(data, schema)
	}

	return data
}

// hasType checks if schema has the specified type
func hasType(schema *jsonschema.Schema, typ string) bool {
	if schema == nil {
		return false
	}
	for _, t := range schema.Types {
		if t == typ {
			return true
		}
	}
	return false
}

// isRequired checks if a property is required
func isRequired(propName string, required []string) bool {
	for _, r := range required {
		if r == propName {
			return true
		}
	}
	return false
}

// applyDefaultsToObject applies default values to an object
func applyDefaultsToObject(data interface{}, schema *jsonschema.Schema) interface{} {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return data
	}

	if schema.Properties == nil {
		return data
	}

	result := make(map[string]interface{})

	// Copy existing data
	for k, v := range dataMap {
		result[k] = v
	}

	// Apply defaults only for non-required missing properties
	for propName, propSchema := range schema.Properties {
		if propSchema == nil {
			continue
		}

		// Skip required properties - they should be provided by the caller
		if isRequired(propName, schema.Required) {
			continue
		}

		if _, exists := result[propName]; !exists {
			// Property doesn't exist, apply default
			if propSchema.Default != nil {
				result[propName] = propSchema.Default
			} else if hasType(propSchema, "object") {
				// For nested objects, recursively apply defaults
				if propSchema.Properties != nil {
					nestedData := ApplyDefaults(map[string]interface{}{}, propSchema)
					if nestedMap, ok := nestedData.(map[string]interface{}); ok && len(nestedMap) > 0 {
						result[propName] = nestedData
					}
				}
			} else if hasType(propSchema, "array") {
				// For arrays, apply defaults to items if schema defines them
				if getItemsSchema(propSchema) != nil {
					result[propName] = ApplyDefaults([]interface{}{}, propSchema)
				} else {
					result[propName] = []interface{}{}
				}
			}
		} else {
			// Property exists, recursively apply defaults if needed
			if hasType(propSchema, "object") {
				// Check if this property schema defines it as an object with properties
				if propSchema.Properties != nil {
					if nestedObj, ok := result[propName].(map[string]interface{}); ok {
						enrichedNested := ApplyDefaults(nestedObj, propSchema)
						if enrichedMap, ok := enrichedNested.(map[string]interface{}); ok {
							result[propName] = enrichedMap
						}
					}
				}
			} else if hasType(propSchema, "array") {
				// Apply defaults to array items
				if getItemsSchema(propSchema) != nil {
					if arr, ok := result[propName].([]interface{}); ok {
						result[propName] = ApplyDefaults(arr, propSchema)
					}
				}
			}
		}
	}

	return result
}

// applyDefaultsToArray applies default values to array items
func applyDefaultsToArray(data interface{}, schema *jsonschema.Schema) interface{} {
	arr, ok := data.([]interface{})
	if !ok {
		return data
	}

	itemsSchema := getItemsSchema(schema)
	if itemsSchema == nil {
		return data
	}

	result := make([]interface{}, len(arr))
	for i, item := range arr {
		// Apply defaults to each array item
		result[i] = ApplyDefaults(item, itemsSchema)
	}

	return result
}

// getItemsSchema extracts the items schema from array schema
// Items can be *Schema, []*Schema, or nil
func getItemsSchema(schema *jsonschema.Schema) *jsonschema.Schema {
	if schema == nil {
		return nil
	}

	// Handle Items2020 (draft 2020-12)
	if schema.Items2020 != nil {
		return schema.Items2020
	}

	// Handle Items (can be *Schema or []*Schema)
	if schema.Items == nil {
		return nil
	}

	switch items := schema.Items.(type) {
	case *jsonschema.Schema:
		return items
	case []*jsonschema.Schema:
		// For tuple validation, we use the first item schema
		// In practice, you might want to handle each position differently
		if len(items) > 0 {
			return items[0]
		}
		return nil
	default:
		return nil
	}
}
