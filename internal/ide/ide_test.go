package ide

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureSSHConfig_CreatesDirAndFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("USERPROFILE", tmpDir)

	info := SSHInfo{
		Host:       "127.0.0.1",
		Port:       2223,
		User:       "codespace",
		KeyFile:    filepath.Join(tmpDir, ".ssh", "codespaces.auto"),
		Alias:      "cs-test-codespace",
		RemotePath: "/workspaces/repo",
	}

	if err := ensureSSHConfig(info); err != nil {
		t.Fatalf("ensureSSHConfig failed: %v", err)
	}

	configPath := filepath.Join(tmpDir, ".ssh", "config")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("cannot read created config: %v", err)
	}
	if !strings.Contains(string(data), "Host cs-test-codespace") {
		t.Errorf("config missing Host entry:\n%s", string(data))
	}
}

func TestEnsureSSHConfig_NoDuplicateEntry(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("USERPROFILE", tmpDir)

	sshDir := filepath.Join(tmpDir, ".ssh")
	os.MkdirAll(sshDir, 0700)
	configPath := filepath.Join(sshDir, "config")
	os.WriteFile(configPath, []byte("Host cs-existing\n  HostName 127.0.0.1\n"), 0644)

	info := SSHInfo{
		Host:       "127.0.0.1",
		Port:       2223,
		User:       "codespace",
		KeyFile:    filepath.Join(tmpDir, ".ssh", "codespaces.auto"),
		Alias:      "cs-existing",
		RemotePath: "/workspaces/repo",
	}

	if err := ensureSSHConfig(info); err != nil {
		t.Fatalf("ensureSSHConfig failed: %v", err)
	}

	data, _ := os.ReadFile(configPath)
	lines := strings.Split(string(data), "\n")
	hostCount := 0
	for _, line := range lines {
		if strings.Contains(line, "Host cs-existing") {
			hostCount++
		}
	}
	if hostCount != 1 {
		t.Errorf("expected 1 Host entry, got %d", hostCount)
	}
}

func TestEnsureSSHConfig_ReturnsErrorOnBadDir(t *testing.T) {
	// USERPROFILE points to a file, not a directory — MkdirAll will fail
	tmpDir := t.TempDir()
	blockPath := filepath.Join(tmpDir, "block")
	os.WriteFile(blockPath, []byte("not-a-dir"), 0644)
	t.Setenv("USERPROFILE", blockPath)

	info := SSHInfo{
		Host:  "127.0.0.1",
		Port:  2223,
		User:  "codespace",
		Alias: "cs-test",
	}

	err := ensureSSHConfig(info)
	if err == nil {
		t.Fatal("expected error when USERPROFILE is a file, got nil")
	}
	if !strings.Contains(err.Error(), "SSH config dir") {
		t.Errorf("error should mention SSH config dir, got: %v", err)
	}
}
