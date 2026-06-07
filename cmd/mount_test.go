package cmd

import (
	"strings"
	"testing"

	"cloud-ide-mount/internal/codespace"
	"cloud-ide-mount/internal/state"
	"cloud-ide-mount/internal/ui"
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

// ─── FilterAvailable ─────────────────────────────────────────────────────

func TestFilterAvailable_AllAvailable(t *testing.T) {
	selected := []codespace.Codespace{
		{Name: "cs-one", State: "Available"},
		{Name: "cs-two", State: "Available"},
	}
	got := filterAvailable(selected)
	if len(got) != 2 {
		t.Errorf("expected 2 available, got %d", len(got))
	}
}

func TestFilterAvailable_SomeSkipped(t *testing.T) {
	selected := []codespace.Codespace{
		{Name: "cs-one", State: "Available"},
		{Name: "cs-two", State: "Shutdown"},
		{Name: "cs-three", State: "Available"},
	}
	got := filterAvailable(selected)
	if len(got) != 2 {
		t.Errorf("expected 2 available, got %d", len(got))
	}
	if got[0].Name != "cs-one" {
		t.Errorf("expected cs-one first, got %s", got[0].Name)
	}
	if got[1].Name != "cs-three" {
		t.Errorf("expected cs-three second, got %s", got[1].Name)
	}
}

func TestFilterAvailable_NoneAvailable(t *testing.T) {
	selected := []codespace.Codespace{
		{Name: "cs-one", State: "Shutdown"},
		{Name: "cs-two", State: "Starting"},
	}
	got := filterAvailable(selected)
	if len(got) != 0 {
		t.Errorf("expected 0 available, got %d", len(got))
	}
}

func TestFilterAvailable_EmptyInput(t *testing.T) {
	got := filterAvailable(nil)
	if len(got) != 0 {
		t.Errorf("expected 0 available for nil input, got %d", len(got))
	}
}

// ─── GroupByDrive ───────────────────────────────────────────────────────

func TestGroupByDrive_SingleDrive(t *testing.T) {
	assignments := []ui.DriveAssignment{
		{Drive: "D:", Codespace: "cs-one"},
		{Drive: "D:", Codespace: "cs-two"},
	}
	groups := groupByDrive(assignments)
	if len(groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(groups))
	}
	if len(groups["D:"]) != 2 {
		t.Errorf("expected 2 codespaces on D:, got %d", len(groups["D:"]))
	}
}

func TestGroupByDrive_MultipleDrives(t *testing.T) {
	assignments := []ui.DriveAssignment{
		{Drive: "D:", Codespace: "cs-one"},
		{Drive: "E:", Codespace: "cs-two"},
		{Drive: "F:", Codespace: "cs-three"},
	}
	groups := groupByDrive(assignments)
	if len(groups) != 3 {
		t.Errorf("expected 3 groups, got %d", len(groups))
	}
	for _, d := range []string{"D:", "E:", "F:"} {
		if len(groups[d]) != 1 {
			t.Errorf("expected 1 codespace on %s, got %d", d, len(groups[d]))
		}
	}
}

func TestGroupByDrive_DriveOrder(t *testing.T) {
	assignments := []ui.DriveAssignment{
		{Drive: "D:", Codespace: "cs-one"},
		{Drive: "E:", Codespace: "cs-two"},
		{Drive: "D:", Codespace: "cs-three"},
	}
	groups := groupByDrive(assignments)
	// D: should have cs-one then cs-three (preserving order)
	if len(groups["D:"]) != 2 {
		t.Errorf("expected 2 on D:, got %d", len(groups["D:"]))
	}
	if groups["D:"][0] != "cs-one" || groups["D:"][1] != "cs-three" {
		t.Errorf("D: order should be [cs-one, cs-three], got %v", groups["D:"])
	}
}

func TestGroupByDrive_EmptyInput(t *testing.T) {
	groups := groupByDrive(nil)
	if len(groups) != 0 {
		t.Errorf("expected empty map for nil input, got %d", len(groups))
	}
}

// ─── InitState ──────────────────────────────────────────────────────────

func TestInitState_NilFields(t *testing.T) {
	s := &state.State{}
	initState(s)
	if s.Remotes == nil {
		t.Error("expected non-nil Remotes after initState")
	}
	if s.Mounts == nil {
		t.Error("expected non-nil Mounts after initState")
	}
}

func TestInitState_AlreadyInitialized(t *testing.T) {
	s := &state.State{
		Remotes: []state.Remote{{Name: "test"}},
		Mounts:  []state.Mount{{Drive: "D:"}},
	}
	initState(s)
	if len(s.Remotes) != 1 {
		t.Errorf("expected 1 remote preserved, got %d", len(s.Remotes))
	}
	if len(s.Mounts) != 1 {
		t.Errorf("expected 1 mount preserved, got %d", len(s.Mounts))
	}
}

// ─── StartCodespaces ────────────────────────────────────────────────────

func TestStartCodespaces_AllAvailable(t *testing.T) {
	// When all codespaces are already Available, startCodespaces is a no-op
	selected := []codespace.Codespace{
		{Name: "cs-one", State: "Available"},
		{Name: "cs-two", State: "Available"},
	}
	err := startCodespaces(selected)
	if err != nil {
		t.Errorf("expected no error for all-available, got %v", err)
	}
}

func TestStartCodespaces_EmptyInput(t *testing.T) {
	err := startCodespaces(nil)
	if err != nil {
		t.Errorf("expected no error for empty input, got %v", err)
	}
}

// mockLookPathError implements the error interface to simulate exec.LookPath error
type mockLookPathError struct {
	name string
}

func (e *mockLookPathError) Error() string {
	return e.name + ": executable file not found"
}
