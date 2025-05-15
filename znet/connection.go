package znet

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/Xaytick/zinx/utils"
	"github.com/Xaytick/zinx/ziface"
)

type Connection struct {
	// 当前连接隶属于的Server
	TCPServer ziface.IServer
	// 当前连接的socket TCP套接字
	Conn *net.TCPConn
	// 当前连接的ID 也可以称作为SessionID，ID全局唯一
	ConnID uint32
	// 当前连接的关闭状态
	isClosed bool
	// 告知当前连接已经退出/停止的channel
	ExitChan chan bool
	// 无缓冲管道，用于读、写两个goroutine之间的消息通信
	msgChan chan []byte
	// 消息管理MsgId和对应处理方法的消息管理模块
	MsgHandler ziface.IMsgHandler
	// 连接属性集合
	property map[string]interface{}
	// 保护当前property的锁
	propertyLock sync.RWMutex
	// 最后一次活动时间
	lastActivityTime time.Time
}

func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandler) *Connection {

	c := &Connection{
		TCPServer:        server,
		Conn:             conn,
		ConnID:           connID,
		MsgHandler:       msgHandler,
		isClosed:         false,
		ExitChan:         make(chan bool, 1),
		msgChan:          make(chan []byte),
		property:         make(map[string]interface{}),
		lastActivityTime: time.Now(), // 初始化时记录当前时间
	}
	// 将新创建的Conn添加到链接管理中
	c.Register()
	return c
}

// 注册Conn到ConnManager中
func (c *Connection) Register() {
	// 将当前新连接添加到ConnManager中
	c.TCPServer.GetConnManager().Add(c)
}

// 从ConnManager中移除Conn
func (c *Connection) UnRegister() {
	// 将当前新连接添加到ConnManager中
	c.TCPServer.GetConnManager().Remove(c)
}

// 写消息Goroutine， 用户将数据发送给客户端
func (c *Connection) StartWriter() {
	fmt.Println("Writer Goroutine is running")
	defer fmt.Println(c.RemoteAddr().String(), "conn writer exit!")
	// 不断的阻塞等待channel的消息，进行写给客户端
	for {
		select {
		case data := <-c.msgChan:
			// 有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send data error", err)
				return
			}
		case <-c.ExitChan:
			// 代表Reader已经退出，此时Writer也要退出
			return
		}
	}
}

// 读消息Goroutine，用于从客户端中读取数据
func (c *Connection) StartReader() {
	fmt.Println("Reader Goroutine is running")
	defer fmt.Println(c.RemoteAddr().String(), "conn reader exit!")
	defer c.Stop()

	for {
		// 创建一个拆包解包的对象
		dp := NewDataPack()
		// 读取客户端的Msg head, 8个字节的二进制流
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("read msg head error ", err)
			break
		}
		// 拆包，放在一个msg中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("server unpack err ", err)
			break
		}
		// 按照dataLen，读取data数据，放在msg.Data中
		var data []byte
		if msg.GetMsgLen() > 0 {
			data = make([]byte, msg.GetMsgLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error ", err)
				break
			}
		}
		msg.SetData(data)

		// 更新最后活动时间
		c.UpdateActivity()

		// 得到当前连接的Request
		req := Request{
			conn: c,
			msg:  msg,
		}
		// 从路由中找到注册绑定的Conn对应的MsgHandler调用
		if utils.GlobalObject.WorkerPoolSize > 0 {
			// 已经启动工作池机制，将消息交给Worker处理
			c.MsgHandler.SendMsgToTaskQueue(&req)
		} else {
			// 从绑定好的消息和对应的处理方法中执行对应的Handle方法
			go c.MsgHandler.DoMsgHandler(&req)
		}
	}
}

// 心跳检测
func (c *Connection) startHeartbeat() {
	// 心跳检测间隔
	interval := time.Duration(utils.GlobalObject.HeartbeatInterval) * time.Second
	// 心跳超时时间
	timeout := time.Duration(utils.GlobalObject.HeartbeatTimeout) * time.Second

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 检查最后活动时间，如果超时则关闭连接
			if time.Since(c.lastActivityTime) > timeout {
				fmt.Printf("心跳超时，关闭连接 ConnID=%d, IP=%s, 最后活动: %s\n",
					c.ConnID, c.RemoteAddr().String(), c.lastActivityTime.Format("2006-01-02 15:04:05"))
				c.Stop()
				return
			}
		case <-c.ExitChan:
			// 连接已关闭，退出心跳检测
			return
		}
	}
}

// 更新最后活动时间
func (c *Connection) UpdateActivity() {
	c.propertyLock.Lock()
	c.lastActivityTime = time.Now()
	c.propertyLock.Unlock()
}

// 获取最后活动时间
func (c *Connection) GetLastActivityTime() time.Time {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()
	return c.lastActivityTime
}

// 提供一个SendMsg方法，将我们要发送给客户端的数据，先进行封包，再发送
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed {
		return nil
	}
	// 将data进行封包
	dp := NewDataPack()
	binaryMsg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return err
	}
	// 将数据发送给channel
	c.msgChan <- binaryMsg
	return nil
}

func (c *Connection) Start() {
	fmt.Println("Conn Start()... ConnID = ", c.ConnID)
	// 启动当前连接的读数据业务
	go c.StartReader()
	// 启动当前连接的写数据业务
	go c.StartWriter()
	// 启动心跳检测
	go c.startHeartbeat()
	// 按照开发者传递进来的创建连接时需要处理的业务，执行hook方法
	c.TCPServer.CallOnConnStart(c)
}

// 停止连接，结束当前连接状态
func (c *Connection) Stop() {
	fmt.Println("Conn Stop()... ConnID = ", c.ConnID)
	// 如果当前连接已经关闭
	if c.isClosed {
		return
	}
	c.isClosed = true
	// 调用开发者注册的该连接的销毁之前需要处理的业务
	c.TCPServer.CallOnConnStop(c)
	// UnRegister方法解除当前连接的注册
	c.UnRegister()
	// 关闭socket连接
	c.Conn.Close()
	// 告知Writer关闭
	c.ExitChan <- true
	// 回收资源
	close(c.msgChan)
	close(c.ExitChan)
}

func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Connection) Send(data []byte) error {
	return nil
}

// 设置连接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	// 添加一个连接属性
	c.property[key] = value
}

// 获取连接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()
	// 读取属性
	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, fmt.Errorf("no property found")
	}
}

// 移除连接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	// 删除属性
	delete(c.property, key)
}
