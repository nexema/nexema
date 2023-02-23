package cmd

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
	"tomasweigenast.com/nexema/tool/nexema"
)

const helpText = `Nexema - binary interchange made simple

Usage:
   nexema command (arguments...)
{{if .Commands}}
Available commands:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
{{end}}
About:
	Made by Tomás Weigenast <tomaswegenast@gmail.com>
	v1.0.0
	Licensed under GPL-3.0
`

var app *cli.App

func init() {
	app = &cli.App{
		CustomAppHelpTemplate: helpText,
		CommandNotFound:       cli.ShowCommandCompletions,
	}

	app.Commands = []cli.Command{
		{
			Name:  "mod",
			Usage: "Manages Nexema projects",
			Subcommands: []cli.Command{
				{
					Name:      "init",
					Usage:     "Initializes a new project",
					ArgsUsage: "[the path where to initialize the project]",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:     "overwrite",
							Usage:    "Overwrites any previous existing nexema.yaml at specified at",
							Required: false,
						},
					},
					Action: func(c *cli.Context) error {
						path := c.Args().First()
						if len(path) == 0 {
							return cli.NewExitError("path is required", 1)
						}

						return modInit(path, c.Bool("overwrite"))
					},
				},
			},
		},
		{
			Name:  "build",
			Usage: "Builds a project and optionally outputs a snapshot file",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "out",
					Usage: "The path to the output folder where to write the snapshot file",
				},
			},
			Action: func(c *cli.Context) error {
				path := c.Args().First()
				if len(path) == 0 {
					return cli.NewExitError("path is required", 1)
				}

				return buildCmd(path, c.String("out"))
			},
		},
		{
			Name:  "generate",
			Usage: "Builds a project and generates source code",

			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "snapshot-file",
					Usage: "generate from a snapshot file",
				},
				cli.StringSliceFlag{
					Required: true,
					Name:     "for",
					Usage:    "the generators to use and their output path",
				},
			},
			Action: func(c *cli.Context) error {

				path := c.Args().First()
				snapshotPath := c.String("snapshot-file")
				if len(path) == 0 {
					return cli.NewExitError("path is required", 1)
				}

				generateFor := c.StringSlice("for")

				return generateCmd(path, snapshotPath, generateFor)
			},
		},
		{
			Name:  "format",
			Usage: "Format all .nex files in the specified project",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "path",
					Usage: "Path to the project directory",
				},
			},
			Action: func(c *cli.Context) error {
				path := c.String("path")
				if path == "" {
					return cli.NewExitError("path is required", 1)
				}
				fmt.Printf("Formatting code for the project at %s...\n", path)
				return nil
			},
		},
	}
}

func Execute() {
	nexema.Run()

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}