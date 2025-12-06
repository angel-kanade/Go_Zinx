package zinterface

type IMsgRouter interface {
	DoMsgHandler(req IRequest)

	AddHandler(msgId uint32, handler IHandler)

	StartWorkerPool()

	SendMsgToTaskQueue(request IRequest)
}
