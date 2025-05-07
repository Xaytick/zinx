package znet

import (
	"fmt"
	"net"
	"zinx/utils"
	"zinx/ziface"
)

type Server struct {
	// 服务器名称
	Name string
	// 服务器绑定的ip版本
	IPVersion string
	// 服务器监听的IP
	IP string
	// 服务器监听的端口
	Port int
	// 当前Server的消息管理模块，用来绑定MsgID和对应的处理业务API关系
	MsgHandler ziface.IMsgHandler
	// 当前Server的连接管理器
	ConnManager ziface.IConnManager
	// 当前Server的连接创建时Hook函数
	OnConnStart func(conn ziface.IConnection)
	// 当前Server的连接断开时的Hook函数
	OnConnStop func(conn ziface.IConnection)
}

func NewServer(name string) *Server {
	s := &Server{
		Name:      utils.GlobalObject.Name,
		IPVersion: "tcp4",
		IP:        utils.GlobalObject.Host,
		Port:      utils.GlobalObject.TcpPort,
		MsgHandler: NewMsgHandler(),
		ConnManager: NewConnManager(),
	}
	return s
}

func (s *Server) Start() {
	fmt.Println("[zinx] Server name:", utils.GlobalObject.Name,
	". Server Listener at IP: ", utils.GlobalObject.Host,
	", Port ", utils.GlobalObject.TcpPort)

	fmt.Println("[zinx] Version:", utils.GlobalObject.Version,
	". MaxConn:", utils.GlobalObject.MaxConn,
	". MaxPackageSize:", utils.GlobalObject.MaxPackageSize)
	
	go func() {
		// 0.启动worker工作池
		s.MsgHandler.StartWorkerPool()
		// 1.获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr err:", err)
			return
		}
		
		// 2.监听服务器的地址
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen ", s.IPVersion, " err", err)
			return
		}
		fmt.Println("start zinx server success, ", s.Name, "listening...")

		var cid uint32 = 0

		// 3.阻塞的等待客户端链接，处理客户端链接业务（读写）
		for {
			// 如果有客户端链接过来，阻塞会返回
			conn, err := listener.AcceptTCP()
			if err!= nil {
				fmt.Println("Accept err:", err)
				continue	
			}

			// 设置服务器最大连接控制，如果超过最大连接，那么则关闭此新的连接
			if s.ConnManager.Size() > utils.GlobalObject.MaxConn {
				fmt.Println("Too many Connections, MaxConn = ", utils.GlobalObject.MaxConn)
				conn.Close()
				continue
			}
			dealConn := NewConnection(s, conn, cid, s.MsgHandler)
			cid++

			go dealConn.Start()
		}	
	}()
}

func (s *Server) AddRouter(msgID uint32, router ziface.IRouter) {
	s.MsgHandler.AddRouter(msgID, router)
	fmt.Println("Added Router successfully!")
}

func (s *Server) Stop() {
	// 将一些服务器的资源、状态或者一些已经开辟的链接信息进行停止或者回收
	s.ConnManager.ClearConns()
	fmt.Println("[STOP] Zinx server name ", s.Name)
}

func (s *Server) Serve() {
	// 启动server的服务功能
	s.Start()

	// TODO 做一些启动服务器之后的额外业务

	// 阻塞状态
	select {}
}

func (s *Server) GetConnManager() ziface.IConnManager {
	return s.ConnManager
}

// 注册OnConnStart钩子函数的方法
func (s *Server) SetOnConnStart(hookFunc func(conn ziface.IConnection)) {
	s.OnConnStart = hookFunc
}

// 注册OnConnStop钩子函数的方法
func (s *Server) SetOnConnStop(hookFunc func(conn ziface.IConnection)) {
	s.OnConnStop = hookFunc
}

// 调用OnConnStart钩子函数的方法
func (s *Server) CallOnConnStart(conn ziface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("---> Call OnConnStart()...")
		s.OnConnStart(conn)
	}
}

// 调用OnConnStop钩子函数的方法
func (s *Server) CallOnConnStop(conn ziface.IConnection) {
	if s.OnConnStop!= nil {
		fmt.Println("---> Call OnConnStop()...")
		s.OnConnStop(conn)
	}
}