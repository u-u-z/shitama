package client

import (
	"github.com/evshiron/shitama/common"
)

type ShardInfo struct {
	common.ShardInfo
	RTT float32 `json:"rtt"`
}
