package ui_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/hapiio/git-profile/internal/ui"
)

// IsTTY must return false in the test environment because stdin is a pipe,
// not a terminal. This guards against accidental blocking reads in CI.
func TestIsTTY_InTestContext(t *testing.T) {
	if ui.IsTTY() {
		t.Error("IsTTY() = true in test context; stdin should be a pipe, not a terminal")
	}
}

func TestStyles_NotNil(t *testing.T) {
	styles := []struct {
		name  string
		value string
	}{
		{"Primary", ui.Primary.Render("test")},
		{"Success", ui.Success.Render("test")},
		{"Warning", ui.Warning.Render("test")},
		{"Danger", ui.Danger.Render("test")},
		{"Muted", ui.Muted.Render("test")},
		{"Bold", ui.Bold.Render("test")},
		{"ActiveMarker", ui.ActiveMarker},
	}

	for _, s := range styles {
		if s.value == "" {
			t.Errorf("Style %q rendered empty string", s.name)
		}
	}
}

// captureStdout replaces os.Stdout with a pipe for the duration of fn and
// returns whatever was written to it.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = orig
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

// captureStderr replaces os.Stderr with a pipe for the duration of fn and
// returns whatever was written to it.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stderr
	os.Stderr = w
	fn()
	w.Close()
	os.Stderr = orig
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestSuccessf_ContainsMessage(t *testing.T) {
	out := captureStdout(t, func() { ui.Successf("hello %s", "world") })
	if !strings.Contains(out, "hello world") {
		t.Errorf("Successf output %q does not contain %q", out, "hello world")
	}
}

func TestInfof_ContainsMessage(t *testing.T) {
	out := captureStdout(t, func() { ui.Infof("loading %d items", 3) })
	if !strings.Contains(out, "loading 3 items") {
		t.Errorf("Infof output %q does not contain %q", out, "loading 3 items")
	}
}

func TestWarningf_ContainsMessage(t *testing.T) {
	out := captureStdout(t, func() { ui.Warningf("deprecated %s", "flag") })
	if !strings.Contains(out, "deprecated flag") {
		t.Errorf("Warningf output %q does not contain %q", out, "deprecated flag")
	}
}

func TestErrorf_WritesToStderr(t *testing.T) {
	out := captureStderr(t, func() { ui.Errorf("something went %s", "wrong") })
	if !strings.Contains(out, "something went wrong") {
		t.Errorf("Errorf stderr %q does not contain %q", out, "something went wrong")
	}
}

func TestSuccessf_NotEmpty(t *testing.T) {
	out := captureStdout(t, func() { ui.Successf("ok") })
	if strings.TrimSpace(out) == "" {
		t.Error("Successf produced empty output")
	}
}
