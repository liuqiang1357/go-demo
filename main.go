package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <command>")
		fmt.Println("Commands:")
		fmt.Println("  test           - Run all tests")
		fmt.Println("  test-pongo2    - Run pongo2 tests only")
		fmt.Println("  test-jsonschema - Run jsonschema tests only")
		fmt.Println("  test-verbose   - Run all tests with verbose output")
		os.Exit(1)
	}

	command := os.Args[1]
	var cmd *exec.Cmd

	switch command {
	case "test":
		cmd = exec.Command("go", "test", "./...")
	case "test-pongo2":
		cmd = exec.Command("go", "test", "-v", "./pkg/pongo2/...")
	case "test-jsonschema":
		cmd = exec.Command("go", "test", "-v", "./pkg/jsonschema/...")
	case "test-verbose":
		cmd = exec.Command("go", "test", "-v", "./...")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}

