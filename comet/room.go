package main

import (
	"goim/libs/proto"
	"sync"
)

type Room struct {
	id     int32     //roomid  唯一
	rLock  sync.RWMutex    //互斥锁
	next   *Channel         // 该房间的所有客户端的Channel 是一个双向链表，复杂度为o(1)，效率比较高。
	drop   bool             // 标示房间是否存活
	Online int // dirty read is ok    房间的channel数量，即房间的在线用户的多少
}

// NewRoom new a room struct, store channel room info.
func NewRoom(id int32) (r *Room) {
	r = new(Room)
	r.id = id
	r.drop = false
	r.next = nil
	r.Online = 0
	return
}

// Put put channel into the room.
func (r *Room) Put(ch *Channel) (err error) {
	r.rLock.Lock()
	if !r.drop {
		if r.next != nil {
			r.next.Prev = ch
		}
		ch.Next = r.next
		ch.Prev = nil
		r.next = ch // insert to header
		r.Online++
	} else {
		err = ErrRoomDroped
	}
	r.rLock.Unlock()
	return
}

// Del delete channel from the room.
func (r *Room) Del(ch *Channel) bool {
	r.rLock.Lock()
	if ch.Next != nil {
		// if not footer
		ch.Next.Prev = ch.Prev
	}
	if ch.Prev != nil {
		// if not header
		ch.Prev.Next = ch.Next
	} else {
		r.next = ch.Next
	}
	r.Online--
	r.drop = (r.Online == 0)
	r.rLock.Unlock()
	return r.drop
}

// Push push msg to the room, if chan full discard it.
func (r *Room) Push(p *proto.Proto) {
	r.rLock.RLock()          //锁是必须的
	for ch := r.next; ch != nil; ch = ch.Next {      //单链表push每一个channel
		ch.Push(p)
	}
	r.rLock.RUnlock()
	return
}

// Close close the room.
func (r *Room) Close() {
	r.rLock.RLock()
	for ch := r.next; ch != nil; ch = ch.Next {
		ch.Close()
	}
	r.rLock.RUnlock()
}
