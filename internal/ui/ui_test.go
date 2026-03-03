package ui_test

import (
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
	// Smoke-test that all exported styles render without panicking.
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
