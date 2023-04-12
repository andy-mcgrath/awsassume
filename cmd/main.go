package main

import (
	"github.com/andy-mcgrath/awsassume/cmd/assume"
	"github.com/andy-mcgrath/awsassume/config"
	"github.com/mitchellh/cli"
	"log"
	"os"
)

const version = "v0.0.1"

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	c := cli.NewCLI("awsassume", version)
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"assume": func() (cli.Command, error) {
			return &assume.Command{
				Ui: &cli.ColoredUi{
					Ui: &cli.BasicUi{
						Writer:      os.Stdout,
						ErrorWriter: os.Stderr,
					},
					OutputColor: cli.UiColorGreen,
					InfoColor:   cli.UiColorYellow,
					ErrorColor:  cli.UiColorRed,
				},
				Config: cfg,
			}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Fatalf("Error executing CLI: %s", err)
	}

	os.Exit(exitStatus)
}
