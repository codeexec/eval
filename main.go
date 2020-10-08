package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kjk/u"
)

func index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	verbosef(ctx, "index: '%s'\n", r.URL)
	fmt.Fprint(w, "Hello strager. You can contact me via https://blog.kowalczyk.info/contactme.html\n")
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

func main() {
	//addToPath("/usr/local/go/bin")
	//printEnv()

	// TODO: do verbose by default
	if true || os.Getenv("VERBOSE") == "true" {
		verbose = true
	}
	http.HandleFunc("/", index)
	http.HandleFunc("/eval", eval)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if isRunningDev() {
		port = "8533"
	}

	addr := ":" + port
	uri := "http://" + addr

	if isRunningDev() {
		verbose = true
		addr = "localhost:" + port
		uri = "http://" + addr
	}
	ctx := context.Background()
	logf(ctx, "starting on '%s' dev: %v, verbose: %v\n", uri, isRunningDev(), verbose)

	if isRunningDev() {
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
