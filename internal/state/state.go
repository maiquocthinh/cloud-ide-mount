package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// fileMu serializes file I/O to prevent races on the shared state file.
var fileMu sync.Mutex

// State represents the persisted application state.
type State struct {
	mu      sync.RWMutex
	Remotes []Remote `json:"remotes"`
	Mounts  []Mount  `json:"mounts"`
}

type Remote struct {
	Name       string `json:"name"`
	Codespace  string `json:"codespace"`
	Port       int    `json:"port"`
	TunnelPid  int    `json:"tunnel_pid"`
	FolderPath string `json:"folder_path"`
}

type Mount struct {
	Drive     string `json:"drive"`
	RclonePid int    `json:"rclone_pid"`
	Remote    string `json:"remote"`
	Codespace string `json:"codespace,omitempty"`
	Mode      string `json:"mode"` // "combined" or "separate"
}

// appRoot returns the application root directory.
// Priority: CLOUD_IDE_MOUNT_ROOT env var → executable directory.
func appRoot() string {
	if root := os.Getenv("CLOUD_IDE_MOUNT_ROOT"); root != "" {
		return root
	}
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func filePath() string {
	return filepath.Join(appRoot(), "config", "state.json")
}

// Load reads state from disk. Returns an empty State if the file does not exist.
func Load() (*State, error) {
	s := &State{}
	data, err := os.ReadFile(filePath())
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("reading state file: %w", err)
	}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("parsing state file: %w", err)
	}
	return s, nil
}

// Save atomically writes state to disk using a temp file + rename.
func Save(s *State) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fileMu.Lock()
	defer fileMu.Unlock()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	dir := filepath.Dir(filePath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating state dir %s: %w", dir, err)
	}

	// Atomic write: write to temp file, then rename
	tmp, err := os.CreateTemp(dir, "state-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing temp file: %w", err)
	}
	tmp.Close()

	if err := os.Rename(tmpPath, filePath()); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renaming temp to state file: %w", err)
	}

	return nil
}

// Remove deletes the state file from disk.
func (s *State) Remove() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fileMu.Lock()
	defer fileMu.Unlock()

	if err := os.Remove(filePath()); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing state file: %w", err)
	}
	return nil
}
