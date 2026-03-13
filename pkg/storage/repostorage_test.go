package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestRepo(t *testing.T) (string, string) {
	t.Helper()
	tmp := t.TempDir()
	gitDir := filepath.Join(tmp, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	return tmp, gitDir
}

func TestNewCreatesDirectories(t *testing.T) {
	tmp, gitDir := setupTestRepo(t)
	s, err := NewRepoStorage(gitDir, tmp)
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{s.AiDir, s.WorkingLogs, s.Snapshots} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", dir)
		}
	}
	if s.AiDir != filepath.Join(gitDir, "ai-attr") {
		t.Errorf("unexpected AiDir: %s", s.AiDir)
	}
}

func TestWorkingLogDir(t *testing.T) {
	tmp, gitDir := setupTestRepo(t)
	s, _ := NewRepoStorage(gitDir, tmp)

	if s.HasWorkingLog("abc123") {
		t.Error("should not have working log yet")
	}

	dir, err := s.WorkingLogDir("abc123")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("working log dir should exist")
	}
	if !s.HasWorkingLog("abc123") {
		t.Error("should have working log now")
	}
}

func TestDeleteWorkingLog(t *testing.T) {
	tmp, gitDir := setupTestRepo(t)
	s, _ := NewRepoStorage(gitDir, tmp)

	s.WorkingLogDir("abc123")
	if !s.HasWorkingLog("abc123") {
		t.Fatal("should have working log")
	}

	if err := s.DeleteWorkingLog("abc123"); err != nil {
		t.Fatal(err)
	}
	if s.HasWorkingLog("abc123") {
		t.Error("should not have working log after delete")
	}
}

func TestDeleteNonexistentWorkingLog(t *testing.T) {
	tmp, gitDir := setupTestRepo(t)
	s, _ := NewRepoStorage(gitDir, tmp)
	if err := s.DeleteWorkingLog("nonexistent"); err != nil {
		t.Fatal(err)
	}
}

func TestToRelativePath(t *testing.T) {
	tmp, gitDir := setupTestRepo(t)
	s, _ := NewRepoStorage(gitDir, tmp)

	if r := s.ToRelativePath("src/main.rs"); r != "src/main.rs" {
		t.Errorf("expected src/main.rs, got %s", r)
	}

	abs := filepath.Join(tmp, "src/main.rs")
	if r := s.ToRelativePath(abs); r != "src/main.rs" {
		t.Errorf("expected src/main.rs, got %s", r)
	}
}

func TestFindGitDir(t *testing.T) {
	tmp := t.TempDir()
	gitDir := filepath.Join(tmp, ".git")
	os.MkdirAll(gitDir, 0755)

	foundGit, foundWork, err := FindGitDir(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if foundGit != gitDir {
		t.Errorf("expected %s, got %s", gitDir, foundGit)
	}
	if foundWork != tmp {
		t.Errorf("expected %s, got %s", tmp, foundWork)
	}

	// From subdir
	sub := filepath.Join(tmp, "src", "deep")
	os.MkdirAll(sub, 0755)
	foundGit, foundWork, err = FindGitDir(sub)
	if err != nil {
		t.Fatal(err)
	}
	if foundGit != gitDir {
		t.Errorf("expected %s, got %s", gitDir, foundGit)
	}
	if foundWork != tmp {
		t.Errorf("expected %s, got %s", tmp, foundWork)
	}
}

func TestFindGitDirNotFound(t *testing.T) {
	tmp := t.TempDir()
	_, _, err := FindGitDir(tmp)
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestSaveAndLoadSnapshot(t *testing.T) {
	tmp, gitDir := setupTestRepo(t)
	s, _ := NewRepoStorage(gitDir, tmp)

	content := "fn main() {\n    println!(\"hello\");\n}\n"
	if err := s.SaveSnapshot("src/main.rs", content); err != nil {
		t.Fatal(err)
	}

	loaded, ok, err := s.LoadSnapshot("src/main.rs")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected snapshot to exist")
	}
	if loaded != content {
		t.Errorf("content mismatch")
	}
}

func TestLoadNonexistentSnapshot(t *testing.T) {
	tmp, gitDir := setupTestRepo(t)
	s, _ := NewRepoStorage(gitDir, tmp)

	_, ok, err := s.LoadSnapshot("does/not/exist.rs")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("expected no snapshot")
	}
}

func TestDeleteSnapshot(t *testing.T) {
	tmp, gitDir := setupTestRepo(t)
	s, _ := NewRepoStorage(gitDir, tmp)

	s.SaveSnapshot("src/main.rs", "content")
	_, ok, _ := s.LoadSnapshot("src/main.rs")
	if !ok {
		t.Fatal("expected snapshot to exist")
	}

	if err := s.DeleteSnapshot("src/main.rs"); err != nil {
		t.Fatal(err)
	}
	_, ok, _ = s.LoadSnapshot("src/main.rs")
	if ok {
		t.Error("expected snapshot to be deleted")
	}
}

func TestListWorkingLogDirs(t *testing.T) {
	tmp, gitDir := setupTestRepo(t)
	s, _ := NewRepoStorage(gitDir, tmp)

	s.WorkingLogDir("sha-aaa")
	s.WorkingLogDir("sha-bbb")
	s.WorkingLogDir("sha-ccc")

	dirs, err := s.ListWorkingLogDirs()
	if err != nil {
		t.Fatal(err)
	}
	if len(dirs) != 3 {
		t.Fatalf("expected 3 dirs, got %d", len(dirs))
	}
	if dirs[0][0] != "sha-aaa" {
		t.Errorf("expected sha-aaa, got %s", dirs[0][0])
	}
	if dirs[1][0] != "sha-bbb" {
		t.Errorf("expected sha-bbb, got %s", dirs[1][0])
	}
	if dirs[2][0] != "sha-ccc" {
		t.Errorf("expected sha-ccc, got %s", dirs[2][0])
	}
}

func TestListWorkingLogDirsEmpty(t *testing.T) {
	tmp, gitDir := setupTestRepo(t)
	s, _ := NewRepoStorage(gitDir, tmp)

	dirs, err := s.ListWorkingLogDirs()
	if err != nil {
		t.Fatal(err)
	}
	if len(dirs) != 0 {
		t.Errorf("expected empty, got %d", len(dirs))
	}
}
