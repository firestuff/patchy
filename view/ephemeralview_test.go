package view_test

import (
	"context"
	"testing"

	"github.com/firestuff/patchy/view"
	"github.com/stretchr/testify/require"
)

func TestEphemeralView(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	v, err := view.NewEphemeralView(ctx, []string{"foo", "bar"})
	require.Nil(t, err)

	msg := <-v.Chan()
	require.Equal(t, []string{"foo", "bar"}, msg)

	v.MustUpdate([]string{"foo", "bar", "zig"})

	msg = <-v.Chan()
	require.Equal(t, []string{"foo", "bar", "zig"}, msg)

	cancel()

	for {
		err := v.Update([]string{"zig", "zag"})

		if _, ok := <-v.Chan(); !ok {
			require.NotNil(t, err)
			break
		}
	}
}
