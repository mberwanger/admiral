package main

import (
	_ "embed"
	"os"

	"github.com/mberwanger/admiral/cli/cmd"
	"github.com/mberwanger/admiral/server/version"
)

//go:embed art.txt
var asciiArt string

func main() {
	cmd.Execute(
		buildInfo(),
		os.Exit,
		os.Args[1:],
	)
}

func buildInfo() version.Info {
	return version.GetVersionInfo(
		version.WithAppDetails("Admiral", "Platform Orchestrator that helps developers build, deploy, and manage their applications more quickly and easily", "https://admiral.io"),
		version.WithASCIIName(asciiArt),
	)
}
