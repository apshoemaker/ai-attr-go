package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/apshoemaker/ai-attr/pkg/cli"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:     "ai-attr",
		Short:   "Line-level AI code attribution via native agent hooks",
		Version: version,
	}

	// checkpoint
	checkpointCmd := &cobra.Command{
		Use:   "checkpoint <agent>",
		Short: "Record a checkpoint (called by agent hooks)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hookInput, _ := cmd.Flags().GetBool("hook-input")
			return cli.RunCheckpoint(args[0], hookInput)
		},
	}
	checkpointCmd.Flags().Bool("hook-input", false, "Read hook input from stdin")
	rootCmd.AddCommand(checkpointCmd)

	// commit
	commitCmd := &cobra.Command{
		Use:   "commit",
		Short: "Consolidate checkpoints into a git note (called by post-commit hook)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunCommit()
		},
	}
	rootCmd.AddCommand(commitCmd)

	// install
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install git hook and agent configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			agentsStr, _ := cmd.Flags().GetString("agents")
			var agents []string
			if agentsStr != "" {
				agents = strings.Split(agentsStr, ",")
			}
			return cli.RunInstall(agents)
		},
	}
	installCmd.Flags().String("agents", "", "Agents to configure (claude,copilot,codex,cline)")
	rootCmd.AddCommand(installCmd)

	// blame
	blameCmd := &cobra.Command{
		Use:   "blame <file>",
		Short: "Display line-level AI attribution for a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonFlag, _ := cmd.Flags().GetBool("json")
			return cli.RunBlame(args[0], jsonFlag)
		},
	}
	blameCmd.Flags().Bool("json", false, "Output as JSON")
	rootCmd.AddCommand(blameCmd)

	// show
	showCmd := &cobra.Command{
		Use:   "show [commit]",
		Short: "Display the attribution note for a commit",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			commit := ""
			if len(args) > 0 {
				commit = args[0]
			}
			return cli.RunShow(commit)
		},
	}
	rootCmd.AddCommand(showCmd)

	// stats
	statsCmd := &cobra.Command{
		Use:   "stats [range]",
		Short: "Display AI composition statistics",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			commitRange := ""
			if len(args) > 0 {
				commitRange = args[0]
			}
			jsonFlag, _ := cmd.Flags().GetBool("json")
			return cli.RunStats(commitRange, jsonFlag)
		},
	}
	statsCmd.Flags().Bool("json", false, "Output as JSON")
	rootCmd.AddCommand(statsCmd)

	// uninstall
	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove hooks and agent configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunUninstall()
		},
	}
	rootCmd.AddCommand(uninstallCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "ai-attr: error: %v\n", err)
		os.Exit(1)
	}
}
