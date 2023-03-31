package api_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/firestuff/patchy/api"
	"github.com/stretchr/testify/require"
)

func TestTemplateGoClient(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	api.Register[mergeTestType](ta.api)

	ctx := context.Background()

	gc, err := ta.pyc.GoClient(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, gc)

	t.Log(gc)

	dir, err := os.MkdirTemp("", "goclient")
	require.NoError(t, err)

	defer os.RemoveAll(dir)

	err = os.WriteFile(filepath.Join(dir, "client.go"), []byte(gc), 0o600)
	require.NoError(t, err)

	runNoError(t, dir, "go", "mod", "init", "test")
	runNoError(t, dir, "go", "mod", "tidy")
	runNoError(t, dir, "go", "vet", ".")
	runNoError(t, dir, "go", "build", ".")
}

func TestTemplateTSClient(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	api.Register[mergeTestType](ta.api)

	ctx := context.Background()

	tc, err := ta.pyc.TSClient(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, tc)

	t.Log(tc)

	dir, err := os.MkdirTemp("", "tsclient")
	require.NoError(t, err)

	defer os.RemoveAll(dir)

	err = os.WriteFile(filepath.Join(dir, "client.ts"), []byte(tc), 0o600)
	require.NoError(t, err)

	runNoError(t, dir, "tsc", "--strict", "client.ts")
}

func runNoError(t *testing.T, dir, name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir
	out, err := cmd.Output()
	require.NoError(t, err, "cmd='%s'\nargs=%v\nout='%s'\nerr='%s'", name, arg, string(out), getStderr(err))
}

func getStderr(err error) string {
	ee := &exec.ExitError{}
	if errors.As(err, &ee) {
		return string(ee.Stderr)
	}

	return ""
}
