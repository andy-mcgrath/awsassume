package main

import (
	"github.com/mitchellh/cli"
	"log"
	"os"
)

const version = "v0.0.1"

func main() {
	c := cli.NewCLI("awsassume", version)
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"assume": func() (cli.Command, error) {
			return &AssumeCommand{
				Ui: &cli.ColoredUi{
					Ui: &cli.BasicUi{
						Writer:      os.Stdout,
						ErrorWriter: os.Stderr,
					},
					OutputColor: cli.UiColorGreen,
					InfoColor:   cli.UiColorYellow,
					ErrorColor:  cli.UiColorRed,
				},
			}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Fatalf("Error executing CLI: %s", err)
	}

	os.Exit(exitStatus)
}
