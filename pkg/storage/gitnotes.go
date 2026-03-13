package storage

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/apshoemaker/ai-attr/pkg/core"
	aierr "github.com/apshoemaker/ai-attr/pkg/errors"
)

const AIAttributionRef = "ai-attribution"

// ExecGit executes a git command in a specific directory and returns its stdout.
func ExecGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			return "", aierr.NewGit(fmt.Sprintf("git %s failed: %s", args[0], stderr))
		}
		return "", aierr.NewGit(fmt.Sprintf("failed to execute git: %v", err))
	}
	return string(out), nil
}

// ExecGitStdin executes a git command with stdin input.
func ExecGitStdin(dir string, stdinData []byte, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(string(stdinData))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", aierr.NewGit(fmt.Sprintf("git %s failed: %s", args[0], string(out)))
	}
	return string(out), nil
}

// NotesAdd adds an authorship note to a commit, overwriting any existing note.
func NotesAdd(repoDir, commitSHA, noteContent string) error {
	refArg := fmt.Sprintf("--ref=%s", AIAttributionRef)
	_, err := ExecGitStdin(repoDir, []byte(noteContent), "notes", refArg, "add", "-f", "-F", "-", commitSHA)
	return err
}

// NotesShow reads the authorship note for a commit.
func NotesShow(repoDir, commitSHA string) (string, bool, error) {
	refArg := fmt.Sprintf("--ref=%s", AIAttributionRef)
	content, err := ExecGit(repoDir, "notes", refArg, "show", commitSHA)
	if err != nil {
		if aie, ok := err.(*aierr.AiAttrError); ok && strings.Contains(aie.Message, "no note found") {
			return "", false, nil
		}
		return "", false, err
	}
	return content, true, nil
}

// GetAuthorship reads and parses the authorship log for a commit.
func GetAuthorship(repoDir, commitSHA string) (*core.AuthorshipLog, bool, error) {
	content, ok, err := NotesShow(repoDir, commitSHA)
	if err != nil || !ok {
		return nil, false, err
	}
	log, err := core.DeserializeFromString(content)
	if err != nil {
		return nil, false, err
	}
	return log, true, nil
}

// NotesRemove removes the authorship note for a commit.
func NotesRemove(repoDir, commitSHA string) error {
	refArg := fmt.Sprintf("--ref=%s", AIAttributionRef)
	_, _ = ExecGit(repoDir, "notes", refArg, "remove", commitSHA)
	return nil
}

// GitShowFile gets file content at a specific commit.
func GitShowFile(repoDir, commit, filePath string) (string, bool, error) {
	spec := fmt.Sprintf("%s:%s", commit, filePath)
	content, err := ExecGit(repoDir, "show", spec)
	if err != nil {
		if aie, ok := err.(*aierr.AiAttrError); ok {
			msg := aie.Message
			if strings.Contains(msg, "does not exist") || strings.Contains(msg, "fatal") {
				return "", false, nil
			}
		}
		return "", false, err
	}
	return content, true, nil
}

// GitBlamePorcelain runs git blame --porcelain on a file.
func GitBlamePorcelain(repoDir, file string) (string, error) {
	return ExecGit(repoDir, "blame", "--porcelain", file)
}

// GetHeadSHA gets the current HEAD commit SHA.
func GetHeadSHA(repoDir string) (string, error) {
	out, err := ExecGit(repoDir, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// GetUserName gets the current git user name.
func GetUserName(repoDir string) (string, error) {
	out, err := ExecGit(repoDir, "config", "user.name")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// GetUserEmail gets the current git user email.
func GetUserEmail(repoDir string) (string, error) {
	out, err := ExecGit(repoDir, "config", "user.email")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// FindRepoRoot finds the repo root from a path.
func FindRepoRoot(start string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = start
	out, err := cmd.Output()
	if err != nil {
		return "", aierr.NewNotAGitRepo()
	}
	return strings.TrimSpace(string(out)), nil
}
