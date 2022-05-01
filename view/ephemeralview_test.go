package view_test

import (
	"testing"

	"github.com/firestuff/patchy/view"
	"github.com/stretchr/testify/require"
)

func TestEphemeralView(t *testing.T) {
	t.Parallel()

	v := view.NewEphemeralView([]string{"foo", "bar"})
	defer v.Close()

	msg := <-v.Chan()
	require.Equal(t, []string{"foo", "bar"}, msg)

	v.MustUpdate([]string{"foo", "bar", "zig"})

	msg = <-v.Chan()
	require.Equal(t, []string{"foo", "bar", "zig"}, msg)
}
