package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	Message   chan string
}

// 监听server实例 有message 就广播给给用
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message
		log.Print("server 收到消息 ", msg)
		this.mapLock.Lock()
		for _, user := range this.OnlineMap {
			user.C <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := fmt.Sprintf("[ %s ] %s : %s", user.Addr, user.Name, msg)
	this.Message <- sendMsg
}

// 处理连接
func (this *Server) Handler(conn net.Conn) {
	fmt.Println("连接成功")
	user := NewUser(conn, this)

	user.OnLine()

	isLive := make(chan bool)

	// 处理用户发送的数据
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.OffLine()
				return
			}

			if err != nil && err != io.EOF {
				log.Panic("conn.Read err: ", err)
				return
			}

			// 提取用户消息 去除"\n"
			msg := string(buf[:n-1])
			user.DoMessage(msg)

			isLive <- true
		}
	}()

	for {
		select {
		case <-isLive:
			// 重置时间
		case <-time.After(time.Second * 60 * 5):

			this.mapLock.Lock()
			delete(this.OnlineMap, user.Name)
			this.mapLock.Unlock()

			user.SendMsg("连接超时，您已经被踢下线！")
			conn.Close()
			user.conn.Close()

			close(user.C) // 存在bug ListenMessage会一直执行
			// 退出
			runtime.Goexit()
			// return
		}

	}
}

func (this *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err: ", err)
		return
	}

	defer listener.Close()
	go this.ListenMessager()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("net.Listen err: ", err)
			continue
		}

		go this.Handler(conn)
	}
}

func NewServer(ip string, port int) *Server {
	return &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}
