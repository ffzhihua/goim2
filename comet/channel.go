package main

import (
	"goim/libs/bufio"
	"goim/libs/proto"
)

// Channel used by message pusher send msg to write goroutine.
//消息推到channel(长连接通道)中 其实就是Session与语言层的channel不一样
type Channel struct {
	RoomId   int32 //每一个channel都有一个roomid
	CliProto Ring  //客户端proto
	signal   chan *proto.Proto
	Writer   bufio.Writer
	Reader   bufio.Reader
	Next     *Channel
	Prev     *Channel
}

func NewChannel(cli, svr int, rid int32) *Channel {
	c := new(Channel)
	c.RoomId = rid
	c.CliProto.Init(cli)
	c.signal = make(chan *proto.Proto, svr)
	return c
}

// Push server push message.
func (c *Channel) Push(p *proto.Proto) (err error) {
	select {
	case c.signal <- p:
	default:
	}
	return
}

// Ready check the channel ready or close?
func (c *Channel) Ready() *proto.Proto {
	return <-c.signal
}

// Signal send signal to the channel, protocol ready.
func (c *Channel) Signal() {
	c.signal <- proto.ProtoReady
}

// Close close the channel.
func (c *Channel) Close() {
	c.signal <- proto.ProtoFinish
}
