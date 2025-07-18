package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	EAviatorStageZero         = 0
	EAviatorStageBet          = 1
	EAviatorStageCashOut      = 2
	EAviatorStageCashOutAward = 3
)

const (
	BET_TIME      = 5 * time.Second
	CASH_OUT_TIME = 5 * time.Second
	AWARD_TIME    = 3 * time.Second
)

type PlayerBetSt struct {
	BetArea     int32
	BetValue    float64
	CashOut     float64
	autoCashOut float64
	hasCashOut  bool
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
	isRobot    bool    // 是否机器人

	BetList []*PlayerBetSt

	conn  *websocket.Conn
	mutex sync.Mutex
}

type AviatorGameContext struct {
	players map[string]*AviatorPlayerInfo
	robots  map[string]*AviatorPlayerInfo

	RecordId          int  // 牌局号(每次下一局累加1)
	isRunning         bool // 是否已启动
	Timer             *time.Timer
	curStateStartTime int64 //当前阶段开始时间
	CurStage          int32 //当前阶段
	CurMultiplier     float64

	CashOuts     []CashOut
	CurrentBets  []Bet
	LastBets     []Bet
	TotalCashOut float64
	TotalBet     float64

	startTime     int64
	endTime       int64
	LastRoundInfo RoundInfo
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
		robots:            make(map[string]*AviatorPlayerInfo, 0),
		curStateStartTime: 0,
		CurStage:          int32(EAviatorStageZero),
		CurMultiplier:     1.0,
		startTime:         0,
		endTime:           0,
	}
}
func (g *AviatorGameContext) Init() {
	g.StartTimer(500*time.Millisecond, g.OnTick)
}

func (g *AviatorGameContext) NewGameInit() {
	g.CurStage = EAviatorStageZero
	g.curStateStartTime = time.Now().UnixMilli()
	g.TotalBet = 0
	g.TotalCashOut = 0
	g.CurMultiplier = 1.0
	g.CashOuts = make([]CashOut, 0)
	g.CurrentBets = make([]Bet, 0)
	g.RecordId = g.RecordId + 1
}

func (g *AviatorGameContext) OnLogin(conn *websocket.Conn, obj map[string]interface{}) {
	playerInfo := &AviatorPlayerInfo{
		conn:      conn,
		Balance:   10000,
		BetList:   make([]*PlayerBetSt, 0),
		IsOffline: false,
		AccountId: "33687&&demo",
		Nickname:  "demo_71815",
		Currency:  "MAD",
	}
	g.players[conn.RemoteAddr().String()] = playerInfo
}
func (g *AviatorGameContext) OnRecv(conn *websocket.Conn, obj map[string]interface{}) {
	switch obj["c"] {
	case "cancelBetHandler":
		var result CancelBetRequest
		params, _ := obj["p"].(map[string]interface{})
		if err := MapToStruct(params, &result); err != nil {
			return
		}
		g.C2sCancelBet(conn, &result)
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
		g.C2sCurrentBetsInfo(conn)
	case "previousRoundInfoHandler":
		g.C2sPreviousRoundInfo(conn)
	case "getHugeWinsInfoHandler":
		var result HugeWinRequest
		params, _ := obj["p"].(map[string]interface{})
		if err := MapToStruct(params, &result); err != nil {
			return
		}
		g.C2sGetHugeWinsInfo(conn, &result)
	case "getTopRoundsInfoHandler":
		var result TopRoundRequest
		params, _ := obj["p"].(map[string]interface{})
		if err := MapToStruct(params, &result); err != nil {
			return
		}
		g.C2sGtTopRoundsInfo(conn, &result)
	case "getTopWinsInfoHandler":
		var result TopWinRequest
		params, _ := obj["p"].(map[string]interface{})
		if err := MapToStruct(params, &result); err != nil {
			return
		}
		g.C2sGetTopWinsInfo(conn, &result)
	default:
		fmt.Printf("⚠️ 未知扩展命令: %s\n", obj["c"])
	}
}

/*
func (g *AviatorGameContext) BetHistory(conn *websocket.Conn) {

}

func (g *AviatorGameContext) RoundFairness(conn *websocket.Conn) {

}

*/

func (g *AviatorGameContext) C2sCancelBet(conn *websocket.Conn, req *CancelBetRequest) {
	playerInfo := g.players[conn.RemoteAddr().String()]
	if playerInfo == nil {
		return
	}

	if g.CurStage != EAviatorStageBet {
		return
	}

	if req.BetID <= 0 || req.BetID > 2 {
		return
	}

	betSt := g.Id2Bet(int32(req.BetID), playerInfo)
	if betSt == nil {
		return
	}

	//扣钱
	playerInfo.Balance -= betSt.BetValue
	g.S2cNewBalance(playerInfo, playerInfo.Balance)
	if g.TotalBet-betSt.BetValue > 0 {
		g.TotalBet -= betSt.BetValue
	}

	g.CancelBet(int32(req.BetID), playerInfo)
	rsp := &CancelBetResponse{
		Code:     200,
		PlayerID: playerInfo.AccountId,
		BetID:    req.BetID,
	}

	result, _ := StructToMap(rsp)
	g.SendToClient(playerInfo, "cancelBet", result)
}

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
	g.S2cNewBalance(playerInfo, playerInfo.Balance)

	g.TotalBet += req.Bet
	playerInfo.BetList = append(playerInfo.BetList, &PlayerBetSt{
		BetArea:     int32(req.BetID),
		BetValue:    req.Bet,
		CashOut:     0,
		autoCashOut: req.AutoCashOut,
		hasCashOut:  false,
	})

	betResponse := &BetResponse{
		Code:         200,
		Bet:          req.Bet,
		BetID:        req.BetID,
		FreeBet:      req.FreeBet,
		PlayerID:     playerInfo.AccountId,
		ProfileImage: playerInfo.ProfileImage,
		Username:     playerInfo.Nickname,
	}

	g.CurrentBets = append(g.CurrentBets, Bet{
		Bet:          req.Bet,
		BetID:        req.BetID,
		IsFreeBet:    req.FreeBet,
		PlayerID:     playerInfo.AccountId,
		ProfileImage: playerInfo.ProfileImage,
		Username:     playerInfo.Nickname,
		Currency:     playerInfo.Currency,
		Payout:       0,
		WinAmount:    0,
		Win:          false,
		RoundBetId:   req.BetID,
	})
	if len(g.CurrentBets) > 50 {
		g.CurrentBets = g.CurrentBets[1:]
	}

	result, _ := StructToMap(betResponse)
	g.SendToClient(playerInfo, "bet", result)
}

func (g *AviatorGameContext) Id2Bet(betId int32, playerInfo *AviatorPlayerInfo) *PlayerBetSt {
	for idx, bet := range playerInfo.BetList {
		if bet.BetArea == betId {
			return playerInfo.BetList[idx]
		}
	}
	return nil
}

func (g *AviatorGameContext) CancelBet(betId int32, playerInfo *AviatorPlayerInfo) {
	for idx, bet := range playerInfo.BetList {
		if bet.BetArea == betId {
			playerInfo.BetList = append(playerInfo.BetList[:idx], playerInfo.BetList[idx+1:]...)
			break
		}
	}
	for idx, bet := range g.CurrentBets {
		if int32(bet.BetID) == betId && playerInfo.AccountId == bet.PlayerID {
			g.CurrentBets = append(g.CurrentBets[:idx], g.CurrentBets[idx+1:]...)
			break
		}
	}
}

func (g *AviatorGameContext) SetCashOut(betId int32, betValue float64, curMultiplier float64, playerInfo *AviatorPlayerInfo) {
	for idx, bet := range playerInfo.BetList {
		if bet.BetArea == betId {
			playerInfo.BetList[idx].hasCashOut = true
			playerInfo.BetList[idx].CashOut = betValue * curMultiplier
		}
	}

	for idx, bet := range g.CurrentBets {
		if int32(bet.BetID) == betId && playerInfo.AccountId == bet.PlayerID {
			g.CurrentBets[idx].Payout = g.CurMultiplier
			g.CurrentBets[idx].WinAmount = betValue * curMultiplier
			g.CurrentBets[idx].Win = true
			break
		}

	}
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

	if betSt.CashOut > 0 {
		return
	}
	CurMultiplier := g.CurMultiplier

	//加钱
	playerInfo.Balance += betSt.BetValue * CurMultiplier
	g.S2cNewBalance(playerInfo, playerInfo.Balance)

	g.SetCashOut(int32(req.BetID), betSt.BetValue, CurMultiplier, playerInfo)
	g.TotalCashOut += betSt.BetValue * CurMultiplier
	g.CashOuts = append(g.CashOuts, CashOut{
		BetID:      req.BetID,
		Multiplier: CurMultiplier,
		PlayerID:   playerInfo.AccountId,
		WinAmount:  betSt.BetValue * CurMultiplier,
	})

	cashOutResponse := CashOutResponse{
		Code:        200,
		Multiplier:  CurMultiplier,
		Cashouts:    make([]CashOutItem, 0),
		OperatorKey: "demo",
	}

	cashOutResponse.Cashouts = append(cashOutResponse.Cashouts, CashOutItem{
		BetAmount:           betSt.BetValue,
		BetID:               req.BetID,
		PlayerID:            playerInfo.AccountId,
		WinAmount:           betSt.BetValue * CurMultiplier,
		IsMaxWinAutoCashOut: false,
	})

	result, _ := StructToMap(cashOutResponse)
	g.SendToClient(playerInfo, "cashOut", result)

}

func (g *AviatorGameContext) C2sGetHugeWinsInfo(conn *websocket.Conn, req *HugeWinRequest) {
	playerInfo := g.players[conn.RemoteAddr().String()]
	if playerInfo == nil {
		return
	}
	topWinsResponse := TopWinsResponse{
		Code:    200,
		TopWins: []TopWin{},
	}

	topWinsResponse.TopWins = append(topWinsResponse.TopWins, TopWin{
		MaxMultiplier:           755.02,
		WinAmount:               7229509.53,
		EndDate:                 1752791583350,
		Payout:                  100,
		IsFreeBet:               false,
		ProfileImage:            "av-5.png",
		Bet:                     72295.09,
		RoundBetId:              2969720161,
		WinAmountInMainCurrency: 799938,
		Zone:                    "aviator_core_inst2_demo1",
		Currency:                "MAD",
		RoundId:                 8280205,
		PlayerId:                2053967,
		Username:                "demo_24529",
	})

	result, _ := StructToMap(topWinsResponse)
	g.SendToClient(playerInfo, "getHugeWinsInfo", result)
}

func (g *AviatorGameContext) C2sGetTopWinsInfo(conn *websocket.Conn, req *TopWinRequest) {
	playerInfo := g.players[conn.RemoteAddr().String()]
	if playerInfo == nil {
		return
	}
	topWinsResponse := TopWinsResponse{
		Code:    200,
		TopWins: []TopWin{},
	}

	topWinsResponse.TopWins = append(topWinsResponse.TopWins, TopWin{
		MaxMultiplier:           755.02,
		WinAmount:               7229509.53,
		EndDate:                 1752791583350,
		Payout:                  100,
		IsFreeBet:               false,
		ProfileImage:            "av-5.png",
		Bet:                     72295.09,
		RoundBetId:              2969720161,
		WinAmountInMainCurrency: 799938,
		Zone:                    "aviator_core_inst2_demo1",
		Currency:                "MAD",
		RoundId:                 8280205,
		PlayerId:                2053967,
		Username:                "demo_24529",
	})

	result, _ := StructToMap(topWinsResponse)
	g.SendToClient(playerInfo, "getTopWinsInfo", result)
}

func (g *AviatorGameContext) C2sGtTopRoundsInfo(conn *websocket.Conn, req *TopRoundRequest) {
	playerInfo := g.players[conn.RemoteAddr().String()]
	if playerInfo == nil {
		return
	}
	topRoundResponse := TopRoundsResponse{
		Code:      200,
		TopRounds: []TopRound{},
	}

	topRoundResponse.TopRounds = append(topRoundResponse.TopRounds, TopRound{
		RoundId:        1,
		RoundStartDate: 1638300000,
		EndDate:        1638300000 + 1000,
		MaxMultiplier:  1.5,
		ServerSeed:     "serverSeed",
		Zone:           "zone",
	})

	result, _ := StructToMap(topRoundResponse)
	g.SendToClient(playerInfo, "getTopRoundsInfo", result)
}

func (g *AviatorGameContext) C2sPreviousRoundInfo(conn *websocket.Conn) {
	playerInfo := g.players[conn.RemoteAddr().String()]
	if playerInfo == nil {
		return
	}

	previousRoundInfo := &PreviousRoundInfo{
		Bets: []Bet{},
		Code: 200,
		RoundInfo: RoundInfo{
			RoundId:        g.RecordId,
			Multiplier:     g.CurMultiplier,
			RoundStartDate: g.startTime,
			RoundEndDate:   g.endTime,
		},
	}

	previousRoundInfo.Bets = append(previousRoundInfo.Bets, g.LastBets...)
	result, _ := StructToMap(previousRoundInfo)
	g.SendToClient(playerInfo, "previousRoundInfoResponse", result)
}

func (g *AviatorGameContext) C2sCurrentBetsInfo(conn *websocket.Conn) {
	playerInfo := g.players[conn.RemoteAddr().String()]
	if playerInfo == nil {
		return
	}

	currentBetsInfo := &CurrentBetsInfo{
		BetsCount:              g.BetsCount(),
		OpenBetsCount:          g.OpenBetsCount(),
		Code:                   200,
		CashOuts:               []CashOut{},
		Bets:                   []Bet{},
		ActivePlayersCount:     g.OnlinePlayers(),
		TopPlayerProfileImages: []string{},
		TotalCashOut:           g.TotalCashOut,
	}

	currentBetsInfo.CashOuts = append(currentBetsInfo.CashOuts, g.CashOuts...)
	currentBetsInfo.Bets = append(currentBetsInfo.Bets, g.CurrentBets...)

	result, _ := StructToMap(currentBetsInfo)
	g.SendToClient(playerInfo, "currentBetsInfo", result)

}

func (g *AviatorGameContext) S2cUpdateCurrentCashOuts() {
	ntf := &UpdateCurrentCashOuts{
		Code:                   200,
		TotalCashOut:           g.TotalCashOut,
		OpenBetsCount:          int(g.TotalBet),
		ActivePlayersCount:     g.OnlinePlayers(),
		TopPlayerProfileImages: []string{},
		CashOuts:               []CashOut{},
	}

	ntf.CashOuts = append(ntf.CashOuts, g.CashOuts...)
	result, _ := StructToMap(ntf)
	g.SendToAllClients("updateCurrentCashOuts", result)
}

func (g *AviatorGameContext) S2cUpdateCurrentBets() {
	ntf := &UpdateCurrentBets{
		BetsCount:              g.BetsCount(),
		Code:                   200,
		ActivePlayersCount:     g.OnlinePlayers(),
		Bets:                   []Bet{},
		TopPlayerProfileImages: []string{},
	}
	ntf.Bets = append(ntf.Bets, g.CurrentBets...)

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

func (g *AviatorGameContext) S2cUpdateCrashX() {
	ntf := &UpdateCrashX{
		Code:   200,
		X:      g.CurMultiplier,
		CrashX: g.CurMultiplier,
	}
	result, _ := StructToMap(ntf)
	g.SendToAllClients("x", result)
}

func (g *AviatorGameContext) OnlinePlayers() int {
	onlinePlayers := 0
	for _, player := range g.players {
		if !player.IsOffline {
			onlinePlayers++
		}
	}

	return onlinePlayers + len(g.robots)
}

func (g *AviatorGameContext) OpenBetsCount() int {
	openBetsCount := 0
	for _, player := range g.players {
		for _, bet := range player.BetList {
			if bet.hasCashOut {
				continue
			}
			openBetsCount++
		}
	}
	return openBetsCount
}
func (g *AviatorGameContext) BetsCount() int {
	betsCount := 0
	for _, player := range g.players {
		if len(player.BetList) > 0 {
			betsCount += len(player.BetList)
		}
	}
	return betsCount
}

func (g *AviatorGameContext) S2cOnlinePlayers() {
	onlinePlayers := g.OnlinePlayers()
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
		TimeLeft:   5000,
	}

	if newStatus == EAviatorStageBet {
		ntf.ServerTime = time.Now().UnixMilli()
		ntf.BetStateEndTime = ntf.ServerTime + BET_TIME.Milliseconds()
	}
	result, _ := StructToMap(ntf)
	g.SendToAllClients("changeState", result)
}

func (g *AviatorGameContext) S2cNewBalance(player *AviatorPlayerInfo, balance float64) {
	ntf := &NewBalance{
		Code:       200,
		NewBalance: balance,
	}

	result, _ := StructToMap(ntf)
	g.SendToClient(player, "newBalance", result)
}

func (g *AviatorGameContext) SendToAllClients(cmd string, data map[string]interface{}) {
	p := map[string]interface{}{
		"p": data,
		"c": cmd,
	}

	packet := BuildSFSMessage(13, 1, p)
	for _, player := range g.players {
		if player.IsOffline || player.isRobot {
			continue
		}

		player.mutex.Lock()
		defer player.mutex.Unlock()
		player.conn.WriteMessage(websocket.BinaryMessage, packet)
	}
}

func (g *AviatorGameContext) SendToClient(player *AviatorPlayerInfo, cmd string, data map[string]interface{}) {
	p := map[string]interface{}{
		"p": data,
		"c": cmd,
	}

	packet := BuildSFSMessage(13, 1, p)

	player.mutex.Lock()
	defer player.mutex.Unlock()
	player.conn.WriteMessage(websocket.BinaryMessage, packet)

	println("SendToClient=", cmd, data)
}

func (g *AviatorGameContext) UpdateStatus(newStatus int32) {
	println("UpdateStatus status=", newStatus)

	g.CurStage = newStatus
	g.curStateStartTime = time.Now().UnixMilli()

	// 服务端虚拟状态不用通知
	if newStatus == EAviatorStageCashOutAward {
		return
	} else {
		g.S2cChangeState(g.CurStage)
	}
}

func (g *AviatorGameContext) GenOdds(interval int64) float64 {
	// 游戏阶段：更新倍数等

	// 将tick转换为实际秒数 (tick * 0.1)
	seconds := float64(interval/200) * 0.2
	// 使用新公式: y = 0.9084 * exp(0.0752 * x)
	odds := 0.99 * math.Exp(0.0752*seconds)
	return odds

}

func (g *AviatorGameContext) OnTick() {
	now := time.Now().UnixMilli()
	interval := now - g.curStateStartTime

	println("onTick ", now, interval, "stage=", g.CurStage)
	switch g.CurStage {
	case EAviatorStageZero:
		g.UpdateStatus(EAviatorStageBet)
	case EAviatorStageBet:
		{
			if interval > BET_TIME.Milliseconds() {
				g.UpdateStatus(EAviatorStageCashOut)
			} else {
				g.AutoRobotBet()
				g.S2cUpdateCurrentBets()
			}
		}
	case EAviatorStageCashOut:
		{
			if interval > CASH_OUT_TIME.Milliseconds() {
				g.DoSettle()
			} else {
				oldCurMultiplier := g.CurMultiplier
				g.CurMultiplier = g.GenOdds(interval)
				if g.CurMultiplier < 1.01 {
					g.CurMultiplier = 1.01
				}

				//判断系统会不会输
				sysWin := g.CacSysWin()
				if sysWin < 0 {
					g.CurMultiplier = oldCurMultiplier
					g.DoSettle()
				} else {
					g.AutoRobotCashOut()
					g.AutoCashOut()
					g.S2cUpdateCurrentCashOuts()
					g.S2cUpdateX()
				}
			}
		}
	case EAviatorStageCashOutAward:
		{
			if interval > CASH_OUT_TIME.Milliseconds() {
				g.DoStart()
				g.UpdateStatus(EAviatorStageBet)
			}
		}
	}
	g.S2cOnlinePlayers()
}

func (g *AviatorGameContext) DoSettle() {

	println("DoSettle")
	g.S2cUpdateCrashX()
	g.S2cRoundChartInfo()

	//清空下注
	for _, player := range g.players {
		player.BetList = []*PlayerBetSt{}
	}
	g.robots = map[string]*AviatorPlayerInfo{}
	g.UpdateStatus(EAviatorStageCashOutAward)
	g.endTime = time.Now().UnixMilli()

	g.LastRoundInfo = RoundInfo{
		RoundId:        g.RecordId,
		RoundStartDate: g.startTime,
		RoundEndDate:   g.endTime,
		Multiplier:     g.CurMultiplier,
	}
}

func (g *AviatorGameContext) DoStart() {
	g.LastBets = make([]Bet, len(g.CurrentBets))
	copy(g.LastBets, g.CurrentBets)
	g.NewGameInit()
	g.startTime = time.Now().UnixMilli()
}

func (g *AviatorGameContext) AutoRobotBet() {
	robotCount := rand.Intn(2) + 1

	for i := 0; i < robotCount; i++ {
		robot := g.CreateRobot()
		betValue := rand.Float64() * 100
		betId := rand.Intn(2) + 1

		g.TotalBet += betValue
		robot.BetList = append(robot.BetList, &PlayerBetSt{
			BetArea:     int32(betId),
			BetValue:    betValue,
			CashOut:     0,
			hasCashOut:  false,
			autoCashOut: (rand.Float64()*float64(rand.Intn(2)) + 1),
		})

		g.CurrentBets = append(g.CurrentBets, Bet{
			Bet:          betValue,
			BetID:        betId,
			IsFreeBet:    false,
			PlayerID:     robot.AccountId,
			ProfileImage: robot.ProfileImage,
			Username:     robot.Nickname,
			Payout:       0,
			WinAmount:    0,
			Win:          false,
			RoundBetId:   betId,
		})
		if len(g.CurrentBets) > 50 {
			g.CurrentBets = g.CurrentBets[1:]
		}
		g.robots[robot.AccountId] = robot
	}
}

func (g *AviatorGameContext) CreateRobot() *AviatorPlayerInfo {

	rand.Seed(time.Now().UnixNano()) // 初始化随机种子

	// 生成 100000-999999 之间的随机数
	randomNum1 := rand.Intn(900000) + 100000
	randomNum2 := rand.Intn(900000) + 100000

	playerInfo := &AviatorPlayerInfo{
		Balance:   0,
		BetList:   make([]*PlayerBetSt, 0),
		IsOffline: false,
		AccountId: fmt.Sprint(randomNum1) + "&&demo",
		Nickname:  "demo" + fmt.Sprint(randomNum2),
	}
	return playerInfo
}

func (g *AviatorGameContext) AutoRobotCashOut() {
	for _, player := range g.robots {
		for idx := range player.BetList {
			bet := player.BetList[idx]
			if bet.hasCashOut {
				continue
			}
			if bet.autoCashOut >= g.CurMultiplier {
				continue
			}
			g.SetCashOut(int32(bet.BetArea), bet.BetValue, g.CurMultiplier, player)

			g.TotalCashOut += bet.BetValue * g.CurMultiplier
			g.CashOuts = append(g.CashOuts, CashOut{
				BetID:      int(bet.BetArea),
				Multiplier: g.CurMultiplier,
				PlayerID:   player.AccountId,
				WinAmount:  bet.BetValue * g.CurMultiplier,
			})
		}
	}
}

func (g *AviatorGameContext) AutoCashOut() {
	for _, player := range g.players {
		for idx := range player.BetList {
			bet := player.BetList[idx]
			if bet.hasCashOut {
				continue
			}
			if bet.autoCashOut > 1 {
				player.Balance += bet.BetValue * g.CurMultiplier
				g.SetCashOut(int32(bet.BetArea), bet.BetValue, g.CurMultiplier, player)
				g.TotalCashOut += bet.BetValue * g.CurMultiplier
				g.CashOuts = append(g.CashOuts, CashOut{
					BetID:      int(bet.BetArea),
					Multiplier: g.CurMultiplier,
					PlayerID:   player.AccountId,
					WinAmount:  bet.BetValue * g.CurMultiplier,
				})
				if player.IsOffline {
					continue
				}
				g.S2cNewBalance(player, player.Balance)
			}
		}
	}
}

func (g *AviatorGameContext) CacSysWin() float64 {
	totalBet := 0.0
	totalCashOut := 0.0
	for _, player := range g.players {
		for _, bet := range player.BetList {
			totalBet += bet.BetValue
			if bet.hasCashOut {
				totalCashOut += bet.CashOut
			}
		}
	}
	return totalBet - totalCashOut
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
