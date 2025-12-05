package znet

import "Go_Zinx/zinterface"

type Request struct {
	// 建立好的链接
	conn zinterface.IConnection
	// 数据
	data []byte
}

func (r *Request) GetConnection() zinterface.IConnection {
	return r.conn
}

func (r *Request) GetData() []byte {
	return r.data
}
