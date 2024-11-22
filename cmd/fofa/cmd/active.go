package cmd

import (
	"errors"
	"fmt"
	"github.com/FofaInfo/GoFOFA"
	"github.com/FofaInfo/GoFOFA/pkg/outformats"
	"github.com/urfave/cli/v2"
	"io"
	"os"
	"sync"
)

var (
	activeTarget string // single active target
	retry        int    // timeout retry count
)

var activeCmd = &cli.Command{
	Name:  "active",
	Usage: "website active",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "url",
			Aliases:     []string{"u"},
			Value:       "",
			Usage:       "probe active for url",
			Destination: &activeTarget,
		},
		&cli.StringFlag{
			Name:        "outFile",
			Aliases:     []string{"o"},
			Usage:       "if not set, write to stdout",
			Destination: &outFile,
		},
		&cli.StringFlag{
			Name:        "inFile",
			Aliases:     []string{"i"},
			Usage:       "input file to build template if not use pipeline mode",
			Destination: &inFile,
		},
		&cli.StringFlag{
			Name:        "format",
			Value:       "csv",
			Usage:       "can be csv/json/xml",
			Destination: &format,
		},
		&cli.IntFlag{
			Name:        "workers",
			Value:       10,
			Usage:       "number of workers",
			Destination: &workers,
		},
		&cli.IntFlag{
			Name:        "retry",
			Value:       3,
			Usage:       "timeout retry count",
			Destination: &retry,
		},
	},
	Action: ActiveAction,
}

func ActiveAction(ctx *cli.Context) error {
	// valid same config
	for _, arg := range ctx.Args().Slice() {
		if arg[0] == '-' {
			return errors.New(fmt.Sprintln("there is args after fofa query:", arg))
		}
	}

	// gen output
	var outTo io.Writer
	if len(outFile) > 0 {
		f, err := os.Create(outFile)
		if err != nil {
			return fmt.Errorf("create outFile %s failed: %w", outFile, err)
		}
		outTo = f
		defer f.Close()
	} else {
		outTo = os.Stdout
	}

	// gen writer
	headFields := []string{"url", "isActive"}
	var writer outformats.OutWriter
	switch format {
	case "csv":
		writer = outformats.NewCSVWriter(outTo)
	case "json":
		writer = outformats.NewJSONWriter(outTo, headFields)
	case "xml":
		writer = outformats.NewXMLWriter(outTo, headFields)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}

	var locker sync.Mutex

	writeURL := func(u string) error {
		// do active
		resp := gofofa.DoHttpCheck(u, retry)
		res := [][]string{{u, fmt.Sprintf("%t", resp.IsActive)}}

		// output
		locker.Lock()
		defer locker.Unlock()
		if err := writer.WriteAll(res); err != nil {
			return err
		}
		writer.Flush()

		return nil
	}

	if len(activeTarget) != 0 {
		return writeURL(activeTarget)
	} else {
		var inf io.Reader
		if inFile != "" {
			f, err := os.Open(inFile)
			if err != nil {
				return err
			}
			defer f.Close()
			inf = f
		} else {
			inf = os.Stdin
		}
		concurrentPipeline(writeURL, inf)
	}

	return nil
}
