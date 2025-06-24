package main

type GameConfig struct {
	MaxBet                  int64                  `json:"maxBet"`
	MinBet                  int64                  `json:"minBet"`
	DefaultLineBetIdx       int64                  `json:"defaultLineBetIdx"`
	DefaultBetLineIdx       int64                  `json:"defaultBetLineIdx"`
	DefaultWaysBetIdx       int64                  `json:"defaultWaysBetIdx"`
	DefaultWaysBetColumnIdx int64                  `json:"defaultWaysBetColumnIdx"`
	DefaultConnectBetIdx    int64                  `json:"defaultConnectBetIdx"`
	DefaultQuantityBetIdx   int64                  `json:"defaultQuantityBetIdx"`
	BetCombinations         map[string]interface{} `json:"betCombinations"`
	SingleBetCombinations   map[string]interface{} `json:"singleBetCombinations"`
	GambleLimit             int64                  `json:"gambleLimit"`
	GambleTimes             int64                  `json:"gambleTimes"`
	GameFeatureCount        int64                  `json:"gameFeatureCount"`
	ExecuteSetting          map[string]interface{} `json:"executeSetting"` // 可展开为嵌套结构
	Denoms                  []int64                `json:"denoms"`
	DefaultDenomIdx         int64                  `json:"defaultDenomIdx"`
	BuyFeature              bool                   `json:"buyFeature"`
	BuyFeatureLimit         int64                  `json:"buyFeatureLimit"`
}

// type ExecuteSetting struct {
// 	SettingId              string                   `json:"settingId"`
// 	BetSpecSetting         map[string]interface{}   `json:"betSpecSetting"`   // 可展开为 BetSpecSetting
// 	GameStateSetting       []map[string]interface{} `json:"gameStateSetting"` // 建议生成 GameStateSetting 结构体
// 	DoubleGameSetting      map[string]interface{}   `json:"doubleGameSetting"`
// 	BoardDisplaySetting    map[string]interface{}   `json:"boardDisplaySetting"`
// 	GameFlowSetting        map[string]interface{}   `json:"gameFlowSetting"`
// 	ReiterateSpinCriterion map[string]interface{}   `json:"reiterateSpinCriterion"`
// 	RValue                 map[string]interface{}   `json:"rValue"`
// }

// type BetSpecSetting struct {
// 	PaymentType      string                 `json:"paymentType"`
// 	ExtraBetTypeList []string               `json:"extraBetTypeList"`
// 	BetSpecification map[string]interface{} `json:"betSpecification"` // 可细化为结构体
// 	BuyFeature       map[string]interface{} `json:"buyFeature"`       // 如 {"BuyFeature_01": 75}
// }
// type GameStateSetting struct {
// 	GameStateType         string                 `json:"gameStateType"`
// 	FrameSetting          FrameSetting           `json:"frameSetting"`
// 	TableSetting          TableSetting           `json:"tableSetting"`
// 	SymbolSetting         SymbolSetting          `json:"symbolSetting"`
// 	LineSetting           map[string]interface{} `json:"lineSetting"`
// 	GameHitPatternSetting map[string]interface{} `json:"gameHitPatternSetting"`
// 	SpecialFeatureSetting map[string]interface{} `json:"specialFeatureSetting"`
// 	ProgressSetting       map[string]interface{} `json:"progressSetting"`
// 	DisplaySetting        map[string]interface{} `json:"displaySetting"`
// 	ExtendSetting         map[string]interface{} `json:"extendSetting"`
// }

// type FrameSetting struct {
// 	ScreenColumn    int64  `json:"screenColumn"`
// 	ScreenRow       int64  `json:"screenRow"`
// 	WheelUsePattern string `json:"wheelUsePattern"`
// }

// type TableSetting struct {
// 	TableCount          int64                      `json:"tableCount"`
// 	TableHitProbability []float64                  `json:"tableHitProbability"`
// 	WheelData           [][]map[string]interface{} `json:"wheelData"` // 可进一步结构化
// }

// type SymbolSetting struct {
// 	SymbolCount     int64     `json:"symbolCount"`
// 	SymbolAttribute []string  `json:"symbolAttribute"`
// 	PayTable        [][]int64 `json:"payTable"`
// 	MixGroupCount   int64     `json:"mixGroupCount"`
// }
// type SpecialFeatureSetting struct {
// 	SpecialFeatureCount int64            `json:"specialFeatureCount"`
// 	SpecialHitInfo      []SpecialHitInfo `json:"specialHitInfo"`
// }

// type SpecialHitInfo struct {
// 	SpecialHitPattern string `json:"specialHitPattern"`
// 	TriggerEvent      string `json:"triggerEvent"`
// 	BasePay           int64  `json:"basePay"`
// }

// type ProgressSetting struct {
// 	TriggerLimitType string       `json:"triggerLimitType"`
// 	StepSetting      StepSetting  `json:"stepSetting"`
// 	StageSetting     StageSetting `json:"stageSetting"`
// 	RoundSetting     RoundSetting `json:"roundSetting"`
// }

// type StepSetting struct {
// 	DefaultStep int64 `json:"defaultStep"`
// 	AddStep     int64 `json:"addStep"`
// 	MaxStep     int64 `json:"maxStep"`
// }

// type StageSetting struct {
// 	DefaultStage int64 `json:"defaultStage"`
// 	AddStage     int64 `json:"addStage"`
// 	MaxStage     int64 `json:"maxStage"`
// }

// type RoundSetting struct {
// 	DefaultRound int64 `json:"defaultRound"`
// 	AddRound     int64 `json:"addRound"`
// 	MaxRound     int64 `json:"maxRound"`
// }

// type ExtendSetting struct {
// 	EliminatedMaxTimes           int64                          `json:"eliminatedMaxTimes"`
// 	ScatterC1Id                  int64                          `json:"scatterC1Id"`
// 	ScatterC2Id                  int64                          `json:"scatterC2Id"`
// 	ScatterMultiplier            []int64                        `json:"scatterMultiplier"`
// 	ScatterMultiplierWeight      []int64                        `json:"scatterMultiplierWeight"`
// 	ScatterMultiplierNoHitWeight []int64                        `json:"scatterMultiplierNoHitWeight"`
// 	TriggerRound                 map[string]TriggerRoundSetting `json:"triggerRound"`
// }

// type TriggerRoundSetting struct {
// 	DefaultRound int64 `json:"defaultRound"`
// 	AddRound     int64 `json:"addRound"`
// 	MaxRound     int64 `json:"maxRound"`
// }
// type DisplaySetting struct {
// 	ReadyHandSetting ReadyHandSetting `json:"readyHandSetting"`
// }

// type ReadyHandSetting struct {
// 	ReadyHandLimitType string   `json:"readyHandLimitType"`
// 	ReadyHandCount     int64    `json:"readyHandCount"`
// 	ReadyHandType      []string `json:"readyHandType"`
// }

// type DoubleGameSetting struct {
// 	DoubleRoundUpperLimit int64   `json:"doubleRoundUpperLimit"`
// 	DoubleBetUpperLimit   int64   `json:"doubleBetUpperLimit"`
// 	Rtp                   float64 `json:"rtp"`
// 	TieRate               float64 `json:"tieRate"`
// }

// type GameFlowSetting struct {
// 	ConditionTableWithoutBoardEnd [][]string `json:"conditionTableWithoutBoardEnd"`
// }
// type OddsIntervalSetting struct {
// 	MinOdds    int64   `json:"minOdds"`
// 	MaxOdds    float64 `json:"maxOdds"`
// 	RejectProb float64 `json:"rejectProb"`
// }

// type ReiterateSpinCriterion struct {
// 	OddsIntervalSetting []OddsIntervalSetting `json:"oddsIntervalSetting"`
// }

// type RValue map[string]int64
