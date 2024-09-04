package readformats

import (
	"encoding/csv"
	"fmt"
	"os"
)

type CSVReader struct {
	Filename      string
	ColumnsToKeep []string
}

func NewCSVReader(filename string) *CSVReader {
	return &CSVReader{Filename: filename}
}

func (c *CSVReader) ReadFile() ([]map[string]string, []string, error) {
	file, err := os.Open(c.Filename)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read header: %v", err)
	}

	keepIndices := make(map[int]string)
	for i, col := range header {
		if len(c.ColumnsToKeep) > 0 {
			for _, keepCol := range c.ColumnsToKeep {
				if col == keepCol {
					keepIndices[i] = col
					break
				}
			}
		}
		keepIndices[i] = col
	}

	var result []map[string]string
	for {
		row, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, nil, fmt.Errorf("failed to read row: %v", err)
		}

		rowData := make(map[string]string)
		for i, value := range row {
			if col, ok := keepIndices[i]; ok {
				rowData[col] = value
			}
		}
		result = append(result, rowData)
	}

	return result, header, nil
}
