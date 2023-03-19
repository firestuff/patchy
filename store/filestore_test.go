package store_test

import (
	"os"
	"testing"

	"github.com/firestuff/patchy/store"
	"github.com/stretchr/testify/require"
)

func TestFileStore(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	defer os.RemoveAll(dir)

	st := store.NewFileStore(dir)

	testStorer(t, st)
}

func TestFileStoreDelete(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	defer os.RemoveAll(dir)

	st := store.NewFileStore(dir)

	testDelete(t, st)
}

func TestFileStoreList(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	defer os.RemoveAll(dir)

	st := store.NewFileStore(dir)

	testList(t, st)
}
