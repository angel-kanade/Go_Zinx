# Go_Zinx 扩展功能实现说明

## 1. 日志系统

### 实现方式

日志系统通过 `utils/logger.go` 实现，提供了多级别日志输出、文件日志支持和文件轮换功能。

#### 核心特性

- **日志级别**：支持 DEBUG、INFO、WARN、ERROR、FATAL 五个级别
- **输出目标**：支持控制台和文件输出
- **文件轮换**：支持设置单个文件最大大小和最大文件数量，自动进行文件轮换
- **日志格式化**：每条日志包含时间戳、日志级别和消息内容

#### 核心代码结构

```go
// Logger 日志结构体
type Logger struct {
    Level            int     // 日志级别
    File             *os.File // 日志文件指针
    MaxFileSize      int64   // 单个文件最大大小（字节）
    MaxFiles         int     // 最大文件数量
    FilePath         string  // 日志文件路径
    FileName         string  // 日志文件名（不包含扩展名）
    FileExt          string  // 日志文件扩展名
    CurrentFileIndex int     // 当前文件索引
}
```

#### 主要方法

- `NewLogger(level int, file string, maxSize int64, maxFiles int)`：创建日志实例
- `Debug(format string, args ...interface{})`：输出调试日志
- `Info(format string, args ...interface{})`：输出信息日志
- `Warn(format string, args ...interface{})`：输出警告日志
- `Error(format string, args ...interface{})`：输出错误日志
- `Fatal(format string, args ...interface{})`：输出致命错误日志并退出程序
- `checkAndRotate()`：检查文件大小并进行轮换
- `rotateFile()`：执行文件轮换
- `deleteOldFiles()`：删除超过最大数量的旧文件

#### 使用方法

```go
// 初始化全局日志（在init()中自动初始化）
// 默认输出到控制台，级别为INFO
// 如需启用文件日志，可修改init()函数：
// GlobalLogger = NewLogger(INFO, "logs/zinx.log", 10*1024*1024, 5) // 10MB per file, max 5 files

// 使用全局日志输出
utils.GlobalLogger.Info("Server started on %s:%d", ip, port)
utils.GlobalLogger.Error("Error occurred: %v", err)
```

## 2. 心跳机制

### 实现方式

心跳机制通过 `znet/heartbeat.go` 实现，用于检测和管理连接的活跃状态，避免僵尸连接占用资源。

#### 核心特性

- **定期检查**：定期检查所有连接的活跃状态
- **超时关闭**：超过指定时间未活动的连接会被自动关闭
- **心跳包处理**：支持处理客户端发送的心跳包，更新连接活跃时间

#### 核心代码结构

```go
// HeartbeatChecker 心跳检测器
type HeartbeatChecker struct {
    connMap      map[uint32]time.Time // 连接ID到最后活跃时间的映射
    checkInterval time.Duration        // 检查间隔
    timeout       time.Duration        // 超时时间
    stopChan      chan struct{}        // 停止通道
    mutex         sync.Mutex           // 互斥锁
}
```

#### 主要方法

- `NewHeartbeatChecker(checkInterval, timeout time.Duration)`：创建心跳检测器
- `Start()`：启动心跳检测
- `Stop()`：停止心跳检测
- `AddConnection(conn zinterface.IConnection)`：添加连接到检测列表
- `RemoveConnection(connID uint32)`：从检测列表中移除连接
- `UpdateActivity(connID uint32)`：更新连接的活跃时间
- `handleHeartbeat()`：处理心跳包，更新连接活跃时间
- `checkConnections()`：定期检查所有连接，关闭超时连接

#### 使用方法

心跳机制在 Server 启动时自动初始化和启动：

```go
// 在NewServer()中初始化心跳检测器
heartbeatChecker := NewHeartbeatChecker(5*time.Second, 30*time.Second) // 每5秒检查一次，超时30秒
heartbeatChecker.Start()
```

连接建立时会自动添加到心跳检测列表，连接关闭时自动移除。

## 3. 性能监控

### 实现方式

性能监控通过 `utils/metrics.go` 实现，用于收集和报告服务器的性能指标。

#### 核心特性

- **指标收集**：收集连接数、消息数、错误数、处理时间等指标
- **实时统计**：实时更新各项指标
- **定时报告**：每分钟输出一次性能报告
- **全局可见**：通过 GlobalMetrics 提供全局访问

#### 核心代码结构

```go
// Metrics 性能指标结构体
type Metrics struct {
    connectionsTotal    int64           // 总连接数
    connectionsCurrent  int64           // 当前连接数
    messagesReceived    int64           // 接收消息数
    messagesSent        int64           // 发送消息数
    errors              int64           // 错误数
    messageHandlingTime time.Duration   // 消息总处理时间
    messageCount        int64           // 处理消息总数
    mutex               sync.RWMutex    // 读写锁
}
```

#### 主要方法

- `InitMetrics()`：初始化性能指标收集器
- `IncrementConnectionsTotal()`：增加总连接数
- `IncrementConnectionsCurrent()`：增加当前连接数
- `DecrementConnectionsCurrent()`：减少当前连接数
- `IncrementMessagesReceived()`：增加接收消息数
- `IncrementMessagesSent()`：增加发送消息数
- `IncrementErrors()`：增加错误数
- `RecordMessageHandlingTime(duration time.Duration)`：记录消息处理时间
- `GetMetricsReport()`：生成性能指标报告

#### 定时报告

在 Server 启动时会启动一个定时报告协程，每分钟输出一次性能指标：

```go
// 在startMetricsReporter()方法中实现
func (s *Server) startMetricsReporter() {
    ticker := time.NewTicker(1 * time.Minute)
    go func() {
        for {
            select {
            case <-ticker.C:
                utils.GlobalLogger.Info(utils.GlobalMetrics.GetMetricsReport())
            case <-s.exitChan:
                ticker.Stop()
                return
            }
        }
    }()
}
```

#### 使用方法

性能指标收集器在 Server 启动时自动初始化：

```go
// 在NewServer()中初始化
utils.InitMetrics()
```

各项指标会在连接建立/关闭、消息接收/发送、错误发生时自动更新，无需手动操作。

## 总结

Go_Zinx 框架通过这三个扩展功能，提供了完整的日志记录、连接管理和性能监控能力，提高了服务器的可靠性、可维护性和性能。这些功能都已经集成到框架中，无需额外配置即可使用，也可以根据需要进行定制。
