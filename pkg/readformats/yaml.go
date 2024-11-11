package readformats

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const maxCapacity = 2048 * 1024

type YAMLReader struct {
	Filename string
}

type Category struct {
	Name    string   `yaml:"name"`
	Filters []string `yaml:"filters"`
}

type Config struct {
	Categories []Category `yaml:"categories"`
}

func NewYAMLReader(filename string) *YAMLReader {
	return &YAMLReader{Filename: filename}
}

func (y *YAMLReader) ReadLines() ([]byte, error) {
	file, err := os.Open(y.Filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	var buffer bytes.Buffer
	buf := make([]byte, maxCapacity)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(buf, maxCapacity)
	lineNumber := 1
	for scanner.Scan() {
		buffer.Write(scanner.Bytes())
		buffer.WriteByte('\n')
		lineNumber++
	}

	if err = scanner.Err(); err != nil {
		return nil, fmt.Errorf("something bad happened in the line %v: %v", lineNumber, err)
	}

	return buffer.Bytes(), nil
}

func (y *YAMLReader) ReadFile() (Config, error) {
	data, err := y.ReadLines()
	if err != nil {
		return Config{}, fmt.Errorf("failed to read file: %v", err)
	}

	var config Config
	if err = yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal yaml: %v", err)
	}

	return config, nil
}
