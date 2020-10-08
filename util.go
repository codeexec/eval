package main

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
)

// auto-determine if we're running in production or in dev
// dev vs. productions affects:
// * database (emulated firestore in dev)
// * http address ("localhosts" in dev)
// * templates are re-loaded in dev
func isRunningDev() bool {
	if runtime.GOOS == "darwin" {
		return true
	}
	if runtime.GOOS == "windows" {
		return true
	}
	if runtime.GOOS != "linux" {
		logf(context.Background(), "Unrecognized runtime.GOOS '%s'\n", runtime.GOOS)
	}
	// we assume this is linux and hence production mode
	// TODO: if we ever expect to run on Linux and not in production mode,
	// this needs updating (e.g. check the user we're running as is "presstige")
	return false
}

func printDir(dir string) {
	ctx := context.Background()
	fn := func(path string, info os.FileInfo, err error) error {
		if info == nil {
			logf(ctx, "%s\n", path)
			return nil
		}
		logf(ctx, "%s: %d\n", path, info.Size())
		return nil
	}
	filepath.Walk(dir, fn)
}

/*
func listFiles(ctx context.Context, dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		logf(ctx, "ioutil.ReadDir('%s') failed with '%s'\n", dir, err)
		return
	}
	for _, fi := range files {
		fmt.Printf("%s %d\n", fi.Name(), fi.Size())
	}
}
*/
