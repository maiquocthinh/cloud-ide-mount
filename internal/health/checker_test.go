package health

import (
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"cloud-ide-mount/internal/codespace"
	"cloud-ide-mount/internal/state"
)

func TestCheckTunnelPort_Alive(t *testing.T) {
	// Start a local listener
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port

	// Override dialTimeout to use the real one
	got := CheckTunnelPort(port)
	if got != Alive {
		t.Errorf("CheckTunnelPort(%d) = %v, want %v", port, got, Alive)
	}
}

func TestCheckTunnelPort_Dead(t *testing.T) {
	// Use a port that's unlikely to be open
	got := CheckTunnelPort(0)
	if got != Dead {
		t.Errorf("CheckTunnelPort(0) = %v, want %v", got, Dead)
	}
}

func TestCheckTunnelPort_DialTimeoutOverride(t *testing.T) {
	// Override dialTimeout to test error handling
	old := dialTimeout
	t.Cleanup(func() { dialTimeout = old })

	dialTimeout = func(network, address string, timeout time.Duration) (net.Conn, error) {
		return nil, errors.New("mock dial error")
	}

	got := CheckTunnelPort(9999)
	if got != Dead {
		t.Errorf("with mock dial error = %v, want %v", got, Dead)
	}
}

func TestCheckProcess_Alive(t *testing.T) {
	// Current process should always exist
	got := CheckProcess(os.Getpid())
	if got != Alive {
		t.Errorf("CheckProcess(%d) = %v, want %v", os.Getpid(), got, Alive)
	}
}

func TestCheckProcess_Dead(t *testing.T) {
	// PID 0 or negative should be dead
	got := CheckProcess(0)
	if got != Dead {
		t.Errorf("CheckProcess(0) = %v, want %v", got, Dead)
	}
}

func TestCheckProcess_NonExistent(t *testing.T) {
	// Try a very large PID that likely doesn't exist
	got := CheckProcess(999999999)
	if got == Alive {
		t.Log("CheckProcess returned alive — likely running on Windows where FindProcess always succeeds")
	}
}

func TestCheckProcess_FindProcessOverride(t *testing.T) {
	old := findProcess
	t.Cleanup(func() { findProcess = old })

	// Mock findProcess to return an error
	findProcess = func(pid int) (*os.Process, error) {
		return nil, errors.New("mock find error")
	}

	got := CheckProcess(42)
	if got != Dead {
		t.Errorf("with mock find error = %v, want %v", got, Dead)
	}
}

func TestCheckMountDrive_Alive(t *testing.T) {
	// Temp dir should always exist
	got := CheckMountDrive(os.TempDir())
	if got != Alive {
		t.Errorf("CheckMountDrive(%q) = %v, want %v", os.TempDir(), got, Alive)
	}
}

func TestCheckMountDrive_Empty(t *testing.T) {
	got := CheckMountDrive("")
	if got != Dead {
		t.Errorf("CheckMountDrive('') = %v, want %v", got, Dead)
	}
}

func TestCheckMountDrive_NonExistent(t *testing.T) {
	got := CheckMountDrive("/nonexistent/path/that/does/not/exist")
	if got != Dead {
		t.Errorf("CheckMountDrive(nonexistent) = %v, want %v", got, Dead)
	}
}

func TestCheckMountDrive_OSStatOverride(t *testing.T) {
	old := osStat
	t.Cleanup(func() { osStat = old })

	// Mock osStat to return an error that's not IsNotExist
	osStat = func(name string) (os.FileInfo, error) {
		return nil, errors.New("permission denied")
	}

	got := CheckMountDrive("C:\\")
	if got != Error {
		t.Errorf("with mock permission error = %v, want %v", got, Error)
	}
}

func TestCheckAll_Empty(t *testing.T) {
	s := &state.State{}
	report := CheckAll(s, nil)

	if len(report.Tunnels) != 0 {
		t.Errorf("expected 0 tunnels, got %d", len(report.Tunnels))
	}
	if len(report.Mounts) != 0 {
		t.Errorf("expected 0 mounts, got %d", len(report.Mounts))
	}
}

func TestCheckAll_WithData(t *testing.T) {
	s := &state.State{
		Remotes: []state.Remote{
			{
				Name:      "cs-test",
				Codespace: "test-codespace",
				Port:      0, // will be dead
				TunnelPid: 999999999,
			},
		},
		Mounts: []state.Mount{
			{
				Drive:     os.TempDir(),
				RclonePid: os.Getpid(),
				Remote:    "test-remote",
			},
		},
	}

	allCs := []codespace.Codespace{
		{Name: "test-codespace", State: "Available"},
	}

	report := CheckAll(s, allCs)

	if len(report.Tunnels) != 1 {
		t.Fatalf("expected 1 tunnel, got %d", len(report.Tunnels))
	}
	if len(report.Mounts) != 1 {
		t.Fatalf("expected 1 mount, got %d", len(report.Mounts))
	}

	// Tunnel port 0 should be dead
	if report.Tunnels[0].PortOpen != Dead {
		t.Errorf("tunnel port 0 should be dead, got %v", report.Tunnels[0].PortOpen)
	}
	// Codespace state should be set
	if report.Tunnels[0].CsState != "Available" {
		t.Errorf("CsState = %q, want %q", report.Tunnels[0].CsState, "Available")
	}

	// Mount drive should be alive (TempDir exists)
	if report.Mounts[0].DriveOK != Alive {
		t.Errorf("mount drive should be alive, got %v", report.Mounts[0].DriveOK)
	}
}

func TestCheckAll_MissingCodespace(t *testing.T) {
	s := &state.State{
		Remotes: []state.Remote{
			{
				Codespace: "unknown-codespace",
				Port:      0,
			},
		},
	}

	// allCs doesn't include the codespace
	allCs := []codespace.Codespace{
		{Name: "other", State: "Available"},
	}

	report := CheckAll(s, allCs)
	if len(report.Tunnels) != 1 {
		t.Fatalf("expected 1 tunnel, got %d", len(report.Tunnels))
	}
	// CsState should be empty (not found in allCs)
	if report.Tunnels[0].CsState != "" {
		t.Errorf("CsState should be empty for unknown codespace, got %q", report.Tunnels[0].CsState)
	}
}

func TestStatus_String(t *testing.T) {
	tests := []struct {
		s    Status
		want string
	}{
		{Alive, "alive"},
		{Dead, "dead"},
		{Error, "error"},
		{Status(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.s.String(); got != tt.want {
			t.Errorf("Status(%d).String() = %q, want %q", tt.s, got, tt.want)
		}
	}
}
