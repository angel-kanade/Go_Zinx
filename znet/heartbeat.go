package znet

import (
	"Go_Zinx/utils"
	"Go_Zinx/zinterface"
	"sync"
	"time"
)

// HeartbeatChecker 心跳检测器
type HeartbeatChecker struct {
	connections    map[uint32]zinterface.IConnection
	lastActiveTime map[uint32]time.Time
	checkInterval  time.Duration
	timeout        time.Duration
	mutex          sync.RWMutex
	stopChan       chan bool
	wg             sync.WaitGroup
}

// NewHeartbeatChecker 创建心跳检测器
func NewHeartbeatChecker(checkInterval, timeout time.Duration) *HeartbeatChecker {
	return &HeartbeatChecker{
		connections:    make(map[uint32]zinterface.IConnection),
		lastActiveTime: make(map[uint32]time.Time),
		checkInterval:  checkInterval,
		timeout:        timeout,
		stopChan:       make(chan bool),
	}
}

// AddConnection 添加连接到心跳检测
func (hc *HeartbeatChecker) AddConnection(conn zinterface.IConnection) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	connID := conn.GetConnId()
	hc.connections[connID] = conn
	hc.lastActiveTime[connID] = time.Now()

	utils.GlobalLogger.Info("Add connection %d to heartbeat checker", connID)
}

// RemoveConnection 从心跳检测中移除连接
func (hc *HeartbeatChecker) RemoveConnection(connID uint32) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	delete(hc.connections, connID)
	delete(hc.lastActiveTime, connID)

	utils.GlobalLogger.Info("Remove connection %d from heartbeat checker", connID)
}

// UpdateActiveTime 更新连接的活动时间
func (hc *HeartbeatChecker) UpdateActiveTime(connID uint32) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	if _, exists := hc.connections[connID]; exists {
		hc.lastActiveTime[connID] = time.Now()
	}
}

// Start 开始心跳检测
func (hc *HeartbeatChecker) Start() {
	hc.wg.Add(1)
	go func() {
		defer hc.wg.Done()
		utils.GlobalLogger.Info("Heartbeat checker started, check interval: %v, timeout: %v", hc.checkInterval, hc.timeout)

		ticker := time.NewTicker(hc.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				hc.checkHeartbeat()
			case <-hc.stopChan:
				utils.GlobalLogger.Info("Heartbeat checker stopped")
				return
			}
		}
	}()
}

// Stop 停止心跳检测
func (hc *HeartbeatChecker) Stop() {
	close(hc.stopChan)
	hc.wg.Wait()
}

// checkHeartbeat 检查心跳
func (hc *HeartbeatChecker) checkHeartbeat() {
	hc.mutex.RLock()
	// 创建需要关闭的连接ID列表
	var toClose []uint32
	now := time.Now()

	for connID, lastActive := range hc.lastActiveTime {
		if now.Sub(lastActive) > hc.timeout {
			toClose = append(toClose, connID)
		}
	}
	hc.mutex.RUnlock()

	// 关闭超时的连接
	for _, connID := range toClose {
		hc.mutex.Lock()
		conn, exists := hc.connections[connID]
		hc.mutex.Unlock()

		if exists {
			utils.GlobalLogger.Warn("Connection %d heartbeat timeout, closing connection", connID)
			conn.Stop()
			hc.RemoveConnection(connID)
		}
	}
}

// IsHeartbeatMsg 检查是否是心跳消息
func IsHeartbeatMsg(msgID uint32) bool {
	// 假设心跳包的消息ID为0
	return msgID == 0
}

// CreateHeartbeatMsg 创建心跳消息
func CreateHeartbeatMsg() zinterface.IMessage {
	return NewMsgPackage(0, []byte("ping"))
}

// HeartbeatHandler 心跳消息处理器
type HeartbeatHandler struct {
	BaseHandler
}

// Handle 处理心跳消息
func (h *HeartbeatHandler) Handle(request zinterface.IRequest) {
	// 简单响应pong
	heartbeatMsg := CreateHeartbeatMsg()
	heartbeatMsg.SetData([]byte("pong"))
	request.GetConnection().SendMsg(heartbeatMsg.GetMsgId(), heartbeatMsg.GetData())

	// 更新连接的活动时间
	conn := request.GetConnection()
	if checker, err := conn.GetProperty("HeartbeatChecker"); err == nil {
		if hc, ok := checker.(*HeartbeatChecker); ok {
			hc.UpdateActiveTime(conn.GetConnId())
		}
	}
}
