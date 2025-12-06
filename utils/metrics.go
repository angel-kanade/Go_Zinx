package utils

import (
	"fmt"
	"sync"
	"time"
)

// Metrics 性能指标收集器
type Metrics struct {
	mu sync.RWMutex

	// 连接相关指标
	ConnectionsTotal   uint64 // 总连接数
	ConnectionsCurrent uint64 // 当前连接数
	ConnectionsClosed  uint64 // 已关闭连接数

	// 消息相关指标
	MessagesReceivedTotal uint64 // 总接收消息数
	MessagesSentTotal     uint64 // 总发送消息数

	// 处理时间相关指标
	MessageHandlingTimeTotal time.Duration // 消息处理总时间
	MessageHandlingCount     uint64        // 消息处理计数

	// 错误相关指标
	ErrorsTotal uint64 // 总错误数
}

// 全局性能指标收集器
var GlobalMetrics *Metrics

// InitMetrics 初始化性能指标收集器
func InitMetrics() {
	GlobalMetrics = &Metrics{}
}

// IncrementConnectionsTotal 增加总连接数
func (m *Metrics) IncrementConnectionsTotal() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ConnectionsTotal++
	m.ConnectionsCurrent++
}

// DecrementConnectionsCurrent 减少当前连接数
func (m *Metrics) DecrementConnectionsCurrent() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ConnectionsCurrent > 0 {
		m.ConnectionsCurrent--
	}
	m.ConnectionsClosed++
}

// IncrementMessagesReceived 增加接收消息数
func (m *Metrics) IncrementMessagesReceived() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesReceivedTotal++
}

// IncrementMessagesSent 增加发送消息数
func (m *Metrics) IncrementMessagesSent() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesSentTotal++
}

// RecordMessageHandlingTime 记录消息处理时间
func (m *Metrics) RecordMessageHandlingTime(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessageHandlingTimeTotal += duration
	m.MessageHandlingCount++
}

// IncrementErrors 增加错误数
func (m *Metrics) IncrementErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ErrorsTotal++
}

// GetAverageMessageHandlingTime 获取平均消息处理时间
func (m *Metrics) GetAverageMessageHandlingTime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.MessageHandlingCount == 0 {
		return 0
	}
	return m.MessageHandlingTimeTotal / time.Duration(m.MessageHandlingCount)
}

// GetMetricsReport 获取性能指标报告
func (m *Metrics) GetMetricsReport() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return fmt.Sprintf(
		`Performance Metrics:
-----------------------------------
Connections:
  Total:       %d
  Current:     %d
  Closed:      %d
Messages:
  Received:    %d
  Sent:        %d
Processing:
  Total Time:  %v
  Average Time: %v
Errors:
  Total:       %d
-----------------------------------
`,
		m.ConnectionsTotal,
		m.ConnectionsCurrent,
		m.ConnectionsClosed,
		m.MessagesReceivedTotal,
		m.MessagesSentTotal,
		m.MessageHandlingTimeTotal,
		m.GetAverageMessageHandlingTime(),
		m.ErrorsTotal,
	)
}
