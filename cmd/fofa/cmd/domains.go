package cmd

import (
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"github.com/weppos/publicsuffix-go/publicsuffix"
	"io"
	"log"
	"os"
	"slices"
	"sort"
	"strings"
)

var (
	withCount = false
	clueMode  = false
)

// domains subcommand
var domainsCmd = &cli.Command{
	Name:                   "domains",
	Usage:                  "extend domains from a domain",
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
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
		&cli.BoolFlag{
			Name:        "withCount",
			Value:       false,
			Usage:       "domain with count",
			Destination: &withCount,
		},
		&cli.BoolFlag{
			Name:        "clue",
			Value:       false,
			Usage:       "clue output",
			Destination: &clueMode,
		},
	},
	Action: DomainsAction,
}

type kv struct {
	Key   string
	Value int
}

func sortByValue(m map[string]int) []kv {
	pairs := make([]kv, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Value > pairs[j].Value
	})
	return pairs
}

// DomainsAction extend domains action
func DomainsAction(ctx *cli.Context) error {
	// valid same config
	for _, arg := range ctx.Args().Slice() {
		if arg[0] == '-' {
			return errors.New(fmt.Sprintln("there is args after fofa query:", arg))
		}
	}

	domain := ctx.Args().First()
	if len(domain) == 0 {
		return errors.New("domain cannot be empty")
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

	validCertQuery := `cert.is_valid=true && cert.is_match=true`

	// do search
	res, err := fofaCli.HostSearch(`domain="`+domain+`" && status_code="200" && `+validCertQuery, size, []string{"certs_domains", "certs_subject_org"}, gofofa.SearchOptions{
		Full:     full,
		UniqByIP: uniqByIP,
	})
	if err != nil {
		return err
	}

	domainMap := make(map[string]int)
	orgMap := make(map[string]int)
	for _, hosts := range res {
		if hosts[0] == "" {
			continue
		}

		ns := strings.Split(strings.ReplaceAll(hosts[0], "*", "www"), ",")
		var domains []string
		for _, hostname := range ns {
			// 提取有效的hostname
			domain, err := publicsuffix.Domain(hostname)
			if err != nil {
				log.Println("parse domain failed:", hostname)
			}

			if domain == "" {
				continue
			}

			if !slices.Contains(domains, domain) {
				domains = append(domains, domain)

				domainMap[domain]++
			}
		}

		if len(hosts[1]) > 0 {
			//certs_subject_org
			org := hosts[1]
			orgMap[org]++
		}
	}

	// output
	if clueMode {
		var clues []string
		for k, _ := range domainMap {
			clues = append(clues, fmt.Sprintf(`domain="%s"`, k))
		}
		var orgClues []string
		for k, _ := range orgMap {
			orgClues = append(orgClues, fmt.Sprintf(`cert.subject.org="%s"`, k))
		}
		clues = append(clues, `( (`+strings.Join(orgClues, " || ")+`) && `+validCertQuery+`)`)

		outTo.Write([]byte(strings.Join(clues, " || ")))
	} else {
		for _, kv := range sortByValue(domainMap) {
			if withCount {
				outTo.Write([]byte(fmt.Sprintf("%s\t%d\n", kv.Key, kv.Value)))
			} else {
				outTo.Write([]byte(kv.Key + "\n"))
			}
		}
	}

	return nil
}
