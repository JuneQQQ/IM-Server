package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	IP   string
	Port int
	// 在线用户列表
	OnlineUserMap map[string]*User
	mapLock       sync.RWMutex
	// 消息广播channel，接收所有用户的消息
	Message chan string
}

// NewServer 创建服务器
func NewServer(ip string, port int) *Server {
	server := &Server{
		IP:   ip,
		Port: port,

		OnlineUserMap: make(map[string]*User),
		Message:       make(chan string),
	}

	return server
}

// Start 服务器启动 对象方法
func (s *Server) Start() {
	// socket listener   创建socket并绑定IP+PORT
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		println("给定地址无法连接！错误原因：", err.Error())
		return
	}
	// close listener socket
	defer func(listen net.Listener) {
		err := listen.Close()
		if err != nil {
			println("已正常关闭~")
		}
	}(listener)

	println("成功绑定服务器地址，接下来开始监听客户端请求~")
	// 启动监听BroadMessage的goroutine
	go s.ListenBroadMessage()
	// accept  服务器开始监听连接请求
	for {
		conn, err := listener.Accept()
		if err != nil {
			println("accept 异常！错误原因：", err)
			continue
		}
		println("连接已创建，客户端信息：", conn.RemoteAddr().String())
		go s.Handler(conn) // 创建goroutine处理conn
	}
}

// Handler 处理连接，此方法由用户专属go程调用
func (s *Server) Handler(conn net.Conn) {
	// 将用户加入OnlineMap中
	user := NewUser(conn, s)
	// 0.启动用户端的消息监听
	go user.listenMessage()
	// 1.用户上线
	user.Online()
	// 2.主动发送广播消息
	go func() {
		// 消息读取缓冲区 最大4KB
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				//user.OffLine() // 用户下线
				return
			}
			if err != nil && err != io.EOF {
				println("Conn Read err:", err)
				return
			}
			// 去除最后的\n  一共n个字节，那么取[0,n-1)
			msg := string(buf[:n-1])
			user.DoMessage(msg)
		}
	}()
	// 3.检测用户存活与否
	for {
		select {
		// 只要用户十秒内发送任何命令或消息即证明存活
		case <-user.isAlive:
			// doNothing 走到这里下面的定时器就会被重置
		case <-time.After(time.Minute * 5):
			user.sendMessage("【提示】" + "[" + user.Name + "]" + "您不太活跃，即将断开连接！")
			time.Sleep(time.Second)
			user.OffLine()
		}
	}
}

// BroadMessage 广播用户消息
func (s *Server) BroadMessage(user *User, msg string) {
	//println("将要广播消息：", msg)
	msg = "【广播】[" + user.Name + "]" + ":" + msg
	s.Message <- msg
}

// ListenBroadMessage  监听Message这个channel
func (s *Server) ListenBroadMessage() {
	for {
		msg := <-s.Message
		// 将消息发送给所有用户
		s.mapLock.Lock()
		for _, user := range s.OnlineUserMap {
			user.C <- msg
		}
		s.mapLock.Unlock()
	}
}

func main() {
	server := NewServer("127.0.0.1", 8788)
	server.Start()
}
