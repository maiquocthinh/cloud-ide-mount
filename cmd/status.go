package cmd

import (
	"fmt"

	"cloud-ide-mount/internal/codespace"
	"cloud-ide-mount/internal/health"
	"cloud-ide-mount/internal/logging"
	"cloud-ide-mount/internal/state"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show health status of tunnels and mounts",
	Long: `Displays the current health of all active SSH tunnels and rclone mounts.
Checks tunnel TCP connectivity, rclone process existence, and mount path accessibility.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		s, err := state.Load()
		if err != nil {
			return fmt.Errorf("loading state: %w", err)
		}
		if s == nil || (len(s.Remotes) == 0 && len(s.Mounts) == 0) {
			fmt.Println("  No active tunnels or mounts.")
			return nil
		}

		allCs, err := codespace.List()
		if err != nil {
			logging.Warn("Cannot fetch codespace list (status will be partial)", "error", err)
			allCs = nil
		}

		report := health.CheckAll(s, allCs)
		displayStatus(report)
		return nil
	},
}

func displayStatus(r health.Report) {
	fmt.Println()
	fmt.Println("  Status Report")
	fmt.Println("  ─────────────")
	fmt.Println()

	if len(r.Tunnels) > 0 {
		fmt.Println("  Tunnels:")
		fmt.Println("  ┌───────────────────────────────────────────────────────────────────┐")
		fmt.Printf("  │ %-30s %5s  %-8s  %-12s │\n", "CODESPACE", "PORT", "TUNNEL", "CS STATE")
		for _, t := range r.Tunnels {
			icon := iconFor(t.PortOpen)
			stateStr := t.CsState
			if stateStr == "" {
				stateStr = "N/A"
			}
			fmt.Printf("  │ %-30s %5d  %-8s  %-12s │\n", truncateStr(t.Codespace, 30), t.Port, icon, stateStr)
		}
		fmt.Println("  └───────────────────────────────────────────────────────────────────┘")
		fmt.Println()
	}

	if len(r.Mounts) > 0 {
		fmt.Println("  Mounts:")
		fmt.Println("  ┌──────────────────────────────────────────────────────────────┐")
		fmt.Printf("  │ %-5s %-8s  %-8s  %-20s │\n", "DRIVE", "RCLONE", "MOUNT", "REMOTE")
		for _, m := range r.Mounts {
			procIcon := iconFor(m.Process)
			driveIcon := iconFor(m.DriveOK)
			fmt.Printf("  │ %-5s %-8s  %-8s  %-20s │\n", m.Drive, procIcon, driveIcon, truncateStr(m.Remote, 20))
		}
		fmt.Println("  └──────────────────────────────────────────────────────────────┘")
		fmt.Println()
	}
}

func iconFor(s health.Status) string {
	switch s {
	case health.Alive:
		return "✅ alive"
	case health.Dead:
		return "❌ dead"
	case health.Error:
		return "⚠️ error"
	default:
		return "? unknown"
	}
}

func truncateStr(s string, n int) string {
	if len(s) > n {
		return s[:n-1] + "…"
	}
	return s
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
