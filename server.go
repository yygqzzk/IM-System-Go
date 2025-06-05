package main

import (
	"fmt"
	"net"
	"strconv"
	"sync"
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
	// 启动监听线程
	go server.Forward()

	return server
}

// 处理连接请求
func (server *Server) Handler(conn net.Conn) {
	// 连接业务
	fmt.Println("server new connection: " + conn.RemoteAddr().String())
	// 创建新用户, 并加入到onlineMap中
	user := NewUser(conn)

	// 通过锁来防止并发问题
	server.mapLock.Lock()
	server.OnlineMap[user.Name] = user
	server.mapLock.Unlock()

	// 用户上线提醒
	server.BroadCast(user, "online ~ ")

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
			fmt.Println("服务端转发消息: " + msg)
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
