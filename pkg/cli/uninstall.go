package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apshoemaker/ai-attr/pkg/installers"
	"github.com/apshoemaker/ai-attr/pkg/storage"
)

// RunUninstall removes hooks and agent configurations.
func RunUninstall() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	_, workDir, err := storage.FindGitDir(cwd)
	if err != nil {
		return err
	}
	return UninstallAt(workDir)
}

// UninstallAt removes hooks and configs for the given repo (testable).
func UninstallAt(repoWorkdir string) error {
	installers.UninstallGitHook(repoWorkdir)
	installers.UninstallClaude(repoWorkdir)

	aiAttrDir := filepath.Join(repoWorkdir, ".git", "ai-attr")
	if _, err := os.Stat(aiAttrDir); err == nil {
		os.RemoveAll(aiAttrDir)
	}

	fmt.Fprintln(os.Stderr, "ai-attr: uninstalled successfully")
	return nil
}
