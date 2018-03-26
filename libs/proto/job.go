package proto

// TODO optimize struct after replace kafka
type KafkaMsg struct {
	OP       string   `json:"op"`               //操作类型
	RoomId   int32    `json:"roomid,omitempty"` //房间号
	ServerId int32    `json:"server,omitempty"` //comet id
	SubKeys  []string `json:"subkeys,omitempty"`
	Msg      []byte   `json:"msg"`
	Ensure   bool     `json:"ensure,omitempty"` //是否强推送(伪强推)
}
