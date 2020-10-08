package main

import (
	"context"
	"fmt"
)

var (
	verbose bool
)

func logf(ctx context.Context, format string, args ...interface{}) {
	s := format
	if len(args) > 0 {
		s = fmt.Sprintf(format, args...)
	}
	fmt.Print(s)
}

func verbosef(ctx context.Context, format string, args ...interface{}) {
	if !verbose {
		return
	}
	logf(ctx, format, args...)
}
