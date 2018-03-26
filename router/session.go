package main

type Session struct {
	seq     int32                     // 序列号自增标记器
	servers map[int32]int32           // seq:server
	rooms   map[int32]map[int32]int32 // roomid:seq:server with specified room id
}

// NewSession new a session struct. store the seq and serverid.
func NewSession(server int) *Session {
	s := new(Session)
	s.servers = make(map[int32]int32, server)
	s.rooms = make(map[int32]map[int32]int32)
	s.seq = 0
	return s
}

func (s *Session) nextSeq() int32 {
	s.seq++
	return s.seq
}

// Put put a session according with sub key.关联一个subKey和session
func (s *Session) Put(server int32) (seq int32) {
	seq = s.nextSeq()
	s.servers[seq] = server
	return
}

// PutRoom put a session in a room according with subkey.关联一个subKey到comet service 以及room
func (s *Session) PutRoom(server int32, roomId int32) (seq int32) {
	var (
		ok   bool
		room map[int32]int32
	)
	seq = s.Put(server)
	if room, ok = s.rooms[roomId]; !ok {
		room = make(map[int32]int32)
		s.rooms[roomId] = room
	}
	room[seq] = server
	return
}

//返回session关联的所有comet service信息
func (s *Session) Servers() (seqs []int32, servers []int32) {
	var (
		i           = len(s.servers)
		seq, server int32
	)
	seqs = make([]int32, i)
	servers = make([]int32, i)
	for seq, server = range s.servers {
		i--
		seqs[i] = seq
		servers[i] = server
	}
	return
}

// Del delete the session by sub key.删除指定的subKey所关联的Session.Servers
func (s *Session) Del(seq int32) (has, empty bool, server int32) {
	if server, has = s.servers[seq]; has {
		delete(s.servers, seq)
	}
	empty = (len(s.servers) == 0)
	return
}

// DelRoom delete the session and room by subkey. 删除指定的subKey、roomid所关联的Session.rooms
func (s *Session) DelRoom(seq int32, roomId int32) (has, empty bool, server int32) {
	var (
		ok   bool
		room map[int32]int32
	)
	has, empty, server = s.Del(seq)
	if room, ok = s.rooms[roomId]; ok {
		delete(room, seq)
		if len(room) == 0 {
			delete(s.rooms, roomId)
		}
	}
	return
}

//返回session所关联的所有comet service信息
func (s *Session) Count() int {
	return len(s.servers)
}
