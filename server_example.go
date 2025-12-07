package main

import (
	"Go_Zinx/utils"
	"Go_Zinx/zinterface"
	"Go_Zinx/znet"
	"fmt"
)

// 定义心跳消息处理器
type HeartbeatHandler struct{}

// 心跳消息的预处理
func (h *HeartbeatHandler) PreHandle(request zinterface.IRequest) {
	// 可以在这里添加一些预处理逻辑，如记录日志
	utils.GlobalLogger.Info("[Heartbeat] PreHandle: Client %d sent heartbeat", request.GetConn().GetConnId())
}

// 心跳消息的主处理
func (h *HeartbeatHandler) Handle(request zinterface.IRequest) {
	// 处理心跳消息，通常只需要更新连接的活跃时间
	// 心跳处理逻辑已经在heartbeat.go中实现，这里可以添加额外处理
	utils.GlobalLogger.Info("[Heartbeat] Handle: Responding to heartbeat from client %d", request.GetConn().GetConnId())
	
	// 获取连接的心跳检测器并更新活跃时间
	if checker, ok := request.GetConn().GetProperty("HeartbeatChecker").(zinterface.IHeartbeatChecker); ok {
		checker.UpdateConnection(request.GetConn())
	}
}

// 心跳消息的后处理
func (h *HeartbeatHandler) PostHandle(request zinterface.IRequest) {
	// 可以在这里添加一些后处理逻辑，如统计心跳次数
}

// 定义只读消息处理器
type ReadOnlyHandler struct{}

func (r *ReadOnlyHandler) PreHandle(request zinterface.IRequest) {
	utils.GlobalLogger.Info("[ReadOnly] PreHandle: Client %d requesting read-only data", request.GetConn().GetConnId())
}

func (r *ReadOnlyHandler) Handle(request zinterface.IRequest) {
	// 处理只读请求，返回一些数据
	responseMsg := "This is read-only data from server"
	request.GetConn().SendMsg(1, []byte(responseMsg))
	utils.GlobalLogger.Info("[ReadOnly] Handle: Sent read-only data to client %d", request.GetConn().GetConnId())
}

func (r *ReadOnlyHandler) PostHandle(request zinterface.IRequest) {
	// 可以在这里添加一些后处理逻辑，如记录请求时间
}

// 定义写操作消息处理器
type WriteHandler struct{}

func (w *WriteHandler) PreHandle(request zinterface.IRequest) {
	utils.GlobalLogger.Info("[Write] PreHandle: Client %d requesting write operation", request.GetConn().GetConnId())
}

func (w *WriteHandler) Handle(request zinterface.IRequest) {
	// 处理写操作请求
	receivedData := string(request.GetData())
	utils.GlobalLogger.Info("[Write] Handle: Received write data from client %d: %s", request.GetConn().GetConnId(), receivedData)
	
	// 返回写操作结果
	responseMsg := fmt.Sprintf("Write operation completed successfully: %s", receivedData)
	request.GetConn().SendMsg(2, []byte(responseMsg))
}

func (w *WriteHandler) PostHandle(request zinterface.IRequest) {
	// 可以在这里添加一些后处理逻辑，如数据持久化确认
}

func main() {
	// 创建一个服务器
	server := znet.NewServer()

	// 注册消息处理器，定义不同的msgId
	// msgId = 0: 心跳消息
	server.AddHandler(0, &HeartbeatHandler{})
	
	// msgId = 1: 只读消息请求
	server.AddHandler(1, &ReadOnlyHandler{})
	
	// msgId = 2: 写操作请求
	server.AddHandler(2, &WriteHandler{})
	
	// msgId = 3: 可以继续添加更多消息类型
	// server.AddHandler(3, &AnotherHandler{})

	// 启动服务器
	server.Serve()
}
