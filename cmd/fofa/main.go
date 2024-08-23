package main

import (
	"fmt"
	"github.com/LubyRuffy/gofofa/cmd/fofa/cmd"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

var (
	version = "v0.2.20"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown" // goreleaser fill

	defaultCommand = "search"
)

func main() {
	app := &cli.App{
		Name:                   "fofa",
		Usage:                  fmt.Sprintf("fofa client on Go %s, commit %s, built at %s", version, commit, date),
		Version:                version,
		UseShortOptionHandling: true,
		EnableBashCompletion:   true,
		Authors: []*cli.Author{
			{
				Name:  "LubyRuffy",
				Email: "lubyruffy@gmail.com",
			},
		},
		Flags:    cmd.GlobalOptions,
		Before:   cmd.BeforAction,
		Commands: cmd.GlobalCommands,
	}

	// default command
	if len(os.Args) > 1 && !cmd.IsValidCommand(os.Args[1]) {
		var newArgs []string
		newArgs = append(newArgs, os.Args[0])
		newArgs = append(newArgs, defaultCommand)
		newArgs = append(newArgs, os.Args[1:]...)
		os.Args = newArgs
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
