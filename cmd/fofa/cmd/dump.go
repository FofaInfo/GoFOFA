package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/LubyRuffy/gofofa"
	"github.com/LubyRuffy/gofofa/pkg/outformats"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"os"
	"strings"
)

// dump subcommand
var dumpCmd = &cli.Command{
	Name:                   "dump",
	Usage:                  "fofa dump data",
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "fields",
			Aliases:     []string{"f"},
			Value:       "host,ip,port",
			Usage:       "visit fofa website for more info",
			Destination: &fieldString,
		},
		&cli.StringFlag{
			Name:        "format",
			Value:       "csv",
			Usage:       "can be csv/json/xml",
			Destination: &format,
		},
		&cli.BoolFlag{
			Name:        "json",
			Aliases:     []string{"j"},
			Usage:       "output use json format",
			Destination: &json,
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
			Usage:       "queries line by line",
			Destination: &inFile,
		},
		&cli.IntFlag{
			Name:        "size",
			Aliases:     []string{"s"},
			Value:       -1,
			Usage:       "-1 means all",
			Destination: &size,
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
		&cli.IntFlag{
			Name:        "batchSize",
			Aliases:     []string{"bs"},
			Value:       1000,
			Usage:       "the amount of data contained in each batch",
			Destination: &batchSize,
		},
	},
	Action: DumpAction,
}

// DumpAction search action
func DumpAction(ctx *cli.Context) error {
	// valid same config
	var queries []string
	query := ctx.Args().First()
	if len(query) > 0 {
		queries = append(queries, query)
	}
	if len(inFile) > 0 {
		// 打开文件
		file, err := os.Open(inFile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		// 创建一个 Scanner 对象来逐行读取文件
		scanner := bufio.NewScanner(file)

		// 逐行读取并打印
		for scanner.Scan() {
			queries = append(queries, scanner.Text())
		}

		// 检查是否有读取错误
		if err = scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}

	if len(queries) == 0 {
		return errors.New("fofa query cannot be empty, use args or -inFile")
	}
	fields := strings.Split(fieldString, ",")
	if len(fields) == 0 {
		return errors.New("fofa fields cannot be empty")
	}

	// gen output
	var outTo io.Writer
	if len(outFile) > 0 {
		var f *os.File
		var err error
		if f, err = os.Create(outFile); err != nil {
			return fmt.Errorf("create outFile %s failed: %w", outFile, err)
		}
		outTo = f
		defer f.Close()
	} else {
		outTo = os.Stdout
	}

	if json {
		format = "json"
	}
	// gen writer
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
	for _, query := range queries {
		log.Println("dump data of query:", query)

		fetchedSize := 0
		err := fofaCli.DumpSearch(query, size, batchSize, fields, func(res [][]string, allSize int) (err error) {
			fetchedSize += len(res)
			log.Printf("size: %d/%d, %.2f%%", fetchedSize, allSize, 100*float32(fetchedSize)/float32(allSize))
			// output
			err = writer.WriteAll(res)
			return err
		}, gofofa.SearchOptions{
			FixUrl:    fixUrl,
			UrlPrefix: urlPrefix,
			Full:      full,
		})
		if err != nil {
			log.Println("fetch error:", err)
			//return err
		}
	}

	return nil
}
