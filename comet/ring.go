package main

import (
	"goim/libs/proto"

	log "github.com/thinkboy/log4go"
)

//ring是一个环形对象池，用于管理协议对象-proto，rp为可读游标，wp为可写游标，内存的大小为4的整数倍。
type Ring struct {
	// read
	rp   uint64 //可读游标
	num  uint64
	mask uint64
	// TODO split cacheline, many cpu cache line size is 64
	// pad [40]byte
	// write
	wp   uint64 //可写游标
	data []proto.Proto
}

func NewRing(num int) *Ring {
	r := new(Ring)
	r.init(uint64(num))
	return r
}

func (r *Ring) Init(num int) {
	r.init(uint64(num))
}

func (r *Ring) init(num uint64) {
	// 2^N
	if num&(num-1) != 0 {
		for num&(num-1) != 0 {
			num &= (num - 1)
		}
		num = num << 1
	}
	r.data = make([]proto.Proto, num)
	r.num = num
	r.mask = r.num - 1
}

//获取一个proto对象的引用，用于读，不会移动可读游标
func (r *Ring) Get() (proto *proto.Proto, err error) {
	if r.rp == r.wp {
		return nil, ErrRingEmpty
	}
	proto = &r.data[r.rp&r.mask]
	return
}

//移动可读游标
func (r *Ring) GetAdv() {
	r.rp++
	if Debug {
		log.Debug("ring rp: %d, idx: %d", r.rp, r.rp&r.mask)
	}
}

//获取一个proto对象的引用，用于写，不会移动可写游标
func (r *Ring) Set() (proto *proto.Proto, err error) {
	log.Debug("ring wp: %d, rp: %d  rnum: %d", r.wp, r.rp, r.num)
	if r.wp-r.rp >= r.num {
		return nil, ErrRingFull
	}
	proto = &r.data[r.wp&r.mask]
	return
}

//移动可写游标
func (r *Ring) SetAdv() {
	r.wp++
	if Debug {
		log.Debug("ring wp: %d, idx: %d", r.wp, r.wp&r.mask)
	}
}

func (r *Ring) Reset() {
	r.rp = 0
	r.wp = 0
	// prevent pad compiler optimization
	// r.pad = [40]byte{}
}
