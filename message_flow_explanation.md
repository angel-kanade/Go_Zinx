# 客户端发送消息与MsgRouter工作原理详解

## 一、客户端发送消息的完整流程

### 1. 客户端准备工作
- **建立TCP连接**：客户端通过`net.Dial("tcp", "127.0.0.1:8888")`与服务器建立TCP连接
- **准备消息内容**：确定要发送的消息ID和消息体内容

### 2. 消息打包（自定义二进制协议）

客户端必须按照服务器定义的协议格式打包消息：

```
+------------------+------------------+------------------+
| 数据长度(4字节)  |   消息ID(4字节)  |      消息体      |
+------------------+------------------+------------------+
```

**打包步骤**：
```go
// 1. 准备消息ID和内容
messageID := uint32(1)
messageData := []byte("Hello Zinx Server!")

// 2. 创建缓冲区
msgLen := uint32(len(messageData))
dataBuff := bytes.NewBuffer([]byte{})

// 3. 写入数据长度（LittleEndian编码）
binary.Write(dataBuff, binary.LittleEndian, msgLen)

// 4. 写入消息ID
binary.Write(dataBuff, binary.LittleEndian, messageID)

// 5. 写入消息体
binary.Write(dataBuff, binary.LittleEndian, messageData)
```

### 3. 发送消息
```go
// 通过TCP连接发送打包好的二进制数据
conn.Write(dataBuff.Bytes())
```

## 二、服务器接收与处理消息流程

### 1. 连接建立后的准备
当客户端连接成功后，服务器会：
- 为该连接创建`Connection`实例
- 启动独立的读协程(`StartReader`)和写协程(`StartWriter`)

### 2. 读协程接收消息

读协程不断从TCP连接读取数据：

```go
// 循环读取客户端消息
for {
    // 1. 读取消息头（固定8字节：数据长度+消息ID）
    headData := make([]byte, dp.GetHeadLen())
    io.ReadFull(c.GetTCPConnection(), headData)
    
    // 2. 解析消息头，获取消息长度和ID
    msg, _ := dp.Unpack(headData)
    
    // 3. 根据消息长度读取消息体
    if msg.GetDataLen() > 0 {
        data := make([]byte, msg.GetDataLen())
        io.ReadFull(c.GetTCPConnection(), data)
        msg.SetData(data)
    }
    
    // 4. 构造Request对象
    req := &Request{conn: c, msg: msg}
    
    // 5. 提交到工作池或直接处理
    server.WorkerPool.AddRequest(req)  // 或直接调用c.Router.DoMsgHandler(req)
}
```

## 三、MsgRouter的作用与工作原理

### 1. MsgRouter是什么？
`MsgRouter`是消息路由分发器，它的核心作用是**根据消息ID将请求分发给对应的处理器**。

### 2. MsgRouter的实现结构

```go
// MsgRouter结构体定义
type MsgRouter struct {
    Apis map[uint32]zinterface.IHandler  // 消息ID到处理器的映射表
}
```

### 3. 消息处理器注册

在服务器启动时，开发者需要为不同的消息ID注册对应的处理器：

```go
// 1. 创建处理器实例
helloHandler := &HelloHandler{}
heartbeatHandler := &HeartbeatHandler{}

// 2. 注册到消息路由器
server.AddHandler(1, helloHandler)  // 消息ID=1 -> HelloHandler处理
server.AddHandler(0, heartbeatHandler)  // 消息ID=0 -> 心跳处理器
```

### 4. 消息路由分发流程

当服务器收到消息后：

1. **构造Request对象**：封装连接和消息
2. **调用路由器处理**：`router.DoMsgHandler(req)`
3. **查找处理器**：根据消息ID在`Apis`映射表中查找对应的处理器
4. **执行处理器**：依次调用处理器的三个方法：
   - `PreHandle(req)`：预处理（如参数校验、日志记录）
   - `Handle(req)`：核心业务处理
   - `PostHandle(req)`：后处理（如清理资源、结果统计）

### 5. 示例：消息处理流程

```
客户端发送：
+------------------+------------------+------------------+
| 数据长度=17      |   消息ID=1       | "Hello Server"   |
+------------------+------------------+------------------+

服务器处理：
1. 读取并解析消息头，得到消息ID=1，数据长度=17
2. 读取17字节的消息体："Hello Server"
3. 构造Request对象
4. 调用router.DoMsgHandler(req)
5. 在Apis映射表中查找ID=1对应的处理器HelloHandler
6. 执行HelloHandler.PreHandle(req)
7. 执行HelloHandler.Handle(req)  // 核心处理逻辑
8. 执行HelloHandler.PostHandle(req)
```

## 四、消息ID的作用

消息ID是整个系统的核心标识，它：

1. **区分不同类型的消息**：每个业务功能对应一个唯一的消息ID
2. **实现消息路由**：MsgRouter根据消息ID找到对应的处理器
3. **简化通信协议**：客户端和服务器通过统一的ID进行通信，无需关心具体实现

## 五、消息处理的三种模式

1. **工作池模式**：请求被发送到工作池，由工作线程处理（推荐用于高并发场景）
2. **协程模式**：直接为每个请求创建一个新的Go协程处理
3. **串行模式**：在当前读协程中直接处理（不推荐，会阻塞其他消息）

## 六、实例解析：心跳消息处理

### 1. 心跳消息定义
- **消息ID**：0
- **消息内容**："ping"（客户端发送）或"pong"（服务器响应）

### 2. 心跳处理器实现

```go
type HeartbeatHandler struct {
    BaseHandler
}

func (h *HeartbeatHandler) Handle(request zinterface.IRequest) {
    // 1. 简单响应"pong"
    heartbeatMsg := znet.CreateHeartbeatMsg()
    heartbeatMsg.SetData([]byte("pong"))
    request.GetConnection().SendMsg(heartbeatMsg.GetMsgId(), heartbeatMsg.GetData())
    
    // 2. 更新连接活动时间
    conn := request.GetConnection()
    if checker, err := conn.GetProperty("HeartbeatChecker"); err == nil {
        if hc, ok := checker.(*HeartbeatChecker); ok {
            hc.UpdateActiveTime(conn.GetConnId())
        }
    }
}
```

### 3. 心跳处理流程

```
客户端发送心跳包：消息ID=0，内容="ping"
1. 服务器接收并解析消息
2. MsgRouter根据ID=0找到HeartbeatHandler
3. 执行HeartbeatHandler.Handle()
4. 服务器返回"pong"响应
5. 更新连接的最后活动时间
```

## 七、总结

### 客户端发送消息步骤：
1. 建立TCP连接
2. 按照协议格式打包消息（数据长度+消息ID+消息体）
3. 通过TCP连接发送二进制数据

### MsgRouter的核心作用：
- **解耦**：将消息接收与处理逻辑分离
- **路由**：根据消息ID分发请求
- **扩展**：方便添加新的消息处理器
- **中间件**：支持预处理和后处理

通过这种机制，服务器可以高效地处理来自不同客户端的各种类型的消息，实现了请求与处理的解耦，提高了系统的可扩展性和维护性。
