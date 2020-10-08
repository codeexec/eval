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
	"github.com/google/shlex"
)

func getClangBuildCommand(cmd string, files []*defs.File) ([]string, error) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		// default build command
		//def := []string{"g++", "-std=c++14", "-lstdc++", "-o", "app"}
		def := []string{"clang++", "-std=c++14", "-o", "app"}
		fileNames := getCppFileNames(files)
		def = append(def, fileNames...)
		return def, nil
	}
	parts, err := shlex.Split(cmd)
	valid := parts[0] == "clang" || parts[0] == "clang++"
	if !valid {
		return nil, fmt.Errorf("got '%s' but expected 'clang' or 'clang++", parts[0])
	}
	return parts, err
}

func findClangExeName(args []string, def string) string {
	for i, s := range args {
		if s == "-o" {
			// TODO: invalid otherwise
			if i+1 < len(args) {
				return args[i+1]
			}
		}
	}
	return def
}

func evalCodeClang(e *defs.EvalRequest, dir string) *defs.EvalResponse {
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
	case "cpp", "clang":
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
			f.Name = genUniqueCppFileName(seenNames)
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

	defExeName := filepath.Base(dir)

	// if the user did over-write :run, run it
	// TODO: also support custom :build
	if m != nil && m.RunCmd != "" {
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

	var exeName string
	{
		args, err := getClangBuildCommand(e.Command, e.Files)
		if err != nil {
			res.Error = err.Error()
			return res
		}
		acmd := args[0]
		cmd := exec.Command(acmd, args[1:]...)
		exeName = findClangExeName(args, defExeName)
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
			return res
		}
	}

	cmd := exec.Command("./" + exeName) // or could add "." to PATH
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
