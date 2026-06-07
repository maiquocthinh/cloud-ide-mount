package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"cloud-ide-mount/internal/codespace"
	"cloud-ide-mount/internal/ide"
	"cloud-ide-mount/internal/state"

	"github.com/spf13/cobra"
)

type entry struct {
	Codespace  string
	Repository string
	Port       int
	RemotePath string
}

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open a codespace in an IDE via SSH",
	Long:  `Opens a mounted codespace in VS Code, IntelliJ IDEA, or Zed via SSH remote protocol.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		s, err := state.Load()
		if err != nil {
			return err
		}
		if s == nil || len(s.Remotes) == 0 {
			return fmt.Errorf("no active tunnels — run mount first")
		}

		allCs, err := codespace.List()
		if err != nil {
			return fmt.Errorf("listing codespaces: %w", err)
		}
		csRepo := map[string]string{}
		for _, cs := range allCs {
			csRepo[cs.Name] = cs.Repository
		}

		var entries []entry
		for _, r := range s.Remotes {
			repo := csRepo[r.Codespace]
			if repo == "" {
				repo = r.Codespace
			}
			entries = append(entries, entry{
				Codespace:  r.Codespace,
				Repository: repo,
				Port:       r.Port,
				RemotePath: workspacePath(repo),
			})
		}

		fmt.Println()
		fmt.Println("  #   CODESPACE          REPO                PORT")
		fmt.Println("  ─   ─────────          ────                ────")
		for i, e := range entries {
			fmt.Printf("  %2d  %-18s %-19s %d\n", i+1,
				truncate(e.Codespace, 18), truncate(e.Repository, 19), e.Port)
		}
		fmt.Println()

		var ent *entry
		if len(entries) == 1 {
			ent = &entries[0]
			fmt.Printf("  Selected: %s\n", ent.Codespace)
		} else {
			n := readInt(fmt.Sprintf("  Select codespace [1-%d]", len(entries)), 1, len(entries))
			ent = &entries[n-1]
		}

		fmt.Printf("  Remote path [%s] → ", ent.RemotePath)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if p := strings.TrimSpace(scanner.Text()); p != "" {
			ent.RemotePath = p
		}

		avail := ide.Available()
		if len(avail) == 0 {
			return fmt.Errorf("no supported IDEs found on PATH")
		}

		fmt.Println()
		fmt.Println("  Available IDEs:")
		for i, id := range avail {
			fmt.Printf("    [%d] %s\n", i+1, id.Name)
		}
		fmt.Println()

		n := readInt(fmt.Sprintf("  Select IDE [1-%d]", len(avail)), 1, len(avail))
		selected := avail[n-1]

		info := ide.SSHInfo{
			Host:       "127.0.0.1",
			Port:       ent.Port,
			User:       "codespace",
			KeyFile:    KeyFile,
			Alias:      "cs-" + codespace.SafeName(ent.Codespace),
			RemotePath: ent.RemotePath,
		}

		fmt.Printf("  Opening %s via SSH with %s...\n", ent.Codespace, selected.Name)
		return ide.Open(selected, info)
	},
}

func workspacePath(repo string) string {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) == 2 {
		return "/workspaces/" + parts[1]
	}
	return "/workspaces/" + repo
}

func readInt(prompt string, min, max int) int {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("%s → ", prompt)
		scanner.Scan()
		raw := strings.TrimSpace(scanner.Text())
		n, err := strconv.Atoi(raw)
		if err == nil && n >= min && n <= max {
			return n
		}
		fmt.Printf("  Enter a number between %d and %d.\n", min, max)
	}
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n-1] + "…"
	}
	return s
}

func init() {
	rootCmd.AddCommand(openCmd)
}
