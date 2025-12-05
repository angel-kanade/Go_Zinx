package main

import "Go_Zinx/znet"

func main() {
	// 创建一个服务器
	s := znet.NewServer()

	// 启动服务器
	s.Serve()
}
