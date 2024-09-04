package gofofa

import (
	"encoding/csv"
	"fmt"
	"github.com/LubyRuffy/gofofa/pkg/readformats"
	"os"
	"strings"
)

type DataCategory struct {
	Type   string
	Record []string
}

func deDuplicates(arr []string) []string {
	exists := make(map[string]bool)

	var result []string
	for _, value := range arr {
		if value == "" {
			continue
		}
		if !exists[value] {
			exists[value] = true
			result = append(result, value)
		}
	}
	return result
}

func Category(configFile, inputFile, category string) error {
	yamlReader := readformats.NewYAMLReader(configFile)
	config, err := yamlReader.ReadFile()
	if err != nil {
		return fmt.Errorf("error reading YAML file: %v", err)
	}

	// 打开 CSV 文件进行读取
	csvReader := readformats.NewCSVReader(inputFile)
	data, header, err := csvReader.ReadFile()
	if err != nil {
		return fmt.Errorf("error opening CSV file: %v", err)
	}

	// 创建输出文件的 writer 映射
	writers := make(map[string]*csv.Writer)
	for _, fileType := range config.FileTypes {
		var file *os.File
		file, err = os.Create(fileType)
		if err != nil {
			return fmt.Errorf("error creating output file: %v", err)
		}
		defer file.Close()
		writers[fileType] = csv.NewWriter(file)
		defer writers[fileType].Flush()
	}

	// 写好表头
	for _, writer := range writers {
		writer.Write(header)
	}

	// 根据分类标准打标签
	for _, record := range data {
		categories := deDuplicates(strings.Split(record[category], ","))
		for _, category := range categories {
			if fileTypeName, ok := config.Categories[category]; ok {
				if writer, ok := writers[config.FileTypes[fileTypeName]]; ok {
					if _, ok := record[fileTypeName]; ok {
						continue
					}
					record[fileTypeName] = fileTypeName
					row := make([]string, len(header))
					for i, headerRow := range header {
						row[i] = record[headerRow]
					}
					writer.Write(row)
				}
			}
		}
	}
	return nil
}
