// Package cmd internal tests cover unexported helper functions that are not
// reachable from the external cmd_test package.
package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestExpandHome_WithTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir:", err)
	}
	got := expandHome("~/foo/bar")
	want := home + "/foo/bar"
	if got != want {
		t.Errorf("expandHome(%q) = %q, want %q", "~/foo/bar", got, want)
	}
}

func TestExpandHome_AbsolutePath(t *testing.T) {
	in := "/absolute/path/key"
	if got := expandHome(in); got != in {
		t.Errorf("expandHome(%q) = %q, want unchanged", in, got)
	}
}

func TestExpandHome_EmptyString(t *testing.T) {
	if got := expandHome(""); got != "" {
		t.Errorf("expandHome(%q) = %q, want %q", "", got, "")
	}
}

func TestExpandHome_TildeOnly(t *testing.T) {
	// "~" without a slash should be returned unchanged.
	if got := expandHome("~"); got != "~" {
		t.Errorf("expandHome(%q) = %q, want unchanged", "~", got)
	}
}

// ---- extractSSHKey ----

func TestExtractSSHKey_StandardFormat(t *testing.T) {
	got := extractSSHKey("ssh -i /home/user/.ssh/id_rsa -F /dev/null")
	if got != "/home/user/.ssh/id_rsa" {
		t.Errorf("extractSSHKey = %q, want %q", got, "/home/user/.ssh/id_rsa")
	}
}

func TestExtractSSHKey_Empty(t *testing.T) {
	if got := extractSSHKey(""); got != "" {
		t.Errorf("extractSSHKey(%q) = %q, want empty", "", got)
	}
}

func TestExtractSSHKey_NoFlag(t *testing.T) {
	if got := extractSSHKey("ssh -F /dev/null"); got != "" {
		t.Errorf("extractSSHKey = %q, want empty", got)
	}
}

func TestExtractSSHKey_TildeKey(t *testing.T) {
	got := extractSSHKey("ssh -i ~/.ssh/id_ed25519 -F /dev/null")
	if got != "~/.ssh/id_ed25519" {
		t.Errorf("extractSSHKey = %q, want %q", got, "~/.ssh/id_ed25519")
	}
}

// ---- quotedList ----

func TestQuotedList_Single(t *testing.T) {
	got := quotedList([]string{"work"})
	if got != `"work"` {
		t.Errorf("quotedList = %q, want %q", got, `"work"`)
	}
}

func TestQuotedList_Multiple(t *testing.T) {
	got := quotedList([]string{"work", "personal"})
	if !strings.Contains(got, `"work"`) || !strings.Contains(got, `"personal"`) {
		t.Errorf("quotedList = %q, want both entries quoted", got)
	}
}

func TestQuotedList_Empty(t *testing.T) {
	got := quotedList([]string{})
	if got != "" {
		t.Errorf("quotedList([]) = %q, want empty string", got)
	}
}
