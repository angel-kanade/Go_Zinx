package zinterface

import "net"

type IConnection interface {
	// 启动连接
	Start()

	// 停止链接
	Stop()

	// 获取绑定的 Socket
	GetTCPConnection() *net.TCPConn

	// 获取ID
	GetConnId() uint32

	// 获取 Address
	RemoteAddr() net.Addr

	// 发送数据
	SendMsg(msgId uint32, data []byte) error

	SetProperty(key string, value any)

	GetProperty(key string) (any, error)

	RemoveProperty(key string)
}
