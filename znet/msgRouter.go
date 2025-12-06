package znet

import (
	"Go_Zinx/zinterface"
	"fmt"
	"strconv"
)

type MsgRouter struct {
	Apis map[uint32]zinterface.IHandler
}

func NewMsgRouter() *MsgRouter {
	return &MsgRouter{Apis: make(map[uint32]zinterface.IHandler)}
}

// 调度执行对应的消息处理方法
func (m *MsgRouter) DoMsgHandler(req zinterface.IRequest) {
	id := req.GetMsgID()

	handler, ok := m.Apis[id]
	if !ok {
		fmt.Println("api msgId =", req.GetMsgID(), "is NOT FOUND!")
	}

	handler.PreHandle(req)
	handler.Handle(req)
	handler.PostHandle(req)

}

// 添加具体逻辑
func (m *MsgRouter) AddHandler(msgId uint32, handler zinterface.IHandler) {
	if _, ok := m.Apis[msgId]; ok {
		// 已经注册
		panic("repeat api, msgId =" + strconv.Itoa(int(msgId)))
	}

	m.Apis[msgId] = handler
	fmt.Println("Add api MsgId =", msgId, "handler =", handler)
}
