package utils

import (
	"encoding/json"
	"os"

	"github.com/Xaytick/zinx/ziface"
)

/*
存储有关zinx的全局参数，供其他模块使用
一些参数可以通过zinx.json由用户进行配置
*/

type GlobalObj struct {
	// server
	TCPServer ziface.IServer // 当前Zinx全局的Server对象
	Host      string         // 当前服务器主机监听的IP
	TcpPort   int            // 当前服务器主机监听的端口号
	Name      string         // 当前服务器的名称
	// zinx
	Version        string // 当前Zinx的版本号
	MaxConn        int    // 当前服务器主机允许的最大连接数
	MaxPackageSize uint32 // 当前Zinx框架数据包的最大值
	WorkerPoolSize uint32 // 业务工作Worker池的大小
	MaxTaskLen     uint32 // 业务工作Worker对应负责的任务队列最大任务存储数量, 允许用户最多开辟多少个worker
}

// 定义一个全局的对外GlobalObj
var GlobalObject *GlobalObj

// 提供一个init方法，初始化当前的GlobalObject
func init() {
	// 如果配置文件没有加载，下面的是默认的值
	GlobalObject = &GlobalObj{
		Name:           "ZinxServerApp",
		Version:        "V0.10",
		TcpPort:        8999,
		Host:           "0.0.0.0",
		MaxConn:        1000,
		MaxPackageSize: 4096,
		WorkerPoolSize: 10,
		MaxTaskLen:     1024,
	}

	// 应该通过zinx.json来加载自定义的参数
	GlobalObject.Reload()
}

// 加载用户自定义的配置文件
func (g *GlobalObj) Reload() {
	data, err := os.ReadFile("conf/zinx.json")
	json.Unmarshal(data, &GlobalObject)
	if err != nil {
		panic(err)
	}

}
