package znet

import (
	"Go_Zinx/utils"
	"Go_Zinx/zinterface"
	"fmt"
	"strconv"
)

type MsgRouter struct {
	Apis           map[uint32]zinterface.IHandler
	TaskQueue      []chan zinterface.IRequest
	WorkerPoolSize uint32
}

func NewMsgRouter() *MsgRouter {
	return &MsgRouter{
		Apis:           make(map[uint32]zinterface.IHandler),
		TaskQueue:      make([]chan zinterface.IRequest, utils.GlobalObject.WorkerPoolSize),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize, // 从全局配置中获取
	}
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

// 启动一个Worker工作池
func (m *MsgRouter) StartWorkerPool() {
	// 根据WorkerPoolSize
	for i := 0; i < int(m.WorkerPoolSize); i++ {
		// init channel
		m.TaskQueue[i] = make(chan zinterface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)

		// start worker
		go m.startOneWorker(i, m.TaskQueue[i])
	}
}

// 启动一个Worker工作流程
func (m *MsgRouter) startOneWorker(workerId int, taskQueue chan zinterface.IRequest) {
	fmt.Println("Worker Id =", workerId, "is starting...")
	for {
		select {
		case req := <-taskQueue:
			m.DoMsgHandler(req)

		}
	}
}

func (m *MsgRouter) SendMsgToTaskQueue(request zinterface.IRequest) {
	// 将消息平均分配给不同的Worker
	// 根据客户端建立的ConnId来进行分配
	workerId := request.GetConnection().GetConnId() % m.WorkerPoolSize
	fmt.Println("Add ConnID =", request.GetConnection().GetConnId(),
		"request MsgId =", request.GetMsgID(),
		"to WorkerId =", workerId)

	// 将消息发送给对应的TaskQueue
	m.TaskQueue[workerId] <- request
}
