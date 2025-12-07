package znet

import (
	"Go_Zinx/utils"
	"Go_Zinx/zinterface"
	"errors"
	"io"
	"net"
	"sync"
	"time"
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
	defer utils.GlobalLogger.Info("connID = %d Writer stopped", c.ConnID)
	utils.GlobalLogger.Info("connID = %d Writer Goroutine is running...", c.ConnID)

	for {
		select {
		case data := <-c.MsgChan:
			if _, err := c.Conn.Write(data); err != nil {
				utils.GlobalLogger.Errorf("Send data error: %v", err)
				utils.GlobalMetrics.IncrementErrors()
				return
			}
			// 更新性能指标：消息发送
			utils.GlobalMetrics.IncrementMessagesSent()
		case <-c.ExitChan:
			// Reader 退出
			return
		}
	}
}

func (c *Connection) StartReader() {
	defer utils.GlobalLogger.Info("connID = %d Reader stopped", c.ConnID)
	defer c.Stop()
	utils.GlobalLogger.Info("connID = %d Reader Goroutine is running...", c.ConnID)

	for {
		dp := utils.NewDataPackUtil()

		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			utils.GlobalLogger.Errorf("read msg head error: %v", err)
			utils.GlobalMetrics.IncrementErrors()
			break
		}

		msg, err := dp.Unpack(headData)
		if err != nil {
			utils.GlobalLogger.Errorf("unpack error: %v", err)
			utils.GlobalMetrics.IncrementErrors()
			break
		}

		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				utils.GlobalLogger.Errorf("read msg data error: %v", err)
				utils.GlobalMetrics.IncrementErrors()
				break
			}
		}

		msg.SetData(data)

		// 更新性能指标：消息接收
		utils.GlobalMetrics.IncrementMessagesReceived()

		req := &Request{
			conn: c,
			msg:  msg,
		}
		// 记录消息处理开始时间
		startTime := time.Now()

		// 获取工作池
		if server, ok := c.TCPServer.(*Server); ok && server.WorkerPool != nil {
			// 使用工作池处理消息
			server.WorkerPool.AddRequest(req)
			// 记录消息处理时间
			utils.GlobalMetrics.RecordMessageHandlingTime(time.Since(startTime))
		} else {
			// 降级方案：直接使用goroutine处理消息
			go func() {
				c.Router.DoMsgHandler(req)
				// 记录消息处理时间
				utils.GlobalMetrics.RecordMessageHandlingTime(time.Since(startTime))
			}()
		}
	}
}

func (c *Connection) Start() {
	utils.GlobalLogger.Info("Conn Start... ConnID = %d", c.ConnID)

	// 启动当前链接的业务
	// Read goroutine
	go c.StartReader()
	// Write goroutine
	go c.StartWriter()

	c.TCPServer.CallOnConnStart(c)
}

func (c *Connection) Stop() {
	utils.GlobalLogger.Info("Conn Stop..., ConnID = %d", c.ConnID)

	if c.isClosed {
		return
	}

	// 从心跳检测器中移除
	if server, ok := c.TCPServer.(*Server); ok && server.HeartbeatChecker != nil {
		server.HeartbeatChecker.RemoveConnection(c.ConnID)
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
		utils.GlobalLogger.Errorf("Pack error msg id = %d", msgId)
		return errors.New("pack error msg")
	}

	c.MsgChan <- binaryMsg

	return nil
}

// GetRouter 获取连接的路由
func (c *Connection) GetRouter() zinterface.IMsgRouter {
	return c.Router
}
