package examples

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/zzliekkas/flow/websocket"
)

// WebSocketExample 展示WebSocket的基本用法
func WebSocketExample() {
	// 创建WebSocket管理器
	manager := websocket.NewManager()

	// 设置消息处理器
	manager.SetMessageHandler(func(conn *websocket.Connection, msg *websocket.Message) error {
		if msg.Type == "chat" && msg.Event == "chat.message" {
			// 收到聊天消息时，广播给所有用户
			log.Printf("收到聊天消息: %v\n", msg.Data)
			manager.Broadcast(msg)
		}
		return nil
	})

	// 设置连接后钩子函数，处理用户加入频道
	manager.SetAfterConnect(func(conn *websocket.Connection) {
		// 用户连接后自动加入 "chat" 频道
		conn.JoinChannel("chat")

		// 通知频道内所有用户有新用户加入
		msg := &websocket.Message{
			Type:    "channel",
			Channel: "chat",
			Event:   "user.joined",
			Data: map[string]interface{}{
				"user_id":   conn.UserID,
				"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
			},
		}
		manager.BroadcastToChannel("chat", msg)
	})

	// 设置断开连接前钩子函数，处理用户离开频道
	manager.SetBeforeDisconnect(func(conn *websocket.Connection) {
		// 获取用户的频道列表
		channels := conn.GetChannels()

		// 为每个频道发送通知
		for _, channel := range channels {
			msg := &websocket.Message{
				Type:    "channel",
				Channel: channel,
				Event:   "user.left",
				Data: map[string]interface{}{
					"user_id":   conn.UserID,
					"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
				},
			}
			manager.BroadcastToChannel(channel, msg)
		}
	})

	// 设置认证函数，从请求中提取用户ID
	manager.SetAuthFunc(func(r *http.Request) (string, map[string]interface{}, error) {
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			return "", nil, fmt.Errorf("缺少 user_id 参数")
		}
		return userID, nil, nil
	})

	// 创建HTTP处理器
	mux := http.NewServeMux()

	// 注册WebSocket升级处理器
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// 通过认证函数自动处理用户ID
		manager.HandleRequest(w, r)
	})

	// 启动HTTP服务器
	go func() {
		log.Println("启动WebSocket服务器在 :8080")
		if err := http.ListenAndServe(":8080", mux); err != nil {
			log.Fatalf("HTTP服务器错误: %v", err)
		}
	}()

	// 使用管理器启动定时任务，每5秒广播一次服务器时间
	go func() {
		for {
			time.Sleep(5 * time.Second)

			// 创建服务器时间消息
			msg := &websocket.Message{
				Type:  "event",
				Event: "server.time",
				Data: map[string]interface{}{
					"time": time.Now().Format(time.RFC3339),
				},
				Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
			}

			// 广播给所有连接的客户端
			manager.Broadcast(msg)
		}
	}()

	// 使用我们的客户端API连接到WebSocket服务器
	go func() {
		// 等待服务器启动
		time.Sleep(1 * time.Second)

		// 创建WebSocket客户端
		client := websocket.NewClient("ws://localhost:8080/ws?user_id=client1")

		// 设置连接成功回调
		client.OnConnect = func() {
			log.Println("客户端已连接")

			// 加入聊天频道
			if err := client.JoinChannel("chat"); err != nil {
				log.Printf("加入频道错误: %v", err)
			}

			// 发送欢迎消息
			client.SendToChannel("chat", "chat.message", map[string]string{
				"text": "大家好！我是客户端1",
			})
		}

		// 设置接收消息回调
		client.OnMessage = func(msg *websocket.Message) {
			log.Printf("收到消息: 类型=%s, 事件=%s, 数据=%v", msg.Type, msg.Event, msg.Data)
		}

		// 连接到服务器
		if err := client.Connect(); err != nil {
			log.Printf("连接错误: %v", err)
		}
	}()

	// 阻塞主线程
	select {}
}
