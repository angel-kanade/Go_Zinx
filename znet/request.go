package znet

import "Go_Zinx/zinterface"

type Request struct {
	// 建立好的链接
	conn zinterface.IConnection
	// 数据
	msg zinterface.IMessage
}

func (r *Request) GetMsgData() []byte {
	return r.msg.GetData()
}

func (r *Request) GetMsgID() uint32 {
	return r.msg.GetMsgId()
}

func (r *Request) GetConnection() zinterface.IConnection {
	return r.conn
}
