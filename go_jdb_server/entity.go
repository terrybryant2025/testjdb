package main

type CashOut struct {
	Code       int     `json:"code"`
	PlayerID   string  `json:"player_id"`
	WinAmount  float64 `json:"winAmount"`
	Multiplier float64 `json:"multiplier"`
	BetID      int     `json:"betId"`
	Currency   string  `json:"currency"`
}

// UpdateCurrentCashOuts represents the response for updating current cashouts
type UpdateCurrentCashOuts struct {
	OpenBetsCount          int       `json:"openBetsCount"`
	Code                   int       `json:"code"`
	CashOuts               []CashOut `json:"cashouts"`
	ActivePlayersCount     int       `json:"activePlayersCount"`
	TotalCashOut           float64   `json:"totalCashOut"`
	TopPlayerProfileImages []string  `json:"topPlayerProfileImages"`
}

// OnlinePlayers represents the online players count response
type OnlinePlayers struct {
	Code          int `json:"code"`
	OnlinePlayers int `json:"onlinePlayers"`
}

// ChangeState represents the state change response
type ChangeState struct {
	NewStateID      int   `json:"newStateId"`
	Code            int   `json:"code"`
	RoundID         int64 `json:"roundId,omitempty"`
	BetStateEndTime int64 `json:"betStateEndTime,omitempty"`
	ServerTime      int64 `json:"serverTime,omitempty"`
	TimeLeft        int64 `json:"timeLeft,omitempty"`
}

// BetRequest represents the bet request
type BetRequest struct {
	Bet         float64 `json:"bet"`
	ClientSeed  string  `json:"clientSeed"`
	BetID       int     `json:"betId"`
	FreeBet     bool    `json:"freeBet"`
	AutoCashOut float64 `json:"autoCashOut"`
}

type CancelBetRequest struct {
	BetID int `json:"betId"`
}

type CancelBetResponse struct {
	BetID    int    `json:"betId"`
	Code     int    `json:"code"`
	PlayerID string `json:"player_id"`
}

// BetResponse represents the bet response
type BetResponse struct {
	Bet          float64 `json:"bet"`
	Code         int     `json:"code"`
	PlayerID     string  `json:"player_id"`
	FreeBet      bool    `json:"freeBet"`
	BetID        int     `json:"betId"`
	ProfileImage string  `json:"profileImage"`
	Username     string  `json:"username"`
}

// NewBalance represents the new balance response
type NewBalance struct {
	Code       int     `json:"code"`
	NewBalance float64 `json:"newBalance"`
}

// CashOutRequest represents the cashout request
type CashOutRequest struct {
	BetID            int   `json:"betId"`
	CurrentTimestamp int64 `json:"currentTimestamp"`
}

type CashOutItem struct {
	BetAmount           float64 `json:"betAmount"`
	WinAmount           float64 `json:"winAmount"`
	PlayerID            string  `json:"player_id"`
	BetID               int     `json:"betId"`
	IsMaxWinAutoCashOut bool    `json:"isMaxWinAutoCashOut"`
}

type CashOutResponse struct {
	Code        int           `json:"code"`
	Cashouts    []CashOutItem `json:"cashouts"`
	Multiplier  float64       `json:"multiplier"`
	OperatorKey string        `json:"operatorKey"`
}

type RoundChartInfo struct {
	Code          int     `json:"code"`
	MaxMultiplier float64 `json:"maxMultiplier"`
	RoundId       int     `json:"roundId"`
}

type Platform struct {
	DeviceInfo string `json:"deviceInfo"`
	UserAgent  string `json:"userAgent"`
	DeviceType string `json:"deviceType"`
}

type RequestItem struct {
	Token        string   `json:"token"`
	Currency     string   `json:"currency"`
	Lang         string   `json:"lang"`
	SessionToken string   `json:"sessionToken"`
	Platform     Platform `json:"platform"`
	Version      string   `json:"version"`
	Jurisdiction string   `json:"jurisdiction"`
}

type LoginReq struct {
	Zn string      `json:"zn"`
	Un string      `json:"un"`
	Pw string      `json:"pw"`
	P  RequestItem `json:"p"`
}

type LoginRsp struct {
	Rs int16           `json:"rs"`
	Zn string          `json:"zn"`
	Un string          `json:"un"`
	Pi int16           `json:"pi"`
	Rl [][]interface{} `json:"rl"` // 或者更具体的 []ResponseItem
	Id int             `json:"id"`
}
type ResponseItem struct {
	Code        int           `json:"0"`
	Type        string        `json:"1"`
	State       string        `json:"2"`
	BoolField3  bool          `json:"3"`
	BoolField4  bool          `json:"4"`
	BoolField5  bool          `json:"5"`
	ShortField6 int16         `json:"6"`
	ShortField7 int16         `json:"7"`
	EmptyArray  []interface{} `json:"8"`
}

type RoundInfo struct {
	Multiplier     float64 `json:"multiplier"`
	RoundStartDate int64   `json:"roundStartDate"` // 毫秒时间戳
	RoundEndDate   int64   `json:"roundEndDate"`   // 毫秒时间戳
	RoundId        int     `json:"roundId"`
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
	UserID       string   `json:"userId"`
	Username     string   `json:"username"`
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

type ChatPromo struct {
	IsEnabled bool `json:"isEnabled"`
}

type ChatRain struct {
	IsEnabled         bool `json:"isEnabled"`
	RainMinBet        int  `json:"rainMinBet"`
	DefaultNumOfUsers int  `json:"defaultNumOfUsers"`
	MinNumOfUsers     int  `json:"minNumOfUsers"`
	MaxNumOfUsers     int  `json:"maxNumOfUsers"`
	RainMaxBet        int  `json:"rainMaxBet"`
}

type Chat struct {
	Promo            ChatPromo `json:"promo"`
	Rain             ChatRain  `json:"rain"`
	IsGifsEnabled    bool      `json:"isGifsEnabled"`
	SendMessageDelay int       `json:"sendMessageDelay"`
	IsEnabled        bool      `json:"isEnabled"`
	MaxMessages      int       `json:"maxMessages"`
	MaxMessageLength int       `json:"maxMessageLength"`
}

type EngagementTools struct {
	IsExternalChatEnabled bool `json:"isExternalChatEnabled"`
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
	BetInputStep                     int             `json:"betInputStep"`
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
	BetOptions                       []int           `json:"betOptions"`
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
	ReturnToPlayer                   float64         `json:"returnToPlayer"`
	IsBalanceValidationEnabled       bool            `json:"isBalanceValidationEnabled"`
	IsHolidayTheme                   bool            `json:"isHolidayTheme"`
	IsGameRulesHaveMultiplierFormula bool            `json:"isGameRulesHaveMultiplierFormula"`
	AccountHistoryActionType         string          `json:"accountHistoryActionType"`
	Chat                             Chat            `json:"chat"`
	IrcDisplayType                   string          `json:"ircDisplayType"`
	GameRulesAutoCashOutType         string          `json:"gameRulesAutoCashOutType"`
}

type LoginInit struct {
	RoundsInfo         []RoundInfo   `json:"roundsInfo"`
	Code               int           `json:"code"`
	ActiveBets         []interface{} `json:"activeBets"`
	OnlinePlayers      int           `json:"onlinePlayers"`
	ActiveFreeBetsInfo []interface{} `json:"activeFreeBetsInfo"`
	User               User          `json:"user"`
	Config             Config        `json:"config"`
	RoundID            int           `json:"roundId"`
	StageID            int           `json:"stageId"`
	CurrentMultiplier  float64       `json:"currentMultiplier"`
}

type CurrentBetsInfo struct {
	BetsCount              int       `json:"betsCount"`
	OpenBetsCount          int       `json:"openBetsCount"`
	Code                   int       `json:"code"`
	CashOuts               []CashOut `json:"cashOuts"`
	ActivePlayersCount     int       `json:"activePlayersCount"`
	Bets                   []Bet     `json:"bets"`
	TopPlayerProfileImages []string  `json:"topPlayerProfileImages"`
	TotalCashOut           float64   `json:"totalCashOut"`
}

type Bet struct {
	Bet          float64 `json:"bet"`
	PlayerID     string  `json:"player_id"`
	BetID        int     `json:"betId"`
	IsFreeBet    bool    `json:"isFreeBet"`
	Currency     string  `json:"currency"`
	ProfileImage string  `json:"profileImage"`
	Username     string  `json:"username"`
	Win          bool    `json:"win"`
	RoundBetId   int     `json:"roundBetId"`
	WinAmount    float64 `json:"winAmount"`
	Payout       float64 `json:"payout"`
}

type UpdateCurrentBets struct {
	BetsCount              int      `json:"betsCount"`
	Code                   int      `json:"code"`
	ActivePlayersCount     int      `json:"activePlayersCount"`
	Bets                   []Bet    `json:"bets"`
	TopPlayerProfileImages []string `json:"topPlayerProfileImages"`
}

type UpdateX struct {
	Code int     `json:"code"`
	X    float64 `json:"x"`
}

type UpdateCrashX struct {
	Code   int     `json:"code"`
	CrashX float64 `json:"crashX"`
	X      float64 `json:"x"`
}

type PreviousRoundInfo struct {
	RoundInfo RoundInfo `json:"roundInfo"`
	Code      int       `json:"code"`
	Bets      []Bet     `json:"bets"`
}
