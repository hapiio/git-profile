// Package cmd_test contains integration tests for git-profile commands.
// Tests use the --config flag to isolate each test's state to a temporary file
// and assert on side effects (config file contents) rather than stdout.
package cmd_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hapiio/git-profile/cmd"
	"github.com/hapiio/git-profile/internal/config"
)

// initGitRepo creates a temporary directory, runs git init inside it, and
// sets a minimal local git identity so git operations succeed.
func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Skipf("git not available: %v", err)
	}
	exec.Command("git", "-C", dir, "config", "user.name", "Test User").Run()        //nolint:errcheck
	exec.Command("git", "-C", dir, "config", "user.email", "test@example.com").Run() //nolint:errcheck
	return dir
}

// chdir changes the working directory for the test and restores it via t.Cleanup.
func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
}

// run executes git-profile with --config <cfgPath> prepended to args.
func run(t *testing.T, cfgPath string, args ...string) error {
	t.Helper()
	all := append([]string{"--config", cfgPath}, args...)
	return cmd.RunArgs(all)
}

// mustRun fails the test if the command returns an error.
func mustRun(t *testing.T, cfgPath string, args ...string) {
	t.Helper()
	if err := run(t, cfgPath, args...); err != nil {
		t.Fatalf("command %v failed: %v", args, err)
	}
}

// loadCfg loads the config at path, failing the test on error.
func loadCfg(t *testing.T, cfgPath string) *config.Config {
	t.Helper()
	m, err := config.NewManager(cfgPath)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	cfg, err := m.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return cfg
}

// tmpCfg returns a path inside t.TempDir() for an isolated config file.
func tmpCfg(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "config.json")
}

// ---- add ----

func TestAdd_Success(t *testing.T) {
	cfg := tmpCfg(t)
	mustRun(t, cfg, "add", "--id", "work", "--name", "Jane Dev", "--email", "jane@company.com")

	loaded := loadCfg(t, cfg)
	p, ok := loaded.Profiles["work"]
	if !ok {
		t.Fatal("profile 'work' not found after add")
	}
	if p.GitUser != "Jane Dev" {
		t.Errorf("GitUser = %q, want %q", p.GitUser, "Jane Dev")
	}
	if p.GitEmail != "jane@company.com" {
		t.Errorf("GitEmail = %q, want %q", p.GitEmail, "jane@company.com")
	}
	if p.ID != "work" {
		t.Errorf("ID = %q, want %q", p.ID, "work")
	}
}

func TestAdd_WithSSHKey(t *testing.T) {
	cfg := tmpCfg(t)
	mustRun(t, cfg, "add",
		"--id", "oss",
		"--name", "Dev OSS",
		"--email", "dev@oss.org",
		"--ssh-key", "~/.ssh/id_oss",
	)

	p := loadCfg(t, cfg).Profiles["oss"]
	if p.SSHKeyPath != "~/.ssh/id_oss" {
		t.Errorf("SSHKeyPath = %q, want %q", p.SSHKeyPath, "~/.ssh/id_oss")
	}
}

func TestAdd_WithGPG(t *testing.T) {
	cfg := tmpCfg(t)
	mustRun(t, cfg, "add",
		"--id", "secure",
		"--name", "Sec Dev",
		"--email", "sec@example.com",
		"--gpg-key", "ABCDEF1234",
		"--sign-commits",
	)

	p := loadCfg(t, cfg).Profiles["secure"]
	if p.GPGKeyID != "ABCDEF1234" {
		t.Errorf("GPGKeyID = %q, want %q", p.GPGKeyID, "ABCDEF1234")
	}
	if !p.SignCommits {
		t.Error("SignCommits = false, want true")
	}
}

func TestAdd_Duplicate_Errors(t *testing.T) {
	cfg := tmpCfg(t)
	mustRun(t, cfg, "add", "--id", "work", "--name", "Jane", "--email", "jane@co.com")

	err := run(t, cfg, "add", "--id", "work", "--name", "Jane2", "--email", "jane2@co.com")
	if err == nil {
		t.Fatal("adding duplicate profile should return error, got nil")
	}
}

func TestAdd_Force_Overwrites(t *testing.T) {
	cfg := tmpCfg(t)
	mustRun(t, cfg, "add", "--id", "work", "--name", "Old Name", "--email", "old@co.com")
	mustRun(t, cfg, "add", "--id", "work", "--name", "New Name", "--email", "new@co.com", "--force")

	p := loadCfg(t, cfg).Profiles["work"]
	if p.GitUser != "New Name" {
		t.Errorf("GitUser after --force = %q, want %q", p.GitUser, "New Name")
	}
	if p.GitEmail != "new@co.com" {
		t.Errorf("GitEmail after --force = %q, want %q", p.GitEmail, "new@co.com")
	}
}

// Cobra reuses the command struct between test calls and pflag does not reset
// variable values to their defaults between parses. To reliably test "missing
// required field" validation we pass the flag explicitly with an empty string,
// which forces pflag to write "" into the variable regardless of prior state.

func TestAdd_EmptyID_Errors(t *testing.T) {
	cfg := tmpCfg(t)
	// IsTTY() returns false in test context, so an empty --id must produce an error.
	err := run(t, cfg, "add", "--id", "", "--name", "Jane", "--email", "jane@co.com")
	if err == nil {
		t.Fatal("add with empty --id should error, got nil")
	}
}

func TestAdd_EmptyName_Errors(t *testing.T) {
	cfg := tmpCfg(t)
	err := run(t, cfg, "add", "--id", "work", "--name", "", "--email", "jane@co.com")
	if err == nil {
		t.Fatal("add with empty --name should error, got nil")
	}
}

func TestAdd_EmptyEmail_Errors(t *testing.T) {
	cfg := tmpCfg(t)
	err := run(t, cfg, "add", "--id", "work", "--name", "Jane", "--email", "")
	if err == nil {
		t.Fatal("add with empty --email should error, got nil")
	}
}

// ---- list ----

func TestList_Empty_NoError(t *testing.T) {
	cfg := tmpCfg(t)
	mustRun(t, cfg, "list")
}

func TestList_WithProfiles_NoError(t *testing.T) {
	cfg := tmpCfg(t)
	mustRun(t, cfg, "add", "--id", "work", "--name", "Jane", "--email", "jane@co.com")
	mustRun(t, cfg, "add", "--id", "personal", "--name", "Jane", "--email", "jane@me.com")
	mustRun(t, cfg, "list")
}

// ---- remove ----

func TestRemove_Success(t *testing.T) {
	cfg := tmpCfg(t)
	mustRun(t, cfg, "add", "--id", "work", "--name", "Jane", "--email", "jane@co.com")
	mustRun(t, cfg, "remove", "--yes", "work")

	loaded := loadCfg(t, cfg)
	if _, ok := loaded.Profiles["work"]; ok {
		t.Error("profile 'work' should not exist after remove")
	}
}

func TestRemove_Multiple(t *testing.T) {
	cfg := tmpCfg(t)
	mustRun(t, cfg, "add", "--id", "a", "--name", "A", "--email", "a@a.com")
	mustRun(t, cfg, "add", "--id", "b", "--name", "B", "--email", "b@b.com")
	mustRun(t, cfg, "add", "--id", "c", "--name", "C", "--email", "c@c.com")
	mustRun(t, cfg, "remove", "--yes", "a", "b")

	loaded := loadCfg(t, cfg)
	if _, ok := loaded.Profiles["a"]; ok {
		t.Error("profile 'a' should be removed")
	}
	if _, ok := loaded.Profiles["b"]; ok {
		t.Error("profile 'b' should be removed")
	}
	if _, ok := loaded.Profiles["c"]; !ok {
		t.Error("profile 'c' should still exist")
	}
}

func TestRemove_NotFound_Errors(t *testing.T) {
	cfg := tmpCfg(t)
	err := run(t, cfg, "remove", "--yes", "nonexistent")
	if err == nil {
		t.Fatal("remove of nonexistent profile should error, got nil")
	}
}

// ---- rename ----

func TestRename_Success(t *testing.T) {
	cfg := tmpCfg(t)
	mustRun(t, cfg, "add", "--id", "old", "--name", "Jane", "--email", "jane@co.com")
	mustRun(t, cfg, "rename", "old", "new")

	loaded := loadCfg(t, cfg)
	if _, ok := loaded.Profiles["old"]; ok {
		t.Error("profile 'old' should not exist after rename")
	}
	p, ok := loaded.Profiles["new"]
	if !ok {
		t.Fatal("profile 'new' should exist after rename")
	}
	if p.ID != "new" {
		t.Errorf("ID field = %q, want %q", p.ID, "new")
	}
	if p.GitUser != "Jane" {
		t.Errorf("GitUser = %q, want %q", p.GitUser, "Jane")
	}
}

func TestRename_PreservesFields(t *testing.T) {
	cfg := tmpCfg(t)
	mustRun(t, cfg, "add",
		"--id", "src",
		"--name", "Dev",
		"--email", "dev@example.com",
		"--ssh-key", "~/.ssh/id_dev",
		"--gpg-key", "GPGKEY123",
		"--sign-commits",
	)
	mustRun(t, cfg, "rename", "src", "dst")

	p := loadCfg(t, cfg).Profiles["dst"]
	if p.SSHKeyPath != "~/.ssh/id_dev" {
		t.Errorf("SSHKeyPath after rename = %q, want %q", p.SSHKeyPath, "~/.ssh/id_dev")
	}
	if p.GPGKeyID != "GPGKEY123" {
		t.Errorf("GPGKeyID after rename = %q, want %q", p.GPGKeyID, "GPGKEY123")
	}
	if !p.SignCommits {
		t.Error("SignCommits should be preserved after rename")
	}
}

func TestRename_SourceNotFound_Errors(t *testing.T) {
	cfg := tmpCfg(t)
	err := run(t, cfg, "rename", "ghost", "newname")
	if err == nil {
		t.Fatal("rename of nonexistent source should error, got nil")
	}
}

func TestRename_TargetExists_Errors(t *testing.T) {
	cfg := tmpCfg(t)
	mustRun(t, cfg, "add", "--id", "a", "--name", "A", "--email", "a@a.com")
	mustRun(t, cfg, "add", "--id", "b", "--name", "B", "--email", "b@b.com")

	err := run(t, cfg, "rename", "a", "b")
	if err == nil {
		t.Fatal("rename to existing target should error, got nil")
	}
}

// ---- version ----

func TestVersion_NoError(t *testing.T) {
	cfg := tmpCfg(t)
	mustRun(t, cfg, "version")
}

// ---- current ----

func TestCurrent_NotInRepo_Errors(t *testing.T) {
	cfg := tmpCfg(t)
	chdir(t, t.TempDir()) // plain directory, not a git repo
	err := run(t, cfg, "current")
	if err == nil {
		t.Fatal("current outside git repo should error, got nil")
	}
}

func TestCurrent_InRepo_NoError(t *testing.T) {
	cfg := tmpCfg(t)
	dir := initGitRepo(t)
	chdir(t, dir)
	mustRun(t, cfg, "current")
}

func TestCurrent_InRepo_WithMatchedProfile_NoError(t *testing.T) {
	cfg := tmpCfg(t)
	dir := initGitRepo(t)
	// Add a profile that matches the git identity set by initGitRepo.
	mustRun(t, cfg, "add", "--id", "test", "--name", "Test User", "--email", "test@example.com")
	chdir(t, dir)
	mustRun(t, cfg, "current")
}

// ---- use ----

func TestUse_NotFound_Errors(t *testing.T) {
	cfg := tmpCfg(t)
	err := run(t, cfg, "use", "ghost")
	if err == nil {
		t.Fatal("use of nonexistent profile should error, got nil")
	}
}

func TestUse_NotInRepo_Errors(t *testing.T) {
	cfg := tmpCfg(t)
	chdir(t, t.TempDir())
	mustRun(t, cfg, "add", "--id", "work", "--name", "Jane", "--email", "jane@co.com")
	err := run(t, cfg, "use", "work")
	if err == nil {
		t.Fatal("use without --global outside git repo should error, got nil")
	}
}

func TestUse_InRepo_Success(t *testing.T) {
	cfg := tmpCfg(t)
	dir := initGitRepo(t)
	// --gpg-key "" resets the pflag variable so GPG settings from earlier tests
	// (cobra reuses the command struct and pflag does not reset between Execute calls)
	// don't leak into this profile and get applied to a git repo.
	mustRun(t, cfg, "add", "--id", "work", "--name", "Jane Dev", "--email", "jane@co.com", "--gpg-key", "")
	chdir(t, dir)
	mustRun(t, cfg, "use", "work")
}

func TestUse_InRepo_WithSSHKey_Success(t *testing.T) {
	cfg := tmpCfg(t)
	dir := initGitRepo(t)
	mustRun(t, cfg, "add", "--id", "oss", "--name", "Dev", "--email", "dev@oss.org", "--ssh-key", "~/.ssh/id_oss", "--gpg-key", "")
	chdir(t, dir)
	mustRun(t, cfg, "use", "oss")
}

// ---- set-default ----

func TestSetDefault_NotFound_Errors(t *testing.T) {
	cfg := tmpCfg(t)
	err := run(t, cfg, "set-default", "ghost")
	if err == nil {
		t.Fatal("set-default of nonexistent profile should error, got nil")
	}
}

func TestSetDefault_NotInRepo_Errors(t *testing.T) {
	cfg := tmpCfg(t)
	chdir(t, t.TempDir())
	mustRun(t, cfg, "add", "--id", "work", "--name", "Jane", "--email", "jane@co.com")
	err := run(t, cfg, "set-default", "work")
	if err == nil {
		t.Fatal("set-default without --global outside git repo should error, got nil")
	}
}

func TestSetDefault_InRepo_Success(t *testing.T) {
	cfg := tmpCfg(t)
	dir := initGitRepo(t)
	mustRun(t, cfg, "add", "--id", "work", "--name", "Jane", "--email", "jane@co.com", "--gpg-key", "")
	chdir(t, dir)
	mustRun(t, cfg, "set-default", "work")
}

// ---- ensure ----

func TestEnsure_NoProfiles_Errors(t *testing.T) {
	cfg := tmpCfg(t)
	// empty config — no profiles at all
	err := run(t, cfg, "ensure")
	if err == nil {
		t.Fatal("ensure with no profiles should error, got nil")
	}
}

func TestEnsure_NoDefault_NoTTY_Errors(t *testing.T) {
	cfg := tmpCfg(t)
	// Use a unique ID that won't match any real gitprofile.default in the user's
	// git config, so both local and global default lookups fall through.
	mustRun(t, cfg, "add", "--id", "norealdeffound", "--name", "Jane", "--email", "jane@co.com")
	// Chdir to a plain dir (not a git repo) so git.GetConfig returns an error
	// and cannot pick up a local gitprofile.default from the project repo.
	chdir(t, t.TempDir())
	err := run(t, cfg, "ensure")
	if err == nil {
		t.Fatal("ensure with no matching default and no TTY should error, got nil")
	}
}

func TestEnsure_WithLocalDefault_InRepo_Success(t *testing.T) {
	cfg := tmpCfg(t)
	dir := initGitRepo(t)
	mustRun(t, cfg, "add", "--id", "work", "--name", "Jane", "--email", "jane@co.com", "--gpg-key", "")
	chdir(t, dir)
	mustRun(t, cfg, "set-default", "work")
	mustRun(t, cfg, "ensure")
}

// ---- install-hooks ----

func TestInstallHooks_NotInRepo_Errors(t *testing.T) {
	cfg := tmpCfg(t)
	chdir(t, t.TempDir())
	err := run(t, cfg, "install-hooks")
	if err == nil {
		t.Fatal("install-hooks outside git repo should error, got nil")
	}
}

func TestInstallHooks_InRepo_CreatesHooks(t *testing.T) {
	cfg := tmpCfg(t)
	dir := initGitRepo(t)
	chdir(t, dir)
	mustRun(t, cfg, "install-hooks")

	for _, hookName := range []string{"prepare-commit-msg", "pre-push"} {
		hookPath := filepath.Join(dir, ".git", "hooks", hookName)
		if _, err := os.Stat(hookPath); os.IsNotExist(err) {
			t.Errorf("hook %q was not created at %s", hookName, hookPath)
		}
	}
}

func TestInstallHooks_Idempotent(t *testing.T) {
	cfg := tmpCfg(t)
	dir := initGitRepo(t)
	chdir(t, dir)
	// Running twice should not error (second run updates existing hooks).
	mustRun(t, cfg, "install-hooks")
	mustRun(t, cfg, "install-hooks")
}

// ---- import ----

func TestImport_InRepo_WithID_Success(t *testing.T) {
	cfg := tmpCfg(t)
	dir := initGitRepo(t)
	chdir(t, dir)
	mustRun(t, cfg, "import", "--id", "imported")

	p := loadCfg(t, cfg).Profiles["imported"]
	if p.GitUser != "Test User" {
		t.Errorf("GitUser = %q, want %q", p.GitUser, "Test User")
	}
	if p.GitEmail != "test@example.com" {
		t.Errorf("GitEmail = %q, want %q", p.GitEmail, "test@example.com")
	}
}

func TestImport_NoID_NonTTY_Errors(t *testing.T) {
	cfg := tmpCfg(t)
	dir := initGitRepo(t)
	chdir(t, dir)
	// Cobra reuses the command struct between RunArgs calls and pflag does not
	// reset flag variables to their defaults between executions. Pass --id ""
	// explicitly so the id variable is written as empty regardless of prior state.
	err := run(t, cfg, "import", "--id", "")
	if err == nil {
		t.Fatal("import with empty --id in non-TTY should error, got nil")
	}
}

func TestImport_DuplicateID_Errors(t *testing.T) {
	cfg := tmpCfg(t)
	dir := initGitRepo(t)
	chdir(t, dir)
	mustRun(t, cfg, "import", "--id", "imported")
	err := run(t, cfg, "import", "--id", "imported")
	if err == nil {
		t.Fatal("import with duplicate profile ID should error, got nil")
	}
}

func TestImport_NotInRepo_Errors(t *testing.T) {
	cfg := tmpCfg(t)
	chdir(t, t.TempDir())
	err := run(t, cfg, "import", "--id", "x")
	if err == nil {
		t.Fatal("import outside git repo (without --global) should error, got nil")
	}
}
