package readformats

import (
	"encoding/csv"
	"io"
	"os"
)

type CSVRow map[string]string

func LoadCSVStreamed(filePath string) ([]CSVRow, []string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	var rows []CSVRow
	headers, err := reader.Read()
	if err != nil {
		return nil, nil, err
	}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		row := make(CSVRow)
		for i, value := range record {
			row[headers[i]] = value
		}
		rows = append(rows, row)
	}
	return rows, headers, nil
}
