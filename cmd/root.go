package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"cloud-ide-mount/internal/logging"

	"github.com/spf13/cobra"
)

var (
	StartPort     int
	KeyFile       string
	CombineRemote string
	Force         bool
)

var rootCmd = &cobra.Command{
	Use:   "cs-mount",
	Short: "Mount GitHub Codespaces as local drives via rclone",
	Long: `A CLI tool for mounting GitHub Codespaces as Windows drive letters.
Uses SSH tunnels and rclone to expose remote codespace filesystems as local drives.`,
	Run: func(cmd *cobra.Command, _ []string) {
		cmd.Help()
	},
}

func Execute() {
	// Initialize logging
	logPath := filepath.Join(logDir(), "cloud-ide-mount.log")
	if err := logging.Init(logging.Options{Level: logging.LevelInfo, Path: logPath}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to init logging: %v\n", err)
	}
	defer logging.Close()

	if err := rootCmd.Execute(); err != nil {
		logging.Error("Command failed", "error", err)
		os.Exit(1)
	}
}

// logDir returns the directory for log files.
// Priority: CLOUD_IDE_MOUNT_ROOT env var → executable directory.
func logDir() string {
	if root := os.Getenv("CLOUD_IDE_MOUNT_ROOT"); root != "" {
		return filepath.Join(root, "logs")
	}
	exe, err := os.Executable()
	if err != nil {
		return "logs"
	}
	return filepath.Join(filepath.Dir(exe), "logs")
}

func init() {
	defaultKey := os.ExpandEnv("${USERPROFILE}\\.ssh\\codespaces.auto")

	rootCmd.PersistentFlags().IntVar(&StartPort, "start-port", 2223, "First SSH tunnel port")
	rootCmd.PersistentFlags().StringVar(&KeyFile, "key-file", defaultKey, "SSH key for rclone SFTP")
	rootCmd.PersistentFlags().StringVar(&CombineRemote, "combine-remote", "codespaces-auto", "rclone combine remote name")
	rootCmd.PersistentFlags().BoolVarP(&Force, "force", "f", false, "Skip confirmation prompts")
}
