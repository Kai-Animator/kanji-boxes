package storage

import (
	"errors"
	"os"
	"path/filepath"
)

const (
	EnvDataDir = "KANJI_BOXES_DIR"
	AppDirName = "kanji-boxes"
)

func ResolveDataDir(override string) (string, error) {
	if override != "" {
		return override, nil
	}

	if env := os.Getenv(EnvDataDir); env != "" {
		return env, nil
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	if base == "" {
		return "", errors.New("user config directory not available")
	}

	return filepath.Join(base, AppDirName), nil
}

func CardsPath(dataDir string) string {
	return filepath.Join(dataDir, "cards.json")
}

func StatePath(dataDir string) string {
	return filepath.Join(dataDir, "state.json")
}
