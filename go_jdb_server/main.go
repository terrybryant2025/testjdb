package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ---------------------------------spin-------------------------------------
type SpinResultWrapper struct {
	SpinResult SpinResult `json:"spinResult"` // 游戏结果
	TS         int64      `json:"ts"`         // 时间戳
	Balance    float64    `json:"balance"`    // 余额
	GameSeq    int64      `json:"gameSeq"`    // 游戏序号
}

type SpinResult struct {
	GameStateCount     int            `json:"gameStateCount"`     // 游戏状态总数
	GameStateResult    []GameState    `json:"gameStateResult"`    // 游戏状态结果列表
	TotalWin           int            `json:"totalWin"`           // 总赢分
	BoardDisplayResult BoardDisplay   `json:"boardDisplayResult"` // 面板显示结果
	GameFlowResult     GameFlowResult `json:"gameFlowResult"`     // 游戏流程结果
}

type GameState struct {
	GameStateId   int           `json:"gameStateId"`           // 游戏状态ID
	CurrentState  int           `json:"currentState"`          // 当前状态
	GameStateType string        `json:"gameStateType"`         // 游戏状态类型
	RoundCount    int           `json:"roundCount"`            // 回合数
	RoundResult   []RoundResult `json:"roundResult,omitempty"` // 回合结果列表
	StateWin      int           `json:"stateWin"`              // 状态赢分
}
type SpecialFeatureResult struct {
	SpecialHitPattern    string   `json:"specialHitPattern,omitempty"`
	TriggerEvent         string   `json:"triggerEvent,omitempty"`
	SpecialScreenHitData [][]bool `json:"specialScreenHitData,omitempty"`
	SpecialScreenWin     int      `json:"specialScreenWin"`
}
type RoundResult struct {
	RoundWin              int                    `json:"roundWin"`                       // 回合赢分
	ScreenResult          ScreenResult           `json:"screenResult"`                   // 屏幕结果
	ExtendGameStateResult ExtendGameStateResult  `json:"extendGameStateResult"`          // 扩展游戏状态
	ProgressResult        ProgressResult         `json:"progressResult"`                 // 进度结果
	DisplayResult         DisplayResult          `json:"displayResult"`                  // 显示结果
	GameResult            GameResult             `json:"gameResult"`                     // 游戏结果
	SpecialFeatureResult  []SpecialFeatureResult `json:"specialFeatureResult,omitempty"` //特殊模式
}

type ScreenResult struct {
	TableIndex   int     `json:"tableIndex"`   // 表格索引
	ScreenSymbol [][]int `json:"screenSymbol"` // 屏幕符号矩阵
	DampInfo     [][]int `json:"dampInfo"`     // 衰减信息
}

type ExtendGameState struct {
}
type ExtendGameStateResult struct {
	ScreenScatterTwoPositionList [][][]int       `json:"screenScatterTwoPositionList,omitempty"` // 散布符号2位置列表
	ScreenMultiplier             []interface{}   `json:"screenMultiplier,omitempty"`             // 屏幕倍数
	RoundMultiplier              int             `json:"roundMultiplier,omitempty"`              // 回合倍数
	ScreenWinsInfo               []ScreenWinInfo `json:"screenWinsInfo,omitempty"`               // 屏幕获胜信息
	GameDescriptor               GameDescriptor  `json:"gameDescriptor,omitempty"`
	// ✅ 补充字段：根据 JSON 中存在的字段添加
	ReSpinFlag       bool    `json:"reSpinFlag,omitempty"`   // 是否触发再旋转
	ReSpinTimes      int     `json:"reSpinTimes,omitempty"`  // 再旋转次数
	ColumnRecord     int     `json:"columnRecord,omitempty"` // 列记录标记
	SquintFlag       bool    `json:"squintFlag,omitempty"`   // 是否为斜视状态
	ExtendWin        int     `json:"extendWin,omitempty"`    // 扩展赢分
	ScreenSymbol     [][]int `json:"screenSymbol,omitempty"` // 屏幕符号矩阵
	DampInfo         [][]int `json:"dampInfo,omitempty"`     // 衰减信息
	TriggerMusicFlag bool    `json:"triggerMusicFlag,omitempty"`
}
type ScreenWinInfo struct {
	PlayerWin         int           `json:"playerWin"`         // 玩家赢分
	QuantityWinResult []interface{} `json:"quantityWinResult"` // 数量获胜结果
	GameWinType       string        `json:"gameWinType"`       // 游戏获胜类型
}

type GameDescriptor struct {
	Version          int               `json:"version"`          // 版本号
	CascadeComponent [][]interface{}   `json:"cascadeComponent"` // 级联组件
	Component        [][]ComponentItem `json:"component"`
}
type Placeholder struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type ComponentItem struct {
	Type         string        `json:"type"`
	Value        string        `json:"value"`
	Placeholders []Placeholder `json:"placeholders"`
}

type ProgressResult struct {
	MaxTriggerFlag bool      `json:"maxTriggerFlag"` // 最大触发标志
	StepInfo       StepInfo  `json:"stepInfo"`       // 步骤信息
	StageInfo      StageInfo `json:"stageInfo"`      // 阶段信息
	RoundInfo      RoundInfo `json:"roundInfo"`      // 回合信息
}

type StepInfo struct {
	CurrentStep int `json:"currentStep"` // 当前步骤
	AddStep     int `json:"addStep"`     // 增加步骤
	TotalStep   int `json:"totalStep"`   // 总步骤
}

type StageInfo struct {
	CurrentStage int `json:"currentStage"` // 当前阶段
	TotalStage   int `json:"totalStage"`   // 总阶段
	AddStage     int `json:"addStage"`     // 增加阶段
}

type RoundInfo struct {
	CurrentRound int `json:"currentRound"` // 当前回合
	TotalRound   int `json:"totalRound"`   // 总回合
	AddRound     int `json:"addRound"`     // 增加回合
}

type DisplayResult struct {
	AccumulateWinResult AccumulateWinResult `json:"accumulateWinResult"` // 累积获胜结果
	ReadyHandResult     ReadyHandResult     `json:"readyHandResult"`     // 准备手牌结果
	BoardDisplayResult  BoardDisplay        `json:"boardDisplayResult"`  // 面板显示结果
}

type AccumulateWinResult struct {
	BeforeSpinFirstStateOnlyBasePayAccWin int `json:"beforeSpinFirstStateOnlyBasePayAccWin"` // 旋转前首状态仅基本支付累积赢分
	AfterSpinFirstStateOnlyBasePayAccWin  int `json:"afterSpinFirstStateOnlyBasePayAccWin"`  // 旋转后首状态仅基本支付累积赢分
	BeforeSpinAccWin                      int `json:"beforeSpinAccWin"`                      // 旋转前累积赢分
	AfterSpinAccWin                       int `json:"afterSpinAccWin"`                       // 旋转后累积赢分
}

type ReadyHandResult struct {
	DisplayMethod [][]bool `json:"displayMethod"` // 显示方法
}

type BoardDisplay struct {
	WinRankType string `json:"winRankType"`         // 获胜等级类型
	ScoreType   string `json:"scoreType,omitempty"` // 分数类型
	DisplayBet  int    `json:"displayBet"`          // 显示投注
}

type GameResult struct {
	PlayerWin int `json:"playerWin"` // 玩家赢分
	// QuantityGameResult     QuantityGameResult `json:"quantityGameResult,omitempty"`     // 数量游戏结果
	CascadeEliminateResult []interface{}   `json:"cascadeEliminateResult,omitempty"` // 级联消除结果
	GameWinType            string          `json:"gameWinType,omitempty"`            // 游戏获胜类型
	LineWinResult          []LineWinResult `json:"lineWinResult"`                    // 线型中奖明细
}

type QuantityGameResult struct {
	PlayerWin         int           `json:"playerWin,omitempty"`         // 玩家赢分
	QuantityWinResult []interface{} `json:"quantityWinResult,omitempty"` // 数量获胜结果
	GameWinType       string        `json:"gameWinType,omitempty"`       // 游戏获胜类型
}

type GameFlowResult struct {
	IsBoardEndFlag       bool  `json:"IsBoardEndFlag"`       // 面板结束标志
	CurrentSystemStateId int   `json:"currentSystemStateId"` // 当前系统状态ID
	SystemStateIdOptions []int `json:"systemStateIdOptions"` // 系统状态ID选项
}

//---------------------------------spin end-------------------------------------

// ---------------------------------super init-----------------------------------
type GameSettingResponse struct {
	MaxBet                  int64          `json:"maxBet"`
	MinBet                  int64          `json:"minBet"`
	DefaultLineBetIdx       int            `json:"defaultLineBetIdx"`
	DefaultBetLineIdx       int            `json:"defaultBetLineIdx"`
	DefaultWaysBetIdx       int            `json:"defaultWaysBetIdx"`
	DefaultWaysBetColumnIdx int            `json:"defaultWaysBetColumnIdx"`
	DefaultConnectBetIdx    int            `json:"defaultConnectBetIdx"`
	DefaultQuantityBetIdx   int            `json:"defaultQuantityBetIdx"`
	BetCombinations         map[string]int `json:"betCombinations"`
	SingleBetCombinations   map[string]int `json:"singleBetCombinations"`
	GambleLimit             int            `json:"gambleLimit"`
	GambleTimes             int            `json:"gambleTimes"`
	GameFeatureCount        int            `json:"gameFeatureCount"`
	ExecuteSetting          ExecuteSetting `json:"executeSetting"`
	Denoms                  []int          `json:"denoms"`
	DefaultDenomIdx         int            `json:"defaultDenomIdx"`
	BuyFeature              bool           `json:"buyFeature"`
	BuyFeatureLimit         int            `json:"buyFeatureLimit"`
}
type ExecuteSetting struct {
	SettingId           string              `json:"settingId"`
	BetSpecSetting      BetSpecSetting      `json:"betSpecSetting"`
	GameStateSetting    []GameStateSetting  `json:"gameStateSetting"`
	DoubleGameSetting   DoubleGameSetting   `json:"doubleGameSetting"`
	BoardDisplaySetting BoardDisplaySetting `json:"boardDisplaySetting"`
	GameFlowSetting     GameFlowSetting     `json:"gameFlowSetting"`
}
type BetSpecSetting struct {
	PaymentType      string           `json:"paymentType"`
	ExtraBetTypeList []string         `json:"extraBetTypeList"`
	BetSpecification BetSpecification `json:"betSpecification"`
}

type BetSpecification struct {
	LineBetList []int  `json:"lineBetList"`
	BetLineList []int  `json:"betLineList"`
	BetType     string `json:"betType"`
}
type GameStateSetting struct {
	GameStateType         string                `json:"gameStateType"`
	FrameSetting          FrameSetting          `json:"frameSetting"`
	TableSetting          TableSetting          `json:"tableSetting"`
	SymbolSetting         SymbolSetting         `json:"symbolSetting"`
	LineSetting           LineSetting           `json:"lineSetting"`
	GameHitPatternSetting GameHitPatternSetting `json:"gameHitPatternSetting"`
	SpecialFeatureSetting SpecialFeatureSetting `json:"specialFeatureSetting"`
	ProgressSetting       ProgressSetting       `json:"progressSetting"`
	DisplaySetting        DisplaySetting        `json:"displaySetting"`
	ExtendSetting         ExtendSetting         `json:"extendSetting"`
}
type FrameSetting struct {
	ScreenColumn    int    `json:"screenColumn"`
	ScreenRow       int    `json:"screenRow"`
	WheelUsePattern string `json:"wheelUsePattern"`
}
type TableSetting struct {
	TableCount          int           `json:"tableCount"`
	TableHitProbability []int         `json:"tableHitProbability"`
	WheelData           [][]WheelSlot `json:"wheelData"` // 三维数组：[table][column][]WheelSlot
}

type WheelSlot struct {
	WheelLength int   `json:"wheelLength"`
	NoWinIndex  []int `json:"noWinIndex"`
	WheelData   []int `json:"wheelData"`
}
type SymbolSetting struct {
	SymbolCount     int      `json:"symbolCount"`
	SymbolAttribute []string `json:"symbolAttribute"`
	PayTable        [][]int  `json:"payTable"`
	MixGroupCount   int      `json:"mixGroupCount"`
	MixGroupSetting []any    `json:"mixGroupSetting"` // 类型未知，先设为 any
}
type LineSetting struct {
	MaxBetLine int     `json:"maxBetLine"`
	LineTable  [][]int `json:"lineTable"`
}
type GameHitPatternSetting struct {
	GameHitPattern    string `json:"gameHitPattern"`
	MaxEliminateTimes int    `json:"maxEliminateTimes"`
}
type SpecialFeatureSetting struct {
	SpecialFeatureCount int              `json:"specialFeatureCount"`
	SpecialHitInfo      []SpecialHitInfo `json:"specialHitInfo"`
}

type SpecialHitInfo struct {
	SpecialHitPattern string `json:"specialHitPattern"`
	TriggerEvent      string `json:"triggerEvent"`
	BasePay           int    `json:"basePay"`
}
type ProgressSetting struct {
	TriggerLimitType string       `json:"triggerLimitType"`
	StepSetting      StepSetting  `json:"stepSetting"`
	StageSetting     StageSetting `json:"stageSetting"`
	RoundSetting     RoundSetting `json:"roundSetting"`
}

type StepSetting struct {
	DefaultStep int `json:"defaultStep"`
	AddStep     int `json:"addStep"`
	MaxStep     int `json:"maxStep"`
}

type StageSetting struct {
	DefaultStage int `json:"defaultStage"`
	AddStage     int `json:"addStage"`
	MaxStage     int `json:"maxStage"`
}

type RoundSetting struct {
	DefaultRound int `json:"defaultRound"`
	AddRound     int `json:"addRound"`
	MaxRound     int `json:"maxRound"`
}
type DisplaySetting struct {
	ReadyHandSetting ReadyHandSetting `json:"readyHandSetting"`
}

type ReadyHandSetting struct {
	ReadyHandLimitType string `json:"readyHandLimitType"`
	ReadyHandCount     int    `json:"readyHandCount"`
}
type ExtendSetting struct {
	// 以下字段为可选字段，具体结构随 gameStateType 不同可能略有差异
	InitialChooseTableIndex int     `json:"initialChooseTableIndex,omitempty"`
	RespinProbability       float64 `json:"respinProbability,omitempty"`
	TargetScreen            [][]int `json:"targetScreen,omitempty"`
	RespinFlag              bool    `json:"respinFlag,omitempty"`
	RespinTableWeight       []int   `json:"respinTableWeight,omitempty"`
	RespinTableChoose       []int   `json:"respinTableChoose,omitempty"`
	RespinColumnIndex       int     `json:"respinColumnIndex,omitempty"`

	DampInfoRange    int   `json:"dampInfoRange"`
	EmptySymbolID    int   `json:"emptySymbolID"`
	C2SymbolID       int   `json:"c2SymbolID,omitempty"`
	DampInfoSymbol   []int `json:"dampInfoSymbol,omitempty"`
	RoundLimit       []int `json:"roundLimit,omitempty"`
	ChooseTableIndex []int `json:"chooseTableIndex,omitempty"`
	FowardNRound     int   `json:"fowardNRound,omitempty"`
	AllRoundOdds     int   `json:"allRoundOdds,omitempty"`
	FowardNRoundOdds int   `json:"fowardNRoundOdds,omitempty"`
	OddsHitPattern   []int `json:"oddsHitPattern,omitempty"`
}
type DoubleGameSetting struct {
	DoubleRoundUpperLimit int     `json:"doubleRoundUpperLimit"`
	DoubleBetUpperLimit   int64   `json:"doubleBetUpperLimit"`
	RTP                   float64 `json:"rtp"`
	TieRate               float64 `json:"tieRate"`
}

type BoardDisplaySetting struct {
	WinRankSetting WinRankSetting `json:"winRankSetting"`
}

type WinRankSetting struct {
	BigWin   int `json:"BigWin"`
	MegaWin  int `json:"MegaWin"`
	UltraWin int `json:"UltraWin"`
}
type GameFlowSetting struct {
	ConditionTableWithoutBoardEnd [][]string `json:"conditionTableWithoutBoardEnd"`
}
type LineWinResult struct {
	LineId         int      `json:"lineId"`         // 中奖线编号
	HitDirection   string   `json:"hitDirection"`   // 命中方向（如 "LeftToRight"）
	IsMixGroupFlag bool     `json:"isMixGroupFlag"` // 是否为混合图标组合
	HitMixGroup    int      `json:"hitMixGroup"`    // 混合组 ID，-1 表示无混组
	HitSymbol      int      `json:"hitSymbol"`      // 命中符号 ID（主符号）
	HitWay         int      `json:"hitWay"`         // 命中的连续列数
	HitOdds        int      `json:"hitOdds"`        // 命中赔率
	LineWin        int      `json:"lineWin"`        // 当前线获得的奖励
	ScreenHitData  [][]bool `json:"screenHitData"`  // 每列是否命中，用于标记画面中奖位置
}
type RoundInfo2 struct {
	MaxMultiplier float64 `json:"maxMultiplier"`
	RoundId       int     `json:"roundId"`
}

type Settings struct {
	Music     bool `json:"music"`
	Sound     bool `json:"sound"`
	SecondBet bool `json:"secondBet"`
	Animation bool `json:"animation"`
}

type User struct {
	Settings     Settings `json:"settings"`
	Balance      float64  `json:"balance"`
	ProfileImage string   `json:"profileImage"`
	UserId       string   `json:"userId"`
	Username     string   `json:"username"`
}

type Payload struct {
	RoundsInfo         []RoundInfo2  `json:"roundsInfo"`
	Code               int           `json:"code"`
	ActiveBets         []interface{} `json:"activeBets"`
	ActiveFreeBetsInfo []interface{} `json:"activeFreeBetsInfo"`
	OnlinePlayers      int           `json:"onlinePlayers"`
	RoundId            int           `json:"roundId"`
	StageId            int           `json:"stageId"`
	CurrentMultiplier  float64       `json:"currentMultiplier"`
	User               User          `json:"user"`
	Config             Config        `json:"config"`
}

type Response struct {
	C string  `json:"c"`
	P Payload `json:"p"`
}
type Config struct {
	IsAutoBetFeatureEnabled          bool            `json:"isAutoBetFeatureEnabled"`
	BetPrecision                     int             `json:"betPrecision"`
	MaxBet                           float64         `json:"maxBet"`
	IsAlderneyModalShownOnInit       bool            `json:"isAlderneyModalShownOnInit"`
	IsCurrencyNameHidden             bool            `json:"isCurrencyNameHidden"`
	IsLoginTimer                     bool            `json:"isLoginTimer"`
	IsClockVisible                   bool            `json:"isClockVisible"`
	IsBetsHistoryEndBalanceEnabled   bool            `json:"isBetsHistoryEndBalanceEnabled"`
	BetInputStep                     float64         `json:"betInputStep"`
	AutoBetOptions                   AutoBetOptions  `json:"autoBetOptions"`
	IsGameRulesHaveMaxWin            bool            `json:"isGameRulesHaveMaxWin"`
	IsBetsHistoryStartBalanceEnabled bool            `json:"isBetsHistoryStartBalanceEnabled"`
	IsMaxUserMultiplierEnabled       bool            `json:"isMaxUserMultiplierEnabled"`
	IsShowActivePlayersWidget        bool            `json:"isShowActivePlayersWidget"`
	BackToHomeActionType             string          `json:"backToHomeActionType"`
	InactivityTimeForDisconnect      int             `json:"inactivityTimeForDisconnect"`
	IsActiveGameFocused              bool            `json:"isActiveGameFocused"`
	IsNetSessionEnabled              bool            `json:"isNetSessionEnabled"`
	FullBetTime                      int             `json:"fullBetTime"`
	MinBet                           float64         `json:"minBet"`
	IsGameRulesHaveMinimumBankValue  bool            `json:"isGameRulesHaveMinimumBankValue"`
	IsShowTotalWinWidget             bool            `json:"isShowTotalWinWidget"`
	IsShowBetControlNumber           bool            `json:"isShowBetControlNumber"`
	BetOptions                       []float64       `json:"betOptions"`
	ModalShownOnInit                 string          `json:"modalShownOnInit"`
	IsLiveBetsAndStatisticsHidden    bool            `json:"isLiveBetsAndStatisticsHidden"`
	OnLockUIActions                  string          `json:"onLockUIActions"`
	IsEmbeddedVideoHidden            bool            `json:"isEmbeddedVideoHidden"`
	IsBetTimerBranded                bool            `json:"isBetTimerBranded"`
	DefaultBetValue                  float64         `json:"defaultBetValue"`
	MaxUserWin                       float64         `json:"maxUserWin"`
	IsUseMaskedUsername              bool            `json:"isUseMaskedUsername"`
	IsShowWinAmountUntilNextRound    bool            `json:"isShowWinAmountUntilNextRound"`
	MultiplierPrecision              int             `json:"multiplierPrecision"`
	AutoCashOut                      AutoCashOut     `json:"autoCashOut"`
	IsMultipleBetsEnabled            bool            `json:"isMultipleBetsEnabled"`
	EngagementTools                  EngagementTools `json:"engagementTools"`
	IsFreeBetsEnabled                bool            `json:"isFreeBetsEnabled"`
	PingIntervalMs                   int             `json:"pingIntervalMs"`
	IsLogoUrlHidden                  bool            `json:"isLogoUrlHidden"`
	ChatApiVersion                   int             `json:"chatApiVersion"`
	Currency                         string          `json:"currency"`
	ShowCrashExampleInRules          bool            `json:"showCrashExampleInRules"`
	IsPodSelectAvailable             bool            `json:"isPodSelectAvailable"`
	ReturnToPlayer                   int             `json:"returnToPlayer"`
	IsBalanceValidationEnabled       bool            `json:"isBalanceValidationEnabled"`
	IsHolidayTheme                   bool            `json:"isHolidayTheme"`
	IsGameRulesHaveMultiplierFormula bool            `json:"isGameRulesHaveMultiplierFormula"`
	AccountHistoryActionType         string          `json:"accountHistoryActionType"`
	Chat                             Chat            `json:"chat"`
	IrcDisplayType                   string          `json:"ircDisplayType"`
	GameRulesAutoCashOutType         string          `json:"gameRulesAutoCashOutType"`
}

type AutoBetOptions struct {
	DecreaseOrExceedStopPointReq bool  `json:"decreaseOrExceedStopPointReq"`
	NumberOfRounds               []int `json:"numberOfRounds"`
}

type AutoCashOut struct {
	MinValue     float64 `json:"minValue"`
	DefaultValue float64 `json:"defaultValue"`
	MaxValue     float64 `json:"maxValue"`
}

type EngagementTools struct {
	IsExternalChatEnabled bool `json:"isExternalChatEnabled"`
}

type Chat struct {
	Promo            Promo   `json:"promo"`
	Rain             Rain    `json:"rain"`
	IsGifsEnabled    bool    `json:"isGifsEnabled"`
	SendMessageDelay float64 `json:"sendMessageDelay"`
	IsEnabled        bool    `json:"isEnabled"`
	MaxMessages      int     `json:"maxMessages"`
	MaxMessageLength int     `json:"maxMessageLength"`
}

type Promo struct {
	IsEnabled bool `json:"isEnabled"`
}

type Rain struct {
	IsEnabled         bool    `json:"isEnabled"`
	RainMinBet        float64 `json:"rainMinBet"`
	DefaultNumOfUsers int     `json:"defaultNumOfUsers"`
	MinNumOfUsers     int     `json:"minNumOfUsers"`
	MaxNumOfUsers     int     `json:"maxNumOfUsers"`
	RainMaxBet        float64 `json:"rainMaxBet"`
}

//---------------------------------init end-------------------------------------

// GameStatus 20 waitstart 30 gaming 40 gameover
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func StringToUint16Array(s string) []uint16 {
	runes := []rune(s)            // 将字符串分割为 Unicode 码点
	utf16s := utf16.Encode(runes) // 编码为 UTF-16 码元
	return utf16s
}
func main() {

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "key", "timestamp", "x-trace-id", "x-token", "client-version", "User-Agent", "Cache-Control", "Pragma", "Expires"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/websocket", wsHandler)
	r.POST("/frontendAPI.do", reportConfigHandler)
	r.POST("/rum", reportRumHandler)
	r.POST("/batchLog", reportLogHandler)
	r.POST("/cache/GetGameResultSetting", GetGameResultSetting)
	// r.POST('https://spf-api.pgh-nmgat.com/mascot/get/notify',
	// r.POST('https://spf-api.pgh-nmgat.com/whitelabel',
	// r.POST('https://spf-api.pgh-nmgat.com/settinggetSetting',

	fmt.Println("🚀 服务启动：")
	fmt.Println("- WebSocket 地址：ws://localhost:3333/websocket")
	r.Run(":3333")
}
func reportRumHandler(c *gin.Context) {
	c.String(200, "1\t1")

}

type LogEntry struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace"`
	Level     string `json:"level"`
	// 你可以根据实际内容扩展其他字段
}

func GetGameResultSetting(c *gin.Context) {

}
func reportLogHandler(c *gin.Context) {
	var logs []LogEntry

	if err := c.ShouldBindJSON(&logs); err != nil {
		c.JSON(400, gin.H{"error": "invalid JSON format", "detail": err.Error()})
		return
	}

	// ✅ 打印日志条数和内容做验证
	fmt.Printf("📥 收到日志数量: %d\n", len(logs))
	for _, log := range logs {
		fmt.Printf("Log ID: %s, Namespace: %s, Level: %s\n", log.ID, log.Namespace, log.Level)
	}
	//返回{"data":"12 has been created."}
	c.JSON(200, gin.H{"data": "12 has been created."})

}

func reportConfigHandler(c *gin.Context) {
	action := c.PostForm("action")

	fmt.Println("🔐 收到 action:", action)
	switch action {
	case "101":
		c.JSON(http.StatusOK, gin.H{
			"status": "0000",
			"data": gin.H{
				"ots": "97a28971-122a-4559-a271-397400e9d84d",
			},
		})
	case "6":
		c.JSON(http.StatusOK, gin.H{
			"status": "0000",
			"data": []gin.H{
				{
					"uid":        "demo001685@XX",
					"userName":   "001685",
					"lvl":        0,
					"userStatus": 0,
					"currency":   "XX",
					"timeZone":   "GMT+08:00",
				},
			},
		})
	case "20":
		c.JSON(http.StatusOK, gin.H{
			"status": "0000",
			"data": gin.H{
				"url": "http://aviator.local.com/aviator/?currency=MAD&operator=demo&jurisdiction=CW&lang=pt&return_url&user=53302&token=OTf6G17K3bzYNpWc1uRqTduR2qovP1BR",
				//"url": "http://abcd.jbp.com/?tpg2tl=1&d=1&isApp=true&gName=TreasureBowl_d65c592&lang=cn&homeUrl=&mute=0&gameType=14&mType=14042&x=e9tkQRED2CBLfk0amYQ8IuupMN-wAZcoIKYGa2P4tE2QLQjl6Qg_MmX6PUb8jPYvi0mYBom7O5nmo0unuC3eivfqG3Y3BIDC",
				//	"url": "http://abcd.abcd.com/?tpg2tl=1&d=1&isApp=true&gName=PopPopCandy_096d45b&lang=cn&homeUrl=&mute=0&gameType=14&mType=14087&x=e9tkQRED2CClQmf9gCvFzgwjLyNIEyHpYaWaJUcxXZAYv4XExx8PPCqeD9kNReoH1u1relEAkvZBJu0EJcsF5wKTlotEyTq7",
				//"url": "http://abcd.super.com/?tpg2tl=1&d=1&isApp=true&gName=LuckyDiamond_8dca129&lang=cn&homeUrl=&mute=0&gameType=14&mType=14054&x=e9tkQRED2CBWLS-LUoWn_VDFlws8ozYiyKUUY8aNoIitinh1Hku72QhwxKxXm9gHzAVLCNkc6pWdBRwpN3fwGN1M1OTa9tGh",
			},
		})
	case "19":
		c.JSON(http.StatusOK, gin.H{
			"status": "0000",
			"data": gin.H{
				"isShowAutoPlay": true,
				"result4": gin.H{
					"currency":         "XX",
					"isDemoAccount":    true,
					"showDemoFeatures": true,
					"isApiAccount":     true,
					"isShowJackpot":    false,
					"isShowCurrency":   true,
					"isShowDollarSign": false,
					"decimalPoint":     2,
					"gameGroup": []int{ //14054
						131, 7, 0, 9, 140, 12, 141, 142, 0, 0, 18, 150, 22, 30, 31, 160, 32, 161, 162, 18, 50, 55, 56, 57, 58, 59, 60, 190, 66, 67, 70, 75, 80, 81, 90, 92, 93, 120,
					},
					"functionList": []string{},
				},
				"result6": gin.H{
					"uid":        "demo000428@XX",
					"userName":   "000428",
					"lvl":        0,
					"userStatus": 0,
					"currency":   "XX",
				},
				"result10": gin.H{
					"status":    "0000",
					"sessionID": []string{"", "", "", "A", ""},

					"zone":     "JDB_ZONE_GAME",
					"gsInfo":   "jdb247.net_443_0",
					"gameType": 14,

					//  "machineType": 14042,//聚宝盆
					//"machineType": 14087, //宝宝甜心
					"machineType": 14054,
					"isRecovery":  false,
					"s0":          "",
					"s1":          "",
					"s2":          "",
					"s3":          "CD4414C0DB1C7818B1B9E175ADFFBC9FA5CDA709FC7815878E4D82530BDAFE4BC6A3CBE41610DDA4082A5F572CCBDF83DDD2527BB4B77464B2E147F01304A7F75E520E6EA4E55D2BB00597F9ABCF27B670E9E79272A0BC2C455B29E2CE458E2F17A6212C0BD3EC11E9867F8D66E378D6082A3B720E2A821E8C142CE6F1B1DC1C8E2E2074D9169F29CF95201512B4E266024A411F92BAE456D6BEE57909110F951CB02116DBA632719778FCEE2F44CE9ED841AA906A5AF7D51FBFB15D068F7C2BA73D0A351243C59208960186D5F5A711560B6BF472F0E370EEBC6A14612EDF669C10C84E40236744BC901F0B0E2AA906FB444CF6736D873C88BF6CBDAC6E8C0F2DFB42464D959A614FBD3ADB264BFB78BA2D090E951B845E1B7E00FAC008A2642A35E720F44EC6435B338D33B125804DE9CF33A7B42EC506DFD3E7EF27BC1C9917FADC3904014E8B3140074492E64187172A04D0A06CEBBF57EA3B0852DFB7FB2D09C1A97385601E23BE42575D8E326D65716B1DD24F07457D20232CF89447A70082593D869179A2FB0C4C9645A3217B3A27BCAD0766F4D588800E5E0B2F695C5A6E0E0B0E5398C25CF054B9D5E0B7F9417BDD97001E8DDE3777DA24F23C00CFDAB548B3E91E27A85788CE2461CC374C8F35202A6C9A1370384E3B355747D2C85010EC2C6765AB07E571320D9A40FE3172FB539D9D5A309CA1547BD53BE1B61AD0FC0F1EEE7DD1F0727DD567F4DF2E13",
					// "s3":       "CD4414C0DB1C78180A701358764DC2753E49C01C5456222A8E4D82530BDAFE4B138D3425C162A2368E5B8E31354B29EC9E51756D2A887EDCDF7BCD82D67FA924877174E55B96280065DFD9F850681C0F70E9E79272A0BC2C455B29E2CE458E2F17A6212C0BD3EC11359573D7B5DC8B84082A3B720E2A821E7FA682E64FC4165BE31D5635679CFD8DECEBFBA665E6DF331878625875AFC8B82342E1068E535C25AF1285F0223E487D34E3873B77607F8FDEA3D1340533A6CB6DD17306B58262BFF15CCBB45CBA640B560B6BF472F0E370EEBC6A14612EDF669C10C84E40236744BC901F0B0E2AA906FB444CF6736D873C88BF6CBDAC6E8C0F29AC7924551BEFD6ACEB645832E623C488455466AE8C2203B476C9331522D85482386EDE3BD45DEB215D83E10E139114CBC4E57A57E6197788A99209EE217131A9C0E9CC41B14244505D6E90CF218329B78AFEED6EE7F8DCB79733E160923EFFC919143A5F26F1DA6F5322C99FB59B39C35F262BDC016F3410E05889F33EE1603B1E3EDC0A88BC09A3B2680270F17A7B98F201C741E6499F7ECE7FBDF1F36EF4F241BE34F6FDAD879EB51496B791603FAC538268EA71E73E00E056684E9BC07A25A1251C8EF556968E0C012E395CFDF395CA6AA045C777A36E963E636BD998409CB0328ECEB740F31905B2DA5F63B51EA8AAAE6FEF0A4CE5AAE59984E5248ED2",
					"s4":       "",
					"gameUid":  "demo000428@XX",
					"gamePass": "2313ee4", //宝宝甜心
					//"gamePass":             "b2ba01f",
					"useSSL":               true,
					"streamingUrl":         gin.H{},
					"achievementServerUrl": "",
					"chatServerUrl":        "",
					"isWSBinary":           false,
				},
				"treatmentGroup": gin.H{},
				"gameNotice": gin.H{
					"showRValueInHelp":        false,
					"showHelpOnRValueSection": false,
					"showGameName":            false,
				},
			},
		})
	case "5":
		c.JSON(200, gin.H{
			"status": "0000",
			"data": []gin.H{
				{
					"uid":        "demo000428@XX",
					"userStatus": 0,
					"ts":         1747360970322,
					"timeZone":   "Asia/Taipei",
					"hitJackpot": []interface{}{},
				},
			},
		})
	case "21":
		c.JSON(200, gin.H{
			"status": "0000",
		})
	case "23":
		c.JSON(200, gin.H{
			"status": "0000",
			"data":   []interface{}{},
		})
	case "24":
		c.JSON(200, gin.H{
			"status": "0000",
			"data": gin.H{
				"enable":       true,
				"availability": 0,
			},
		})
	}
}
func HexForGoDeclaration(data []byte) string {
	return fmt.Sprintf("hexData := \"%s\"", strings.ToUpper(hex.EncodeToString(data)))
}
func wsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("WebSocket Upgrade 错误:", err)
		return
	}
	defer conn.Close()

	fmt.Println("✅ WebSocket 客户端连接成功")

	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("读取消息失败:", err)
			break
		}

		if messageType == websocket.BinaryMessage {
			fmt.Println("📥 收到二进制消息:", len(data))
			// // //打印收到的数据
			// fmt.Printf("字节: % x\n", data)
			// 传入 bytes.Reader，跳过前4字节
			reader := bytes.NewReader(data[4:])
			decoded, _ := DecodeSFSObject(reader, data[4:])

			// decoded, consumed := DecodeSFSObject(data2[3:])
			// fmt.Println("📥 consumed:", consumed)
			//fmt.Printf("🧩 解码结果: %+v\n", decoded)
			HandleSFSMessage(conn, decoded)

		}
	}
}

func HandleSFSMessage(conn *websocket.Conn, obj map[string]interface{}) {
	aVal, ok := obj["a"]
	if !ok {
		fmt.Println("❌ 没有找到 'a' 字段")
		return
	}

	// 👉 统一把 a 转成 int
	var aInt int
	switch v := aVal.(type) {
	case int:
		aInt = v
	case int16:
		aInt = int(v)
	case int32:
		aInt = int(v)
	case float64:
		aInt = int(v)
	default:
		fmt.Printf("❌ 无法识别 'a' 字段类型: %T (%v)\n", v, v)
		return
	}

	pVal, ok := obj["p"]
	if !ok {
		fmt.Println("❌ 没有找到 'p' 字段")
		return
	}
	pMap, ok := pVal.(map[string]interface{})
	if !ok {
		fmt.Println("❌ 'p' 字段不是 map 类型")
		return
	}

	fmt.Printf("🎯 收到消息 a=%d, p内容=%+v\n", aInt, pMap)

	switch aInt {
	case 0: //握手
		handleHandshake(conn, pMap)

	case 1: //登陆
		handleLogin(conn, pMap)
		// CallExtensionResponse(conn, pMap)
	// case 2:
	// 	handleJoinRoom(conn, pMap)
	case 29: //心跳
		handleHeartbeat(conn, pMap)
	case 13: //嵌套协议
		handleCallExtension(conn, pMap)
	default:
		fmt.Printf("⚠️ 未知消息编号: %d\n", aInt)
	}
}

func DecodeSFSObject(reader *bytes.Reader, fullData []byte) (map[string]interface{}, int) {
	startLen := reader.Len()
	result := make(map[string]interface{})

	var fieldCount uint16
	if err := binary.Read(reader, binary.BigEndian, &fieldCount); err != nil {
		fmt.Println("❌ 字段数量读取失败:", err)
		return result, 0
	}
	// fmt.Printf("📦 字段数量: %d\n", fieldCount)

	for i := 0; i < int(fieldCount); i++ {
		// offset := len(fullData) - reader.Len()
		// fmt.Printf("\n🧩 解析字段 %d, 偏移: %d, 剩余: %d 字节\n", i+1, offset, reader.Len())

		// remainingBytes := fullData[offset:]
		// fmt.Printf("📦 剩余原始字节: % X\n", remainingBytes)

		var nameLen uint16
		if err := binary.Read(reader, binary.BigEndian, &nameLen); err != nil {
			fmt.Println("❌ 字段名长度读取失败:", err)
			break
		}
		// fmt.Printf("长度: %d 字节\n", nameLen)

		nameBytes := make([]byte, nameLen)
		if _, err := io.ReadFull(reader, nameBytes); err != nil {
			fmt.Println("❌ 字段名读取失败:", err)
			break
		}
		fieldName := string(nameBytes)

		fieldType, err := reader.ReadByte()
		if err != nil {
			fmt.Println("❌ 字段类型读取失败:", err)
			break
		}
		// fmt.Printf("🔑 字段名: %s, 类型: 0x%02X\n", fieldName, fieldType)

		switch fieldType {
		case TypeNull:
			result[fieldName] = nil
			fmt.Println("✅ null")
		case TypeBool:
			b, err := reader.ReadByte()
			if err != nil {
				fmt.Println("❌ bool 读取失败:", err)
				break
			}
			result[fieldName] = b != 0
			// fmt.Printf("✅ bool: %v\n", b != 0)

		case 0x02: // BYTE
			b, err := reader.ReadByte()
			if err != nil {
				fmt.Println("❌ byte 读取失败:", err)
				break
			}
			result[fieldName] = b
			// fmt.Printf("✅ byte: %d\n", b)

		case 0x03: // SHORT
			var val int16
			if err := binary.Read(reader, binary.BigEndian, &val); err != nil {
				fmt.Println("❌ short 读取失败:", err)
				break
			}
			result[fieldName] = val
			// fmt.Printf("✅ short: %d\n", val)
		case TypeInt:
			var val int32
			if err := binary.Read(reader, binary.BigEndian, &val); err != nil {
				fmt.Println("❌ int 读取失败:", err)
				break
			}
			result[fieldName] = val
			// fmt.Printf("✅ int: %d\n", val)

		case TypeLong:
			var val int64
			if err := binary.Read(reader, binary.BigEndian, &val); err != nil {
				fmt.Println("❌ long 读取失败:", err)
				break
			}
			result[fieldName] = val
			// fmt.Printf("✅ long: %d\n", val)

		case TypeFloat:
			var val float32
			if err := binary.Read(reader, binary.BigEndian, &val); err != nil {
				fmt.Println("❌ float 读取失败:", err)
				break
			}
			result[fieldName] = val
			// fmt.Printf("✅ float: %f\n", val)

		case TypeDouble:
			var val float64
			if err := binary.Read(reader, binary.BigEndian, &val); err != nil {
				fmt.Println("❌ double 读取失败:", err)
				break
			}
			result[fieldName] = val
			// fmt.Printf("✅ double: %f\n", val)
		case 0x08: // UTF_STRING
			var strlen uint16
			if err := binary.Read(reader, binary.BigEndian, &strlen); err != nil {
				fmt.Println("❌ 字符串长度读取失败:", err)
				break
			}
			str := make([]byte, strlen)
			if _, err := io.ReadFull(reader, str); err != nil {
				fmt.Println("❌ 字符串读取失败:", err)
				break
			}
			result[fieldName] = string(str)
			//fmt.Printf("✅ string: %s\n", string(str))
		case TypeUtfStringArray:
			var count int16
			if err := binary.Read(reader, binary.BigEndian, &count); err != nil {
				fmt.Println("❌ UTF_STRING_ARRAY 长度读取失败:", err)
				break
			}
			arr := make([]string, count)
			for i := int16(0); i < count; i++ {
				var l uint16
				binary.Read(reader, binary.BigEndian, &l)
				b := make([]byte, l)
				io.ReadFull(reader, b)
				arr[i] = string(b)
			}
			result[fieldName] = arr
			fmt.Printf("✅ UTF_STRING_ARRAY: %+v\n", arr)

		case TypeIntArray:
			var count int16
			if err := binary.Read(reader, binary.BigEndian, &count); err != nil {
				fmt.Println("❌ INT_ARRAY 长度读取失败:", err)
				break
			}
			arr := make([]int32, count)
			for i := int16(0); i < count; i++ {
				binary.Read(reader, binary.BigEndian, &arr[i])
			}
			result[fieldName] = arr
			// fmt.Printf("✅ INT_ARRAY: %+v\n", arr)

		case TypeDoubleArray:
			var count int16
			if err := binary.Read(reader, binary.BigEndian, &count); err != nil {
				fmt.Println("❌ DOUBLE_ARRAY 长度读取失败:", err)
				break
			}
			arr := make([]float64, count)
			for i := int16(0); i < count; i++ {
				binary.Read(reader, binary.BigEndian, &arr[i])
			}
			result[fieldName] = arr
			// fmt.Printf("✅ DOUBLE_ARRAY: %+v\n", arr)
		case 0x12: // NESTED SFSObject
			// fmt.Printf("🧬 嵌套字段 %s 开始递归解析...\n", fieldName)
			subStart := len(fullData) - reader.Len()
			subResult, _ := DecodeSFSObject(reader, fullData[subStart:])
			result[fieldName] = subResult
			// fmt.Printf("✅ 嵌套字段 %s 完成\n", fieldName)
		case TypeSFSArray:
			var count int16
			if err := binary.Read(reader, binary.BigEndian, &count); err != nil {
				fmt.Println("❌ SFS_ARRAY 长度读取失败:", err)
				break
			}
			// fmt.Printf("🔁 SFSArray 长度: %d\n", count)
			arr := make([]interface{}, count)
			for i := int16(0); i < count; i++ {
				typ, err := reader.ReadByte()
				if err != nil {
					fmt.Println("❌ SFSArray 元素类型读取失败:", err)
					break
				}
				// 👇 递归伪装字段名处理：用 index 作为临时字段名
				fakeMap := map[string]interface{}{}
				fakeField := fmt.Sprintf("%d", i)
				switch typ {
				case TypeSFSObject:
					subStart := len(fullData) - reader.Len()
					subObj, _ := DecodeSFSObject(reader, fullData[subStart:])
					arr[i] = subObj
				default:
					reader.UnreadByte()
					DecodeSFSObjectElement(reader, fullData, fakeField, typ, fakeMap)
					arr[i] = fakeMap[fakeField]
				}
			}
			result[fieldName] = arr
			// fmt.Printf("✅ SFS_ARRAY: %+v\n", arr)
		default:
			fmt.Printf("⚠️ 不支持字段类型: 0x%02X (%s)\n", fieldType, fieldName)
		}
	}

	consumed := startLen - reader.Len()
	// fmt.Printf("✅ 解码完成, 消耗字节: %d\n", consumed)
	return result, consumed
}
func DecodeSFSObjectElement(reader *bytes.Reader, fullData []byte, fieldName string, fieldType byte, result map[string]interface{}) {
	switch fieldType {
	case TypeNull:
		result[fieldName] = nil

	case TypeBool:
		b, _ := reader.ReadByte()
		result[fieldName] = b != 0

	case TypeByte:
		b, _ := reader.ReadByte()
		result[fieldName] = b

	case TypeShort:
		var val int16
		binary.Read(reader, binary.BigEndian, &val)
		result[fieldName] = val

	case TypeInt:
		var val int32
		binary.Read(reader, binary.BigEndian, &val)
		result[fieldName] = val

	case TypeLong:
		var val int64
		binary.Read(reader, binary.BigEndian, &val)
		result[fieldName] = val

	case TypeFloat:
		var val float32
		binary.Read(reader, binary.BigEndian, &val)
		result[fieldName] = val

	case TypeDouble:
		var val float64
		binary.Read(reader, binary.BigEndian, &val)
		result[fieldName] = val

	case TypeUtfString:
		var strlen uint16
		binary.Read(reader, binary.BigEndian, &strlen)
		str := make([]byte, strlen)
		io.ReadFull(reader, str)
		result[fieldName] = string(str)

	case TypeUtfStringArray:
		var count int16
		binary.Read(reader, binary.BigEndian, &count)
		arr := make([]string, count)
		for i := int16(0); i < count; i++ {
			var l uint16
			binary.Read(reader, binary.BigEndian, &l)
			b := make([]byte, l)
			io.ReadFull(reader, b)
			arr[i] = string(b)
		}
		result[fieldName] = arr

	case TypeIntArray:
		var count int16
		binary.Read(reader, binary.BigEndian, &count)
		arr := make([]int32, count)
		for i := int16(0); i < count; i++ {
			binary.Read(reader, binary.BigEndian, &arr[i])
		}
		result[fieldName] = arr

	case TypeDoubleArray:
		var count int16
		binary.Read(reader, binary.BigEndian, &count)
		arr := make([]float64, count)
		for i := int16(0); i < count; i++ {
			binary.Read(reader, binary.BigEndian, &arr[i])
		}
		result[fieldName] = arr

	case TypeSFSObject:
		subStart := len(fullData) - reader.Len()
		obj, _ := DecodeSFSObject(reader, fullData[subStart:])
		result[fieldName] = obj

	default:
		fmt.Printf("⚠️ DecodeSFSObjectElement 暂不支持字段类型: 0x%02X\n", fieldType)
	}
}

func writeString(buf *bytes.Buffer, s string) {
	binary.Write(buf, binary.BigEndian, uint16(len(s)))
	buf.Write([]byte(s))
}

type SFSObject struct {
	fields []SFSField
}

type SFSField struct {
	Name  string
	Type  byte
	Value interface{}
}

// const (
//
//	TypeByte      = 0x02
//	TypeShort     = 0x03
//	TypeInt       = 0x04
//	TypeUtfString = 0x08
//	TypeSFSObject = 0x12
//
// )
const (
	TypeNull           = 0x00 // NULL
	TypeBool           = 0x01 // BOOL
	TypeByte           = 0x02 // BYTE
	TypeShort          = 0x03 // SHORT
	TypeInt            = 0x04 // INT
	TypeLong           = 0x05 // LONG
	TypeFloat          = 0x06 // FLOAT
	TypeDouble         = 0x07 // DOUBLE
	TypeUtfString      = 0x08 // UTF_STRING
	TypeBoolArray      = 0x09 // BOOL_ARRAY
	TypeByteArray      = 0x0A // BYTE_ARRAY
	TypeShortArray     = 0x0B // SHORT_ARRAY
	TypeIntArray       = 0x0C // INT_ARRAY
	TypeLongArray      = 0x0D // LONG_ARRAY
	TypeFloatArray     = 0x0E // FLOAT_ARRAY
	TypeDoubleArray    = 0x0F // DOUBLE_ARRAY
	TypeUtfStringArray = 0x10 // UTF_STRING_ARRAY
	TypeSFSArray       = 0x11 // SFS_ARRAY
	TypeSFSObject      = 0x12 // SFS_OBJECT
	TypeText           = 0x14 // TEXT
)

func NewSFSObject() *SFSObject {
	return &SFSObject{}
}

func (s *SFSObject) Put(name string, typeId byte, value interface{}) {
	s.fields = append(s.fields, SFSField{Name: name, Type: typeId, Value: value})
}

func (s *SFSObject) ToBinary() []byte {
	buf := new(bytes.Buffer)

	// 写 SFSObject type
	buf.WriteByte(0x12)

	// 写字段数量
	binary.Write(buf, binary.BigEndian, uint16(len(s.fields)))

	// 写每个字段
	for _, f := range s.fields {
		binary.Write(buf, binary.BigEndian, uint16(len(f.Name)))
		buf.Write([]byte(f.Name))
		buf.WriteByte(f.Type)

		switch f.Type {
		case TypeByte:
			buf.WriteByte(f.Value.(byte))
		case TypeShort:
			binary.Write(buf, binary.BigEndian, f.Value.(int16))
		case TypeInt:
			binary.Write(buf, binary.BigEndian, f.Value.(int32))
		case TypeUtfString:
			str := f.Value.(string)
			binary.Write(buf, binary.BigEndian, uint16(len(str)))
			buf.Write([]byte(str))
		case TypeSFSObject:
			child := f.Value.(*SFSObject)
			buf.Write(child.ToBinary())
		default:
			panic(fmt.Sprintf("Unsupported field type: %d", f.Type))
		}
	}

	return buf.Bytes()
}

func OnPacketWrite(controllerId byte, actionId int16, paramPayload []byte) []byte {
	top := new(bytes.Buffer)

	// 写顶层 SFSObject
	top.WriteByte(0x12)                            // SFSObject 标识
	binary.Write(top, binary.BigEndian, uint16(3)) // 字段数量3

	// 👉 p 字段 (嵌套SFSObject)
	writeString(top, "p")
	top.WriteByte(TypeSFSObject)
	top.Write(paramPayload)

	// 👉 a 字段 (short)
	writeString(top, "a")
	top.WriteByte(TypeShort)
	binary.Write(top, binary.BigEndian, actionId)

	// 👉 c 字段 (byte)
	writeString(top, "c")
	top.WriteByte(TypeByte)
	top.WriteByte(controllerId)

	// 包装 header
	final := new(bytes.Buffer)
	final.WriteByte(0x80)                                    // 固定
	binary.Write(final, binary.BigEndian, uint16(top.Len())) // 长度2字节
	final.Write(top.Bytes())

	return final.Bytes()
}

func OnPacketWriteHandshakeFix() []byte {
	buf := new(bytes.Buffer)

	// 顶层 SFSObject
	buf.WriteByte(0x12)                            // SFSObject
	binary.Write(buf, binary.BigEndian, uint16(3)) // 字段数量3

	// 1️⃣ 写 p 字段
	writeString(buf, "p")
	buf.WriteByte(TypeSFSObject)

	// p字段内部 (注意这里 p内部先打字段数量3)
	{
		pInner := new(bytes.Buffer)

		// p内部字段数量
		binary.Write(pInner, binary.BigEndian, uint16(3))

		// ct字段
		writeString(pInner, "ct")
		pInner.WriteByte(TypeInt)
		binary.Write(pInner, binary.BigEndian, int32(1024))

		// ms字段
		writeString(pInner, "ms")
		pInner.WriteByte(TypeInt)
		binary.Write(pInner, binary.BigEndian, int32(500000))

		// tk字段
		writeString(pInner, "tk")
		pInner.WriteByte(TypeUtfString)
		tk := "24f1ff9beba507db9394ff37e6123ee0"
		binary.Write(pInner, binary.BigEndian, uint16(len(tk)))
		pInner.Write([]byte(tk))

		// 写入 p字段内容
		buf.Write(pInner.Bytes())
	}

	// 2️⃣ 写 a 字段 cmd
	writeString(buf, "a")
	buf.WriteByte(TypeShort)
	binary.Write(buf, binary.BigEndian, int16(0))

	// 3️⃣ 写 c 字段
	writeString(buf, "c")
	buf.WriteByte(TypeByte)
	buf.WriteByte(0)

	// 外部 header
	final := new(bytes.Buffer)
	final.WriteByte(0x80)
	binary.Write(final, binary.BigEndian, uint16(buf.Len()))
	final.Write(buf.Bytes())

	return final.Bytes()
}

// 通用打包发送函数
func BuildSFSMessage(a int16, c interface{}, p map[string]interface{}) []byte {
	buf := new(bytes.Buffer)

	// 写 SFSObject 类型
	buf.WriteByte(TypeSFSObject)
	binary.Write(buf, binary.BigEndian, uint16(3)) // 外层字段数量固定为3个：p, a, c

	// 1️⃣ 写 p 字段（payload）
	writeString(buf, "p")
	buf.WriteByte(TypeSFSObject)

	pInner := BuildSFSObject(p)
	buf.Write(pInner)

	// 2️⃣ 写 a 字段（action）
	writeString(buf, "a")
	buf.WriteByte(TypeShort)
	binary.Write(buf, binary.BigEndian, a)

	// 3️⃣ 写 c 字段（controller ID，可为 byte 或 string）
	writeString(buf, "c")
	switch v := c.(type) {
	case byte:
		buf.WriteByte(TypeByte)
		buf.WriteByte(v)
	case string:
		buf.WriteByte(TypeUtfString)
		binary.Write(buf, binary.BigEndian, uint16(len(v)))
		buf.Write([]byte(v))
	case []string:
		buf.WriteByte(TypeUtfStringArray)
		binary.Write(buf, binary.BigEndian, int16(len(v)))
		for _, s := range v {
			binary.Write(buf, binary.BigEndian, uint16(len(s)))
			buf.Write([]byte(s))
		}
	case int:
		if v >= 0 && v <= 255 {
			buf.WriteByte(TypeByte)
			buf.WriteByte(byte(v))
		} else if v >= -32768 && v <= 32767 {
			buf.WriteByte(TypeShort)
			binary.Write(buf, binary.BigEndian, int16(v))
		} else {
			fmt.Printf("⚠️ c 类型 int 超出支持范围: %d，默认使用 byte=0\n", v)
			buf.WriteByte(TypeByte)
			buf.WriteByte(0)
		}
	default:
		fmt.Printf("⚠️ 不支持的 c 类型: %T (%v)，默认使用 byte=0\n", v, v)
		buf.WriteByte(TypeByte)
		buf.WriteByte(0)
	}

	// 封装头部
	final := new(bytes.Buffer)
	final.WriteByte(0x80)
	binary.Write(final, binary.BigEndian, uint16(buf.Len()))
	final.Write(buf.Bytes())
	// if a == 13 {
	// 	fmt.Println("📤 SFSMessage (a == 13):")
	// 	fmt.Println(final.Bytes())
	// }
	fmt.Println("发送消息", final.Len())

	return final.Bytes()
}

// 构建嵌套的 SFSObject（二进制）
func BuildSFSObject(obj map[string]interface{}) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint16(len(obj)))

	for key, val := range obj {
		writeString(buf, key)

		switch v := val.(type) {
		case int:
			buf.WriteByte(TypeInt)
			binary.Write(buf, binary.BigEndian, int32(v))
		case int16:
			buf.WriteByte(TypeShort)
			binary.Write(buf, binary.BigEndian, v)
		case int32:
			buf.WriteByte(TypeInt)
			binary.Write(buf, binary.BigEndian, v)
		case byte:
			buf.WriteByte(TypeByte)
			buf.WriteByte(v)
		case bool:
			buf.WriteByte(TypeBool)
			if v {
				buf.WriteByte(1)
			} else {
				buf.WriteByte(0)
			}
		case float64:
			buf.WriteByte(TypeDouble)
			binary.Write(buf, binary.BigEndian, v)
		case float32:
			buf.WriteByte(TypeFloat)
			binary.Write(buf, binary.BigEndian, v)
		case int64:
			buf.WriteByte(TypeLong)
			binary.Write(buf, binary.BigEndian, v)
		case string:
			buf.WriteByte(TypeUtfString)
			binary.Write(buf, binary.BigEndian, uint16(len(v)))
			buf.Write([]byte(v))
		case []interface{}:
			buf.WriteByte(TypeSFSArray)
			writeSFSArray(buf, v)
		case map[string]interface{}:
			buf.WriteByte(TypeSFSObject)
			inner := BuildSFSObject(v)
			buf.Write(inner)

		case []float64:
			buf.WriteByte(TypeDoubleArray)
			binary.Write(buf, binary.BigEndian, int16(len(v))) // 写入元素数量
			for _, f := range v {
				binary.Write(buf, binary.BigEndian, f) // 写入每个 float64 值
			}
			// ✅ 新增支持 []int → INT_ARRAY
		case []int:
			buf.WriteByte(TypeIntArray)
			binary.Write(buf, binary.BigEndian, int16(len(v)))
			for _, i := range v {
				binary.Write(buf, binary.BigEndian, int32(i))
			}
		case []int32:
			buf.WriteByte(TypeIntArray)
			binary.Write(buf, binary.BigEndian, int16(len(v))) // 若和 []int 一致用 int16
			for _, i := range v {
				binary.Write(buf, binary.BigEndian, int32(i)) // 直接写 int32 值
			}
		case []uint16:
			buf.WriteByte(TypeShortArray)
			binary.Write(buf, binary.BigEndian, int16(len(v))) // 元素个数
			for _, s := range v {
				binary.Write(buf, binary.BigEndian, s) // 每个 uint16 元素
			}
		case []byte:
			buf.WriteByte(TypeByteArray)
			binary.Write(buf, binary.BigEndian, int32(len(v))) // 4字节长度
			buf.Write(v)                                       // 写入原始字节
		// ✅ 新增支持 []string → UTF_STRING_ARRAY
		case []string:
			buf.WriteByte(TypeUtfStringArray)
			binary.Write(buf, binary.BigEndian, int16(len(v)))
			for _, s := range v {
				binary.Write(buf, binary.BigEndian, uint16(len(s)))
				buf.Write([]byte(s))
			}
		case []map[string]interface{}:
			buf.WriteByte(TypeSFSArray)
			binary.Write(buf, binary.BigEndian, int16(len(v))) // 数组长度

			for _, item := range v {
				buf.WriteByte(TypeSFSObject)  // 每个元素类型
				inner := BuildSFSObject(item) // 递归构造
				buf.Write(inner)
			}

		default:
			fmt.Printf("⚠️ 暂不支持类型: %T (%v)\n", v, v)
		}
	}

	return buf.Bytes()
}

// 写入 SFSArray
func writeSFSArray(buf *bytes.Buffer, arr []interface{}) {
	binary.Write(buf, binary.BigEndian, int16(len(arr))) // 先写元素数量

	for _, item := range arr {
		switch v := item.(type) {
		case int:
			buf.WriteByte(TypeInt)
			binary.Write(buf, binary.BigEndian, int32(v))
		case int32:
			buf.WriteByte(TypeInt)
			binary.Write(buf, binary.BigEndian, v)
		case int16:
			buf.WriteByte(TypeShort)
			binary.Write(buf, binary.BigEndian, v)
		case int64:
			buf.WriteByte(TypeLong)
			binary.Write(buf, binary.BigEndian, v)
		case byte:
			buf.WriteByte(TypeByte)
			buf.WriteByte(v)
		case bool:
			buf.WriteByte(TypeBool)
			if v {
				buf.WriteByte(0x01)
			} else {
				buf.WriteByte(0x00)
			}
		case float32:
			buf.WriteByte(TypeFloat)
			binary.Write(buf, binary.BigEndian, v)
		case float64:
			buf.WriteByte(TypeDouble)
			binary.Write(buf, binary.BigEndian, v)

		case string:
			buf.WriteByte(TypeUtfString)
			binary.Write(buf, binary.BigEndian, uint16(len(v)))
			buf.Write([]byte(v))
		case map[string]interface{}:
			buf.WriteByte(TypeSFSObject)
			inner := BuildSFSObject(v)
			buf.Write(inner)
		case []interface{}: // 允许嵌套数组
			buf.WriteByte(TypeSFSArray)
			writeSFSArray(buf, v)
		case []int:
			buf.WriteByte(TypeIntArray)
			binary.Write(buf, binary.BigEndian, int16(len(v)))
			for _, i := range v {
				binary.Write(buf, binary.BigEndian, int32(i))
			}
		case []float64:
			buf.WriteByte(TypeDoubleArray)
			binary.Write(buf, binary.BigEndian, int16(len(v)))
			for _, f := range v {
				binary.Write(buf, binary.BigEndian, f)
			}
		case []string:
			buf.WriteByte(TypeUtfStringArray)
			binary.Write(buf, binary.BigEndian, int16(len(v)))
			for _, s := range v {
				binary.Write(buf, binary.BigEndian, uint16(len(s)))
				buf.Write([]byte(s))
			}

		default:
			fmt.Printf("⚠️ SFSArray中暂不支持类型: %T (%v)\n", v, v)
		}
	}
}
func handleHeartbeat(conn *websocket.Conn, obj map[string]interface{}) {
	p := map[string]interface{}{}
	packet := BuildSFSMessage(29, 0, p)
	conn.WriteMessage(websocket.BinaryMessage, packet)
}
func handleHandshake(conn *websocket.Conn, obj map[string]interface{}) {
	p := map[string]interface{}{
		"ct": 2147483647,
		"ms": 500000,
		"tk": "f56cf347bca59b945ea8b9fe4a1af0e6",
	}

	packet := BuildSFSMessage(0, 0, p)
	conn.WriteMessage(websocket.BinaryMessage, packet)
}
func handleLogin(conn *websocket.Conn, obj map[string]interface{}) {

	// roomList := []interface{}{
	// 	[]interface{}{2, "SLOT_ROOM", "default", true, false, false, int16(1839), int16(5000), []interface{}{}, int16(0), int16(0)},
	// 	[]interface{}{3, "PUSOYS_LOBBY", "default", false, false, false, int16(21), int16(5000), []interface{}{}},
	// 	[]interface{}{4, "TONGITS_LOBBY", "default", false, false, false, int16(29), int16(5000), []interface{}{}},
	// 	[]interface{}{5, "RUMMY_LOBBY", "default", false, false, false, int16(0), int16(5000), []interface{}{}},
	// 	[]interface{}{6, "RUNNING_GAME", "default", false, false, false, int16(16), int16(5000), []interface{}{}},
	// 	[]interface{}{236, "18020", "default", false, false, false, int16(1), int16(5000), []interface{}{}},
	// 	[]interface{}{237, "18021", "default", false, false, false, int16(0), int16(5000), []interface{}{}},
	// 	[]interface{}{238, "SINGLE_SPIN", "default", false, false, false, int16(8), int16(5000), []interface{}{}},
	// 	[]interface{}{239, "18026", "default", false, false, false, int16(6), int16(5000), []interface{}{}},
	// 	[]interface{}{240, "MINES", "default", false, false, false, int16(91), int16(5000), []interface{}{}},
	// 	[]interface{}{241, "CASINO_ROOM", "default", false, false, false, int16(0), int16(5000), []interface{}{}},
	// 	[]interface{}{242, "18022", "default", false, false, false, int16(0), int16(5000), []interface{}{}},
	// }

	// // 构造返回数据 map[payload]
	// p := map[string]interface{}{
	// 	"rs": int16(0),        // 登录成功
	// 	"zn": "JDB_ZONE_GAME", // 区域名
	// 	"un": obj["un"],       // 用户名
	// 	"pi": int16(0),        // playerId
	// 	"rl": roomList,        // 房间列表
	// 	"id": int32(1928827),  // 用户 ID
	// }
	roomList := []interface{}{
		[]interface{}{0, "game_state", "default", false, false, false, 0, 20, []interface{}{}},
	}
	// 构造返回数据 map[payload]
	p := map[string]interface{}{
		"rs": int16(0),                   // 登录成功
		"zn": "aviator_core_inst2_demo1", // 区域名
		"un": obj["un"],                  // 用户名
		"pi": int16(0),                   // playerId
		"rl": roomList,                   // 房间列表
		"id": int32(1928827),             // 用户 ID
	}
	// 构造封包并发送
	packet := BuildSFSMessage(1, 0, p)
	conn.WriteMessage(websocket.BinaryMessage, packet)

	fmt.Println("✅ 已发送 Login 响应")
	AfterLogin(conn, obj)
}
func handleUserCountChange(conn *websocket.Conn, obj map[string]interface{}) {

	p := map[string]interface{}{
		"r":  int32(3),  // Room ID，使用 int32
		"uc": int16(20), // 用户数量，short = int16
	}
	// 构造封包并发送
	packet := BuildSFSMessage(1001, 0, p)
	conn.WriteMessage(websocket.BinaryMessage, packet)

	fmt.Println("✅ 已发送 Login 响应")
}

func AfterLogin(conn *websocket.Conn, obj map[string]interface{}) {
	p := map[string]interface{}{
		"c": "init",
		"p": map[string]interface{}{
			"roundsInfo": []map[string]interface{}{
				{"maxMultiplier": 2.25, "roundId": 8241979},
				{"maxMultiplier": 1.35, "roundId": 8241977},
				{"maxMultiplier": 1.37, "roundId": 8241974},
				{"maxMultiplier": 1.0, "roundId": 8241972},
				{"maxMultiplier": 2.83, "roundId": 8241969},
				{"maxMultiplier": 2.37, "roundId": 8241965},
				{"maxMultiplier": 2.52, "roundId": 8241964},
				{"maxMultiplier": 1.0, "roundId": 8241963},
				{"maxMultiplier": 4.41, "roundId": 8241960},
				{"maxMultiplier": 1.74, "roundId": 8241958},
				{"maxMultiplier": 1.98, "roundId": 8241955},
				{"maxMultiplier": 1.03, "roundId": 8241954},
				{"maxMultiplier": 2.41, "roundId": 8241952},
				{"maxMultiplier": 1.52, "roundId": 8241951},
				{"maxMultiplier": 1.45, "roundId": 8241949},
				{"maxMultiplier": 2.37, "roundId": 8241946},
				{"maxMultiplier": 1.29, "roundId": 8241944},
				{"maxMultiplier": 1.12, "roundId": 8241943},
				{"maxMultiplier": 18.16, "roundId": 8241939},
				{"maxMultiplier": 1.0, "roundId": 8241938},
				{"maxMultiplier": 1.8, "roundId": 8241937},
				{"maxMultiplier": 4.05, "roundId": 8241935},
				{"maxMultiplier": 178.86, "roundId": 8241931},
				{"maxMultiplier": 1.73, "roundId": 8241928},
				{"maxMultiplier": 1.77, "roundId": 8241926},
			},
			"code":               200,
			"activeBets":         []interface{}{}, // 空 SFSArray
			"activeFreeBetsInfo": []interface{}{}, // 空 SFSArray
			"onlinePlayers":      1329,
			"user": map[string]interface{}{
				"settings": map[string]interface{}{
					"music":     false,
					"sound":     false,
					"secondBet": true,
					"animation": true,
				},
				"balance":      5000.0,
				"profileImage": "av-21.png",
				"userId":       "33687&&demo",
				"username":     "demo_71815",
			},
			"config": map[string]interface{}{
				"isAutoBetFeatureEnabled":        true,
				"betPrecision":                   2,
				"maxBet":                         1000.0,
				"isAlderneyModalShownOnInit":     false,
				"isCurrencyNameHidden":           false,
				"isLoginTimer":                   false,
				"isClockVisible":                 false,
				"isBetsHistoryEndBalanceEnabled": false,
				"betInputStep":                   1.0,
				"autoBetOptions": map[string]interface{}{
					"decreaseOrExceedStopPointReq": true,
					"numberOfRounds":               []int{10, 20, 50, 100},
				},
				"isGameRulesHaveMaxWin":            false,
				"isBetsHistoryStartBalanceEnabled": false,
				"isMaxUserMultiplierEnabled":       false,
				"isShowActivePlayersWidget":        true,
				"backToHomeActionType":             "navigate",
				"inactivityTimeForDisconnect":      0,
				"isActiveGameFocused":              false,
				"isNetSessionEnabled":              false,
				"fullBetTime":                      5000,
				"minBet":                           1.0,
				"isGameRulesHaveMinimumBankValue":  false,
				"isShowTotalWinWidget":             true,
				"isShowBetControlNumber":           false,
				"betOptions":                       []float64{10, 20, 50, 100},
				"modalShownOnInit":                 "none",
				"isLiveBetsAndStatisticsHidden":    false,
				"onLockUIActions":                  "cancelBet",
				"isEmbeddedVideoHidden":            false,
				"isBetTimerBranded":                true,
				"defaultBetValue":                  1.0,
				"maxUserWin":                       100000.0,
				"isUseMaskedUsername":              true,
				"isShowWinAmountUntilNextRound":    false,
				"multiplierPrecision":              2,
				"autoCashOut": map[string]interface{}{
					"minValue":     1.01,
					"defaultValue": 1.1,
					"maxValue":     100.0,
				},
				"isMultipleBetsEnabled": true,
				"engagementTools": map[string]interface{}{
					"isExternalChatEnabled": false,
				},
				"isFreeBetsEnabled":                true,
				"pingIntervalMs":                   15000,
				"isLogoUrlHidden":                  false,
				"chatApiVersion":                   2,
				"currency":                         "MAD",
				"showCrashExampleInRules":          false,
				"isPodSelectAvailable":             true,
				"returnToPlayer":                   97,
				"isBalanceValidationEnabled":       true,
				"isHolidayTheme":                   false,
				"isGameRulesHaveMultiplierFormula": false,
				"accountHistoryActionType":         "navigate",
				"chat": map[string]interface{}{
					"promo": map[string]interface{}{
						"isEnabled": true,
					},
					"rain": map[string]interface{}{
						"isEnabled":         false,
						"rainMinBet":        1.0,
						"defaultNumOfUsers": 5,
						"minNumOfUsers":     3,
						"maxNumOfUsers":     10,
						"rainMaxBet":        100.0,
					},
					"isGifsEnabled":    true,
					"sendMessageDelay": 5000.0,
					"isEnabled":        false,
					"maxMessages":      70,
					"maxMessageLength": 160,
				},
				"ircDisplayType":           "modal",
				"gameRulesAutoCashOutType": "default",
			},
			"roundId":           8241983,
			"stageId":           2,
			"currentMultiplier": 1.17,
		},
	}

	packet := BuildSFSMessage(13, 1, p)
	conn.WriteMessage(websocket.BinaryMessage, packet)
}
func handleCallExtension(conn *websocket.Conn, obj map[string]interface{}) {
	// 从 obj 中提取扩展名、参数、请求ID
	cmd, _ := obj["c"].(string)
	params, _ := obj["p"].(map[string]interface{})
	reqId, _ := obj["r"]

	fmt.Printf("📨 CallExtension: cmd=%s, reqId=%v, params=%v\n", cmd, reqId, params)

	switch cmd {
	case "betHandler":
		handleBet(conn, params)
	case "cashOutHandler":
		handleCashout(conn, params)
	case "openCellHandler":
		handleOpenCell(conn, params)

	case "autoPlayHandler":
		handleAutoPlay(conn, params)
	case "gameLogin":
		handleGameLogin(conn, params)
	case "h5.init":
		handleH5Init(conn, params)
		handleUserCountChange(conn, params)

	case "h5.spin":
		handleH5spin(conn, params)

	case "h5.feature":
		handleH5feature(conn, params)
	case "GEN_HEARTBEAT":
		handleGENHeartbeat(conn, params)
	default:
		fmt.Printf("⚠️ 未知扩展命令: %s\n", cmd)
	}
}
func handleH5feature(conn *websocket.Conn, obj map[string]interface{}) {
	fmt.Printf("obj: %+v\n", obj)

	// 从obj中提取所有参数
	// entity, _ := obj["entity"].(map[string]interface{})
	// featureId, _ := entity["featureId"].(map[string]interface{})

}

func handleGENHeartbeat(conn *websocket.Conn, obj map[string]interface{}) {
	p := map[string]interface{}{
		"p": map[string]interface{}{
			"heartbeat": int64(1747365913855),
		},
		"c": "heartbeat",
	}
	packet := BuildSFSMessage(13, 1, p)
	conn.WriteMessage(websocket.BinaryMessage, packet)
}
func handleH5spin(conn *websocket.Conn, obj map[string]interface{}) {
	// 打印obj对象
	fmt.Printf("obj: %+v\n", obj)

	// 从obj中提取所有参数
	entity, _ := obj["entity"].(map[string]interface{})
	// 内嵌的下注请求结构
	// (utf_string) betType: LineGame
	// 投注类型，此处为“线型游戏”模式（常见于老虎机）
	//(int) betLine: 1
	// 下注的线数，例如下注 1 条赔付线
	// (int) lineBet: 10
	// 每条线上的投注额，例如每线下注 10（单位为 denom）
	betRequest, _ := entity["betRequest"].(map[string]interface{})

	// 从betRequest中提取参数
	betType, _ := betRequest["betType"].(string)          // QuantityGame LineGame // 投注类型，此处为“线型游戏”模式（常见于老虎机）
	quantityBet, _ := betRequest["quantityBet"].(float64) // 1

	// 从entity中提取其他参数
	buyFeatureType, _ := entity["buyFeatureType"]      // null 是否购买特殊功能
	denom, _ := entity["denom"].(float64)              // 10 投注额
	extraBetType, _ := entity["extraBetType"].(string) // NoExtraBet “无额外投注” 模式（NoExtraBet）
	gameStateId, _ := entity["gameStateId"].(float64)  // 0 一般用于同步状态（如开始、进行中、结算）
	playerBet, _ := entity["playerBet"].(float64)      // 20 玩家实际下注总额（一般是 denom × lineBet × betLine）

	fmt.Printf("解析参数: betType=%s, quantityBet=%v, buyFeatureType=%v, denom=%v, extraBetType=%s, gameStateId=%v, playerBet=%v\n",
		betType, quantityBet, buyFeatureType, denom, extraBetType, gameStateId, playerBet)
	//   spinResultStr := "{"spinResult":{"gameStateCount":3,"gameStateResult":[{"gameStateId":0,"currentState":1,"gameStateType":"GS_001","roundCount":0,"stateWin":0},{"gameStateId":1,"currentState":2,"gameStateType":"GS_161","roundCount":1,"roundResult":[{"roundWin":0,"screenResult":{"tableIndex":0,"screenSymbol":[[4,10,10,6,8],[10,10,3,3,9],[9,8,8,5,5],[8,3,3,6,6],[5,6,6,8,8],[10,4,4,9,9]],"dampInfo":[[4,8],[6,9],[9,6],[8,0],[5,10],[10,2]]},"extendGameStateResult":{"screenScatterTwoPositionList":[[[0,0,0,0,0],[0,0,0,0,0],[0,0,0,0,0],[0,0,0,0,0],[0,0,0,0,0],[0,0,0,0,0]]],"screenMultiplier":[],"roundMultiplier":1,"screenWinsInfo":[{"playerWin":0,"quantityWinResult":[],"gameWinType":"QuantityGame"}],"extendWin":0,"gameDescriptor":{"version":1,"cascadeComponent":[[null]]}},"progressResult":{"maxTriggerFlag":true,"stepInfo":{"currentStep":1,"addStep":0,"totalStep":1},"stageInfo":{"currentStage":1,"totalStage":1,"addStage":0},"roundInfo":{"currentRound":1,"totalRound":1,"addRound":0}},"displayResult":{"accumulateWinResult":{"beforeSpinFirstStateOnlyBasePayAccWin":0,"afterSpinFirstStateOnlyBasePayAccWin":0,"beforeSpinAccWin":0,"afterSpinAccWin":0},"readyHandResult":{"displayMethod":[[false],[false],[false],[false],[false],[false]]},"boardDisplayResult":{"winRankType":"Nothing","displayBet":0}},"gameResult":{"playerWin":0,"quantityGameResult":{"playerWin":0,"quantityWinResult":[],"gameWinType":"QuantityGame"},"cascadeEliminateResult":[],"gameWinType":"CascadeGame"}}],"stateWin":0},{"gameStateId":5,"currentState":3,"gameStateType":"GS_002","roundCount":0,"stateWin":0}],"totalWin":0,"boardDisplayResult":{"winRankType":"Nothing","scoreType":"Nothing","displayBet":20},"gameFlowResult":{"IsBoardEndFlag":true,"currentSystemStateId":5,"systemStateIdOptions":[0]}},"ts":1747793347489,"balance":1999.78,"gameSeq":7480749037627}"

	spinResultStr := ""

	switch extraBetType {
	default:
		spinResultStr = GetSpinResult(conn, obj)
	}

	p := map[string]interface{}{
		"p": map[string]interface{}{
			"code":   "spinResponse",
			"entity": []byte(spinResultStr),
		},
		"c": "h5.spinResponse",
	}

	// 发送响应
	packet := BuildSFSMessage(13, 1, p)
	fmt.Println("发送spinResponse")
	conn.WriteMessage(websocket.BinaryMessage, packet)
}

func handleH5Init(conn *websocket.Conn, obj map[string]interface{}) {
	// fmt.Println("解析结果11:", entityStr)
	// entityBytes := StringToUint16Array(entityStr)
	// fmt.Println("解析结果11:", entityBytes)
	entityStr := getGameConfig()
	p := map[string]interface{}{
		"p": map[string]interface{}{
			"code":   "initResponse",
			"entity": []byte(entityStr),
		},
		"c": "h5.initResponse",
	}
	packet := BuildSFSMessage(13, 1, p)
	//打印packet
	fmt.Println("发送initResponse")
	conn.WriteMessage(websocket.BinaryMessage, packet)

}
func handleGameLogin(conn *websocket.Conn, obj map[string]interface{}) {

	p := map[string]interface{}{
		"p": map[string]interface{}{
			"loginRoom": "SLOT_ROOM",
			"data":      true,
			"balance":   2000.0,
			"testMode":  false,
			"serverId":  "04",
			"ts":        int64(1747365908909),
		},
		"c": "gameLoginReturn",
	}

	packet := BuildSFSMessage(13, 1, p)
	conn.WriteMessage(websocket.BinaryMessage, packet)

}
func handleBet(conn *websocket.Conn, obj map[string]interface{}) {
	// bet, _ := params["bet"].(float64)
	// clientSeed, _ := params["clientSeed"].(string)
	// mines, _ := params["minesAmount"].(float64) // 若为 int，则改为 int 类型判断
	p := map[string]interface{}{
		"p": map[string]interface{}{
			"code":       200,
			"winAmount":  0.3,
			"balance":    2999.7,
			"multiplier": 1.0,
			"nextTile":   1.1,
			"currency":   "USD",
			"win":        false,
			"cashout":    0.3,
		},
		"c": "betResponse",
	}

	packet := BuildSFSMessage(13, 1, p)
	conn.WriteMessage(websocket.BinaryMessage, packet)
}
func handleCashout(conn *websocket.Conn, obj map[string]interface{}) {
	p := map[string]interface{}{
		"p": map[string]interface{}{
			"mines": []interface{}{
				map[string]interface{}{
					"columns":    2,
					"rows":       2,
					"cellNumber": 7,
				},
				map[string]interface{}{
					"columns":    4,
					"rows":       2,
					"cellNumber": 9,
				},
				map[string]interface{}{
					"columns":    3,
					"rows":       3,
					"cellNumber": 13,
				},
			},
			"betAmount":      0.3,
			"code":           200,
			"winAmount":      0.0,
			"balance":        3000.0,
			"multiplier":     0.0,
			"coefficientSum": 0.0,
			"currency":       "USD",
			"win":            false,
		},
		"c": "cashOutResponse",
	}

	packet := BuildSFSMessage(13, 1, p)
	conn.WriteMessage(websocket.BinaryMessage, packet)
}
func handleOpenCell(conn *websocket.Conn, obj map[string]interface{}) {
	p := map[string]interface{}{
		"p": map[string]interface{}{
			"code":     200,
			"nextTile": 1.25,
			"win":      true,
			"cashout":  0.33,
			"openCell": map[string]interface{}{
				"columns":    3,
				"rows":       3,
				"cellNumber": 13,
			},
		},
		"c": "openCellResponse",
	}

	packet := BuildSFSMessage(13, 1, p)
	conn.WriteMessage(websocket.BinaryMessage, packet)
}
func handleAutoPlay(conn *websocket.Conn, obj map[string]interface{}) {
	p := map[string]interface{}{
		"p": map[string]interface{}{
			"endGame": map[string]interface{}{
				"mines": []interface{}{
					map[string]interface{}{
						"columns":    1,
						"rows":       5,
						"cellNumber": 21,
					},
					map[string]interface{}{
						"columns":    3,
						"rows":       5,
						"cellNumber": 23,
					},
					map[string]interface{}{
						"columns":    4,
						"rows":       2,
						"cellNumber": 9,
					},
				},
				"betAmount":      0.3,
				"code":           200,
				"winAmount":      0.33,
				"balance":        3000.06,
				"multiplier":     1.1,
				"coefficientSum": 1.1,
				"currency":       "USD",
				"win":            true,
			},
			"code": 200,
		},
		"c": "autoPlayResponse",
	}

	packet := BuildSFSMessage(13, 1, p)
	conn.WriteMessage(websocket.BinaryMessage, packet)
}

func stringToUTF16BEBytes(s string) []byte {
	runes := []rune(s)
	buf := bytes.NewBuffer(nil)
	for _, r := range runes {
		buf.WriteByte(byte(r >> 8))   // 高字节
		buf.WriteByte(byte(r & 0xFF)) // 低字节
	}
	return buf.Bytes()
}

func stringToUtf16Bytes(input string) []byte {
	runes := []rune(input)
	result := make([]byte, len(runes)*2) // 每个 UTF-16 编码占 2 个字节

	for i, r := range runes {
		result[i*2] = byte(r >> 8)     // 高字节
		result[i*2+1] = byte(r & 0xFF) // 低字节
	}
	return result
}
func GetSpinResult(conn *websocket.Conn, obj map[string]interface{}) string {
	playerBet, _ := obj["playerBet"].(float64) //实际下注额
	GS_001 := GameState{
		GameStateId:   0,        // 游戏状态ID
		CurrentState:  1,        // 当前状态
		GameStateType: "GS_001", // 游戏状态类型
		RoundCount:    0,        // 回合数
		StateWin:      0,        // 该状态获得的奖金
	}
	GS_002 := GameState{
		GameStateId:   3,        // 游戏状态ID
		CurrentState:  3,        // 当前状态
		GameStateType: "GS_002", // 游戏状态类型
		RoundCount:    0,        // 回合数
		StateWin:      0,        // 该状态获得的奖金
	}
	///GS_161_1 := getGS161()
	GS_112_1, Special := getGS112(playerBet)
	GS_113 := GameState{}
	gameStateResult := []GameState{}
	systemStateIdOptions := []int{0}
	totalWin := GS_112_1.StateWin
	if len(Special) > 0 {
		fmt.Println("特殊模式")
		GS_113 = getGS113(playerBet)
		GS_002.CurrentState = 4
		totalWin = GS_113.StateWin

		gameStateResult = []GameState{
			GS_001,
			GS_112_1,
			GS_113,
			GS_002,
		}
		systemStateIdOptions = []int{0, 997}
	} else {
		fmt.Println("普通模式")
		gameStateResult = []GameState{
			GS_001,
			GS_112_1,
			GS_002,
		}
	}
	// 构建最顶层对象
	result := SpinResultWrapper{
		TS:      time.Now().UnixMilli(), // 当前时间戳(毫秒)
		Balance: 1999.78,                // 玩家余额
		GameSeq: 7480749037627,          // 游戏序列号
		SpinResult: SpinResult{
			GameStateCount:  len(gameStateResult), // 游戏状态总数
			GameStateResult: gameStateResult,
			TotalWin:        totalWin, // 总奖金
			BoardDisplayResult: BoardDisplay{
				WinRankType: "Nothing",      // 获奖等级类型
				ScoreType:   "Nothing",      // 分数类型
				DisplayBet:  int(playerBet), // 用户本次 spin 投注额
			},
			GameFlowResult: GameFlowResult{
				IsBoardEndFlag:       true,                 // 面板是否结束标志
				CurrentSystemStateId: 3,                    // 当前系统状态ID
				SystemStateIdOptions: systemStateIdOptions, // 系统状态ID选项
			},
		},
	}

	// 转换为 JSON 字符串
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(jsonBytes))
	return string(jsonBytes)
}

func getSpecialFeature(screen [][]int, damps [][]int) []SpecialFeatureResult {
	SpecialFeature := []SpecialFeatureResult{}
	if screen[0][0] == 0 && screen[1][0] == 1 && screen[2][0] == 0 {
		SpecialFeature = []SpecialFeatureResult{
			{
				SpecialHitPattern: "HP_95",
				TriggerEvent:      "Trigger_01",
				SpecialScreenHitData: [][]bool{
					{true},
					{true},
					{true},
				},
				SpecialScreenWin: 0,
			},
		}
	}

	return SpecialFeature
}
func getGS112(playerBet float64) (GameState, []SpecialFeatureResult) {
	screen, dampInfos := GetSreenResult(true)
	ExtendGameResult := getExtendGameStateResult(screen, dampInfos) //判断要不要再次旋转
	SpecialFeature := getSpecialFeature(screen, dampInfos)
	win := 0
	symbol := -1
	win1, symbol1 := GetRoundWin(screen)
	symbol = symbol1
	win = win1
	if ExtendGameResult.ReSpinFlag == true {
		SpecialFeature = getSpecialFeature(ExtendGameResult.ScreenSymbol, ExtendGameResult.DampInfo)
		win2, symbol2 := GetRoundWin(ExtendGameResult.ScreenSymbol)
		win = win2 + win1
		symbol = symbol2
	}

	if len(SpecialFeature) <= 0 {
		fmt.Println("没有特殊模式")
		SpecialFeature = []SpecialFeatureResult{}
	}
	fmt.Println("特殊模式", SpecialFeature)

	lineWin := getLineWinResult(win, symbol, playerBet)

	GS_112 := GameState{
		GameStateId:   1,        // 游戏状态ID
		CurrentState:  2,        // 当前状态
		GameStateType: "GS_112", // 游戏状态类型
		RoundCount:    1,        // 回合数
		RoundResult: []RoundResult{ //每个回合的详情
			{
				RoundWin: win, // 该回合获得的奖金
				ScreenResult: ScreenResult{
					TableIndex:   0,         // 当前使用的赔率表或盘面索引                                               // 表格索引 init 那个状态里的 第0 个表
					ScreenSymbol: screen,    // [][]int{{4, 10, 10, 6, 8}, {10, 10, 3, 3, 9}, {9, 8, 8, 5, 5}, {8, 3, 3, 6, 6}, {5, 6, 6, 8, 8}, {10, 4, 4, 9, 9}}, // 屏幕显示的符号 每列5
					DampInfo:     dampInfos, // // 每列前后的遮罩或滑出符号（如滑动特效）
				},

				ExtendGameStateResult: ExtendGameResult,

				ProgressResult: ProgressResult{
					MaxTriggerFlag: true,                                                   // 是否触发最大进度或终点标志                                                  // 是否达到最大触发次数
					StepInfo:       StepInfo{CurrentStep: 1, AddStep: 0, TotalStep: 1},     // 步骤信息
					StageInfo:      StageInfo{CurrentStage: 1, TotalStage: 1, AddStage: 0}, // 阶段信息
					RoundInfo:      RoundInfo{CurrentRound: 1, TotalRound: 1, AddRound: 0}, // 回合信息
				},
				DisplayResult: DisplayResult{
					AccumulateWinResult: AccumulateWinResult{
						AfterSpinAccWin:                       0, // 累旋转后累积赢分
						AfterSpinFirstStateOnlyBasePayAccWin:  0, // 旋转后首状态仅基本支付累积赢分
						BeforeSpinFirstStateOnlyBasePayAccWin: 0, // 旋转前首状态仅基本支付累积赢分
						BeforeSpinAccWin:                      0, // 旋转前累积赢分

					}, // 累积奖金结果
					ReadyHandResult:    ReadyHandResult{DisplayMethod: [][]bool{{false}, {false}, {false}}}, // 是否显示“听牌”/接近中奖的提示
					BoardDisplayResult: BoardDisplay{WinRankType: "Nothing", DisplayBet: 0},                 // 无任何奖型      // 当前显示投注（本 round）                                 // 面板显示结果
				},
				GameResult: GameResult{
					PlayerWin:     win,        // 玩家获得的奖金
					LineWinResult: lineWin,    // 空的中奖线 slice
					GameWinType:   "LineGame", // 本轮属于线型游戏结果
				},
				SpecialFeatureResult: SpecialFeature,
			},
		},
		StateWin: win, // 该状态获得的奖金
	}
	return GS_112, SpecialFeature
}
func getGS113(playerBet float64) GameState {

	//	screen, dampInfos := getSpecialDataWithDampInfo(true)
	var roundResults []RoundResult
	AfterSpinAccWin := 0
	AfterSpinFirstStateOnlyBasePayAccWin := 0
	BeforeSpinFirstStateOnlyBasePayAccWin := 0
	BeforeSpinAccWin := 0
	for {
		screen, dampInfos := getSpecialDataWithDampInfo(true)
		win, symbol := GetRoundWin(screen)
		ExtendGameResult := getExtendGameStateResult(screen, dampInfos) //判断要不要再次旋转
		lineWin := getLineWinResult(win, symbol, playerBet)
		AfterSpinFirstStateOnlyBasePayAccWin += win
		AfterSpinAccWin += win
		ExtendGameResult.ExtendWin = 0
		ExtendGameResult.TriggerMusicFlag = false
		// 添加当前结果到数组
		roundResults = append(roundResults, RoundResult{
			RoundWin: win, // 该回合获得的奖金
			ScreenResult: ScreenResult{
				TableIndex:   0,         // 当前使用的赔率表或盘面索引                                               // 表格索引 init 那个状态里的 第0 个表
				ScreenSymbol: screen,    // [][]int{{4, 10, 10, 6, 8}, {10, 10, 3, 3, 9}, {9, 8, 8, 5, 5}, {8, 3, 3, 6, 6}, {5, 6, 6, 8, 8}, {10, 4, 4, 9, 9}}, // 屏幕显示的符号 每列5
				DampInfo:     dampInfos, // // 每列前后的遮罩或滑出符号（如滑动特效）
			},

			ExtendGameStateResult: ExtendGameResult,

			ProgressResult: ProgressResult{
				MaxTriggerFlag: false,                                                                                          //todo                                                                                    // 是否触发最大进度或终点标志                                                  // 是否达到最大触发次数
				StepInfo:       StepInfo{CurrentStep: 1, AddStep: 0, TotalStep: 1},                                             // 步骤信息
				StageInfo:      StageInfo{CurrentStage: 1, TotalStage: 1, AddStage: 0},                                         // 阶段信息
				RoundInfo:      RoundInfo{CurrentRound: len(roundResults) + 1, TotalRound: len(roundResults) + 1, AddRound: 1}, // 回合信息
			},
			DisplayResult: DisplayResult{
				AccumulateWinResult: AccumulateWinResult{
					AfterSpinAccWin:                       AfterSpinAccWin,                       // 累旋转后累积赢分
					AfterSpinFirstStateOnlyBasePayAccWin:  AfterSpinFirstStateOnlyBasePayAccWin,  // 旋转后首状态仅基本支付累积赢分
					BeforeSpinFirstStateOnlyBasePayAccWin: BeforeSpinFirstStateOnlyBasePayAccWin, // 旋转前首状态仅基本支付累积赢分
					BeforeSpinAccWin:                      BeforeSpinAccWin,                      // 旋转前累积赢分

				}, // 累积奖金结果
				ReadyHandResult:    ReadyHandResult{DisplayMethod: [][]bool{{false}, {false}, {false}}}, // 是否显示“听牌”/接近中奖的提示
				BoardDisplayResult: BoardDisplay{WinRankType: "Nothing", DisplayBet: 0},                 // 无任何奖型      // 当前显示投注（本 round）                                 // 面板显示结果
			},
			GameResult: GameResult{
				PlayerWin:     win,        // 玩家获得的奖金
				LineWinResult: lineWin,    // 空的中奖线 slice
				GameWinType:   "LineGame", // 本轮属于线型游戏结果
			},
		})
		BeforeSpinFirstStateOnlyBasePayAccWin += win
		BeforeSpinAccWin += win
		// 判断是否满足结束条件：screen[1][0] == 2
		if screen[1][0] == 2 {
			break
		}
	}

	GS_113 := GameState{
		GameStateId:   2,        // 游戏状态ID
		CurrentState:  3,        // 当前状态
		GameStateType: "GS_113", // 游戏状态类型
		RoundCount:    1,        // 回合数
		RoundResult:   roundResults,
		StateWin:      AfterSpinAccWin, // 该状态获得的奖金
	}
	GS_113.RoundCount = len(GS_113.RoundResult)
	return GS_113
}

func getLineWinResult(win int, symbol int, bet float64) []LineWinResult {
	if win <= 0 {
		return []LineWinResult{}
	}

	odds := payTable[symbol][2] // 获取该符号的3连赔率（假设第3个位置是3连）

	result := LineWinResult{
		LineId:         0, // 第0行
		HitDirection:   "LeftToRight",
		IsMixGroupFlag: false,
		HitMixGroup:    -1,
		HitSymbol:      symbol,                   // 命中符号
		HitWay:         3,                        // 命中3格
		HitOdds:        odds,                     // 赔率
		LineWin:        int(float64(odds) * bet), // 奖励 = 赔率 × 下注额
		ScreenHitData: [][]bool{
			{true},
			{true},
			{true},
		},
	}

	return []LineWinResult{result}
}

func getExtendGameStateResult(screen [][]int, damps [][]int) ExtendGameStateResult {
	result := ExtendGameStateResult{
		ReSpinFlag:   false, // 是否触发再旋转
		ReSpinTimes:  0,     // 再旋转次数
		ColumnRecord: 0,     // 列记录
		SquintFlag:   false, // 是否斜视偏移
		ExtendWin:    0,     // 扩展赢分

	}
	if screen[0][0] != 0 || screen[2][0] != 0 {
		return result
	}
	rand := rand.Float64()
	if screen[1][0] == 9 && rand > 0.5 { //50%概率触发再次旋转
		result.ReSpinFlag = true
		result.ReSpinTimes++
		result.ColumnRecord = 2
		result.SquintFlag = true
		screenTwo, dampInfos := ReSpinClumnTwo(false)
		result.ScreenSymbol = [][]int{ // 屏幕显示的符号
			{screen[0][0]}, {screenTwo}, {screen[2][0]}, // 每列一个符号
		}
		result.DampInfo = [][]int{ // 每列对应的阻尼信息
			damps[0], dampInfos, damps[2],
		}
		gamedesc := GameDescriptor{
			Version: 1,
			Component: [][]ComponentItem{
				{
					{
						Type:  "labelWithPlaceholders",
						Value: "getRespinGameWithTimes",
						Placeholders: []Placeholder{
							{
								Type:  "text",
								Value: "1",
							},
						},
					},
				},
			},
		}
		result.GameDescriptor = gamedesc
	} else {
		result.SquintFlag = true
	}
	return result
}
func getGameConfig() string {
	resp := GameSettingResponse{
		MaxBet:                  math.MaxInt64, // 使用 int64 的最大值
		MinBet:                  0,
		DefaultLineBetIdx:       0,
		DefaultBetLineIdx:       0,
		DefaultWaysBetIdx:       -1,
		DefaultWaysBetColumnIdx: -1,
		DefaultConnectBetIdx:    -1,
		DefaultQuantityBetIdx:   -1,
		BetCombinations: map[string]int{
			"10_1_NoExtraBet": 10,
			"1_1_NoExtraBet":  1,
			"2_1_NoExtraBet":  2,
			"4_1_NoExtraBet":  4,
			"5_1_NoExtraBet":  5,
		},
		SingleBetCombinations: map[string]int{
			"10_10_1_NoExtraBet": 10,
			"10_1_1_NoExtraBet":  1,
			"10_2_1_NoExtraBet":  2,
			"10_4_1_NoExtraBet":  4,
			"10_5_1_NoExtraBet":  5,
		},
		GambleLimit:      0,
		GambleTimes:      0,
		GameFeatureCount: 3,
		Denoms:           []int{10},
		DefaultDenomIdx:  0,
		BuyFeature:       true,
		BuyFeatureLimit:  2147483647,

		ExecuteSetting: ExecuteSetting{
			SettingId: "v3_14054_05_01_001",
			BetSpecSetting: BetSpecSetting{
				PaymentType:      "PT_001",
				ExtraBetTypeList: []string{"NoExtraBet"},
				BetSpecification: BetSpecification{
					LineBetList: []int{1, 2, 4, 5, 10},
					BetLineList: []int{1},
					BetType:     "LineGame",
				},
			},
			GameStateSetting: []GameStateSetting{
				{
					GameStateType: "GS_112",
					FrameSetting: FrameSetting{
						ScreenColumn:    3,
						ScreenRow:       1,
						WheelUsePattern: "Dependent",
					},
					TableSetting: TableSetting{
						TableCount:          2,
						TableHitProbability: []int{1, 0},
						WheelData:           WheelData,
					},
					SymbolSetting: SymbolSetting{
						SymbolCount:     10,
						SymbolAttribute: []string{"Wild_01", "FreeGame_01", "FreeGame_02", "M1", "M2", "M3", "M4", "M5", "M6", "M7"},
						PayTable:        payTable,
						MixGroupCount:   0,
						MixGroupSetting: []any{},
					},
					LineSetting: LineSetting{
						MaxBetLine: 1,
						LineTable: [][]int{
							{0, 0, 0},
						},
					},
					GameHitPatternSetting: GameHitPatternSetting{
						GameHitPattern:    "LineGame_LeftToRight",
						MaxEliminateTimes: 0,
					},
					SpecialFeatureSetting: SpecialFeatureSetting{
						SpecialFeatureCount: 1,
						SpecialHitInfo: []SpecialHitInfo{
							{
								SpecialHitPattern: "HP_95",
								TriggerEvent:      "Trigger_01",
								BasePay:           0,
							},
						},
					},
					ProgressSetting: ProgressSetting{
						TriggerLimitType: "RoundLimit",
						StepSetting:      StepSetting{DefaultStep: 1, AddStep: 0, MaxStep: 1},
						StageSetting:     StageSetting{DefaultStage: 1, AddStage: 0, MaxStage: 1},
						RoundSetting:     RoundSetting{DefaultRound: 1, AddRound: 0, MaxRound: 1},
					},
					DisplaySetting: DisplaySetting{
						ReadyHandSetting: ReadyHandSetting{
							ReadyHandLimitType: "NoReadyHandLimit",
							ReadyHandCount:     0,
						},
					},
					ExtendSetting: ExtendSetting{
						InitialChooseTableIndex: 1,
						RespinProbability:       0.311,
						TargetScreen: [][]int{
							{0},
							{9},
							{0},
						},
						RespinFlag:        false,
						RespinTableWeight: []int{1},
						RespinTableChoose: []int{2},
						RespinColumnIndex: 1,
						DampInfoRange:     2,
						EmptySymbolID:     9,
					},
				},
				{
					GameStateType: "GS_113",
					FrameSetting: FrameSetting{
						ScreenColumn:    3,
						ScreenRow:       1,
						WheelUsePattern: "Dependent",
					},
					TableSetting: TableSetting{
						TableCount:          4,
						TableHitProbability: []int{1, 0, 0, 0},
						WheelData: [][]WheelSlot{

							{
								{
									WheelLength: 1,
									NoWinIndex:  []int{0},
									WheelData:   []int{0},
								},
								{
									WheelLength: 20,
									NoWinIndex:  []int{0},
									WheelData:   []int{3, 4, 5, 6, 7, 8, 4, 5, 6, 7, 8, 5, 6, 7, 8, 6, 7, 8, 7, 8},
								},
								{
									WheelLength: 1,
									NoWinIndex:  []int{0},
									WheelData:   []int{0},
								},
							},
							{
								{
									WheelLength: 1,
									NoWinIndex:  []int{0},
									WheelData:   []int{0},
								},
								{
									WheelLength: 25,
									NoWinIndex:  []int{0},
									WheelData:   []int{3, 4, 2, 5, 6, 8, 2, 7, 8, 6, 5, 4, 8, 5, 7, 6, 2, 7, 8, 7, 8, 2, 7, 8, 7},
								},
								{
									WheelLength: 1,
									NoWinIndex:  []int{0},
									WheelData:   []int{0},
								},
							},
							{
								{
									WheelLength: 1,
									NoWinIndex:  []int{0},
									WheelData:   []int{0},
								},
								{
									WheelLength: 62,
									NoWinIndex:  []int{0},
									WheelData:   []int{0, 3, 4, 5, 2, 2, 6, 7, 8, 2, 2, 6, 7, 8, 2, 2, 5, 7, 6, 2, 2, 3, 7, 8, 2, 2, 7, 8, 8, 2, 2, 3, 3, 4, 5, 2, 2, 6, 7, 8, 2, 2, 6, 7, 8, 2, 2, 5, 7, 6, 2, 2, 3, 7, 8, 2, 2, 7, 8, 8, 2, 2},
								},
								{
									WheelLength: 1,
									NoWinIndex:  []int{0},
									WheelData:   []int{0},
								},
							},
							{
								{
									WheelLength: 1,
									NoWinIndex:  []int{0},
									WheelData:   []int{0},
								},
								{
									WheelLength: 62,
									NoWinIndex:  []int{0},
									WheelData:   []int{0, 3, 4, 5, 2, 2, 6, 7, 8, 2, 2, 6, 7, 8, 2, 2, 5, 7, 6, 2, 2, 3, 7, 8, 2, 2, 7, 8, 8, 2, 2, 3, 3, 4, 5, 2, 2, 6, 7, 8, 2, 2, 6, 7, 8, 2, 2, 5, 7, 6, 2, 2, 3, 7, 8, 2, 2, 7, 8, 8, 2, 2},
								},
								{
									WheelLength: 1,
									NoWinIndex:  []int{0},
									WheelData:   []int{0},
								},
							},
						},
					},
					SymbolSetting: SymbolSetting{
						SymbolCount:     10,
						SymbolAttribute: []string{"Wild_01", "FreeGame_01", "FreeGame_02", "M1", "M2", "M3", "M4", "M5", "M6", "M7"},
						PayTable: [][]int{
							{0, 0, 200}, {0, 0, 0}, {0, 0, 0}, {0, 0, 30}, {0, 0, 20},
							{0, 0, 15}, {0, 0, 8}, {0, 0, 5}, {0, 0, 2}, {0, 0, 0},
						},
						MixGroupCount:   0,
						MixGroupSetting: []any{},
					},
					LineSetting: LineSetting{
						MaxBetLine: 1,
						LineTable:  [][]int{{0, 0, 0}},
					},
					GameHitPatternSetting: GameHitPatternSetting{
						GameHitPattern:    "LineGame_LeftToRight",
						MaxEliminateTimes: 0,
					},
					SpecialFeatureSetting: SpecialFeatureSetting{
						SpecialFeatureCount: 1,
						SpecialHitInfo: []SpecialHitInfo{
							{SpecialHitPattern: "HP_96", TriggerEvent: "Trigger_01", BasePay: 0},
						},
					},
					ProgressSetting: ProgressSetting{
						TriggerLimitType: "RoundLimit",
						StepSetting:      StepSetting{DefaultStep: 1, AddStep: 0, MaxStep: 1},
						StageSetting:     StageSetting{DefaultStage: 1, AddStage: 0, MaxStage: 1},
						RoundSetting:     RoundSetting{DefaultRound: 1, AddRound: 1, MaxRound: 20},
					},
					DisplaySetting: DisplaySetting{
						ReadyHandSetting: ReadyHandSetting{
							ReadyHandLimitType: "NoReadyHandLimit",
							ReadyHandCount:     0,
						},
					},
					ExtendSetting: ExtendSetting{
						DampInfoRange:    2,
						EmptySymbolID:    9,
						C2SymbolID:       2,
						DampInfoSymbol:   []int{0, 3, 4, 5, 6, 7, 8},
						RoundLimit:       []int{5, 12, 19, 20},
						ChooseTableIndex: []int{1, 2, 3, 4},
						FowardNRound:     5,
						AllRoundOdds:     50,
						FowardNRoundOdds: 20,
						OddsHitPattern:   []int{5, 5, 15},
					},
				},
			},

			// INSERT_YOUR_REWRITE_HERE
			DoubleGameSetting: DoubleGameSetting{
				DoubleRoundUpperLimit: 5,
				DoubleBetUpperLimit:   1000000000,
				RTP:                   0.96,
				TieRate:               0.1,
			},
			BoardDisplaySetting: BoardDisplaySetting{
				WinRankSetting: WinRankSetting{
					BigWin:   30,
					MegaWin:  75,
					UltraWin: 200,
				},
			},
			GameFlowSetting: GameFlowSetting{
				ConditionTableWithoutBoardEnd: [][]string{
					{"CD_False", "CD_True", "CD_False"},
					{"CD_False", "CD_False", "CD_01"},
					{"CD_False", "CD_False", "CD_False"},
				},
			},
		},
	}

	// 转换为 JSON 字符串
	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(jsonBytes))
	return string(jsonBytes)
}
