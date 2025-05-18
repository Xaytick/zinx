package znet

import (
	"fmt"
	"sync"

	"github.com/Xaytick/zinx/ziface"
)

type ConnManager struct {
	connections map[uint32]ziface.IConnection // 管理的连接集合
	connLock    sync.RWMutex                  // 读写连接集合的读写锁

	userToConn map[uint]uint32 // userID (uint) -> connID (uint32)
	userLock   sync.RWMutex    // 保护 userToConn
}

// 创建ConnManager
func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint32]ziface.IConnection),
		userToConn:  make(map[uint]uint32),
	}
}

// 添加连接
func (cm *ConnManager) Add(conn ziface.IConnection) {
	cm.connLock.Lock()
	cm.connections[conn.GetConnID()] = conn
	// 获取当前连接数，必须在释放写锁之前或持有读锁的情况下进行
	// 为避免在Println中调用Size()再次锁connLock，我们在这里直接获取长度
	currentSize := len(cm.connections)
	cm.connLock.Unlock() // 在调用 fmt.Println 和 cm.Size() 之前解锁

	fmt.Println("connID = ", conn.GetConnID(),
		"add to ConnManager successfully: conn num = ", currentSize) // 使用保存的长度值
}

// 删除连接
func (cm *ConnManager) Remove(conn ziface.IConnection) {
	// 保护共享资源map,加写锁
	cm.connLock.Lock()
	connID := conn.GetConnID()
	delete(cm.connections, connID)
	cm.connLock.Unlock() // Unlock before fmt.Println

	// Also remove from userToConn map
	userIDVal, err := conn.GetProperty("userID") // Assuming userID is set as a property on the connection
	if err == nil {
		if userID, ok := userIDVal.(uint); ok {
			cm.userLock.Lock()
			// Check if the current connID in map matches the one being removed for this userID
			if mappedConnID, userFound := cm.userToConn[userID]; userFound && mappedConnID == connID {
				delete(cm.userToConn, userID)
				fmt.Println("UserID = ", userID, " removed from userToConn map.")
			}
			cm.userLock.Unlock()
		}
	}

	fmt.Println("connID = ", connID,
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
	cm.connLock.RLock() // Use RLock for read-only operations like len()
	size := len(cm.connections)
	cm.connLock.RUnlock()
	return size
}

// 清除并终止所有连接
func (cm *ConnManager) ClearConns() {
	cm.connLock.Lock()
	// Create a slice of connections to stop AFTER releasing the lock
	connsToStop := make([]ziface.IConnection, 0, len(cm.connections))
	for _, conn := range cm.connections {
		connsToStop = append(connsToStop, conn)
	}
	// Clear the connections map while still holding the lock
	cm.connections = make(map[uint32]ziface.IConnection)
	cm.connLock.Unlock() // Release connLock before stopping connections

	cm.userLock.Lock()
	cm.userToConn = make(map[uint]uint32) // Clear userToConn map
	cm.userLock.Unlock()                  // Release userLock

	// Now stop each connection. This will call Remove, which will try to lock, but it should be fine now.
	for _, conn := range connsToStop {
		conn.Stop() // This will call Remove(), which will attempt its own locking.
	}

	fmt.Println("Clear All Connections and User Mappings initiated. Actual removal happens in conn.Stop().")
	// Note: The final count might be 0 after all Stop() calls complete if they are synchronous.
	// The log here reflects the state after maps are cleared but before all Stop() might have finished if async.
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

// 设置连接的UserID
func (cm *ConnManager) SetConnByUserID(connID uint32, userID uint) {
	// First, ensure the connection actually exists
	cm.connLock.RLock()
	_, ok := cm.connections[connID]
	cm.connLock.RUnlock()

	if !ok {
		fmt.Printf("SetConnByUserID: ConnID %d does not exist. Cannot associate userID %d.\n", connID, userID)
		return
	}

	cm.userLock.Lock()
	defer cm.userLock.Unlock()
	cm.userToConn[userID] = connID
	fmt.Printf("Associated UserID %d with ConnID %d\n", userID, connID)
}

// 清除特定UserID的映射
func (cm *ConnManager) ClearConnByUserID(userID uint) {
	cm.userLock.Lock()
	defer cm.userLock.Unlock()
	if _, ok := cm.userToConn[userID]; ok {
		delete(cm.userToConn, userID)
		fmt.Printf("Cleared UserID %d from userToConn map.\n", userID)
	} else {
		fmt.Printf("ClearConnByUserID: UserID %d not found in userToConn map.\n", userID)
	}
}

// 根据UserID获取连接 (Optimized with userToConn map)
func (cm *ConnManager) GetConnByUserID(userID uint) ziface.IConnection {
	cm.userLock.RLock()
	connID, ok := cm.userToConn[userID]
	cm.userLock.RUnlock()

	if !ok {
		return nil
	}

	cm.connLock.RLock()
	defer cm.connLock.RUnlock()
	conn, connOk := cm.connections[connID]
	if !connOk {
		fmt.Printf("GetConnByUserID: Found connID %d for userID %d, but connID not in connections map.\n", connID, userID)
		return nil
	}
	return conn
}
