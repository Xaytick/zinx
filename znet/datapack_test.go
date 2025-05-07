package znet

import (
	"fmt"
	"io"
	"net"
	"testing"
)

/*
	封包拆包实例的测试
*/

func TestDataPack(t *testing.T) {
	/* 模拟的服务器
	 */

	// 创建socketTCP
	listener, err := net.Listen("tcp", "127.0.0.1:8999")
	if err != nil {
		fmt.Println("server listen err:", err)
		return
	}

	// 创建一个go承载，负责从客户端处理业务
	go func() {
		// 从客户端读取数据，拆包处理
		for {
			conn, err := listener.Accept()
			if err!= nil {
				fmt.Println("server accept err:", err)		
			}		

			go func(conn net.Conn) {
				// 处理客户端请求
				// ----> 拆包的过程 <----
				// 定义一个拆包的对象dp
				dp := NewDataPack()
				for {
					// 1. 先读出流中的head部分
					headData := make([]byte, dp.GetHeadLen())
					_, err := io.ReadFull(conn, headData)
					if err != nil {
						fmt.Println("read head error")
						return
					}
					msgHead, err := dp.Unpack(headData)
					if err != nil {
						fmt.Println("server unpack err:", err)
						return
					}
					// msg是有data数据的，需要再次读取data数据
					// 2. 再根据head中的dataLen，读取data内容
					if msgHead.GetMsgLen() > 0 {
						msg := msgHead.(*Message)
						msg.Data = make([]byte, msg.GetMsgLen())
						
						_, err = io.ReadFull(conn, msg.Data)
						if err != nil {
							fmt.Println("server unpack data err:", err)
							return
						}
						// 完整的一个消息已经读取完毕
						fmt.Println("----> Recv Msg: ID = ", msg.Id,
						", len = ", msg.DataLen, 
						", data = ", string(msg.Data))		
					}
				}
				
			}(conn)
		}
	}()
	

	/* 模拟的客户端
	 */
	conn, err := net.Dial("tcp", "127.0.0.1:8999")
	if err != nil {
		fmt.Println("client dial err:", err)
		return
	}

	// 创建一个封包对象dp
	dp := NewDataPack()

	// 模拟黏包过程，封装两个msg一同发送
	// 封装第一个msg1包
	msg1 := &Message{
		Id: 1,
		DataLen: 5,
		Data: []byte{'h', 'e', 'l', 'l', 'o'},	
	}
	sendData1, err := dp.Pack(msg1)
	if err != nil {
		fmt.Println("client pack msg1 err:", err)
		return
	}
	// 封装第二个msg2包
	msg2 := &Message{
		Id: 2,
		DataLen: 7,
		Data: []byte{'w', 'o', 'r', 'l', 'd', '!', '!'},	
	}
	sendData2, err := dp.Pack(msg2)
	if err != nil {
		fmt.Println("client pack msg2 err:", err)
		return	
	}
	// 将两个包一次发给服务器
	sendData1 = append(sendData1, sendData2...)
	conn.Write(sendData1)

}