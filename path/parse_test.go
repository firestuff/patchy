//nolint: testpackage
package path

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseTime(t *testing.T) {
	t.Parallel()

	start := time.Now()

	now, err := parse("now", time.Time{})
	require.Nil(t, err)

	end := time.Now()

	require.WithinRange(t, now.(*timeVal).time, start, end)
}
