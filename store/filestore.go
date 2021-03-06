package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/firestuff/patchy/metadata"
)

type FileStore struct {
	root string
}

func NewFileStore(root string) *FileStore {
	return &FileStore{
		root: root,
	}
}

func (s *FileStore) Write(t string, obj any) error {
	id := getSafeID(metadata.GetMetadata(obj).ID)
	dir := filepath.Join(s.root, t, id[:4])

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, fmt.Sprintf("%s.*", id))
	if err != nil {
		return err
	}
	defer tmp.Close()

	enc := json.NewEncoder(tmp)
	enc.SetEscapeHTML(false)

	err = enc.Encode(obj)
	if err != nil {
		return err
	}

	err = tmp.Close()
	if err != nil {
		return err
	}

	err = os.Rename(tmp.Name(), filepath.Join(dir, id))
	if err != nil {
		return err
	}

	return nil
}

func (s *FileStore) Delete(t string, id string) error {
	safeID := getSafeID(id)
	dir := filepath.Join(s.root, t, safeID[:4])

	return os.Remove(filepath.Join(dir, safeID))
}

func (s *FileStore) Read(t string, id string, factory func() any) (any, error) {
	safeID := getSafeID(id)
	dir := filepath.Join(s.root, t, safeID[:4])

	obj := factory()

	err := s.read(filepath.Join(dir, safeID), obj)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, err
	}

	return obj, nil
}

func (s *FileStore) List(t string, factory func() any) ([]any, error) {
	dir := filepath.Join(s.root, t)
	fsys := os.DirFS(dir)

	ret := []any{}

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if strings.Contains(d.Name(), ".") {
			// Temporary file
			return nil
		}

		obj := factory()

		err = s.read(filepath.Join(dir, path), obj)
		if err != nil {
			return err
		}

		ret = append(ret, obj)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (s *FileStore) read(path string, obj any) error {
	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fh.Close()

	dec := json.NewDecoder(fh)
	dec.DisallowUnknownFields()

	return dec.Decode(obj)
}

func getSafeID(id string) string {
	// TODO: Make this an hmac to prevent partial collision DoS attacks
	h := sha256.New()
	h.Write([]byte(id))

	return hex.EncodeToString(h.Sum(nil))
}
