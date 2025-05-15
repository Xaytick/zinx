package znet

import (
	"fmt"

	"github.com/Xaytick/zinx/utils"
	"github.com/Xaytick/zinx/ziface"
)

// HeartbeatRouter 心跳消息处理路由
type HeartbeatRouter struct {
	BaseRouter
}

// Handle 处理心跳消息
func (hr *HeartbeatRouter) Handle(request ziface.IRequest) {
	// 获取连接
	conn := request.GetConnection()
	// 更新活动时间(虽然在StartReader中已经更新过，这里是为了代码的完整性)
	if c, ok := conn.(*Connection); ok {
		c.UpdateActivity()
	}

	// 回复一个PONG消息
	err := conn.SendMsg(utils.PONG_MSG_ID, []byte("pong"))
	if err != nil {
		fmt.Println("回复心跳消息失败:", err)
	} else {
		fmt.Printf("心跳响应: PONG 发送至 ConnID=%d\n", conn.GetConnID())
	}
}
