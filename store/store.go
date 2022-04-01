package store

import "encoding/json"
import "fmt"
import "os"
import "path/filepath"

import "github.com/firestuff/patchy/metadata"

type Store struct {
	root string
}

func NewStore(root string) *Store {
	return &Store{
		root: root,
	}
}

func (s *Store) Write(t string, obj any) error {
	dir := filepath.Join(s.root, t)
	filename := metadata.GetMetadata(obj).GetSafeId()

	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, fmt.Sprintf("%s.*", filename))
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

	err = os.Rename(tmp.Name(), filepath.Join(dir, filename))
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Delete(t string, obj any) error {
	dir := filepath.Join(s.root, t)
	filename := metadata.GetMetadata(obj).GetSafeId()
	return os.Remove(filepath.Join(dir, filename))
}

func (s *Store) Read(t string, obj any) error {
	dir := filepath.Join(s.root, t)
	filename := metadata.GetMetadata(obj).GetSafeId()
	return s.read(filepath.Join(dir, filename), obj)
}

func (s *Store) List(t string, factory func() any) ([]any, error) {
	dir := filepath.Join(s.root, t)

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	ret := []any{}

	for _, entry := range files {
		if entry.IsDir() {
			continue
		}

		obj := factory()

		err = s.read(filepath.Join(dir, entry.Name()), obj)
		if err != nil {
			return nil, err
		}

		ret = append(ret, obj)
	}

	return ret, nil
}

func (s *Store) read(path string, obj any) error {
	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fh.Close()

	dec := json.NewDecoder(fh)
	dec.DisallowUnknownFields()

	return dec.Decode(obj)
}
