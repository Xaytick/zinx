package znet

import (
	"bytes"
	"encoding/binary"

	"github.com/Xaytick/zinx/utils"
	"github.com/Xaytick/zinx/ziface"

	"github.com/pkg/errors"
)

/*
用于处理TCP粘包问题，
面向TCP连接的数据流，通过封包和解包来解决
测试结果：
=== RUN   TestDataPack
--- PASS: TestDataPack (0.00s)
----> Recv Msg: ID =  1 , len =  5 , data =  hello
----> Recv Msg: ID =  2 , len =  7 , data =  world!!
PASS
ok      zinx/znet       0.639s
*/
type DataPack struct {
}

func NewDataPack() *DataPack {
	return &DataPack{}
}

func (dp *DataPack) GetHeadLen() uint32 {
	//DataLen uint32(4字节) + ID uint32(4字节)
	return 8
}

func (dp *DataPack) Pack(msg ziface.IMessage) ([]byte, error) {
	//创建一个存放bytes字节的缓冲
	databuf := bytes.NewBuffer([]byte{})

	// 将datalen写进databuf中
	if err := binary.Write(databuf, binary.LittleEndian, msg.GetMsgLen()); err != nil {
		return nil, err
	}

	// 将MsgId写进databuf中
	if err := binary.Write(databuf, binary.LittleEndian, msg.GetMsgId()); err != nil {
		return nil, err
	}

	// 将data数据写进databuf中
	if err := binary.Write(databuf, binary.LittleEndian, msg.GetData()); err != nil {
		return nil, err
	}

	return databuf.Bytes(), nil
}

// 拆包方法(将包的head信息读出来)之后再根据head信息里的data长度，再进行一次读包，将data读出来
func (dp *DataPack) Unpack(binaryData []byte) (ziface.IMessage, error) {
	// 创建一个从输入二进制数据的ioReader
	databuf := bytes.NewReader(binaryData)
	// 只解压head的信息，得到dataLen和msgID
	msg := &Message{}
	// 读dataLen
	if err := binary.Read(databuf, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}

	// 读msgID
	if err := binary.Read(databuf, binary.LittleEndian, &msg.Id); err != nil {
		return nil, err
	}

	// 判断DataLen是否已经超出了我们允许的最大包长度
	if utils.GlobalObject.MaxPackageSize > 0 && msg.DataLen > utils.GlobalObject.MaxPackageSize {
		return nil, errors.New("Too large msg data received")
	}

	return msg, nil
}
