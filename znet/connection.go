package znet

import (
	"Go_Zinx/utils"
	"Go_Zinx/zinterface"
	"fmt"
	"net"
)

type Connection struct {
	Conn *net.TCPConn

	ConnID uint32

	isClosed bool

	// 去告知链接已退出的channel
	ExitChan chan bool

	// 该链接处理的方法Router
	Router zinterface.IHandler
}

func NewConnection(conn *net.TCPConn, connID uint32, router zinterface.IHandler) *Connection {
	c := &Connection{
		Conn:     conn,
		ConnID:   connID,
		isClosed: false,
		ExitChan: make(chan bool, 1),
		Router:   router,
	}
	return c
}

func (c *Connection) StartReader() {
	defer fmt.Println("connID =", c.ConnID, "Reader stopped")
	defer c.Stop()
	fmt.Println("Reader Goroutine is running...")

	for {
		buf := make([]byte, utils.GlobalObject.MaxPackageSize)

		_, err := c.Conn.Read(buf)

		if err != nil {
			fmt.Println("Receive buffer error", err)
			continue
		}

		// 得到当前conn数据的Req
		req := &Request{
			conn: c,
			data: buf,
		}

		// 从路由中，找到注册绑定的Conn对应的Router调用
		go func(req zinterface.IRequest) {
			c.Router.PreHandle(req)
			c.Router.Handle(req)
			c.Router.PostHandle(req)
		}(req)
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
	return nil
}
