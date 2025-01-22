package cmd

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

var (
	dedupString string
)

var dedupCmd = &cli.Command{
	Name:                   "dedup",
	Usage:                  "remove duplicate tool",
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "dedup",
			Aliases:     []string{"d"},
			Usage:       "remove duplicate according to the fields",
			Destination: &dedupString,
		},
		&cli.StringFlag{
			Name:        "inFile",
			Aliases:     []string{"i"},
			Usage:       "input duplicate file",
			Destination: &inFile,
		},
		&cli.StringFlag{
			Name:        "outFile",
			Aliases:     []string{"o"},
			Value:       "duplicate.csv",
			Usage:       "write csv file",
			Destination: &outFile,
		},
	},

	Action: deduplicateAction,
}

func isCSV(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), ".csv")
}

func readCSV(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return records, nil
}

func writeCSV(filePath string, records [][]string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, record := range records {
		if err = writer.Write(record); err != nil {
			return err
		}
	}
	return nil
}

// indexOf 查找字段在表头中的索引
func indexOf(headers []string, field string) int {
	for i, header := range headers {
		if header == field {
			return i
		}
	}
	return -1
}

func deduplicates(records [][]string, fields []string) ([][]string, error) {
	if len(records) < 2 {
		return nil, errors.New("deduplicate failed: CSV file is empty")
	}

	// 获取字段索引
	fieldIndexes := make([]int, 0)
	for _, field := range fields {
		index := indexOf(fields, field)
		if index == -1 {
			return nil, fmt.Errorf("field '%s' not found in headers", field)
		}
		fieldIndexes = append(fieldIndexes, index)
	}

	// 去重逻辑
	seen := make(map[string]bool)
	uniqueRows := [][]string{records[0]}

	for _, row := range records[1:] {
		keyParts := make([]string, len(fields))
		for i, idx := range fieldIndexes {
			keyParts[i] = row[idx]
		}
		key := strings.Join(keyParts, "|")
		if !seen[key] {
			seen[key] = true
			uniqueRows = append(uniqueRows, row)
		}
	}

	return uniqueRows, nil
}

func deduplicateAction(ctx *cli.Context) error {
	// valid same config
	if len(ctx.Args().Slice()) > 0 {
		return errors.New("Use -h to find help ")
	}

	duplicate := strings.Split(dedupString, ",")
	if len(duplicate) == 0 || len(inFile) == 0 {
		return errors.New("flag needs arguments: -d field -i target.csv")
	}

	if !isCSV(inFile) {
		return errors.New("invalid file type, please input a csv file type")
	}

	records, err := readCSV(inFile)
	if err != nil {
		return errors.New("read csv file failed: " + err.Error())
	}

	uniqueRecords, err := deduplicates(records, duplicate)
	if err != nil {
		return errors.New("deduplicates failed: " + err.Error())
	}

	err = writeCSV(outFile, uniqueRecords)
	if err != nil {
		return errors.New("write csv file failed: " + err.Error())
	}

	fmt.Println("The deduplicated CSV file has been created:", outFile)

	return nil
}
