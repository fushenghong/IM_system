package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int
	//在线用户的列表
	OnlineMap map[string]*User
	maplock   sync.RWMutex //添加一个读写互斥锁
	//消息广播的channel
	Message chan string //通道里面传输的数据类型是字符串类型。该通道属于server端通道。
}

//创建一个server的接口，也可以把NewServer()函数理解成构造函数，主要是起初始化作用。该函数主要作用是初始化一个server对象。
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User), //只为OnlineMap字段，开辟了内存空间，但是没有进行赋值。
		Message:   make(chan string),      //同上，只开辟了内存空间未进行赋值。
	}
	return server //将创建的server对象，返回到函数调用处。
}

//ListenMessager()函数，加上go关键字便是一个goroutine了。主要用来监听Message这个server端channel，一旦有消息就发送给全部的在线user。
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message //从Message信道读取到信息，赋值给变量msg。
		//将msg发送给全部在线user
		this.maplock.Lock()                  //上锁
		for _, cli := range this.OnlineMap { //遍历用户列表，将msg信息发送给所有的user。
			cli.C <- msg //将从Message信道中读取到的信息，在写入客户端信道。每一个客户端都有一个channel。
		}
		this.maplock.Unlock() //解锁
	}
}

//广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg //字符串的拼接，使用+进行字符串的拼接
	this.Message <- sendMsg                                  //将信息写入通道Message。
}

//连接成功之后，进行业务的处理
func (this *Server) Handler(conn net.Conn) {
	//...当前连接的业务
	//fmt.Println("链接建立成功")
	user := NewUser(conn, this) //得到一个用户对象
	////用户上线，将用户加入到onlineMap中
	//this.maplock.Lock() //上锁
	//this.OnlineMap[user.Name] = user
	//this.maplock.Unlock() //解锁
	////广播当前用户上线消息
	//this.BroadCast(user, "Online success")
	user.Online() //用户上线业务

	////监听用户是否活跃的channel
	//isLive := make(chan bool)
	//接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096) //创建一个缓冲区
		for {
			n, err := conn.Read(buf) //将读取到的信息写入缓冲区中。并返回读取的字节数和错误信息
			if n == 0 {
				//this.BroadCast(user, "下线") //如果读取的字节数为0，则向所有的客户端广播用户下线
				user.Offline() //用户下线业务
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("conn read err:", err)
				return
			}
			//提取用户的消息（去除'\n'）
			msg := string(buf[:n-1]) //将字节切片转换为字符串
			////将得到的消息进行广播
			//this.BroadCast(user, msg)
			//用户针对msg进行消息处理
			user.DoMessage(msg)
			////用户的任意消息，代表当前用户是一个活跃的。
			//isLive <- true
		}
	}()
	//当前handler阻塞
	//for {
	select {}
	//	case <-isLive:
	//	//当前用户是活跃的，应该重置定时器。
	//	//不做任何事情，为了激活select，更新下面的定时器。
	//
	//	case <-time.After(time.Second * 10):
	//		//已经超时
	//		//将当前的User强制关闭
	//		user.SendMsg("You have been kicked")
	//		//销毁用的资源
	//		close(user.C)
	//		//退出当前Handler
	//		return
	//	}
	//}
}

//启动服务器的接口
func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	//fmt.Sprintf("%s:%d", this.Ip, this.Port)相当于格式化输入，这是一个字符串。相当于127.0.0.1:8888
	if err != nil {
		fmt.Println("net.listen err:", err)
		return
	}
	//close listen socket
	defer listener.Close()
	//启动监听Message的goroutine
	go this.ListenMessager() //使用go关键字启动一个协程，该协程用来监听Message通道。若该通道有任何信息向所有User广播。
	for {
		//当调用监听器的Accept方法时，流程会被阻塞，直到某个客户端程序与当前程序建立TCP连接。
		//此时，Accept方法会返回两个结果值：第一个结果值代表了当前TCP连接的net.Conn类型值，
		//而第二个结果值依然是一个error类型的值。
		conn, err := listener.Accept() //也就是说，当得到conn相当于客户端已经和服务器建立连接了。
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue //跳出此次for循环
		}
		//do handler
		go this.Handler(conn) //tcp连接建立完成之后，通过开启一个协程来进行业务处理。
	}
}
