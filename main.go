package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/naohman/whenistheq/client"
	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:        "whenIsTheQ",
		Usage:       "whenistheq nexttrain --addr localhost:8080 --station R17 --line Q --direction downtown",
		Description: "tells you when the next train is",
		Commands: []*cli.Command{
			{
				Name:   "nexttrain",
				Usage:  "Tells you when the next train is",
				Action: WhenIsTheQ,
				After:  checkUpDown,
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
						Name:    "system",
						Aliases: []string{"sys"},
						Usage:   "the subway system to query",
						Value:   "us-ny-subway",
					},
					&cli.BoolFlag{
						Name:    "uptown",
						Aliases: []string{"u"},
						Usage:   "get the time of the next uptown train",
						Value:   false,
					},
					&cli.BoolFlag{
						Name:    "downtown",
						Aliases: []string{"d"},
						Usage:   "get the time of the next downtown train",
						Value:   false,
					},
					&cli.BoolFlag{
						Name:  "diff",
						Usage: "print the duration remaining to the next train instead of the absolute",
						Value: false,
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
	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func WhenIsTheQ(_ context.Context, c *cli.Command) error {
	tClient := client.NewClient(c.String("addr"), c.String("system"))
	direction := client.UPTOWN
	if c.Bool("downtown") {
		direction = client.DOWNTOWN
	}
	departure, err := tClient.GetNextDeparture(c.String("station"), c.String("line"), direction)
	if err != nil {
		return err
	}
	if c.Bool("diff") {
		duration := time.Until(departure)
		fmt.Printf("%02d:%02d\n", int(duration.Minutes()), int(duration.Seconds())%60)
	} else {
		fmt.Println(departure.Format("03:04:05 PM"))
	}
	return nil
}

func checkUpDown(_ context.Context, c *cli.Command) error {
	if c.Bool("uptown") == c.Bool("downtown") {
		return fmt.Errorf("must specify exactly one of -uptown or -downtown")
	}
	return nil
}
