package view_test

import (
	"testing"

	"github.com/firestuff/patchy/view"
	"github.com/stretchr/testify/require"
)

func TestFilterView(t *testing.T) {
	t.Parallel()

	ev := view.NewEphemeralView([]string{"foo", "bar"})

	fv := view.NewFilterView[[]string](ev, func(in []string) (out []string) {
		for _, s := range in {
			if s != "bar" {
				out = append(out, s)
			}
		}

		return
	})

	msg := <-fv.Chan()
	require.Equal(t, []string{"foo"}, msg)

	ev.MustUpdate([]string{"foo", "bar", "zig"})

	msg = <-fv.Chan()
	require.Equal(t, []string{"foo", "zig"}, msg)

	ev.Close()

	_, ok := <-fv.Chan()
	require.False(t, ok)
}
