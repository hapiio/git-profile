package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hapiio/git-profile/internal/config"
)

func newTestManager(t *testing.T) *config.Manager {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	m, err := config.NewManager(path)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return m
}

func TestLoadEmpty(t *testing.T) {
	m := newTestManager(t)
	cfg, err := m.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(cfg.Profiles))
	}
	if cfg.Version != 1 {
		t.Errorf("expected version 1, got %d", cfg.Version)
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	m := newTestManager(t)

	want := &config.Config{
		Version: 1,
		Profiles: map[string]config.Profile{
			"work": {
				ID:       "work",
				GitUser:  "Jane Work",
				GitEmail: "jane@company.com",
			},
			"personal": {
				ID:         "personal",
				GitUser:    "Jane Doe",
				GitEmail:   "jane@example.com",
				SSHKeyPath: "/home/jane/.ssh/id_personal",
			},
		},
	}

	if err := m.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := m.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(got.Profiles) != len(want.Profiles) {
		t.Fatalf("got %d profiles, want %d", len(got.Profiles), len(want.Profiles))
	}

	for id, wp := range want.Profiles {
		gp, ok := got.Profiles[id]
		if !ok {
			t.Errorf("profile %q not found after round-trip", id)
			continue
		}
		if gp.GitUser != wp.GitUser {
			t.Errorf("profile %q: GitUser = %q, want %q", id, gp.GitUser, wp.GitUser)
		}
		if gp.GitEmail != wp.GitEmail {
			t.Errorf("profile %q: GitEmail = %q, want %q", id, gp.GitEmail, wp.GitEmail)
		}
		if gp.SSHKeyPath != wp.SSHKeyPath {
			t.Errorf("profile %q: SSHKeyPath = %q, want %q", id, gp.SSHKeyPath, wp.SSHKeyPath)
		}
	}
}

func TestSortedIDs(t *testing.T) {
	cfg := &config.Config{
		Version: 1,
		Profiles: map[string]config.Profile{
			"zebra":    {ID: "zebra"},
			"alpha":    {ID: "alpha"},
			"work":     {ID: "work"},
			"personal": {ID: "personal"},
		},
	}

	ids := cfg.SortedIDs()
	want := []string{"alpha", "personal", "work", "zebra"}
	for i, id := range ids {
		if id != want[i] {
			t.Errorf("SortedIDs[%d] = %q, want %q", i, id, want[i])
		}
	}
}

func TestSaveAtomic(t *testing.T) {
	m := newTestManager(t)

	cfg := &config.Config{
		Version:  1,
		Profiles: map[string]config.Profile{"test": {ID: "test", GitUser: "T", GitEmail: "t@t.com"}},
	}

	if err := m.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// tmp file must not remain after successful save.
	if _, err := os.Stat(m.Path() + ".tmp"); !os.IsNotExist(err) {
		t.Error("tmp file still exists after Save")
	}

	// Config file must exist with restricted permissions.
	info, err := os.Stat(m.Path())
	if err != nil {
		t.Fatalf("Stat config: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("config permissions = %o, want 0600", perm)
	}
}

func TestLoadMalformed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	if err := os.WriteFile(path, []byte("{bad json}"), 0o600); err != nil {
		t.Fatal(err)
	}

	m, _ := config.NewManager(path)
	_, err := m.Load()
	if err == nil {
		t.Fatal("expected error loading malformed JSON, got nil")
	}
}

func TestMigrate_BackfillsProfileIDs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	// write a config where profile IDs are missing (old format).
	raw := `{"version":0,"profiles":{"work":{"git_user":"Jane","git_email":"j@c.com"}}}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}

	m, _ := config.NewManager(path)
	cfg, err := m.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Version != 1 {
		t.Errorf("version after migration = %d, want 1", cfg.Version)
	}
	if cfg.Profiles["work"].ID != "work" {
		t.Errorf("profile ID not backfilled: got %q, want %q", cfg.Profiles["work"].ID, "work")
	}
}

func TestProfile_AllFields_RoundTrip(t *testing.T) {
	m := newTestManager(t)

	want := config.Profile{
		ID:          "gpg-profile",
		GitUser:     "Security Dev",
		GitEmail:    "sec@example.com",
		SSHKeyPath:  "/home/user/.ssh/id_ed25519",
		GPGKeyID:    "ABC123DEF456",
		SignCommits: true,
	}

	cfg := &config.Config{
		Version:  1,
		Profiles: map[string]config.Profile{want.ID: want},
	}
	if err := m.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := m.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	got, ok := loaded.Profiles[want.ID]
	if !ok {
		t.Fatalf("profile %q not found after round-trip", want.ID)
	}
	if got.GPGKeyID != want.GPGKeyID {
		t.Errorf("GPGKeyID = %q, want %q", got.GPGKeyID, want.GPGKeyID)
	}
	if got.SignCommits != want.SignCommits {
		t.Errorf("SignCommits = %v, want %v", got.SignCommits, want.SignCommits)
	}
	if got.SSHKeyPath != want.SSHKeyPath {
		t.Errorf("SSHKeyPath = %q, want %q", got.SSHKeyPath, want.SSHKeyPath)
	}
}

func TestSortedIDs_Empty(t *testing.T) {
	cfg := &config.Config{Version: 1, Profiles: map[string]config.Profile{}}
	ids := cfg.SortedIDs()
	if len(ids) != 0 {
		t.Errorf("SortedIDs on empty config = %v, want []", ids)
	}
}

func TestSortedIDs_Single(t *testing.T) {
	cfg := &config.Config{
		Version:  1,
		Profiles: map[string]config.Profile{"only": {ID: "only"}},
	}
	ids := cfg.SortedIDs()
	if len(ids) != 1 || ids[0] != "only" {
		t.Errorf("SortedIDs = %v, want [only]", ids)
	}
}

func TestSave_Overwrite(t *testing.T) {
	m := newTestManager(t)

	// save initial config.
	cfg1 := &config.Config{
		Version:  1,
		Profiles: map[string]config.Profile{"a": {ID: "a", GitUser: "A", GitEmail: "a@a.com"}},
	}
	if err := m.Save(cfg1); err != nil {
		t.Fatalf("first Save: %v", err)
	}

	// overwrite with a different config.
	cfg2 := &config.Config{
		Version:  1,
		Profiles: map[string]config.Profile{"b": {ID: "b", GitUser: "B", GitEmail: "b@b.com"}},
	}
	if err := m.Save(cfg2); err != nil {
		t.Fatalf("second Save: %v", err)
	}

	got, err := m.Load()
	if err != nil {
		t.Fatalf("Load after overwrite: %v", err)
	}
	if _, ok := got.Profiles["a"]; ok {
		t.Error("old profile 'a' should not exist after overwrite")
	}
	if _, ok := got.Profiles["b"]; !ok {
		t.Error("new profile 'b' should exist after overwrite")
	}
}

func TestNewManager_CreatesDirectory(t *testing.T) {
	base := t.TempDir()
	// use a nested path that doesn't exist yet.
	path := filepath.Join(base, "nested", "deep", "config.json")
	m, err := config.NewManager(path)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	// save something so the file is created.
	cfg := &config.Config{Version: 1, Profiles: map[string]config.Profile{}}
	if err := m.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("config file should exist after Save")
	}
}

func TestManager_Path(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	m, err := config.NewManager(path)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	if m.Path() != path {
		t.Errorf("Path() = %q, want %q", m.Path(), path)
	}
}
