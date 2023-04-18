package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/dchest/uniuri"
	"github.com/firestuff/patchy"
	"github.com/firestuff/patchy/api"
	"github.com/firestuff/patchy/patchyc"
)

type testType struct {
	api.Metadata
	Text string `json:"text"`
	Num  int64  `json:"num"`
}

func main() {
	dbname := fmt.Sprintf("file:%s?mode=memory&cache=shared", uniuri.New())

	api, err := patchy.NewSQLiteAPI(dbname)
	if err != nil {
		panic(err)
	}

	patchy.Register[testType](api)

	api.ServeFiles("/_test/*filepath", http.Dir("../browser/"))

	err = api.ListenSelfCert("[::]:8080")
	if err != nil {
		panic(err)
	}

	hn, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	baseURL := fmt.Sprintf("https://%s/", net.JoinHostPort(hn, strconv.Itoa(api.Addr().Port)))

	log.Printf("listening on: %s", baseURL)

	go func() {
		ctx := context.Background()

		pyc := patchyc.NewClient(baseURL).
			SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}) //nolint:gosec

		tc, err := pyc.TSClient(ctx)
		if err != nil {
			panic(err)
		}

		err = os.WriteFile("../browser/client.ts", []byte(tc), 0o600)
		if err != nil {
			panic(err)
		}

		cmd := exec.Command("tsc", "--pretty")
		cmd.Dir = "../browser"

		out, err := cmd.Output()
		if err != nil {
			panic(string(out))
		}

		log.Printf("build complete")
	}()

	err = api.Serve()
	if err != nil {
		panic(err)
	}
}
