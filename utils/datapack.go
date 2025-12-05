package utils

import (
	"Go_Zinx/zinterface"
	"Go_Zinx/znet"
	"bytes"
	"encoding/binary"
	"errors"
)

type DataPack struct {
}

func NewDataPackUtil() *DataPack {
	return &DataPack{}
}

func (dp *DataPack) GetHeadLen() uint32 {
	return 4 + 4
}

func (dp *DataPack) Pack(msg zinterface.IMessage) ([]byte, error) {
	// 定义一个缓冲
	dataBuff := bytes.NewBuffer([]byte{})

	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetDataLen()); err != nil {
		return nil, err
	}
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgId()); err != nil {
		return nil, err
	}
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetData()); err != nil {
		return nil, err
	}

	return dataBuff.Bytes(), nil
}

func (dp *DataPack) Unpack(data []byte) (zinterface.IMessage, error) {
	// 创建一个输入二进制数据的IO-Reader
	dataBuff := bytes.NewReader(data)

	msg := &znet.Message{}

	// 先读Head-Len
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}

	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.Id); err != nil {
		return nil, err
	}

	if GlobalObject.MaxPackageSize > 0 && msg.DataLen > GlobalObject.MaxPackageSize {
		return nil, errors.New("msg Data Is Too Large")
	}

	return msg, nil
}
