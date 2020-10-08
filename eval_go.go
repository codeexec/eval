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

func getGoBuildCommand(cmd string) ([]string, error) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		// default build command
		return []string{"go", "build", "-o", "app"}, nil
	}
	parts, err := shlex.Split(cmd)
	if parts[0] != "go" {
		return nil, fmt.Errorf("got '%s' but expected 'go'", parts[0])
	}
	return parts, err
}

func findGoExeName(args []string, def string) string {
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

func genUniqueGoFileName(seenNames map[string]bool) string {
	for i := 0; i < 100; i++ {
		name := "main.go"
		if i > 0 {
			name = fmt.Sprintf("main-%d.go", i)
		}
		if !seenNames[name] {
			seenNames[name] = true
			return name
		}
	}
	panic("couldn't generate unique Go name")
}

func evalCodeGo(e *defs.EvalRequest, dir string) *defs.EvalResponse {
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
	case "go":
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
			f.Name = genUniqueGoFileName(seenNames)
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
	{
		// synthesize go.mod if doesn't exist
		hasGoMod := false
		for _, f := range e.Files {
			if f.Name == "go.mod" {
				hasGoMod = true
				break
			}
		}
		if !hasGoMod {
			defExeName = "test"
			cmd := exec.Command("go", "mod", "init", "test")
			cmd.Dir = dir
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			verbosef(ctx, "> %s\n", cmd)
			err := cmd.Run()
			// TODO: maybe ignore this error
			if err != nil {
				res.Stdout = stdout.String()
				res.Stderr = stderr.String()
				res.Error = err.Error()
				return res
			}
		}
	}

	// if the user did over-write :run, run it
	// TODO: also support custom :build
	if m != nil && m.RunCmd != "" {
		logf(ctx, "running custom run command: '%s'\n", m.RunCmd)
		args, err := getGoBuildCommand(m.RunCmd)
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
		args, err := getGoBuildCommand(e.Command)
		if err != nil {
			res.Error = err.Error()
			return res
		}
		acmd := args[0]
		cmd := exec.Command(acmd, args[1:]...)
		exeName = findGoExeName(args, defExeName)
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
