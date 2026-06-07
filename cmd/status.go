package cmd

import (
	"fmt"
	"os"

	"cloud-ide-mount/internal/state"
	"cloud-ide-mount/internal/tunnel"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show mount/tunnel health",
	RunE: func(_ *cobra.Command, _ []string) error {
		s, err := state.Load()
		if err != nil {
			return err
		}
		fmt.Println()
		if s == nil {
			fmt.Println("  No state file. Nothing mounted.")
			return nil
		}

		fmt.Println("  DRIVE   MODE       LABEL                          TUNNEL   PID")
		fmt.Println("  ─────   ────       ─────                          ──────   ───")

		for _, m := range s.Mounts {
			_, driveErr := os.Stat(m.Drive + "\\")
			driveOk := driveErr == nil

			label := "(combined)"
			if m.Mode != "combined" {
				label = m.Codespace
			}

			var relatedRemotes []state.Remote
			if m.Mode == "separate" {
				for _, r := range s.Remotes {
					if r.Codespace == m.Codespace {
						relatedRemotes = append(relatedRemotes, r)
					}
				}
			} else {
				relatedRemotes = s.Remotes
			}

			tunnelOk := true
			for _, r := range relatedRemotes {
				if !tunnel.PortOpen(r.Port) {
					tunnelOk = false
					break
				}
			}

			driveStr := m.Drive
			if !driveOk {
				driveStr = m.Drive + " ✗"
			}
			tunStr := "open"
			if !tunnelOk {
				tunStr = "closed"
			}

			fmt.Printf("  %-7s %-10s %-32s %-8s %d\n", driveStr, m.Mode, label, tunStr, m.RclonePid)
		}

		fmt.Println()
		fmt.Println("  Remotes:")
		for _, r := range s.Remotes {
			portStatus := "open"
			if !tunnel.PortOpen(r.Port) {
				portStatus = "closed"
			}
			fmt.Printf("    %-30s port %-6d %s\n", r.Name, r.Port, portStatus)
		}
		fmt.Println()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
