//nolint:testpackage
package path

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseTimeNow(t *testing.T) {
	t.Parallel()

	start := time.Now()

	now, err := parse("now", time.Time{})
	require.NoError(t, err)

	end := time.Now()

	require.WithinRange(t, now.(*timeVal).time, start, end)
}

func TestParseTimeDate(t *testing.T) {
	t.Parallel()

	parsed, err := parse("2022-11-01-08:00", time.Time{})
	require.NoError(t, err)

	tv := parsed.(*timeVal)
	require.Equal(t, tv.precision, 24*time.Hour)
}

func TestParseTimeSecond(t *testing.T) {
	t.Parallel()

	parsed, err := parse("2022-11-01T05:06:07Z", time.Time{})
	require.NoError(t, err)

	tv := parsed.(*timeVal)
	require.Equal(t, tv.precision, 1*time.Second)
}
