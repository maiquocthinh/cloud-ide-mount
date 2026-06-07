package cmd

import (
	"fmt"
	"os"

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
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	defaultKey := os.ExpandEnv("${USERPROFILE}\\.ssh\\codespaces.auto")

	rootCmd.PersistentFlags().IntVar(&StartPort, "start-port", 2223, "First SSH tunnel port")
	rootCmd.PersistentFlags().StringVar(&KeyFile, "key-file", defaultKey, "SSH key for rclone SFTP")
	rootCmd.PersistentFlags().StringVar(&CombineRemote, "combine-remote", "codespaces-auto", "rclone combine remote name")
	rootCmd.PersistentFlags().BoolVarP(&Force, "force", "f", false, "Skip confirmation prompts")
}
