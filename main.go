package main

import (
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli"
)

var (
	Config *tConfig //global config struct
)

func main() {
	// Setting up fancy CLI stuff.
	app := cli.NewApp()
	app.Name = "gOTP"
	app.HelpName = "gotp"
	app.Usage = "Generate OTP for your hipster tool"
	app.UsageText = "gotp [global options] command [command option]"
	app.Version = "0.0.1"
	app.Compiled = time.Now()
	app.Copyright = "(c) 2016 Magnus Kaiser & Contributers\n\n   MIT License (see https://opensource.org/licenses/MIT)"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Magnus Kaiser",
			Email: "xoc@4xoc.com",
		},
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "Return debug information",
		},
	}

	// Always reading & checking config first.
	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			checkConfiguration(true)
		} else {
			checkConfiguration(false)
		}

		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:        "add",
			Aliases:     []string{"a"},
			Description: "This will start a new interactive session to add a new OTP configuration if *no* options are set.",
			Usage:       "Add a new OPT configuration",
			Action:      AddOTP,
			ArgsUsage:   " ",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "uri, u",
					Usage: "(required) URI to parse that holds the OTP information. MUST be wrappen in \"",
				},
				cli.StringFlag{
					Name:  "label, l",
					Usage: "(required if not within URI) Set a label (name) to identify the OTP",
				},
				cli.StringFlag{
					Name:  "description, d",
					Usage: "(optional) Set a description for this OTP",
				},
			},
		},
		{
			Name:        "get",
			Aliases:     []string{"g"},
			Description: "Returns the current OTP based on the name given. The name must match an existing configuration entry. Empty string is returned when the name doesn't match any known configuration.",
			Usage:       "Get an OTP for an existing OTP integration",
			Action:      GetOTP,
			ArgsUsage:   "[name]",
		},
		{
			Name:        "list",
			Aliases:     []string{"l"},
			Description: "Returns a list of all configured OTPs.",
			Usage:       "Get a list of all integrated OTPs",
			Action:      ListOTP,
			ArgsUsage:   "[name]",
		},
		{
			Name:        "delete",
			Aliases:     []string{"d"},
			Description: "Deletes the configuration identified by name. Returns an error message when no configuration with the given name exists. Use -f to force deletion without being asked for confirmation.",
			Usage:       "Delete an existing OTP configuration",
			Action:      DeleteOTP,
			ArgsUsage:   "[name]",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "Don't ask for confirmation"},
			},
		},
	}

	app.CommandNotFound = func(c *cli.Context, command string) {
		fmt.Fprintf(c.App.Writer, "Unknown command '%s'. Use gotp help for further details.\n", command)
	}

	// No matter what arguments have been passed on, on initial setup
	// we do not do more than just setting up the environment.
	app.Run(os.Args)
}
