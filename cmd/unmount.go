package cmd

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"cloud-ide-mount/internal/executil"
	"cloud-ide-mount/internal/logging"
	"cloud-ide-mount/internal/rclone"
	"cloud-ide-mount/internal/state"
	"cloud-ide-mount/internal/ui"

	"github.com/spf13/cobra"
)

var stopFlag bool

var unmountCmd = &cobra.Command{
	Use:   "unmount",
	Short: "Unmount mounted codespace drives",
	RunE: func(_ *cobra.Command, _ []string) error {
		s, err := state.Load()
		if err != nil {
			return err
		}
		if s == nil || len(s.Mounts) == 0 {
			logging.Info("No active mounts found.")
			return nil
		}

		toUnmount := ui.SelectMountsToUnmount(s.Mounts)
		if len(toUnmount) == 0 {
			logging.Info("Nothing selected.")
			return nil
		}

		if !Force {
			var labels []string
			for _, m := range toUnmount {
				labels = append(labels, m.Drive)
			}
			if !ui.Confirm(fmt.Sprintf("Unmount %s? [y/N]", strings.Join(labels, ", "))) {
				logging.Info("Canceled.")
				return nil
			}
		}

		fmt.Println()

		unmountedDrives := map[string]bool{}
		for _, m := range toUnmount {
			unmountedDrives[m.Drive] = true
			logging.Info(fmt.Sprintf("Stopping rclone: %s (PID %d)", m.Drive, m.RclonePid), "drive", m.Drive, "pid", m.RclonePid)
			if err := executil.KillProcess(m.RclonePid, 5*time.Second); err != nil {
				logging.Warn(fmt.Sprintf("error stopping rclone PID %d: %v", m.RclonePid, err), "pid", m.RclonePid, "error", err)
			}
		}

		var remainingMounts []state.Mount
		for _, m := range s.Mounts {
			if !unmountedDrives[m.Drive] {
				remainingMounts = append(remainingMounts, m)
			}
		}

		remainingCodespaces := map[string]bool{}
		for _, m := range remainingMounts {
			if m.Codespace != "" {
				remainingCodespaces[m.Codespace] = true
			}
		}

		// Collect codespaces that will be stopped
		var codespacesToStop []string
		for _, r := range s.Remotes {
			if !remainingCodespaces[r.Codespace] && r.Codespace != "" {
				codespacesToStop = append(codespacesToStop, r.Codespace)
			}
		}

		var remainingRemotes []state.Remote
		for _, r := range s.Remotes {
			if remainingCodespaces[r.Codespace] {
				remainingRemotes = append(remainingRemotes, r)
				continue
			}
			if r.TunnelPid > 0 {
				logging.Info(fmt.Sprintf("Stopping tunnel: %s (PID %d)", r.Codespace, r.TunnelPid), "codespace", r.Codespace, "pid", r.TunnelPid)
				if err := executil.KillProcess(r.TunnelPid, 5*time.Second); err != nil {
					logging.Warn(fmt.Sprintf("error stopping tunnel PID %d: %v", r.TunnelPid, err), "pid", r.TunnelPid, "error", err)
				}
			}
			killPortListeners(r.Port)
			if err := rclone.DeleteRemote(r.Name); err != nil {
				logging.Warn(fmt.Sprintf("error deleting remote %s: %v", r.Name, err), "remote", r.Name, "error", err)
			}
		}

		if len(remainingMounts) == 0 {
			if err := rclone.DeleteRemote(CombineRemote); err != nil {
				logging.Warn(fmt.Sprintf("error deleting combine remote %s: %v", CombineRemote, err), "remote", CombineRemote, "error", err)
			}
			if err := s.Remove(); err != nil {
				return fmt.Errorf("clearing state file: %w", err)
			}
			fmt.Println()
			logging.Info("All unmounted. State cleared.")
		} else {
			s.Remotes = remainingRemotes
			s.Mounts = remainingMounts
			if err := state.Save(s); err != nil {
				return fmt.Errorf("saving state after unmount: %w", err)
			}
			fmt.Println()
			logging.Info(fmt.Sprintf("Unmounted. %d drive(s) still active.", len(remainingMounts)), "activeMounts", len(remainingMounts))
		}

		if stopFlag && len(codespacesToStop) > 0 {
			for _, name := range codespacesToStop {
				logging.Info(fmt.Sprintf("Stopping codespace: %s...", name), "codespace", name)
				execCmdOutput("gh", "cs", "stop", "-c", name)
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(unmountCmd)
	unmountCmd.Flags().BoolVarP(&stopFlag, "stop", "s", false, "Stop codespace(s) after unmounting")
}

func killPortListeners(port int) {
	out, err := exec.Command("powershell", "-c",
		fmt.Sprintf(`Get-NetTCPConnection -LocalPort %d -State Listen -ErrorAction SilentlyContinue | ForEach-Object { $_.OwningProcess }`, port),
	).Output()
	if err != nil {
		return
	}
	pidStr := strings.TrimSpace(string(out))
	if pidStr == "" {
		return
	}
	var pid int
	fmt.Sscanf(pidStr, "%d", &pid)
	if pid > 0 {
		if err := executil.KillProcess(pid, 5*time.Second); err != nil {
			logging.Warn(fmt.Sprintf("error killing port listener PID %d: %v", pid, err), "pid", pid, "error", err)
		}
	}
}
