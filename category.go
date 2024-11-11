package gofofa

import (
	"encoding/csv"
	"fmt"
	"github.com/LubyRuffy/gofofa/pkg/readformats"
	"github.com/expr-lang/expr"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DataCategory struct {
	Type   string
	Record []string
}

type CategoryOptions struct {
	Unique       bool   // is the classification unique
	RelationFile string // relation file
	SourceField  string // source field
	TargetField  string // target field
}

func evaluateExpressions(filters []string, data readformats.CSVRow) (bool, error) {
	// 处理数据
	env := make(map[string]interface{})
	for i, value := range data {
		env[i] = value
	}

	// 添加过滤器内置方法
	env["CONTAIN"] = strings.Contains

	for _, filter := range filters {
		program, err := expr.Compile(filter, expr.Env(env))
		if err != nil {
			return false, err
		}

		output, err := expr.Run(program, env)
		if err != nil {
			return false, err
		}

		if output.(bool) {
			return true, nil
		}
	}
	return false, nil
}

func Category(configFile, inputFile string, options ...CategoryOptions) (string, error) {
	var (
		unique bool
	)

	if len(options) > 0 {
		unique = options[0].Unique
	}

	yamlReader := readformats.NewYAMLReader(configFile)
	config, err := yamlReader.ReadFile()
	if err != nil {
		return "", fmt.Errorf("error reading YAML file: %v", err)
	}

	// 打开 CSV 文件进行读取
	data, header, err := readformats.LoadCSVStreamed(inputFile)
	if err != nil {
		return "", fmt.Errorf("error opening CSV file: %v", err)
	}

	if len(data) == 0 {
		return "", fmt.Errorf("no data found in input file: %s", inputFile)
	}

	// 创建输出文件的 writer 映射
	writers := make(map[string]*csv.Writer)
	// 创建一个存放result的文件夹
	resultDir := filepath.Join("result", fmt.Sprintf(time.Now().Format("20060102150405")))
	if err = os.MkdirAll(resultDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("error creating results directory: %v", err)
	}

	for _, cate := range config.Categories {
		// 创建分类文件writer
		fileName := filepath.Join(resultDir, cate.Name+".csv")
		var file *os.File
		file, err = os.Create(fileName)
		if err != nil {
			return "", fmt.Errorf("error creating output file: %v", err)
		}
		defer file.Close()
		writers[cate.Name] = csv.NewWriter(file)
		defer writers[cate.Name].Flush()
	}

	// 写好表头
	for _, writer := range writers {
		writer.Write(header)
	}

	counts := make(map[string]int)

	// 根据分类标准打标签
	var match bool
	for _, recordMap := range data {
		for _, cate := range config.Categories {
			match, err = evaluateExpressions(cate.Filters, recordMap)
			if err != nil {
				log.Println("Error evaluating expressions:", err)
				continue
			}

			if match {
				row := make([]string, len(header))
				for i, headerRow := range header {
					row[i] = fmt.Sprintf("%v", recordMap[headerRow])
				}
				err = writers[cate.Name].Write(row)
				if err != nil {
					return "", fmt.Errorf("error writing record: %v", err)
				}
				// 匹配成功，写入分类文件
				counts[cate.Name]++
				if unique {
					break
				}
			}
		}

	}

	for _, cate := range config.Categories {
		fmt.Println("[-] Matches category:", cate.Name, ", Length:", counts[cate.Name])
	}

	return resultDir, nil
}
