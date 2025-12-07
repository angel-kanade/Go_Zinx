# 常用Handler注册与Writer使用说明

## 一、已注册的常用Handler

我已经为您注册了5个常用的Handler，每个Handler都使用了Connection的Writer来发送响应：

| MsgId | Handler名称 | 功能描述 |
|-------|-------------|----------|
| 0 | HeartbeatHandler | 心跳消息处理，接收PING并返回PONG |
| 1 | EchoHandler | 回显功能，返回客户端发送的相同数据 |
| 2 | DataTransferHandler | 数据传输处理，将接收的字符串转为大写 |
| 3 | CloseConnectionHandler | 关闭连接请求处理 |
| 4 | StatusQueryHandler | 服务器状态查询 |

## 二、Writer的使用方式

在Go_Zinx中，Writer的使用非常简单，主要通过`Connection.SendMsg()`方法实现：

```go
// 发送消息的标准方式，内部使用了Writer
responseData := []byte("Hello, Client!")
err := conn.SendMsg(msgId, responseData)
if err != nil {
    // 错误处理
}
```

`SendMsg()`方法的工作原理：
1. 调用`DataPack.Pack()`将消息打包成二进制格式
2. 将打包好的二进制数据发送到`MsgChan`通道
3. `StartWriter`协程从`MsgChan`接收数据并通过TCP连接发送

## 三、运行测试步骤

### 1. 启动服务器
```bash
go run server_full_example.go
```

服务器启动后会显示：
```
[Server] Registered all handlers successfully
[Server] Available MsgIds:
  - MsgId 0: Heartbeat (Ping/Pong)
  - MsgId 1: Echo (Send back received data)
  - MsgId 2: Data Transfer (Process and return data)
  - MsgId 3: Close Connection (Request to close connection)
  - MsgId 4: Status Query (Get server status)
[Server] Starting Go_Zinx Server...
```

### 2. 运行客户端测试
```bash
go run client_test.go
```

客户端会依次测试所有Handler，并显示结果：
```
Connected to server!

=== Testing Echo (MsgId=1) ===
Echo Response - MsgId: 1, Data: ECHO: Hello, Go_Zinx!

=== Testing Data Transfer (MsgId=2) ===
Transfer Response - MsgId: 2, Data: PROCESSED: THIS IS TEST DATA TO BE PROCESSED

=== Testing Heartbeat (MsgId=0) ===
Heartbeat Response - MsgId: 0, Data: PONG

=== Testing Status Query (MsgId=4) ===
Status Response - MsgId: 4, Data: {
    "server_name": "Go_Zinx_Server",
    "version": "1.0.0",
    "uptime": "5.123456789s",
    "active_connections": 1
}

=== Testing Close Connection (MsgId=3) ===
Close Response - MsgId: 3, Data: Connection closing...

All tests completed!
```

## 四、Writer的工作流程

```
1. Handler调用conn.SendMsg(msgId, data)
2. SendMsg()打包消息：DataLen + MsgId + Data
3. 将打包好的数据发送到MsgChan通道
4. StartWriter协程从MsgChan接收数据
5. StartWriter通过TCP连接发送数据到客户端
```

## 五、Handler的扩展方式

如果您需要添加更多的Handler，可以按照以下步骤：

1. 创建一个结构体实现`zinterface.IHandler`接口
2. 为结构体实现`PreHandle()`、`Handle()`、`PostHandle()`方法
3. 在服务器启动时使用`server.AddHandler(msgId, &YourHandler{})`注册

示例：
```go
// 自定义Handler
type YourHandler struct{}

func (h *YourHandler) PreHandle(request zinterface.IRequest) {
    // 预处理逻辑
}

func (h *YourHandler) Handle(request zinterface.IRequest) {
    // 核心处理逻辑
    request.GetConn().SendMsg(100, []byte("Custom Response"))
}

func (h *YourHandler) PostHandle(request zinterface.IRequest) {
    // 后处理逻辑
}

// 注册Handler
server.AddHandler(100, &YourHandler{})
```

## 六、注意事项

1. **MsgId唯一性**：确保每个Handler使用唯一的MsgId
2. **错误处理**：始终检查`SendMsg()`的返回错误
3. **性能考虑**：避免在Handler中执行耗时操作，建议使用工作池
4. **并发安全**：Handler中的代码需要考虑并发安全问题
