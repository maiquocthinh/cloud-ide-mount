package state

import (
	"encoding/json"
	"os"
	"path/filepath"
)

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

type State struct {
	Remotes []Remote `json:"remotes"`
	Mounts  []Mount  `json:"mounts"`
}

func filePath() string {
	return filepath.Join(os.TempDir(), "cloud-ide-mount-state.json")
}

func Load() (*State, error) {
	data, err := os.ReadFile(filePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func Save(s *State) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath(), data, 0644)
}

func Remove() error {
	if err := os.Remove(filePath()); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
