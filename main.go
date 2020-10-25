package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kjk/u"
)

// /
func index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	verbosef(ctx, "index: '%s'\n", r.URL)
	fmt.Fprint(w, "Hello stranger. You can contact me via https://blog.kowalczyk.info/contactme.html\n")
}

// /ping
func ping(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	verbosef(ctx, "ping: '%s'\n", r.URL)
	fmt.Fprint(w, "pong\n")
}

func serveJSON(w http.ResponseWriter, r *http.Request, code int, v interface{}) {
	ctx := r.Context()
	d, err := json.Marshal(v)
	if err != nil {
		logf(ctx, "json.Marshal() failed with '%s'", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "json.Marshal() failed with '%s'", err)
		return
	}

	w.Header().Set("content-type", "text/json")
	w.WriteHeader(code)
	_, _ = w.Write(d)
}

var (
	flgDeployGcr      bool
	flgBuildDocker    bool
	flgRunDockerLocal bool
	flgRunUnitTests   bool
	// in dev:
	// * we run emulataed firestore database
	// * run on "localhost" http address and a different http address
	// * templates are re-loaded
	flgDev     bool
	flgVerbose bool
)

func main() {
	//addToPath("/usr/local/go/bin")
	//printEnv()

	flag.BoolVar(&flgDeployGcr, "deploy-gcr", false, "builds docker image for gcr and deploys it to gcr")
	flag.BoolVar(&flgBuildDocker, "build-docker", false, "builder docker image locally eval-multi-20_04")
	flag.BoolVar(&flgRunDockerLocal, "run-docker", false, "build and run docker images locally. can access on http://localhost:8080")
	flag.BoolVar(&flgRunUnitTests, "unit-tests", false, "run unit tests")
	flag.BoolVar(&flgVerbose, "verbose", false, "run one of the do commands")
	flag.Parse()

	// TODO: do verbose by default
	if true || os.Getenv("VERBOSE") == "true" {
		flgVerbose = true
	}
	if !flgDev {
		flgDev = isRunningWindows()
	}

	if flgRunUnitTests {
		runUnitTests()
		return
	}

	if flgBuildDocker {
		buildDockerLocal("main")
		return
	}

	if flgDeployGcr {
		deployGcr()
		return
	}

	if flgRunDockerLocal {
		buildDockerLocal("multi")
		runDockerLocal("multi")
		return
	}

	// no point running on Windows
	if isRunningWindows() {
		flag.Usage()
		return
	}

	http.HandleFunc("/", index)
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/eval", eval)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// TODO: maybe not necessary
	if flgDev {
		flgVerbose = true
		port = "8533"
	}

	verbose = flgVerbose

	addr := ":" + port
	uri := "http://" + addr

	if flgDev {
		addr = "localhost:" + port
		uri = "http://" + addr
	}
	ctx := context.Background()
	logf(ctx, "starting on '%s' dev: %v, verbose: %v\n", uri, flgDev, verbose)

	if isRunningWindows() {
		go func() {
			time.Sleep(time.Second)
			u.OpenBrowser(uri)
		}()
	}

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		logf(ctx, "http.ListendAndServe() failed with '%s'\n", err)
	}
}
