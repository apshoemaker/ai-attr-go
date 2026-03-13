package installers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// InstallCopilot installs Copilot agent hooks via .github/hooks/ai-attr.json.
func InstallCopilot(repoWorkdir, binaryPath string) error {
	hooksDir := filepath.Join(repoWorkdir, ".github", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(hooksDir, "ai-attr.json")
	command := fmt.Sprintf("%s checkpoint copilot --hook-input", binaryPath)

	config := map[string]interface{}{
		"agent": "copilot",
		"hooks": map[string]interface{}{
			"PreToolUse": map[string]interface{}{
				"matcher": "insert_edit_into_file|create_file",
				"command": command,
			},
			"PostToolUse": map[string]interface{}{
				"matcher": "insert_edit_into_file|create_file",
				"command": command,
			},
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}
