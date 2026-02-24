package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadJSONNonExistent(t *testing.T) {
	var data map[string]string
	err := ReadJSON("/nonexistent/path.json", &data)
	if err != nil {
		t.Fatalf("expected no error for non-existent file, got %v", err)
	}
	if data != nil {
		t.Fatalf("expected nil data for non-existent file, got %v", data)
	}
}

func TestReadJSONValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	content := `{"key": "value", "count": 42}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	var data struct {
		Key   string `json:"key"`
		Count int    `json:"count"`
	}
	if err := ReadJSON(path, &data); err != nil {
		t.Fatalf("read json: %v", err)
	}

	if data.Key != "value" {
		t.Errorf("expected key=value, got %s", data.Key)
	}
	if data.Count != 42 {
		t.Errorf("expected count=42, got %d", data.Count)
	}
}

func TestReadJSONInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.json")

	content := `{invalid json}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	var data map[string]string
	err := ReadJSON(path, &data)
	if err == nil {
		t.Fatalf("expected error for invalid json, got nil")
	}
}

func TestWriteJSONAtomicCreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output.json")

	data := map[string]string{"hello": "world"}
	if err := WriteJSONAtomic(path, data); err != nil {
		t.Fatalf("write json: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file should exist: %v", err)
	}

	var loaded map[string]string
	if err := ReadJSON(path, &loaded); err != nil {
		t.Fatalf("read json: %v", err)
	}

	if loaded["hello"] != "world" {
		t.Errorf("expected hello=world, got %s", loaded["hello"])
	}
}

func TestWriteJSONAtomicCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "nested", "output.json")

	data := map[string]int{"count": 100}
	if err := WriteJSONAtomic(path, data); err != nil {
		t.Fatalf("write json: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file should exist: %v", err)
	}
}

func TestWriteJSONAtomicOverwrites(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "overwrite.json")

	// 最初のデータ
	data1 := map[string]string{"version": "1"}
	if err := WriteJSONAtomic(path, data1); err != nil {
		t.Fatalf("write json 1: %v", err)
	}

	// 上書き
	data2 := map[string]string{"version": "2"}
	if err := WriteJSONAtomic(path, data2); err != nil {
		t.Fatalf("write json 2: %v", err)
	}

	var loaded map[string]string
	if err := ReadJSON(path, &loaded); err != nil {
		t.Fatalf("read json: %v", err)
	}

	if loaded["version"] != "2" {
		t.Errorf("expected version=2, got %s", loaded["version"])
	}
}

func TestWriteJSONAtomicInvalidValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.json")

	// チャンネルはJSONにエンコードできない
	ch := make(chan int)
	err := WriteJSONAtomic(path, ch)
	if err == nil {
		t.Fatalf("expected error for unencodable value, got nil")
	}

	// ファイルが作成されていないことを確認
	if _, statErr := os.Stat(path); statErr == nil {
		t.Errorf("file should not exist after encoding error")
	}
}
