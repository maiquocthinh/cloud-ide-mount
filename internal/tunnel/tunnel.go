package tunnel

import (
	"fmt"
	"net"
	"os/exec"
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

func NextFreePort(from int) int {
	p := from
	for PortOpen(p) {
		p++
	}
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

// GetCsSshPort probes the SSH port used by a codespace tunnel.
// It connects to the tunnel, reads the SSH banner, then probes port 22.
func GetCsSshPort(csName string, localPort int) int {
	for attempt := 0; attempt < 10; attempt++ {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", localPort), 2*time.Second)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		buf := make([]byte, 64)
		n, _ := conn.Read(buf)
		conn.Close()

		if n > 0 && strings.HasPrefix(string(buf[:n]), "SSH-") {
			// Probe port 22
			probe := exec.Command("gh", "cs", "ssh", "-c", csName, "--",
				"-o", "ConnectTimeout=3",
				"-p", "22",
				"-o", "BatchMode=yes",
				"-o", "StrictHostKeyChecking=no",
				"exit")
			if err := probe.Run(); err == nil {
				return 22
			}
			return 2222
		}
		time.Sleep(500 * time.Millisecond)
	}
	return 2222
}
