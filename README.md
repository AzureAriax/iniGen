# iniGen
- go语言ini文件读取生成工具
- 在go项目中使用：
```bash
go get github.com/AzureAriax/iniGen@v1.0.1
```
- 以下是使用方法:
``` go
package main

import "github.com/AzureAriax/iniGen/iniGen"

func main() {
	m := map[string]string{
		"service": "./config/conf.go",
		"mysql":   "./config/conf.go",
		"redis":   "./cache/common.go",
		"MongoDB": "./config/conf.go",
	}
	iniGen.Gen("./config/config.ini", m)
}

```
