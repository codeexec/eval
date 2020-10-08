package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/codeexec/eval/defs"
)

func dbgEval(e *defs.EvalRequest) {
	if !verbose {
		return
	}
	ctx := context.Background()
	// for _, f := range e.Files {
	// 	verbosef(ctx, "File: %s\nContent:\n%s\n", f.Name, f.Content)
	// }
	verbosef(ctx, "Language: '%s'\n", e.Language)
	if e.Command != "" {
		verbosef(ctx, "Command: '%s'\n", e.Command)
	}
	for _, f := range e.Files {
		verbosef(ctx, "File: %s, size: %d\n", f.Name, len(f.Content))
	}
}

func eval(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method == http.MethodGet {
		fmt.Fprint(w, "eval!\n")
		return
	}
	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "ioutil.ReadAll() failed with '%s'\n", err)
		return
	}
	verbosef(ctx, "eval: '%s'\n", r.URL)
	var eval defs.EvalRequest
	err = json.Unmarshal(d, &eval)
	if err != nil {
		fmt.Fprintf(w, "json.Unmarshal() failed with '%s'\n", err)
		return
	}

	var rsp *defs.EvalResponse
	dir, err := ioutil.TempDir("/tmp", "src")
	if err != nil {
		rsp.Error = fmt.Sprintf("ioutil.TempDir() failed with '%s'", err)
		serveJSON(w, r, http.StatusOK, rsp)
		return
	}
	verbosef(ctx, "created directory '%s'\n", dir)
	defer os.RemoveAll(dir)

	dbgEval(&eval)

	lang := strings.ToLower(eval.Language)

	switch lang {
	case "go":
		rsp = evalCodeGo(&eval, dir)
	case "cpp", "gcc":
		rsp = evalCodeGcc(&eval, dir)
	case "clang":
		rsp = evalCodeClang(&eval, dir)
	case "node":
		rsp = evalCodeNode(&eval, dir)
	default:
		rsp = &defs.EvalResponse{
			Error: fmt.Sprintf("Unsupported language '%s'", eval.Language),
		}
	}

	rsp.Files = getEvalFiles(ctx, dir)

	serveJSON(w, r, http.StatusOK, rsp)
}

func getEvalFiles(ctx context.Context, dir string) []*defs.EvalFile {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		logf(ctx, "ioutil.ReadDir('%s') failed with '%s'\n", dir, err)
		return nil
	}
	var res []*defs.EvalFile
	for _, fi := range files {
		f := &defs.EvalFile{
			Name: fi.Name(),
			Size: fi.Size(),
		}
		res = append(res, f)
	}
	return res
}

/*
func printEnv() {
	ctx := context.Background()
	for _, e := range os.Environ() {
		logf(ctx, "%s\n", e)
	}
}

func addToPath(dir string) {
	ctx := context.Background()
	path := os.Getenv("PATH")
	if strings.Contains(path, dir) {
		logf(ctx, "'%s' already in PATH\n", dir)
		return
	}
	path = path + ":" + dir
	os.Setenv("PATH", path)
}
*/
