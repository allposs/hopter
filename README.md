# hopter
> hopter 是一个gin+logrus+配置进行封装的脚手架,方便快速开发。

# 使用
```bash
go get github.com/allposs/hopter
```
main.go 
```go
package main

import (
	web "github.com/allposs/hopter"
)

func main() {
	web.New(web.NewConfig()).Run()
}
```
# 配置文件
config/config.yaml
```yaml
server:
  port: '8000'
  ip: '0.0.0.0'
```

