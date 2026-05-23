package main

import (
	"context"
	"os"
	"runtime/debug"

	"github.com/arloliu/gocensus/internal/cli"
)

var version = "dev"

func main() {
	os.Exit(cli.Run(context.Background(), os.Args[1:], os.Stdout, os.Stderr, effectiveVersion(version, readBuildInfo())))
}

func readBuildInfo() *debug.BuildInfo {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}
	return info
}

func effectiveVersion(buildVersion string, info *debug.BuildInfo) string {
	if buildVersion != "" && buildVersion != "dev" {
		return buildVersion
	}
	if info != nil && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	if buildVersion != "" {
		return buildVersion
	}
	return "dev"
}
