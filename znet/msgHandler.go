package znet

import (
	"fmt"
	"sync/atomic"

	"github.com/Xaytick/zinx/utils"
	"github.com/Xaytick/zinx/ziface"
)

/*
	消息处理模块的实现
*/

type MsgHandler struct {
	// 存放每个MsgID 所对应的处理方法
	Apis map[uint32]ziface.IRouter
	// 负责worker取任务的消息队列
	TaskQueue []chan ziface.IRequest
	// 负责worker池的worker数量
	WorkerPoolSize uint32
}

// 初始化,创建MsgHandler方法
func NewMsgHandler() *MsgHandler {
	return &MsgHandler{
		Apis:           make(map[uint32]ziface.IRouter),
		TaskQueue:      make([]chan ziface.IRequest, utils.GlobalObject.WorkerPoolSize),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize, // 从全局配置中获取
	}
}

// 调度,执行对应的Router消息处理方法
func (mh *MsgHandler) DoMsgHandler(Request ziface.IRequest) {
	// 1.从Request中找到msgID
	handler, ok := mh.Apis[Request.GetMsgID()]
	if !ok {
		fmt.Println("api msgID = ", Request.GetMsgID(), " is not found and need registry!")
		return
	}
	// 2.根据msgID调度对应的router业务
	handler.PreHandle(Request)
	handler.Handle(Request)
	handler.PostHandle(Request)
}

// 为消息添加具体的处理逻辑
func (mh *MsgHandler) AddRouter(msgID uint32, router ziface.IRouter) {
	// 1.判断当前msg绑定的API处理方法是否已经存在
	if _, ok := mh.Apis[msgID]; ok {
		// id已经注册
		panic(fmt.Sprintf("repeat api, msgID = %d", msgID))
	}

	// 2.添加msg与api的绑定关系
	mh.Apis[msgID] = router
	fmt.Println("Add api msgID = ", msgID)
}

// 启动一个Worker工作池(开启工作池的动作只能发生一次，一个zinx框架只能有一个worker工作池)
func (mh *MsgHandler) StartWorkerPool() {
	// 根据workerPoolSize 分别开启Worker，每个Worker用一个go来承载
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		// 一个worker被启动
		// 1.当前的worker对应的channel消息队列，开辟空间，第i个worker就用第i个channel
		mh.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalObject.MaxTaskLen)
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}
}

// 启动一个Worker工作流程
func (mh *MsgHandler) StartOneWorker(workerID int, taskQueue chan ziface.IRequest) {
	fmt.Println("Worker ID = ", workerID, " is started...")
	// 不断的阻塞等待对应消息队列的消息
	for request := range taskQueue {
		// 如果有消息过来，出列的就是一个客户端的Request，执行当前Request所绑定的业务
		mh.DoMsgHandler(request)
	}
}

// 将消息交给TaskQueue，由Worker进行处理
var rrIndex uint32 = 0

func (mh *MsgHandler) SendMsgToTaskQueue(request ziface.IRequest) {
	// 轮询分配worker, 使用原子操作保证线程安全
	idx := atomic.AddUint32(&rrIndex, 1)
	workerID := idx % mh.WorkerPoolSize
	fmt.Println("Add ConnID = ", request.GetConnection().GetConnID(),
		"request MsgID = ", request.GetMsgID(),
		"to workerID = ", workerID)
	// 将消息发送给对应的worker的TaskQueue
	mh.TaskQueue[workerID] <- request
}
