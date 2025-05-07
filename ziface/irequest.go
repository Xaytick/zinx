package ziface

/*
	将请求的一个连接和数据封装到一个Request中，
*/

type IRequest interface {
	// 得到当前连接
	GetConnection() IConnection
	// 得到请求的消息数据
	GetData() []byte
	// 得到请求的消息的ID
	GetMsgID() uint32
}