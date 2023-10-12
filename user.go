package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	go user.ListenMessage()
	return user
}

// 监听当前用户的channel的方法
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		log.Print("用户", this.Name, "准备发送消息", msg, len(msg), "\n", this.conn, "\n")
		this.conn.Write([]byte(msg + "\n"))
	}
}

// 用户上线
func (this *User) OnLine() {
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "已上线！")
}

// 发送消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 用户下线
func (this *User) OffLine() {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "已下线！")
}

// 用户处理消息
func (this *User) DoMessage(msg string) {
	log.Println(msg, msg == "who")
	if msg == "who" {
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := fmt.Sprintf("[ %s ] %s : 在线\n", user.Addr, user.Name)
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename:" {
		name := strings.Split(msg, ":")[1]

		_, ok := this.server.OnlineMap[name]

		if ok {
			this.SendMsg("名字[ " + name + "] 已存在！\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.Name = name
			this.server.OnlineMap[name] = this
			this.server.mapLock.Unlock()
			this.SendMsg("名字已经修改为: " + name + "\n")
		}

	} else {
		this.server.BroadCast(this, msg)
	}

}
