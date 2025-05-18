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
	cm.userLock.Lock() // Lock userToConn as well
	defer cm.connLock.Unlock()
	defer cm.userLock.Unlock()

	for connID, conn := range cm.connections {
		conn.Stop()
		delete(cm.connections, connID)
	}
	cm.userToConn = make(map[uint]uint32) // Clear userToConn map

	// Simplified logging
	fmt.Println("Clear All Connections and User Mappings successfully: conn num = ", len(cm.connections), ", user map size = ", len(cm.userToConn))
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
	_, connExists := cm.connections[connID]
	cm.connLock.RUnlock()

	if !connExists {
		fmt.Printf("SetConnByUserID: ConnID %d does not exist. Cannot associate userID %d.\n", connID, userID)
		return
	}

	cm.userLock.Lock()
	defer cm.userLock.Unlock()
	// Optional: Handle if userID is already mapped to another connID (e.g. single login policy)
	// For now, simply overwrite.
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
		return nil // UserID not found in map
	}

	// Now get the connection using the connID
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()
	conn, connOk := cm.connections[connID]
	if !connOk {
		// This case should ideally not happen if Remove() correctly cleans userToConn.
		// But as a safeguard:
		fmt.Printf("GetConnByUserID: Found connID %d for userID %d, but connID not in connections map.\n", connID, userID)
		// Clean up the stale entry from userToConn
		// To do this safely without deadlocking, we might need a more complex lock or deferred unlock sequence.
		// For now, just returning nil. A more robust solution would handle this cleanup.
		return nil
	}
	return conn
}
