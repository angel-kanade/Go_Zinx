package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// 日志级别定义
const (
	DEBUG = iota
	INFO
	WARN
	ERROR
	FATAL
)

// 日志级别名称
var levelNames = []string{
	"DEBUG",
	"INFO",
	"WARN",
	"ERROR",
	"FATAL",
}

// Logger 日志结构体
type Logger struct {
	Level            int
	File             *os.File
	MaxFileSize      int64  // 单个文件最大大小（字节）
	MaxFiles         int    // 最大文件数量
	FilePath         string // 日志文件路径
	FileName         string // 日志文件名（不包含扩展名）
	FileExt          string // 日志文件扩展名
	CurrentFileIndex int    // 当前文件索引
}

// NewLogger 创建一个新的日志实例
func NewLogger(level int, file string, maxSize int64, maxFiles int) *Logger {
	var f *os.File
	var filePath, fileName, fileExt string
	var currentIndex int

	if file != "" {
		filePath = filepath.Dir(file)
		fileName = filepath.Base(file)
		fileExt = filepath.Ext(fileName)
		fileName = fileName[:len(fileName)-len(fileExt)]

		// 如果文件存在，获取文件信息以确定当前大小和索引
		if _, err := os.Stat(file); err == nil {
			// 文件存在，检查是否需要轮换
			f, err = os.OpenFile(file, os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				fmt.Printf("Open log file %s error: %v\n", file, err)
				f = os.Stdout
			} else {
				// 检查文件大小
				fileInfo, _ := f.Stat()
				if fileInfo.Size() >= maxSize {
					// 需要轮换，关闭当前文件
					f.Close()
					currentIndex = 1
					// 创建新文件
					newFile := fmt.Sprintf("%s/%s.%d%s", filePath, fileName, currentIndex, fileExt)
					f, err = os.OpenFile(newFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
					if err != nil {
						fmt.Printf("Create log file %s error: %v\n", newFile, err)
						f = os.Stdout
					}
				}
			}
		} else {
			// 文件不存在，创建新文件
			f, err = os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				fmt.Printf("Create log file %s error: %v\n", file, err)
				f = os.Stdout
			}
		}
	} else {
		f = os.Stdout
	}

	return &Logger{
		Level:            level,
		File:             f,
		MaxFileSize:      maxSize,
		MaxFiles:         maxFiles,
		FilePath:         filePath,
		FileName:         fileName,
		FileExt:          fileExt,
		CurrentFileIndex: currentIndex,
	}
}

// checkAndRotate 检查文件大小并进行轮换
func (l *Logger) checkAndRotate() {
	if l.File == os.Stdout {
		return
	}

	fileInfo, err := l.File.Stat()
	if err != nil {
		return
	}

	if fileInfo.Size() >= l.MaxFileSize {
		l.rotateFile()
	}
}

// rotateFile 执行文件轮换
func (l *Logger) rotateFile() {
	// 关闭当前文件
	l.File.Close()

	// 增加文件索引
	l.CurrentFileIndex++

	// 创建新文件名
	newFileName := fmt.Sprintf("%s/%s.%d%s", l.FilePath, l.FileName, l.CurrentFileIndex, l.FileExt)

	// 创建新文件
	var err error
	l.File, err = os.OpenFile(newFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Create new log file %s error: %v\n", newFileName, err)
		l.File = os.Stdout
		return
	}

	// 删除旧文件
	l.deleteOldFiles()
}

// deleteOldFiles 删除超过最大数量的旧文件
func (l *Logger) deleteOldFiles() {
	if l.MaxFiles <= 0 {
		return
	}

	// 计算需要保留的最小索引
	minIndex := l.CurrentFileIndex - l.MaxFiles + 1
	if minIndex <= 0 {
		return
	}

	// 删除旧文件
	for i := 1; i < minIndex; i++ {
		oldFileName := fmt.Sprintf("%s/%s.%d%s", l.FilePath, l.FileName, i, l.FileExt)
		os.Remove(oldFileName)
	}
}

// log 内部日志输出方法
func (l *Logger) log(level int, format string, args ...interface{}) {
	if level < l.Level {
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	levelName := levelNames[level]
	msg := fmt.Sprintf(format, args...)

	logMsg := fmt.Sprintf("%s [%s] %s\n", now, levelName, msg)

	// 写入日志
	l.File.WriteString(logMsg)

	// 检查文件大小并进行轮换
	l.checkAndRotate()

	// 如果是致命错误，输出后退出程序
	if level == FATAL {
		os.Exit(1)
	}
}

// Debug 输出调试日志
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info 输出信息日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn 输出警告日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error 输出错误日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Errorf 输出格式化错误日志
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal 输出致命错误日志并退出程序
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
}

// Close 关闭日志文件
func (l *Logger) Close() {
	if l.File != os.Stdout {
		l.File.Close()
	}
}

// 全局日志实例
var GlobalLogger *Logger

// 初始化全局日志
func init() {
	// 默认输出到控制台，级别为INFO
	// 如果要启用文件日志，可修改这里的参数
	// 参数说明：日志级别，日志文件路径，单个文件最大大小（字节），最大文件数量
	// GlobalLogger = NewLogger(INFO, "logs/zinx.log", 10*1024*1024, 5) // 示例：10MB per file, max 5 files
	GlobalLogger = NewLogger(INFO, "", 0, 0) // 默认输出到控制台
}
