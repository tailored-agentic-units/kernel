package memory_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/tailored-agentic-units/kernel/memory"
)

func TestFileStore_List_EmptyDir(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)

	keys, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("List() returned %d keys, want 0", len(keys))
	}
}

func TestFileStore_List_MissingRoot(t *testing.T) {
	store := memory.NewFileStore(filepath.Join(t.TempDir(), "nonexistent"))

	keys, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("List() returned %d keys, want 0", len(keys))
	}
}

func TestFileStore_List_PopulatedDir(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "memory/global.md", "global notes")
	writeTestFile(t, root, "skills/go-patterns/SKILL.md", "skill def")
	writeTestFile(t, root, "agents/explorer.json", "{}")

	store := memory.NewFileStore(root)
	keys, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	want := []string{
		"agents/explorer.json",
		"memory/global.md",
		"skills/go-patterns/SKILL.md",
	}
	if len(keys) != len(want) {
		t.Fatalf("List() returned %d keys, want %d", len(keys), len(want))
	}
	for i, key := range keys {
		if key != want[i] {
			t.Errorf("List()[%d] = %q, want %q", i, key, want[i])
		}
	}
}

func TestFileStore_List_SkipsHidden(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "visible.md", "content")
	writeTestFile(t, root, ".hidden", "secret")
	writeTestFile(t, root, ".hiddendir/file.md", "nested secret")

	store := memory.NewFileStore(root)
	keys, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("List() returned %d keys, want 1", len(keys))
	}
	if keys[0] != "visible.md" {
		t.Errorf("List()[0] = %q, want %q", keys[0], "visible.md")
	}
}

func TestFileStore_Load(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "memory/global.md", "global notes")
	writeTestFile(t, root, "agents/explorer.json", `{"name":"explorer"}`)

	store := memory.NewFileStore(root)

	entries, err := store.Load(context.Background(), "memory/global.md", "agents/explorer.json")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("Load() returned %d entries, want 2", len(entries))
	}

	if entries[0].Key != "memory/global.md" {
		t.Errorf("entries[0].Key = %q, want %q", entries[0].Key, "memory/global.md")
	}
	if string(entries[0].Value) != "global notes" {
		t.Errorf("entries[0].Value = %q, want %q", string(entries[0].Value), "global notes")
	}

	if entries[1].Key != "agents/explorer.json" {
		t.Errorf("entries[1].Key = %q, want %q", entries[1].Key, "agents/explorer.json")
	}
	if string(entries[1].Value) != `{"name":"explorer"}` {
		t.Errorf("entries[1].Value = %q, want %q", string(entries[1].Value), `{"name":"explorer"}`)
	}
}

func TestFileStore_Load_KeyNotFound(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)

	_, err := store.Load(context.Background(), "nonexistent.md")
	if !errors.Is(err, memory.ErrKeyNotFound) {
		t.Errorf("Load() error = %v, want %v", err, memory.ErrKeyNotFound)
	}
}

func TestFileStore_Save(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)

	entries := []memory.Entry{
		{Key: "memory/global.md", Value: []byte("global notes")},
		{Key: "skills/go-patterns/SKILL.md", Value: []byte("skill def")},
	}

	if err := store.Save(context.Background(), entries...); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify files written
	got, err := os.ReadFile(filepath.Join(root, "memory", "global.md"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "global notes" {
		t.Errorf("file content = %q, want %q", string(got), "global notes")
	}

	got, err = os.ReadFile(filepath.Join(root, "skills", "go-patterns", "SKILL.md"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "skill def" {
		t.Errorf("file content = %q, want %q", string(got), "skill def")
	}
}

func TestFileStore_Save_Overwrite(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)

	if err := store.Save(context.Background(), memory.Entry{Key: "note.md", Value: []byte("v1")}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := store.Save(context.Background(), memory.Entry{Key: "note.md", Value: []byte("v2")}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(root, "note.md"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "v2" {
		t.Errorf("file content = %q, want %q", string(got), "v2")
	}
}

func TestFileStore_Delete(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "memory/global.md", "content")

	store := memory.NewFileStore(root)

	if err := store.Delete(context.Background(), "memory/global.md"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "memory", "global.md")); !os.IsNotExist(err) {
		t.Error("file should not exist after Delete")
	}

	// Parent directory should be cleaned up
	if _, err := os.Stat(filepath.Join(root, "memory")); !os.IsNotExist(err) {
		t.Error("empty parent directory should be removed after Delete")
	}
}

func TestFileStore_Delete_NonExistent(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)

	if err := store.Delete(context.Background(), "nonexistent.md"); err != nil {
		t.Errorf("Delete() error = %v, want nil for missing key", err)
	}
}

func TestFileStore_Delete_PreservesParentWithSiblings(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "memory/a.md", "content a")
	writeTestFile(t, root, "memory/b.md", "content b")

	store := memory.NewFileStore(root)

	if err := store.Delete(context.Background(), "memory/a.md"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Parent should still exist because b.md is there
	if _, err := os.Stat(filepath.Join(root, "memory")); os.IsNotExist(err) {
		t.Error("parent directory should be preserved when sibling files exist")
	}
}

func TestFileStore_RoundTrip(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)

	original := []memory.Entry{
		{Key: "memory/global.md", Value: []byte("global notes")},
		{Key: "skills/go-patterns/SKILL.md", Value: []byte("skill definition")},
		{Key: "agents/explorer.json", Value: []byte(`{"tools":["grep"]}`)},
	}

	if err := store.Save(context.Background(), original...); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	keys, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	loaded, err := store.Load(context.Background(), keys...)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded) != len(original) {
		t.Fatalf("Load() returned %d entries, want %d", len(loaded), len(original))
	}

	got := make(map[string]string, len(loaded))
	for _, entry := range loaded {
		got[entry.Key] = string(entry.Value)
	}
	for _, entry := range original {
		val, ok := got[entry.Key]
		if !ok {
			t.Errorf("key %q not found in loaded entries", entry.Key)
			continue
		}
		if val != string(entry.Value) {
			t.Errorf("key %q: value = %q, want %q", entry.Key, val, string(entry.Value))
		}
	}
}

// writeTestFile creates a file with the given content under root.
func writeTestFile(t *testing.T, root, key, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
