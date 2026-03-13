package cli

import (
	"fmt"
	"os"

	"github.com/apshoemaker/ai-attr/pkg/storage"
)

// RunShow displays the raw attribution note for a commit.
func RunShow(commit string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	sha := commit
	if sha == "" {
		sha, err = storage.GetHeadSHA(cwd)
		if err != nil {
			return err
		}
	}

	content, ok, err := storage.NotesShow(cwd, sha)
	if err != nil {
		return err
	}
	if ok {
		fmt.Print(content)
	} else {
		shortSHA := sha
		if len(shortSHA) > 8 {
			shortSHA = shortSHA[:8]
		}
		fmt.Fprintf(os.Stderr, "No attribution note found for %s (%s)\n", shortSHA, sha)
	}
	return nil
}
