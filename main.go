// Package main is the entry point for the ado CLI.
// The version variable is injected at build time via ldflags.
package main

import "github.com/gyurisc/adocli/cmd"

// version is set by goreleaser or Makefile via -ldflags "-X main.version=..."
var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
