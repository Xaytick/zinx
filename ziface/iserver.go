package ziface

type IServer interface {
	// 启动服务器
	Start()
	// 停止服务器
	Stop()
	// 运行服务器
	Serve()
	// 路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
	AddRouter(msgID uint32, router IRouter)
	// 获取连接管理器
	GetConnManager() IConnManager
	// 设置该Server的连接创建时Hook函数
	SetOnConnStart(func(conn IConnection))
	// 设置该Server的连接断开时的Hook函数
	SetOnConnStop(func(conn IConnection))
	// 调用连接OnConnStart Hook函数
	CallOnConnStart(conn IConnection)
	// 调用连接OnConnStop Hook函数
	CallOnConnStop(conn IConnection)
}