package znet

import (
	"Go_Zinx/zinterface"
)

// 定义BaseHandler，以后实现Handler时，先嵌入BaseHandler基类，然后重写就好
// 类似Java中Abstract类
type BaseHandler struct {
}

func (br *BaseHandler) PreHandle(req zinterface.IRequest) {

}

func (br *BaseHandler) Handle(req zinterface.IRequest) {

}

func (br *BaseHandler) PostHandle(req zinterface.IRequest) {

}
