package znet

import (
	"Go_Zinx/utils"
	"Go_Zinx/zinterface"
	"errors"
	"fmt"
	"io"
	"net"
)

type Connection struct {
	Conn *net.TCPConn

	ConnID uint32

	isClosed bool

	// 去告知链接已退出的channel
	ExitChan chan bool

	// 无缓冲管道
	MsgChan chan []byte

	// 该链接处理的方法Router
	Router zinterface.IMsgRouter
}

func NewConnection(conn *net.TCPConn, connID uint32, router zinterface.IMsgRouter) *Connection {
	c := &Connection{
		Conn:     conn,
		ConnID:   connID,
		isClosed: false,
		ExitChan: make(chan bool, 1),
		MsgChan:  make(chan []byte),
		Router:   router,
	}
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

		if utils.GlobalObject.WorkerPoolSize > 0 {
			// 已经开启工作池机制
			c.Router.SendMsgToTaskQueue(req)
		} else {
			go c.Router.DoMsgHandler(req)
		}
	}

}

func (c *Connection) Start() {
	fmt.Println("Conn Start... ConnID =", c.ConnID)

	// 启动当前链接的业务
	// Read goroutine
	go c.StartReader()
	// Write goroutine
	go c.StartWriter()
}

func (c *Connection) Stop() {
	fmt.Println("Conn Stop..., ConnID =", c.ConnID)

	if c.isClosed {
		return
	}

	close(c.MsgChan)
	c.isClosed = true
	c.Conn.Close()
	c.ExitChan <- true
	close(c.ExitChan)
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
