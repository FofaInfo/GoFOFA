package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/LubyRuffy/gofofa"
	"github.com/LubyRuffy/gofofa/pkg/outformats"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"os"
	"strings"
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
			Name:        "target",
			Aliases:     []string{"t"},
			Value:       "",
			Usage:       "probe active for targets",
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
	var writer outformats.OutWriter
	writer = outformats.NewCSVWriter(outTo)

	var locker sync.Mutex

	writeLink := func(links []string) error {
		var res [][]string

		for _, l := range links {
			var result []string
			resp := gofofa.DoHttpCheck(l, retry)
			result = append(result, l, fmt.Sprintf("%t", resp.IsActive))
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
