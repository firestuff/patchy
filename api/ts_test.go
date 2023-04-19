package api_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/firestuff/patchy/api"
	"github.com/stretchr/testify/require"
)

func TestTSNode(t *testing.T) { //nolint:tparallel
	t.Parallel()

	testTS(t, "node", true, testPathNode)
}

func TestTSFirefox(t *testing.T) { //nolint:tparallel
	t.Parallel()

	testTS(t, "browser", false, testPathFirefox)
}

func TestTSChrome(t *testing.T) { //nolint:tparallel
	t.Parallel()

	testTS(t, "browser", false, testPathChrome)
}

func testTS(t *testing.T, env string, parallel bool, runner func(*testing.T, string)) {
	dir := buildTS(t, env)
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	paths, err := filepath.Glob(filepath.Join(dir, "*_test.js"))
	require.NoError(t, err)

	for _, path := range paths {
		path := path

		t.Run(
			filepath.Base(path),
			func(t *testing.T) {
				if parallel {
					t.Parallel()
				}

				runner(t, path)
			},
		)
	}
}

func testPathNode(t *testing.T, path string) {
	ta := newTestAPI(t)
	defer ta.shutdown(t)

	env := map[string]string{
		"NODE_DEBUG":                   os.Getenv("NODE_DEBUG"),
		"NODE_NO_WARNINGS":             "1",
		"NODE_TLS_REJECT_UNAUTHORIZED": "0",
		"BASE_URL":                     ta.baseURL,
	}

	ctx := context.Background()

	runNoError(ctx, t, filepath.Dir(path), env, "node", "--enable-source-maps", filepath.Base(path))

	ta.checkTests(t)
}

func testPathFirefox(t *testing.T, path string) {
	ta := newTestAPIInsecure(t, func(a *api.API) {
		a.ServeFiles("/_ts_test/*filepath", http.Dir(filepath.Dir(path)))
		a.HandlerFunc("GET", "/_ts_test.html", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `<!DOCTYPE html>
<title>patchy Browser Tests</title>
<link rel="icon" href="data:,">

<script type="module" src="_ts_test/%s"></script>
`, filepath.Base(path))
		})
	})
	defer ta.shutdown(t)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-ta.testDone
		cancel()
	}()

	runNoError(ctx, t, "", nil, "firefox", "--headless", "--no-remote", fmt.Sprintf("%s_ts_test.html", ta.baseURL))

	ta.checkTests(t)
}

func testPathChrome(t *testing.T, path string) {
	ta := newTestAPIInsecure(t, func(a *api.API) {
		a.ServeFiles("/_ts_test/*filepath", http.Dir(filepath.Dir(path)))
		a.HandlerFunc("GET", "/_ts_test.html", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `<!DOCTYPE html>
<title>patchy Browser Tests</title>
<link rel="icon" href="data:,">

<script type="module" src="_ts_test/%s"></script>
`, filepath.Base(path))
		})
	})
	defer ta.shutdown(t)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-ta.testDone
		cancel()
	}()

	runNoError(ctx, t, "", nil, "google-chrome", "--headless", "--disable-gpu", "--remote-debugging-port=9222", fmt.Sprintf("%s_ts_test.html", ta.baseURL))

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

	runNoError(ctx, t, dir, nil, "tsc", "--pretty")

	return dir
}
