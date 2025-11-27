package pongo2

import (
	"strings"
	"testing"

	"github.com/flosch/pongo2/v6"
)

func TestJSONGeneration(t *testing.T) {
	templateString := `{
  "name": "{{ name }}",
  "age": {{ age }},
  "email": "{{ email }}",
  "active": {{ active }},
  "tags": [
{% for tag in tags %}
    "{{ tag }}"{% if not forloop.last %},{% endif %}
{% endfor %}
  ],
  "metadata": {
    "created_at": "{{ metadata.created_at }}",
    "version": {{ metadata.version }}
  }
}`

	template, err := pongo2.FromString(templateString)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	context := pongo2.Context{
		"name":  "John Doe",
		"age":   30,
		"email": "john@example.com",
		"active": true,
		"tags":   []string{"golang", "testing", "pongo2"},
		"metadata": map[string]interface{}{
			"created_at": "2024-01-15T10:30:00Z",
			"version":    1,
		},
	}

	output, err := template.Execute(context)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	// Verify output contains expected values
	if !strings.Contains(output, "John Doe") {
		t.Error("Output should contain 'John Doe'")
	}
	if !strings.Contains(output, "30") {
		t.Error("Output should contain age '30'")
	}
	if !strings.Contains(output, "golang") {
		t.Error("Output should contain tag 'golang'")
	}
	if !strings.Contains(output, "2024-01-15T10:30:00Z") {
		t.Error("Output should contain created_at timestamp")
	}

	t.Logf("Generated JSON:\n%s", output)
}

func TestXMLGeneration(t *testing.T) {
	templateString := `<?xml version="1.0" encoding="UTF-8"?>
<user>
  <name>{{ name }}</name>
  <age>{{ age }}</age>
  <email>{{ email }}</email>
  <active>{{ active|lower }}</active>
  <tags>
{% for tag in tags %}
    <tag>{{ tag }}</tag>
{% endfor %}
  </tags>
  <metadata>
    <created_at>{{ metadata.created_at }}</created_at>
    <version>{{ metadata.version }}</version>
  </metadata>
</user>`

	template, err := pongo2.FromString(templateString)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	context := pongo2.Context{
		"name":  "Jane Smith",
		"age":   28,
		"email": "jane@example.com",
		"active": true,
		"tags":   []string{"xml", "template", "generation"},
		"metadata": map[string]interface{}{
			"created_at": "2024-01-15T10:30:00Z",
			"version":    2,
		},
	}

	output, err := template.Execute(context)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	// Verify output contains expected values
	if !strings.Contains(output, "Jane Smith") {
		t.Error("Output should contain 'Jane Smith'")
	}
	if !strings.Contains(output, "<user>") {
		t.Error("Output should contain XML root element")
	}
	if !strings.Contains(output, "xml") {
		t.Error("Output should contain tag 'xml'")
	}
	if !strings.Contains(output, "true") {
		t.Error("Output should contain active status")
	}

	t.Logf("Generated XML:\n%s", output)
}

func TestMarkdownGeneration(t *testing.T) {
	templateString := `# {{ title }}

**Author:** {{ author }}  
**Date:** {{ date }}  
**Status:** {{ status }}

## Description

{{ description }}

## Features

{% for feature in features %}
- {{ feature.name }}: {{ feature.description }}
{% endfor %}

## Code Example

` + "```" + `{{ code_language }}
{{ code_example }}
` + "```" + `

## Tags

{% for tag in tags %}
- ` + "`" + `{{ tag }}` + "`" + `
{% endfor %}

---

*Generated at {{ generated_at }}*`

	template, err := pongo2.FromString(templateString)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	context := pongo2.Context{
		"title":       "Pongo2 Template Engine",
		"author":      "Test User",
		"date":        "2024-01-15",
		"status":      "Active",
		"description": "A powerful template engine for Go, inspired by Django templates.",
		"features": []map[string]string{
			{"name": "Template Inheritance", "description": "Support for template inheritance and includes"},
			{"name": "Filters", "description": "Rich set of built-in filters"},
			{"name": "Control Structures", "description": "Loops, conditionals, and more"},
		},
		"code_language": "go",
		"code_example": `template := pongo2.Must(pongo2.FromString("Hello {{ name }}!"))
output := template.Execute(pongo2.Context{"name": "World"})`,
		"tags":        []string{"golang", "template", "pongo2"},
		"generated_at": "2024-01-15T10:30:00Z",
	}

	output, err := template.Execute(context)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	// Verify output contains expected values
	if !strings.Contains(output, "# Pongo2 Template Engine") {
		t.Error("Output should contain title as heading")
	}
	if !strings.Contains(output, "Test User") {
		t.Error("Output should contain author")
	}
	if !strings.Contains(output, "Template Inheritance") {
		t.Error("Output should contain feature names")
	}
	if !strings.Contains(output, "```go") {
		t.Error("Output should contain code block with language")
	}
	if !strings.Contains(output, "golang") {
		t.Error("Output should contain tags")
	}

	t.Logf("Generated Markdown:\n%s", output)
}

