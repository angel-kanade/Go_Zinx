package znet

import (
	"Go_Zinx/zinterface"
	"fmt"
	"net"
)

type Connection struct {
	Conn *net.TCPConn

	ConnID uint32

	isClosed bool

	// 当前链接绑定的业务方法API
	handleApi zinterface.HandleFunc

	// 去告知链接已退出的channel
	ExitChan chan bool
}

func NewConnection(conn *net.TCPConn, connID uint32, callback_api zinterface.HandleFunc) *Connection {
	c := &Connection{
		Conn:      conn,
		ConnID:    connID,
		isClosed:  false,
		handleApi: callback_api,
		ExitChan:  make(chan bool, 1),
	}
	return c
}

func (c *Connection) StartReader() {
	defer fmt.Println("connID =", c.ConnID, "Reader stopped")
	defer c.Stop()
	fmt.Println("Reader Goroutine is running...")

	for {
		buf := make([]byte, 512)

		cnt, err := c.Conn.Read(buf)

		if err != nil {
			fmt.Println("Receive buffer error", err)
			continue
		}

		// 调用HandleAPi
		if err = c.handleApi(c.Conn, buf, cnt); err != nil {
			fmt.Println("ConnID", c.ConnID, "handle failed", err)
			break
		}
	}

}

func (c Connection) Start() {
	fmt.Println("Conn Start... ConnID =", c.ConnID)

	// 启动当前链接的业务
	// Read goroutine
	go c.StartReader()
	// TODO Write goroutine
}

func (c Connection) Stop() {
	fmt.Println("Conn Stop..., ConnID =", c.ConnID)

	if c.isClosed {
		return
	}

	c.isClosed = true
	c.Conn.Close()
	close(c.ExitChan)
	return
}

func (c Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

func (c Connection) GetConnId() uint32 {
	return c.ConnID
}

func (c Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c Connection) Send(data []byte) error {

}
