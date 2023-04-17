package api_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/firestuff/patchy/api"
	"github.com/stretchr/testify/require"
)

type complexTestType struct {
	api.Metadata
	A string
	B int
	C []string
	D nestedType
	E *nestedType
}

type nestedType struct {
	F []int
	G string
}

func TestTemplateGoClient(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	api.Register[complexTestType](ta.api)

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

	runNoError(t, dir, nil, "go", "mod", "init", "test")
	runNoError(t, dir, nil, "go", "mod", "tidy")
	runNoError(t, dir, nil, "go", "vet", ".")
	runNoError(t, dir, nil, "go", "build", ".")
}

func runNoError(t *testing.T, dir string, env map[string]string, name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir

	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	out, err := cmd.Output()
	t.Logf("cmd='%s'\nargs=%v\nout='%s'\nerr='%s'", name, arg, string(out), getStderr(err))
	require.NoError(t, err)
}

func getStderr(err error) string {
	ee := &exec.ExitError{}
	if errors.As(err, &ee) {
		return string(ee.Stderr)
	}

	return ""
}
