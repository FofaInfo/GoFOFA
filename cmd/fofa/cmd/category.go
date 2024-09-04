package cmd

import (
	"errors"
	"fmt"
	"github.com/LubyRuffy/gofofa"
	"github.com/LubyRuffy/gofofa/pkg/readformats"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

const (
	ConfigFileName = "config.yaml"
)

var categoryCmd = &cli.Command{
	Name:                   "category",
	Usage:                  "classify data according to config",
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "inFile",
			Aliases:     []string{"i"},
			Usage:       "input data file",
			Destination: &inFile,
		},
	},

	Action: categoryAction,
}

func categoryAction(ctx *cli.Context) error {
	// 检测无效参数
	if len(ctx.Args().Slice()) > 0 {
		return errors.New("invalid arguments")
	}

	// 查找当前目录下是否有config.yaml文件
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %s", err.Error())
	}

	found := false
	filepath.Walk(currentDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && info.Name() == ConfigFileName {
			found = true
			return filepath.SkipDir // 找到后停止查找
		}
		return nil
	})

	if !found {
		return errors.New("not found config.yaml")
	}

	// 检测config文件内容是否合规
	yamlReader := readformats.NewYAMLReader(ConfigFileName)
	config, err := yamlReader.ReadFile()
	if err != nil {
		return fmt.Errorf("error reading YAML file: %v", err)
	}
	if len(config.Categories) == 0 {
		return errors.New("categories not be empty")
	}
	for _, fileType := range config.FileTypes {
		if filepath.Ext(fileType) != ".csv" {
			return errors.New("file type only .csv supported")
		}
	}

	// 检测input是否为空
	if len(inFile) == 0 {
		return errors.New("no input file")
	}

	err = gofofa.Category(ConfigFileName, inFile)
	if err != nil {
		return fmt.Errorf("error category: %s", err.Error())
	}

	return nil
}
