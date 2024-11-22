package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/LubyRuffy/gofofa"
	"github.com/LubyRuffy/gofofa/pkg/outformats"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"golang.org/x/time/rate"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

var (
	fieldString   string // fieldString
	size          int    // fetch size
	format        string // out format
	outFile       string // out file
	inFile        string // in file
	deductMode    string // deduct Mode
	fixUrl        bool   // each host fix as url, like 1.1.1.1,80 will change to http://1.1.1.1
	urlPrefix     string // each host fix as url, like 1.1.1.1,80 will change to http://1.1.1.1
	full          bool   // search result for over a year
	batchSize     int    // amount of data contained in each batch, only for dump
	json          bool   // out format as json for short
	uniqByIP      bool   // group by ip
	workers       int    // number of workers
	ratePerSecond int    // fofa request per second
	template      string // template in pipeline mode
	checkActive   int    // probe website is existed, add isActive field
	deWildcard    int    // number of wildcard domains retained
	filter        string // filter data by rules
	dedupHost     bool   // deduplicate by host
	headline      bool   // add headline for csv
	batchType     string // batch query, can be ip/domain
)

const (
	IPMax     = 100 // maximum number of splicing ips
	DomainMax = 50  // maximum number of splicing domains
)

// search subcommand
var searchCmd = &cli.Command{
	Name:                   "search",
	Usage:                  "fofa host search",
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "fields",
			Aliases:     []string{"f"},
			Value:       "ip,port",
			Usage:       "visit fofa website for more info",
			Destination: &fieldString,
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
		&cli.IntFlag{
			Name:        "size",
			Aliases:     []string{"s"},
			Value:       100,
			Usage:       "if DeductModeFree set, select free limit size automatically",
			Destination: &size,
		},
		&cli.StringFlag{
			Name:        "deductMode",
			Value:       "DeductModeFree",
			Usage:       "DeductModeFree or DeductModeFCoin",
			Destination: &deductMode,
		},
		&cli.BoolFlag{
			Name:        "fixUrl",
			Value:       false,
			Usage:       "each host fix as url, like 1.1.1.1,80 will change to http://1.1.1.1",
			Destination: &fixUrl,
		},
		&cli.StringFlag{
			Name:        "urlPrefix",
			Value:       "",
			Usage:       "prefix of url, default is http://, can be redis:// and so on ",
			Destination: &urlPrefix,
		},
		&cli.BoolFlag{
			Name:        "full",
			Value:       false,
			Usage:       "search result for over a year",
			Destination: &full,
		},
		&cli.BoolFlag{
			Name:        "uniqByIP",
			Value:       false,
			Usage:       "uniq by ip",
			Destination: &uniqByIP,
		},
		&cli.IntFlag{
			Name:        "workers",
			Value:       10,
			Usage:       "number of workers",
			Destination: &workers,
		},
		&cli.IntFlag{
			Name:        "rate",
			Value:       2,
			Usage:       "fofa query per second",
			Destination: &ratePerSecond,
		},
		&cli.StringFlag{
			Name:        "template",
			Value:       "ip={}",
			Usage:       "template in pipeline mode",
			Destination: &template,
		},
		&cli.StringFlag{
			Name:        "inFile",
			Aliases:     []string{"i"},
			Usage:       "input file to build template if not use pipeline mode",
			Destination: &inFile,
		},
		&cli.IntFlag{
			Name:        "checkActive",
			Value:       -1,
			Usage:       "probe website is existed, add isActive field",
			Destination: &checkActive,
		},
		&cli.IntFlag{
			Name:        "deWildcard",
			Value:       -1,
			Usage:       "number of wildcard domains retained",
			Destination: &deWildcard,
		},
		&cli.StringFlag{
			Name:        "filter",
			Value:       "",
			Usage:       "filter data by rules",
			Destination: &filter,
		},
		&cli.BoolFlag{
			Name:        "dedupHost",
			Value:       false,
			Usage:       "deduplicate by host",
			Destination: &dedupHost,
		},
		&cli.BoolFlag{
			Name:        "headline",
			Value:       false,
			Usage:       "add headline for csv",
			Destination: &headline,
		},
		&cli.StringFlag{
			Name:        "batchType",
			Value:       "",
			Usage:       "batch query, can be ip/domain",
			Destination: &batchType,
		},
	},
	Action: SearchAction,
}

func fieldIndex(fields []string, fieldName string) int {
	for i, f := range fields {
		if f == fieldName {
			return i
		}
	}
	return -1
}

func hashField(fields []string, fieldName string) bool {
	for _, f := range fields {
		if f == fieldName {
			return true
		}
	}
	return false
}

func hasBodyField(fields []string) bool {
	return hashField(fields, "body")
}

func pipelineProcess(writeQuery func(query string) error, in io.Reader, batchType string) {
	// 并发模式
	wg := sync.WaitGroup{}
	queries := make(chan string, workers)
	limiter := rate.NewLimiter(rate.Limit(ratePerSecond), 5)

	worker := func(queries <-chan string, wg *sync.WaitGroup) {
		for q := range queries {
			tmpQuery := strings.ReplaceAll(template, "{}", q)
			if err := limiter.Wait(context.Background()); err != nil {
				fmt.Println("Error: ", err)
			}
			if err := writeQuery(tmpQuery); err != nil {
				log.Println("[WARNING]", err)
			}
			wg.Done()
		}
	}
	for w := 0; w < workers; w++ {
		go worker(queries, &wg)
	}

	scanner := bufio.NewScanner(in)
	num := 0
	if batchType != "" {
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
			num++
			if (batchType == "ip" && num >= IPMax) || (batchType == "domain" && num >= DomainMax) {
				wg.Add(1)
				queries <- constructQuery(batchType, lines)
				num = 0
				lines = nil
			}
		}
		if len(lines) > 0 {
			wg.Add(1)
			queries <- constructQuery(batchType, lines)
		}
	} else {
		for scanner.Scan() {
			line := scanner.Text()
			wg.Add(1)
			queries <- line
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

	wg.Wait()
}

func constructQuery(queryType string, queries []string) string {
	var queryBuilder strings.Builder
	for i, query := range queries {
		if i > 0 {
			queryBuilder.WriteString(" || ")
		}
		queryBuilder.WriteString(fmt.Sprintf("%s=%s", queryType, query))
	}
	return queryBuilder.String()
}

// SearchAction search action
func SearchAction(ctx *cli.Context) error {
	// valid same config
	for _, arg := range ctx.Args().Slice() {
		if arg[0] == '-' {
			return errors.New(fmt.Sprintln("there is args after fofa query:", arg))
		}
	}

	query := ctx.Args().First()
	if len(query) == 0 {
		//return errors.New("fofa query cannot be empty")
		log.Println("not set fofa query, now in pipeline mode....")
		if template == "" {
			return errors.New("template cannot be empty in pipeline mode")
		}
	}

	fields := strings.Split(fieldString, ",")
	if len(fields) == 0 {
		return errors.New("fofa fields cannot be empty")
	}

	// headline只允许在format=csv的情况下使用
	if headline && format != "csv" && len(outFile) > 0 {
		return errors.New("headline param is only allowed if format is csv, outFile not be empty")
	}

	// deWildcard不能为0
	if deWildcard == 0 {
		return errors.New("deWildcard param cannot be zero")
	}

	// isActive不能为0
	if checkActive == 0 {
		return errors.New("isActive param cannot be zero")
	}

	// batchType检验
	if batchType != "" && batchType != "ip" && batchType != "domain" {
		return errors.New("batchType param has to be one of ip/domain")
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
	var headFields = fields
	if checkActive > 0 {
		headFields = append(headFields, "isActive")
	}
	if hasBodyField(fields) && format == "csv" {
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

	if headline && format == "csv" && len(outFile) > 0 {
		// 写入表头
		err := writer.WriteAll([][]string{headFields})
		if err != nil {
			return err
		}
	}

	var locker sync.Mutex

	writeQuery := func(query string) error {
		log.Println("query fofa of:", query)
		// do search
		res, err := fofaCli.HostSearch(query, size, fields, gofofa.SearchOptions{
			FixUrl:      fixUrl,
			UrlPrefix:   urlPrefix,
			Full:        full,
			UniqByIP:    uniqByIP,
			CheckActive: checkActive,
			DeWildcard:  deWildcard,
			Filter:      filter,
			DedupHost:   dedupHost,
		})
		if err != nil {
			return err
		}

		// output
		locker.Lock()
		defer locker.Unlock()
		if err = writer.WriteAll(res); err != nil {
			return err
		}
		writer.Flush()

		return nil
	}

	if query != "" {
		return writeQuery(query)
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
		pipelineProcess(writeQuery, inf, batchType)
	}

	return nil
}
