package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
			fmt.Println("No active mounts found.")
			return nil
		}

		toUnmount := ui.SelectMountsToUnmount(s.Mounts)
		if len(toUnmount) == 0 {
			fmt.Println("Nothing selected.")
			return nil
		}

		if !Force {
			var labels []string
			for _, m := range toUnmount {
				labels = append(labels, m.Drive)
			}
			if !ui.Confirm(fmt.Sprintf("Unmount %s? [y/N]", strings.Join(labels, ", "))) {
				fmt.Println("Canceled.")
				return nil
			}
		}

		fmt.Println()

		unmountedDrives := map[string]bool{}
		for _, m := range toUnmount {
			unmountedDrives[m.Drive] = true
			fmt.Printf("  Stopping rclone: %s (PID %d)\n", m.Drive, m.RclonePid)
			killPid(m.RclonePid)
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
				fmt.Printf("  Stopping tunnel: %s (PID %d)\n", r.Codespace, r.TunnelPid)
				killPid(r.TunnelPid)
			}
			killPortListeners(r.Port)
			rclone.DeleteRemote(r.Name)
		}

		if len(remainingMounts) == 0 {
			rclone.DeleteRemote(CombineRemote)
			s.Remove()
			fmt.Println()
			fmt.Println("  All unmounted. State cleared.")
		} else {
			s.Remotes = remainingRemotes
			s.Mounts = remainingMounts
			state.Save(s)
			fmt.Println()
			fmt.Printf("  Unmounted. %d drive(s) still active.\n", len(remainingMounts))
		}

		if stopFlag && len(codespacesToStop) > 0 {
			for _, name := range codespacesToStop {
				fmt.Printf("  Stopping codespace: %s...\n", name)
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

func killPid(pid int) {
	p, err := os.FindProcess(pid)
	if err == nil {
		p.Kill()
	}
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
		killPid(pid)
	}
}
