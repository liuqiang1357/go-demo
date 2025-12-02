package jsonschema

import (
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// Package jsonschema provides JSON schema default value application utilities.

// ApplyDefaults applies default values from schema to JSON data.
// Important rules:
//   - data == nil is treated as JSON null and is preserved as-is (no defaults applied)
//   - Defaults are only applied to *missing* object properties (non-required properties only)
//   - Required properties never receive defaults (they must be explicitly provided)
//   - Explicit null values are preserved and do not receive defaults
//   - Defaults are recursively applied to nested objects and arrays
func ApplyDefaults(data interface{}, schema *jsonschema.Schema) interface{} {
	if schema == nil {
		return data
	}

	// Explicit null should be preserved and not receive defaults
	if data == nil {
		return nil
	}

	// Handle $ref: resolve reference first
	schema = resolveRef(schema)

	// Handle combination schemas: allOf, oneOf, anyOf
	if len(schema.AllOf) > 0 {
		return applyDefaultsWithCombination(data, schema.AllOf, schema, "allOf")
	}
	if len(schema.OneOf) > 0 {
		return applyDefaultsWithCombination(data, schema.OneOf, schema, "oneOf")
	}
	if len(schema.AnyOf) > 0 {
		return applyDefaultsWithCombination(data, schema.AnyOf, schema, "anyOf")
	}

	// Check if it's an object schema (has properties, even without explicit type)
	if schema.Properties != nil {
		if obj, ok := data.(map[string]interface{}); ok {
			return applyDefaultsToObject(obj, schema)
		}
		// Type mismatch: return original data
		return data
	}

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
	if len(required) == 0 {
		return false
	}
	for _, r := range required {
		if r == propName {
			return true
		}
	}
	return false
}

// applyDefaultsToObject applies default values to an object.
// Only non-required properties that are missing will receive defaults.
// Required properties are skipped and must be explicitly provided.
func applyDefaultsToObject(data interface{}, schema *jsonschema.Schema) interface{} {
	dataMap, ok := data.(map[string]interface{})
	if !ok || schema.Properties == nil {
		return data
	}

	result := make(map[string]interface{})
	for k, v := range dataMap {
		result[k] = v
	}

	for propName, propSchema := range schema.Properties {
		// Skip required properties - they must be explicitly provided, no defaults applied
		if propSchema == nil || isRequired(propName, schema.Required) {
			continue
		}

		existingValue, exists := result[propName]
		if !exists {
			// Property doesn't exist (non-required): apply default or recursively process
			// ApplyDefaults handles $ref internally, so we can use it directly
			if value := applyDefaultsForProperty(nil, propSchema); shouldAddValue(value) {
				result[propName] = value
			}
		} else if existingValue != nil {
			// Property exists and is not nil: recursively apply defaults to nested structures
			// Preserve nil values as-is (user explicitly provided null)
			result[propName] = applyDefaultsForProperty(existingValue, propSchema)
		}
	}

	return result
}

// applyDefaultsForProperty applies defaults to a property value based on its schema
func applyDefaultsForProperty(value interface{}, propSchema *jsonschema.Schema) interface{} {
	if propSchema == nil {
		return value
	}

	// For nil values (property missing), try to infer type from schema to create empty structure
	if value == nil {
		// ApplyDefaults will handle $ref and combination keywords.
		// Here we just need a hint whether we should start from an empty object/array.
		resolvedSchema := resolveRef(propSchema)

		if resolvedSchema != nil {
			// Direct object/array hints from this schema
			if resolvedSchema.Properties != nil || hasType(resolvedSchema, "object") {
				value = map[string]interface{}{}
			} else if hasType(resolvedSchema, "array") {
				value = []interface{}{}
			} else {
				// If it's a combination schema, try to infer from its children
				children := append(append(resolvedSchema.AllOf, resolvedSchema.AnyOf...), resolvedSchema.OneOf...)
				for _, child := range children {
					child = resolveRef(child)
					if child == nil {
						continue
					}
					if child.Properties != nil || hasType(child, "object") {
						value = map[string]interface{}{}
						break
					}
					if hasType(child, "array") {
						value = []interface{}{}
						break
					}
				}

				// If we still don't know the structure, but schema has a default, use it directly
				if value == nil && resolvedSchema.Default != nil {
					return resolvedSchema.Default
				}
			}
		}

		// If we still couldn't infer object/array and there's no default, keep it nil
		if value == nil {
			return nil
		}
	}

	return ApplyDefaults(value, propSchema)
}

// resolveRef resolves $ref recursively
func resolveRef(schema *jsonschema.Schema) *jsonschema.Schema {
	if schema == nil {
		return nil
	}
	for schema.Ref != nil {
		schema = schema.Ref
	}
	return schema
}

// shouldAddValue checks if a value should be added to the result
// Returns false for nil values and empty objects/arrays
func shouldAddValue(value interface{}) bool {
	if value == nil {
		return false
	}
	// Check if it's an empty object
	if obj, ok := value.(map[string]interface{}); ok {
		return len(obj) > 0
	}
	// Check if it's an empty array
	if arr, ok := value.([]interface{}); ok {
		return len(arr) > 0
	}
	return true
}

// applyDefaultsToArray applies default values to array items
func applyDefaultsToArray(data interface{}, schema *jsonschema.Schema) interface{} {
	arr, ok := data.([]interface{})
	if !ok {
		return data
	}

	result := make([]interface{}, len(arr))
	for i, item := range arr {
		// Get schema for this specific position (handles tuple validation)
		itemsSchema := getItemsSchemaForIndex(schema, i)
		if itemsSchema != nil {
			// Apply defaults to array item
			// Note: We preserve the processed value even if it becomes empty/nil, as this is user-provided data
			result[i] = ApplyDefaults(item, itemsSchema)
		} else {
			// No schema for this item, keep original value
			result[i] = item
		}
	}

	return result
}

// getItemsSchemaForIndex extracts the items schema for a specific array index
// Handles both list validation (single schema) and tuple validation (array of schemas)
// For tuple validation, returns the schema at the given index, or the last schema if index exceeds
// For list validation, returns the single schema for all indices
func getItemsSchemaForIndex(schema *jsonschema.Schema, index int) *jsonschema.Schema {
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
		// List validation: same schema for all items
		return items
	case []*jsonschema.Schema:
		// Tuple validation: different schema for each position
		if len(items) == 0 {
			return nil
		}
		// If index is within bounds, use the schema at that position
		if index < len(items) {
			return items[index]
		}
		// If index exceeds tuple length, use the last schema (additionalItems behavior)
		// This follows JSON Schema spec: additional items beyond tuple length use the last schema
		return items[len(items)-1]
	default:
		return nil
	}
}

// applyDefaultsWithCombination applies defaults from combination schemas (allOf/oneOf/anyOf)
func applyDefaultsWithCombination(data interface{}, subschemas []*jsonschema.Schema, baseSchema *jsonschema.Schema, mode string) interface{} {
	var schemasToApply []*jsonschema.Schema

	switch mode {
	case "allOf":
		// allOf: apply all subschemas
		schemasToApply = subschemas
	case "oneOf":
		// oneOf: find exactly one matching schema
		var matching []*jsonschema.Schema
		for _, s := range subschemas {
			if s.Validate(data) == nil {
				matching = append(matching, s)
			}
		}
		if len(matching) == 1 {
			schemasToApply = matching
		} else {
			// Graceful degradation: apply all if no unique match
			schemasToApply = subschemas
		}
	case "anyOf":
		// anyOf: find matching schemas
		var matching []*jsonschema.Schema
		for _, s := range subschemas {
			if s.Validate(data) == nil {
				matching = append(matching, s)
			}
		}
		if len(matching) > 0 {
			schemasToApply = matching
		} else {
			// Graceful degradation: apply all if none match
			schemasToApply = subschemas
		}
	}

	// Apply defaults from selected schemas sequentially
	result := data
	for _, s := range schemasToApply {
		result = ApplyDefaults(result, s)
	}

	return applyDefaultsToBaseSchema(result, baseSchema)
}

// applyDefaultsToBaseSchema applies defaults from the base schema (properties, etc.)
// This is used after applying defaults from combination schemas (allOf/anyOf/oneOf)
func applyDefaultsToBaseSchema(data interface{}, schema *jsonschema.Schema) interface{} {
	if schema.Properties != nil {
		if obj, ok := data.(map[string]interface{}); ok {
			return applyDefaultsToObject(obj, schema)
		}
		return data
	}

	if hasType(schema, "array") {
		return applyDefaultsToArray(data, schema)
	}

	return data
}
