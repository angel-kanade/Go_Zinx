package main

import (
	"Go_Zinx/utils"
	"Go_Zinx/znet"
	"fmt"
	"net"
	"time"
)

// 按自定义二进制协议打包消息
func packMessage(msgId uint32, data []byte) ([]byte, error) {
	dp := utils.NewDataPackUtil()
	msg := znet.NewMsgPackage(msgId, data)
	return dp.Pack(msg)
}

func main() {
	// 1. 建立TCP连接
	conn, err := net.Dial("tcp", "127.0.0.1:8999")
	if err != nil {
		fmt.Println("Failed to connect to server:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to server!")

	// 2. 测试回显功能 (MsgId=1)
	fmt.Println("\n=== Testing Echo (MsgId=1) ===")
	echoData := []byte("Hello, Go_Zinx!")
	echoMsg, err := packMessage(1, echoData)
	if err != nil {
		fmt.Println("Failed to pack echo message:", err)
		return
	}
	_, err = conn.Write(echoMsg)
	if err != nil {
		fmt.Println("Failed to send echo message:", err)
		return
	}
	
	// 接收回显响应
	echoResp := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	readLen, err := conn.Read(echoResp)
	if err != nil {
		fmt.Println("Failed to receive echo response:", err)
		return
	}
	
	// 解析响应
	dp := utils.NewDataPackUtil()
	respMsg, err := dp.Unpack(echoResp[:readLen])
	if err != nil {
		fmt.Println("Failed to unpack echo response:", err)
		return
	}
	fmt.Printf("Echo Response - MsgId: %d, Data: %s\n", respMsg.GetMsgId(), string(respMsg.GetData()))

	// 3. 测试数据传输功能 (MsgId=2)
	fmt.Println("\n=== Testing Data Transfer (MsgId=2) ===")
	transferData := []byte("this is test data to be processed")
	transferMsg, err := packMessage(2, transferData)
	if err != nil {
		fmt.Println("Failed to pack transfer message:", err)
		return
	}
	_, err = conn.Write(transferMsg)
	if err != nil {
		fmt.Println("Failed to send transfer message:", err)
		return
	}
	
	// 接收传输响应
	transferResp := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	readLen, err = conn.Read(transferResp)
	if err != nil {
		fmt.Println("Failed to receive transfer response:", err)
		return
	}
	
	respMsg, err = dp.Unpack(transferResp[:readLen])
	if err != nil {
		fmt.Println("Failed to unpack transfer response:", err)
		return
	}
	fmt.Printf("Transfer Response - MsgId: %d, Data: %s\n", respMsg.GetMsgId(), string(respMsg.GetData()))

	// 4. 测试心跳功能 (MsgId=0)
	fmt.Println("\n=== Testing Heartbeat (MsgId=0) ===")
	heartbeatMsg, err := packMessage(0, []byte("PING"))
	if err != nil {
		fmt.Println("Failed to pack heartbeat message:", err)
		return
	}
	_, err = conn.Write(heartbeatMsg)
	if err != nil {
		fmt.Println("Failed to send heartbeat message:", err)
		return
	}
	
	// 接收心跳响应
	heartbeatResp := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	readLen, err = conn.Read(heartbeatResp)
	if err != nil {
		fmt.Println("Failed to receive heartbeat response:", err)
		return
	}
	
	respMsg, err = dp.Unpack(heartbeatResp[:readLen])
	if err != nil {
		fmt.Println("Failed to unpack heartbeat response:", err)
		return
	}
	fmt.Printf("Heartbeat Response - MsgId: %d, Data: %s\n", respMsg.GetMsgId(), string(respMsg.GetData()))

	// 5. 测试状态查询功能 (MsgId=4)
	fmt.Println("\n=== Testing Status Query (MsgId=4) ===")
	statusMsg, err := packMessage(4, []byte("status"))
	if err != nil {
		fmt.Println("Failed to pack status message:", err)
		return
	}
	_, err = conn.Write(statusMsg)
	if err != nil {
		fmt.Println("Failed to send status message:", err)
		return
	}
	
	// 接收状态响应
	statusResp := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	readLen, err = conn.Read(statusResp)
	if err != nil {
		fmt.Println("Failed to receive status response:", err)
		return
	}
	
	respMsg, err = dp.Unpack(statusResp[:readLen])
	if err != nil {
		fmt.Println("Failed to unpack status response:", err)
		return
	}
	fmt.Printf("Status Response - MsgId: %d, Data: %s\n", respMsg.GetMsgId(), string(respMsg.GetData()))

	// 6. 测试关闭连接功能 (MsgId=3)
	fmt.Println("\n=== Testing Close Connection (MsgId=3) ===")
	closeMsg, err := packMessage(3, []byte("close"))
	if err != nil {
		fmt.Println("Failed to pack close message:", err)
		return
	}
	_, err = conn.Write(closeMsg)
	if err != nil {
		fmt.Println("Failed to send close message:", err)
		return
	}
	
	// 接收关闭确认
	closeResp := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	readLen, err = conn.Read(closeResp)
	if err != nil {
		fmt.Println("Failed to receive close response:", err)
		return
	}
	
	respMsg, err = dp.Unpack(closeResp[:readLen])
	if err != nil {
		fmt.Println("Failed to unpack close response:", err)
		return
	}
	fmt.Printf("Close Response - MsgId: %d, Data: %s\n", respMsg.GetMsgId(), string(respMsg.GetData()))

	fmt.Println("\nAll tests completed!")
}
