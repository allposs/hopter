package hopter

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"time"
)

// loadConfigFile 读取配置文件
func loadConfigFile() []byte {
	dir, _ := os.Getwd()
	file := dir + "/config/config.yaml"
	b, err := os.ReadFile(file)
	if err != nil {
		Fatal("系统初始化异常:服务器读取配置文件异常，%v", err)
	}
	return b
}

// isExist 文件或目录是否存在
// return false 表示文件不存在
func isExist(name string) bool {
	_, err := os.Stat(name)
	return err == nil || os.IsExist(err)
}

// makeDirAll 创建日志目录
func makeDirAll(logPath string) error {
	logDir := path.Dir(logPath)
	if !isExist(logDir) {
		if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
			return fmt.Errorf("create <%s> error: %s", logDir, err)
		}
	}
	return nil
}

// isWindow 是否是windows系统
func isWindow() bool {
	return runtime.GOOS == "windows"
}

// getNowDateTime 获取当前的日期时间
func getNowDateTime() string {
	return time.Now().Format(TimestampFormat)
}
