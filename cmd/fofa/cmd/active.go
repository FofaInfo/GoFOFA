package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"git.gobies.org/goby/httpclient"
	"github.com/LubyRuffy/gofofa/pkg/outformats"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

var activeTarget string // single active target

var activeCmd = &cli.Command{
	Name:  "active",
	Usage: "website active",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "target",
			Aliases:     []string{"t"},
			Value:       "",
			Usage:       "probe active for targets",
			Destination: &activeTarget,
		},
		&cli.StringFlag{
			Name:        "format",
			Value:       "csv",
			Usage:       "can be csv/json/xml",
			Destination: &format,
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
		&cli.IntFlag{
			Name:        "workers",
			Value:       10,
			Usage:       "number of workers",
			Destination: &workers,
		},
	},
	Action: ActiveAction,
}

func checkActive(t string) bool {
	fURL := httpclient.NewFixUrl(t)
	cfg := httpclient.NewGetRequestConfig("/")
	_, err := httpclient.DoHttpRequest(fURL, cfg)
	if err != nil {
		return false
	}
	return true
}

func pipelineLink(writeLink func(links []string) error, in io.Reader) {
	// 并发模式
	wg := sync.WaitGroup{}
	links := make(chan string, workers)

	worker := func(links <-chan string, wg *sync.WaitGroup) {
		for l := range links {
			tmpLink := []string{l}
			if err := writeLink(tmpLink); err != nil {
				log.Println("[WARNING]", err)
			}
			wg.Done()
		}
	}
	for w := 0; w < workers; w++ {
		go worker(links, &wg)
	}

	scanner := bufio.NewScanner(in)
	for scanner.Scan() { // internally, it advances token based on sperator
		line := scanner.Text()
		wg.Add(1)
		links <- line
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

	wg.Wait()
}

func ActiveAction(ctx *cli.Context) error {
	// valid same config
	for _, arg := range ctx.Args().Slice() {
		if arg[0] == '-' {
			return errors.New(fmt.Sprintln("there is args after fofa query:", arg))
		}
	}

	var targets []string
	if len(activeTarget) > 0 {
		targets = strings.Split(activeTarget, ",")
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

	// gen writer
	var writer outformats.OutWriter
	switch format {
	case "csv":
		writer = outformats.NewCSVWriter(outTo)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}

	var locker sync.Mutex

	writeLink := func(links []string) error {
		var res [][]string

		for _, l := range links {
			var result []string
			isActive := checkActive(l)
			result = append(result, l, fmt.Sprintf("%t", isActive))
			res = append(res, result)
		}

		locker.Lock()
		defer locker.Unlock()
		if err := writer.WriteAll(res); err != nil {
			return err
		}
		writer.Flush()

		return nil
	}

	if len(targets) != 0 {
		return writeLink(targets)
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
		pipelineLink(writeLink, inf)
	}

	return nil
}
