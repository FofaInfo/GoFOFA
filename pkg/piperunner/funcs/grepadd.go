package funcs

import (
	"bytes"
	"errors"
	"github.com/lubyruffy/gofofa/pkg/pipeast"
	"github.com/lubyruffy/gofofa/pkg/piperunner/corefuncs"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"os"
	"regexp"
	"strings"
	"text/template"
)

func grepAddHook(fi *pipeast.FuncInfo) string {
	tmpl, err := template.New("grep_add").Parse(`AddField(GetRunner(), map[string]interface{}{
    "from": map[string]interface{}{
        "method": "grep",
        "field": {{ .Field }},
        "value": {{ .Value }},
    },
    "name": {{ .Name }},
})`)
	if err != nil {
		panic(err)
	}
	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, struct {
		Field string
		Value string
		Name  string
	}{
		Field: fi.Params[0].String(),
		Value: fi.Params[1].String(),
		Name:  fi.Params[2].String(),
	})
	if err != nil {
		panic(err)
	}
	return tpl.String()
}

type addFieldFrom struct {
	Method  string `json:"method"`
	Field   string
	Value   string
	Options string
}

type addFieldParams struct {
	Name  string
	Value *string       // 可以没有，就取from
	From  *addFieldFrom // 可以没有，就取Value
}

func addField(p corefuncs.Runner, params map[string]interface{}) (string, []string) {

	var err error
	var options addFieldParams
	if err = mapstructure.Decode(params, &options); err != nil {
		panic(err)
	}

	var addValue string
	var addRegex *regexp.Regexp

	var newLines []string
	EachLine(p.GetLastFile(), func(line string) error {
		var newLine string
		if options.Value != nil {
			if addValue == "" {
				addValue = *options.Value
			}
			newLine, _ = sjson.Set(line, options.Name, addValue)
		} else {
			switch options.From.Method {
			case "grep":
				if addRegex == nil {
					addRegex, err = regexp.Compile(options.From.Value)
					if err != nil {
						panic(err)
					}
				}
				res := addRegex.FindAllStringSubmatch(gjson.Get(line, options.From.Field).String(), -1)
				newLine, err = sjson.Set(line, options.Name, res)
				if err != nil {
					panic(err)
				}
			default:
				panic(errors.New("unknown from type"))
			}
		}
		newLines = append(newLines, newLine)
		return nil
	})

	return WriteTempFile(".json", func(f *os.File) {
		content := strings.Join(newLines, "\n")
		n, err := f.WriteString(content)
		if err != nil {
			panic(err)
		}
		if n != len(content) {
			panic("write string failed")
		}
	}), nil
}

func init() {
	corefuncs.RegisterWorkflow("grep_add", grepAddHook, "AddField", addField) // grep匹配再新增字段
}