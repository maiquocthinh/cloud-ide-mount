package health

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"cloud-ide-mount/internal/codespace"
	"cloud-ide-mount/internal/state"
)

// ─── Status type ───────────────────────────────────────────────────

// Status represents the health status of a component.
type Status int

const (
	Alive Status = iota
	Dead
	Error
)

func (s Status) String() string {
	switch s {
	case Alive:
		return "alive"
	case Dead:
		return "dead"
	case Error:
		return "error"
	default:
		return "unknown"
	}
}

// ─── Health data types ─────────────────────────────────────────────

// TunnelStatus represents the health of a single tunnel/remote.
type TunnelStatus struct {
	Codespace string
	Port      int
	Pid       int
	PortOpen  Status // TCP port check
	CsState   string // codespace state from gh cs list, empty if unavailable
}

// MountStatus represents the health of a single mount.
type MountStatus struct {
	Drive     string
	RclonePid int
	Remote    string
	Process   Status // rclone process existence
	DriveOK   Status // mount path accessibility
}

// Report is the full health report.
type Report struct {
	Tunnels []TunnelStatus
	Mounts  []MountStatus
}

// ─── Function vars (mockable in tests) ────────────────────────────

var (
	findProcess = os.FindProcess
	osStat      = os.Stat
	dialTimeout = net.DialTimeout
)

// ─── Individual check functions ────────────────────────────────────

// CheckTunnelPort checks if a TCP port on localhost is open.
func CheckTunnelPort(port int) Status {
	conn, err := dialTimeout("tcp",
		fmt.Sprintf("127.0.0.1:%d", port),
		2*time.Second,
	)
	if err != nil {
		return Dead
	}
	conn.Close()
	return Alive
}

// CheckProcess checks if a process with the given PID exists.
// On Windows, os.FindProcess always succeeds even for dead processes,
// so this may report false positives on that platform.
func CheckProcess(pid int) Status {
	if pid <= 0 {
		return Dead
	}
	_, err := findProcess(pid)
	if err != nil {
		return Dead
	}
	return Alive
}

// CheckMountDrive checks if a mount drive or path is accessible.
func CheckMountDrive(drive string) Status {
	if drive == "" {
		return Dead
	}
	path := drive
	if runtime.GOOS == "windows" && !strings.HasSuffix(path, "\\") {
		path += "\\"
	}
	_, err := osStat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Dead
		}
		return Error
	}
	return Alive
}

// ─── Full health check ─────────────────────────────────────────────

// CheckAll performs all health checks and returns a Report.
func CheckAll(s *state.State, allCs []codespace.Codespace) Report {
	// Build codespace state map
	csStates := map[string]string{}
	for _, cs := range allCs {
		csStates[cs.Name] = cs.State
	}

	// Check tunnels
	tunnels := make([]TunnelStatus, len(s.Remotes))
	for i, r := range s.Remotes {
		tunnels[i] = TunnelStatus{
			Codespace: r.Codespace,
			Port:      r.Port,
			Pid:       r.TunnelPid,
			PortOpen:  CheckTunnelPort(r.Port),
			CsState:   csStates[r.Codespace],
		}
	}

	// Check mounts
	mounts := make([]MountStatus, len(s.Mounts))
	for i, m := range s.Mounts {
		mounts[i] = MountStatus{
			Drive:     m.Drive,
			RclonePid: m.RclonePid,
			Remote:    m.Remote,
			Process:   CheckProcess(m.RclonePid),
			DriveOK:   CheckMountDrive(m.Drive),
		}
	}

	return Report{Tunnels: tunnels, Mounts: mounts}
}
