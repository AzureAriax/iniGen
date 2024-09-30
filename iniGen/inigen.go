package iniGen

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
	FilePath    string
}

func Gen(iniFilePath string, outputPaths map[string]string) {
	fmt.Println("代码生成开始!")
	structs, err := parseIniFile(iniFilePath, outputPaths)
	if err != nil {
		log.Fatalf("解析 ini 文件失败: %v", err)
	}
	for _, structInfo := range structs {
		if err := generateCode(structInfo); err != nil {
			log.Fatalf("生成代码失败: %v", err)
		}
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

func parseIniFile(filepath string, outputPaths map[string]string) ([]StructInfo, error) {
	cfg, err := ini.Load(filepath)
	if err != nil {
		return nil, err
	}

	var structs []StructInfo
	for _, section := range cfg.Sections() {
		// 跳过默认 section
		if section.Name() == "DEFAULT" {
			continue
		}

		outputPath, ok := outputPaths[section.Name()]
		if !ok {
			return nil, fmt.Errorf("未找到服务节 %s 对应的输出路径", section.Name())
		}

		structInfo := StructInfo{
			StructName:  section.Name(),
			FuncName:    fmt.Sprint("Load", strings.Title(section.Name())),
			SectionName: section.Name(),
			Fields:      make(map[string]string),
			FilePath:    outputPath,
		}

		for _, key := range section.Keys() {
			structInfo.Fields[key.Name()] = iniTypeToGo(key.Value())
		}
		structs = append(structs, structInfo)
	}

	return structs, nil
}

func generateCode(structInfo StructInfo) error {
	// 模板
	const templateText = `
package {{.StructName}}

import (
    "gopkg.in/ini.v1"
    "log"
)

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

func init() {
    file, err := ini.Load("config.ini")
    if err != nil {
        log.Fatalf("无法加载配置文件: %v", err)
    }
    _, err = {{.FuncName}}(file)
    if err != nil {
        log.Fatalf("初始化 {{.StructName}} 配置失败: %v", err)
    }
}
`
	tmpl, err := template.New("config").Funcs(template.FuncMap{
		"title": strings.Title,
	}).Parse(templateText)
	if err != nil {
		return err
	}
	// 直接使用 structInfo.FilePath 作为输出路径
	outFile, err := os.Create(structInfo.FilePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return tmpl.Execute(outFile, structInfo)
}
