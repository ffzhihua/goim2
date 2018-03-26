package main

import (
	"github.com/garyburd/redigo/redis"
	log "github.com/thinkboy/log4go"
	"goim/libs/define"
	//"strconv"
)

// developer could implement "Auth" interface for decide how get userId, or roomId
type Auther interface {
	Auth(token string) (userId int64, roomId int32)
}

type DefaultAuther struct {
}

func NewDefaultAuther() *DefaultAuther {
	return &DefaultAuther{}
}

// func (a *DefaultAuther) Auth(token string) (userId int64, roomId int32) {
// 	var err error
// 	if userId, err = strconv.ParseInt(token, 10, 64); err != nil {
// 		userId = 0
// 		roomId = define.NoRoom
// 	} else {
// 		roomId = 1 // only for debug
// 	}
// 	return
// }
func (a *DefaultAuther) Auth(token string) (userId int64, roomId int32) {
	var err error
	c, err := redis.Dial("tcp", "127.0.0.1:6380")
	if err != nil {
		log.Info("redis connect: \"%s\"", err)
		return
	}
	// //密码授权
	// c.Do("AUTH", "123456")
	// c.Do("SET", "a", 134)
	var key = "p_BOCACHE.action.session_data_teacher.session_id_" + token
	session_raw, err := redis.String(c.Do("HGET", key, "raw"))
	log.Info("session_raw: \"%s\"", session_raw)
	defer c.Close()
	if session_raw == "" {
		userId = 0
		roomId = define.NoRoom
	} else {
		res := a.ELFHash(session_raw)
		userId = int64(res)
		roomId = 1
	}
	return
}

func (a *DefaultAuther) ELFHash(str string) (key uint) {
	var hash uint = 0
	var x uint = 0
	for _, v := range []byte(str) {
		hash = (hash << 4) + uint(v)
		if x = hash & 0xF0000000; x != 0 {
			hash ^= (x >> 24)
			hash &= ^x
		}
	}
	return hash & 0x7FFFFFFF
}
