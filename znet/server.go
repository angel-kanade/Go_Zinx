package znet

import (
	"Go_Zinx/utils"
	"Go_Zinx/zinterface"
	"fmt"
	"net"
	"time"
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

	// 心跳检测器
	HeartbeatChecker *HeartbeatChecker

	// 退出通道
	exitChan chan struct{}
}

func (s *Server) SetOnConnStart(f func(connection zinterface.IConnection)) {
	s.OnConnStart = f
}

func (s *Server) SetOnConnStop(f func(connection zinterface.IConnection)) {
	s.OnConnStop = f
}

func (s *Server) CallOnConnStart(connection zinterface.IConnection) {
	if s.OnConnStart != nil {
		utils.GlobalLogger.Info("=========Call OnConnStart()==========")
		s.OnConnStart(connection)
	}
	// 更新性能指标：连接建立
	utils.GlobalMetrics.IncrementConnectionsTotal()
}

func (s *Server) CallOnConnStop(connection zinterface.IConnection) {
	if s.OnConnStop != nil {
		utils.GlobalLogger.Info("=========Call OnConnStop()===========")
		s.OnConnStop(connection)
	}
	// 更新性能指标：连接关闭
	utils.GlobalMetrics.DecrementConnectionsCurrent()
}

// Server添加一个Handler
func (s *Server) AddHandler(msgId uint32, handler zinterface.IHandler) {
	s.msgRouter.AddHandler(msgId, handler)
	fmt.Println("Add router Success!")
}

func (s *Server) Start() {
	utils.GlobalLogger.Info("[Start] Server Listener at Address: %s:%d", s.IP, s.Port)

	// 开辟一个 go 协程处理服务器启动，防止阻塞
	go func() {
		// 1. 获取一个TCP的Addr:Port
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			utils.GlobalLogger.Error("Resolve TCP Address error: %v", err)
			return
		}

		// 普通监听
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			utils.GlobalLogger.Error("Start Server Listener failed: %v", err)
			return
		}

		defer listener.Close()

		utils.GlobalLogger.Info("start Zinx Server success %s Listening", s.Name)

		var cid uint32
		cid = 0
		// 3. 阻塞等待客户端连接，处理业务
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				utils.GlobalLogger.Error("Accept Error: %v", err)
				continue
			}

			if s.connManager.Len() >= utils.GlobalObject.MaxConn {
				//  给客户端响应超出最大连接
				utils.GlobalLogger.Warn("Too Many Connections MaxConn = %d", utils.GlobalObject.MaxConn)
				conn.Close()
				continue
			}

			cid++
			dealConn := NewConnection(s, conn, cid, s.msgRouter)

			// 将连接添加到心跳检测
			s.HeartbeatChecker.AddConnection(dealConn)
			dealConn.SetProperty("HeartbeatChecker", s.HeartbeatChecker)

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
	utils.GlobalLogger.Info("[STOP] Server's ConnManager is closing")
	s.connManager.ClearConn()

	// 停止心跳检测器
	if s.HeartbeatChecker != nil {
		s.HeartbeatChecker.Stop()
	}

	// 通知退出
	close(s.exitChan)
}

// startMetricsReporter 启动性能指标报告器
func (s *Server) startMetricsReporter() {
	go func() {
		ticker := time.NewTicker(60 * time.Second) // 每分钟报告一次
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 输出性能报告
				utils.GlobalLogger.Info(utils.GlobalMetrics.GetMetricsReport())
			case <-s.exitChan:
				return
			}
		}
	}()
}

func NewServer() zinterface.IServer {
	// 初始化性能指标收集器
	utils.InitMetrics()
	// 创建心跳检测器，每5秒检查一次，超时30秒
	heartbeatChecker := NewHeartbeatChecker(5*time.Second, 30*time.Second)
	heartbeatChecker.Start()

	s := &Server{
		Name:             utils.GlobalObject.Name,
		IPVersion:        "tcp4",
		IP:               utils.GlobalObject.Host,
		Port:             utils.GlobalObject.TCPPort,
		msgRouter:        NewMsgRouter(),
		connManager:      NewConnManager(),
		HeartbeatChecker: heartbeatChecker,
		exitChan:         make(chan struct{}),
	}

	// 添加心跳包处理
	s.AddHandler(0, &HeartbeatHandler{})

	// 启动性能指标报告器
	s.startMetricsReporter()

	return s
}

func (s *Server) GetConnManager() zinterface.IConnManager {
	return s.connManager
}
