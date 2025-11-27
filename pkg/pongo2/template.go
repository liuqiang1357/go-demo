package pongo2

// This package provides template generation utilities using pongo2.
// See template_test.go for usage examples and tests.

import (
	"encoding/json"

	"github.com/flosch/pongo2/v6"
)

// Register common filters used in templates.
func init() {
	// to_json encodes a value as a JSON fragment string.
	// It is intended to be used inside JSON templates, e.g.:
	//   "name": {{ name|to_json }}
	// so that quotes and special characters are safely escaped.
	// The returned value is marked as "safe" to prevent further HTML escaping.
	pongo2.RegisterFilter("to_json", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		b, err := json.Marshal(in.Interface())
		if err != nil {
			// In current pongo2 versions, the Error struct is generally constructed
			// via helper functions; here we simply return the original error message
			// as a template error without relying on concrete fields.
			return nil, &pongo2.Error{Sender: "filter:to_json: " + err.Error()}
		}
		// Mark the JSON fragment as safe so it won't be HTML-escaped again.
		return pongo2.AsSafeValue(string(b)), nil
	})
}
