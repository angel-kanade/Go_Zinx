# MsgRouter 与 Handler 注册说明

## 一、MsgRouter 中的 Handler 由谁管理？

**是的，MsgRouter 中的所有 Handler 都是由服务器注册和维护的**。

### 注册流程：
1. **服务器创建**：`server := znet.NewServer()`
2. **定义 Handler**：实现 `zinterface.IHandler` 接口（PreHandle/Handle/PostHandle）
3. **注册 Handler**：通过 `server.AddHandler(msgId, handler)` 方法注册

### 代码示例：
```go
// 定义消息处理器
type ReadOnlyHandler struct{}

func (r *ReadOnlyHandler) PreHandle(request zinterface.IRequest) { /*...*/ }
func (r *ReadOnlyHandler) Handle(request zinterface.IRequest) { /*...*/ }
func (r *ReadOnlyHandler) PostHandle(request zinterface.IRequest) { /*...*/ }

// 注册处理器到服务器
server.AddHandler(1, &ReadOnlyHandler{})  // msgId=1 -> ReadOnlyHandler
```

## 二、MsgId 的定义与使用

**是的，服务器会预先定义各种 MsgId**，用于区分不同类型的消息。

### 常见的 MsgId 定义：
| MsgId | 消息类型 | 用途 |
|-------|----------|------|
| 0     | 心跳消息 | 维持连接活性，检测超时 |
| 1     | 只读消息 | 客户端请求读取数据 |
| 2     | 写操作消息 | 客户端请求写入数据 |
| 3+    | 自定义消息 | 根据业务需求定义 |

### 消息处理流程：
1. 客户端发送包含 MsgId 的消息
2. 服务器接收并解析消息头，获取 MsgId
3. MsgRouter 根据 MsgId 查找对应的 Handler
4. 依次执行 Handler 的 PreHandle → Handle → PostHandle 方法

## 三、Handler 的生命周期

每个 Handler 都有三个处理阶段，便于实现统一的处理逻辑：

1. **PreHandle**：消息的预处理（如日志记录、参数验证）
2. **Handle**：消息的核心处理逻辑
3. **PostHandle**：消息的后处理（如结果统计、资源清理）

## 四、示例：完整的消息处理链路

```
客户端发送：
[消息头: DataLen=10, MsgId=1] [消息体: "read_data"]

服务器处理：
1. 接收并解析消息 → MsgId=1
2. MsgRouter 查找 → ReadOnlyHandler
3. 执行 ReadOnlyHandler.PreHandle() → 记录日志
4. 执行 ReadOnlyHandler.Handle() → 处理读请求
5. 执行 ReadOnlyHandler.PostHandle() → 统计请求
6. 返回响应 → [消息头: DataLen=20, MsgId=1] [消息体: "read_data_result"]
```

## 五、注意事项

1. **MsgId 的唯一性**：每个 MsgId 只能注册一个 Handler
2. **消息格式统一**：客户端和服务器必须使用相同的消息格式和 MsgId 定义
3. **心跳消息**：MsgId=0 通常预留给心跳消息，由心跳检测器自动处理
4. **业务扩展性**：可以根据业务需求定义任意数量的自定义 MsgId
