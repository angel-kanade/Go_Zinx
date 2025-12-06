package zinterface

type IConnManager interface {
	AddConn(conn IConnection)
	RemoteConn(connId uint32)
	GetConn(connId uint32) (IConnection, error)
	Len() int
	ClearConn()
}
