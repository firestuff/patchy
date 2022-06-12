package view_test

import (
	"context"
	"testing"

	"github.com/firestuff/patchy/view"
	"github.com/stretchr/testify/require"
)

func TestFilterView(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	ev, err := view.NewEphemeralView(ctx, []string{"foo", "bar"})
	require.Nil(t, err)

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

	cancel()

	for {
		err := ev.Update([]string{"zig", "zag"})

		if _, ok := <-fv.Chan(); !ok {
			require.NotNil(t, err)
			break
		}
	}
}
