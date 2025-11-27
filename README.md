# Go Demo Project

A Go project for testing various libraries and their usage patterns using Go's testing framework.

## Libraries Tested

### 1. flosch/pongo2

Template engine for generating strings in various formats (JSON, XML, Markdown).

**Features:**
- Template inheritance and includes
- Rich set of filters
- Control structures (loops, conditionals)
- Django-inspired syntax

### 2. santhosh-tekuri/jsonschema

JSON Schema validation and manipulation library.

**Features:**
- JSON Schema validation
- Dynamic schema compilation
- Default value application
- Support for draft-07 schema

## Usage

### Prerequisites

- Go 1.21 or higher

### Installation

```bash
go mod download
```

### Running Tests

#### Run All Tests

```bash
# Run all tests
go test ./...

# With verbose output
go test -v ./...
```

#### Run Specific Test Packages

```bash
# Test pongo2 only
go test -v ./pkg/pongo2/...

# Test jsonschema only
go test -v ./pkg/jsonschema/...
```

#### Run Specific Tests

```bash
# Run a specific test function
go test -v ./pkg/pongo2 -run TestJSONGeneration

# Run tests matching a pattern
go test -v ./pkg/jsonschema -run TestDefault
```

## Project Structure

```
go-demo/
├── go.mod                       # Go module definition
├── .gitignore                   # Git ignore rules
├── .gitattributes               # Git attributes for line endings
├── pkg/
│   ├── pongo2/
│   │   ├── template.go          # Package documentation
│   │   └── template_test.go     # Pongo2 template tests
│   └── jsonschema/
│       ├── validator.go         # Package documentation
│       └── validator_test.go    # JSON Schema validation tests
└── README.md                    # This file
```

## Test Coverage

### Pongo2 Tests

- **TestJSONGeneration**: Tests JSON template generation with variables, loops, and nested objects
- **TestXMLGeneration**: Tests XML document generation with proper formatting
- **TestMarkdownGeneration**: Tests Markdown generation with headings, lists, and code blocks

### JSON Schema Tests

#### Validation Tests

- **Valid JSON**: Tests successful validation of correct data
- **Missing Required Field**: Tests validation failure for missing required fields
- **Wrong Type**: Tests validation failure for type mismatches
- **Constraint Violation**: Tests validation failure for constraint violations (e.g., max value)
- **Empty Tags Array**: Tests validation failure for array minimum items constraint

#### Default Value Tests

- **Partial JSON**: Tests applying defaults to JSON with missing fields
- **Empty JSON Object**: Tests applying all defaults to an empty object
- **Nested Object**: Tests applying defaults to nested objects recursively

## Examples

### Pongo2 - JSON Generation

Generates structured JSON output from template with dynamic data, including arrays and nested objects.

### Pongo2 - XML Generation

Generates XML documents from templates with proper formatting and element nesting.

### Pongo2 - Markdown Generation

Creates Markdown documents with dynamic content, lists, code blocks, and metadata.

### JSON Schema - Validation

Validates JSON data against a schema with comprehensive test cases covering:
- Valid data scenarios
- Missing required fields
- Type mismatches
- Constraint violations

### JSON Schema - Default Values

Applies default values from schema to incomplete JSON data:
- Partial JSON objects
- Empty objects
- Nested objects with recursive default application

