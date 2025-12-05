package zinterface

type IHandler interface {
	// preHandle
	PreHandle(request IRequest)
	// handle
	Handle(request IRequest)
	// postHandle
	PostHandle(request IRequest)
}
