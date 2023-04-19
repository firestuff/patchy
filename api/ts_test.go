package api_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTSNode(t *testing.T) {
	t.Parallel()

	dir := buildTS(t, "node")
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	paths, err := filepath.Glob(filepath.Join(dir, "*_test.js"))
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
		"NODE_DEBUG":                   os.Getenv("NODE_DEBUG"),
		"NODE_NO_WARNINGS":             "1",
		"NODE_TLS_REJECT_UNAUTHORIZED": "0",
		"BASE_URL":                     ta.baseURL,
	}

	runNoError(t, filepath.Dir(path), env, "node", "--enable-source-maps", filepath.Base(path))

	ta.checkTests(t)
}

func buildTS(t *testing.T, env string) string {
	dir, err := os.MkdirTemp("", "ts_test")
	require.NoError(t, err)

	paths, err := filepath.Glob("ts_test/*")
	require.NoError(t, err)

	for _, path := range paths {
		src, err := filepath.Abs(path)
		require.NoError(t, err)

		base := filepath.Base(path)

		if strings.Contains(base, ":") {
			parts := strings.SplitN(base, ":", 2)

			if parts[0] == env {
				base = parts[1]
			} else {
				continue
			}
		}

		err = os.Symlink(src, filepath.Join(dir, base))
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
