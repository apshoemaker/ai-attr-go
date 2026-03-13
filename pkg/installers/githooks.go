package installers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const aiAttrMarker = "# ai-attr post-commit hook"

// InstallGitHook installs the git post-commit hook.
func InstallGitHook(repoWorkdir, binaryPath string) error {
	hooksDir := filepath.Join(repoWorkdir, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return err
	}

	hookPath := filepath.Join(hooksDir, "post-commit")
	hookLine := fmt.Sprintf("%s\n%s commit\n", aiAttrMarker, binaryPath)

	if data, err := os.ReadFile(hookPath); err == nil {
		existing := string(data)
		if strings.Contains(existing, aiAttrMarker) {
			return nil // already installed
		}
		newContent := fmt.Sprintf("%s\n%s", strings.TrimRight(existing, "\n"), hookLine)
		if err := os.WriteFile(hookPath, []byte(newContent), 0755); err != nil {
			return err
		}
	} else {
		content := fmt.Sprintf("#!/bin/sh\n%s", hookLine)
		if err := os.WriteFile(hookPath, []byte(content), 0755); err != nil {
			return err
		}
	}

	return os.Chmod(hookPath, 0755)
}

// UninstallGitHook removes the ai-attr line from the post-commit hook.
func UninstallGitHook(repoWorkdir string) error {
	hookPath := filepath.Join(repoWorkdir, ".git", "hooks", "post-commit")
	data, err := os.ReadFile(hookPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	var newLines []string
	skipNext := false

	for _, line := range lines {
		if skipNext {
			skipNext = false
			continue
		}
		if strings.Contains(line, aiAttrMarker) {
			skipNext = true
			continue
		}
		if strings.Contains(line, "ai-attr commit") {
			continue
		}
		newLines = append(newLines, line)
	}

	result := strings.TrimSpace(strings.Join(newLines, "\n"))
	if result == "" || result == "#!/bin/sh" {
		return os.Remove(hookPath)
	}
	return os.WriteFile(hookPath, []byte(result+"\n"), 0755)
}
