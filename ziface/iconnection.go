package ziface

import "net"

type IConnection interface {
	
	// 启动连接，让当前连接开始工作
	Start()

	//停止连接，结束当前连接状态
	Stop()

	//获取当前连接绑定的socket conn
	GetTCPConnection() *net.TCPConn

	//获取当前连接模块的连接ID
	GetConnID() uint32

	//获取远程客户端的TCP状态 IP port
	RemoteAddr() net.Addr

	//发送数据，将数据发送给远程的客户端
	Send(data []byte) error

	//发送消息，将我们对客户端定义的消息进行发送
	SendMsg(msgId uint32, data []byte) error

	//设置连接属性
	SetProperty(key string, value interface{})

	//获取连接属性
	GetProperty(key string) (interface{}, error)

	//移除连接属性
	RemoveProperty(key string)
}

type HandleFunc func(*net.TCPConn, []byte, int) error