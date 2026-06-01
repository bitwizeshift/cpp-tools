package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/pflag"
	"rodusek.dev/pkg/cpp-tools/internal/diagnostic"
	"rodusek.dev/pkg/cpp-tools/internal/gitdiff"
	"rodusek.dev/pkg/cpp-tools/internal/linefilter"
)

func main() {
	fs := pflag.CommandLine
	var cached bool
	var loggerFlag diagnostic.LoggerFlag
	fs.BoolVar(&cached, "cached", false, "Use the staged changes instead of the working tree")
	loggerFlag.RegisterFlags(fs)

	pflag.Parse()

	logger := loggerFlag.Logger(os.Stderr)

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	diff, err := gitdiff.DefaultLister.List(ctx, &gitdiff.ListOptions{
		Indexed: cached,
		Paths:   pflag.Args(),
	})
	if err != nil {
		logger.Error(ctx, &diagnostic.Diagnostic{
			ID:      "bad-diff",
			Title:   "Unable to parse git-diff output",
			Message: err.Error(),
		})
		os.Exit(1)
	}

	filters := linefilter.FromDiff(diff...)
	if len(filters) == 0 {
		fmt.Println("[]")
	} else {
		bytes, err := json.Marshal(filters)
		if err != nil {
			panic(err) // this one shouldn't happen
		}
		fmt.Println(string(bytes))
	}
}
