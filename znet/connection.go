package znet

import (
	"Go_Zinx/utils"
	"Go_Zinx/zinterface"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
)

type Connection struct {
	// 隶属Server
	TCPServer zinterface.IServer

	Conn *net.TCPConn

	ConnID uint32

	isClosed bool

	// 去告知链接已退出的channel
	ExitChan chan bool

	// 无缓冲管道
	MsgChan chan []byte

	// 该链接处理的方法Router
	Router zinterface.IMsgRouter

	properties     map[string]any
	propertiesLock sync.RWMutex
}

func (c *Connection) SetProperty(key string, value any) {
	c.propertiesLock.Lock()
	defer c.propertiesLock.Unlock()
	c.properties[key] = value
}

func (c *Connection) GetProperty(key string) (any, error) {
	c.propertiesLock.RLock()
	defer c.propertiesLock.RUnlock()
	if value, ok := c.properties[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("No Property Found!")
	}
}

func (c *Connection) RemoveProperty(key string) {
	c.propertiesLock.Lock()
	defer c.propertiesLock.Unlock()
	delete(c.properties, key)
}

func NewConnection(server zinterface.IServer, conn *net.TCPConn, connID uint32, router zinterface.IMsgRouter) *Connection {
	c := &Connection{
		TCPServer:      server,
		Conn:           conn,
		ConnID:         connID,
		isClosed:       false,
		ExitChan:       make(chan bool, 1),
		MsgChan:        make(chan []byte),
		Router:         router,
		properties:     make(map[string]any),
		propertiesLock: sync.RWMutex{},
	}

	// 将conn加入到ConnManager中
	server.GetConnManager().AddConn(c)
	return c
}

func (c *Connection) StartWriter() {
	defer fmt.Println("connID =", c.ConnID, "Writer stopped")
	fmt.Println("Writer Goroutine is running...")

	for {
		select {
		case data := <-c.MsgChan:
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send data error", err)
				return
			}
		case <-c.ExitChan:
			// Reader 退出
			return
		}
	}
}

func (c *Connection) StartReader() {
	defer fmt.Println("connID =", c.ConnID, "Reader stopped")
	defer c.Stop()
	fmt.Println("Reader Goroutine is running...")

	for {
		//buf := make([]byte, utils.GlobalObject.MaxPackageSize)
		//
		//_, err := c.Conn.Read(buf)
		//
		//if err != nil {
		//	fmt.Println("Receive buffer error", err)
		//	continue
		//}
		//
		//// 得到当前conn数据的Req
		//req := &Request{
		//	conn: c,
		//	data: buf,
		//}

		dp := utils.NewDataPackUtil()

		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("read msg head error", err)
			break
		}

		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error", err)
		}

		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error", err)
				break
			}
		}

		msg.SetData(data)

		req := &Request{
			conn: c,
			msg:  msg,
		}
		// 从路由中，找到注册绑定的Conn对应的Router调用
		go c.Router.DoMsgHandler(req)
	}

}

func (c *Connection) Start() {
	fmt.Println("Conn Start... ConnID =", c.ConnID)

	// 启动当前链接的业务
	// Read goroutine
	go c.StartReader()
	// Write goroutine
	go c.StartWriter()

	c.TCPServer.CallOnConnStart(c)
}

func (c *Connection) Stop() {
	fmt.Println("Conn Stop..., ConnID =", c.ConnID)

	if c.isClosed {
		return
	}

	close(c.MsgChan)
	c.isClosed = true
	c.TCPServer.CallOnConnStop(c)
	c.Conn.Close()
	c.ExitChan <- true
	close(c.ExitChan)

	c.TCPServer.GetConnManager().RemoteConn(c.ConnID)

	return
}

func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

func (c *Connection) GetConnId() uint32 {
	return c.ConnID
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed {
		return errors.New("Connection is closed")
	}

	dp := utils.NewDataPackUtil()

	binaryMsg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id =", msgId)
		return errors.New("pack error msg")
	}

	c.MsgChan <- binaryMsg

	return nil
}
