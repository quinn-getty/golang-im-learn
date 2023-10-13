package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	Conn       net.Conn
	flag       int
}

func NewClient(ip string, port int) *Client {
	client := &Client{
		ServerIp:   ip,
		ServerPort: port,
		flag:       999,
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		log.Print("net.Dial err: ", err)
	}

	client.Conn = conn
	return client
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器的IP(默认: 127.0.0.1)")
	flag.IntVar(&serverPort, "port", 2222, "设置服务器的Port(默认: 2222)")
}

func (client *Client) UpdateName() bool {
	log.Println(">>>>>>>>>please input user name<<<<<<<<")
	fmt.Scanln(&client.Name)
	sendMsg := "rename:" + client.Name + "\n"
	// log.Print("rename Msg:", sendMsg, "\n")
	if _, err := client.Conn.Write([]byte(sendMsg)); err != nil {
		log.Print("client.Conn.Write", err)
		return false
	}

	return true
}

func (client *Client) PublicChat() {
	var msg string
	log.Println(">>>>>>>>>please input chat content ,[exit]quit<<<<<<<<")
	fmt.Scanln(&msg)
	for msg != "exit" {
		if msg != "" {
			if _, err := client.Conn.Write([]byte(msg)); err != nil {
				log.Print("client.Conn.Write error", err)
				break
			}
			msg = ""
			log.Println(">>>>>>>>>please input chat content ,[exit]quit<<<<<<<<")
			fmt.Scanln(&msg)
		}
	}
}

func (this *Client) SelectUsers() {
	_, err := this.Conn.Write([]byte("who\n"))
	if err != nil {
		log.Print("client.Conn.Write", err)
	}
}

func (client *Client) PrivateChat() {
	client.SelectUsers()
	var name string
	var chatMsg string
	log.Println(">>>>>>>>> please input user name [exit] <<<<<<<<")
	fmt.Scanln(&name)
	for name != "exit" {
		log.Println(">>>>>>>>> please input message![exit] <<<<<<<<")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			client.Conn.Write([]byte("@" + name + ":" + chatMsg + "\n\n"))
		}
	}

}

// 处理返回的信息
func (client *Client) DealResponse() {
	io.Copy(os.Stdout, client.Conn)
}

func (client *Client) menu() bool {
	var flag int
	log.Print("\n1. 公聊\n", "2. 私聊\n", "3. 修改用户名\n", "0. 退出")

	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	}
	log.Println("请输入合法数字")
	return false
}

// func (Client *Client)

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}
		switch client.flag {
		case 1:
			// log.Println("公聊")
			client.PublicChat()
			break
		case 2:
			log.Println("私聊")
			break
		case 3:
			log.Println("修改用户名")
			client.UpdateName()
			break

		}
	}
}

func main() {
	// 解析命令
	flag.Parse()

	client := NewClient(serverIp, serverPort)

	if client == nil {
		log.Println(">>>>>>> 连接失败")
		return
	}

	go client.DealResponse()

	log.Println(">>>>>>> 连接成功")

	client.Run()
	// select {}
}
