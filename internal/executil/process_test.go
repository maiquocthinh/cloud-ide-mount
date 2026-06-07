package executil

import (
	"os/exec"
	"testing"
	"time"
)

func TestKillProcess(t *testing.T) {
	// Start a long-running sleep process that we can kill
	cmd := exec.Command("sleep", "60")
	if err := cmd.Start(); err != nil {
		t.Fatalf("starting test process: %v", err)
	}

	pid := cmd.Process.Pid

	// Kill it with a reasonable timeout
	start := time.Now()
	err := KillProcess(pid, 5*time.Second)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("KillProcess() error: %v", err)
	}

	// Ensure the kill completed quickly (no arbitrary sleep).
	// If we used a plain sleep(1) it would take at least 1s.
	if elapsed > 3*time.Second {
		t.Errorf("KillProcess() took %v, expected < 3s", elapsed)
	}
}

func TestKillProcessNonExistent(t *testing.T) {
	// Use a PID that's guaranteed to not exist.
	err := KillProcess(0, time.Second)
	if err == nil {
		t.Error("KillProcess(0) expected error, got nil")
	}
}
