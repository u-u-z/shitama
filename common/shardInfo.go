package common

type ShardInfo struct {
	Addr     string `json:"addr"`
	IP       string `json:"ip"`
	EchoPort int    `json:"echoPort"`
}
