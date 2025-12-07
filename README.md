# Go_Zinx

一个基于Go语言开发的轻量级、高性能网络框架，是基于刘丹冰老师的[Zinx-V1.0](https://www.bilibili.com/video/BV1wE411d7th/?spm_id_from=333.337.search-card.all.click)拓展开发（在此感谢！），专注于TCP通信场景，提供简洁易用的API和丰富的功能组件。


## 我做了什么
- 修改了项目架构及部分名称，更符合个人习惯
- 解决了原先代码中部分并发安全问题
- 完善 Worker 工作池的创建和管理逻辑，添加了动态调整工作线程数的功能
- 新增了心跳检测功能
- 实现了日志功能，支持控制台和文件日志输出，并支持简单的性能监控


## 项目特点

- **高性能**：基于Go协程和通道实现高效的并发处理
- **易用性**：简洁的API设计，快速上手开发网络应用
- **可扩展**：模块化设计，支持自定义消息处理器和扩展功能
- **稳定性**：完善的错误处理和资源管理机制
- **功能丰富**：内置心跳检测、连接管理、工作池等核心组件



### 简单服务器示例

```go
package main

import (
    "Go_Zinx/znet"
    "Go_Zinx/zinterface"
)

// EchoHandler 回显处理器
type EchoHandler struct {
    znet.BaseHandler
}

func (h *EchoHandler) Handle(req zinterface.IRequest) {
    // 获取请求数据
    data := req.GetMsgData()
    // 回显数据
    req.GetConnection().SendMsg(1, data)
}

func main() {
    // 创建服务器实例
    server := znet.NewServer()
    
    // 注册消息处理器
    server.AddHandler(1, &EchoHandler{})
    
    // 启动服务器
    server.Serve()
}
```

### 简单客户端示例

```go
package main

import (
    "fmt"
    "net"
    "time"
)

func main() {
    // 连接服务器
    conn, err := net.Dial("tcp", "127.0.0.1:8888")
    if err != nil {
        fmt.Println("连接失败:", err)
        return
    }
    defer conn.Close()
    
    // 发送消息
    msg := []byte("Hello Zinx Server!")
    conn.Write([]byte{0, 12, 0, 1})
    conn.Write(msg)
    
    // 接收响应
    buf := make([]byte, 1024)
    n, err := conn.Read(buf)
    if err != nil {
        fmt.Println("接收失败:", err)
        return
    }
    fmt.Println("收到响应:", string(buf[4:n]))
}
```

## 核心组件

### 1. Server
- 服务器核心组件，负责监听端口、接受连接
- 管理连接、消息路由和工作池
- 提供连接建立/关闭的钩子函数

### 2. Connection
- 封装TCP连接，提供读写协程
- 支持属性存储和连接状态管理
- 处理消息的接收、解析和发送

### 3. MsgRouter
- 消息路由组件，根据消息ID分发请求
- 支持三段式处理流程：PreHandle → Handle → PostHandle

### 4. Handler
- 消息处理器接口，定义了消息处理的三个阶段
- 可通过嵌入BaseHandler实现自定义处理器

### 5. HeartbeatChecker
- 心跳检测组件，定期检查连接活跃度
- 自动关闭超时未活动的连接
- 支持更新连接活动时间

### 6. WorkerPool
- 工作池组件，管理工作线程
- 支持核心线程和最大线程数配置
- 自动回收空闲线程，优化资源使用

## 完整示例

查看 `server_full_example.go` 和 `client_example.go` 获取完整的服务器和客户端示例代码。

## 配置说明

配置文件位于 `resource/config.json`，支持以下配置项：

```json
{
  "Name": "ZinxServerApp",
  "Host": "0.0.0.0",
  "TCPPort": 8888,
  "MaxConn": 1000,
  "WorkerPool": {
    "CoreWorkers": 4,
    "MaxWorkers": 8,
    "QueueSize": 100,
    "IdleTimeout": 60
  }
}
```

## 性能指标

框架内置性能监控功能，每分钟输出一次性能报告，包括：
- 总连接数
- 当前连接数
- 消息处理延迟
- 工作池状态

## 项目结构

```
├── znet/           # 核心框架代码
│   ├── server.go      # 服务器实现
│   ├── connection.go  # 连接管理
│   ├── msgRouter.go   # 消息路由
│   ├── handler.go     # 处理器接口
│   ├── heartbeat.go   # 心跳检测
│   ├── workerpool.go  # 工作池
│   └── message.go     # 消息定义
├── zinterface/     # 接口定义
├── utils/          # 工具函数
│   ├── datapack.go    # 数据打包/解包
│   ├── globalobj.go   # 全局配置
│   └── logger.go      # 日志工具
├── resource/       # 资源文件
│   └── config.json    # 配置文件
├── server_example.go      # 简单服务器示例
├── server_full_example.go # 完整服务器示例
├── client_example.go      # 客户端示例
└── README.md             # 项目说明文档
```

