package store

import "encoding/json"
import "fmt"
import "io/fs"
import "os"
import "path/filepath"
import "strings"

import "github.com/firestuff/patchy/metadata"

type LocalStore struct {
	root string
}

func NewLocalStore(root string) Storer {
	return &LocalStore{
		root: root,
	}
}

func (s *LocalStore) Write(t string, obj any) error {
	id := metadata.GetMetadata(obj).GetSafeId()
	dir := filepath.Join(s.root, t, id[:4])

	err := os.MkdirAll(dir, 0700)
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

func (s *LocalStore) Delete(t string, id string) error {
	safeId := metadata.GetSafeId(id)
	dir := filepath.Join(s.root, t, safeId[:4])
	return os.Remove(filepath.Join(dir, safeId))
}

func (s *LocalStore) Read(t string, obj any) error {
	id := metadata.GetMetadata(obj).GetSafeId()
	dir := filepath.Join(s.root, t, id[:4])
	return s.read(filepath.Join(dir, id), obj)
}

func (s *LocalStore) List(t string, factory func() any) ([]any, error) {
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

func (s *LocalStore) read(path string, obj any) error {
	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fh.Close()

	dec := json.NewDecoder(fh)
	dec.DisallowUnknownFields()

	return dec.Decode(obj)
}
