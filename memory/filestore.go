package memory

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type fileStore struct {
	root string
}

// NewFileStore creates a Store backed by the filesystem. Keys map 1:1 to
// relative file paths under root.
func NewFileStore(root string) Store {
	return &fileStore{root: root}
}

func (s *fileStore) List(_ context.Context) ([]string, error) {
	var keys []string

	err := filepath.WalkDir(s.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) && path == s.root {
				return fs.SkipAll
			}
			return err
		}

		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(s.root, path)
		if err != nil {
			return err
		}
		keys = append(keys, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLoadFailed, err)
	}

	return keys, nil
}

func (s *fileStore) Load(_ context.Context, keys ...string) ([]Entry, error) {
	entries := make([]Entry, 0, len(keys))

	for _, key := range keys {
		path := filepath.Join(s.root, filepath.FromSlash(key))
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
			}
			return nil, fmt.Errorf("%w: %s: %v", ErrLoadFailed, key, err)
		}
		entries = append(entries, Entry{Key: key, Value: data})
	}

	return entries, nil
}

func (s *fileStore) Save(_ context.Context, entries ...Entry) error {
	for _, e := range entries {
		path := filepath.Join(s.root, filepath.FromSlash(e.Key))

		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("%w: %s: %v", ErrSaveFailed, e.Key, err)
		}

		tmp, err := os.CreateTemp(dir, ".tmp-*")
		if err != nil {
			return fmt.Errorf("%w: %s: %v", ErrSaveFailed, e.Key, err)
		}
		tmpName := tmp.Name()

		if _, err := tmp.Write(e.Value); err != nil {
			tmp.Close()
			os.Remove(tmpName)
			return fmt.Errorf("%w: %s: %v", ErrSaveFailed, e.Key, err)
		}
		if err := tmp.Close(); err != nil {
			os.Remove(tmpName)
			return fmt.Errorf("%w: %s: %v", ErrSaveFailed, e.Key, err)
		}

		if err := os.Rename(tmpName, path); err != nil {
			os.Remove(tmpName)
			return fmt.Errorf("%w: %s: %v", ErrSaveFailed, e.Key, err)
		}
	}

	return nil
}

func (s *fileStore) Delete(_ context.Context, keys ...string) error {
	for _, key := range keys {
		path := filepath.Join(s.root, filepath.FromSlash(key))
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("delete failed: %s: %w", key, err)
		}

		dir := filepath.Dir(path)
		for dir != s.root {
			if err := os.Remove(dir); err != nil {
				break
			}
			dir = filepath.Dir(dir)
		}
	}

	return nil
}
