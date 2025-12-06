package utils

import (
	"Go_Zinx/zinterface"
)

// 存储配置参数类
type GlobalObj struct {
	TCPServer      zinterface.IServer
	Host           string
	TCPPort        int
	Name           string
	Version        string
	MaxConn        int
	MaxPackageSize uint32
	// 日志相关配置
	LogLevel int    // 日志级别
	LogFile  string // 日志文件路径
}

var GlobalObject *GlobalObj

// 解析JSON参数
func (g *GlobalObj) Reload() {
	// 暂时使用默认配置，后续可以添加简单的JSON配置读取
	// 这里保留方法是为了兼容现有代码
}

func init() {
	// 默认数值
	GlobalObject = &GlobalObj{
		TCPServer:      nil,
		Host:           "127.0.0.1",
		TCPPort:        8888,
		Name:           "kanade's server",
		Version:        "Latest",
		MaxConn:        1000,
		MaxPackageSize: 1024,
		// 日志默认配置
		LogLevel: INFO,
		LogFile:  "",
	}

	// 尝试从JSON读取配置
	GlobalObject.Reload()

	// 初始化日志
	if GlobalObject.LogFile != "" {
		// 如果配置了日志路径，则使用文件日志，默认文件大小10MB，最多10个文件
		GlobalLogger = NewLogger(GlobalObject.LogLevel, GlobalObject.LogFile, 10*1024*1024, 10)
	} else {
		// 否则使用控制台日志
		GlobalLogger = NewLogger(GlobalObject.LogLevel, "", 0, 0)
	}
}
