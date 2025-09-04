package main

import (
	"fmt"
	"os"

	"github.com/naohman/whenistheq/client"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "whenIsTheQ",
		Usage:       "whenistheq nexttrain --addr localhost:8080 --station R17 --line Q --direction downtown",
		Description: "tells you when the next train is",
		Commands: []*cli.Command{
			{
				Name:   "nexttrain",
				Usage:  "Tells you when the next train is",
				Action: WhenIsTheQ,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "addr",
						Aliases: []string{"a"},
						Usage:   "address of the Transiter server's API",
						Value:   "localhost:8083",
					},
					&cli.StringFlag{
						Name:    "station",
						Aliases: []string{"s"},
						Usage:   "the station code to query",
						Value:   "",
					},
					&cli.StringFlag{
						Name:    "line",
						Aliases: []string{"l"},
						Usage:   "the subway line to query",
						Value:   "",
					},
					&cli.StringFlag{
						Name:    "direction",
						Aliases: []string{"d"},
						Usage:   "the direction of the train (uptown/downtown)",
						Value:   "",
					},
					&cli.StringFlag{
						Name:    "system",
						Aliases: []string{"sys"},
						Usage:   "the subway system to query",
						Value:   "us-ny-subway",
					},
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func WhenIsTheQ(c *cli.Context) error {
	client := client.NewClient(c.String("addr"), c.String("system"))
	time, err := client.GetNextDeparture(c.String("station"), c.String("line"), c.String("direction"))
	if err != nil {
		return err
	}
	fmt.Println(time.Format("03:04:05 PM"))
	return nil
}
