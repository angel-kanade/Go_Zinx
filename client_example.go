package main

import (
	"Go_Zinx/utils"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// 发送消息函数
func sendMessage(conn net.Conn, msgId uint32, data []byte) error {
	// 按照自定义二进制协议打包消息
	msgLen := uint32(len(data))
	dataBuff := bytes.NewBuffer([]byte{})

	// 写入数据长度
	binary.Write(dataBuff, binary.LittleEndian, msgLen)
	// 写入消息ID
	binary.Write(dataBuff, binary.LittleEndian, msgId)
	// 写入消息体
	binary.Write(dataBuff, binary.LittleEndian, data)

	// 发送打包好的消息
	_, err := conn.Write(dataBuff.Bytes())
	return err
}

// 接收消息函数
func receiveMessage(conn net.Conn) ([]byte, error) {
	// 1. 读取8字节的消息头
	header := make([]byte, 8)
	_, err := conn.Read(header)
	if err != nil {
		return nil, err
	}

	// 2. 解析消息头，获取消息长度
	dp := utils.NewDataPackUtil()
	msg, err := dp.Unpack(header)
	if err != nil {
		return nil, err
	}

	// 3. 如果有消息体，读取消息体
	if msg.GetDataLen() > 0 {
		data := make([]byte, msg.GetDataLen())
		_, err := conn.Read(data)
		if err != nil {
			return nil, err
		}
		msg.SetData(data)
	}

	return msg.GetData(), nil
}

// 客户端发送消息示例
func main() {
	// 1. 与服务器建立TCP连接
	conn, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		fmt.Println("连接服务器失败:", err)
		return
	}
	defer conn.Close()

	fmt.Println("成功连接到服务器")

	// 2. 测试心跳消息 (MsgId=0)
	fmt.Println("\n=== 测试心跳消息 (MsgId=0) ===")
	if err := sendMessage(conn, 0, []byte("ping")); err != nil {
		fmt.Println("发送心跳消息失败:", err)
		return
	}

	if response, err := receiveMessage(conn); err == nil {
		fmt.Printf("收到心跳响应: %s\n", string(response))
	}

	// 3. 测试回显消息 (MsgId=1)
	fmt.Println("\n=== 测试回显消息 (MsgId=1) ===")
	echoMsg := []byte("Hello Zinx Server!")
	if err := sendMessage(conn, 1, echoMsg); err != nil {
		fmt.Println("发送回显消息失败:", err)
		return
	}

	if response, err := receiveMessage(conn); err == nil {
		fmt.Printf("收到回显响应: %s\n", string(response))
	}

	// 4. 测试数据传输消息 (MsgId=2)
	fmt.Println("\n=== 测试数据传输消息 (MsgId=2) ===")
	dataMsg := []byte("Request: Process this data")
	if err := sendMessage(conn, 2, dataMsg); err != nil {
		fmt.Println("发送数据传输消息失败:", err)
		return
	}

	if response, err := receiveMessage(conn); err == nil {
		fmt.Printf("收到数据传输响应: %s\n", string(response))
	}

	// 5. 测试状态查询消息 (MsgId=4)
	fmt.Println("\n=== 测试状态查询消息 (MsgId=4) ===")
	if err := sendMessage(conn, 4, []byte("getStatus")); err != nil {
		fmt.Println("发送状态查询消息失败:", err)
		return
	}

	if response, err := receiveMessage(conn); err == nil {
		// 格式化JSON输出
		var statusMap map[string]interface{}
		if err := json.Unmarshal(response, &statusMap); err == nil {
			if prettyJSON, err := json.MarshalIndent(statusMap, "", "  "); err == nil {
				fmt.Printf("收到状态查询响应:\n%s\n", string(prettyJSON))
			} else {
				fmt.Printf("收到状态查询响应: %s\n", string(response))
			}
		} else {
			fmt.Printf("收到状态查询响应: %s\n", string(response))
		}
	}

	// 6. 测试关闭连接消息 (MsgId=3)
	fmt.Println("\n=== 测试关闭连接消息 (MsgId=3) ===")
	if err := sendMessage(conn, 3, []byte("closeConnection")); err != nil {
		fmt.Println("发送关闭连接消息失败:", err)
		return
	}

	fmt.Println("关闭连接请求已发送")

	// 等待服务器响应
	time.Sleep(1 * time.Second)
}
