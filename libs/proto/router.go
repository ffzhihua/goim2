package proto

//增加用户：
type PutArg struct {
	UserId int64
	Server int32
	RoomId int32
}

type PutReply struct {
	Seq int32 // 序列号
}

//移除用户：
type DelArg struct {
	UserId int64
	Seq    int32
	RoomId int32
}

type DelReply struct {
	Has bool //	是否存在目标用户
}

// 剔除comet server
type DelServerArg struct {
	Server int32
}

// 获取用户信息
type GetArg struct {
	UserId int64
}

// 获取Router的所有信息
type GetReply struct {
	Seqs    []int32
	Servers []int32
}

type GetAllReply struct {
	UserIds  []int64
	Sessions []*GetReply
}

type MGetArg struct {
	UserIds []int64
}

type MGetReply struct {
	UserIds  []int64
	Sessions []*GetReply
}

// 返回所有连接个数
type CountReply struct {
	Count int32
}

// 获取特定房间的所有连接
type RoomCountArg struct {
	RoomId int32
}

type RoomCountReply struct {
	Count int32
}

// 获取所有房间的连接个数
type AllRoomCountReply struct {
	Counter map[int32]int32
}

// 获取所有的comet server个数
type AllServerCountReply struct {
	Counter map[int32]int32
}

// 获取所有的用户个数
type UserCountArg struct {
	UserId int64
}

type UserCountReply struct {
	Count int32
}
