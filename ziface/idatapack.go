package ziface

/*
用于处理TCP粘包问题，
面向TCP连接的数据流，通过封包和解包来解决
*/

type IDataPack interface {
	// 获取包头的长度方法
	GetHeadLen() uint32
	// 封包方法
	Pack(msg IMessage) ([]byte, error)
	// 拆包方法
	Unpack([]byte) (IMessage, error)	

}