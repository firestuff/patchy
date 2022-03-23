package storebus

import "encoding/json"
import "fmt"
import "os"
import "path/filepath"

type Store struct {
	root string
}

func NewStore(root string) *Store {
	return &Store{
		root: root,
	}
}

func (s *Store) Write(t string, obj Object) error {
	dir := filepath.Join(s.root, t)
	filename := ObjectSafeId(obj)

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

func (s *Store) Delete(t string, obj Object) error {
	dir := filepath.Join(s.root, t)
	filename := ObjectSafeId(obj)
	return os.Remove(filepath.Join(dir, filename))
}

func (s *Store) Read(t string, obj Object) error {
	dir := filepath.Join(s.root, t)
	filename := ObjectSafeId(obj)

	fh, err := os.Open(filepath.Join(dir, filename))
	if err != nil {
		return err
	}
	defer fh.Close()

	dec := json.NewDecoder(fh)
	dec.DisallowUnknownFields()

	err = dec.Decode(obj)
	if err != nil {
		return err
	}

	return nil
}
