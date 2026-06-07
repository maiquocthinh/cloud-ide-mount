package cmd

import (
	"strings"
	"testing"

	"cloud-ide-mount/internal/codespace"
)

func TestFindFolderPath(t *testing.T) {
	paths := []codespace.UpstreamPath{
		{Cs: codespace.Codespace{Name: "cs-one", Repository: "org/repo-a"}, FolderPath: "org/repo-a"},
		{Cs: codespace.Codespace{Name: "cs-two", Repository: "org/repo-b"}, FolderPath: "org/repo-b"},
	}

	tests := []struct {
		name     string
		csName   string
		expected string
	}{
		{"found match", "cs-one", "org/repo-a"},
		{"found second", "cs-two", "org/repo-b"},
		{"not found returns safe name", "cs-unknown", "cs-unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findFolderPath(paths, tt.csName)
			if got != tt.expected {
				t.Errorf("findFolderPath(%q) = %q, want %q", tt.csName, got, tt.expected)
			}
		})
	}
}

func TestDetectSSHPort_DefaultOnError(t *testing.T) {
	// When exec fails (e.g. no gh CLI), detectSSHPort should return 22 (default)
	port := detectSSHPort("non-existent-cs")
	if port != 22 {
		t.Errorf("expected default port 22 on exec failure, got %d", port)
	}
}

func TestExecLook_NotFound(t *testing.T) {
	_, err := execLook("this-command-does-not-exist-12345")
	if err == nil {
		t.Error("expected error for non-existent command, got nil")
	}
}

func TestCheckDeps_MissingDeps(t *testing.T) {
	// Save original exec.LookPath used by checkDeps via execLook
	// We can only verify the function returns error for missing deps
	// This test uses the actual exec.LookPath
	origExecLook := execLook

	// Override execLook to simulate missing gh
	execLook = func(name string) (string, error) {
		if name == "gh" {
			return "", &mockLookPathError{name: name}
		}
		return origExecLook(name)
	}
	defer func() { execLook = origExecLook }()

	err := checkDeps()
	if err == nil {
		t.Error("expected error when gh is missing, got nil")
	}
	if !strings.Contains(err.Error(), "gh not found") {
		t.Errorf("error should mention gh, got: %v", err)
	}
}

func TestCheckDeps_MissingRclone(t *testing.T) {
	origExecLook := execLook

	execLook = func(name string) (string, error) {
		if name == "rclone" {
			return "", &mockLookPathError{name: name}
		}
		return origExecLook(name)
	}
	defer func() { execLook = origExecLook }()

	err := checkDeps()
	if err == nil {
		t.Error("expected error when rclone is missing, got nil")
	}
	if !strings.Contains(err.Error(), "rclone not found") {
		t.Errorf("error should mention rclone, got: %v", err)
	}
}

// mockLookPathError implements the error interface to simulate exec.LookPath error
type mockLookPathError struct {
	name string
}

func (e *mockLookPathError) Error() string {
	return e.name + ": executable file not found"
}
