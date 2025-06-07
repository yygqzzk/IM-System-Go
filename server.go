package main

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.Mutex

	// 消息广播 channel
	Exchange chan string
}

// 创建一个server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Exchange:  make(chan string),
	}

	return server
}

// 处理连接请求
func (server *Server) Handler(conn net.Conn) {
	// 连接业务
	fmt.Println("server new connection: " + conn.RemoteAddr().String())
	// 创建新用户, 并加入到onlineMap中
	user := NewUser(conn, server)
	user.Online()

	// 监听用户活跃的channel
	isActive := make(chan bool)

	// 接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)

			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}
			// 提取用户的消息
			msg := string(buf[:n-1])
			// 广播消息
			user.doMessage(msg)

			// 刷新活跃时间
			isActive <- true
		}
	}()

	// 阻塞handler
	for {
		select {
		// 收到true消息时,自动刷新定时器时间
		case <-isActive:
		// 配置定时器
		case <-time.After(time.Minute * 10):
			// 超时关闭用户连接
			fmt.Printf("%s timeout, close connection ... \n", conn.RemoteAddr().String())

			// 消息提示
			user.SendMsg("timeout close connection")

			// 关闭消息channel
			close(user.Inbox)

			// 关闭连接
			err := conn.Close()
			if err != nil {
				fmt.Println("Close conn err:", err)
				return
			}
			return
		}
	}

}

func (server *Server) BroadCast(user *User, msg string) {
	msg = "[" + user.Addr + "]" + user.Name + ": " + msg
	server.Exchange <- msg
}

// 监听Exchange, 并进行消息转发
func (server *Server) Forward() {
	for msg := range server.Exchange {
		server.mapLock.Lock()
		for _, v := range server.OnlineMap {
			//fmt.Println("服务端转发消息: " + msg)
			v.Inbox <- msg
		}
		server.mapLock.Unlock()
	}
}

// 启动服务器接口
func (server *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	fmt.Println("server: " + server.Ip + ":" + strconv.Itoa(server.Port) + " start success")

	// 启动监听线程
	go server.Forward()

	// accept
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err:", err)
			continue
		}
		// do handler
		go server.Handler(conn)
	}

	// close listen socket
	defer listener.Close()
}
