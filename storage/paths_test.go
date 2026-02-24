package storage

import (
	"os"
	"strings"
	"testing"
)

func TestResolveDataDirOverride(t *testing.T) {
	dir, err := ResolveDataDir("/custom/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir != "/custom/path" {
		t.Errorf("expected /custom/path, got %s", dir)
	}
}

func TestResolveDataDirFromEnv(t *testing.T) {
	original := os.Getenv(EnvDataDir)
	defer os.Setenv(EnvDataDir, original)

	if err := os.Setenv(EnvDataDir, "/env/path"); err != nil {
		t.Fatalf("set env: %v", err)
	}

	dir, err := ResolveDataDir("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir != "/env/path" {
		t.Errorf("expected /env/path, got %s", dir)
	}
}

func TestResolveDataDirDefault(t *testing.T) {
	original := os.Getenv(EnvDataDir)
	defer os.Setenv(EnvDataDir, original)

	// 環境変数をクリア
	os.Unsetenv(EnvDataDir)

	dir, err := ResolveDataDir("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// デフォルトパスにAppDirNameが含まれることを確認
	if !strings.HasSuffix(dir, AppDirName) {
		t.Errorf("expected path to end with %s, got %s", AppDirName, dir)
	}
}

func TestCardsPath(t *testing.T) {
	path := CardsPath("/data/dir")
	expected := "/data/dir/cards.json"
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestStatePath(t *testing.T) {
	path := StatePath("/data/dir")
	expected := "/data/dir/state.json"
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}
