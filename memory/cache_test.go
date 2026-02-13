package memory_test

import (
	"context"
	"sync"
	"testing"

	"github.com/tailored-agentic-units/kernel/memory"
)

func TestCache_Bootstrap_IndexOnly(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)
	writeTestFile(t, root, "memory/global.md", "notes")
	writeTestFile(t, root, "skills/a/SKILL.md", "skill")

	cache := memory.NewCache(store)
	if err := cache.Bootstrap(context.Background()); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	// Index populated
	if !cache.Has("memory/global.md") {
		t.Error("Has(memory/global.md) = false, want true")
	}
	if !cache.Has("skills/a/SKILL.md") {
		t.Error("Has(skills/a/SKILL.md) = false, want true")
	}

	// Content not loaded
	if _, ok := cache.Get("memory/global.md"); ok {
		t.Error("Get(memory/global.md) should return false before Resolve")
	}
}

func TestCache_Bootstrap_WithPrefixes(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)
	writeTestFile(t, root, "memory/global.md", "notes")
	writeTestFile(t, root, "skills/a/SKILL.md", "skill")

	cache := memory.NewCache(store)
	if err := cache.Bootstrap(context.Background(), "memory/"); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	// memory/ prefix loaded
	val, ok := cache.Get("memory/global.md")
	if !ok {
		t.Fatal("Get(memory/global.md) = false, want true after Bootstrap with prefix")
	}
	if string(val) != "notes" {
		t.Errorf("Get(memory/global.md) = %q, want %q", string(val), "notes")
	}

	// skills/ not loaded (not in prefix)
	if _, ok := cache.Get("skills/a/SKILL.md"); ok {
		t.Error("Get(skills/a/SKILL.md) should return false, not in bootstrap prefix")
	}

	// But indexed
	if !cache.Has("skills/a/SKILL.md") {
		t.Error("Has(skills/a/SKILL.md) = false, want true")
	}
}

func TestCache_Bootstrap_EmptyStore(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)

	cache := memory.NewCache(store)
	if err := cache.Bootstrap(context.Background(), "memory/"); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	if len(cache.Keys()) != 0 {
		t.Errorf("Keys() returned %d keys, want 0", len(cache.Keys()))
	}
}

func TestCache_Resolve(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)
	writeTestFile(t, root, "skills/a/SKILL.md", "skill content")

	cache := memory.NewCache(store)
	if err := cache.Bootstrap(context.Background()); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	// Not yet loaded
	if _, ok := cache.Get("skills/a/SKILL.md"); ok {
		t.Fatal("Get() should return false before Resolve")
	}

	// Resolve
	if err := cache.Resolve(context.Background(), "skills/a/SKILL.md"); err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	val, ok := cache.Get("skills/a/SKILL.md")
	if !ok {
		t.Fatal("Get() should return true after Resolve")
	}
	if string(val) != "skill content" {
		t.Errorf("Get() = %q, want %q", string(val), "skill content")
	}
}

func TestCache_Resolve_SkipsCached(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)
	writeTestFile(t, root, "memory/global.md", "original")

	cache := memory.NewCache(store)
	if err := cache.Bootstrap(context.Background(), "memory/"); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	// Modify the underlying file
	writeTestFile(t, root, "memory/global.md", "modified")

	// Resolve should skip already-cached key
	if err := cache.Resolve(context.Background(), "memory/global.md"); err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	val, _ := cache.Get("memory/global.md")
	if string(val) != "original" {
		t.Errorf("Resolve should skip cached, got %q, want %q", string(val), "original")
	}
}

func TestCache_Set_And_Get(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)
	cache := memory.NewCache(store)

	cache.Set("memory/session.md", []byte("session notes"))

	val, ok := cache.Get("memory/session.md")
	if !ok {
		t.Fatal("Get() should return true after Set")
	}
	if string(val) != "session notes" {
		t.Errorf("Get() = %q, want %q", string(val), "session notes")
	}

	// Should also be in index
	if !cache.Has("memory/session.md") {
		t.Error("Has() should return true after Set")
	}
}

func TestCache_Get_DefensiveCopy(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)
	cache := memory.NewCache(store)

	cache.Set("key", []byte("original"))

	val, _ := cache.Get("key")
	val[0] = 'X'

	got, _ := cache.Get("key")
	if string(got) != "original" {
		t.Errorf("Get() returned mutable reference, got %q after mutation", string(got))
	}
}

func TestCache_Set_DefensiveCopy(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)
	cache := memory.NewCache(store)

	input := []byte("original")
	cache.Set("key", input)
	input[0] = 'X'

	val, _ := cache.Get("key")
	if string(val) != "original" {
		t.Errorf("Set() did not copy input, got %q after mutation", string(val))
	}
}

func TestCache_Delete(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)
	cache := memory.NewCache(store)

	cache.Set("key", []byte("value"))
	cache.Delete("key")

	if _, ok := cache.Get("key"); ok {
		t.Error("Get() should return false after Delete")
	}
	if cache.Has("key") {
		t.Error("Has() should return false after Delete")
	}
}

func TestCache_Keys(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)
	writeTestFile(t, root, "b.md", "b")
	writeTestFile(t, root, "a.md", "a")
	writeTestFile(t, root, "c.md", "c")

	cache := memory.NewCache(store)
	if err := cache.Bootstrap(context.Background()); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	keys := cache.Keys()
	want := []string{"a.md", "b.md", "c.md"}
	if len(keys) != len(want) {
		t.Fatalf("Keys() returned %d keys, want %d", len(keys), len(want))
	}
	for i, key := range keys {
		if key != want[i] {
			t.Errorf("Keys()[%d] = %q, want %q", i, key, want[i])
		}
	}
}

func TestCache_Entries(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)
	writeTestFile(t, root, "memory/a.md", "a")
	writeTestFile(t, root, "memory/b.md", "b")
	writeTestFile(t, root, "skills/x/SKILL.md", "x")

	cache := memory.NewCache(store)
	if err := cache.Bootstrap(context.Background(), "memory/", "skills/"); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	entries := cache.Entries("memory/")
	if len(entries) != 2 {
		t.Fatalf("Entries(memory/) returned %d entries, want 2", len(entries))
	}
	if entries[0].Key != "memory/a.md" {
		t.Errorf("entries[0].Key = %q, want %q", entries[0].Key, "memory/a.md")
	}
	if entries[1].Key != "memory/b.md" {
		t.Errorf("entries[1].Key = %q, want %q", entries[1].Key, "memory/b.md")
	}
}

func TestCache_Entries_OnlyCached(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)
	writeTestFile(t, root, "memory/a.md", "a")

	cache := memory.NewCache(store)
	if err := cache.Bootstrap(context.Background()); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	// Indexed but not loaded — should not appear in Entries
	entries := cache.Entries("memory/")
	if len(entries) != 0 {
		t.Errorf("Entries() returned %d entries for unloaded prefix, want 0", len(entries))
	}
}

func TestCache_Flush(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)
	cache := memory.NewCache(store)

	cache.Set("memory/new.md", []byte("new content"))

	if err := cache.Flush(context.Background()); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	// Verify persisted to store
	entries, err := store.Load(context.Background(), "memory/new.md")
	if err != nil {
		t.Fatalf("Store.Load() error = %v", err)
	}
	if string(entries[0].Value) != "new content" {
		t.Errorf("persisted value = %q, want %q", string(entries[0].Value), "new content")
	}
}

func TestCache_Flush_DeletesRemoved(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)
	writeTestFile(t, root, "memory/old.md", "old")

	cache := memory.NewCache(store)
	if err := cache.Bootstrap(context.Background(), "memory/"); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	cache.Delete("memory/old.md")

	if err := cache.Flush(context.Background()); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	// Verify removed from store
	_, err := store.Load(context.Background(), "memory/old.md")
	if err == nil {
		t.Error("Store.Load() should error after flushing deleted key")
	}
}

func TestCache_Flush_OnlyDirty(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)
	writeTestFile(t, root, "memory/existing.md", "original")

	cache := memory.NewCache(store)
	if err := cache.Bootstrap(context.Background(), "memory/"); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	// Flush without modifications
	if err := cache.Flush(context.Background()); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	// Original file untouched
	entries, err := store.Load(context.Background(), "memory/existing.md")
	if err != nil {
		t.Fatalf("Store.Load() error = %v", err)
	}
	if string(entries[0].Value) != "original" {
		t.Errorf("value = %q, want %q", string(entries[0].Value), "original")
	}
}

func TestCache_Flush_ClearsDirtyTracking(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)
	cache := memory.NewCache(store)

	cache.Set("key", []byte("v1"))
	if err := cache.Flush(context.Background()); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	// Modify the file directly on disk
	writeTestFile(t, root, "key", "v2-on-disk")

	// Second flush should not re-save the key (dirty tracking was cleared)
	if err := cache.Flush(context.Background()); err != nil {
		t.Fatalf("second Flush() error = %v", err)
	}

	entries, err := store.Load(context.Background(), "key")
	if err != nil {
		t.Fatalf("Store.Load() error = %v", err)
	}
	if string(entries[0].Value) != "v2-on-disk" {
		t.Errorf("second flush overwrote file, got %q, want %q", string(entries[0].Value), "v2-on-disk")
	}
}

func TestCache_Concurrent_GetSet(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)
	cache := memory.NewCache(store)
	const n = 100

	var wg sync.WaitGroup
	wg.Add(2 * n)

	for range n {
		go func() {
			defer wg.Done()
			cache.Set("key", []byte("value"))
		}()
		go func() {
			defer wg.Done()
			cache.Get("key")
		}()
	}
	wg.Wait()
}

func TestCache_Concurrent_SetDelete(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)
	cache := memory.NewCache(store)
	const n = 100

	var wg sync.WaitGroup
	wg.Add(2 * n)

	for range n {
		go func() {
			defer wg.Done()
			cache.Set("key", []byte("value"))
		}()
		go func() {
			defer wg.Done()
			cache.Delete("key")
		}()
	}
	wg.Wait()
}

func TestCache_Concurrent_HasKeys(t *testing.T) {
	root := t.TempDir()
	store := memory.NewFileStore(root)
	cache := memory.NewCache(store)
	const n = 100

	var wg sync.WaitGroup
	wg.Add(3 * n)

	for range n {
		go func() {
			defer wg.Done()
			cache.Set("key", []byte("value"))
		}()
		go func() {
			defer wg.Done()
			cache.Has("key")
		}()
		go func() {
			defer wg.Done()
			cache.Keys()
		}()
	}
	wg.Wait()
}
