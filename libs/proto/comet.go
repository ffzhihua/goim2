package proto

//Push RPC模块

//心跳：
type NoArg struct {
}

//心跳：
type NoReply struct {
}

// 单播
type PushMsgArg struct {
	Key string //subKey
	P   Proto
}

// 把某条消息推送给多个subKey
type PushMsgsArg struct {
	Key    string
	PMArgs []*PushMsgArg
}

type PushMsgsReply struct {
	Index int32
}

// 多播
type MPushMsgArg struct {
	Keys []string
	P    Proto
}

type MPushMsgReply struct {
	Index int32
}

// 广播
type MPushMsgsArg struct {
	PMArgs []*PushMsgArg
}

type MPushMsgsReply struct {
	Index int32
}

type BoardcastArg struct {
	P Proto
}

// 把某条消息推送给某个房间的所有channels
type BoardcastRoomArg struct {
	RoomId int32
	P      Proto
}

type RoomsReply struct {
	RoomIds map[int32]struct{}
}
