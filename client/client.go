package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type Client struct {
	ServerIP   string
	ServerPort int
	Name       string
	conn       net.Conn
	module     int
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIP:   serverIp,
		ServerPort: serverPort,
		Name:       "游客",
		conn:       nil,
		module:     9999,
	}
	// 链接 server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		println("net dail error:", err)
		return nil
	}
	client.conn = conn

	return client
}

var ServerIP string
var ServerPort int

func init() {
	flag.StringVar(&ServerIP, "ip", "127.0.0.1", "指定服务器IP")
	flag.IntVar(&ServerPort, "port", 8788, "指定服务器PORT")
}

func (c *Client) Run() {
	for c.module != 0 {
		for !c.menu() {
		}
		// go 的 switch 不需要 break
		switch c.module {
		case 1:
			// 广播模式
			c.BroadChat()
		case 2:
			// 私聊模式
			c.PrivateChat()
		case 3:
			// 改名
			c.Rename()
		}
		time.Sleep(time.Second)
		println("\nDONE，请选择下一轮操作")
	}
}

func (c *Client) BroadChat() {
	var msg string
	fmt.Print(">>>>>>>请输入消息，exit退出：")
	fmt.Scanln(&msg)
	for msg != "exit" {
		if len(msg) != 0 {
			msg += "\n"
			_, err := c.conn.Write([]byte(msg))
			if err != nil {
				println("ERROR:", err)
				return
			}
		}
		time.Sleep(time.Second)
		msg = ""
		fmt.Print(">>>>>请输入消息，exit退出：")
		fmt.Scanln(&msg)
	}
}

func (c *Client) menu() bool {
	fmt.Println("1.广播模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	var module int
	_, _ = fmt.Scanln(&module)
	if module >= 0 && module <= 4 {
		c.module = module
		return true
	} else {
		println("输入不合法，请重新输入")
		return false
	}
}

func (c *Client) Rename() bool {
	fmt.Print(">>>>>>>>>请输入用户名：")
	fmt.Scanln(&c.Name)
	msg := "rename " + c.Name + "\n"
	n, err := c.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("输入有误：", err.Error())
		return false
	}
	println(n, err)
	return true
}

func (c *Client) GetResponse() {
	//将client.conn的输出copy到stdout标准输出上 该方法阻塞
	io.Copy(os.Stdout, c.conn)
}

func (c *Client) SelectUsers() {
	msg := "users"
	_, err := c.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
}
func (c *Client) PrivateChat() {
	var remoteName string
	var msg string
	c.SelectUsers()

	for remoteName != "exit" {
		fmt.Print(">>>>>>>>>>>>请输入聊天对象：")
		fmt.Scanln(&remoteName)

		for msg != "exit" {
			if len(msg) != 0 {
				msg := "to " + remoteName + " " + msg + "\n"
				_, err := c.conn.Write([]byte(msg))
				if err != nil {
					println("ERROR:", err)
					return
				}
			}
			msg = ""
			fmt.Print(">>>>>>>>>>请输入消息内容,exit退出：")
			fmt.Scanln(&msg)
		}
		c.SelectUsers()
		fmt.Print(">>>>>>>>>>>>请输入聊天对象：")
		fmt.Scanln(&remoteName)
	}

}

func main() {
	client := NewClient(ServerIP, ServerPort)
	if client == nil {
		println("连接服务器失败！")
	} else {
		println("连接服务器成功~")
	}
	go client.GetResponse()

	// 启动客户端服务
	client.Run()
}
