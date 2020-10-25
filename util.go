package main

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
)

func isRunningWindows() bool {
	return runtime.GOOS == "windows"
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
