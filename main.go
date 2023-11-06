package main

import (
	"fmt"
	"os"

	"github.com/stateful/runme/internal/cmd"
	"github.com/stateful/runme/internal/version"
	"go.uber.org/zap"
)

func root() int {
	loggercfg := zap.NewDevelopmentConfig()
	loggercfg.OutputPaths = []string{
		"./runme.log",
	}
	logger, _ := loggercfg.Build()

	root := cmd.Root()
	root.Version = fmt.Sprintf("%s (%s) on %s", version.BuildVersion, version.Commit, version.BuildDate)
	if err := root.Execute(); err != nil {
		logger.Debug("running root command failed", zap.Error(err))
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	return 0
}

func main() {
	os.Exit(root())
}
