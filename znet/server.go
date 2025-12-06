package znet

import (
	"Go_Zinx/utils"
	"Go_Zinx/zinterface"
	"fmt"
	"net"
)

// IServer的接口实现，定义一个Server的服务器模块
type Server struct {
	Name      string
	IPVersion string
	IP        string
	Port      int

	// 当前的Server注册的Router
	msgRouter zinterface.IMsgRouter

	// connManager
	connManager zinterface.IConnManager

	// hook
	OnConnStart func(conn zinterface.IConnection)
	OnConnStop  func(conn zinterface.IConnection)
}

func (s *Server) SetOnConnStart(f func(connection zinterface.IConnection)) {
	s.OnConnStart = f
}

func (s *Server) SetOnConnStop(f func(connection zinterface.IConnection)) {
	s.OnConnStop = f
}

func (s *Server) CallOnConnStart(connection zinterface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("=========Call OnConnStart()==========")
		s.OnConnStart(connection)
	}
}

func (s *Server) CallOnConnStop(connection zinterface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("=========Call OnConnStop()===========")
		s.OnConnStop(connection)
	}
}

// Server添加一个Handler
func (s *Server) AddHandler(msgId uint32, handler zinterface.IHandler) {
	s.msgRouter.AddHandler(msgId, handler)
	fmt.Println("Add router Success!")
}

func (s *Server) Start() {
	fmt.Printf("[Start] Server Listener at Address: %s:%d\n", s.IP, s.Port)

	// 开辟一个 go 协程处理服务器启动，防止阻塞
	go func() {
		// 1. 获取一个TCP的Addr:Port
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("Resolve TCP Address error", err)
		}
		// 2. 监听
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("Start Server Listener failed", err)
		}

		fmt.Println("start Zinx Server success", s.Name, "Listening")

		var cid uint32
		cid = 0
		// 3. 阻塞等待客户端连接，处理业务
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept Error", err)
				continue
			}

			if s.connManager.Len() >= utils.GlobalObject.MaxConn {
				//  给客户端响应超出最大连接
				fmt.Println("Too Many Connections MaxConn =", utils.GlobalObject.MaxConn)
				conn.Close()
				continue
			}

			cid++
			dealConn := NewConnection(s, conn, cid, s.msgRouter)

			// 启动链接业务处理
			go dealConn.Start()
		}
	}()

}

func (s *Server) Serve() {
	s.Start()

	// TODO 做一些启动服务器之外的额外业务

	// 保持阻塞状态
	select {}
}

func (s *Server) Stop() {
	fmt.Println("[STOP] Server's ConnManager is closing")
	s.connManager.ClearConn()
}

func NewServer() zinterface.IServer {
	s := &Server{
		Name:        utils.GlobalObject.Name,
		IPVersion:   utils.GlobalObject.Version,
		IP:          utils.GlobalObject.Host,
		Port:        utils.GlobalObject.TCPPort,
		msgRouter:   NewMsgRouter(),
		connManager: NewConnManager(),
	}

	return s
}

func (s *Server) GetConnManager() zinterface.IConnManager {
	return s.connManager
}
