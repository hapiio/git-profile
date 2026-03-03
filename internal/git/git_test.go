package git_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/hapiio/git-profile/internal/git"
)

// initRepo creates a bare-minimum git repo in a temp dir and returns its path.
func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	run("init")
	run("config", "user.name", "Test User")
	run("config", "user.email", "test@test.com")
	return dir
}

// chdir changes the working directory for the duration of the test.
func chdir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir(%q): %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(old); err != nil {
			t.Errorf("Chdir back to %q: %v", old, err)
		}
	})
}

// ---- IsRepo ----

func TestIsRepo_InsideRepo(t *testing.T) {
	dir := initRepo(t)
	chdir(t, dir)
	if !git.IsRepo() {
		t.Error("IsRepo() = false inside a git repo, want true")
	}
}

func TestIsRepo_OutsideRepo(t *testing.T) {
	chdir(t, t.TempDir()) // plain dir, not a git repo
	if git.IsRepo() {
		t.Error("IsRepo() = true outside a git repo, want false")
	}
}

// ---- Dir ----

func TestDir_ReturnsNonEmpty(t *testing.T) {
	dir := initRepo(t)
	chdir(t, dir)

	got, err := git.Dir()
	if err != nil {
		t.Fatalf("Dir() error: %v", err)
	}
	if got == "" {
		t.Error("Dir() returned empty string")
	}
}

func TestDir_OutsideRepo_ReturnsError(t *testing.T) {
	chdir(t, t.TempDir())
	if _, err := git.Dir(); err == nil {
		t.Error("Dir() outside a repo should return an error, got nil")
	}
}

// ---- SetConfig / GetConfig ----

func TestSetAndGetConfig_Local(t *testing.T) {
	dir := initRepo(t)
	chdir(t, dir)

	const key, val = "test.mykey", "hello-world"
	if err := git.SetConfig("local", key, val); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}

	got, err := git.GetConfig(key)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if got != val {
		t.Errorf("GetConfig = %q, want %q", got, val)
	}
}

func TestGetConfig_KeyAbsent_ReturnsError(t *testing.T) {
	dir := initRepo(t)
	chdir(t, dir)

	_, err := git.GetConfig("nonexistent.key.xyz")
	if err == nil {
		t.Error("GetConfig on absent key should return an error, got nil")
	}
}

func TestSetConfig_MultipleKeys(t *testing.T) {
	dir := initRepo(t)
	chdir(t, dir)

	pairs := [][2]string{
		{"user.name", "Override Name"},
		{"user.email", "override@example.com"},
	}
	for _, kv := range pairs {
		if err := git.SetConfig("local", kv[0], kv[1]); err != nil {
			t.Fatalf("SetConfig(%q): %v", kv[0], err)
		}
	}
	for _, kv := range pairs {
		got, err := git.GetConfig(kv[0])
		if err != nil {
			t.Fatalf("GetConfig(%q): %v", kv[0], err)
		}
		if got != kv[1] {
			t.Errorf("GetConfig(%q) = %q, want %q", kv[0], got, kv[1])
		}
	}
}

// ---- UnsetConfig ----

func TestUnsetConfig_KeyPresent(t *testing.T) {
	dir := initRepo(t)
	chdir(t, dir)

	const key = "test.unsetme"
	if err := git.SetConfig("local", key, "value"); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	if err := git.UnsetConfig("local", key); err != nil {
		t.Fatalf("UnsetConfig: %v", err)
	}
	if _, err := git.GetConfig(key); err == nil {
		t.Error("GetConfig after UnsetConfig should return error, got nil")
	}
}

func TestUnsetConfig_KeyAbsent_NoError(t *testing.T) {
	dir := initRepo(t)
	chdir(t, dir)

	// Unsetting a key that was never set must not return an error.
	if err := git.UnsetConfig("local", "never.set.key"); err != nil {
		t.Errorf("UnsetConfig on absent key: %v (want nil)", err)
	}
}

// ---- GetGlobalConfig ----

func TestGetGlobalConfig_KeyAbsent_ReturnsError(t *testing.T) {
	// Reading a key that almost certainly does not exist in the real global config.
	_, err := git.GetGlobalConfig("gitprofile.veryrandom.key.xyz.test")
	if err == nil {
		t.Error("GetGlobalConfig on absent key should return error, got nil")
	}
}
