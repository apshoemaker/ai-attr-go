package cli

import (
	"fmt"
	"os"

	"github.com/apshoemaker/ai-attr/pkg/installers"
	"github.com/apshoemaker/ai-attr/pkg/storage"
)

// RunInstall installs git hook and agent configurations.
func RunInstall(agents []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	_, workDir, err := storage.FindGitDir(cwd)
	if err != nil {
		return err
	}

	binaryPath, err := os.Executable()
	if err != nil {
		return err
	}

	return InstallAt(workDir, binaryPath, agents)
}

// InstallAt installs hooks for the given repo (testable).
func InstallAt(repoWorkdir, binaryPath string, agents []string) error {
	if err := installers.InstallGitHook(repoWorkdir, binaryPath); err != nil {
		return err
	}

	if len(agents) == 0 {
		agents = []string{"claude"}
	}

	for _, agent := range agents {
		switch agent {
		case "claude":
			if err := installers.InstallClaude(repoWorkdir, binaryPath); err != nil {
				return err
			}
		case "copilot":
			if err := installers.InstallCopilot(repoWorkdir, binaryPath); err != nil {
				return err
			}
		case "cline":
			if err := installers.InstallCline(repoWorkdir, binaryPath); err != nil {
				return err
			}
		case "codex":
			if err := installers.InstallCodex(binaryPath, ""); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown agent: %s. Supported: claude, copilot, cline, codex", agent)
		}
	}

	fmt.Fprintln(os.Stderr, "ai-attr: installed successfully")
	return nil
}
