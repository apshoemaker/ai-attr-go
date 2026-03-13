package installers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// InstallCodex installs Codex CLI notify hook in a codex config dir.
// configDir overrides the default ~/.codex/ directory (for testing).
func InstallCodex(binaryPath, configDir string) error {
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not determine home directory: %w", err)
		}
		configDir = filepath.Join(home, ".codex")
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.toml")
	notifyLine := fmt.Sprintf("notify = \"%s checkpoint codex --hook-input\"", binaryPath)

	data, err := os.ReadFile(configPath)
	if err == nil {
		content := string(data)
		if strings.Contains(content, "ai-attr") {
			// Update existing line
			var newLines []string
			for _, line := range strings.Split(content, "\n") {
				if strings.Contains(line, "ai-attr") && strings.HasPrefix(strings.TrimSpace(line), "notify") {
					newLines = append(newLines, notifyLine)
				} else {
					newLines = append(newLines, line)
				}
			}
			return os.WriteFile(configPath, []byte(strings.Join(newLines, "\n")), 0644)
		}
		// Append
		newContent := fmt.Sprintf("%s\n%s\n", strings.TrimRight(content, "\n"), notifyLine)
		return os.WriteFile(configPath, []byte(newContent), 0644)
	}

	// Create new
	return os.WriteFile(configPath, []byte(notifyLine+"\n"), 0644)
}
