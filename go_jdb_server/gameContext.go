package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	EAviatorStageZero    = 0
	EAviatorStageBet     = 1
	EAviatorStageCashOut = 2
	EAviatorStageReady   = 3
)

const (
	BET_TIME      = 5 * time.Second
	CASH_OUT_TIME = 5 * time.Second
	AWARD_TIME    = 3 * time.Second
	READY_TIME    = 5 * time.Second
)

type PlayerBetSt struct {
	BetArea   uint32
	BetValue  float64
	IsCashout bool
}

type AviatorPlayerInfo struct {
	ChannelId    int64  // 渠道ID
	Pid          int64  // 玩家ID
	Nickname     string // 玩家别名，暂时写死了
	AccountId    string // 玩家信息
	Currency     string // 货币类型
	PlayerType   int64  // 玩家类型 1.正常账号  2.试玩账号
	ProfileImage string

	Balance    float64 // 余额
	Rtp        int64   // 当前RTP
	RtpLevel   int64   // Rtp等级
	ChannelRtp int64   // 渠道RTP
	IsOffline  bool    // 是否离线
	Token      string  // 用户token
	AutoBet    bool    // 是否自动下注

	BetList []*PlayerBetSt

	conn  *websocket.Conn
	mutex sync.Mutex
}

type AviatorGameContext struct {
	players map[string]*AviatorPlayerInfo

	RecordId          int  // 牌局号(每次下一局累加1)
	isRunning         bool // 是否已启动
	Timer             *time.Timer
	curStateStartTime int64 //当前阶段开始时间
	CurStage          int32 //当前阶段
	CurMultiplier     float64

	CashOuts     []CashOut
	TotalCashOut float64
	TotalBet     float64
	IsAwarding   bool
}

func StructToMap(obj interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// 将 map[string]interface{} 转换为结构体
func MapToStruct(data map[string]interface{}, result interface{}) error {
	// 将 map 转换为 JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("map to json error: %v", err)
	}

	// 将 JSON 解析到结构体
	if err := json.Unmarshal(jsonData, result); err != nil {
		return fmt.Errorf("json to struct error: %v", err)
	}

	return nil
}

func NewGameContext() *AviatorGameContext {
	return &AviatorGameContext{
		players:           make(map[string]*AviatorPlayerInfo, 0),
		curStateStartTime: 0,
		CurStage:          int32(EAviatorStageZero),
		CurMultiplier:     1.0,
	}
}
func (g *AviatorGameContext) Init() {
	g.StartTimer(1000*time.Millisecond, g.OnTick)
}

func (g *AviatorGameContext) NewGameInit() {
	g.CurStage = EAviatorStageReady
	g.curStateStartTime = time.Now().UnixMilli()
	g.TotalBet = 0
	g.TotalCashOut = 0
	g.CurMultiplier = 1.0
	g.CashOuts = make([]CashOut, 0)
	g.IsAwarding = false
}

func (g *AviatorGameContext) OnLogin(conn *websocket.Conn, obj map[string]interface{}) {
	playerInfo := &AviatorPlayerInfo{
		conn:    conn,
		Balance: 10000,
	}
	g.players[conn.RemoteAddr().String()] = playerInfo
}
func (g *AviatorGameContext) OnRecv(conn *websocket.Conn, obj map[string]interface{}) {
	switch obj["c"] {
	case "betHandler":
		var result BetRequest
		params, _ := obj["p"].(map[string]interface{})
		if err := MapToStruct(params, &result); err != nil {
			return
		}
		g.C2sBet(conn, &result)
	case "cashOutHandler":
		var result CashOutRequest
		params, _ := obj["p"].(map[string]interface{})
		if err := MapToStruct(params, &result); err != nil {
			return
		}
		g.C2sCashOut(conn, &result)
	case "currentBetsInfoHandler":
		g.CurrentBetsInfo(conn)
	case "previousRoundInfo":
	case "getHugeWinsInfo":
	case "getTopRoundsInfo":
	case "getTopWinsInfo":
	case "betHistory":
	default:
		fmt.Printf("⚠️ 未知扩展命令: %s\n", obj["c"])
	}
}

/*
func (g *AviatorGameContext) BetHistory(conn *websocket.Conn) {

}

func (g *AviatorGameContext) RoundFairness(conn *websocket.Conn) {

}

func (g *AviatorGameContext) getHugeWinsInfo(c context.Context) {

}

func (g *AviatorGameContext) getTopWinsInfo(c context.Context) {

}

func (g *AviatorGameContext) getTopRoundsInfo(c context.Context) {

}

func (g *AviatorGameContext) PreviousRoundInfo(c context.Context) {

}
*/
func (g *AviatorGameContext) C2sBet(conn *websocket.Conn, req *BetRequest) {
	playerInfo := g.players[conn.RemoteAddr().String()]
	if playerInfo == nil {
		return
	}

	if g.CurStage != EAviatorStageBet {
		return
	}

	if req.Bet <= 0 || req.Bet > playerInfo.Balance || req.BetID <= 0 || req.BetID > 2 {
		return
	}

	betSt := g.Id2Bet(int32(req.BetID), playerInfo)
	if betSt != nil {
		return
	}

	//扣钱
	playerInfo.Balance -= req.Bet
	g.S2cNewBalance(conn, playerInfo.Balance)

	g.TotalBet += req.Bet
	playerInfo.BetList = append(playerInfo.BetList, &PlayerBetSt{
		BetArea:  uint32(req.BetID),
		BetValue: req.Bet,
	})

	betResponse := &BetResponse{
		Bet:          req.Bet,
		BetID:        req.BetID,
		FreeBet:      req.FreeBet,
		PlayerID:     playerInfo.AccountId,
		ProfileImage: playerInfo.ProfileImage,
		Username:     playerInfo.Nickname,
	}

	result, _ := StructToMap(betResponse)
	g.SendToClient(conn, "bet", result)
}

func (g *AviatorGameContext) Id2Bet(betId int32, playerInfo *AviatorPlayerInfo) *PlayerBetSt {
	for idx, bet := range playerInfo.BetList {
		if bet.BetArea == uint32(betId) {
			return playerInfo.BetList[idx]
		}
	}
	return nil
}

func (g *AviatorGameContext) C2sCashOut(conn *websocket.Conn, req *CashOutRequest) {
	playerInfo := g.players[conn.RemoteAddr().String()]
	if playerInfo == nil {
		return
	}

	if g.CurStage != EAviatorStageCashOut {
		return
	}

	betSt := g.Id2Bet(int32(req.BetID), playerInfo)
	if betSt == nil {
		return
	}

	if betSt.IsCashout {
		return
	}
	CurMultiplier := g.CurMultiplier

	//加钱
	playerInfo.Balance += betSt.BetValue * CurMultiplier
	g.S2cNewBalance(conn, playerInfo.Balance)

	betSt.IsCashout = true
	g.TotalCashOut += betSt.BetValue * CurMultiplier
	g.CashOuts = append(g.CashOuts, CashOut{
		BetID:      req.BetID,
		Multiplier: CurMultiplier,
		PlayerID:   playerInfo.AccountId,
		WinAmount:  betSt.BetValue * CurMultiplier,
	})

	cashOut := &CashOut{
		PlayerID:   playerInfo.AccountId,
		BetID:      req.BetID,
		Multiplier: g.CurMultiplier,
	}

	result, _ := StructToMap(cashOut)
	g.SendToClient(conn, "cashOut", result)

}
func (g *AviatorGameContext) CurrentBetsInfo(conn *websocket.Conn) {
	currentBetsInfo := &CurrentBetsInfo{
		BetsCount:              88,
		OpenBetsCount:          56,
		Code:                   200,
		CashOuts:               []CashOut{},
		Bets:                   []Bet{},
		ActivePlayersCount:     333,
		TopPlayerProfileImages: []string{},
		TotalCashOut:           666.5,
	}

	currentBetsInfo.CashOuts = append(currentBetsInfo.CashOuts, g.CashOuts...)

	result, _ := StructToMap(currentBetsInfo)
	g.SendToClient(conn, "currentBetsInfo", result)
}

func (g *AviatorGameContext) S2cUpdateCurrentCashOuts() {
	ntf := &UpdateCurrentCashOuts{
		Code:                   200,
		TotalCashOut:           g.TotalCashOut,
		OpenBetsCount:          int(g.TotalBet),
		ActivePlayersCount:     0,
		TopPlayerProfileImages: []string{},
		CashOuts:               []CashOut{},
	}

	ntf.CashOuts = append(ntf.CashOuts, g.CashOuts...)
	result, _ := StructToMap(ntf)
	g.SendToAllClients("updateCurrentCashOuts", result)
}

func (g *AviatorGameContext) S2cUpdateCurrentBets() {
	ntf := &UpdateCurrentBets{
		BetsCount:              33,
		Code:                   200,
		ActivePlayersCount:     100,
		Bets:                   []Bet{},
		TopPlayerProfileImages: []string{},
	}
	result, _ := StructToMap(ntf)
	g.SendToAllClients("updateCurrentBets", result)
}

func (g *AviatorGameContext) S2cRoundChartInfo() {
	ntf := &RoundChartInfo{
		Code:          200,
		MaxMultiplier: g.CurMultiplier,
		RoundId:       g.RecordId,
	}
	result, _ := StructToMap(ntf)
	g.SendToAllClients("roundChartInfo", result)
}

func (g *AviatorGameContext) S2cUpdateX() {
	ntf := &UpdateX{
		Code: 200,
		X:    g.CurMultiplier,
	}
	result, _ := StructToMap(ntf)
	g.SendToAllClients("x", result)
}
func (g *AviatorGameContext) S2cOnlinePlayers() {
	onlinePlayers := 0
	for _, player := range g.players {
		if !player.IsOffline {
			onlinePlayers++
		}
	}
	ntf := &OnlinePlayers{
		Code:          200,
		OnlinePlayers: onlinePlayers,
	}
	result, _ := StructToMap(ntf)
	g.SendToAllClients("onlinePlayers", result)
}

func (g *AviatorGameContext) S2cChangeState(newStatus int32) {
	ntf := &ChangeState{
		Code:       200,
		NewStateID: int(newStatus),
		RoundID:    int64(g.RecordId),
	}

	if newStatus == EAviatorStageBet {
		ntf.ServerTime = time.Now().UnixMilli()
		ntf.BetStateEndTime = ntf.ServerTime + BET_TIME.Milliseconds()
	}
	result, _ := StructToMap(ntf)
	g.SendToAllClients("changeState", result)
}

func (g *AviatorGameContext) S2cNewBalance(conn *websocket.Conn, balance float64) {
	ntf := &NewBalance{
		Code:       200,
		NewBalance: balance,
	}

	result, _ := StructToMap(ntf)
	g.SendToClient(conn, "newBalance", result)
}

func (g *AviatorGameContext) SendToAllClients(cmd string, data map[string]interface{}) {
	p := map[string]interface{}{
		"p": data,
		"c": cmd,
	}

	packet := BuildSFSMessage(13, 1, p)
	for _, player := range g.players {
		if player.IsOffline {
			continue
		}

		player.mutex.Lock()
		defer player.mutex.Unlock()
		player.conn.WriteMessage(websocket.BinaryMessage, packet)
	}
}

func (g *AviatorGameContext) SendToClient(conn *websocket.Conn, cmd string, data map[string]interface{}) {
	p := map[string]interface{}{
		"p": data,
		"c": cmd,
	}

	packet := BuildSFSMessage(13, 1, p)
	conn.WriteMessage(websocket.BinaryMessage, packet)
}

func (g *AviatorGameContext) UpdateStatus(newStatus int32, oldStatus int32) {
	println("UpdateStatus status=", newStatus)

	if newStatus == EAviatorStageReady {
		g.NewGameInit()
	} else {
		g.CurStage = newStatus
		g.curStateStartTime = time.Now().UnixMilli()
	}

	g.S2cChangeState(newStatus)
}

func (g *AviatorGameContext) OnTick() {
	now := time.Now().UnixMilli()
	println("onTick ", now, g.curStateStartTime, now-g.curStateStartTime, BET_TIME.Milliseconds(), "stage=", g.CurStage)

	switch g.CurStage {
	case EAviatorStageBet:
		{
			if now-g.curStateStartTime > BET_TIME.Milliseconds() {
				g.UpdateStatus(EAviatorStageCashOut, EAviatorStageBet)
			} else {
				g.S2cUpdateCurrentBets()
			}
		}
	case EAviatorStageCashOut:
		{
			if now-g.curStateStartTime > CASH_OUT_TIME.Milliseconds()+AWARD_TIME.Microseconds() {
				g.UpdateStatus(EAviatorStageReady, EAviatorStageCashOut)
			} else if now-g.curStateStartTime > CASH_OUT_TIME.Milliseconds() {
				if !g.IsAwarding {
					g.IsAwarding = true
					g.S2cRoundChartInfo()
				}
			} else {
				g.CurMultiplier = g.CurMultiplier + 0.1
				g.S2cUpdateCurrentCashOuts()
				g.S2cUpdateX()
			}
		}
	case EAviatorStageReady:
		{
			if now-g.curStateStartTime > READY_TIME.Milliseconds() {
				g.UpdateStatus(EAviatorStageBet, EAviatorStageReady)
			}
		}
	}
	g.S2cOnlinePlayers()
}

// StartTimer 启动定时器
func (g *AviatorGameContext) StartTimer(interval time.Duration, callback func()) {
	if g.IsRunning() {
		return // 已经启动，直接返回
	}
	g.isRunning = true
	g.Timer = time.NewTimer(interval)
	go func() {
		for range g.Timer.C {
			callback()
			g.Timer.Reset(interval)
		}
	}()
}

// StopTimer 停止定时器
func (g *AviatorGameContext) StopTimer() {
	if g.Timer != nil {
		g.Timer.Stop()
		g.isRunning = false
	}
}

// IsRunning 获取运行状态
func (g *AviatorGameContext) IsRunning() bool {
	return g.isRunning
}
