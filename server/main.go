package main

import (
	_ "embed"
	"os"

	"go.admiral.io/admiral/server/cmd/app"
	"go.admiral.io/admiral/server/version"
)

//go:embed art.txt
var asciiArt string

func main() {
	app.Execute(
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
