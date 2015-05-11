package main

import (
	"os"
	"runtime"

	"github.com/codegangsta/cli"

	"github.com/erikh/s3util/get"
	"github.com/erikh/s3util/put"
)

const VERSION = "0.0.1"
const AUTHOR = "Erik Hollensbe <erik@hollensbe.org>"

var standardFlags = []cli.Flag{
	cli.StringFlag{
		Name:   "host",
		Value:  "",
		Usage:  "The host to use when engaging with S3.",
		EnvVar: "S3_HOST",
	},
	cli.StringFlag{
		Name:   "region",
		Value:  "",
		Usage:  "The region to use when engaging with S3.",
		EnvVar: "S3_REGION",
	},
	cli.IntFlag{
		Name:   "concurrency",
		Value:  20,
		Usage:  "The amount of parallel workers to use.",
		EnvVar: "S3_CONCURRENCY",
	},
	cli.StringFlag{
		Name:   "access-key",
		Value:  "",
		Usage:  "The AWS access key to use.",
		EnvVar: "AWS_ACCESS_KEY_ID",
	},
	cli.StringFlag{
		Name:   "secret-key",
		Value:  "",
		Usage:  "The AWS secret key to use.",
		EnvVar: "AWS_SECRET_ACCESS_KEY",
	},
}

var commands = []cli.Command{
	cli.Command{
		Name:        "put",
		Usage:       "put [directory] [target s3 url]",
		Description: "Upload files to S3",
		Flags:       standardFlags,
		Action:      put.NewPut().PutCommand,
	},
	cli.Command{
		Name:        "get",
		Usage:       "get [s3 url] [target directory]",
		Description: "Download files from S3",
		Flags:       standardFlags,
		Action:      get.NewGet().GetCommand,
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
	runtime.GOMAXPROCS(runtime.NumCPU())
	makeApp().Run(os.Args)
}
