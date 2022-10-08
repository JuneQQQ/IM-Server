package main

import (
	"fmt"
	"net"
	"runtime"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server

	isAlive chan bool
}

func (u *User) listenMessage() {
	for {
		// 父go程存活的情况下
		msg := <-u.C
		_, err := u.conn.Write([]byte(msg + "\n"))
		if err != nil {
			println("【系统信息】"+"连接异常，即将退出！异常原因：", err.Error())
			println("可能的原因：用户已被踢出")
			return
		}
	}
}

func NewUser(conn net.Conn, server *Server) *User {
	user := &User{
		Name:    conn.RemoteAddr().String(),
		Addr:    conn.RemoteAddr().String(),
		C:       make(chan string),
		conn:    conn,
		server:  server,
		isAlive: make(chan bool),
	}

	return user
}

// Online 用户上线
func (u *User) Online() {
	u.server.mapLock.Lock()
	u.server.OnlineUserMap[u.Name] = u
	u.server.mapLock.Unlock()
	// 上线消息广播
	u.server.BroadMessage(u, "已上线")
}

// OffLine 用户下线
func (u *User) OffLine() {
	u.server.mapLock.Lock()
	delete(u.server.OnlineUserMap, u.Name)
	u.server.mapLock.Unlock()
	// 清理资源
	close(u.C)
	// 断开conn
	err := u.conn.Close()
	if err != nil {
		println("断开用户连接时异常，异常信息：", err.Error())
		return
	}
	// 给其他用户广播下线消息
	u.server.BroadMessage(u, "已下线")
	runtime.Goexit()
}

// 给当前用户发消息
func (u *User) sendMessage(msg string) {
	u.C <- msg
}

// DoMessage 处理消息
func (u *User) DoMessage(msg string) {
	orders := strings.Split(msg, " ")
	u.isAlive <- true // 存活标志，channel只要有值就可以

	// 错误处理
	defer func() {
		if r := recover(); r != nil {
			u.sendMessage("【系统信息】" + "命令有误，请重新输入")              // 用户提示命令错误
			fmt.Printf("【ERROR】"+"输入的命令：%s,本次异常：%v", orders, r) // 给服务器发送异常信息
		}
	}()

	switch orders[0] {
	case "users":
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineUserMap {
			onLineMessage := "【系统信息】" + "[" + user.Name + "]" + ":" + "在线.."
			u.sendMessage(onLineMessage)
		}
		u.server.mapLock.Unlock()
	case "rename":
		if len(orders) == 2 && len(orders[1]) >= 3 && len(orders[1]) <= 10 {
			_, ok := u.server.OnlineUserMap[orders[1]]
			if ok {
				u.sendMessage("【系统信息】名称已被使用，请重新输入")
			} else {
				u.changeUserName(orders[1])
				u.sendMessage("【系统信息】更名成功~您的新名称是:" + orders[1])
			}
		} else {
			u.sendMessage("【系统信息】参数不合法，请重新输入，用户名长度在3-10之间，不能含有空格及换行等字符")
		}
	case "to":
		if len(orders) == 3 {
			target, ok := u.server.OnlineUserMap[orders[1]]
			fmt.Printf("%v\n", u.server.OnlineUserMap)
			fmt.Printf("%v\n", target)
			if ok {
				// 存在，可以发送给该用户
				target.sendMessage("【新消息】[" + u.Name + "]:" + orders[1])
				u.sendMessage("【系统信息】" + "消息已发送~")
			} else {
				u.sendMessage("【系统信息】用户 " + orders[1] + " 不存在")
			}
		} else {
			u.sendMessage("【系统信息】参数不合法，请重新输入")
		}
	default:
		u.server.BroadMessage(u, msg)
	}
}

// 更改用户名
func (u *User) changeUserName(newName string) {
	u.server.mapLock.Lock()
	delete(u.server.OnlineUserMap, u.Name)
	u.Name = newName
	u.server.OnlineUserMap[newName] = u
	u.server.mapLock.Unlock()
}
