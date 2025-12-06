package utils

import (
	"Go_Zinx/zinterface"
	"bytes"
	"encoding/binary"
	"errors"
)

type DataPack struct {
}

// message 是一个实现了zinterface.IMessage接口的结构体
// 用于在utils包内部处理消息打包和解包
// 避免直接依赖znet包的Message结构体
// 实现zinterface.IMessage接口
type message struct {
	Id      uint32
	DataLen uint32
	Data    []byte
}

func (m *message) GetMsgId() uint32 {
	return m.Id
}

func (m *message) GetDataLen() uint32 {
	return m.DataLen
}

func (m *message) GetData() []byte {
	return m.Data
}

func (m *message) SetMsgId(id uint32) {
	m.Id = id
}

func (m *message) SetDataLen(len uint32) {
	m.DataLen = len
}

func (m *message) SetData(data []byte) {
	m.Data = data
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

	// 创建一个实现了zinterface.IMessage接口的结构体
	msg := &message{Id: 0, DataLen: 0, Data: nil}

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
