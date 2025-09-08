package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"strings"
	"time"

	"github.com/golang/freetype/truetype"
	"github.com/naohman/whenistheq/client"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v3"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
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
				Usage:       "whenistheq next_train --station R17 --line Q --direction downtown",
				Description: "Tells you when the next train is",
				Action:      NextTrain,
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
					&cli.StringFlag{
						Name:    "destination",
						Aliases: []string{"D"},
						Usage:   "The destination of the line to query",
						Value:   "",
					},
					&cli.StringFlag{
						Name:    "direction",
						Aliases: []string{"d"},
						Usage:   "The direction of the train (Manhattan, Outbound, Uptown, Downtown, etc)",
						Value:   "",
					},
					&cli.BoolFlag{
						Name:  "diff",
						Usage: "print the duration remaining to the next train instead of the absolute",
						Value: false,
					},
				},
			},
			{
				Name:        "icon",
				Description: "Print a PNG icon for an MTA subway line",
				Usage:       "whenistheq icon --line Q --output q.png",
				Action:      Icon,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "line",
						Aliases:  []string{"l"},
						Usage:    "the subway line to query",
						Value:    "",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "output",
						Aliases:  []string{"o"},
						Usage:    "the file to write the icon to",
						Value:    "",
						Required: true,
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
	lineSelector, err := makeLineSelector(tClient, c)
	if err != nil {
		return err
	}
	departure, err := tClient.GetNextDeparture(c.String("station"), lineSelector)
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

func makeLineSelector(tClient *client.Client, c *cli.Command) (*client.LineSelector, error) {
	line := c.String("line")
	if line == "" {
		return nil, fmt.Errorf("--line is required")
	}
	selector := &client.LineSelector{
		Line: line,
	}
	if (c.String("destination") == "") == (c.String("direction") == "") {
		return nil, fmt.Errorf("must set exactly one of --direction, --destination")
	}
	if dir := c.String("direction"); dir != "" {
		selector.Direction = &client.HeadsignMatcher{Headsign: dir}
	}
	if dest := c.String("destination"); dest != "" {
		stop, err := tClient.GetStop(dest)
		if err != nil {
			return nil, err
		}
		selector.Direction = client.NewStationMatcher(stop)
	}
	return selector, nil
}

func Icon(_ context.Context, c *cli.Command) error {
	tClient := client.NewClient(c.String("addr"), c.String("system"))
	route, err := tClient.GetLine(c.String("line"))
	if err != nil {
		return err
	}
	size := 64
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	radius := size / 2
	drawCircle(img, radius, radius, radius, route.Color())

	goFont, err := truetype.Parse(goregular.TTF)
	if err != nil {
		panic(err)
	}
	drawer := &font.Drawer{
		Dst: img,
		Src: image.NewUniform(color.White),
		Face: truetype.NewFace(goFont, &truetype.Options{
			Size:    40,
			DPI:     72,
			Hinting: font.HintingFull,
		}),
	}
	width := drawer.MeasureString(route.Id)
	wordStart := radius - (width.Ceil() / 2)
	drawer.Dot = fixed.Point26_6{
		X: fixed.I(wordStart),
		Y: fixed.I(44),
	}
	drawer.DrawString(route.Id)

	f, err := os.OpenFile(c.String("output"), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	png.Encode(f, img)
	return nil
}

func drawCircle(img draw.Image, x0, y0, r int, c color.RGBA) {
	for x := x0 - r; x <= x0+r; x++ {
		dx2 := (x - x0) * (x - x0)
		for y := y0 - r; y <= y0+r; y++ {
			dy2 := (y - y0) * (y - y0)
			dist := math.Sqrt(float64(dx2 + dy2))

			if int(dist) < r {
				img.Set(x, y, c)
			}
		}
	}
}
