package main

import (
	"bufio"
	// "encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	proto1 "goim/libs/proto"
	"math/rand"
	"net"
	"os"
	"time"
)

//数据包类型
const (
	HEART_BEAT_PACKET = 0x00
	REPORT_PACKET     = 0x01
)

//默认的服务器地址
var (
	server = "127.0.0.1:8089"
)

//客户端对象
type TcpClient struct {
	connection *net.TCPConn
	hawkServer *net.TCPAddr
	stopChan   chan struct{}
}

func main() {
	//拿到服务器地址信息
	hawkServer, err := net.ResolveTCPAddr("tcp", server)
	if err != nil {
		fmt.Printf("hawk server [%s] resolve error: [%s]", server, err.Error())
		os.Exit(1)
	}
	//连接服务器
	connection, err := net.DialTCP("tcp", nil, hawkServer)
	if err != nil {
		fmt.Printf("connect to hawk server error: [%s]", err.Error())
		os.Exit(1)
	}
	client := &TcpClient{
		connection: connection,
		hawkServer: hawkServer,
		stopChan:   make(chan struct{}),
	}
	//启动接收
	go client.receivePackets()

	//发送心跳的goroutine
	go func() {
		heartBeatTick := time.Tick(2 * time.Second)
		for {
			select {
			case <-heartBeatTick:
				client.sendHeartPacket()
			case <-client.stopChan:
				return
			}
		}
	}()

	//测试用的，开300个goroutine每秒发送一个包
	for i := 0; i < 1; i++ {
		go func() {
			sendTimer := time.After(2 * time.Second)
			for {
				select {
				case <-sendTimer:
					client.sendReportPacket()
					sendTimer = time.After(1 * time.Second)
				case <-client.stopChan:
					return
				}
			}
		}()
	}
	//等待退出
	<-client.stopChan
}

// 接收数据包
func (client *TcpClient) receivePackets() {
	reader := bufio.NewReader(client.connection)
	for {
		//承接上面说的服务器端的偷懒，我这里读也只是以\n为界限来读区分包
		msg, err := reader.ReadString('\n')
		if err != nil {
			//在这里也请处理如果服务器关闭时的异常
			close(client.stopChan)
			break
		}
		fmt.Print(msg)
	}
}

//发送数据包{\"ver\":1,\"op\":7,\"seq\":0,\"body\":{\"test\":1111}}
//仔细看代码其实这里做了两次json的序列化，有一次其实是不需要的
func (client *TcpClient) sendReportPacket() {
	var (
		ver       = int32(1)
		operation = int32(7)
		seqid     = int32(1)
		body      = []byte("{\"test\":1111}")
	)

	//这一次其实可以不需要，在封包的地方把类型和数据传进去即可
	packet := &proto1.ImProto{
		Ver:       &ver,
		Operation: &operation,
		SeqId:     &seqid,
		Body:      body,
	}
	sendBytes, err := proto.Marshal(packet)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("message-len: %s\n", len(sendBytes))
	//发送
	client.connection.Write(sendBytes)
	fmt.Println("Send metric data success!")
}

//发送心跳包，与发送数据包一样
func (client *TcpClient) sendHeartPacket() {

	var (
		ver       = int32(1)
		operation = int32(2)
		seqid     = int32(2)
		body      = []byte("{\"test\":1111}")
	)
	reportPacket := &proto1.ImProto{
		Ver:       &ver,
		Operation: &operation,
		SeqId:     &seqid,
		Body:      body,
	}
	sendBytes, err := proto.Marshal(reportPacket)
	fmt.Println("sharkhand:%s", reportPacket)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("sharkhand-len: %s\n", len(sendBytes))
	client.connection.Write(sendBytes)
	fmt.Println("Send heartbeat data success!")
}

//拿一串随机字符
func getRandString() string {
	length := rand.Intn(50)
	strBytes := make([]byte, length)
	for i := 0; i < length; i++ {
		strBytes[i] = byte(rand.Intn(26) + 97)
	}
	return string(strBytes)
}
