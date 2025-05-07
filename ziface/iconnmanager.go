package ziface

/*
	连接管理模块的抽象层
*/
type IConnManager interface {
	//添加连接
	Add(conn IConnection)
	//删除连接
	Remove(conn IConnection)
	//根据connID获取连接
	Get(connID uint32) (IConnection, error)
	//得到当前连接总数
	Size() int
	//清除并终止所有连接
	ClearConns()
}