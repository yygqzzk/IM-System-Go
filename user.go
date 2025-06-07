package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	Inbox  chan string
	conn   net.Conn
	server *Server
}

func (user *User) Online() {
	// 通过锁来防止并发问题
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()

	// 用户上线提醒
	user.doMessage("online ~ ")
}

func (user *User) Offline() {

	// 删除在线用户
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()

	//广播下线消息
	user.doMessage("offline ~ ")
}

// 给当前User对应的客户端发送消息
func (user *User) SendMsg(msg string) {
	user.conn.Write([]byte(msg))
}

func (user *User) doMessage(msg string) {
	// 查询在线用户
	if msg == "who" {
		user.server.mapLock.Lock()
		for _, v := range user.server.OnlineMap {
			onlineMsg := "[" + v.Addr + "]" + v.Name + ": Online ... \n"
			user.SendMsg(onlineMsg)
		}
		user.server.mapLock.Unlock()
	} else if len(msg) > 7 && strings.HasPrefix(msg, "rename|") {
		// 修改用户名
		newName := strings.Split(msg, "|")[1]

		// 判断newName是否重复
		_, exist := user.server.OnlineMap[newName]
		if exist {
			info := fmt.Sprintf("%s has been taken \n", newName)
			user.SendMsg(info)
		} else {
			// 修改时需要加锁
			user.server.mapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.Name = newName
			user.server.OnlineMap[newName] = user
			user.server.mapLock.Unlock()

			info := fmt.Sprintf("name has been updated: %s \n", newName)
			user.SendMsg(info)
		}
	} else {
		user.server.BroadCast(user, msg)
	}

}

func NewUser(conn net.Conn, server *Server) *User {
	user := &User{
		Name:   conn.RemoteAddr().String(),
		Addr:   conn.RemoteAddr().String(),
		Inbox:  make(chan string),
		conn:   conn,
		server: server,
	}
	// 用户监听消息
	go user.ListenMessage()
	return user
}

// 用户读取消息，并返回客户端
func (user *User) ListenMessage() {
	for info := range user.Inbox {
		//fmt.Println("用户收取消息: " + info)
		_, err := user.conn.Write([]byte(info + "\r\n"))
		if err != nil {
			fmt.Println("写入数据失败: " + err.Error())
			continue
		}
	}
}
