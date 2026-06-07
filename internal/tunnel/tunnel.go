package tunnel

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"cloud-ide-mount/internal/executil"
)

func PortOpen(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func WaitPort(port int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if PortOpen(port) {
			return true
		}
		time.Sleep(time.Second)
	}
	return false
}

// NextFreePort returns the next free port starting from 'from'.
// Deprecated: Use AllocatePort for atomic allocation that avoids TOCTOU races.
func NextFreePort(from int) int {
	ap, err := AllocatePort(from)
	if err != nil {
		return from
	}
	p := ap.Port
	ap.Close()
	return p
}

func StartTunnel(csName string, localPort, remotePort int) (*exec.Cmd, error) {
	args := []string{
		"cs", "ssh", "-c", csName,
		"--", "-N", "-L", fmt.Sprintf("127.0.0.1:%d:127.0.0.1:%d", localPort, remotePort),
	}
	cmd := exec.Command("gh", args...)
	cmd.SysProcAttr = executil.SysProcAttr()
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting tunnel for %s: %w", csName, err)
	}
	return cmd, nil
}

// execSSHCommand is a function variable for mocking in tests.
var execSSHCommand = defaultExecSSHCommand

func defaultExecSSHCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// DetectSSHPort attempts to detect the SSH port inside a codespace by reading
// /etc/ssh/sshd_config via gh cs ssh. It retries up to 3 times and defaults to 22.
func DetectSSHPort(csName string) int {
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Second)
		}

		// First try without sudo (sshd_config is usually world-readable)
		out, err := execSSHCommand("gh", "cs", "ssh", "-c", csName, "--",
			"cat", "/etc/ssh/sshd_config")
		if err != nil {
			// Fallback: try with sudo
			out, err = execSSHCommand("gh", "cs", "ssh", "-c", csName, "--",
				"sudo", "cat", "/etc/ssh/sshd_config")
			if err != nil {
				continue
			}
		}

		port := parseSSHPort(out)
		if port != 0 {
			return port
		}
	}

	return 22
}

// parseSSHPort extracts the SSH Port directive from sshd_config content.
// Returns 0 if not found or invalid.
func parseSSHPort(config string) int {
	for _, line := range strings.Split(config, "\n") {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Match "Port <number>" (case-insensitive)
		if !strings.HasPrefix(strings.ToLower(line), "port ") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			port, err := strconv.Atoi(parts[1])
			if err == nil && port > 0 && port < 65536 {
				return port
			}
		}
	}
	return 0
}
