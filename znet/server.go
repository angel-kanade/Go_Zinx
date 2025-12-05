package znet

import (
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

	// 当前的Server注册的链接处理业务
	Handler zinterface.IHandler
}

// Server添加一个Handler
func (s *Server) AddHandler(handler zinterface.IHandler) {
	s.Handler = handler
	fmt.Println("Add Handler Success!")
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

			cid++
			dealConn := NewConnection(conn, cid, s.Handler)

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

}

func NewServer(name string) zinterface.IServer {
	s := &Server{
		Name:      name,
		IPVersion: "tcp4",
		IP:        "127.0.0.1",
		Port:      8888,
		Handler:   nil,
	}

	return s
}
