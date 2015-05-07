package main

import (
	"os"

	"github.com/codegangsta/cli"
)

const VERSION = "0.0.1"
const AUTHOR = "Erik Hollensbe <erik@hollensbe.org>"

var standardFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "host",
		Value: "",
		Usage: "The host to use when engaging with S3.",
	},
	cli.StringFlag{
		Name:  "region",
		Value: "",
		Usage: "The region to use when engaging with S3.",
	},
	cli.IntFlag{
		Name:  "concurrency",
		Value: 20,
		Usage: "The amount of parallel workers to use.",
	},
}

var commands = []cli.Command{
	cli.Command{
		Name:        "put",
		Usage:       "put [directory] [target s3 url]",
		Description: "Upload files to S3",
		Flags:       standardFlags,
		Action:      newput().putCommand,
	},
	cli.Command{
		Name:        "get",
		Usage:       "get [s3 url] [target directory]",
		Description: "Download files from S3",
		Flags:       standardFlags,
		Action:      newget().getCommand,
	},
}

func makeApp() *cli.App {
	app := cli.NewApp()

	app.Name = "s3util"
	app.Usage = "Control S3 with parallel transfers"
	app.Version = VERSION
	app.Author = AUTHOR
	app.Commands = commands

	return app
}

func main() {
	makeApp().Run(os.Args)
}
