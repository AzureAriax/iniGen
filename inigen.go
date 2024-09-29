package inigen

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/template"

	"gopkg.in/ini.v1"
)

type StructInfo struct {
	StructName  string
	Fields      map[string]string
	FuncName    string
	SectionName string
}

func gen() {
	fmt.Println("代码生成开始!")
	iniFilePath := "./config/config.ini"
	structs, err := parseIniFile(iniFilePath)
	if err != nil {
		log.Fatalf("解析 ini 文件失败: %v", err)
	}
	if err := generateCode(structs); err != nil {
		log.Fatalf("生成代码失败: %v", err)
	}
	log.Println("配置加载函数生成成功！")
}

func iniTypeToGo(value string) string {
	if _, err := strconv.Atoi(value); err == nil {
		return "int"
	}
	if _, err := strconv.ParseBool(value); err == nil {
		return "bool"
	}
	return "string"
}

func parseIniFile(filepath string) ([]StructInfo, error) {
	cfg, err := ini.Load(filepath)
	if err != nil {
		return nil, err
	}
	var structs []StructInfo
	for _, section := range cfg.Sections() {
		//跳过默认 section
		if section.Name() == "DEFAULT" {
			continue
		}

		structInfo := StructInfo{
			StructName:  section.Name(),
			FuncName:    fmt.Sprint("Load", strings.Title(section.Name())),
			SectionName: section.Name(),
			Fields:      make(map[string]string),
		}

		for _, key := range section.Keys() {
			structInfo.Fields[key.Name()] = iniTypeToGo(key.Value())
		}
		structs = append(structs, structInfo)
	}
	return structs, nil
}
func generateCode(structs []StructInfo) error {
	// 模板
	const templateText = `
package config

import (
    "gopkg.in/ini.v1"
    "log"
)

{{range .}}
// {{.FuncName}} 加载 {{.StructName}} 的配置
type {{.StructName}} struct {
    {{- range $key, $type := .Fields }}
    {{ $key | title }} {{ $type }}
    {{- end }}
}

func {{.FuncName}}(file *ini.File) (*{{.StructName}}, error) {
    cfg := &{{.StructName}}{}
    err := file.Section("{{.SectionName}}").MapTo(cfg)
    if err != nil {
        log.Printf("加载 {{.StructName}} 配置失败: %v", err)
        return nil, err
    }
    return cfg, nil
}
{{end}}
`
	tmpl, err := template.New("config").Funcs(template.FuncMap{
		"title": strings.Title,
	}).Parse(templateText)
	if err != nil {
		return err
	}

	outFile, err := os.Create("config/config.go")
	if err != nil {
		return err
	}
	defer outFile.Close()

	return tmpl.Execute(outFile, structs)
}
