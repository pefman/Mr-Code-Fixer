package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// TestRunner detects and runs tests for different project types
type TestRunner struct {
	RepoPath string
}

func NewTestRunner(repoPath string) *TestRunner {
	return &TestRunner{RepoPath: repoPath}
}

// DetectTestCommand finds the appropriate test command for the project
func (t *TestRunner) DetectTestCommand() (string, bool) {
	// Check for package.json (Node.js)
	if _, err := os.Stat(filepath.Join(t.RepoPath, "package.json")); err == nil {
		// Check if npm test script exists
		return "npm test", true
	}
	
	// Check for go.mod (Go)
	if _, err := os.Stat(filepath.Join(t.RepoPath, "go.mod")); err == nil {
		return "go test ./...", true
	}
	
	// Check for requirements.txt or setup.py (Python)
	if _, err := os.Stat(filepath.Join(t.RepoPath, "requirements.txt")); err == nil {
		return "python -m pytest", true
	}
	if _, err := os.Stat(filepath.Join(t.RepoPath, "setup.py")); err == nil {
		return "python -m pytest", true
	}
	
	// Check for Cargo.toml (Rust)
	if _, err := os.Stat(filepath.Join(t.RepoPath, "Cargo.toml")); err == nil {
		return "cargo test", true
	}
	
	// Check for pom.xml (Maven/Java)
	if _, err := os.Stat(filepath.Join(t.RepoPath, "pom.xml")); err == nil {
		return "mvn test", true
	}
	
	// Check for build.gradle (Gradle/Java)
	if _, err := os.Stat(filepath.Join(t.RepoPath, "build.gradle")); err == nil {
		return "gradle test", true
	}
	
	// Check for composer.json (PHP)
	if _, err := os.Stat(filepath.Join(t.RepoPath, "composer.json")); err == nil {
		return "php vendor/bin/phpunit", true
	}
	
	return "", false
}

// RunTests executes the detected test command
func (t *TestRunner) RunTests() (bool, string, error) {
	testCmd, found := t.DetectTestCommand()
	if !found {
		return true, "No tests detected - skipping", nil
	}
	
	fmt.Printf("\n🧪 Running tests: %s\n", testCmd)
	
	// Split command into parts
	parts := strings.Fields(testCmd)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = t.RepoPath
	
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	
	if err != nil {
		return false, outputStr, fmt.Errorf("tests failed: %w", err)
	}
	
	return true, outputStr, nil
}

// TestResult contains the outcome of running tests
type TestResult struct {
	Passed  bool
	Output  string
	Command string
}

func (t *TestRunner) Execute() *TestResult {
	cmd, found := t.DetectTestCommand()
	if !found {
		return &TestResult{
			Passed:  true,
			Output:  "No tests detected",
			Command: "",
		}
	}
	
	passed, output, _ := t.RunTests()
	return &TestResult{
		Passed:  passed,
		Output:  output,
		Command: cmd,
	}
}
