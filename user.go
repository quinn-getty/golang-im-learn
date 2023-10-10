package main

import (
	"log"
	"net"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
}

func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
	}
	go user.ListenMessage()
	return user
}

// 监听当前用户的channel的方法
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		log.Print("用户", this.Name, "准备发送消息", msg)
		this.conn.Write([]byte(msg + "\n"))
	}
}
