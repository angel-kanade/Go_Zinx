package main

import (
	"Go_Zinx/utils"
	"Go_Zinx/zinterface"
	"Go_Zinx/znet"
	"fmt"
	"strings"
	"time"
)

// 2. 回显消息处理器 - 用于测试Writer
type EchoHandler struct{}

func (e *EchoHandler) PreHandle(request zinterface.IRequest) {
	utils.GlobalLogger.Info("[Echo] PreHandle: Client %d sent echo request", request.GetConnection().GetConnId())
}

func (e *EchoHandler) Handle(request zinterface.IRequest) {
	// 获取客户端发送的数据
	receivedData := string(request.GetMsgData())
	utils.GlobalLogger.Info("[Echo] Handle: Received from client %d: %s", request.GetConnection().GetConnId(), receivedData)

	// 使用Writer发送回显响应
	responseData := []byte(fmt.Sprintf("ECHO: %s", receivedData))
	err := request.GetConnection().SendMsg(1, responseData) // 使用SendMsg即使用了Writer
	if err != nil {
		utils.GlobalLogger.Error("[Echo] Failed to send echo response: %v", err)
		return
	}

	utils.GlobalLogger.Info("[Echo] Handle: Sent echo response to client %d", request.GetConnection().GetConnId())
}

func (e *EchoHandler) PostHandle(request zinterface.IRequest) {
	// 可以在这里记录回显请求的统计信息
}

// 3. 数据传输消息处理器
type DataTransferHandler struct{}

func (d *DataTransferHandler) PreHandle(request zinterface.IRequest) {
	utils.GlobalLogger.Info("[DataTransfer] PreHandle: Client %d sent data transfer request", request.GetConnection().GetConnId())
}

func (d *DataTransferHandler) Handle(request zinterface.IRequest) {
	// 处理数据传输请求
	receivedData := string(request.GetMsgData())
	utils.GlobalLogger.Info("[DataTransfer] Handle: Processing data from client %d: %s", request.GetConnection().GetConnId(), receivedData)

	// 模拟数据处理
	processedData := strings.ToUpper(receivedData)

	// 使用Writer发送处理后的数据
	responseData := []byte(fmt.Sprintf("PROCESSED: %s", processedData))
	err := request.GetConnection().SendMsg(2, responseData) // 使用SendMsg即使用了Writer
	if err != nil {
		utils.GlobalLogger.Error("[DataTransfer] Failed to send processed data: %v", err)
		return
	}

	utils.GlobalLogger.Info("[DataTransfer] Handle: Sent processed data to client %d", request.GetConnection().GetConnId())
}

func (d *DataTransferHandler) PostHandle(request zinterface.IRequest) {
	// 可以在这里更新数据传输统计
}

// 4. 关闭连接消息处理器
type CloseConnectionHandler struct{}

func (c *CloseConnectionHandler) PreHandle(request zinterface.IRequest) {
	utils.GlobalLogger.Info("[CloseConnection] PreHandle: Client %d requested connection close", request.GetConnection().GetConnId())
}

func (c *CloseConnectionHandler) Handle(request zinterface.IRequest) {
	// 发送关闭确认
	responseData := []byte("Connection closing...")
	err := request.GetConnection().SendMsg(3, responseData) // 使用SendMsg即使用了Writer
	if err != nil {
		utils.GlobalLogger.Error("[CloseConnection] Failed to send close confirmation: %v", err)
	}

	// 延迟关闭连接，确保客户端收到确认
	go func(conn zinterface.IConnection) {
		time.Sleep(500 * time.Millisecond)
		conn.Stop()
	}(request.GetConnection())

	utils.GlobalLogger.Info("[CloseConnection] Handle: Closing connection for client %d", request.GetConnection().GetConnId())
}

func (c *CloseConnectionHandler) PostHandle(request zinterface.IRequest) {
	// 可以在这里记录连接关闭原因
}

// 5. 状态查询消息处理器
type StatusQueryHandler struct{}

func (s *StatusQueryHandler) PreHandle(request zinterface.IRequest) {
	utils.GlobalLogger.Info("[StatusQuery] PreHandle: Client %d requested server status", request.GetConnection().GetConnId())
}

func (s *StatusQueryHandler) Handle(request zinterface.IRequest) {
	// 构建服务器状态信息
	// 简化实现，直接使用固定的活跃连接数为1
	statusInfo := fmt.Sprintf(`{
	\"server_name\": \"Go_Zinx_Server\",
	\"version\": \"1.0.0\",
	\"uptime\": \"%s\",
	\"active_connections\": 1
}`, time.Since(startTime).String())

	// 使用Writer发送状态信息
	err := request.GetConnection().SendMsg(4, []byte(statusInfo)) // 使用SendMsg即使用了Writer
	if err != nil {
		utils.GlobalLogger.Error("[StatusQuery] Failed to send status: %v", err)
		return
	}

	utils.GlobalLogger.Info("[StatusQuery] Handle: Sent server status to client %d", request.GetConnection().GetConnId())
}

func (s *StatusQueryHandler) PostHandle(request zinterface.IRequest) {
	// 可以在这里记录状态查询次数
}

var startTime time.Time

func main() {
	startTime = time.Now()

	// 创建服务器实例
	server := znet.NewServer()

	// 注册连接钩子函数
	server.SetOnConnStart(func(conn zinterface.IConnection) {
		utils.GlobalLogger.Info("[Hook] Connection established: Client %d", conn.GetConnId())
	})

	server.SetOnConnStop(func(conn zinterface.IConnection) {
		utils.GlobalLogger.Info("[Hook] Connection closed: Client %d", conn.GetConnId())
	})

	// 注册常用的Handler，全部使用Connection的Writer发送响应
	// MsgId=1: 回显消息
	server.AddHandler(1, &EchoHandler{})
	// MsgId=2: 数据传输消息
	server.AddHandler(2, &DataTransferHandler{})
	// MsgId=3: 关闭连接消息
	server.AddHandler(3, &CloseConnectionHandler{})
	// MsgId=4: 状态查询消息
	server.AddHandler(4, &StatusQueryHandler{})

	utils.GlobalLogger.Info("[Server] Registered all handlers successfully")
	utils.GlobalLogger.Info("[Server] Available MsgIds:")
	utils.GlobalLogger.Info("  - MsgId 0: Heartbeat (Ping/Pong)")
	utils.GlobalLogger.Info("  - MsgId 1: Echo (Send back received data)")
	utils.GlobalLogger.Info("  - MsgId 2: Data Transfer (Process and return data)")
	utils.GlobalLogger.Info("  - MsgId 3: Close Connection (Request to close connection)")
	utils.GlobalLogger.Info("  - MsgId 4: Status Query (Get server status)")

	// 启动服务器
	utils.GlobalLogger.Info("[Server] Starting Go_Zinx Server...")
	server.Serve()
}
