package hopter

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

const (
	//FileNameDateFormat 日志文件名的默认日期格式
	FileNameDateFormat = "%Y%m%d"
	//TimestampFormat 日志条目中的默认日期时间格式
	TimestampFormat = "2006-01-02 15:04:05"
	//Text 普通文本格式日志
	Text = "text"
	//JSON json格式日志
	JSON = "json"
	//DataKey json日志条目中 数据字段都会作为该字段的嵌入字段
	DataKey = "data"
)

var (
	logs               *Klogger
	fileNameDateFormat string // 日志文件名的日期格式
	timestampFormat    string // 日志条目中的日期时间格式
)

// option 日志配置参数
type option struct {
	// log 路径
	LogPath string `yaml:"logPath"`
	// 日志类型 json|text
	LogType string `yaml:"logType"`
	//是否不同类型分文件存储
	IsClassSubFile bool `yaml:"isClassSubFile"`
	// 文件名的日期格式
	FileNameDateFormat string `yaml:"fileNameDateFormat"`
	// 是否前台打印日志
	IsForeground bool `yaml:"isForeground"`
	// 日志中日期时间格式
	TimestampFormat string `yaml:"timestampFormat"`
	// 日志级别
	LogLevel string `yaml:"logLevel"`
	// 日志最长保存多久
	MaxAge time.Duration `yaml:"maxAge"`
	// 日志默认多长时间轮转一次
	RotationTime time.Duration `yaml:"rotationTime"`
	// 是否开启记录文件名和行号
	IsEnableRecordFileInfo bool `yaml:"isEnableRecordFileInfo"`
	// 文件名和行号字段名
	FileInfoField string `yaml:"fileInfoField"`
	// json日志是否美化输出
	JSONPrettyPrint bool `yaml:"jsonPrettyPrint"`
	// json日志条目中 数据字段都会作为该字段的嵌入字段
	JSONDataKey string `json:"jsonDataKey"`
}

// Klogger 日志引擎
type Klogger struct {
	*logrus.Logger
	enableRecordFileInfo bool
}

// Logger 获取Logger
func Logger() *Klogger {
	return logs
}

func newLogger(option *option) (*logrus.Logger, error) {
	if option.LogPath == "" {
		dir, _ := os.Getwd()
		path := dir + "/logs/server.log"
		option.LogPath = path
	}
	if err := makeDirAll(option.LogPath); err != nil {
		return nil, err
	}
	if option.FileNameDateFormat == "" {
		fileNameDateFormat = FileNameDateFormat
	} else {
		fileNameDateFormat = option.FileNameDateFormat
	}
	if option.TimestampFormat == "" {
		timestampFormat = TimestampFormat
	} else {
		timestampFormat = option.TimestampFormat
	}
	log := logrus.New()
	log.SetOutput(io.Discard)
	if option.IsForeground {
		log.SetOutput(os.Stderr)
	}
	level, err := logrus.ParseLevel(option.LogLevel)
	if err != nil {
		Warn("系统初始化异常:获取日志级别异常，%v", err)
		level = logrus.InfoLevel
	}
	log.SetLevel(level)
	switch option.LogType {
	case JSON:
		format := &logrus.JSONFormatter{
			TimestampFormat: timestampFormat,
			PrettyPrint:     option.JSONPrettyPrint,
		}
		if option.JSONDataKey != "" {
			format.DataKey = option.JSONDataKey
		}
		log.Formatter = format
	default:
		log.Formatter = &logrus.TextFormatter{
			TimestampFormat: timestampFormat,
		}
	}
	return log, nil
}

// integrate 返回Logger
// 日志类型是: 普通文本日志|JSON日志 全部级别都写入到同一个文件
func integrate(option *option) (*Klogger, error) {
	log, err := newLogger(option)
	if err != nil {
		return nil, err
	}
	writer := new(rotatelogs.RotateLogs)
	if isWindow() {
		writer, err = rotatelogs.New(
			fmt.Sprintf("%s-%s", option.LogPath, fileNameDateFormat),
			rotatelogs.WithMaxAge(option.MaxAge),
			rotatelogs.WithRotationTime(option.RotationTime),
		)
	} else {
		absPath, err := filepath.Abs(option.LogPath)
		if err != nil {
			return nil, fmt.Errorf("rotatelogs.New error: %s", err)
		}
		writer, err = rotatelogs.New(
			fmt.Sprintf("%s-%s", absPath, fileNameDateFormat),
			rotatelogs.WithMaxAge(option.MaxAge),
			rotatelogs.WithRotationTime(option.RotationTime),
			rotatelogs.WithLinkName(absPath),
		)
	}
	if err != nil {
		return nil, err
	}

	fileHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, log.Formatter)

	log.Hooks.Add(fileHook)
	logs = &Klogger{
		log,
		option.IsEnableRecordFileInfo,
	}
	return logs, nil
}

func newRotateLog(option *option, levelStr string) (*rotatelogs.RotateLogs, error) {
	var (
		err      error
		filename string
		writer   *rotatelogs.RotateLogs
		absPath  string
	)

	filename = fmt.Sprintf("%s.%s", option.LogPath, levelStr)
	if isWindow() {
		writer, err = rotatelogs.New(
			fmt.Sprintf("%s.%s", filename, fileNameDateFormat),
			rotatelogs.WithMaxAge(option.MaxAge),
			rotatelogs.WithRotationTime(option.RotationTime),
		)
	} else {
		absPath, err = filepath.Abs(filename)
		if err != nil {
			return nil, fmt.Errorf("rotatelogs.New error: %s", err)
		}

		writer, err = rotatelogs.New(
			fmt.Sprintf("%s.%s", absPath, fileNameDateFormat),
			rotatelogs.WithMaxAge(option.MaxAge),
			rotatelogs.WithRotationTime(option.RotationTime),
			rotatelogs.WithLinkName(absPath),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("rotatelogs.New error: %s", err)
	}

	return writer, nil
}

// separate 不同级别的日志输出到不同的文件
func separate(option *option) (*Klogger, error) {
	log, err := newLogger(option)
	if err != nil {
		return nil, err
	}
	debugWriter, err := newRotateLog(option, "debug")
	if err != nil {
		return nil, err
	}
	infoWriter, err := newRotateLog(option, "info")
	if err != nil {
		return nil, err
	}
	warnWriter, err := newRotateLog(option, "warn")
	if err != nil {
		return nil, err
	}
	errorWriter, err := newRotateLog(option, "error")
	if err != nil {
		return nil, err
	}
	fatalWriter, err := newRotateLog(option, "fatal")
	if err != nil {
		return nil, err
	}
	panicWriter, err := newRotateLog(option, "panic")
	if err != nil {
		return nil, err
	}
	fileHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: debugWriter, // 为不同级别设置不同的输出目的
		logrus.InfoLevel:  infoWriter,
		logrus.WarnLevel:  warnWriter,
		logrus.ErrorLevel: errorWriter,
		logrus.FatalLevel: fatalWriter,
		logrus.PanicLevel: panicWriter,
	}, log.Formatter)

	log.Hooks.Add(fileHook)
	logs = &Klogger{
		log,
		option.IsEnableRecordFileInfo,
	}
	return logs, nil
}

// initLog 初始化日志
func initLog(option *option) (*Klogger, error) {
	if option.IsClassSubFile {
		return separate(option)
	}
	return integrate(option)
}

// LogMode logger接口实现
func (l *Klogger) LogMode(logger.LogLevel) logger.Interface {
	return l
}

// Debug Debug级别日志写入
func (l *Klogger) Debug(ctx context.Context, message string, args ...interface{}) {
	l.WithContext(ctx).Debugf(message, args...)
}

// Info Info级别日志写入
func (l *Klogger) Info(ctx context.Context, message string, args ...interface{}) {
	l.WithContext(ctx).Infof(message, args...)
}

// Warn Warn级别日志写入
func (l *Klogger) Warn(ctx context.Context, message string, args ...interface{}) {
	l.WithContext(ctx).Warnf(message, args...)
}

// Error Error级别日志写入
func (l *Klogger) Error(ctx context.Context, message string, args ...interface{}) {
	l.WithContext(ctx).Errorf(message, args...)
}

// Fatal Fatal级别日志写入
func (l *Klogger) Fatal(ctx context.Context, message string, args ...interface{}) {
	l.WithContext(ctx).Fatalf(message, args...)
}

// Panic Panic级别日志写入
func (l *Klogger) Panic(ctx context.Context, message string, args ...interface{}) {
	l.WithContext(ctx).Panicf(message, args...)
}

// Trace Trace级别日志写入
func (l *Klogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, _ := fc()
	fields := logrus.Fields{}
	//if l.SourceField != "" {
	//	fields[l.SourceField] = utils.FileWithLineNum()
	//}
	//if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) && l.SkipErrRecordNotFound) {
	//	fields[logs.ErrorKey] = err
	//	l.log.WithContext(ctx).WithFields(fields).Errorf("%s [%s]", sql, elapsed)
	//	return
	//}
	//if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
	//	l.log.WithContext(ctx).WithFields(fields).Warnf("%s [%s]", sql, elapsed)
	//	return
	//}
	l.WithContext(ctx).WithFields(fields).Infof("%s [%s]", sql, elapsed)
}

// Debug Debug级别日志写入
func Debug(message string, args ...any) {
	logs.Debugf(message, args...)
}

// Info Info级别日志写入
func Info(message string, args ...any) {
	if logs == nil {
		logrus.Infof(message, args...)
		return
	}
	logs.Infof(message, args...)
}

// Warn Warn级别日志写入
func Warn(message string, args ...any) {
	if logs == nil {
		logrus.Warnf(message, args...)
		return
	}
	logs.Warnf(message, args...)
}

// Error Error级别日志写入
func Error(message string, args ...any) {
	if logs == nil {
		logrus.Errorf(message, args...)
		return
	}
	logs.Errorf(message, args...)
}

// Fatal Fatal级别日志写入
func Fatal(message string, args ...any) {
	if logs == nil {
		logrus.Fatalf(message, args...)
		return
	}
	logs.Fatalf(message, args...)
}

// Panic Panic级别日志写入
func Panic(message string, args ...any) {
	if logs == nil {
		logrus.Panicf(message, args...)
		return
	}
	logs.Panicf(message, args...)
}

// LogMiddleware 日志插件
func LogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		//开始时间
		startTime := time.Now()
		//结束时间
		endTime := time.Now()
		//执行时间
		latencyTime := endTime.Sub(startTime)
		//请求方式
		reqMethod := c.Request.Method
		//请求路由
		reqURI := c.Request.RequestURI
		// 状态码
		statusCode := c.Writer.Status()
		//请求ip
		clientIP := c.ClientIP()
		//请求参数
		reqParams := c.Request.Body
		//请求ua
		reqUa := c.Request.UserAgent()
		var resultBody logrus.Fields
		resultBody = make(map[string]interface{})
		resultBody["requestUri"] = reqURI
		resultBody["clientIp"] = clientIP
		resultBody["body"] = reqParams
		resultBody["userAgent"] = reqUa
		resultBody["requestMethod"] = reqMethod
		resultBody["startTime"] = startTime
		resultBody["endTime"] = endTime
		resultBody["latencyTime"] = latencyTime
		resultBody["statusCode"] = statusCode
		logs.WithFields(resultBody).Info()
	}
}
