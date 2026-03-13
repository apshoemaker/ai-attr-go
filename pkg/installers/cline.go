package installers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// InstallCline installs Cline hooks via .clinerules/hooks/ai-attr.json.
func InstallCline(repoWorkdir, binaryPath string) error {
	hooksDir := filepath.Join(repoWorkdir, ".clinerules", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(hooksDir, "ai-attr.json")
	command := fmt.Sprintf("%s checkpoint cline --hook-input", binaryPath)

	config := map[string]interface{}{
		"agent": "cline",
		"hooks": map[string]interface{}{
			"PreToolUse": map[string]interface{}{
				"command": command,
			},
			"PostToolUse": map[string]interface{}{
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
