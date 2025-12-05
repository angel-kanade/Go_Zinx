package znet

import (
	"Go_Zinx/zinterface"
)

// 定义BaseRouter，以后实现Router时，先嵌入BaseRouter基类，然后重写就好
// 类似Java中Abstract类
type BaseRouter struct {
}

func (br *BaseRouter) PreHandle(req zinterface.IRequest) {

}

func (br *BaseRouter) Handle(req zinterface.IRequest) {

}

func (br *BaseRouter) PostHandle(req zinterface.IRequest) {

}
