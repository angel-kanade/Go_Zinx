package znet

import (
	"Go_Zinx/zinterface"
	"errors"
	"fmt"
	"sync"
)

type ConnManager struct {
	connections map[uint32]zinterface.IConnection
	connLock    sync.RWMutex
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint32]zinterface.IConnection),
		connLock:    sync.RWMutex{},
	}
}

func (c *ConnManager) AddConn(conn zinterface.IConnection) {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	c.connections[conn.GetConnId()] = conn
	fmt.Println("connection add to ConnManager")
}

func (c *ConnManager) RemoteConn(connId uint32) {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	delete(c.connections, connId)
	fmt.Println("connId =", connId, "remove from ConnManager")
}

func (c *ConnManager) GetConn(connId uint32) (zinterface.IConnection, error) {
	c.connLock.RLock()
	defer c.connLock.RUnlock()
	if conn, ok := c.connections[connId]; !ok {
		return nil, errors.New("Connection Not Found")
	} else {
		return conn, nil
	}
}

func (c *ConnManager) Len() int {
	return len(c.connections)
}

func (c *ConnManager) ClearConn() {
	c.connLock.Lock()
	defer c.connLock.Unlock()

	for connId, conn := range c.connections {
		conn.Stop()
		delete(c.connections, connId)
	}

	fmt.Println("Clear All connections success!")
}
