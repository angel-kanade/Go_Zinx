package zinterface

type IRequest interface {
	GetConnection() IConnection

	GetData() []byte
}
