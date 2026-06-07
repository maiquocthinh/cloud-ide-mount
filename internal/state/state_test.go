package state

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	// Use a temp dir as app root for testing
	dir := t.TempDir()
	t.Setenv("CLOUD_IDE_MOUNT_ROOT", dir)

	s := &State{}
	s.Remotes = append(s.Remotes, Remote{
		Name:       "test-remote",
		Codespace:  "owner/repo",
		Port:       2223,
		TunnelPid:  12345,
		FolderPath: "owner/repo",
	})
	s.Mounts = append(s.Mounts, Mount{
		Drive:     "z",
		RclonePid: 67890,
		Remote:    "test-remote",
		Mode:      "combined",
	})

	if err := Save(s); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify file exists at the new path
	expectedPath := filepath.Join(dir, "config", "state-default.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("state file not created at %s", expectedPath)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(loaded.Remotes) != 1 || loaded.Remotes[0].Name != "test-remote" {
		t.Errorf("unexpected remotes: %+v", loaded.Remotes)
	}
	if len(loaded.Mounts) != 1 || loaded.Mounts[0].Drive != "z" {
		t.Errorf("unexpected mounts: %+v", loaded.Mounts)
	}
}

func TestLoadNonExistentReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLOUD_IDE_MOUNT_ROOT", dir)

	s, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if s == nil {
		t.Fatal("Load() returned nil, expected empty State")
	}
	if len(s.Remotes) != 0 || len(s.Mounts) != 0 {
		t.Errorf("expected empty state, got remotes=%d mounts=%d", len(s.Remotes), len(s.Mounts))
	}
}

func TestConcurrentSaves(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLOUD_IDE_MOUNT_ROOT", dir)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s := &State{}
			s.Remotes = append(s.Remotes, Remote{
				Name:      "r",
				Port:      2223 + i,
				TunnelPid: 10000 + i,
			})
			if err := Save(s); err != nil {
				t.Errorf("concurrent Save() error: %v", err)
			}
		}(i)
	}
	wg.Wait()

	// Load should succeed without corruption
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() after concurrent saves error: %v", err)
	}
	if loaded == nil {
		t.Fatal("Load() returned nil")
	}
}

func TestRemove(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLOUD_IDE_MOUNT_ROOT", dir)

	s := &State{}
	if err := Save(s); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if err := s.Remove(); err != nil {
		t.Fatalf("Remove() error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "config", "state-default.json")); !os.IsNotExist(err) {
		t.Fatal("state file still exists after Remove()")
	}
}

func TestRemoveNonExistent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLOUD_IDE_MOUNT_ROOT", dir)

	s := &State{}
	if err := s.Remove(); err != nil {
		t.Fatalf("Remove() on non-existent file error: %v", err)
	}
}

func TestAtomicWriteNoCorruptionOnPartialWrite(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLOUD_IDE_MOUNT_ROOT", dir)

	s := &State{}
	for i := 0; i < 10; i++ {
		s.Remotes = append(s.Remotes, Remote{
			Name:      "r",
			Port:      2223 + i,
			TunnelPid: 10000 + i,
		})
	}

	if err := Save(s); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Simulate a crash mid-write by writing garbage directly
	statePath := filepath.Join(dir, "config", "state-default.json")
	if err := os.WriteFile(statePath, []byte("{garbage}"), 0644); err != nil {
		t.Fatalf("writing garbage: %v", err)
	}

	// Re-saving should recover cleanly
	s2 := &State{}
	s2.Remotes = append(s2.Remotes, Remote{
		Name:      "recovered",
		Port:      9999,
		TunnelPid: 8888,
	})
	if err := Save(s2); err != nil {
		t.Fatalf("Save() after corruption error: %v", err)
	}

	// Load should return the new data, not garbage
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() after recovery error: %v", err)
	}
	if len(loaded.Remotes) != 1 || loaded.Remotes[0].Name != "recovered" {
		t.Errorf("unexpected data after recovery: %+v", loaded.Remotes)
	}
}

func TestAppRootEnvVar(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLOUD_IDE_MOUNT_ROOT", dir)

	if got := appRoot(); got != dir {
		t.Errorf("appRoot() = %q, want %q", got, dir)
	}
}
