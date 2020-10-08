package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/codeexec/eval/defs"
)

func genUniqueJSFileName(seenNames map[string]bool) string {
	for i := 0; i < 100; i++ {
		name := "main.js"
		if i > 0 {
			name = fmt.Sprintf("main-%d.js", i)
		}
		if !seenNames[name] {
			seenNames[name] = true
			return name
		}
	}
	panic("couldn't generate unique node name")
}

func isJSFile(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".js":
		return true
	}
	return false
}

func getJSFileNames(files []*defs.File) []string {
	var res []string
	for _, f := range files {
		if isJSFile(f.Name) {
			res = append(res, f.Name)
		}
	}
	return res
}

func evalCodeNode(e *defs.EvalRequest, dir string) *defs.EvalResponse {
	var err error
	ctx := context.Background()
	res := &defs.EvalResponse{}
	timeStart := time.Now()
	defer func() {
		dur := time.Since(timeStart)
		res.DurationMS = float64(dur) / float64(time.Millisecond)
	}()

	if len(e.Files) == 0 {
		res.Error = "There are not files"
		return res
	}
	lang := strings.ToLower(e.Language)
	if lang == "" {
		res.Error = "'lang' not specified"
		return res
	}
	switch lang {
	case "node":
		// known languages
	default:
		res.Error = fmt.Sprintf("'%s' is not a supported language", lang)
		return res
	}

	seenNames := map[string]bool{}
	for _, f := range e.Files {
		s := f.Name
		if s == "" {
			continue
		}
		if seenNames[s] {
			// TODO: maybe could be more forgiving and auto-generate unique name
			res.Error = fmt.Sprintf("Duplicate file naem '%s'", s)
			return res
		}
		seenNames[s] = true
	}

	var m *EvalMeta
	for _, f := range e.Files {
		if f.Name == "" {
			f.Name = genUniqueJSFileName(seenNames)
		}
		path := filepath.Join(dir, f.Name)
		err = ioutil.WriteFile(path, []byte(f.Content), 0644)
		if err != nil {
			res.Error = err.Error()
			return res
		}
		if m == nil {
			m = parseMetaFromText(f.Content)
		}
	}

	// if the user did over-write :run, run it
	if m != nil && m.RunCmd != "" {
		panic("NYI")
		logf(ctx, "running custom run command: '%s'\n", m.RunCmd)
		args, err := getClangBuildCommand(m.RunCmd, nil)
		if err != nil {
			res.Error = err.Error()
			return res
		}
		acmd := args[0]
		cmd := exec.Command(acmd, args[1:]...)
		cmd.Dir = dir
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		verbosef(ctx, "> %s\n", cmd)
		err = cmd.Run()
		res.Stdout = stdout.String()
		res.Stderr = stderr.String()
		if err != nil {
			res.Error = err.Error()
		}
		// this over-writes
		return res
	}

	// TODO: need to figure out which file to pick if more than one
	files := getJSFileNames(e.Files)
	cmd := exec.Command("node", files[0])
	cmd.Dir = dir
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if len(e.Stdin) > 0 {
		cmd.Stdin = bytes.NewBufferString(e.Stdin)
	}
	verbosef(ctx, "> %s\n", cmd)
	err = cmd.Run()
	res.Stdout = stdout.String()
	res.Stderr = stderr.String()
	if err != nil {
		printDir(dir)
		res.Error = err.Error()
	}
	return res
}
