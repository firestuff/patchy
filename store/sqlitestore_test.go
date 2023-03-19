package store_test

import (
	"testing"

	"github.com/firestuff/patchy/store"
	"github.com/stretchr/testify/require"
)

func TestSQLiteStore(t *testing.T) {
	t.Parallel()

	st, err := store.NewSQLiteStore("file:testStore?mode=memory&cache=shared")
	require.NoError(t, err)
	defer st.Close()

	testStorer(t, st)
}

func TestSQLiteDelete(t *testing.T) {
	t.Parallel()

	st, err := store.NewSQLiteStore("file:testDelete?mode=memory&cache=shared")
	require.NoError(t, err)
	defer st.Close()

	testDelete(t, st)
}

func TestSQLiteList(t *testing.T) {
	t.Parallel()

	st, err := store.NewSQLiteStore("file:testList?mode=memory&cache=shared")
	require.NoError(t, err)
	defer st.Close()

	testList(t, st)
}
