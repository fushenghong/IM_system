package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

//创建一个客户端结构体
type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	Conn       net.Conn
	flag       int
}

func NewClient(serverIp string, ServerPort int) *Client {
	//创建客户端对象
	client := &Client{ //字段Name和字段Conn未进行赋值操作。
		ServerIp:   serverIp,
		ServerPort: ServerPort,
		flag:       999,
	}
	//连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, ServerPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.Conn = conn //将客户端与服务器连接得到的conn赋值给Client中的Conn字段。
	//返回对象
	return client
}

//客户端菜单显示
func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.群聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")
	_, err := fmt.Scanln(&flag)
	if err != nil {
		fmt.Println("从键盘上输入字符错误：", err)
		return false
	}
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>>>>>>请输入合法范围内的数字>>>>>>>>>>>")
		return false
	}
}

//修改用户名
func (client *Client) UpdateName() bool { //返回一个bool类型的值，来表示修改用户名是否成功。
	fmt.Println(">>>>>>>请输入用户名:")
	fmt.Scanln(&client.Name)
	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.Conn.Write([]byte(sendMsg)) //将客户端输入的内容，写入连接中。服务器能从连接中读取客户端的信息。
	if err != nil {
		fmt.Println("client Conn Write err:", err)
		return false
	}
	return true
}

//选择群聊模式
func (client *Client) PublicChat() {
	//提示用户输入信息
	var chatMsg string
	fmt.Println(">>>>>>>请输入聊天内容，exit退出。。。")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		//发给服务端
		//消息不为空则发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.Conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn write err:", err)
				break //跳出for循环
			}
		}
		chatMsg = ""
		fmt.Println(">>>>>>>>请输入聊天内容，exit退出。。。")
		fmt.Scanln(&chatMsg)
	}
}

//查询在线用户
func (client *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := client.Conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn write err:", err)
		return
	}
}

//选择私聊模式
func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string
	client.SelectUsers()
	fmt.Println(">>>>>>>>请输入聊天对象[用户名],exit退出:")
	fmt.Scanln(&remoteName)
	for remoteName != "exit" {
		fmt.Println(">>>>>请输入消息内容，exit退出：")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			//消息不为空则发送
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.Conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn write err:", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println(">>>>>>>>>>请输入消息内容，exit退出：")
			fmt.Scanln(&chatMsg)
		}
		client.SelectUsers()
		fmt.Println(">>>>>>>>请输入聊天对象[用户名],exit退出:")
		fmt.Scanln(&remoteName)
	}
}

//处理server回应的消息，直接显示到标准输出即可。
func (client *Client) DealResponse() {
	//一旦client.conn有数据，就直接copy到stdout标准输出上，永久阻塞监听。
	io.Copy(os.Stdout, client.Conn)
}

//定义一个方法用来处理client端业务
func (client *Client) Run() {
	for client.flag != 0 { //client.flag默认是999，如果输入0，会将client.flag的值变为0。此时，很明显不满足
		//外层for循环的条件，client.Run()方法结束，main主函数结束，客户端结束。
		//如果是输入65，返回false，client.flag=65。满足外层for循环的条件，接着里面的for循环条件满足，执行for循环里面的内容。里面内容
		//为空。往下执行，没有满足条件的代码，Run()方法，结束。
		for client.menu() != true {
		}
		//根据不同的模式处理不同的业务
		switch client.flag {
		case 1:
			//群聊模式
			fmt.Println("群聊模式选择。。。")
			client.PublicChat()
			break //break主要是用来跳出离它最近的for循环的。
		case 2:
			//私聊模式
			fmt.Println("私聊模式选择。。。")
			client.PrivateChat()
			break //break主要是用来跳出离它最近的for循环的。
		case 3:
			//更新用户名
			fmt.Println("更新用户名选择。。。")
			client.UpdateName()
			break //break主要是用来跳出离它最近的for循环的。
		}
	}
}

//设置两个全局变量
var serverIp string
var serverPort int

//通过./client -ip 127.0.0.1 -port 8888可以来连接服务器。相当于nc 127.0.0.1 8888了。
func init() { //init函数在main函数之前执行。
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器ip地址(默认是127.0.0.1)")
	//init函数的主要作用就是进行参数的绑定
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口号(默认是8888)")
}

func main() {
	//命令行解析
	flag.Parse() //只有先进行命令行解析，才能进行后续步骤。
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>>连接服务器失败。。。")
		return
	}
	go client.DealResponse()
	fmt.Println(">>>>>>>>>连接服务器成功。。。")
	//启动客户端的业务
	//select {} //让程序在此阻塞，让main主函数不会结束。
	client.Run()
}
