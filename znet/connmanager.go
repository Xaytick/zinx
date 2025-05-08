package znet

import (
	"fmt"
	"sync"

	"github.com/Xaytick/zinx/ziface"
)

type ConnManager struct {
	connections map[uint32]ziface.IConnection // 管理的连接集合

	connLock sync.RWMutex // 读写连接集合的读写锁

}

// 创建ConnManager
func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint32]ziface.IConnection),
	}
}

// 添加连接
func (cm *ConnManager) Add(conn ziface.IConnection) {
	// 保护共享资源map,加写锁
	cm.connLock.Lock()
	defer cm.connLock.Unlock()
	// 将conn加入到ConnManager中
	cm.connections[conn.GetConnID()] = conn
	fmt.Println("connID = ", conn.GetConnID(),
		"add to ConnManager successfully: conn num = ", cm.Size())
}

// 删除连接
func (cm *ConnManager) Remove(conn ziface.IConnection) {
	// 保护共享资源map,加写锁
	cm.connLock.Lock()
	defer cm.connLock.Unlock()
	// 删除连接信息
	delete(cm.connections, conn.GetConnID())
	fmt.Println("connID = ", conn.GetConnID(),
		"remove successfully: conn num = ", cm.Size())
}

// 根据connID获取连接
func (cm *ConnManager) Get(connID uint32) (ziface.IConnection, error) {
	// 保护共享资源map,加读锁
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()

	if conn, ok := cm.connections[connID]; ok {
		return conn, nil
	} else {
		return nil, fmt.Errorf("connID = %d is not exist", connID)
	}

}

// 得到当前连接总数
func (cm *ConnManager) Size() int {
	return len(cm.connections)
}

// 清除并终止所有连接
func (cm *ConnManager) ClearConns() {
	// 保护共享资源map,加写锁
	cm.connLock.Lock()
	defer cm.connLock.Unlock()
	// 停止并删除全部的连接信息
	for connID, conn := range cm.connections {
		// 停止
		conn.Stop()
		// 删除
		delete(cm.connections, connID)
	}
	if cm.Size() == 0 {
		fmt.Println("Clear All Connections successfully: conn num = ", cm.Size())
	} else {
		fmt.Println("Clear All Connections failed: conn num = ", cm.Size())
	}

}

// 获取所有连接
func (cm *ConnManager) All() []ziface.IConnection {
    cm.connLock.RLock()
    defer cm.connLock.RUnlock()
    conns := make([]ziface.IConnection, 0, len(cm.connections))
    for _, conn := range cm.connections {
        conns = append(conns, conn)
    }
    return conns
}
