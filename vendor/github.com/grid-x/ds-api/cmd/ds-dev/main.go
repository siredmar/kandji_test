package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "ds-dev"
	app.Usage = "Dev productivity booster"

	app.Commands = []cli.Command{
		{
			Name:  "api",
			Usage: "API related tasks",
			Subcommands: []cli.Command{
				{
					Name:  "new",
					Usage: "Create a new API version",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "version, v",
							Value: time.Now().Format("2006-01-02"),
							Usage: "API version to use",
						},
						cli.StringFlag{
							Name:  "type, t",
							Value: "",
							Usage: "Management (mgmt) or device (dev) API",
						},
						cli.StringFlag{
							Name:  "handler-filename",
							Value: "handler.go",
							Usage: "Overwrite the default handler filename",
						},
						cli.StringFlag{
							Name:  "service-filename",
							Value: "service.go",
							Usage: "Overwrite the default service filename",
						},
						cli.StringFlag{
							Name:  "handlers",
							Value: "",
							Usage: "The handlers to create comma separated of the form <group>:<handler>:<method>. Method is optional and is GET by default.",
						},
						cli.StringFlag{
							Name:  "base-path",
							Value: "api",
							Usage: "Overwrite the value of the API basepath",
						},
					},
					Action: func(ctx *cli.Context) error {
						if ctx.IsSet("handlers") {
							handlers := ctx.String("handlers")
							version := ctx.String("version")

							handlerFileName := ctx.String("handler-filename")
							serviceFileName := ctx.String("service-filename")

							var apiPath string
							if ctx.IsSet("type") {
								t := strings.ToLower(ctx.String("type"))
								switch t {
								case "management", "mgmt":
									apiPath = "management"
								case "device", "dev":
									apiPath = "device"
								default:
									return fmt.Errorf("invalid API type %s", t)
								}

							} else {
								return fmt.Errorf("API type not given. Expecting 'management' or 'device'.")
							}

							basePath := filepath.Join(ctx.String("base-path"), apiPath)

							apiGen, err := NewAPIGenerator(basePath, handlerFileName, serviceFileName, version, handlers)
							if err != nil {
								return err
							}

							dirs, gofiles := apiGen.GetGenFiles()

							for _, d := range dirs {
								os.MkdirAll(d, 0700)
							}

							return writeGenFiles(gofiles)
						} else {
							return fmt.Errorf("no handlers given")
						}
					},
				},
			},
		},
		{
			Name:  "repository",
			Usage: "Generate repositories",
			Subcommands: []cli.Command{
				{
					Name:  "new",
					Usage: "Generate a new repository",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "plural",
							Value: "",
						},
						cli.StringFlag{
							Name:  "singular",
							Value: "",
						},
						cli.StringFlag{
							Name:  "attributes",
							Value: "",
						},
						cli.StringFlag{
							Name:  "repository.base-path",
							Value: "pkg/postgres",
							Usage: "",
						},
						cli.StringFlag{
							Name:  "model.base-path",
							Value: "pkg/model",
							Usage: "",
						},
						cli.BoolFlag{
							Name: "no-migration",
						},
						cli.BoolFlag{
							Name: "no-model",
						},
						cli.BoolFlag{
							Name: "no-tests",
						},
					},
					Action: func(ctx *cli.Context) error {
						if !ctx.IsSet("plural") {
							return fmt.Errorf("plural name not set")
						}
						if !ctx.IsSet("singular") {
							return fmt.Errorf("singular name not set")
						}
						attributes := ctx.String("attributes")
						plural := ctx.String("plural")

						rg, err := NewRepositoryGenerator(
							ctx.String("repository.base-path"),
							ctx.String("model.base-path"),
							plural,
							ctx.String("singular"),
							strings.ToLower(plural)+".go",
							attributes,
							!ctx.Bool("no-migration"),
							!ctx.Bool("no-model"),
							!ctx.Bool("no-tests"),
						)
						if err != nil {
							return err
						}
						dirs, gofiles := rg.GetGenFiles()
						for _, d := range dirs {
							os.MkdirAll(d, 0700)
						}

						return writeGenFiles(gofiles)
					},
				},
			},
		},
		{
			Name:  "migration",
			Usage: "Generate migrations",
			Subcommands: []cli.Command{
				{
					Name:  "new",
					Usage: "Generate a new migration file",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name",
							Value: "",
						},
						cli.StringFlag{
							Name:  "columns",
							Value: "",
						},
						cli.StringFlag{
							Name:  "repository.base-path",
							Value: "pkg/postgres",
							Usage: "",
						},
					},
					Action: func(ctx *cli.Context) error {
						if !ctx.IsSet("name") {
							return fmt.Errorf("name not set")
						}
						attributes := ctx.String("columns")
						name := ctx.String("name")

						rg, err := NewMigrationGenerator(
							ctx.String("repository.base-path"),
							name,
							attributes,
						)
						if err != nil {
							return err
						}
						dirs, files := rg.GetGenFiles()
						for _, d := range dirs {
							os.MkdirAll(d, 0700)
						}

						return writeGenFiles(files)
					},
				},
			},
		},
		{
			Name:  "routes",
			Usage: "Generate routes",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "type, t",
					Value: "",
					Usage: "Management (mgmt) or device (dev) API",
				},
				cli.StringFlag{
					Name:  "base-path",
					Value: "api",
					Usage: "Overwrite the value of the API basepath",
				},
				cli.StringFlag{
					Name:  "since, s",
					Value: "",
					Usage: "Generate routes since version",
				},
				cli.StringFlag{
					Name:  "output, o",
					Value: "",
					Usage: "Overwrite the output filename",
				},
			},
			Action: func(ctx *cli.Context) error {
				var subRoute string
				if ctx.IsSet("type") {
					t := strings.ToLower(ctx.String("type"))
					switch t {
					case "management", "mgmt":
						subRoute = "management"
					case "device", "dev":
						subRoute = "device"
					default:
						return fmt.Errorf("invalid API type %s", t)
					}

				} else {
					return fmt.Errorf("API type not given. Expecting 'management' or 'device'.")
				}

				if ctx.IsSet("since") {
					_, err := time.Parse("2006-01-02", ctx.String("since"))
					if err != nil {
						return fmt.Errorf("since has not format YYYY-MM-DD")
					}
				}
				basePath := filepath.Join(ctx.String("base-path"), subRoute)

				routes, err := NewRoutesGenerator(basePath, subRoute, ctx.String("since"), ctx.String("output"))
				if err != nil {
					return err
				}

				return writeGenFiles(routes.GetGenFiles())
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
