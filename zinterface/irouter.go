package zinterface

type IRouter interface {
	// preHandle
	PreHandle(request IRequest)
	// handle
	Handle(request IRequest)
	// postHandle
	PostHandle(request IRequest)
}
