package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("expected non-nil default config")
	}
	if cfg.MaxDepth <= 0 {
		t.Errorf("expected MaxDepth > 0, got %d", cfg.MaxDepth)
	}
	if len(cfg.IgnoreDirs) == 0 {
		t.Error("expected non-empty IgnoreDirs")
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine user home dir")
	}

	p := ExpandPath("~/test")
	expected := filepath.Join(home, "test")
	if p != expected {
		t.Errorf("expected %s, got %s", expected, p)
	}

	p2 := ExpandPath("/absolute/path")
	if p2 != "/absolute/path" {
		t.Errorf("expected untouched absolute path, got %s", p2)
	}
}
