package gofofa

import (
	"fmt"
	"github.com/LubyRuffy/gofofa/pkg/readformats"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestCategory(t *testing.T) {
	// 创建一个临时目录以便测试
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()

	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// 初始CSV内容
	tmpCSVFile, err := os.CreateTemp("", "testFile-*.csv")
	assert.Nil(t, err)
	defer os.Remove(tmpCSVFile.Name())

	csvContent := "host,category\nimap.jaas.ac.cn:110,\"电子邮件系统,,其他企业应用\"\npop.jaas.ac.cn:110,电子邮件系统\nimap.mail.suda.edu.cn:995,\"数据证书,电子邮件系统\"\nntp.suda.edu.cn:123,其他支撑系统\nsmtp.jaas.ac.cn:465,数据证书"
	_, err = tmpCSVFile.WriteString(csvContent)
	assert.Nil(t, err)
	tmpCSVFile.Close()

	// 初始YAML内容
	tmpYAMLFile, err := os.CreateTemp("", "config-*.yaml")
	assert.Nil(t, err)
	defer os.Remove(tmpYAMLFile.Name())

	yamlContent := "categories:\n  - name: \"hard\"\n    filters:\n      - \"CONTAIN(category, '数据证书')\"\n\n  - name: \"soft\"\n    filters:\n      - \"CONTAIN(category, '其他支撑系统')\"\n\n  - name: \"buss\"\n    filters:\n      - \"CONTAIN(category, '电子邮件系统')\"\n      - \"CONTAIN(category, '其他企业应用')\"\n"
	_, err = tmpYAMLFile.WriteString(yamlContent)
	assert.Nil(t, err)
	tmpYAMLFile.Close()

	errYAMLFile, err := os.CreateTemp("", "error_config-*.yaml")
	assert.Nil(t, err)
	defer os.Remove(errYAMLFile.Name())

	errYAMLContent := "categories:\n  - name: \"hard\"\n    filters:\n      - \"CONTAIN(category, '数据证书')\"\n\n  - name: \"invalid|soft\"\n    filters:\n      - \"CONTAIN(category, '其他支撑系统')\"\n\n  - name: \"buss\"\n    filters:\n      - \"CONTAIN(category, '电子邮件系统')\"\n      - \"CONTAIN(category, '其他企业应用')\"\n"
	_, err = errYAMLFile.WriteString(errYAMLContent)
	assert.Nil(t, err)
	errYAMLFile.Close()

	// 错误检测
	_, err = Category("dsfhdksajfhsdkjfh", tmpCSVFile.Name(), CategoryOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error reading YAML file")

	_, err = Category(tmpYAMLFile.Name(), "dsfhdksajfhsdkjfh", CategoryOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error opening CSV file")

	_, err = Category(errYAMLFile.Name(), tmpCSVFile.Name(), CategoryOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error creating output file")

	// 正确检测
	resultDir, err := Category(tmpYAMLFile.Name(), tmpCSVFile.Name(), CategoryOptions{})
	assert.Nil(t, err)

	// 验证 result 文件夹存在
	resultPath := filepath.Join(tempDir, "result")
	assert.DirExists(t, resultPath, "result directory not exist")

	// 验证时间文件夹是否存在
	assert.DirExists(t, resultDir, resultDir+" directory not exist")

	// 验证新文件是否生成及内容是否正确
	newFilenames := []string{"buss.csv", "soft.csv", "hard.csv"}

	for _, newFilename := range newFilenames {
		newFilename = filepath.Join(resultDir, newFilename)
		assert.FileExists(t, newFilename, newFilename+" not exist")
		// 验证新文件内容
		actualContent, _, err := readformats.LoadCSVStreamed(newFilename)
		if err != nil {
			t.Fatalf("Failed to read processed file: %v", err)
		}

		switch newFilename {
		case filepath.Join(resultDir, "buss.csv"):
			assert.Equal(t, 3, len(actualContent), "Processed file content mismatch: expected length 3, got "+fmt.Sprint(len(actualContent)))
			continue
		case filepath.Join(resultDir, "soft.csv"):
			assert.Equal(t, 1, len(actualContent), "Processed file content mismatch: expected length 1, got "+fmt.Sprint(len(actualContent)))
			continue
		case filepath.Join(resultDir, "hard.csv"):
			assert.Equal(t, 2, len(actualContent), "Processed file content mismatch: expected length 2, got "+fmt.Sprint(len(actualContent)))
			continue
		default:
			t.Errorf("Processed file was not created: %s", newFilename)
		}
	}
}
