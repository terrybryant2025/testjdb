package main

import (
	cmap "github.com/orcaman/concurrent-map/v2"
	"time"
)

var RicherUserData cmap.ConcurrentMap[string, *RicherBettingInfo] // 全局存储用户数据

type RicherBettingInfo struct {
	ChannelId  int64  // 渠道ID
	Pid        int64  // 玩家ID
	Nickname   string // 玩家别名，暂时写死了
	AccountId  string // 玩家信息
	Currency   string // 货币类型
	PlayerType int64  // 玩家类型 1.正常账号  2.试玩账号

	Balance    float64 // 余额
	MaxWin     float64 // 最大赢
	Rtp        int64   // 当前RTP
	RtpLevel   int64   // Rtp等级
	ChannelRtp int64   // 渠道RTP
	BaseRate   float64 // 当前基础倍率
	Win        float64 // 当前赢钱
	IsOffline  bool    // 是否离线
	Token      string  // 用户token
	AutoBet    bool    // 是否自动下注
	Timer      *time.Timer
}
