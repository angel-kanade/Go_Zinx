package utils

import (
	"Go_Zinx/zinterface"
	"encoding/json"
	"os"
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
}

var GlobalObject *GlobalObj

// 解析JSON参数
func (g *GlobalObj) Reload() {
	file, err := os.ReadFile("./resource/config.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(file, &g)

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
	}

	// 尝试从JSON读取配置
	GlobalObject.Reload()
}
