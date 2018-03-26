package main

import (
	"goim/libs/hash/ketama"
	inet "goim/libs/net"
	"goim/libs/net/xrpc"
	"goim/libs/proto"
	"strconv"

	log "github.com/thinkboy/log4go"
	"strings"
)

var (
	routerServiceMap = map[string]*xrpc.Clients{}
	routerRing       *ketama.HashRing
)

const (
	routerService               = "RouterRPC"
	routerServicePing           = "RouterRPC.Ping"
	routerServicePut            = "RouterRPC.Put"
	routerServiceDel            = "RouterRPC.Del"
	routerServiceDelServer      = "RouterRPC.DelServer"
	routerServiceAllRoomCount   = "RouterRPC.AllRoomCount"
	routerServiceAllServerCount = "RouterRPC.AllServerCount"
	routerServiceGet            = "RouterRPC.Get"
	routerServiceMGet           = "RouterRPC.MGet"
	routerServiceGetAll         = "RouterRPC.GetAll"
)

func InitRouter(addrs map[string]string) (err error) {
	var (
		network, addr string
	)
	routerRing = ketama.NewRing(ketama.Base)
	for serverId, bind := range addrs {
		var rpcOptions []xrpc.ClientOptions
		for _, bind = range strings.Split(bind, ",") {
			if network, addr, err = inet.ParseNetwork(bind); err != nil {
				log.Error("inet.ParseNetwork() error(%v)", err)
				return
			}
			options := xrpc.ClientOptions{
				Proto: network,
				Addr:  addr,
			}
			rpcOptions = append(rpcOptions, options)
		}
		// rpc clients
		rpcClient := xrpc.Dials(rpcOptions)
		// ping & reconnect
		rpcClient.Ping(routerServicePing)
		routerRing.AddNode(serverId, 1)
		routerServiceMap[serverId] = rpcClient
		log.Info("router rpc connect: %v ", rpcOptions)
	}
	routerRing.Bake()
	return
}

//通过router serverid获取对应的router service client
func getRouterByServer(server string) (*xrpc.Clients, error) {
	if client, ok := routerServiceMap[server]; ok {
		return client, nil
	} else {
		return nil, ErrRouter
	}
}

//通过userID获取router service client
func getRouterByUID(userID int64) (*xrpc.Clients, error) {
	return getRouterByServer(routerRing.Hash(strconv.FormatInt(userID, 10)))
}

//通过userID获取到Router 节点
func getRouterNode(userID int64) string {
	return routerRing.Hash(strconv.FormatInt(userID, 10))
}

//注册、验证用户，返回一个自增序列号(cometServiceID_incr)
func connect(userID int64, server, roomId int32) (seq int32, err error) {
	var (
		args   = proto.PutArg{UserId: userID, Server: server, RoomId: roomId}
		reply  = proto.PutReply{}
		client *xrpc.Clients
	)
	if client, err = getRouterByUID(userID); err != nil {
		return
	}
	if err = client.Call(routerServicePut, &args, &reply); err != nil {
		log.Error("c.Call(\"%s\",\"%v\") error(%v)", routerServicePut, args, err)
	} else {
		seq = reply.Seq
	}
	return
}

//注销一个用户
func disconnect(userID int64, seq, roomId int32) (has bool, err error) {
	var (
		args   = proto.DelArg{UserId: userID, Seq: seq, RoomId: roomId}
		reply  = proto.DelReply{}
		client *xrpc.Clients
	)
	if client, err = getRouterByUID(userID); err != nil {
		return
	}
	if err = client.Call(routerServiceDel, &args, &reply); err != nil {
		log.Error("c.Call(\"%s\",\"%v\") error(%v)", routerServiceDel, args, err)
	} else {
		has = reply.Has
	}
	return
}

//下线一个comet service
func delServer(server int32) (err error) {
	var (
		args   = proto.DelServerArg{Server: server}
		reply  = proto.NoReply{}
		client *xrpc.Clients
	)
	for _, client = range routerServiceMap {
		if err = client.Call(routerServiceDelServer, &args, &reply); err != nil {
			log.Error("c.Call(\"%s\",\"%v\") error(%v)", routerServiceDelServer, args, err)
		}
	}
	return
}

//获取所有房间个数
func allRoomCount(client *xrpc.Clients) (counter map[int32]int32, err error) {
	var (
		args  = proto.NoArg{}
		reply = proto.AllRoomCountReply{}
	)
	if err = client.Call(routerServiceAllRoomCount, &args, &reply); err != nil {
		log.Error("c.Call(\"%s\", nil) error(%v)", routerServiceAllRoomCount, err)
	} else {
		counter = reply.Counter
	}
	return
}

//获取所有comet-service
func allServerCount(client *xrpc.Clients) (counter map[int32]int32, err error) {
	var (
		args  = proto.NoArg{}
		reply = proto.AllServerCountReply{}
	)
	if err = client.Call(routerServiceAllServerCount, &args, &reply); err != nil {
		log.Error("c.Call(\"%s\", nil) error(%v)", routerServiceAllServerCount, err)
	} else {
		counter = reply.Counter
	}
	return
}

//通过用户ID生成subKey；同一个用户可以同时多处登陆或者同处多实例登陆，它们都会被同等对待。
func genSubKey(userId int64) (res map[int32][]string) {
	var (
		err    error
		i      int
		ok     bool
		key    string
		keys   []string
		args   = proto.GetArg{UserId: userId}
		reply  = proto.GetReply{}
		client *xrpc.Clients
	)
	res = make(map[int32][]string)
	if client, err = getRouterByUID(userId); err != nil {
		return
	}
	if err = client.Call(routerServiceGet, &args, &reply); err != nil {
		log.Error("client.Call(\"%s\",\"%v\") error(%v)", routerServiceGet, args, err)
		return
	}
	for i = 0; i < len(reply.Servers); i++ {
		key = encode(userId, reply.Seqs[i])
		if keys, ok = res[reply.Servers[i]]; !ok {
			keys = []string{}
		}
		keys = append(keys, key)
		res[reply.Servers[i]] = keys
	}
	return
}

//并行获取多个用户信息，返回值为map[comet.serverId][]subkey.
func getSubKeys(res chan *proto.MGetReply, serverId string, userIds []int64) {
	var (
		args  = proto.MGetArg{UserIds: userIds}
		reply = proto.MGetReply{}
	)
	if client, err := getRouterByServer(serverId); err == nil {
		if err = client.Call(routerServiceMGet, &args, &reply); err != nil {
			log.Error("client.Call(\"%s\",\"%v\") error(%v)", routerServiceMGet, args, err)
			res <- nil
		}
	}
	res <- &reply
}

func genSubKeys(userIds []int64) (divide map[int32][]string) {
	var (
		i, j, k      int
		node, subkey string
		subkeys      []string
		server       int32
		session      *proto.GetReply
		reply        *proto.MGetReply
		uid          int64
		ids          []int64
		ok           bool
		m            = make(map[string][]int64)
		res          = make(chan *proto.MGetReply, 1)
	)
	divide = make(map[int32][]string) //map[comet.serverId][]subkey
	for i = 0; i < len(userIds); i++ {
		node = getRouterNode(userIds[i])
		if ids, ok = m[node]; !ok {
			ids = []int64{}
		}
		ids = append(ids, userIds[i])
		m[node] = ids
	}
	for node, ids = range m {
		go getSubKeys(res, node, ids)
	}
	k = len(m)
	for k > 0 {
		k--
		if reply = <-res; reply == nil {
			continue
		}
		for j = 0; j < len(reply.UserIds); j++ {
			session = reply.Sessions[j]
			uid = reply.UserIds[j]
			for i = 0; i < len(session.Seqs); i++ {
				subkey = encode(uid, session.Seqs[i])
				server = session.Servers[i]
				if subkeys, ok = divide[server]; !ok {
					subkeys = []string{subkey}
				} else {
					subkeys = append(subkeys, subkey)
				}
				divide[server] = subkeys
			}
		}
	}
	return
}
