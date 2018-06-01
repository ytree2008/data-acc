package main

import (
	"github.com/RSE-Cambridge/data-acc/pkg/version"
	"github.com/urfave/cli"
	"log"
	"os"
)

func stripFunctionArg(systemArgs []string) []string {
	if len(systemArgs) > 2 && systemArgs[1] == "--function" {
		return append(systemArgs[0:1], systemArgs[2:]...)
	}
	return systemArgs
}

var token = cli.StringFlag{
	Name:  "token, t",
	Usage: "Job ID or Persistent Buffer name",
}
var job = cli.StringFlag{
	Name:  "job",
	Usage: "Path to burst buffer request file.",
}
var caller = cli.StringFlag{
	Name:  "caller",
	Usage: "The system that called the CLI, e.g. Slurm.",
}
var user = cli.IntFlag{
	Name:  "user",
	Usage: "Linux user id that owns the buffer.",
}
var groupid = cli.IntFlag{
	Name:  "groupid",
	Usage: "Linux group id that owns the buffer, defaults to match the user.",
}
var capacity = cli.StringFlag{
	Name:  "capacity",
	Usage: "A request of the form <pool>:<int><units> where units could be GiB or TiB.",
}

func runCli(args []string) error {
	app := cli.NewApp()
	app.Name = "FakeWarp CLI"
	app.Usage = "This CLI is used to integrate data-acc with Slurm's Burst Buffer plugin."
	app.Version = version.VERSION

	app.Commands = []cli.Command{
		{
			Name:   "pools",
			Usage:  "List all the buffer pools",
			Action: listPools,
		},
		{
			Name:   "show_instances",
			Usage:  "List the buffer instances.",
			Action: showInstances,
		},
		{
			Name:   "show_sessions",
			Usage:  "List the buffer sessions.",
			Action: showSessions,
		},
		{
			Name:  "teardown",
			Usage: "Destroy the given buffer.",
			Flags: []cli.Flag{token, job,
				cli.BoolFlag{
					Name: "hurry",
				},
			},
			Action: teardown,
		},
		{
			Name:   "job_process",
			Usage:  "Initial call to validate buffer script",
			Flags:  []cli.Flag{job},
			Action: jobProcess,
		},
		{
			Name:   "setup",
			Usage:  "Create transient burst buffer, called after waiting for enough free capacity.",
			Flags:  []cli.Flag{token, job, caller, user, groupid, capacity},
			Action: setup,
		},
		{
			Name:  "real_size",
			Usage: "Report actual size of created buffer.",
		},
		{
			Name:  "data_in",
			Usage: "Copy data into given buffer.",
		},
		{
			Name:  "paths",
			Usage: "Environment variables describing where the buffer will be mounted.",
		},
		{
			Name:  "pre_run",
			Usage: "Attach given buffers to compute nodes specified.",
		},
		{
			Name:  "post_run",
			Usage: "Detach buffers before releasing compute nodes.",
		},
		{
			Name:  "data_out",
			Usage: "Copy data out of buffer.",
		},
		{
			Name:  "create_persistent",
			Usage: "Create a persistent buffer.",
		},
		{
			Name:   "show_configurations",
			Usage:  "Returns fake data to keep burst buffer plugin happy.",
			Action: showConfigurations,
		},
	}

	return app.Run(stripFunctionArg(args))
}

func main() {
	if err := runCli(os.Args); err != nil {
		log.Fatal(err)
	}
}