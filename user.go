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
	// for range 会识别是否关闭
	for msg := range this.C {
		log.Print("用户", this.Name, "准备发送消息", msg, len(msg), "\n", this.conn, "\n")
		this.conn.Write([]byte(msg + "\n"))
	}
	log.Println("----")
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
	//  查询在线
	if msg == "who" {
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := fmt.Sprintf("[ %s ] %s : 在线\n", user.Addr, user.Name)
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
		return
	}

	// 修改名字
	if len(msg) > 7 && msg[:7] == "rename:" {
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

		return
	}

	if len(msg) > 3 && msg[:2] == "@:" {
		name := strings.Split(msg, ":")[1]

		remoteUser, ok := this.server.OnlineMap[name]
		if !ok {
			this.SendMsg("用户[" + name + "]不存在\n")
			return
		}

		msg := strings.Split(msg, ":")[2]
		if msg == "" {
			this.SendMsg("私聊消息为空\n")
			return
		}

		remoteUser.SendMsg(fmt.Sprintf("【%s】对你说：%s", this.Name, msg))
		return
	}

	// 私聊 @name:message

	this.server.BroadCast(this, msg)

}
