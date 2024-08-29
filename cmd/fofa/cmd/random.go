package cmd

import (
	"errors"
	"fmt"
	"github.com/LubyRuffy/gofofa"
	"github.com/LubyRuffy/gofofa/pkg/outformats"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"math/rand"
	"os"
	"strings"
	"time"
)

var (
	sleepMS int
)

// random subcommand
var randomCmd = &cli.Command{
	Name:                   "random",
	Usage:                  "fofa random data generator",
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "fields",
			Aliases:     []string{"f"},
			Value:       "ip,port,host,header,title,server,lastupdatetime",
			Usage:       "visit fofa website for more info",
			Destination: &fieldString,
		},
		&cli.StringFlag{
			Name:        "format",
			Value:       "json",
			Usage:       "can be csv/json/xml",
			Destination: &format,
		},
		&cli.IntFlag{
			Name:        "size",
			Aliases:     []string{"s"},
			Value:       1,
			Usage:       "-1 means never stop",
			Destination: &size,
		},
		&cli.IntFlag{
			Name:        "sleep",
			Value:       1000,
			Usage:       "ms",
			Destination: &sleepMS,
		},
		&cli.BoolFlag{
			Name:        "fixUrl",
			Value:       false,
			Usage:       "each host fix as url, like 1.1.1.1,80 will change to http://1.1.1.1",
			Destination: &fixUrl,
		},
		&cli.StringFlag{
			Name:        "urlPrefix",
			Value:       "http://",
			Usage:       "prefix of url, default is http://, can be redis:// and so on ",
			Destination: &urlPrefix,
		},
		&cli.BoolFlag{
			Name:        "full",
			Value:       false,
			Usage:       "search result for over a year",
			Destination: &full,
		},
	},
	Action: randomAction,
}

// randomAction random action
func randomAction(ctx *cli.Context) error {
	// valid same config
	query := ctx.Args().First()
	if len(query) == 0 {
		query = "type=subdomain"
	}

	fields := strings.Split(fieldString, ",")
	if len(fields) == 0 {
		return errors.New("fofa fields cannot be empty")
	}
	hostIndex := -1
	if ctx.Bool("verbose") {
		if !hashField(fields, "host") {
			logrus.Warnln("verbose mode, so add host to fields automatically")
			fields = append(fields, "host")
		}
		hostIndex = fieldIndex(fields, "host")
	}

	// gen writer
	outTo := os.Stdout
	var writer outformats.OutWriter
	if hasBodyField(fields) && format == "csv" {
		logrus.Warnln("fields contains body, so change format to json")
		writer = outformats.NewJSONWriter(outTo, fields)
	} else {
		switch format {
		case "csv":
			writer = outformats.NewCSVWriter(outTo)
		case "json":
			writer = outformats.NewJSONWriter(outTo, fields)
		case "xml":
			writer = outformats.NewXMLWriter(outTo, fields)
		default:
			return fmt.Errorf("unknown format: %s", format)
		}
	}

	// do search
	for i := 0; i < size || size == -1; i++ {
		newQuery := query
		if !strings.HasPrefix(newQuery, "host=") && !strings.HasPrefix(query, "ip=") {
			max := time.Now()
			min := max.AddDate(-1, 0, 0)
			delta := max.Unix() - min.Unix()
			sec := rand.Int63n(delta) + min.Unix()
			ts := time.Unix(sec, 0).Format("2006-01-02 15:04:05")
			newQuery = newQuery + ` && before="` + ts + `"`
		}

		res, err := fofaCli.HostSearch(newQuery, 1, fields, gofofa.SearchOptions{
			FixUrl:    fixUrl,
			UrlPrefix: urlPrefix,
			Full:      full,
		})
		if err != nil {
			return err
		}

		if ctx.Bool("verbose") {
			logrus.Debugln("host:", res[0][hostIndex])
		}

		// output
		if err = writer.WriteAll(res); err != nil {
			return err
		}

		// 不是最后一次
		if i < size-1 {
			if sleepMS > 0 {
				time.Sleep(time.Duration(sleepMS))
			}
		}
	}

	return nil
}
