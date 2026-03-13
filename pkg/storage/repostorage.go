package storage

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	aierr "github.com/apshoemaker/ai-attr/pkg/errors"
)

// RepoStorage manages the .git/ai-attr/ directory structure.
type RepoStorage struct {
	AiDir         string
	RepoWorkdir   string
	WorkingLogs   string
	Snapshots     string
	PostSnapshots string
}

// NewRepoStorage creates a RepoStorage for the given git directory and work tree.
func NewRepoStorage(gitDir, repoWorkdir string) (*RepoStorage, error) {
	aiDir := filepath.Join(gitDir, "ai-attr")
	s := &RepoStorage{
		AiDir:         aiDir,
		RepoWorkdir:   repoWorkdir,
		WorkingLogs:   filepath.Join(aiDir, "working_logs"),
		Snapshots:     filepath.Join(aiDir, "snapshots"),
		PostSnapshots: filepath.Join(aiDir, "post_snapshots"),
	}
	if err := s.ensureDirectories(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *RepoStorage) ensureDirectories() error {
	for _, dir := range []string{s.AiDir, s.WorkingLogs, s.Snapshots, s.PostSnapshots} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return aierr.NewIO(err)
		}
	}
	return nil
}

// HasWorkingLog checks if a working log exists for the given commit.
func (s *RepoStorage) HasWorkingLog(sha string) bool {
	_, err := os.Stat(filepath.Join(s.WorkingLogs, sha))
	return err == nil
}

// WorkingLogDir gets or creates the working log directory for a base commit.
func (s *RepoStorage) WorkingLogDir(sha string) (string, error) {
	dir := filepath.Join(s.WorkingLogs, sha)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", aierr.NewIO(err)
	}
	return dir, nil
}

// DeleteWorkingLog deletes the working log for a base commit.
func (s *RepoStorage) DeleteWorkingLog(sha string) error {
	dir := filepath.Join(s.WorkingLogs, sha)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(dir)
}

// ListWorkingLogDirs lists all working log directories, returning (sha, path) pairs.
func (s *RepoStorage) ListWorkingLogDirs() ([][2]string, error) {
	if _, err := os.Stat(s.WorkingLogs); os.IsNotExist(err) {
		return nil, nil
	}
	entries, err := os.ReadDir(s.WorkingLogs)
	if err != nil {
		return nil, aierr.NewIO(err)
	}
	var dirs [][2]string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, [2]string{e.Name(), filepath.Join(s.WorkingLogs, e.Name())})
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i][0] < dirs[j][0] })
	return dirs, nil
}

// SaveSnapshot saves a file snapshot (pre-edit content) for diffing later.
func (s *RepoStorage) SaveSnapshot(repoRelativePath, content string) error {
	key := SnapshotKey(repoRelativePath)
	return os.WriteFile(filepath.Join(s.Snapshots, key), []byte(content), 0644)
}

// LoadSnapshot loads a previously saved snapshot.
func (s *RepoStorage) LoadSnapshot(repoRelativePath string) (string, bool, error) {
	key := SnapshotKey(repoRelativePath)
	data, err := os.ReadFile(filepath.Join(s.Snapshots, key))
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, aierr.NewIO(err)
	}
	return string(data), true, nil
}

// DeleteSnapshot deletes a snapshot.
func (s *RepoStorage) DeleteSnapshot(repoRelativePath string) error {
	key := SnapshotKey(repoRelativePath)
	path := filepath.Join(s.Snapshots, key)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}

// SavePostSnapshot saves a post-edit snapshot.
func (s *RepoStorage) SavePostSnapshot(repoRelativePath, content string) error {
	key := SnapshotKey(repoRelativePath)
	return os.WriteFile(filepath.Join(s.PostSnapshots, key), []byte(content), 0644)
}

// LoadPostSnapshot loads a post-edit snapshot.
func (s *RepoStorage) LoadPostSnapshot(repoRelativePath string) (string, bool, error) {
	key := SnapshotKey(repoRelativePath)
	data, err := os.ReadFile(filepath.Join(s.PostSnapshots, key))
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, aierr.NewIO(err)
	}
	return string(data), true, nil
}

// ClearAllPostSnapshots clears all post-edit snapshots.
func (s *RepoStorage) ClearAllPostSnapshots() error {
	entries, err := os.ReadDir(s.PostSnapshots)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return aierr.NewIO(err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			if err := os.Remove(filepath.Join(s.PostSnapshots, e.Name())); err != nil {
				return aierr.NewIO(err)
			}
		}
	}
	return nil
}

// SaveSessionModel saves the model name for a session.
func (s *RepoStorage) SaveSessionModel(sessionID, model string) error {
	path := filepath.Join(s.AiDir, fmt.Sprintf("session-model-%s", sessionID))
	return os.WriteFile(path, []byte(model), 0644)
}

// LoadSessionModel loads the cached model name for a session.
func (s *RepoStorage) LoadSessionModel(sessionID string) (string, bool, error) {
	path := filepath.Join(s.AiDir, fmt.Sprintf("session-model-%s", sessionID))
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, aierr.NewIO(err)
	}
	return strings.TrimSpace(string(data)), true, nil
}

// ToRelativePath converts a file path to repo-relative.
func (s *RepoStorage) ToRelativePath(filePath string) string {
	if !filepath.IsAbs(filePath) {
		return filePath
	}
	rel, err := filepath.Rel(s.RepoWorkdir, filePath)
	if err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	// Try with canonicalized paths
	absWorkdir, err1 := filepath.EvalSymlinks(s.RepoWorkdir)
	absPath, err2 := filepath.EvalSymlinks(filePath)
	if err1 == nil && err2 == nil {
		rel, err := filepath.Rel(absWorkdir, absPath)
		if err == nil && !strings.HasPrefix(rel, "..") {
			return rel
		}
	}
	return filePath
}

// SnapshotKey returns a SHA256 hash of the repo-relative path, used as snapshot filename.
func SnapshotKey(repoRelativePath string) string {
	h := sha256.Sum256([]byte(repoRelativePath))
	return fmt.Sprintf("%x", h[:])
}

// FindGitDir finds the git directory by walking up from the given path.
func FindGitDir(start string) (gitDir, workDir string, err error) {
	current := start
	if !filepath.IsAbs(current) {
		cwd, e := os.Getwd()
		if e != nil {
			return "", "", aierr.NewIO(e)
		}
		current = filepath.Join(cwd, current)
	}

	for {
		gitPath := filepath.Join(current, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return gitPath, current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", "", aierr.NewNotAGitRepo()
		}
		current = parent
	}
}
