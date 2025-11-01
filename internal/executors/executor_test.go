package executors

import (
	"os"
	"path/filepath"
	"testing"
	
	"sqlfuse/internal/stmts/stmts"
)

func TestNewCmdExecutor(t *testing.T) {
	exec := NewCmdExecutor("test-exec", "/usr/bin/test-exec")
	
	if exec.Name() != "test-exec" {
		t.Errorf("Expected name 'test-exec', got '%s'", exec.Name())
	}
	
	if exec.Path() != "/usr/bin/test-exec" {
		t.Errorf("Expected path '/usr/bin/test-exec', got '%s'", exec.Path())
	}
}

func TestCmdExecutor_BuildCmd(t *testing.T) {
	exec := NewCmdExecutor("test", "/bin/echo")
	
	// Test without seed
	cmd := exec.BuildCmd([]string{"arg1", "arg2"}, nil)
	
	if cmd.Path != "/bin/echo" {
		t.Errorf("Expected command path '/bin/echo', got '%s'", cmd.Path)
	}
	
	if len(cmd.Args) != 3 || cmd.Args[1] != "arg1" || cmd.Args[2] != "arg2" {
		t.Errorf("Expected args [echo, arg1, arg2], got %v", cmd.Args)
	}
}

func TestCmdExecutor_BuildCmd_WithSeed(t *testing.T) {
	exec := NewCmdExecutor("test", "/bin/echo")
	
	// Test with seed
	seed := int64(42)
	cmd := exec.BuildCmd([]string{"arg1"}, &seed)
	
	// Seed should be prepended as --seed=42
	if len(cmd.Args) < 3 {
		t.Fatalf("Expected at least 3 args, got %d", len(cmd.Args))
	}
	
	if cmd.Args[1] != "--seed=42" {
		t.Errorf("Expected first arg '--seed=42', got '%s'", cmd.Args[1])
	}
	
	if cmd.Args[2] != "arg1" {
		t.Errorf("Expected second arg 'arg1', got '%s'", cmd.Args[2])
	}
}

func TestResolveExecPath_EmptyName(t *testing.T) {
	tmpDir := t.TempDir()
	
	_, err := ResolveExecPath(tmpDir, "")
	if err == nil {
		t.Error("Expected error for empty executable name")
	}
	
	if err.Error() != "empty executable name" {
		t.Errorf("Expected 'empty executable name' error, got '%v'", err)
	}
}

func TestResolveExecPath_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	
	_, err := ResolveExecPath(tmpDir, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent executable")
	}
	
	if err.Error() != "executable not found in output directory" {
		t.Errorf("Expected 'executable not found' error, got '%v'", err)
	}
}

func TestResolveExecPath_IsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)
	
	_, err := ResolveExecPath(tmpDir, "subdir")
	if err == nil {
		t.Error("Expected error for directory path")
	}
	
	if err.Error() != "executable path is a directory" {
		t.Errorf("Expected 'path is a directory' error, got '%v'", err)
	}
}

func TestResolveExecPath_NotExecutable(t *testing.T) {
	tmpDir := t.TempDir()
	notExec := filepath.Join(tmpDir, "notexec")
	
	// Create a non-executable file
	f, err := os.Create(notExec)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	f.Close()
	
	// Set mode without execute permissions
	os.Chmod(notExec, 0644)
	
	_, err = ResolveExecPath(tmpDir, "notexec")
	if err == nil {
		t.Error("Expected error for non-executable file")
	}
	
	if err.Error() != "file is not executable" {
		t.Errorf("Expected 'not executable' error, got '%v'", err)
	}
}

func TestResolveExecPath_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	execPath := filepath.Join(tmpDir, "myexec")
	
	// Create an executable file
	f, err := os.Create(execPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	f.Close()
	
	// Set execute permissions
	os.Chmod(execPath, 0755)
	
	resolved, err := ResolveExecPath(tmpDir, "myexec")
	if err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	
	absExpected, _ := filepath.Abs(execPath)
	if resolved != absExpected {
		t.Errorf("Expected path '%s', got '%s'", absExpected, resolved)
	}
}

func TestResolveExecPath_OutsideDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Try to access a file outside the output directory
	_, err := ResolveExecPath(tmpDir, "../outside")
	if err == nil {
		t.Error("Expected error for path outside output directory")
	}
	
	// The error could be either "must reside inside" or "executable not found"
	// depending on whether the file exists or not
	errMsg := err.Error()
	if errMsg != "executable must reside inside output directory" && 
	   errMsg != "executable not found in output directory" {
		t.Errorf("Expected security or not found error, got '%v'", err)
	}
}

func TestResolveExecPath_AbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	execPath := filepath.Join(tmpDir, "absexec")
	
	// Create an executable file
	f, err := os.Create(execPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	f.Close()
	
	os.Chmod(execPath, 0755)
	
	// Resolve using absolute path
	resolved, err := ResolveExecPath(tmpDir, execPath)
	if err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	
	absExpected, _ := filepath.Abs(execPath)
	if resolved != absExpected {
		t.Errorf("Expected path '%s', got '%s'", absExpected, resolved)
	}
}

func TestDiscoverExecutors_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	
	executors, err := DiscoverExecutors(tmpDir)
	if err != nil {
		t.Fatalf("Expected no error, got '%v'", err)
	}
	
	if len(executors) != 0 {
		t.Errorf("Expected 0 executors in empty directory, got %d", len(executors))
	}
}

func TestDiscoverExecutors_WithExecutables(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create two executable files
	exec1 := filepath.Join(tmpDir, "exec1")
	exec2 := filepath.Join(tmpDir, "exec2")
	
	for _, path := range []string{exec1, exec2} {
		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		f.Close()
		os.Chmod(path, 0755)
	}
	
	// Create a non-executable file (should be ignored)
	notExec := filepath.Join(tmpDir, "notexec")
	f, err := os.Create(notExec)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	f.Close()
	os.Chmod(notExec, 0644)
	
	// Create a subdirectory (should be ignored)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)
	
	executors, err := DiscoverExecutors(tmpDir)
	if err != nil {
		t.Fatalf("Expected no error, got '%v'", err)
	}
	
	if len(executors) != 2 {
		t.Errorf("Expected 2 executors, got %d", len(executors))
	}
	
	// Verify executor names
	names := make(map[string]bool)
	for _, exec := range executors {
		names[exec.Name()] = true
	}
	
	if !names["exec1"] || !names["exec2"] {
		t.Errorf("Expected executors 'exec1' and 'exec2', got %v", names)
	}
}

func TestDiscoverExecutors_InvalidDirectory(t *testing.T) {
	_, err := DiscoverExecutors("/nonexistent/path/to/directory")
	if err == nil {
		t.Error("Expected error for invalid directory")
	}
}

func TestStringsHasDotDot(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"simple/path", false},
		{"..", true},
		{"normal/path", false},
	}
	
	for _, tt := range tests {
		result := stringsHasDotDot(tt.input)
		if result != tt.expected {
			t.Errorf("stringsHasDotDot(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestWeightSetter_Interface(t *testing.T) {
	// This test verifies that the WeightSetter interface can be implemented
	// The interface requires SetWeights with StmtType map
	
	var _ WeightSetter = (*mockWeightSetter)(nil)
}

// mockWeightSetter is a minimal implementation of WeightSetter for testing
type mockWeightSetter struct{}

func (m *mockWeightSetter) SetWeights(weights map[stmts.StmtType]uint64) {
	// Minimal implementation for interface compliance
}
