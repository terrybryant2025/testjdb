package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var GameContextMap = make(map[string]*RicherGameContext)

var seq int64

type RicherGameContext struct {
	//gameService *SpiritService

	// 房间基础信息
	ClientID         string // 用户conn clientID
	PlayerInfo       *RicherBettingInfo
	TableId          int         // 座号
	mu               sync.Mutex  // 房间锁，防并发
	RoomTimer        *time.Timer // 操作定时器
	RoomTimeoutTimes int         // 房间操作超时次数

	// 一个房间下的统计数据
	RecordId int // 牌局号(每次下一局累加1)

	// 一次下注的游戏更新数据,输或cash out后 置空
	BetChip      int     // 带入金额（分）
	BetMulti     int     // 带入倍数
	AllGold      int     // 订单盈亏最终金额
	OrderId      string  // 订单ID
	Intervention bool    // 是否干预
	MaxWinProfit float64 // 当局最多可赢
	MaxWinRatio  float64 // 可翻出的最大赔率(<=0 一定要输， >0算出只能赔多少)

	CurrentStatus int // 游戏状态 0游戏待开始   10游戏开始
	GameStateId   int // 游戏状态id
}

func NewRicherGameContext(clientID string, tableId int) *RicherGameContext {
	//playerInfo, ok := RicherUserData.Get(clientID)
	//if !ok {
	//	return nil
	//}

	gameContext := &RicherGameContext{
		ClientID: clientID,
		//PlayerInfo: playerInfo,
		TableId: tableId,
	}
	// 初始化每局游戏的状态字段
	//gameContext.nextGame(context.Background())

	GameContextMap[clientID] = gameContext
	return gameContext
}

func GetSRicherGameContext(clientID string) *RicherGameContext {
	if v, ok := GameContextMap[clientID]; ok {
		return v
	}
	return nil
}

func (g *RicherGameContext) Spin(c context.Context, denom string, extraBetType string, gameStateId string, playerBet int, playerBetMulti int) (*SpinResultWrapper, error) {
	g.BetChip = playerBet
	g.BetMulti = playerBetMulti

	// 获取
	idx := rand.Intn(5000) + 1 // 0 ~ 4999
	idx = 459
	idx = idx - 1 // slice从0开始
	game := GameList[idx]
	seq++
	fmt.Printf("现在命中的是%d", idx+1)

	// 计算金额
	g.AllGold = g.BetMulti * game.Winnings // 下注倍数 * 游戏记录盈利 = 当局盈利

	// 订单成功，添加返回值
	spinResultWrapper := SpinResultWrapper{
		SpinResult: SpinResult{
			GameStateCount:  0,
			GameStateResult: make([]GameState, 0),
			TotalWin:        game.Winnings,
			BoardDisplayResult: BoardDisplay{
				WinRankType: "Nothing",
				ScoreType:   "Nothing",
				DisplayBet:  g.BetChip, // 下注金额，单位分
			},
			GameFlowResult: g.setGameFlowResult(c, game),
		},
		TS:      time.Now().UnixMilli(),
		Balance: 10000, // 用户余额
		GameSeq: seq,
	}

	state := 0
	spinResultWrapper.SpinResult.GameStateResult = append(spinResultWrapper.SpinResult.GameStateResult, *g.SetGameStateId0(c))
	state++
	spinResultWrapper.SpinResult.GameStateResult = append(spinResultWrapper.SpinResult.GameStateResult, *g.SetGameStatId1(c, game)) // 记录第一次滚轮返回内容
	if game.Card.FreeGameNum > 0 {
		spinResultWrapper.SpinResult.GameStateResult = append(spinResultWrapper.SpinResult.GameStateResult, *g.SetGameStatId2(c, game)) // 记录免费游戏返回内容
	}
	spinResultWrapper.SpinResult.GameStateResult = append(spinResultWrapper.SpinResult.GameStateResult, *g.SetGameStatId3(c, len(spinResultWrapper.SpinResult.GameStateResult)+1))
	spinResultWrapper.SpinResult.GameStateCount = len(spinResultWrapper.SpinResult.GameStateResult)

	return &spinResultWrapper, nil
}

func (g *RicherGameContext) SendSpinResult(card *Card) {

}

func (g *RicherGameContext) SetGameStateId0(c context.Context) *GameState {
	return &GameState{
		GameStateId:   StartGameState,
		CurrentState:  1,
		GameStateType: StateMap[StartGameState].GameStateType,
		RoundCount:    0,
		RoundResult:   nil,
		StateWin:      0,
	}
}

// SetGameStatId1 第一次滚动的结果
func (g *RicherGameContext) SetGameStatId1(c context.Context, res *RicherSpinResult) *GameState {
	state := &GameState{
		GameStateId:   BetState,
		CurrentState:  2,
		GameStateType: StateMap[BetState].GameStateType,
		RoundCount:    0,
		StateWin:      res.Card.BaseScore * g.BetMulti, // 这个状态下赢取的金额
	}

	state.RoundResult = g.SetRoundResult(c, res, false)
	state.RoundCount = len(state.RoundResult)
	return state
}

// SetGameStatId2 免费游戏滚动的结果
func (g *RicherGameContext) SetGameStatId2(c context.Context, res *RicherSpinResult) *GameState {
	state := &GameState{
		GameStateId:   FreeGameState,
		CurrentState:  3,
		GameStateType: StateMap[FreeGameState].GameStateType,
		RoundCount:    0,
		StateWin:      res.Card.FreeGameTotalScore * g.BetMulti,
	}

	state.RoundResult = g.SetRoundResult(c, res, true)
	state.RoundCount = len(state.RoundResult)
	return state
}

func (g *RicherGameContext) SetGameStatId3(c context.Context, currentState int) *GameState {
	state := &GameState{
		GameStateId:   EndState,
		CurrentState:  currentState,
		GameStateType: StateMap[EndState].GameStateType,
		RoundCount:    0,
		StateWin:      0,
	}
	return state
}

func (g *RicherGameContext) SetRoundResult(c context.Context, res *RicherSpinResult, isFreeGame bool) []RoundResult {
	gameCard := make([]*Card, 0)
	var totalRound int
	if isFreeGame {
		gameCard = res.Card.FreeGameCard
		totalRound = res.Card.FreeGameNum
	} else {
		gameCard = append(gameCard, res.Card)
		totalRound = 1
	}

	data := make([]RoundResult, 0, len(gameCard))
	for idx, card := range gameCard {
		var roundRes = RoundResult{
			RoundWin: card.BaseScore * g.BetMulti,
			ScreenResult: ScreenResult{
				TableIndex:   0,
				ScreenSymbol: convertCardValMatrix(res.Card.Rolls),
				DampInfo:     g.setDampInfo(c),
			},
			ProgressResult: ProgressResult{
				MaxTriggerFlag: false,
				StepInfo: StepInfo{
					CurrentStep: 1,
					AddStep:     0,
					TotalStep:   1,
				},
				StageInfo: StageInfo{
					CurrentStage: 1,
					TotalStage:   1,
					AddStage:     0,
				},
				RoundInfo: RoundInfo{
					CurrentRound: idx + 1,
					TotalRound:   totalRound,
					AddRound:     0,
				},
			},
			DisplayResult: DisplayResult{
				AccumulateWinResult: AccumulateWinResult{
					BeforeSpinFirstStateOnlyBasePayAccWin: 0,
					AfterSpinFirstStateOnlyBasePayAccWin:  0,
					BeforeSpinAccWin:                      0,
					AfterSpinAccWin:                       0,
				},
				ReadyHandResult: ReadyHandResult{
					DisplayMethod: [][]bool{
						{false},
						{false},
						{false},
						{false},
						{false},
					},
				},
				BoardDisplayResult: BoardDisplay{
					WinRankType: "Nothing",
					ScoreType:   "",
					DisplayBet:  0, // 好像一直都是0
				},
			},
			GameResult: GameResult{
				PlayerWin:    card.BaseScore * g.BetMulti, // 一次滚动 赢取的金额
				WayWinResult: g.setWayWinResult(c, card),
				GameWinType:  "WayGame",
			},
		}

		if isFreeGame {
			roundRes.ExtendGameState = &ExtendGameState{
				ScreenScatterTwoPositionList: nil,
				ScreenMultiplier:             nil,
				RoundMultiplier:              0,
				ScreenWinsInfo:               nil,
				ExtendWin:                    0,
				GameDescriptor: GameDescriptor{
					Version: 1,
					Component: [][]TypVal{
						{
							TypVal{
								Type:  "label",
								Value: "odds",
							},
							TypVal{
								Type:  "label",
								Value: "colon",
							},
							TypVal{
								Type:  "text",
								Value: fmt.Sprintf("%d", idx+1),
							},
						},
					},
				},
			}
		} else if card.FreeGameNum > 0 {
			roundRes.SpecialFeatureResult = &SpecialFeatureResult{
				SpecialHitPattern:    "HP_88",
				TriggerEvent:         "Trigger_01",
				SpecialScreenHitData: card.CardScreenHit[Treasure],
				SpecialScreenWin:     0,
			}
			roundRes.DisplayResult.ReadyHandResult.DisplayMethod[5][0] = true // 不知道指的什么，看的返回值是这个
		}

		data = append(data, roundRes)
	}

	return data
}

func (g *RicherGameContext) setWayWinResult(c context.Context, card *Card) []*WayWinData {
	if card.BaseScore <= 0 {
		return make([]*WayWinData, 0)
	}

	data := make([]*WayWinData, 0, len(card.CardSerialRollCount))

	for cardVal, rollCount := range card.CardSerialRollCount {
		if rollCount >= 3 { // 连续大于3轮才能获得积分
			if cardVal == Treasure && rollCount != 5 {
				continue
			}

			combTotal := 1 // 多少种组合
			for _, count := range card.CardCounts[cardVal] {
				if count > 0 {
					combTotal *= count
				}
			}

			wwr := WayWinData{
				SymbolId:      int(cardVal),
				HitDirection:  "LeftToRight",
				HitNumber:     rollCount,
				HitCount:      combTotal,
				HitOdds:       oddsMap[cardVal][rollCount],
				SymbolWin:     oddsMap[cardVal][rollCount] * combTotal * g.BetMulti,
				ScreenHitData: card.CardScreenHit[cardVal],
			}
			data = append(data, &wwr)
		}
	}
	return data
}

func (g *RicherGameContext) setGameFlowResult(c context.Context, game *RicherSpinResult) GameFlowResult {
	res := GameFlowResult{
		IsBoardEndFlag:       true,
		CurrentSystemStateId: 3,
	}

	res.SystemStateIdOptions = []int{0}
	if game.Winnings > 0 {
		res.SystemStateIdOptions = append(res.SystemStateIdOptions, 997)
	}

	return res
}

func (g *RicherGameContext) setDampInfo(c context.Context) [][]int {
	data := [][]int{
		{7, 4},
		{7, 7},
		{5, 4},
		{8, 2},
		{8, 7},
	}

	return data
}

func convertCardValMatrix(val [5][3]CardVal) [][]int {
	result := make([][]int, len(val))
	for i := range val {
		result[i] = make([]int, len(val[i]))
		for j := range val[i] {
			result[i][j] = int(val[i][j])
		}
	}
	return result
}
