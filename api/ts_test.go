package api_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTSNode(t *testing.T) {
	t.Parallel()

	dir := buildTS(t)
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	paths, err := filepath.Glob(filepath.Join(dir, "*.js"))
	require.NoError(t, err)

	for _, path := range paths {
		path := path

		t.Run(
			path,
			func(t *testing.T) {
				t.Parallel()
				testPath(t, path)
			},
		)
	}
}

func testPath(t *testing.T, path string) {
	ta := newTestAPI(t)
	defer ta.shutdown(t)

	env := map[string]string{
		"NODE_DEBUG":       os.Getenv("NODE_DEBUG"),
		"NODE_NO_WARNINGS": "1",
		"BASE_URL":         ta.baseURL,
	}

	runNoError(t, filepath.Dir(path), env, "node", "--test", "--enable-source-maps", filepath.Base(path))
}

func buildTS(t *testing.T) string {
	dir, err := os.MkdirTemp("", "ts_test")
	require.NoError(t, err)

	paths, err := filepath.Glob("ts_test/*")
	require.NoError(t, err)

	for _, path := range paths {
		abs, err := filepath.Abs(path)
		require.NoError(t, err)

		err = os.Symlink(abs, filepath.Join(dir, filepath.Base(path)))
		require.NoError(t, err)
	}

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	tc, err := ta.pyc.TSClient(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, tc)

	err = os.WriteFile(filepath.Join(dir, "client.ts"), []byte(tc), 0o600)
	require.NoError(t, err)

	runNoError(t, dir, nil, "tsc", "--pretty")

	return dir
}
