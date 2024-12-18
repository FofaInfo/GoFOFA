package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/FofaInfo/GoFOFA"
	"github.com/FofaInfo/GoFOFA/pkg/outformats"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

var (
	browserURL  string
	browserTags string
)

var browserCmd = &cli.Command{
	Name:  "jsRender",
	Usage: "website js render",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "url",
			Aliases:     []string{"u"},
			Value:       "",
			Usage:       "parsing the url through the browser",
			Destination: &browserURL,
		},
		&cli.StringFlag{
			Name:        "tags",
			Aliases:     []string{"t"},
			Value:       "title",
			Usage:       "print tag content, can be title/body",
			Destination: &browserTags,
		},
		&cli.StringFlag{
			Name:        "outFile",
			Aliases:     []string{"o"},
			Usage:       "if not set, write to stdin",
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
			Value:       2,
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
	Action: BrowserAction,
}

func mapToSliceOne(data map[string]interface{}, keys []string) [][]string {
	// 将一个 map[string]interface{} 变成一行 [][]string
	var values []string
	for _, key := range keys {
		if value, exists := data[key]; exists {
			values = append(values, fmt.Sprintf("%v", value))
		}
	}
	return [][]string{values}
}

func concurrentPipeline(writeData func(url string) error, in io.Reader) {
	// 并发模式
	wg := sync.WaitGroup{}
	urls := make(chan string, workers)

	worker := func(urls <-chan string, wg *sync.WaitGroup) {
		for u := range urls {
			if err := writeData(u); err != nil {
				log.Println("[WARNING]", err)
			}
			wg.Done()
		}
	}
	for w := 0; w < workers; w++ {
		go worker(urls, &wg)
	}

	scanner := bufio.NewScanner(in)
	for scanner.Scan() { // internally, it advances token based on sperator
		line := scanner.Text()
		wg.Add(1)
		urls <- line
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

	wg.Wait()
}

func BrowserAction(ctx *cli.Context) error {
	if len(ctx.Args().Slice()) > 0 {
		return errors.New("please use -h to view usage")
	}

	tags := strings.Split(browserTags, ",")

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

	headFields := []string{"url"}
	headFields = append(headFields, tags...)
	var writer outformats.OutWriter
	if hasBodyField(tags) && format == "csv" {
		logrus.Warnln("fields contains body, so change format to json")
		writer = outformats.NewJSONWriter(outTo, headFields)
	} else {
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
	}

	var locker sync.Mutex

	writeURL := func(u string) error {
		log.Println("js render of:", u)
		// do jsBrowser
		b := gofofa.NewWorkerBrowser(u, retry)
		body, err := b.Run()
		if err != nil {
			return err
		}
		if body == nil {
			return nil
		}

		// output
		locker.Lock()
		defer locker.Unlock()
		if err = writer.WriteAll(mapToSliceOne(body, headFields)); err != nil {
			return err
		}
		writer.Flush()
		return nil
	}

	if browserURL != "" {
		return writeURL(browserURL)
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
