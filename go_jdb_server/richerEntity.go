package main

//
//// 最外层结构
//type SpinResultResponse struct {
//	SpinResult SpinResult `json:"spinResult"`
//}
//
//// 主要的spin结果结构
//type SpinResult struct {
//	GameStateCount     int                `json:"gameStateCount"`
//	GameStateResult    []GameStateResult  `json:"gameStateResult"`
//	TotalWin           int                `json:"totalWin"`
//	BoardDisplayResult BoardDisplayResult `json:"boardDisplayResult"`
//	GameFlowResult     GameFlowResult     `json:"gameFlowResult"`
//	Ts                 int64              `json:"ts"`
//	Balance            float64            `json:"balance"`
//	GameSeq            int64              `json:"gameSeq"`
//}
//
//// 每个游戏状态结果
//type GameStateResult struct {
//	GameStateId   int           `json:"gameStateId"`
//	CurrentState  int           `json:"currentState"`
//	GameStateType string        `json:"gameStateType"`
//	RoundCount    int           `json:"roundCount"`
//	StateWin      int           `json:"stateWin"`
//	RoundResult   []RoundResult `json:"roundResult,omitempty"`
//}
//
//// 每轮的结果
//type RoundResult struct {
//	RoundWin       int            `json:"roundWin"`
//	ScreenResult   ScreenResult   `json:"screenResult"`
//	DampInfo       [][]int        `json:"dampInfo"`
//	ProgressResult ProgressResult `json:"progressResult"`
//	DisplayResult  DisplayResult  `json:"displayResult"`
//	GameResult     GameResult     `json:"gameResult"`
//}
//
//// 输赢屏幕表现
//type ScreenResult struct {
//	TableIndex   int     `json:"tableIndex"`
//	ScreenSymbol [][]int `json:"screenSymbol"`
//}
//
//// 进度结果
//type ProgressResult struct {
//	MaxTriggerFlag bool      `json:"maxTriggerFlag"`
//	StepInfo       StepInfo  `json:"stepInfo"`
//	StageInfo      StageInfo `json:"stageInfo"`
//	RoundInfo      RoundInfo `json:"roundInfo"`
//}
//
//type StepInfo struct {
//	CurrentStep int `json:"currentStep"`
//	AddStep     int `json:"addStep"`
//	TotalStep   int `json:"totalStep"`
//}
//
//type StageInfo struct {
//	CurrentStage int `json:"currentStage"`
//	TotalStage   int `json:"totalStage"`
//	AddStage     int `json:"addStage"`
//}
//
//type RoundInfo struct {
//	CurrentRound int `json:"currentRound"`
//	TotalRound   int `json:"totalRound"`
//	AddRound     int `json:"addRound"`
//}
//
//// 显示结果
//type DisplayResult struct {
//	AccumulateWinResult AccumulateWinResult `json:"accumulateWinResult"`
//	ReadyHandResult     ReadyHandResult     `json:"readyHandResult"`
//	BoardDisplayResult  BoardDisplayResult  `json:"boardDisplayResult"`
//}
//
//type AccumulateWinResult struct {
//	BeforeSpinFirstStateOnlyBasePayAccWin int `json:"beforeSpinFirstStateOnlyBasePayAccWin"`
//	AfterSpinFirstStateOnlyBasePayAccWin  int `json:"afterSpinFirstStateOnlyBasePayAccWin"`
//	BeforeSpinAccWin                      int `json:"beforeSpinAccWin"`
//	AfterSpinAccWin                       int `json:"afterSpinAccWin"`
//}
//
//type ReadyHandResult struct {
//	DisplayMethod [][]bool `json:"displayMethod"`
//}
//
//type BoardDisplayResult struct {
//	WinRankType string `json:"winRankType"`
//	DisplayBet  int    `json:"displayBet"`
//}
//
//// 游戏结果
//type GameResult struct {
//	PlayerWin    int            `json:"playerWin"`
//	WayWinData []WayWinData `json:"wayWinResult"`
//	GameWinType  string         `json:"gameWinType"`
//}
//
//type WayWinData struct {
//	SymbolId      int      `json:"symbolId"`
//	HitDirection  string   `json:"hitDirection"`
//	HitNumber     int      `json:"hitNumber"`
//	HitCount      int      `json:"hitCount"`
//	HitOdds       int      `json:"hitOdds"`
//	SymbolWin     int      `json:"symbolWin"`
//	ScreenHitData [][]bool `json:"screenHitData"`
//}
//
//// 其他的结果结构体
//type BoardDisplayResult struct {
//	WinRankType string `json:"winRankType"`
//	ScoreType   string `json:"scoreType"`
//	DisplayBet  int    `json:"displayBet"`
//}
//
//type GameFlowResult struct {
//	IsBoardEndFlag       bool  `json:"IsBoardEndFlag"`
//	CurrentSystemStateId int   `json:"currentSystemStateId"`
//	SystemStateIdOptions []int `json:"systemStateIdOptions"`
//}
