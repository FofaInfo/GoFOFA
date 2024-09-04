package gofofa

import (
	"github.com/LubyRuffy/gofofa/pkg/readformats"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestCategory(t *testing.T) {
	// 初始CSV内容
	tmpCSVFile, err := os.CreateTemp("", "testFile-*.csv")
	assert.Nil(t, err)
	defer os.Remove(tmpCSVFile.Name())

	csvContent := "Host,Category\nimap.jaas.ac.cn:110,\"电子邮件系统,,其他企业应用\"\npop.jaas.ac.cn:110,电子邮件系统\nimap.mail.suda.edu.cn:995,\"数据证书,电子邮件系统\"\nntp.suda.edu.cn:123,其他支撑系统\nsmtp.jaas.ac.cn:465,数据证书"
	_, err = tmpCSVFile.WriteString(csvContent)
	assert.Nil(t, err)
	tmpCSVFile.Close()

	// 初始YAML内容
	tmpYAMLFile, err := os.CreateTemp("", "config-*.yaml")
	assert.Nil(t, err)
	defer os.Remove(tmpYAMLFile.Name())

	yamlContent := "categories:\n  数据证书: \"hard\"\n  其他支撑系统: \"soft\"\n  电子邮件系统: \"buss\"\n  其他企业应用: \"buss\"\n\nfile_types:\n  soft: \"soft.csv\"\n  hard: \"hard.csv\"\n  buss: \"buss.csv\""
	_, err = tmpYAMLFile.WriteString(yamlContent)
	assert.Nil(t, err)
	tmpYAMLFile.Close()

	errYAMLFile, err := os.CreateTemp("", "error_config-*.yaml")
	assert.Nil(t, err)
	defer os.Remove(errYAMLFile.Name())

	errYAMLContent := "categories:\n  数据证书: \"hard\"\n  其他支撑系统: \"soft\"\n  电子邮件系统: \"buss\"\n  其他企业应用: \"buss\"\n\nfile_types:\n  soft: \"invalid|soft.csv\"\n  hard: \"hard.csv\"\n  buss: \"buss.csv\""
	_, err = errYAMLFile.WriteString(errYAMLContent)
	assert.Nil(t, err)
	errYAMLFile.Close()

	// 错误检测
	err = Category("dsfhdksajfhsdkjfh", tmpCSVFile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error reading YAML file")

	err = Category(tmpYAMLFile.Name(), "dsfhdksajfhsdkjfh")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error opening CSV file")

	err = Category(errYAMLFile.Name(), tmpCSVFile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error creating output file")

	// 正确检测
	err = Category(tmpYAMLFile.Name(), tmpCSVFile.Name())
	assert.Nil(t, err)

	// 验证新文件是否生成及内容是否正确
	newFilenames := []string{"buss.csv", "soft.csv", "hard.csv"}
	for _, newFilename := range newFilenames {
		if _, err := os.Stat(newFilename); os.IsNotExist(err) {
			t.Errorf("Processed file was not created: %s", newFilename)
		} else {
			defer os.Remove(newFilename) // 测试结束后删除新文件

			// 验证新文件内容
			csvReader := readformats.NewCSVReader(newFilename)
			actualContent, _, err := csvReader.ReadFile()
			if err != nil {
				t.Fatalf("Failed to read processed file: %v", err)
			}

			switch newFilename {
			case "buss.csv":
				if len(actualContent) != 3 {
					t.Errorf("Processed file content mismatch: expected length 3, got %d", len(actualContent))
				}
				continue
			case "soft.csv":
				if len(actualContent) != 1 {
					t.Errorf("Processed file content mismatch: expected 1, got %d", len(actualContent))
				}
				continue
			case "hard.csv":
				if len(actualContent) != 2 {
					t.Errorf("Processed file content mismatch: expected 2, got %d", len(actualContent))
				}
				continue
			default:
				t.Errorf("Processed file was not created: %s", newFilename)
			}
		}
	}

}
