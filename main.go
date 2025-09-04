package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/naohman/whenistheq/client"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:        "whenIsTheQ",
		Description: "tells you when the next train is",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "addr",
				Aliases: []string{"a"},
				Usage:   "address of the Transiter server's API",
				Value:   "http://localhost:8080",
			},
			&cli.StringFlag{
				Name:    "system",
				Aliases: []string{"s"},
				Usage:   "the subway system to query",
				Value:   "us-ny-subway",
			},
		},
		Commands: []*cli.Command{
			{
				Name:        "station_lookup",
				Description: "Looks up the id of a station",
				Usage:       "whenistheq station_lookup Broadway Junction",
				Action:      StationLookup,
			},
			{
				Name:        "next_train",
				Usage:       "whenistheq next_train --station R17 --line Q --downtown",
				Description: "Tells you when the next train is",
				Action:      NextTrain,
				After:       checkUpDown,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "station",
						Aliases:  []string{"S"},
						Usage:    "the station code to query",
						Value:    "",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "line",
						Aliases:  []string{"l"},
						Usage:    "the subway line to query",
						Value:    "",
						Required: true,
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
				},
			},
		},
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func NextTrain(_ context.Context, c *cli.Command) error {
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

func StationLookup(_ context.Context, c *cli.Command) error {
	tClient := client.NewClient(c.String("addr"), c.String("system"))
	key := strings.Join(c.Args().Slice(), " ")
	stations, err := tClient.FindStationCode(key)
	if err != nil {
		return err
	}
	if len(stations) == 0 {
		return fmt.Errorf("no stations found matching %s", key)
	}
	t := table.New("ID", "Name", "Lines")
	for _, s := range stations {
		t.AddRow(s.Id, s.Name, strings.Join(s.Lines(), ", "))
	}
	t.Print()
	return nil
}

func checkUpDown(_ context.Context, c *cli.Command) error {
	if c.Bool("uptown") == c.Bool("downtown") {
		return fmt.Errorf("must specify exactly one of -uptown or -downtown")
	}
	return nil
}
