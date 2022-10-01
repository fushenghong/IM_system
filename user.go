package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server //在User结构体中嵌入了一个Server结构体，相当于面向对象中的继承。通过server字段便可访问Server结构体
	//的所有字段和方法。
}

//创建一个用户的API，NewUser实际上就是一个构造函数，对User对象进行初始化操作。
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String() //获取用户的网络终端地址，该网络终端地址以字符串的格式显示。
	user := &User{
		Name:   userAddr, //将Name字段，赋值为userAddr
		Addr:   userAddr, //将Addr字段，赋值为userAddr
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	//启动监听当前user channel消息的goroutine。
	go user.ListenMessage() //每一个用户都有一个通道，在初始化用户的时候，就可以开辟一个通道用来监听该用户的通道信息。
	return user
}

//用户的上线业务
func (this *User) Online() {
	//用户上线，将用户加入到onlineMap中
	this.server.maplock.Lock()
	this.server.OnlineMap[this.Name] = this //这个地方的this指的是指向User结构体的指针变量
	this.server.maplock.Unlock()
	//广播当前用户上线消息
	this.server.BroadCast(this, "online success")
}

//用户的下线业务
func (this *User) Offline() {
	//用户下线，将用户从onlineMap中删除
	this.server.maplock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.maplock.Unlock()
	//广播当前用户下线消息
	this.server.BroadCast(this, "offline success")
}

//给当前user对应的客户端发送消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg)) //将在线用户信息写入连接。
}

//用户处理消息的业务
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//查询当前在线用户都有哪些
		this.server.maplock.Lock()
		for _, user := range this.server.OnlineMap { //因为在onelineMap里面的用户都是在线用户，所以将该map里面的用户信息，反馈到查询的客户端即可。
			onlieMsg := "[" + user.Addr + "]" + user.Name + ":" + "online...\n"
			this.SendMsg(onlieMsg)
		}
		this.server.maplock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" { //简单理解就是客户端输入rename|洪富生 时，执行该语句。
		//此if分支是用来修改用户名的。
		//消息格式：rename|张三
		newName := strings.Split(msg, "|")[1] //strings.Split()函数，返回的是一个字符串切片。
		//将rename|洪富生，在|处进行切割。切割成两个子串。rename和洪富生。在将之存入[]string中。strings.Split(msg, "|")[1]
		//意思是将洪富生这个子串赋值给变量newName。
		//判断name是否存在
		_, ok := this.server.OnlineMap[newName] //通过用户名查找map中是否存在该用户。
		if ok {
			this.SendMsg("The current user name is occupied")
		} else {
			this.server.maplock.Lock()
			delete(this.server.OnlineMap, this.Name) //删除原用户对应的user对象。
			this.server.OnlineMap[newName] = this    //通过新用户名赋予一个新的user对象。
			this.server.maplock.Unlock()
			this.Name = newName //将新的用户名，赋值给User结构体的Name字段。
			this.SendMsg("You have updated your user name" + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//消息格式：	to|张三|消息内容
		//1、获取对方用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsg("Message format error")
			return
		}
		//2、根据用户名，得到对方User对象。
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("The user does not exist")
			return
		}
		//3、获取消息内容，通过对方的User对象将消息内容发送过去。
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("No message content")
			return
		}
		remoteUser.SendMsg(this.Name + "to you say:" + content)
	} else { //当满足第一个条件，但不满足第二个条件时，执行此else中的语句。
		this.server.BroadCast(this, msg)
	}
}

//监听当前user channel的方法，一旦有消息，就直接发送给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C //一旦从通道中可以读取数据，将读取到的数据放入变量msg中。
		_, err := this.conn.Write([]byte(msg + "\n"))
		if err != nil {
			fmt.Println("this.conn.Write err:", err)
			//Write()方法用于把[]byte类型的切片中的数据写入到conn对象中
			return
		}
	}
}
