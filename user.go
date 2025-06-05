package main

import (
	"fmt"
	"net"
)

type User struct {
	Name  string
	Addr  string
	Inbox chan string
	conn  net.Conn
}

func NewUser(conn net.Conn) *User {
	user := &User{
		Name:  conn.RemoteAddr().String(),
		Addr:  conn.RemoteAddr().String(),
		Inbox: make(chan string),
		conn:  conn,
	}
	// 用户监听消息
	go user.ListenMessage()
	return user
}

// 用户读取消息，并返回
func (user *User) ListenMessage() {
	for info := range user.Inbox {
		fmt.Println("用户收取消息: " + info)
		_, err := user.conn.Write([]byte(info + "\r\n"))
		if err != nil {
			fmt.Println("写入数据失败: " + err.Error())
			continue
		}
	}
}
