package proto

// 用于comet发送客户端的校验信息
type ConnArg struct {
	Token  string
	Server int32
}

// logic 校验应答
type ConnReply struct {
	Key    string
	RoomId int32
}

// 用于comet发送客户端连接下线
type DisconnArg struct {
	Key    string
	RoomId int32
}

// 应答客户端下线
type DisconnReply struct {
	Has bool
}
