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
	deduplicateString string
)

var duplicateCmd = &cli.Command{
	Name:                   "deduplicate",
	Usage:                  "remove duplicate tool",
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "duplicate",
			Aliases:     []string{"d"},
			Usage:       "remove duplicate according to the fields",
			Destination: &deduplicateString,
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

func deduplicate(records [][]string, fieldName string) ([][]string, error) {
	headers := records[0]
	var index int
	found := false
	for i, header := range headers {
		if header == fieldName {
			index = i
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("field %s not exist", fieldName)
	}

	seen := make(map[string]bool)
	var uniqueRecords [][]string
	uniqueRecords = append(uniqueRecords, headers)

	for _, record := range records[1:] {
		key := record[index]
		if !seen[key] {
			uniqueRecords = append(uniqueRecords, record)
			seen[key] = true
		}
	}
	return uniqueRecords, nil
}

func deduplicates(records [][]string, fields []string) ([][]string, error) {
	var uniqueRecords [][]string
	for _, field := range fields {
		uniq, err := deduplicate(records, field)
		if err != nil {
			return nil, errors.New("deduplicate failed: " + err.Error())
		}
		uniqueRecords = append(uniqueRecords, uniq...)
	}

	return uniqueRecords, nil
}

func deduplicateAction(ctx *cli.Context) error {
	// valid same config
	if len(ctx.Args().Slice()) > 0 {
		return errors.New("Use -h to find help ")
	}

	duplicate := strings.Split(deduplicateString, ",")
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
