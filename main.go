package main

/*
这是还没有进行客户端实现的情况下，已实现功能：
1、群聊
2、修改用户名
3、私聊
*/
func main() {
	server := NewServer("127.0.0.1", 8888)
	server.Start()
}
