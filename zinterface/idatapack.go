package zinterface

// 封包、拆包的接口
type IDataPack interface {
	GetHeadLen() uint32

	Pack(msg IMessage) ([]byte, error)

	Unpack([]byte) (IMessage, error)
}
