package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
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

type RoundResult struct {
	RoundWin             int                   `json:"roundWin"`              // 回合赢分
	ScreenResult         ScreenResult          `json:"screenResult"`          // 屏幕结果
	ExtendGameState      *ExtendGameState      `json:"extendGameStateResult"` // 扩展游戏状态
	SpecialFeatureResult *SpecialFeatureResult `json:"specialFeatureResult"`  // 特殊游戏状态
	ProgressResult       ProgressResult        `json:"progressResult"`        // 进度结果
	DisplayResult        DisplayResult         `json:"displayResult"`         // 显示结果
	GameResult           GameResult            `json:"gameResult"`            // 游戏结果
}

type ScreenResult struct {
	TableIndex   int     `json:"tableIndex"`   // 表格索引
	ScreenSymbol [][]int `json:"screenSymbol"` // 屏幕符号矩阵
	DampInfo     [][]int `json:"dampInfo"`     // 衰减信息
}

type ExtendGameState struct {
	ScreenScatterTwoPositionList [][][]int       `json:"screenScatterTwoPositionList"` // 散布符号2位置列表
	ScreenMultiplier             []interface{}   `json:"screenMultiplier"`             // 屏幕倍数
	RoundMultiplier              int             `json:"roundMultiplier"`              // 回合倍数
	ScreenWinsInfo               []ScreenWinInfo `json:"screenWinsInfo"`               // 屏幕获胜信息
	ExtendWin                    int             `json:"extendWin"`                    // 扩展赢分
	GameDescriptor               GameDescriptor  `json:"gameDescriptor"`               // 游戏描述符

	// 新增
	RoundOdds int `json:"roundOdds"`
}

type ScreenWinInfo struct {
	PlayerWin         int           `json:"playerWin"`         // 玩家赢分
	QuantityWinResult []interface{} `json:"quantityWinResult"` // 数量获胜结果
	GameWinType       string        `json:"gameWinType"`       // 游戏获胜类型
}

type GameDescriptor struct {
	Version          int             `json:"version"`          // 版本号
	CascadeComponent [][]interface{} `json:"cascadeComponent"` // 级联组件

	// 新增
	Component [][]TypVal `json:"component"`
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
	PlayerWin              int                 `json:"playerWin"`              // 玩家赢分
	QuantityGameResult     *QuantityGameResult `json:"quantityGameResult"`     // 数量游戏结果
	WayWinResult           []*WayWinData       `json:"wayWinResult"`           // 滚轮游戏结果
	CascadeEliminateResult []interface{}       `json:"cascadeEliminateResult"` // 级联消除结果
	GameWinType            string              `json:"gameWinType"`            // 游戏获胜类型
}

type QuantityGameResult struct {
	PlayerWin         int           `json:"playerWin"`         // 玩家赢分
	QuantityWinResult []interface{} `json:"quantityWinResult"` // 数量获胜结果
	GameWinType       string        `json:"gameWinType"`       // 游戏获胜类型
}

type GameFlowResult struct {
	IsBoardEndFlag       bool  `json:"IsBoardEndFlag"`       // 面板结束标志
	CurrentSystemStateId int   `json:"currentSystemStateId"` // 当前系统状态ID
	SystemStateIdOptions []int `json:"systemStateIdOptions"` // 系统状态ID选项
}

// 新增

type SpecialFeatureResult struct {
	SpecialHitPattern    string   `json:"specialHitPattern"`
	TriggerEvent         string   `json:"triggerEvent"`
	SpecialScreenHitData [][]bool `json:"specialScreenHitData"`
	SpecialScreenWin     int      `json:"specialScreenWin"`
}

type WayWinData struct {
	SymbolId      int      `json:"symbolId"`      // 命中的符号
	HitDirection  string   `json:"hitDirection"`  // 命中描述
	HitNumber     int      `json:"hitNumber"`     // 滚轮命中数
	HitCount      int      `json:"hitCount"`      // 组合数
	HitOdds       int      `json:"hitOdds"`       // 命中赔率
	SymbolWin     int      `json:"symbolWin"`     // 赢得积分
	ScreenHitData [][]bool `json:"screenHitData"` // 命中分布
}

type TypVal struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

//---------------------------------spin end-------------------------------------

// ---------------------------------init-------------------------------------
type EntitySetting struct {
	MaxBet                  int64                  `json:"maxBet"`
	MinBet                  int                    `json:"minBet"`
	DefaultLineBetIdx       int                    `json:"defaultLineBetIdx"`
	DefaultBetLineIdx       int                    `json:"defaultBetLineIdx"`
	DefaultWaysBetIdx       int                    `json:"defaultWaysBetIdx"`
	DefaultWaysBetColumnIdx int                    `json:"defaultWaysBetColumnIdx"`
	DefaultConnectBetIdx    int                    `json:"defaultConnectBetIdx"`
	DefaultQuantityBetIdx   int                    `json:"defaultQuantityBetIdx"`
	BetCombinations         map[string]int         `json:"betCombinations"`
	SingleBetCombinations   map[string]int         `json:"singleBetCombinations"`
	GambleLimit             int                    `json:"gambleLimit"`
	GambleTimes             int                    `json:"gambleTimes"`
	GameFeatureCount        int                    `json:"gameFeatureCount"`
	ExecuteSetting          map[string]interface{} `json:"executeSetting"`
	Denoms                  []int                  `json:"denoms"`
	DefaultDenomIdx         int                    `json:"defaultDenomIdx"`
	BuyFeature              bool                   `json:"buyFeature"`
	BuyFeatureLimit         int                    `json:"buyFeatureLimit"`
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

	importFromExcel("gameRecord.xlsx")

	r.GET("/websocket", wsHandler)
	r.POST("/frontendAPI.do", reportConfigHandler)
	r.POST("/rum", reportRumHandler)
	r.POST("/batchLog", reportLogHandler)

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
				//"url": "http://abcd.jbp.com/?tpg2tl=1&d=1&isApp=true&gName=TreasureBowl_d65c592&lang=cn&homeUrl=&mute=0&gameType=14&mType=14042&x=e9tkQRED2CBLfk0amYQ8IuupMN-wAZcoIKYGa2P4tE2QLQjl6Qg_MmX6PUb8jPYvi0mYBom7O5nmo0unuC3eivfqG3Y3BIDC",
				// "url": "http://abcd.abcd.com/?tpg2tl=1&d=1&isApp=true&gName=PopPopCandy_096d45b&lang=cn&homeUrl=&mute=0&gameType=14&mType=14087&x=e9tkQRED2CClQmf9gCvFzgwjLyNIEyHpYaWaJUcxXZAYv4XExx8PPCqeD9kNReoH1u1relEAkvZBJu0EJcsF5wKTlotEyTq7",
				//"url": "http://abcd.super.com/?tpg2tl=1&d=1&isApp=true&gName=LuckyDiamond_8dca129&lang=cn&homeUrl=&mute=0&gameType=14&mType=14054&x=e9tkQRED2CBWLS-LUoWn_VDFlws8ozYiyKUUY8aNoIitinh1Hku72QhwxKxXm9gHzAVLCNkc6pWdBRwpN3fwGN1M1OTa9tGh",
				"url": "http://richer.local.com/?tpg2tl=1&d=1&isApp=true&gName=MoneybagsMan_77bfbeb&lang=cn&homeUrl=&mute=0&gameType=14&mType=14047&x=e9tkQRED2CDzvXCEmELwok7aC0pft80PCPH4pX0ueZcp-O6091x8jjqrIY33XP8Yi8f_UJMK5xENkkJT37okjdxreL3Xfxwq",
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
					"gameGroup": []int{
						130, 131, 7, 0, 9, 140, 12, 141, 142, 0, 0, 18, 22, 30, 31,
						160, 32, 161, 162, 50, 55, 56, 57, 58, 59, 60, 66, 67, 75,
						80, 81, 90, 92, 93, 120,
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
					"sessionID": []string{"", "", "", "CD4414C0DB1C78180A701358764DC2753E49C01C5456222A8E4D82530BDAFE4B138D3425C162A2368E5B8E31354B29EC9E51756D2A887EDCDF7BCD82D67FA924877174E55B96280065DFD9F850681C0F70E9E79272A0BC2C455B29E2CE458E2F17A6212C0BD3EC11359573D7B5DC8B84082A3B720E2A821E7FA682E64FC4165BE31D5635679CFD8DECEBFBA665E6DF331878625875AFC8B82342E1068E535C25AF1285F0223E487D34E3873B77607F8FDEA3D1340533A6CB6DD17306B58262BFF15CCBB45CBA640B560B6BF472F0E370EEBC6A14612EDF669C10C84E40236744BC901F0B0E2AA906FB444CF6736D873C88BF6CBDAC6E8C0F29AC7924551BEFD6ACEB645832E623C488455466AE8C2203B476C9331522D85482386EDE3BD45DEB215D83E10E139114CBC4E57A57E6197788A99209EE217131A9C0E9CC41B14244505D6E90CF218329B78AFEED6EE7F8DCB79733E160923EFFC919143A5F26F1DA6F5322C99FB59B39C35F262BDC016F3410E05889F33EE1603B1E3EDC0A88BC09A3B2680270F17A7B98F201C741E6499F7ECE7FBDF1F36EF4F241BE34F6FDAD879EB51496B791603FAC538268EA71E73E00E056684E9BC07A25A1251C8EF556968E0C012E395CFDF395CA6AA045C777A36E963E636BD998409CB0328ECEB740F31905B2DA5F63B51EA8AAAE6FEF0A4CE5AAE59984E5248ED2", ""},

					//"sessionID":   []string{"", "", "", "CD4414C0DB1C7818C6847120848D9E6C20E23EFADFC2C5898E4D82530BDAFE4B2B93DE64ACFE2C4D8E0E5BA7F133CF34F79004B011968D674EDAF5945E9FCDC3CE997A5FAFD16DD2DBD60BA559C97A5E70E9E79272A0BC2C455B29E2CE458E2F17A6212C0BD3EC1138D876E96D26C4AA082A3B720E2A821E389A4FFFA66847E23C877471C2342E9D35D8D1634CFF3595024A411F92BAE456F72C425B183F3D0EB487274164CB29A3AF24C913805B63E3FE86171D722E5D2A31D99815BA487A39215D83E10E139114CBC4E57A57E61977C68B65CD11882AAFDEE14AEE61B82AFDAB0C59EDE90950F3EE618A9EF0F05F24737CB437DCE314DBC919143A5F26F1DA6F5322C99FB59B39C35F262BDC016F3410E05889F33EE1603B1E3EDC0A88BC09A3B2680270F17A7B98F201C741E6499F7ECE7FBDF1F36EF4F241BE34F6FDAD879EB51496B791603F171284C23C767A5C5FCEC67FD57E47C725A1251C8EF55696656A7F8515FF80F6384E3B355747D2C85010EC2C6765AB07F34C85DC54D10E60494FBB363CE4A51BA1547BD53BE1B61AD0FC0F1EEE7DD1F0727DD567F4DF2E13", ""},
					"zone":     "JDB_ZONE_GAME",
					"gsInfo":   "jdb1688.net_443_0",
					"gameType": 14,

					// "machineType": 14042,//聚宝盆
					// "machineType": 14087,//宝宝甜心
					//"machineType": 14054,
					"machineType": 14047, // 富豪哥
					"isRecovery":  false,
					"s0":          "",
					"s1":          "",
					"s2":          "",
					//"s3":          "CD4414C0DB1C78189CC2637019E4862285246E049B693B188E4D82530BDAFE4BBCD64B3F3B290B050DBFDD229E56A5B9B0890C2E7209895748F9A9FB97F406617146A4234BC0F8BD1C0B4CA0D07638E870E9E79272A0BC2C455B29E2CE458E2F17A6212C0BD3EC1138D876E96D26C4AA082A3B720E2A821E7FA682E64FC4165B5605E4BAB81C9BAE3F817B45359E709B024A411F92BAE456D6BEE57909110F951CB02116DBA632719778FCEE2F44CE9ED841AA906A5AF7D51FBFB15D068F7C2BA73D0A351243C59208960186D5F5A711560B6BF472F0E370EEBC6A14612EDF669C10C84E40236744BC901F0B0E2AA906FB444CF6736D873C88BF6CBDAC6E8C0F2DFB42464D959A614FBD3ADB264BFB78BA2D090E951B845E1B7E00FAC008A2642A35E720F44EC6435B338D33B125804DE9CF33A7B42EC506DFD3E7EF27BC1C9903DF678C070092DDBC25C1F5DC0F4D74BCCA48493D656543B27B4BBB4BF5F244E3EBD0515D79CD4223BE42575D8E326D65716B1DD24F07457D20232CF89447A70082593D869179A2FB0C4C9645A3217B3A27BCAD0766F4D588800E5E0B2F695C5A6E0E0B0E5398C25CF054B9D5E0B7F9417BDD97001E8DDE76FE9B7C0098E1192E4121E469AB51135788CE2461CC374C092BD5409AF59386C8AA1562D84244FE2204B067F6EDC535E06480915887C28CED3CA01A2BF058D895DC900678AF487E28D9DE2EEFF07E85F70476E3FD9671FC",
					"s3":       "CD4414C0DB1C78180A701358764DC2753E49C01C5456222A8E4D82530BDAFE4B138D3425C162A2368E5B8E31354B29EC9E51756D2A887EDCDF7BCD82D67FA924877174E55B96280065DFD9F850681C0F70E9E79272A0BC2C455B29E2CE458E2F17A6212C0BD3EC11359573D7B5DC8B84082A3B720E2A821E7FA682E64FC4165BE31D5635679CFD8DECEBFBA665E6DF331878625875AFC8B82342E1068E535C25AF1285F0223E487D34E3873B77607F8FDEA3D1340533A6CB6DD17306B58262BFF15CCBB45CBA640B560B6BF472F0E370EEBC6A14612EDF669C10C84E40236744BC901F0B0E2AA906FB444CF6736D873C88BF6CBDAC6E8C0F29AC7924551BEFD6ACEB645832E623C488455466AE8C2203B476C9331522D85482386EDE3BD45DEB215D83E10E139114CBC4E57A57E6197788A99209EE217131A9C0E9CC41B14244505D6E90CF218329B78AFEED6EE7F8DCB79733E160923EFFC919143A5F26F1DA6F5322C99FB59B39C35F262BDC016F3410E05889F33EE1603B1E3EDC0A88BC09A3B2680270F17A7B98F201C741E6499F7ECE7FBDF1F36EF4F241BE34F6FDAD879EB51496B791603FAC538268EA71E73E00E056684E9BC07A25A1251C8EF556968E0C012E395CFDF395CA6AA045C777A36E963E636BD998409CB0328ECEB740F31905B2DA5F63B51EA8AAAE6FEF0A4CE5AAE59984E5248ED2",
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
	case "13":
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
			// fmt.Println("📥 收到二进制消息:", len(data))
			// // //打印收到的数据
			// fmt.Printf("字节: % x\n", data)
			// 传入 bytes.Reader，跳过前4字节
			reader := bytes.NewReader(data[4:])
			decoded, _ := DecodeSFSObject(reader, data[4:])

			// decoded, consumed := DecodeSFSObject(data2[3:])
			// fmt.Println("📥 consumed:", consumed)
			fmt.Printf("🧩 解码结果: %+v\n", decoded)
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
			fmt.Printf("✅ string: %s\n", string(str))
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

	roomList := []interface{}{
		[]interface{}{2, "SLOT_ROOM", "default", true, false, false, int16(1839), int16(5000), []interface{}{}, int16(0), int16(0)},
		[]interface{}{3, "PUSOYS_LOBBY", "default", false, false, false, int16(21), int16(5000), []interface{}{}},
		[]interface{}{4, "TONGITS_LOBBY", "default", false, false, false, int16(29), int16(5000), []interface{}{}},
		[]interface{}{5, "RUMMY_LOBBY", "default", false, false, false, int16(0), int16(5000), []interface{}{}},
		[]interface{}{6, "RUNNING_GAME", "default", false, false, false, int16(16), int16(5000), []interface{}{}},
		[]interface{}{236, "18020", "default", false, false, false, int16(1), int16(5000), []interface{}{}},
		[]interface{}{237, "18021", "default", false, false, false, int16(0), int16(5000), []interface{}{}},
		[]interface{}{238, "SINGLE_SPIN", "default", false, false, false, int16(8), int16(5000), []interface{}{}},
		[]interface{}{239, "18026", "default", false, false, false, int16(6), int16(5000), []interface{}{}},
		[]interface{}{240, "MINES", "default", false, false, false, int16(91), int16(5000), []interface{}{}},
		[]interface{}{241, "CASINO_ROOM", "default", false, false, false, int16(0), int16(5000), []interface{}{}},
		[]interface{}{242, "18022", "default", false, false, false, int16(0), int16(5000), []interface{}{}},
	}

	// 构造返回数据 map[payload]
	p := map[string]interface{}{
		"rs": int16(0),        // 登录成功
		"zn": "JDB_ZONE_GAME", // 区域名
		"un": obj["un"],       // 用户名
		"pi": int16(0),        // playerId
		"rl": roomList,        // 房间列表
		"id": int32(1928827),  // 用户 ID
	}

	// 构造封包并发送
	packet := BuildSFSMessage(1, 0, p)
	conn.WriteMessage(websocket.BinaryMessage, packet)

	fmt.Println("✅ 已发送 Login 响应")
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
func CallExtensionResponse(conn *websocket.Conn, obj map[string]interface{}) {
	fullGameConfig := map[string]interface{}{
		"displayedAutoCashOutTimer":      int64(10),
		"isActiveGameFocused":            false,
		"isRuleUnfinishedGame":           false,
		"minRoundDurationInMillis":       int64(0),
		"isLoginTimer":                   false,
		"isGameNavigationEnabled":        true,
		"isBetsHistoryEndBalanceEnabled": false,
		"activeGame":                     "mines",
		"minBet":                         0.1,
		"houseEdge":                      3.0,
		"accountHistoryActionType":       "navigate",
		"defaultBetValue":                0.3,
		"isNeedToShowOnLoginModalNotRegulatedByAlderney": false,
		"showPaytableOnStart":                            false,
		"currency":                                       "USD",
		"isBalanceValidationEnabled":                     true,
		"overallAutoCashOutTimer":                        int64(30),
		"gameList": []interface{}{
			"dice", "plinko", "goal", "hi-lo", "mines", "keno",
			"mini-roulette", "hotline", "balloon",
		},
		"maxUserWin":                           10000.0,
		"isCurrencyNameHidden":                 false,
		"fastBets":                             []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 1.2, 2, 4, 10, 20, 50, 100},
		"pingIntervalMs":                       int64(15000),
		"isShowWinAmountUntilNextRound":        false,
		"isShowMultiplierExplanation":          false,
		"isHideFreeBetsInUserMenu":             false,
		"isAutoBetFeatureEnabled":              true,
		"isShowRtp":                            false,
		"operatorHomeButtonFrontEndActionType": "navigate",
		"backToHomeActionType":                 "navigate",
		"isShowLastRoundStateUntilNextRound":   false,
		"betPrecision":                         int32(2),
		"isClockVisible":                       false,
		"isBetsHistoryStartBalanceEnabled":     false,
		"inactivityTimeForDisconnect":          int64(0),
		"isFreeBetDepositEnabled":              false,
		"isMaxWinAm":                           false,
		"maxBet":                               100.0,
		"smallScreenWarning":                   false,
		"autoBetNumberOfRoundsList":            []int32{3, 10, 25, 100, 200, 500},
	}
	p := map[string]interface{}{
		"p": map[string]interface{}{
			"gameConfig": map[string]interface{}{
				"coefficients": map[string]interface{}{
					"1":  []float64{1.01, 1.05, 1.1, 1.15, 1.21, 1.27, 1.34, 1.42, 1.51, 1.61, 1.73, 1.86, 2.02, 2.2, 2.42, 2.69, 3.03, 3.46, 4.04, 4.84, 6.06, 8.08, 12.12, 24.24},
					"2":  []float64{1.05, 1.15, 1.25, 1.38, 1.53, 1.7, 1.9, 2.13, 2.42, 2.77, 3.19, 3.73, 4.4, 5.29, 6.46, 8.08, 10.39, 13.85, 19.4, 29.1, 48.5, 97, 291},
					"3":  []float64{1.1, 1.25, 1.44, 1.67, 1.95, 2.3, 2.73, 3.28, 3.98, 4.9, 6.12, 7.8, 10.14, 13.52, 18.59, 26.55, 39.83, 63.74, 111.55, 223.1, 557.75, 2231},
					"4":  []float64{1.15, 1.38, 1.67, 2.05, 2.53, 3.16, 4, 5.15, 6.74, 8.98, 12.25, 17.16, 24.78, 37.18, 58.43, 97.38, 175.29, 350.58, 818.03, 2454.1, 12270.5},
					"5":  []float64{1.21, 1.53, 1.95, 2.53, 3.32, 4.43, 6.01, 8.32, 11.79, 17.16, 25.74, 40.04, 65.07, 111.54, 204.5, 409.01, 920.28, 2454.09, 8589.34, 51536.09},
					"6":  []float64{1.27, 1.7, 2.3, 3.16, 4.43, 6.33, 9.25, 13.88, 21.45, 34.32, 57.2, 100.1, 185.91, 371.83, 818.03, 2045.08, 6135.25, 24541, 171787},
					"7":  []float64{1.34, 1.9, 2.73, 4, 6.01, 9.25, 14.65, 23.97, 40.75, 72.45, 135.86, 271.72, 588.73, 1412.96, 3885.65, 12952.19, 58284.87, 466278.99},
					"8":  []float64{1.42, 2.13, 3.28, 5.15, 8.32, 13.88, 23.97, 43.15, 81.51, 163.03, 349.35, 815.17, 2119.45, 6358.35, 23313.95, 116569.75, 1049127.75},
					"9":  []float64{1.51, 2.42, 3.98, 6.74, 11.79, 21.45, 40.75, 81.51, 173.22, 395.94, 989.85, 2771.58, 9007.66, 36030.64, 198168.57, 1981685.74},
					"10": []float64{1.61, 2.77, 4.9, 8.98, 17.16, 34.32, 72.45, 163.03, 395.94, 1055.84, 3167.52, 11086.35, 48040.86, 288245.19, 3170697.19},
					"11": []float64{1.73, 3.19, 6.12, 12.25, 25.74, 57.2, 135.86, 349.35, 989.85, 3167.52, 11878.23, 55431.76, 360306.5, 4323678},
					"12": []float64{1.86, 3.73, 7.8, 17.16, 40.04, 100.1, 271.72, 815.17, 2771.58, 11086.35, 55431.76, 388022.38, 5044291},
					"13": []float64{2.02, 4.4, 10.14, 24.78, 65.07, 185.91, 588.73, 2119.44, 9007.66, 48040.86, 360306.49, 5044290.99},
					"14": []float64{2.2, 5.29, 13.52, 37.18, 111.55, 371.83, 1412.96, 6358.35, 36030.65, 288245.2, 4323678},
					"15": []float64{2.42, 6.46, 18.59, 58.43, 204.5, 818.03, 3885.65, 23313.94, 198168.57, 3170697.19},
					"16": []float64{2.69, 8.08, 26.55, 97.38, 409.01, 2045.08, 12952.19, 116569.74, 1981685.74},
					"17": []float64{3.03, 10.39, 39.83, 175.29, 920.28, 6135.25, 58284.87, 1049127.75},
					"18": []float64{3.46, 13.85, 63.74, 350.58, 2454.1, 24541, 466279},
					"19": []float64{4.04, 19.4, 111.55, 818.03, 8589.35, 171787},
					"20": []float64{4.85, 29.1, 223.1, 2454.1, 51536.1},
				},
				"defaultMinesAmount": 3,
			},
			"code":     200,
			"freeBets": []interface{}{},
			"user": map[string]interface{}{
				"settings": map[string]interface{}{
					"music": false,
					"sound": true,
				},
				"balance":  3000.0,
				"avatar":   "av-11.png",
				"userId":   "2008320",
				"username": "demo_75809",
			},
			"config": fullGameConfig, // 需你在代码中另行定义 fullGameConfig 内容

		},
		"c": "init",
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
	//entity, _ := obj["entity"].(map[string]interface{})
	//betRequest, _ := entity["betRequest"].(map[string]interface{})
	//
	//// 从betRequest中提取参数
	//betType, _ := betRequest["betType"].(string)          // QuantityGame
	//quantityBet, _ := betRequest["quantityBet"].(float64) // 1
	//
	//// 从entity中提取其他参数
	//buyFeatureType, _ := entity["buyFeatureType"]      // null
	//denom, _ := entity["denom"].(float64)              // 10
	//extraBetType, _ := entity["extraBetType"].(string) // NoExtraBet
	//gameStateId, _ := entity["gameStateId"].(float64)  // 0
	//playerBet, _ := entity["playerBet"].(float64)      // 20
	//
	//fmt.Printf("解析参数: betType=%s, quantityBet=%v, buyFeatureType=%v, denom=%v, extraBetType=%s, gameStateId=%v, playerBet=%v\n",
	//	betType, quantityBet, buyFeatureType, denom, extraBetType, gameStateId, playerBet)
	//  spinResultStr := "{"spinResult":{"gameStateCount":3,"gameStateResult":[{"gameStateId":0,"currentState":1,"gameStateType":"GS_001","roundCount":0,"stateWin":0},{"gameStateId":1,"currentState":2,"gameStateType":"GS_161","roundCount":1,"roundResult":[{"roundWin":0,"screenResult":{"tableIndex":0,"screenSymbol":[[4,10,10,6,8],[10,10,3,3,9],[9,8,8,5,5],[8,3,3,6,6],[5,6,6,8,8],[10,4,4,9,9]],"dampInfo":[[4,8],[6,9],[9,6],[8,0],[5,10],[10,2]]},"extendGameStateResult":{"screenScatterTwoPositionList":[[[0,0,0,0,0],[0,0,0,0,0],[0,0,0,0,0],[0,0,0,0,0],[0,0,0,0,0],[0,0,0,0,0]]],"screenMultiplier":[],"roundMultiplier":1,"screenWinsInfo":[{"playerWin":0,"quantityWinResult":[],"gameWinType":"QuantityGame"}],"extendWin":0,"gameDescriptor":{"version":1,"cascadeComponent":[[null]]}},"progressResult":{"maxTriggerFlag":true,"stepInfo":{"currentStep":1,"addStep":0,"totalStep":1},"stageInfo":{"currentStage":1,"totalStage":1,"addStage":0},"roundInfo":{"currentRound":1,"totalRound":1,"addRound":0}},"displayResult":{"accumulateWinResult":{"beforeSpinFirstStateOnlyBasePayAccWin":0,"afterSpinFirstStateOnlyBasePayAccWin":0,"beforeSpinAccWin":0,"afterSpinAccWin":0},"readyHandResult":{"displayMethod":[[false],[false],[false],[false],[false],[false]]},"boardDisplayResult":{"winRankType":"Nothing","displayBet":0}},"gameResult":{"playerWin":0,"quantityGameResult":{"playerWin":0,"quantityWinResult":[],"gameWinType":"QuantityGame"},"cascadeEliminateResult":[],"gameWinType":"CascadeGame"}}],"stateWin":0},{"gameStateId":5,"currentState":3,"gameStateType":"GS_002","roundCount":0,"stateWin":0}],"totalWin":0,"boardDisplayResult":{"winRankType":"Nothing","scoreType":"Nothing","displayBet":20},"gameFlowResult":{"IsBoardEndFlag":true,"currentSystemStateId":5,"systemStateIdOptions":[0]}},"ts":1747793347489,"balance":1999.78,"gameSeq":7480749037627}"

	//spinResultStr := GetSpinResult(conn, obj)

	// 从obj中提取所有参数
	entity, _ := obj["entity"].(map[string]interface{})
	betRequest, _ := entity["betRequest"].(map[string]interface{})

	// 从betRequest中提取参数
	//betType, _ := betRequest["betType"].(string)  // QuantityGame
	//betColumn, _ := betRequest["betColumn"].(int) // 1
	wayBet, _ := betRequest["wayBet"].(int32) // 1

	// 从entity中提取其他参数
	//buyFeatureType, _ := entity["buyFeatureType"]      // null
	denom, _ := entity["denom"].(string)               // 10
	extraBetType, _ := entity["extraBetType"].(string) // NoExtraBet
	gameStateId, _ := entity["gameStateId"].(string)   // 0
	playerBet, _ := entity["playerBet"].(string)       // 20

	bet, _ := strconv.Atoi(playerBet)
	gameContext := NewRicherGameContext("test1", 1)
	spinResult, _ := gameContext.Spin(context.Background(), denom, extraBetType, gameStateId, bet, int(wayBet)) // bet/betColumn*10
	spinResultStr, _ := json.Marshal(spinResult)

	p := map[string]interface{}{
		"p": map[string]interface{}{
			"code":   "spinResponse",
			"entity": spinResultStr,
			//"entity": []byte(spinResultStr),
		},
		"c": "h5.spinResponse",
	}

	// 发送响应
	packet := BuildSFSMessage(13, 1, p)
	fmt.Println("发送spinResponse")
	conn.WriteMessage(websocket.BinaryMessage, packet)
}

func handleH5Init(conn *websocket.Conn, obj map[string]interface{}) {
	//entityStr := "{\"maxBet\":9223372036854775807,\"minBet\":0,\"defaultLineBetIdx\":-1,\"defaultBetLineIdx\":-1,\"defaultWaysBetIdx\":-1,\"defaultWaysBetColumnIdx\":-1,\"defaultConnectBetIdx\":-1,\"defaultQuantityBetIdx\":0,\"betCombinations\":{\"10_0_NoExtraBet\":200,\"1_0_NoExtraBet\":20,\"2_0_NoExtraBet\":40,\"3_0_NoExtraBet\":60,\"5_0_NoExtraBet\":100},\"singleBetCombinations\":{\"10_10_0_NoExtraBet\":200,\"10_1_0_NoExtraBet\":20,\"10_2_0_NoExtraBet\":40,\"10_3_0_NoExtraBet\":60,\"10_5_0_NoExtraBet\":100},\"gambleLimit\":0,\"gambleTimes\":0,\"gameFeatureCount\":3,\"executeSetting\":{\"settingId\":\"v3_14087_05_01_201\",\"betSpecSetting\":{\"paymentType\":\"PT_033\",\"extraBetTypeList\":[\"NoExtraBet\"],\"betSpecification\":{\"quantityBetList\":[1,2,3,5,10],\"betType\":\"QuantityGame\"},\"buyFeature\":{\"BuyFeature_01\":75}},\"gameStateSetting\":[{\"gameStateType\":\"GS_161\",\"frameSetting\":{\"screenColumn\":6,\"screenRow\":5,\"wheelUsePattern\":\"PositionDependence\"},\"tableSetting\":{\"tableCount\":2,\"tableHitProbability\":[0.8,0.2],\"wheelData\":[[{\"wheelLength\":100,\"noWinIndex\":[0],\"wheelData\":[8,2,2,7,7,9,4,4,3,3,7,7,4,4,10,10,5,5,10,10,0,6,6,8,8,5,5,3,4,4,9,9,5,5,8,8,4,4,10,10,6,8,8,5,5,10,10,3,3,0,2,2,4,10,10,7,7,2,2,6,6,7,7,8,8,5,5,9,9,10,10,2,2,2,8,8,10,10,3,3,9,9,4,4,10,10,10,9,9,4,4,10,10,2,2,9,9,9,9,8]},{\"wheelLength\":100,\"noWinIndex\":[0],\"wheelData\":[9,9,7,7,7,8,10,10,5,5,7,7,3,3,8,8,5,5,10,10,0,7,7,4,4,7,7,7,4,4,9,9,2,2,10,10,7,7,9,9,8,8,5,5,6,6,5,5,9,9,3,3,8,8,7,7,9,4,4,10,10,0,8,8,5,5,6,6,10,10,3,3,9,9,8,8,4,4,10,10,9,9,5,5,9,9,5,5,8,8,10,10,0,7,7,4,4,9,9,6]},{\"wheelLength\":100,\"noWinIndex\":[0],\"wheelData\":[3,3,10,10,6,7,9,9,5,5,7,7,3,3,10,10,5,5,9,9,0,7,7,4,4,7,7,7,4,4,9,9,2,2,10,10,7,7,9,9,8,8,5,5,6,6,5,5,9,9,3,3,8,8,7,7,9,4,4,10,10,0,8,8,5,5,6,6,10,10,8,8,5,5,6,6,9,9,4,4,8,8,10,10,9,9,5,5,8,8,10,10,2,7,7,4,4,7,7,9]},{\"wheelLength\":100,\"noWinIndex\":[0],\"wheelData\":[7,0,4,4,9,10,10,10,7,7,10,10,2,2,8,8,3,3,6,6,0,10,10,7,7,6,6,6,4,4,7,7,3,3,8,8,5,5,6,6,8,8,3,3,6,6,9,9,4,4,10,10,8,8,9,9,5,5,6,6,6,0,8,8,3,3,10,10,10,8,8,7,7,4,4,9,9,6,6,3,3,10,9,9,7,7,10,10,9,9,4,7,7,10,10,5,5,7,7,10]},{\"wheelLength\":100,\"noWinIndex\":[0],\"wheelData\":[9,9,6,6,6,5,0,9,6,6,10,10,6,6,7,7,10,10,5,5,0,8,8,6,6,8,8,8,4,4,9,9,5,5,9,9,10,10,8,8,7,7,5,5,4,4,8,8,10,10,2,2,3,3,0,4,4,6,6,9,9,0,8,8,5,5,9,9,10,10,4,4,6,6,3,3,10,10,8,8,9,9,10,10,5,5,0,9,9,4,4,10,10,7,7,2,2,6,6,6]},{\"wheelLength\":100,\"noWinIndex\":[0],\"wheelData\":[10,5,5,8,8,0,5,5,2,2,8,8,10,6,6,6,10,10,9,9,0,10,10,3,3,6,6,6,4,4,7,7,3,3,8,8,9,9,6,6,8,8,5,5,6,6,5,5,10,10,9,9,8,8,7,7,2,2,8,8,7,7,0,6,6,5,5,8,8,4,4,3,3,10,10,8,8,9,9,2,2,10,10,4,4,9,9,2,7,7,4,4,4,6,6,5,5,6,6,6]}],[{\"wheelLength\":100,\"noWinIndex\":[0],\"wheelData\":[8,2,2,7,7,9,4,4,3,3,7,7,4,4,10,10,5,5,10,10,0,6,6,8,8,5,5,3,4,4,9,9,5,5,8,8,4,4,10,10,6,8,8,5,5,10,10,3,3,0,2,2,4,10,10,7,7,2,2,6,6,7,7,8,8,5,5,9,9,10,10,2,2,2,8,8,10,10,3,3,9,9,4,4,10,10,10,9,9,4,4,10,10,2,2,9,9,9,9,8]},{\"wheelLength\":100,\"noWinIndex\":[0],\"wheelData\":[9,9,7,7,7,8,10,10,5,5,7,7,3,3,8,8,5,5,10,10,0,7,7,4,4,7,7,7,4,4,9,9,2,2,10,10,7,7,9,9,8,8,5,5,6,6,5,5,9,9,3,3,8,8,7,7,9,4,4,10,10,0,8,8,5,5,6,6,10,10,3,3,9,9,8,8,4,4,10,10,9,9,5,5,9,9,5,5,8,8,10,10,0,7,7,4,4,9,9,6]},{\"wheelLength\":100,\"noWinIndex\":[0],\"wheelData\":[3,3,10,10,6,7,9,9,5,5,7,7,3,3,10,10,5,5,9,9,0,7,7,4,4,7,7,7,4,4,9,9,2,2,10,10,7,7,9,9,8,8,5,5,6,6,5,5,9,9,3,3,8,8,7,7,9,4,4,10,10,0,8,8,5,5,6,6,10,10,8,8,5,5,6,6,9,9,4,4,8,8,10,10,9,9,5,5,8,8,10,10,2,7,7,4,4,7,7,9]},{\"wheelLength\":100,\"noWinIndex\":[0],\"wheelData\":[7,0,4,4,9,10,10,10,7,7,10,10,2,2,8,8,3,3,6,6,0,10,10,7,7,6,6,6,4,4,7,7,3,3,8,8,5,5,6,6,8,8,3,3,6,6,9,9,4,4,10,10,8,8,9,9,5,5,6,6,6,0,8,8,3,3,10,10,10,8,8,7,7,4,4,9,9,6,6,3,3,10,9,9,7,7,10,10,9,9,4,7,7,10,10,5,5,7,7,10]},{\"wheelLength\":100,\"noWinIndex\":[0],\"wheelData\":[9,9,6,6,6,5,0,9,6,6,10,10,6,6,7,7,10,10,5,5,0,8,8,6,6,8,8,8,4,4,9,9,5,5,9,9,10,10,8,8,7,7,5,5,4,4,8,8,10,10,2,2,3,3,0,4,4,6,6,9,9,0,8,8,5,5,9,9,10,10,4,4,6,6,3,3,10,10,8,8,9,9,10,10,5,5,0,9,9,4,4,10,10,7,7,2,2,6,6,6]},{\"wheelLength\":100,\"noWinIndex\":[0],\"wheelData\":[10,5,5,8,8,0,5,5,2,2,8,8,10,6,6,6,10,10,9,9,0,10,10,3,3,6,6,6,4,4,7,7,3,3,8,8,9,9,6,6,8,8,5,5,6,6,5,5,10,10,9,9,8,8,7,7,2,2,8,8,7,7,0,6,6,5,5,8,8,4,4,3,3,10,10,8,8,9,9,2,2,10,10,4,4,9,9,2,7,7,4,4,4,6,6,5,5,6,6,6]}]]},\"symbolSetting\":{\"symbolCount\":11,\"symbolAttribute\":[\"FreeGame_01\",\"BonusGame_01\",\"M1\",\"M2\",\"M3\",\"M4\",\"A\",\"K\",\"Q\",\"J\",\"TE\"],\"payTable\":[[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,200,200,500,500,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000],[0,0,0,0,0,0,0,50,50,200,200,500,500,500,500,500,500,500,500,500,500,500,500,500,500,500,500,500,500,500],[0,0,0,0,0,0,0,40,40,100,100,300,300,300,300,300,300,300,300,300,300,300,300,300,300,300,300,300,300,300],[0,0,0,0,0,0,0,30,30,40,40,240,240,240,240,240,240,240,240,240,240,240,240,240,240,240,240,240,240,240],[0,0,0,0,0,0,0,20,20,30,30,200,200,200,200,200,200,200,200,200,200,200,200,200,200,200,200,200,200,200],[0,0,0,0,0,0,0,16,16,24,24,160,160,160,160,160,160,160,160,160,160,160,160,160,160,160,160,160,160,160],[0,0,0,0,0,0,0,10,10,20,20,100,100,100,100,100,100,100,100,100,100,100,100,100,100,100,100,100,100,100],[0,0,0,0,0,0,0,8,8,18,18,80,80,80,80,80,80,80,80,80,80,80,80,80,80,80,80,80,80,80],[0,0,0,0,0,0,0,5,5,15,15,40,40,40,40,40,40,40,40,40,40,40,40,40,40,40,40,40,40,40]],\"mixGroupCount\":0},\"lineSetting\":{\"maxBetLine\":0},\"gameHitPatternSetting\":{\"gameHitPattern\":\"QuantityGame\",\"maxEliminateTimes\":0},\"specialFeatureSetting\":{\"specialFeatureCount\":3,\"specialHitInfo\":[{\"specialHitPattern\":\"HP_109\",\"triggerEvent\":\"Trigger_01\",\"basePay\":3},{\"specialHitPattern\":\"HP_110\",\"triggerEvent\":\"Trigger_02\",\"basePay\":5},{\"specialHitPattern\":\"HP_124\",\"triggerEvent\":\"Trigger_03\",\"basePay\":100}]},\"progressSetting\":{\"triggerLimitType\":\"NoLimit\",\"stepSetting\":{\"defaultStep\":1,\"addStep\":0,\"maxStep\":1},\"stageSetting\":{\"defaultStage\":1,\"addStage\":0,\"maxStage\":1},\"roundSetting\":{\"defaultRound\":1,\"addRound\":0,\"maxRound\":1}},\"displaySetting\":{\"readyHandSetting\":{\"readyHandLimitType\":\"NoReadyHandLimit\",\"readyHandCount\":1,\"readyHandType\":[\"ReadyHand_34\"]}},\"extendSetting\":{\"eliminatedMaxTimes\":999,\"scatterC1Id\":0,\"scatterC2Id\":1,\"scatterMultiplier\":[2,3,5,8,10,12,15,18,20,25,35,50,100],\"scatterMultiplierWeight\":[100,100,1000,200,120,600,50,30,20,10,5,4,2],\"scatterMultiplierNoHitWeight\":[200,250,300,500,350,200,150,100,80,30,20,4,2],\"triggerRound\":{\"Trigger_01\":{\"defaultRound\":1,\"addRound\":0,\"maxRound\":1},\"Trigger_02\":{\"defaultRound\":1,\"addRound\":0,\"maxRound\":1},\"Trigger_03\":{\"defaultRound\":1,\"addRound\":0,\"maxRound\":1}}}},{\"gameStateType\":\"GS_161\",\"frameSetting\":{\"screenColumn\":6,\"screenRow\":5,\"wheelUsePattern\":\"PositionDependence\"},\"tableSetting\":{\"tableCount\":1,\"tableHitProbability\":[1.0],\"wheelData\":[[{\"wheelLength\":85,\"noWinIndex\":[0],\"wheelData\":[3,3,7,7,9,9,0,6,6,5,5,6,6,10,10,9,9,4,4,10,10,8,8,1,7,7,3,3,10,10,9,9,6,6,2,2,5,5,0,10,10,7,7,9,9,9,5,5,6,6,8,8,3,3,7,7,10,10,9,9,1,5,5,9,9,8,10,10,4,4,4,10,10,3,3,8,8,10,10,2,2,9,9,10,10]},{\"wheelLength\":85,\"noWinIndex\":[0],\"wheelData\":[8,4,4,9,9,0,3,3,10,10,7,7,6,6,9,9,4,4,10,10,9,9,6,6,6,10,10,5,5,7,7,9,9,8,8,5,5,0,10,10,7,7,1,5,5,8,8,2,2,6,6,3,3,8,8,9,9,10,10,1,6,6,9,9,2,2,2,4,4,10,10,10,3,3,8,8,10,10,2,2,9,9,9,9,8]},{\"wheelLength\":85,\"noWinIndex\":[0],\"wheelData\":[8,4,4,7,7,0,7,7,3,3,6,6,8,8,4,4,5,5,7,7,6,6,3,3,3,7,7,6,6,4,4,8,8,9,9,2,2,6,6,5,5,4,4,7,7,3,3,9,9,10,10,1,7,7,9,9,2,2,8,8,6,6,9,9,10,10,1,4,4,5,5,7,7,9,9,3,3,1,8,8,10,10,5,5,8]},{\"wheelLength\":85,\"noWinIndex\":[0],\"wheelData\":[6,5,5,7,7,0,4,4,5,5,9,9,6,6,5,5,5,9,9,8,8,10,10,7,7,8,8,2,2,1,6,6,10,10,9,9,8,8,5,5,10,10,9,9,10,10,8,8,0,9,9,2,2,6,6,3,3,1,7,7,4,4,8,8,10,10,9,9,3,3,10,10,6,6,2,2,5,5,6,6,8,8,9,9,6]},{\"wheelLength\":85,\"noWinIndex\":[0],\"wheelData\":[6,6,0,5,5,10,10,5,5,9,9,6,6,10,10,3,3,7,7,8,8,9,9,10,10,1,8,8,4,4,6,6,2,2,2,0,8,8,3,3,6,6,10,10,2,2,6,4,4,9,9,5,5,7,7,1,6,6,8,8,9,9,5,5,10,10,9,9,8,8,7,7,2,2,6,6,0,5,5,9,9,4,4,10,10]},{\"wheelLength\":85,\"noWinIndex\":[0],\"wheelData\":[0,2,2,9,9,6,6,10,10,1,6,6,5,5,4,4,10,10,8,8,7,7,3,3,1,8,8,10,10,9,9,1,4,4,8,8,5,5,10,10,9,9,7,7,0,8,8,2,2,9,9,4,4,10,10,7,7,7,3,3,6,6,5,5,1,9,9,7,7,4,4,8,8,9,9,3,3,6,6,8,8,5,5,6,6]}]]},\"symbolSetting\":{\"symbolCount\":11,\"symbolAttribute\":[\"FreeGame_01\",\"BonusGame_01\",\"M1\",\"M2\",\"M3\",\"M4\",\"A\",\"K\",\"Q\",\"J\",\"TE\"],\"payTable\":[[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,200,200,500,500,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000],[0,0,0,0,0,0,0,50,50,200,200,500,500,500,500,500,500,500,500,500,500,500,500,500,500,500,500,500,500,500],[0,0,0,0,0,0,0,40,40,100,100,300,300,300,300,300,300,300,300,300,300,300,300,300,300,300,300,300,300,300],[0,0,0,0,0,0,0,30,30,40,40,240,240,240,240,240,240,240,240,240,240,240,240,240,240,240,240,240,240,240],[0,0,0,0,0,0,0,20,20,30,30,200,200,200,200,200,200,200,200,200,200,200,200,200,200,200,200,200,200,200],[0,0,0,0,0,0,0,16,16,24,24,160,160,160,160,160,160,160,160,160,160,160,160,160,160,160,160,160,160,160],[0,0,0,0,0,0,0,10,10,20,20,100,100,100,100,100,100,100,100,100,100,100,100,100,100,100,100,100,100,100],[0,0,0,0,0,0,0,8,8,18,18,80,80,80,80,80,80,80,80,80,80,80,80,80,80,80,80,80,80,80],[0,0,0,0,0,0,0,5,5,15,15,40,40,40,40,40,40,40,40,40,40,40,40,40,40,40,40,40,40,40]],\"mixGroupCount\":0},\"lineSetting\":{\"maxBetLine\":0},\"gameHitPatternSetting\":{\"gameHitPattern\":\"QuantityGame\",\"maxEliminateTimes\":0},\"specialFeatureSetting\":{\"specialFeatureCount\":4,\"specialHitInfo\":[{\"specialHitPattern\":\"HP_108\",\"triggerEvent\":\"ReTrigger_01\",\"basePay\":0},{\"specialHitPattern\":\"HP_109\",\"triggerEvent\":\"ReTrigger_02\",\"basePay\":0},{\"specialHitPattern\":\"HP_110\",\"triggerEvent\":\"ReTrigger_03\",\"basePay\":0},{\"specialHitPattern\":\"HP_124\",\"triggerEvent\":\"ReTrigger_04\",\"basePay\":0}]},\"progressSetting\":{\"triggerLimitType\":\"NoLimit\",\"stepSetting\":{\"defaultStep\":1,\"addStep\":0,\"maxStep\":1},\"stageSetting\":{\"defaultStage\":1,\"addStage\":0,\"maxStage\":1},\"roundSetting\":{\"defaultRound\":10,\"addRound\":5,\"maxRound\":30}},\"displaySetting\":{\"readyHandSetting\":{\"readyHandLimitType\":\"NoReadyHandLimit\",\"readyHandCount\":1,\"readyHandType\":[\"ReadyHand_34\"]}},\"extendSetting\":{\"eliminatedMaxTimes\":999,\"scatterC1Id\":0,\"scatterC2Id\":1,\"scatterMultiplier\":[2,3,5,8,10,12,15,18,20,25,35,50,100],\"scatterMultiplierWeight\":[2100,1500,800,200,100,50,20,10,8,5,3,2,1],\"scatterMultiplierNoHitWeight\":[400,400,500,300,200,100,80,50,30,20,10,4,2],\"triggerRound\":{\"ReTrigger_01\":{\"defaultRound\":10,\"addRound\":5,\"maxRound\":30},\"ReTrigger_02\":{\"defaultRound\":10,\"addRound\":5,\"maxRound\":30},\"ReTrigger_03\":{\"defaultRound\":10,\"addRound\":5,\"maxRound\":30},\"ReTrigger_04\":{\"defaultRound\":10,\"addRound\":5,\"maxRound\":30}}}},{\"gameStateType\":\"GS_161\",\"frameSetting\":{\"screenColumn\":6,\"screenRow\":5,\"wheelUsePattern\":\"FeatureGenerator_01\"},\"tableSetting\":{\"tableCount\":1,\"tableHitProbability\":[1.0],\"wheelData\":[[{\"wheelLength\":100,\"noWinIndex\":[0],\"wheelData\":[8,2,2,7,7,9,4,4,3,3,7,7,4,4,10,10,5,5,10,10,0,6,6,8,8,5,5,3,4,4,9,9,5,5,8,8,4,4,10,10,6,8,8,5,5,10,10,3,3,0,2,2,4,10,10,7,7,2,2,6,6,7,7,8,8,5,5,9,9,10,10,2,2,2,8,8,10,10,3,3,9,9,4,4,10,10,10,9,9,4,4,10,10,2,2,9,9,9,9,8]},{\"wheelLength\":100,\"noWinIndex\":[0],\"wheelData\":[9,9,7,7,7,8,10,10,5,5,7,7,3,3,8,8,5,5,10,10,0,7,7,4,4,7,7,7,4,4,9,9,2,2,10,10,7,7,9,9,8,8,5,5,6,6,5,5,9,9,3,3,8,8,7,7,9,4,4,10,10,0,8,8,5,5,6,6,10,10,3,3,9,9,8,8,4,4,10,10,9,9,5,5,9,9,5,5,8,8,10,10,0,7,7,4,4,9,9,6]},{\"wheelLength\":100,\"noWinIndex\":[0],\"wheelData\":[3,3,10,10,6,7,9,9,5,5,7,7,3,3,10,10,5,5,9,9,0,7,7,4,4,7,7,7,4,4,9,9,2,2,10,10,7,7,9,9,8,8,5,5,6,6,5,5,9,9,3,3,8,8,7,7,9,4,4,10,10,0,8,8,5,5,6,6,10,10,8,8,5,5,6,6,9,9,4,4,8,8,10,10,9,9,5,5,8,8,10,10,2,7,7,4,4,7,7,9]},{\"wheelLength\":100,\"noWinIndex\":[0],\"wheelData\":[7,0,4,4,9,10,10,10,7,7,10,10,2,2,8,8,3,3,6,6,0,10,10,7,7,6,6,6,4,4,7,7,3,3,8,8,5,5,6,6,8,8,3,3,6,6,9,9,4,4,10,10,8,8,9,9,5,5,6,6,6,0,8,8,3,3,10,10,10,8,8,7,7,4,4,9,9,6,6,3,3,10,9,9,7,7,10,10,9,9,4,7,7,10,10,5,5,7,7,10]},{\"wheelLength\":100,\"noWinIndex\":[0],\"wheelData\":[9,9,6,6,6,5,0,9,6,6,10,10,6,6,7,7,10,10,5,5,0,8,8,6,6,8,8,8,4,4,9,9,5,5,9,9,10,10,8,8,7,7,5,5,4,4,8,8,10,10,2,2,3,3,0,4,4,6,6,9,9,0,8,8,5,5,9,9,10,10,4,4,6,6,3,3,10,10,8,8,9,9,10,10,5,5,0,9,9,4,4,10,10,7,7,2,2,6,6,6]},{\"wheelLength\":100,\"noWinIndex\":[0],\"wheelData\":[10,5,5,8,8,0,5,5,2,2,8,8,10,6,6,6,10,10,9,9,0,10,10,3,3,6,6,6,4,4,7,7,3,3,8,8,9,9,6,6,8,8,5,5,6,6,5,5,10,10,9,9,8,8,7,7,2,2,8,8,7,7,0,6,6,5,5,8,8,4,4,3,3,10,10,8,8,9,9,2,2,10,10,4,4,9,9,2,7,7,4,4,4,6,6,5,5,6,6,6]}]],\"screenControlSetting\":[{\"scatterId\":0,\"scatterPatternHitWeight\":[0,0,0,0,10000,77,5],\"scatterTargetColumn\":[0,1,2,3,4,5],\"repeatScatter\":false,\"continuous\":false}]},\"symbolSetting\":{\"symbolCount\":11,\"symbolAttribute\":[\"FreeGame_01\",\"BonusGame_01\",\"M1\",\"M2\",\"M3\",\"M4\",\"A\",\"K\",\"Q\",\"J\",\"TE\"],\"payTable\":[[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,50,50,200,200,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000],[0,0,0,0,0,0,0,30,30,100,100,500,500,500,500,500,500,500,500,500,500,500,500,500,500,500,500,500,500,500],[0,0,0,0,0,0,0,20,20,80,80,300,300,300,300,300,300,300,300,300,300,300,300,300,300,300,300,300,300,300],[0,0,0,0,0,0,0,10,10,30,30,240,240,240,240,240,240,240,240,240,240,240,240,240,240,240,240,240,240,240],[0,0,0,0,0,0,0,8,8,20,20,200,200,200,200,200,200,200,200,200,200,200,200,200,200,200,200,200,200,200],[0,0,0,0,0,0,0,5,5,10,10,160,160,160,160,160,160,160,160,160,160,160,160,160,160,160,160,160,160,160],[0,0,0,0,0,0,0,3,3,8,8,100,100,100,100,100,100,100,100,100,100,100,100,100,100,100,100,100,100,100],[0,0,0,0,0,0,0,2,2,5,5,80,80,80,80,80,80,80,80,80,80,80,80,80,80,80,80,80,80,80],[0,0,0,0,0,0,0,1,1,3,3,40,40,40,40,40,40,40,40,40,40,40,40,40,40,40,40,40,40,40]],\"mixGroupCount\":0},\"lineSetting\":{\"maxBetLine\":0},\"gameHitPatternSetting\":{\"gameHitPattern\":\"QuantityGame\",\"maxEliminateTimes\":0},\"specialFeatureSetting\":{\"specialFeatureCount\":3,\"specialHitInfo\":[{\"specialHitPattern\":\"HP_109\",\"triggerEvent\":\"Trigger_01\",\"basePay\":3},{\"specialHitPattern\":\"HP_110\",\"triggerEvent\":\"Trigger_02\",\"basePay\":5},{\"specialHitPattern\":\"HP_124\",\"triggerEvent\":\"Trigger_03\",\"basePay\":100}]},\"progressSetting\":{\"triggerLimitType\":\"NoLimit\",\"stepSetting\":{\"defaultStep\":1,\"addStep\":0,\"maxStep\":1},\"stageSetting\":{\"defaultStage\":1,\"addStage\":0,\"maxStage\":1},\"roundSetting\":{\"defaultRound\":1,\"addRound\":0,\"maxRound\":1}},\"displaySetting\":{\"readyHandSetting\":{\"readyHandLimitType\":\"NoReadyHandLimit\",\"readyHandCount\":1,\"readyHandType\":[\"ReadyHand_34\"]}},\"extendSetting\":{\"eliminatedMaxTimes\":999,\"scatterC1Id\":0,\"scatterC2Id\":1,\"scatterMultiplier\":[2,3,5,8,10,12,15,18,20,25,35,50,100],\"scatterMultiplierWeight\":[100,100,1000,200,120,600,50,30,20,10,5,4,2],\"scatterMultiplierNoHitWeight\":[200,250,300,500,350,200,150,100,80,30,20,4,2],\"triggerRound\":{\"Trigger_01\":{\"defaultRound\":1,\"addRound\":0,\"maxRound\":1},\"Trigger_02\":{\"defaultRound\":1,\"addRound\":0,\"maxRound\":1},\"Trigger_03\":{\"defaultRound\":1,\"addRound\":0,\"maxRound\":1}},\"noWinScreen\":{\"4\":[[[5,5,0,4,4],[5,10,10,10,0],[3,6,6,5,5],[6,6,0,7,7],[10,10,4,4,6],[0,7,7,3,3]],[[8,5,5,0,9],[5,9,9,0,6],[5,9,9,3,3],[4,4,10,2,2],[6,6,9,9,0],[3,3,10,10,0]],[[0,6,6,8,8],[4,10,10,9,9],[5,5,9,9,0],[6,6,0,10,10],[8,8,4,4,6],[9,0,10,10,3]],[[5,5,0,4,4],[9,9,0,6,6],[5,5,6,6,9],[7,3,3,10,10],[9,0,8,8,5],[3,3,10,10,0]],[[4,10,10,7,7],[9,8,8,10,3],[5,8,10,10,0],[3,10,10,0,7],[6,9,9,0,8],[6,6,0,7,7]]],\"5\":[[[5,5,9,9,10],[0,8,8,5,5],[0,9,9,4,4],[6,6,0,10,10],[0,8,8,6,6],[8,8,7,7,0]],[[0,9,9,10,10],[5,5,6,6,4],[5,9,9,0,6],[0,7,7,5,5],[5,5,6,6,0],[10,10,0,7,7]],[[10,0,6,6,8],[5,5,9,9,3],[4,10,10,0,8],[6,0,8,8,3],[6,6,0,8,8],[8,7,7,0,6]],[[9,9,4,4,0],[10,0,8,8,5],[5,8,10,10,0],[6,6,0,7,7],[7,7,5,5,6],[5,5,6,6,0]]],\"6\":[[[0,6,6,8,8],[4,4,10,10,0],[9,9,0,9,9],[3,3,6,6,0],[6,0,8,8,6],[10,9,9,0,10]],[[10,0,6,6,8],[0,8,8,5,5],[4,4,10,10,0],[3,3,6,6,0],[0,8,8,5,5],[8,7,7,0,6]],[[0,7,7,5,5],[9,0,6,6,4],[10,10,0,8,8],[0,7,7,5,5],[5,6,6,0,7],[6,0,7,7,3]],[[4,4,0,7,7],[5,9,9,0,6],[0,6,6,4,4],[6,0,7,7,3],[6,0,7,7,5],[3,3,10,10,0]]]}}},{\"gameStateType\":\"GS_161\",\"frameSetting\":{\"screenColumn\":6,\"screenRow\":5,\"wheelUsePattern\":\"PositionDependence\"},\"tableSetting\":{\"tableCount\":2,\"tableHitProbability\":[0.7,0.3],\"wheelData\":[[{\"wheelLength\":85,\"noWinIndex\":[0],\"wheelData\":[3,3,7,7,9,9,0,6,6,5,5,6,6,10,10,9,9,4,4,10,10,8,8,1,7,7,3,3,10,10,9,9,6,6,2,2,5,5,0,10,10,7,7,9,9,9,5,5,6,6,8,8,3,3,7,7,10,10,9,9,1,5,5,9,9,8,10,10,4,4,4,10,10,3,3,8,8,10,10,2,2,9,9,10,10]},{\"wheelLength\":85,\"noWinIndex\":[0],\"wheelData\":[8,4,4,9,9,0,3,3,10,10,7,7,6,6,9,9,4,4,10,10,9,9,6,6,6,10,10,5,5,7,7,9,9,8,8,5,5,0,10,10,7,7,1,5,5,8,8,2,2,6,6,3,3,8,8,9,9,10,10,1,6,6,9,9,2,2,2,4,4,10,10,10,3,3,8,8,10,10,2,2,9,9,9,9,8]},{\"wheelLength\":85,\"noWinIndex\":[0],\"wheelData\":[8,4,4,7,7,0,7,7,3,3,6,6,8,8,4,4,5,5,7,7,6,6,3,3,3,7,7,6,6,4,4,8,8,9,9,2,2,6,6,5,5,4,4,7,7,3,3,9,9,10,10,1,7,7,9,9,2,2,8,8,6,6,9,9,10,10,1,4,4,5,5,7,7,9,9,3,3,1,8,8,10,10,5,5,8]},{\"wheelLength\":85,\"noWinIndex\":[0],\"wheelData\":[6,5,5,7,7,0,4,4,5,5,9,9,6,6,5,5,5,9,9,8,8,10,10,7,7,8,8,2,2,1,6,6,10,10,9,9,8,8,5,5,10,10,9,9,10,10,8,8,0,9,9,2,2,6,6,3,3,1,7,7,4,4,8,8,10,10,9,9,3,3,10,10,6,6,2,2,5,5,6,6,8,8,9,9,6]},{\"wheelLength\":85,\"noWinIndex\":[0],\"wheelData\":[6,6,0,5,5,10,10,5,5,9,9,6,6,10,10,3,3,7,7,8,8,9,9,10,10,1,8,8,4,4,6,6,2,2,2,0,8,8,3,3,6,6,10,10,2,2,6,4,4,9,9,5,5,7,7,1,6,6,8,8,9,9,5,5,10,10,9,9,8,8,7,7,2,2,6,6,0,5,5,9,9,4,4,10,10]},{\"wheelLength\":85,\"noWinIndex\":[0],\"wheelData\":[0,2,2,9,9,6,6,10,10,1,6,6,5,5,4,4,10,10,8,8,7,7,3,3,1,8,8,10,10,9,9,1,4,4,8,8,5,5,10,10,9,9,7,7,0,8,8,2,2,9,9,4,4,10,10,7,7,7,3,3,6,6,5,5,1,9,9,7,7,4,4,8,8,9,9,3,3,6,6,8,8,5,5,6,6]}],[{\"wheelLength\":85,\"noWinIndex\":[0],\"wheelData\":[3,3,7,7,9,9,0,6,6,5,5,6,6,10,10,9,9,4,4,10,10,8,8,1,7,7,3,3,10,10,9,9,6,6,2,2,5,5,0,10,10,7,7,9,9,9,5,5,6,6,8,8,3,3,7,7,10,10,9,9,1,5,5,9,9,8,10,10,4,4,4,10,10,3,3,8,8,10,10,2,2,9,9,10,10]},{\"wheelLength\":85,\"noWinIndex\":[0],\"wheelData\":[8,4,4,9,9,0,3,3,10,10,7,7,6,6,9,9,4,4,10,10,9,9,6,6,6,10,10,5,5,7,7,9,9,8,8,5,5,0,10,10,7,7,1,5,5,8,8,2,2,6,6,3,3,8,8,9,9,10,10,1,6,6,9,9,2,2,2,4,4,10,10,10,3,3,8,8,10,10,2,2,9,9,9,9,8]},{\"wheelLength\":85,\"noWinIndex\":[0],\"wheelData\":[8,4,4,7,7,0,7,7,3,3,6,6,8,8,4,4,5,5,7,7,6,6,3,3,3,7,7,6,6,4,4,8,8,9,9,2,2,6,6,5,5,4,4,7,7,3,3,9,9,10,10,1,7,7,9,9,2,2,8,8,6,6,9,9,10,10,1,4,4,5,5,7,7,9,9,3,3,1,8,8,10,10,5,5,8]},{\"wheelLength\":85,\"noWinIndex\":[0],\"wheelData\":[6,5,5,7,7,0,4,4,5,5,9,9,6,6,5,5,5,9,9,8,8,10,10,7,7,8,8,2,2,1,6,6,10,10,9,9,8,8,5,5,10,10,9,9,10,10,8,8,0,9,9,2,2,6,6,3,3,1,7,7,4,4,8,8,10,10,9,9,3,3,10,10,6,6,2,2,5,5,6,6,8,8,9,9,6]},{\"wheelLength\":85,\"noWinIndex\":[0],\"wheelData\":[6,6,0,5,5,10,10,5,5,9,9,6,6,10,10,3,3,7,7,8,8,9,9,10,10,1,8,8,4,4,6,6,2,2,2,0,8,8,3,3,6,6,10,10,2,2,6,4,4,9,9,5,5,7,7,1,6,6,8,8,9,9,5,5,10,10,9,9,8,8,7,7,2,2,6,6,0,5,5,9,9,4,4,10,10]},{\"wheelLength\":85,\"noWinIndex\":[0],\"wheelData\":[0,2,2,9,9,6,6,10,10,1,6,6,5,5,4,4,10,10,8,8,7,7,3,3,1,8,8,10,10,9,9,1,4,4,8,8,5,5,10,10,9,9,7,7,0,8,8,2,2,9,9,4,4,10,10,7,7,7,3,3,6,6,5,5,1,9,9,7,7,4,4,8,8,9,9,3,3,6,6,8,8,5,5,6,6]}]]},\"symbolSetting\":{\"symbolCount\":11,\"symbolAttribute\":[\"FreeGame_01\",\"BonusGame_01\",\"M1\",\"M2\",\"M3\",\"M4\",\"A\",\"K\",\"Q\",\"J\",\"TE\"],\"payTable\":[[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,200,200,500,500,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000,1000],[0,0,0,0,0,0,0,50,50,200,200,500,500,500,500,500,500,500,500,500,500,500,500,500,500,500,500,500,500,500],[0,0,0,0,0,0,0,40,40,100,100,300,300,300,300,300,300,300,300,300,300,300,300,300,300,300,300,300,300,300],[0,0,0,0,0,0,0,30,30,40,40,240,240,240,240,240,240,240,240,240,240,240,240,240,240,240,240,240,240,240],[0,0,0,0,0,0,0,20,20,30,30,200,200,200,200,200,200,200,200,200,200,200,200,200,200,200,200,200,200,200],[0,0,0,0,0,0,0,16,16,24,24,160,160,160,160,160,160,160,160,160,160,160,160,160,160,160,160,160,160,160],[0,0,0,0,0,0,0,10,10,20,20,100,100,100,100,100,100,100,100,100,100,100,100,100,100,100,100,100,100,100],[0,0,0,0,0,0,0,8,8,18,18,80,80,80,80,80,80,80,80,80,80,80,80,80,80,80,80,80,80,80],[0,0,0,0,0,0,0,5,5,15,15,40,40,40,40,40,40,40,40,40,40,40,40,40,40,40,40,40,40,40]],\"mixGroupCount\":0},\"lineSetting\":{\"maxBetLine\":0},\"gameHitPatternSetting\":{\"gameHitPattern\":\"QuantityGame\",\"maxEliminateTimes\":0},\"specialFeatureSetting\":{\"specialFeatureCount\":4,\"specialHitInfo\":[{\"specialHitPattern\":\"HP_108\",\"triggerEvent\":\"ReTrigger_01\",\"basePay\":0},{\"specialHitPattern\":\"HP_109\",\"triggerEvent\":\"ReTrigger_02\",\"basePay\":0},{\"specialHitPattern\":\"HP_110\",\"triggerEvent\":\"ReTrigger_03\",\"basePay\":0},{\"specialHitPattern\":\"HP_124\",\"triggerEvent\":\"ReTrigger_04\",\"basePay\":0}]},\"progressSetting\":{\"triggerLimitType\":\"NoLimit\",\"stepSetting\":{\"defaultStep\":1,\"addStep\":0,\"maxStep\":1},\"stageSetting\":{\"defaultStage\":1,\"addStage\":0,\"maxStage\":1},\"roundSetting\":{\"defaultRound\":10,\"addRound\":5,\"maxRound\":30}},\"displaySetting\":{\"readyHandSetting\":{\"readyHandLimitType\":\"NoReadyHandLimit\",\"readyHandCount\":1,\"readyHandType\":[\"ReadyHand_34\"]}},\"extendSetting\":{\"eliminatedMaxTimes\":999,\"scatterC1Id\":0,\"scatterC2Id\":1,\"scatterMultiplier\":[2,3,5,8,10,12,15,18,20,25,35,50,100],\"scatterMultiplierWeight\":[2000,1500,800,200,100,50,20,10,8,5,3,2,1],\"scatterMultiplierNoHitWeight\":[400,400,500,300,200,100,80,50,30,20,10,4,2],\"triggerRound\":{\"ReTrigger_01\":{\"defaultRound\":10,\"addRound\":5,\"maxRound\":30},\"ReTrigger_02\":{\"defaultRound\":10,\"addRound\":5,\"maxRound\":30},\"ReTrigger_03\":{\"defaultRound\":10,\"addRound\":5,\"maxRound\":30},\"ReTrigger_04\":{\"defaultRound\":10,\"addRound\":5,\"maxRound\":30}}}}],\"doubleGameSetting\":{\"doubleRoundUpperLimit\":5,\"doubleBetUpperLimit\":1000000000,\"rtp\":0.96,\"tieRate\":0.1},\"boardDisplaySetting\":{\"winRankSetting\":{\"BigWin\":10,\"MegaWin\":35,\"UltraWin\":80}},\"gameFlowSetting\":{\"conditionTableWithoutBoardEnd\":[[\"CD_False\",\"CD_38\",\"CD_False\",\"CD_37\",\"CD_False\"],[\"CD_False\",\"CD_False\",\"CD_12\",\"CD_False\",\"CD_False\"],[\"CD_False\",\"CD_False\",\"CD_False\",\"CD_False\",\"CD_False\"],[\"CD_False\",\"CD_False\",\"CD_False\",\"CD_False\",\"CD_12\"],[\"CD_False\",\"CD_False\",\"CD_False\",\"CD_False\",\"CD_False\"]]},\"reiterateSpinCriterion\":{\"oddsIntervalSetting\":[{\"minOdds\":0.0,\"maxOdds\":1.0E-4,\"rejectProb\":0.38},{\"minOdds\":1.0E-4,\"maxOdds\":1.0,\"rejectProb\":0.65},{\"minOdds\":1.0,\"maxOdds\":2.0,\"rejectProb\":0.99},{\"minOdds\":2.0,\"maxOdds\":3.0,\"rejectProb\":0.99},{\"minOdds\":3.0,\"maxOdds\":4.0,\"rejectProb\":0.9},{\"minOdds\":4.0,\"maxOdds\":5.0,\"rejectProb\":0.6},{\"minOdds\":5.0,\"maxOdds\":6.0,\"rejectProb\":0.5},{\"minOdds\":6.0,\"maxOdds\":7.0,\"rejectProb\":0.2},{\"minOdds\":50.0,\"maxOdds\":60.0,\"rejectProb\":0.1},{\"minOdds\":60.0,\"maxOdds\":70.0,\"rejectProb\":0.35},{\"minOdds\":70.0,\"maxOdds\":80.0,\"rejectProb\":0.4},{\"minOdds\":80.0,\"maxOdds\":90.0,\"rejectProb\":0.4},{\"minOdds\":90.0,\"maxOdds\":100.0,\"rejectProb\":0.3},{\"minOdds\":100.0,\"maxOdds\":110.0,\"rejectProb\":0.35},{\"minOdds\":110.0,\"maxOdds\":120.0,\"rejectProb\":0.3},{\"minOdds\":120.0,\"maxOdds\":130.0,\"rejectProb\":0.2},{\"minOdds\":130.0,\"maxOdds\":140.0,\"rejectProb\":0.2},{\"minOdds\":140.0,\"maxOdds\":150.0,\"rejectProb\":0.2},{\"minOdds\":150.0,\"maxOdds\":160.0,\"rejectProb\":0.1}]},\"rValue\":{\"NoExtraBet\":28934,\"BuyFeature_01\":28898}},\"denoms\":[10],\"defaultDenomIdx\":0,\"buyFeature\":true,\"buyFeatureLimit\":2147483647}"
	entityStr := "{\"maxBet\":9223372036854775807,\"defaultWaysBetIdx\":0,\"singleBetCombinations\":{\"10_10_5_NoExtraBet\":500,\"10_1_5_NoExtraBet\":50,\"10_2_5_NoExtraBet\":100,\"10_3_5_NoExtraBet\":150,\"10_5_5_NoExtraBet\":250},\"minBet\":0,\"gambleTimes\":0,\"defaultLineBetIdx\":-1,\"defaultConnectBetIdx\":-1,\"defaultQuantityBetIdx\":-1,\"gameFeatureCount\":4,\"executeSetting\":{\"settingId\":\"v3_14047_05_01_001\",\"betSpecSetting\":{\"paymentType\":\"PT_004\",\"extraBetTypeList\":[\"NoExtraBet\"],\"betSpecification\":{\"wayBetList\":[1,2,3,5,10],\"betColumnList\":[5],\"betType\":\"WayGame\"}},\"gameStateSetting\":[{\"gameStateType\":\"GS_003\",\"frameSetting\":{\"screenColumn\":5,\"screenRow\":3,\"wheelUsePattern\":\"Dependent\"},\"tableSetting\":{\"tableCount\":1,\"tableHitProbability\":[1.0],\"wheelData\":[[{\"wheelLength\":142,\"noWinIndex\":[0],\"wheelData\":[8,7,1,9,9,4,7,4,5,7,3,8,2,2,7,1,6,3,8,4,6,1,3,5,7,3,7,5,3,9,4,1,9,4,1,9,6,6,6,2,7,7,7,1,9,3,2,2,2,9,1,4,9,6,5,5,5,1,6,7,8,3,6,7,1,9,8,6,7,2,8,7,1,9,6,8,9,6,1,8,8,8,2,9,7,6,8,7,1,9,4,7,9,6,5,7,3,8,2,9,7,1,6,3,8,9,6,1,3,7,6,5,9,3,7,1,5,3,6,4,1,9,4,6,7,2,1,7,9,9,3,3,3,1,4,4,4,9,6,5,5,5]},{\"wheelLength\":139,\"noWinIndex\":[0],\"wheelData\":[7,7,7,1,8,8,8,5,5,5,1,7,5,8,3,4,8,1,6,4,7,8,3,7,6,0,3,7,4,5,8,1,5,8,4,7,9,8,2,9,5,7,9,1,7,6,5,7,8,1,7,9,0,0,0,2,2,1,4,4,4,8,7,1,8,5,6,6,5,0,7,5,8,3,3,3,1,8,5,7,8,7,1,5,8,7,5,1,7,5,2,7,8,9,7,1,3,7,5,8,1,7,4,8,1,6,2,2,8,3,7,6,1,5,3,7,4,5,8,1,5,8,4,9,8,2,7,5,1,7,9,1,7,8,5,0,9,9,5]},{\"wheelLength\":142,\"noWinIndex\":[0],\"wheelData\":[4,4,5,1,2,8,8,4,1,6,4,4,6,5,7,8,0,3,7,1,8,3,9,2,8,5,9,1,5,9,6,8,1,7,4,6,7,1,9,4,4,8,8,9,1,3,3,9,8,3,1,8,6,6,4,8,6,4,0,0,0,9,5,5,7,1,9,8,5,7,7,9,2,2,1,8,9,8,4,9,3,1,9,4,9,2,8,4,9,8,4,9,3,8,1,4,9,9,4,1,9,4,4,6,5,7,4,8,3,9,8,3,9,2,8,5,9,1,5,9,6,8,1,2,2,6,7,1,9,4,0,4,8,1,3,9,8,3,1,8,6,6]},{\"wheelLength\":132,\"noWinIndex\":[0],\"wheelData\":[5,5,5,2,1,7,3,4,1,9,4,6,5,4,2,5,5,1,6,5,1,6,3,3,6,8,1,5,5,6,1,3,3,3,1,6,4,4,4,6,3,8,8,8,1,9,9,0,5,7,7,1,3,6,7,5,7,0,5,5,1,6,3,0,6,9,3,6,5,5,5,2,6,6,6,1,7,3,1,4,9,2,5,5,6,1,5,6,3,3,6,0,0,0,4,8,1,6,5,5,6,1,3,3,3,1,2,2,2,6,3,8,8,8,1,9,9,0,5,7,7,7,5,3,6,5,8,6,5,7,3,6]},{\"wheelLength\":121,\"noWinIndex\":[0],\"wheelData\":[7,6,9,1,4,7,2,8,6,4,9,6,1,4,4,6,8,3,7,5,5,1,2,2,2,9,8,8,0,3,6,9,4,4,4,8,6,6,6,1,8,5,5,5,7,7,7,1,8,3,9,1,6,3,8,9,2,2,7,6,1,9,9,9,8,7,6,3,8,1,9,3,6,8,4,2,7,6,9,1,4,7,2,9,1,6,9,3,4,4,0,0,0,3,7,1,3,3,3,9,8,8,1,5,7,7,1,4,4,4,0,8,6,6,5,8,3,9,1,1,1]}]]},\"symbolSetting\":{\"symbolCount\":10,\"symbolAttribute\":[\"Wild_01\",\"FreeGame_01\",\"M1\",\"M2\",\"M3\",\"M4\",\"A\",\"K\",\"Q\",\"J\"],\"payTable\":[[0,0,0,0,0],[0,0,0,0,0],[0,0,75,150,400],[0,0,50,150,300],[0,0,40,100,250],[0,0,30,100,200],[0,0,15,30,125],[0,0,15,30,125],[0,0,10,20,100],[0,0,10,20,100]],\"mixGroupCount\":0,\"mixGroupSetting\":[]},\"gameHitPatternSetting\":{\"gameHitPattern\":\"WayGame_LeftToRight\",\"maxEliminateTimes\":0},\"specialFeatureSetting\":{\"specialFeatureCount\":3,\"specialHitInfo\":[{\"specialHitPattern\":\"HP_88\",\"triggerEvent\":\"Trigger_01\",\"basePay\":0},{\"specialHitPattern\":\"HP_89\",\"triggerEvent\":\"Trigger_02\",\"basePay\":0},{\"specialHitPattern\":\"HP_90\",\"triggerEvent\":\"Trigger_03\",\"basePay\":0}]},\"progressSetting\":{\"triggerLimitType\":\"RoundLimit\",\"stepSetting\":{\"defaultStep\":1,\"addStep\":0,\"maxStep\":1},\"stageSetting\":{\"defaultStage\":1,\"addStage\":0,\"maxStage\":1},\"roundSetting\":{\"defaultRound\":1,\"addRound\":0,\"maxRound\":1}},\"displaySetting\":{\"readyHandSetting\":{\"readyHandLimitType\":\"NoReadyHandLimit\",\"readyHandCount\":1,\"readyHandType\":[\"ReadyHand_27\"]}}},{\"gameStateType\":\"GS_095\",\"frameSetting\":{\"screenColumn\":5,\"screenRow\":3,\"wheelUsePattern\":\"Dependent\"},\"tableSetting\":{\"tableCount\":1,\"tableHitProbability\":[1.0],\"wheelData\":[[{\"wheelLength\":118,\"noWinIndex\":[0],\"wheelData\":[8,7,1,2,7,4,5,7,8,2,5,7,1,6,3,8,4,6,1,3,7,2,7,5,3,9,4,1,9,2,7,6,6,2,9,7,7,1,9,2,3,8,6,2,9,8,1,4,9,6,5,5,1,7,8,3,6,7,1,9,8,6,7,2,5,7,9,6,1,2,6,9,2,6,8,7,1,9,4,7,9,6,5,7,3,8,2,9,7,1,6,3,8,9,6,1,2,7,6,5,9,3,7,1,5,3,6,4,1,9,4,6,7,3,1,4,4,4]},{\"wheelLength\":115,\"noWinIndex\":[0],\"wheelData\":[7,7,7,1,8,4,0,8,6,1,4,7,8,4,7,3,5,8,1,5,8,4,7,9,8,2,0,6,7,9,1,7,9,5,7,8,1,5,9,7,2,2,1,4,4,4,8,7,1,8,5,6,3,7,6,5,0,7,8,3,7,1,3,8,5,6,8,7,1,5,8,7,5,1,2,7,8,9,7,1,3,7,5,8,7,4,8,1,6,2,2,8,3,7,6,1,5,7,4,5,8,1,4,9,8,2,1,7,9,0,8,7,1,5,9]},{\"wheelLength\":115,\"noWinIndex\":[0],\"wheelData\":[4,4,5,2,8,8,4,1,6,7,4,5,6,7,8,0,3,7,1,8,3,9,2,8,5,7,1,5,7,6,8,1,4,6,7,1,9,8,0,3,9,1,3,9,8,3,1,8,6,6,0,7,6,4,8,9,5,4,1,8,7,2,8,7,9,5,2,1,9,8,3,9,8,4,9,8,4,9,3,8,1,4,9,9,4,1,7,6,5,7,8,3,9,8,3,6,8,5,9,1,5,9,6,7,2,1,9,6,4,8,1,3,9,8,3]},{\"wheelLength\":110,\"noWinIndex\":[0],\"wheelData\":[5,5,5,2,1,7,3,5,1,9,4,8,5,4,2,7,6,5,9,1,3,6,8,1,5,7,6,1,3,3,3,4,6,1,2,3,6,4,8,6,1,7,9,6,5,7,7,1,3,6,7,5,7,6,3,6,9,3,6,0,2,6,6,1,6,7,3,1,9,5,6,6,2,5,8,6,1,5,6,8,5,6,2,2,2,6,8,3,7,8,1,9,6,0,5,7,7,8,5,3,9,8,6,5,1,7,3,6,8,9]},{\"wheelLength\":100,\"noWinIndex\":[0],\"wheelData\":[7,6,9,1,4,7,2,8,6,4,9,6,1,4,8,6,4,5,3,7,9,1,2,2,2,9,8,3,6,9,4,4,4,8,6,6,6,5,1,4,8,7,5,1,8,3,9,1,3,8,9,2,6,9,2,1,7,9,3,8,9,0,3,9,6,8,4,9,6,1,4,7,2,8,9,1,4,9,6,4,3,9,7,1,2,3,7,9,1,3,3,3,8,8,5,4,8,9,1,3]}]]},\"symbolSetting\":{\"symbolCount\":10,\"symbolAttribute\":[\"Wild_01\",\"FreeGame_01\",\"M1\",\"M2\",\"M3\",\"M4\",\"A\",\"K\",\"Q\",\"J\"],\"payTable\":[[0,0,0,0,0],[0,0,0,0,0],[0,0,75,150,400],[0,0,50,150,300],[0,0,40,100,250],[0,0,30,100,200],[0,0,15,30,125],[0,0,15,30,125],[0,0,10,20,100],[0,0,10,20,100]],\"mixGroupCount\":0,\"mixGroupSetting\":[]},\"gameHitPatternSetting\":{\"gameHitPattern\":\"WayGame_LeftToRight\",\"maxEliminateTimes\":0},\"specialFeatureSetting\":{\"specialFeatureCount\":3,\"specialHitInfo\":[{\"specialHitPattern\":\"HP_88\",\"triggerEvent\":\"ReTrigger_01\",\"basePay\":0},{\"specialHitPattern\":\"HP_89\",\"triggerEvent\":\"ReTrigger_01\",\"basePay\":0},{\"specialHitPattern\":\"HP_90\",\"triggerEvent\":\"ReTrigger_01\",\"basePay\":0}]},\"progressSetting\":{\"triggerLimitType\":\"RoundLimit\",\"stepSetting\":{\"defaultStep\":1,\"addStep\":0,\"maxStep\":1},\"stageSetting\":{\"defaultStage\":1,\"addStage\":0,\"maxStage\":1},\"roundSetting\":{\"defaultRound\":12,\"addRound\":12,\"maxRound\":50}},\"displaySetting\":{\"readyHandSetting\":{\"readyHandLimitType\":\"NoReadyHandLimit\",\"readyHandCount\":1,\"readyHandType\":[\"ReadyHand_27\"]}},\"extendSetting\":{\"oddsRadix\":1,\"oddsAddition\":1,\"triggerOddsRadix\":{\"Trigger_01\":1,\"Trigger_02\":2,\"Trigger_03\":3},\"freeGameFlag\":false,\"buyFeatureFlag\":false}}],\"doubleGameSetting\":{\"doubleRoundUpperLimit\":5,\"doubleBetUpperLimit\":1000000000,\"rtp\":0.96,\"tieRate\":0.1},\"boardDisplaySetting\":{\"winRankSetting\":{\"BigWin\":30,\"MegaWin\":220,\"UltraWin\":350}},\"gameFlowSetting\":{\"conditionTableWithoutBoardEnd\":[[\"CD_False\",\"CD_True\",\"CD_False\"],[\"CD_False\",\"CD_False\",\"CD_15\"],[\"CD_False\",\"CD_False\",\"CD_False\"]]},\"rValue\":{\"NoExtraBet\":28859}},\"denoms\":[10],\"defaultDenomIdx\":0,\"defaultBetLineIdx\":-1,\"betCombinations\":{\"10_5_NoExtraBet\":500,\"1_5_NoExtraBet\":50,\"2_5_NoExtraBet\":100,\"3_5_NoExtraBet\":150,\"5_5_NoExtraBet\":250},\"gambleLimit\":0,\"buyFeatureLimit\":2147483647,\"buyFeature\":true,\"defaultWaysBetColumnIdx\":0}"
	// fmt.Println("解析结果11:", entityStr)
	// entityBytes := StringToUint16Array(entityStr)
	// fmt.Println("解析结果11:", entityBytes)
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
			"ts":        time.Now().UnixMilli(),
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
	GS_001 := GameState{
		GameStateId:   0,        // 游戏状态ID
		CurrentState:  1,        // 当前状态
		GameStateType: "GS_001", // 游戏状态类型
		RoundCount:    0,        // 回合数
		StateWin:      0,        // 该状态获得的奖金
	}
	GS_002 := GameState{
		GameStateId:   5,        // 游戏状态ID
		CurrentState:  3,        // 当前状态
		GameStateType: "GS_002", // 游戏状态类型
		RoundCount:    0,        // 回合数
		StateWin:      0,        // 该状态获得的奖金
	}
	GS_161_1 := getGS161()

	// 构建最顶层对象
	result := SpinResultWrapper{
		TS:      time.Now().UnixMilli(), // 当前时间戳(毫秒)
		Balance: 1999.78,                // 玩家余额
		GameSeq: 7480749037627,          // 游戏序列号
		SpinResult: SpinResult{
			GameStateCount: 3, // 游戏状态总数
			GameStateResult: []GameState{
				GS_001,
				GS_161_1,
				GS_002,
			},
			TotalWin: 0, // 总奖金
			BoardDisplayResult: BoardDisplay{
				WinRankType: "Nothing", // 获奖等级类型
				ScoreType:   "Nothing", // 分数类型
				DisplayBet:  20,        // 显示的投注额
			},
			GameFlowResult: GameFlowResult{
				IsBoardEndFlag:       true,     // 面板是否结束标志
				CurrentSystemStateId: 5,        // 当前系统状态ID
				SystemStateIdOptions: []int{0}, // 系统状态ID选项
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

func getGS161() GameState {
	GS_161 := GameState{
		GameStateId:   1,        // 游戏状态ID
		CurrentState:  2,        // 当前状态
		GameStateType: "GS_161", // 游戏状态类型
		RoundCount:    1,        // 回合数
		RoundResult: []RoundResult{ //每个回合的详情
			{
				RoundWin: 0, // 该回合获得的奖金
				ScreenResult: ScreenResult{
					TableIndex:   1,                                                         // 表格索引 init 那个状态里的 第0 个表
					ScreenSymbol: getRandomData(false),                                      // [][]int{{4, 10, 10, 6, 8}, {10, 10, 3, 3, 9}, {9, 8, 8, 5, 5}, {8, 3, 3, 6, 6}, {5, 6, 6, 8, 8}, {10, 4, 4, 9, 9}}, // 屏幕显示的符号 每列5
					DampInfo:     [][]int{{4, 8}, {6, 9}, {9, 6}, {8, 0}, {5, 10}, {10, 2}}, // 信息
				},
				ExtendGameState: &ExtendGameState{
					ScreenMultiplier: []interface{}{},
					ScreenScatterTwoPositionList: [][][]int{
						{
							{0, 0, 0, 0, 0},
							{0, 0, 0, 0, 0},
							{0, 0, 0, 0, 0},
							{0, 0, 0, 0, 0},
							{0, 0, 0, 0, 0},
							{0, 0, 0, 0, 0},
						},
					}, // Scatter符号位置列表
					RoundMultiplier: 1, // 回合倍数
					ScreenWinsInfo: []ScreenWinInfo{
						{PlayerWin: 0, GameWinType: "QuantityGame", QuantityWinResult: []interface{}{}}, // 屏幕获奖信息
					},
					ExtendWin: 0, // 额外奖金
					GameDescriptor: GameDescriptor{
						Version:          1,                      // 版本号
						CascadeComponent: [][]interface{}{{nil}}, // 消除组件
					},
				},
				ProgressResult: ProgressResult{
					MaxTriggerFlag: true,                                                   // 是否达到最大触发次数
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
					ReadyHandResult:    ReadyHandResult{DisplayMethod: [][]bool{{false}, {false}, {false}, {false}, {false}, {false}}}, // 准备手牌结果
					BoardDisplayResult: BoardDisplay{WinRankType: "Nothing", DisplayBet: 0},                                            // 面板显示结果
				},
				GameResult: GameResult{
					CascadeEliminateResult: []interface{}{}, //掉落信息
					PlayerWin:              0,               // 玩家获得的奖金
					QuantityGameResult: &QuantityGameResult{
						PlayerWin:         0,              // 数量游戏获得的奖金
						GameWinType:       "QuantityGame", // 游戏获奖类型
						QuantityWinResult: []interface{}{},
					},
					GameWinType: "CascadeGame", // 游戏类型
				},
			},
		},
		StateWin: 0, // 该状态获得的奖金
	}
	return GS_161
}

func getGameConfig() string {
	// 构建最顶层对象
	result := EntitySetting{
		MaxBet:                  math.MaxInt64, // 使用 int64 的最大值
		MinBet:                  0,
		DefaultLineBetIdx:       -1,
		DefaultBetLineIdx:       -1,
		DefaultWaysBetIdx:       -1,
		DefaultWaysBetColumnIdx: -1,
		DefaultConnectBetIdx:    -1,
		DefaultQuantityBetIdx:   0,
		BetCombinations: map[string]int{
			"10_0_NoExtraBet": 200,
			"1_0_NoExtraBet":  20,
			"2_0_NoExtraBet":  40,
			"3_0_NoExtraBet":  60,
			"5_0_NoExtraBet":  100,
		},
		SingleBetCombinations: map[string]int{
			"10_10_0_NoExtraBet": 200,
			"10_1_0_NoExtraBet":  20,
			"10_2_0_NoExtraBet":  40,
			"10_3_0_NoExtraBet":  60,
			"10_5_0_NoExtraBet":  100,
		},
		GambleLimit:      0,
		GambleTimes:      0,
		GameFeatureCount: 3,
		ExecuteSetting: map[string]interface{}{
			"settingId": "v3_14087_05_01_201",
			"betSpecSetting": map[string]interface{}{
				"paymentType":      "PT_033",
				"extraBetTypeList": []string{"NoExtraBet"},
				"betSpecification": map[string]interface{}{
					"quantityBetList": []int{1, 2, 3, 5, 10},
					"betType":         "QuantityGame",
				},
				"buyFeature": map[string]interface{}{
					"BuyFeature_01": 75,
				},
			},
			"gameStateSetting": []map[string]interface{}{
				map[string]interface{}{
					"gameStateType": "GS_161",
					"frameSetting": map[string]interface{}{
						"screenColumn":    6,
						"screenRow":       5,
						"wheelUsePattern": "PositionDependence",
					},
					"tableSetting": map[string]interface{}{
						"tableCount":          2,
						"tableHitProbability": []float64{0.8, 0.2}, // 可填 []float64{0.8, 0.2}
						"wheelData": [][]map[string]interface{}{
							{
								map[string]interface{}{
									"wheelLength": 100,
									"noWinIndex":  []int{0},
									"wheelData":   []int{8, 2, 2, 7, 7, 9, 4, 4, 3, 3, 7, 7, 4, 4, 10, 10, 5, 5, 10, 10, 0, 6, 6, 8, 8, 5, 5, 3, 4, 4, 9, 9, 5, 5, 8, 8, 4, 4, 10, 10, 6, 8, 8, 5, 5, 10, 10, 3, 3, 0, 2, 2, 4, 10, 10, 7, 7, 2, 2, 6, 6, 7, 7, 8, 8, 5, 5, 9, 9, 10, 10, 2, 2, 2, 8, 8, 10, 10, 3, 3, 9, 9, 4, 4, 10, 10, 10, 9, 9, 4, 4, 10, 10, 2, 2, 9, 9, 9, 9, 8}, // 数据较长，省略内容
								},
								map[string]interface{}{
									"wheelLength": 100,
									"noWinIndex":  []int{0},
									"wheelData":   []int{9, 9, 7, 7, 7, 8, 10, 10, 5, 5, 7, 7, 3, 3, 8, 8, 5, 5, 10, 10, 0, 7, 7, 4, 4, 7, 7, 7, 4, 4, 9, 9, 2, 2, 10, 10, 7, 7, 9, 9, 8, 8, 5, 5, 6, 6, 5, 5, 9, 9, 3, 3, 8, 8, 7, 7, 9, 4, 4, 10, 10, 0, 8, 8, 5, 5, 6, 6, 10, 10, 3, 3, 9, 9, 8, 8, 4, 4, 10, 10, 9, 9, 5, 5, 9, 9, 5, 5, 8, 8, 10, 10, 0, 7, 7, 4, 4, 9, 9, 6}, // 数据较长，省略内容
								},
								map[string]interface{}{
									"wheelLength": 100,
									"noWinIndex":  []int{0},
									"wheelData":   []int{3, 3, 10, 10, 6, 7, 9, 9, 5, 5, 7, 7, 3, 3, 10, 10, 5, 5, 9, 9, 0, 7, 7, 4, 4, 7, 7, 7, 4, 4, 9, 9, 2, 2, 10, 10, 7, 7, 9, 9, 8, 8, 5, 5, 6, 6, 5, 5, 9, 9, 3, 3, 8, 8, 7, 7, 9, 4, 4, 10, 10, 0, 8, 8, 5, 5, 6, 6, 10, 10, 8, 8, 5, 5, 6, 6, 9, 9, 4, 4, 8, 8, 10, 10, 9, 9, 5, 5, 8, 8, 10, 10, 2, 7, 7, 4, 4, 7, 7, 9}, // 数据较长，省略内容
								},
								map[string]interface{}{
									"wheelLength": 100,
									"noWinIndex":  []int{0},
									"wheelData":   []int{7, 0, 4, 4, 9, 10, 10, 10, 7, 7, 10, 10, 2, 2, 8, 8, 3, 3, 6, 6, 0, 10, 10, 7, 7, 6, 6, 6, 4, 4, 7, 7, 3, 3, 8, 8, 5, 5, 6, 6, 8, 8, 3, 3, 6, 6, 9, 9, 4, 4, 10, 10, 8, 8, 9, 9, 5, 5, 6, 6, 6, 0, 8, 8, 3, 3, 10, 10, 10, 8, 8, 7, 7, 4, 4, 9, 9, 6, 6, 3, 3, 10, 9, 9, 7, 7, 10, 10, 9, 9, 4, 7, 7, 10, 10, 5, 5, 7, 7, 10}, // 数据较长，省略内容
								},
								map[string]interface{}{
									"wheelLength": 100,
									"noWinIndex":  []int{0},
									"wheelData":   []int{9, 9, 6, 6, 6, 5, 0, 9, 6, 6, 10, 10, 6, 6, 7, 7, 10, 10, 5, 5, 0, 8, 8, 6, 6, 8, 8, 8, 4, 4, 9, 9, 5, 5, 9, 9, 10, 10, 8, 8, 7, 7, 5, 5, 4, 4, 8, 8, 10, 10, 2, 2, 3, 3, 0, 4, 4, 6, 6, 9, 9, 0, 8, 8, 5, 5, 9, 9, 10, 10, 4, 4, 6, 6, 3, 3, 10, 10, 8, 8, 9, 9, 10, 10, 5, 5, 0, 9, 9, 4, 4, 10, 10, 7, 7, 2, 2, 6, 6, 6}, // 数据较长，省略内容
								},
								map[string]interface{}{
									"wheelLength": 100,
									"noWinIndex":  []int{0},
									"wheelData":   []int{10, 5, 5, 8, 8, 0, 5, 5, 2, 2, 8, 8, 10, 6, 6, 6, 10, 10, 9, 9, 0, 10, 10, 3, 3, 6, 6, 6, 4, 4, 7, 7, 3, 3, 8, 8, 9, 9, 6, 6, 8, 8, 5, 5, 6, 6, 5, 5, 10, 10, 9, 9, 8, 8, 7, 7, 2, 2, 8, 8, 7, 7, 0, 6, 6, 5, 5, 8, 8, 4, 4, 3, 3, 10, 10, 8, 8, 9, 9, 2, 2, 10, 10, 4, 4, 9, 9, 2, 7, 7, 4, 4, 4, 6, 6, 5, 5, 6, 6, 6}, // 数据较长，省略内容
								},
							},
							{
								map[string]interface{}{
									"wheelLength": 100,
									"noWinIndex":  []int{0},
									"wheelData":   []int{8, 2, 2, 7, 7, 9, 4, 4, 3, 3, 7, 7, 4, 4, 10, 10, 5, 5, 10, 10, 0, 6, 6, 8, 8, 5, 5, 3, 4, 4, 9, 9, 5, 5, 8, 8, 4, 4, 10, 10, 6, 8, 8, 5, 5, 10, 10, 3, 3, 0, 2, 2, 4, 10, 10, 7, 7, 2, 2, 6, 6, 7, 7, 8, 8, 5, 5, 9, 9, 10, 10, 2, 2, 2, 8, 8, 10, 10, 3, 3, 9, 9, 4, 4, 10, 10, 10, 9, 9, 4, 4, 10, 10, 2, 2, 9, 9, 9, 9, 8}, // 数据较长，省略内容
								},
								map[string]interface{}{
									"wheelLength": 100,
									"noWinIndex":  []int{0},
									"wheelData":   []int{9, 9, 7, 7, 7, 8, 10, 10, 5, 5, 7, 7, 3, 3, 8, 8, 5, 5, 10, 10, 0, 7, 7, 4, 4, 7, 7, 7, 4, 4, 9, 9, 2, 2, 10, 10, 7, 7, 9, 9, 8, 8, 5, 5, 6, 6, 5, 5, 9, 9, 3, 3, 8, 8, 7, 7, 9, 4, 4, 10, 10, 0, 8, 8, 5, 5, 6, 6, 10, 10, 3, 3, 9, 9, 8, 8, 4, 4, 10, 10, 9, 9, 5, 5, 9, 9, 5, 5, 8, 8, 10, 10, 0, 7, 7, 4, 4, 9, 9, 6}, // 数据较长，省略内容
								},
								map[string]interface{}{
									"wheelLength": 100,
									"noWinIndex":  []int{0},
									"wheelData":   []int{3, 3, 10, 10, 6, 7, 9, 9, 5, 5, 7, 7, 3, 3, 10, 10, 5, 5, 9, 9, 0, 7, 7, 4, 4, 7, 7, 7, 4, 4, 9, 9, 2, 2, 10, 10, 7, 7, 9, 9, 8, 8, 5, 5, 6, 6, 5, 5, 9, 9, 3, 3, 8, 8, 7, 7, 9, 4, 4, 10, 10, 0, 8, 8, 5, 5, 6, 6, 10, 10, 8, 8, 5, 5, 6, 6, 9, 9, 4, 4, 8, 8, 10, 10, 9, 9, 5, 5, 8, 8, 10, 10, 2, 7, 7, 4, 4, 7, 7, 9}, // 数据较长，省略内容
								},
								map[string]interface{}{
									"wheelLength": 100,
									"noWinIndex":  []int{0},
									"wheelData":   []int{7, 0, 4, 4, 9, 10, 10, 10, 7, 7, 10, 10, 2, 2, 8, 8, 3, 3, 6, 6, 0, 10, 10, 7, 7, 6, 6, 6, 4, 4, 7, 7, 3, 3, 8, 8, 5, 5, 6, 6, 8, 8, 3, 3, 6, 6, 9, 9, 4, 4, 10, 10, 8, 8, 9, 9, 5, 5, 6, 6, 6, 0, 8, 8, 3, 3, 10, 10, 10, 8, 8, 7, 7, 4, 4, 9, 9, 6, 6, 3, 3, 10, 9, 9, 7, 7, 10, 10, 9, 9, 4, 7, 7, 10, 10, 5, 5, 7, 7, 10}, // 数据较长，省略内容
								},
								map[string]interface{}{
									"wheelLength": 100,
									"noWinIndex":  []int{0},
									"wheelData":   []int{9, 9, 6, 6, 6, 5, 0, 9, 6, 6, 10, 10, 6, 6, 7, 7, 10, 10, 5, 5, 0, 8, 8, 6, 6, 8, 8, 8, 4, 4, 9, 9, 5, 5, 9, 9, 10, 10, 8, 8, 7, 7, 5, 5, 4, 4, 8, 8, 10, 10, 2, 2, 3, 3, 0, 4, 4, 6, 6, 9, 9, 0, 8, 8, 5, 5, 9, 9, 10, 10, 4, 4, 6, 6, 3, 3, 10, 10, 8, 8, 9, 9, 10, 10, 5, 5, 0, 9, 9, 4, 4, 10, 10, 7, 7, 2, 2, 6, 6, 6}, // 数据较长，省略内容
								},
								map[string]interface{}{
									"wheelLength": 100,
									"noWinIndex":  []int{0},
									"wheelData":   []int{10, 5, 5, 8, 8, 0, 5, 5, 2, 2, 8, 8, 10, 6, 6, 6, 10, 10, 9, 9, 0, 10, 10, 3, 3, 6, 6, 6, 4, 4, 7, 7, 3, 3, 8, 8, 9, 9, 6, 6, 8, 8, 5, 5, 6, 6, 5, 5, 10, 10, 9, 9, 8, 8, 7, 7, 2, 2, 8, 8, 7, 7, 0, 6, 6, 5, 5, 8, 8, 4, 4, 3, 3, 10, 10, 8, 8, 9, 9, 2, 2, 10, 10, 4, 4, 9, 9, 2, 7, 7, 4, 4, 4, 6, 6, 5, 5, 6, 6, 6}, // 数据较长，省略内容
								},
							},
						},
					},
					"symbolSetting": map[string]interface{}{
						"symbolCount":     11,
						"symbolAttribute": []string{"FreeGame_01", "BonusGame_01", "M1", "M2", "M3", "M4", "A", "K", "Q", "J", "TE"},
						"payTable": [][]int{
							{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
							{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
							{0, 0, 0, 0, 0, 0, 0, 200, 200, 500, 500, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000},
							{0, 0, 0, 0, 0, 0, 0, 50, 50, 200, 200, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500},
							{0, 0, 0, 0, 0, 0, 0, 40, 40, 100, 100, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300},
							{0, 0, 0, 0, 0, 0, 0, 30, 30, 40, 40, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240},
							{0, 0, 0, 0, 0, 0, 0, 20, 20, 30, 30, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200},
							{0, 0, 0, 0, 0, 0, 0, 16, 16, 24, 24, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160},
							{0, 0, 0, 0, 0, 0, 0, 10, 10, 20, 20, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100},
							{0, 0, 0, 0, 0, 0, 0, 8, 8, 18, 18, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80},
							{0, 0, 0, 0, 0, 0, 0, 5, 5, 15, 15, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40},
						},

						"mixGroupCount": 0,
					},
					"lineSetting": map[string]interface{}{
						"maxBetLine": 0,
					},
					"gameHitPatternSetting": map[string]interface{}{
						"gameHitPattern":    "QuantityGame",
						"maxEliminateTimes": 0,
					},
					"specialFeatureSetting": map[string]interface{}{
						"specialFeatureCount": 3,
						"specialHitInfo": []map[string]interface{}{
							{
								"specialHitPattern": "HP_109",
								"triggerEvent":      "Trigger_01",
								"basePay":           3,
							},
							{
								"specialHitPattern": "HP_110",
								"triggerEvent":      "Trigger_02",
								"basePay":           5,
							},
							{
								"specialHitPattern": "HP_124",
								"triggerEvent":      "Trigger_03",
								"basePay":           100,
							},
						},
					},
					"progressSetting": map[string]interface{}{
						"triggerLimitType": "NoLimit",
						"stepSetting": map[string]interface{}{
							"defaultStep": 1,
							"addStep":     0,
							"maxStep":     1,
						},
						"stageSetting": map[string]interface{}{
							"defaultStage": 1,
							"addStage":     0,
							"maxStage":     1,
						},
						"roundSetting": map[string]interface{}{
							"defaultRound": 1,
							"addRound":     0,
							"maxRound":     1,
						},
					},
					"displaySetting": map[string]interface{}{
						"readyHandSetting": map[string]interface{}{
							"readyHandLimitType": "NoReadyHandLimit",
							"readyHandCount":     1,
							"readyHandType":      []string{"ReadyHand_34"},
						},
					},

					"extendSetting": map[string]interface{}{
						"eliminatedMaxTimes":           999,
						"scatterC1Id":                  0,
						"scatterC2Id":                  1,
						"scatterMultiplier":            []int{2, 3, 5, 8, 10, 12, 15, 18, 20, 25, 35, 50, 100},
						"scatterMultiplierWeight":      []int{100, 100, 1000, 200, 120, 600, 50, 30, 20, 10, 5, 4, 2},
						"scatterMultiplierNoHitWeight": []int{200, 250, 300, 500, 350, 200, 150, 100, 80, 30, 20, 4, 2},
						"triggerRound": map[string]interface{}{
							"Trigger_01": map[string]interface{}{
								"defaultRound": 1,
								"addRound":     0,
								"maxRound":     1,
							},
							"Trigger_02": map[string]interface{}{
								"defaultRound": 1,
								"addRound":     0,
								"maxRound":     1,
							},
							"Trigger_03": map[string]interface{}{
								"defaultRound": 1,
								"addRound":     0,
								"maxRound":     1,
							},
						},
					},
				},
				map[string]interface{}{
					"gameStateType": "GS_161",
					"frameSetting": map[string]interface{}{
						"screenColumn":    6,
						"screenRow":       5,
						"wheelUsePattern": "PositionDependence",
					},
					"tableSetting": map[string]interface{}{
						"tableCount":          1,
						"tableHitProbability": []float64{1}, // 可填 []float64{0.8, 0.2}
						"wheelData": [][]map[string]interface{}{
							{
								map[string]interface{}{
									"wheelLength": 85,
									"noWinIndex":  []int{0},
									"wheelData":   []int{8, 2, 2, 7, 7, 9, 4, 4, 3, 3, 7, 7, 4, 4, 10, 10, 5, 5, 10, 10, 0, 6, 6, 8, 8, 5, 5, 3, 4, 4, 9, 9, 5, 5, 8, 8, 4, 4, 10, 10, 6, 8, 8, 5, 5, 10, 10, 3, 3, 0, 2, 2, 4, 10, 10, 7, 7, 2, 2, 6, 6, 7, 7, 8, 8, 5, 5, 9, 9, 10, 10, 2, 2, 2, 8, 8, 10, 10, 3, 3, 9, 9, 4, 4, 10, 10, 10, 9, 9, 4, 4, 10, 10, 2, 2, 9, 9, 9, 9, 8}, // 数据较长，省略内容
								},
								map[string]interface{}{
									"wheelLength": 85,
									"noWinIndex":  []int{0},
									"wheelData":   []int{9, 9, 7, 7, 7, 8, 10, 10, 5, 5, 7, 7, 3, 3, 8, 8, 5, 5, 10, 10, 0, 7, 7, 4, 4, 7, 7, 7, 4, 4, 9, 9, 2, 2, 10, 10, 7, 7, 9, 9, 8, 8, 5, 5, 6, 6, 5, 5, 9, 9, 3, 3, 8, 8, 7, 7, 9, 4, 4, 10, 10, 0, 8, 8, 5, 5, 6, 6, 10, 10, 3, 3, 9, 9, 8, 8, 4, 4, 10, 10, 9, 9, 5, 5, 9, 9, 5, 5, 8, 8, 10, 10, 0, 7, 7, 4, 4, 9, 9, 6}, // 数据较长，省略内容
								},
								map[string]interface{}{
									"wheelLength": 85,
									"noWinIndex":  []int{0},
									"wheelData":   []int{3, 3, 10, 10, 6, 7, 9, 9, 5, 5, 7, 7, 3, 3, 10, 10, 5, 5, 9, 9, 0, 7, 7, 4, 4, 7, 7, 7, 4, 4, 9, 9, 2, 2, 10, 10, 7, 7, 9, 9, 8, 8, 5, 5, 6, 6, 5, 5, 9, 9, 3, 3, 8, 8, 7, 7, 9, 4, 4, 10, 10, 0, 8, 8, 5, 5, 6, 6, 10, 10, 8, 8, 5, 5, 6, 6, 9, 9, 4, 4, 8, 8, 10, 10, 9, 9, 5, 5, 8, 8, 10, 10, 2, 7, 7, 4, 4, 7, 7, 9}, // 数据较长，省略内容
								},
								map[string]interface{}{
									"wheelLength": 85,
									"noWinIndex":  []int{0},
									"wheelData":   []int{7, 0, 4, 4, 9, 10, 10, 10, 7, 7, 10, 10, 2, 2, 8, 8, 3, 3, 6, 6, 0, 10, 10, 7, 7, 6, 6, 6, 4, 4, 7, 7, 3, 3, 8, 8, 5, 5, 6, 6, 8, 8, 3, 3, 6, 6, 9, 9, 4, 4, 10, 10, 8, 8, 9, 9, 5, 5, 6, 6, 6, 0, 8, 8, 3, 3, 10, 10, 10, 8, 8, 7, 7, 4, 4, 9, 9, 6, 6, 3, 3, 10, 9, 9, 7, 7, 10, 10, 9, 9, 4, 7, 7, 10, 10, 5, 5, 7, 7, 10}, // 数据较长，省略内容
								},
								map[string]interface{}{
									"wheelLength": 85,
									"noWinIndex":  []int{0},
									"wheelData":   []int{9, 9, 6, 6, 6, 5, 0, 9, 6, 6, 10, 10, 6, 6, 7, 7, 10, 10, 5, 5, 0, 8, 8, 6, 6, 8, 8, 8, 4, 4, 9, 9, 5, 5, 9, 9, 10, 10, 8, 8, 7, 7, 5, 5, 4, 4, 8, 8, 10, 10, 2, 2, 3, 3, 0, 4, 4, 6, 6, 9, 9, 0, 8, 8, 5, 5, 9, 9, 10, 10, 4, 4, 6, 6, 3, 3, 10, 10, 8, 8, 9, 9, 10, 10, 5, 5, 0, 9, 9, 4, 4, 10, 10, 7, 7, 2, 2, 6, 6, 6}, // 数据较长，省略内容
								},
								map[string]interface{}{
									"wheelLength": 85,
									"noWinIndex":  []int{0},
									"wheelData":   []int{10, 5, 5, 8, 8, 0, 5, 5, 2, 2, 8, 8, 10, 6, 6, 6, 10, 10, 9, 9, 0, 10, 10, 3, 3, 6, 6, 6, 4, 4, 7, 7, 3, 3, 8, 8, 9, 9, 6, 6, 8, 8, 5, 5, 6, 6, 5, 5, 10, 10, 9, 9, 8, 8, 7, 7, 2, 2, 8, 8, 7, 7, 0, 6, 6, 5, 5, 8, 8, 4, 4, 3, 3, 10, 10, 8, 8, 9, 9, 2, 2, 10, 10, 4, 4, 9, 9, 2, 7, 7, 4, 4, 4, 6, 6, 5, 5, 6, 6, 6}, // 数据较长，省略内容
								},
							},
						},
					},
					"symbolSetting": map[string]interface{}{
						"symbolCount":     11,
						"symbolAttribute": []string{"FreeGame_01", "BonusGame_01", "M1", "M2", "M3", "M4", "A", "K", "Q", "J", "TE"},
						"payTable": [][]int{
							{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
							{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
							{0, 0, 0, 0, 0, 0, 0, 200, 200, 500, 500, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000},
							{0, 0, 0, 0, 0, 0, 0, 50, 50, 200, 200, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500},
							{0, 0, 0, 0, 0, 0, 0, 40, 40, 100, 100, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300},
							{0, 0, 0, 0, 0, 0, 0, 30, 30, 40, 40, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240},
							{0, 0, 0, 0, 0, 0, 0, 20, 20, 30, 30, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200},
							{0, 0, 0, 0, 0, 0, 0, 16, 16, 24, 24, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160},
							{0, 0, 0, 0, 0, 0, 0, 10, 10, 20, 20, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100},
							{0, 0, 0, 0, 0, 0, 0, 8, 8, 18, 18, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80},
							{0, 0, 0, 0, 0, 0, 0, 5, 5, 15, 15, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40},
						},

						"mixGroupCount": 0,
					},
					"lineSetting": map[string]interface{}{
						"maxBetLine": 0,
					},
					"gameHitPatternSetting": map[string]interface{}{
						"gameHitPattern":    "QuantityGame",
						"maxEliminateTimes": 0,
					},
					"specialFeatureSetting": map[string]interface{}{
						"specialFeatureCount": 4,
						"specialHitInfo": []map[string]interface{}{
							{
								"specialHitPattern": "HP_108",
								"triggerEvent":      "Trigger_01",
								"basePay":           0,
							},
							{
								"specialHitPattern": "HP_109",
								"triggerEvent":      "Trigger_02",
								"basePay":           0,
							},
							{
								"specialHitPattern": "HP_110",
								"triggerEvent":      "Trigger_03",
								"basePay":           0,
							},
							{
								"specialHitPattern": "HP_124",
								"triggerEvent":      "Trigger_04",
								"basePay":           0,
							},
						},
					},
					"progressSetting": map[string]interface{}{
						"triggerLimitType": "NoLimit",
						"stepSetting": map[string]interface{}{
							"defaultStep": 1,
							"addStep":     0,
							"maxStep":     1,
						},
						"stageSetting": map[string]interface{}{
							"defaultStage": 1,
							"addStage":     0,
							"maxStage":     1,
						},
						"roundSetting": map[string]interface{}{
							"defaultRound": 10,
							"addRound":     5,
							"maxRound":     30,
						},
					},
					"displaySetting": map[string]interface{}{
						"readyHandSetting": map[string]interface{}{
							"readyHandLimitType": "NoReadyHandLimit",
							"readyHandCount":     1,
							"readyHandType":      []string{"ReadyHand_34"},
						},
					},

					"extendSetting": map[string]interface{}{
						"eliminatedMaxTimes":           999,
						"scatterC1Id":                  0,
						"scatterC2Id":                  1,
						"scatterMultiplier":            []int{2, 3, 5, 8, 10, 12, 15, 18, 20, 25, 35, 50, 100},
						"scatterMultiplierWeight":      []int{2100, 1500, 800, 200, 100, 50, 20, 10, 8, 5, 3, 2, 1},
						"scatterMultiplierNoHitWeight": []int{400, 400, 500, 300, 200, 100, 80, 50, 30, 20, 10, 4, 2},
						"triggerRound": map[string]interface{}{
							"Trigger_01": map[string]interface{}{
								"defaultRound": 10,
								"addRound":     5,
								"maxRound":     30,
							},
							"Trigger_02": map[string]interface{}{
								"defaultRound": 10,
								"addRound":     5,
								"maxRound":     30,
							},
							"Trigger_03": map[string]interface{}{
								"defaultRound": 10,
								"addRound":     5,
								"maxRound":     30,
							},
							"Trigger_04": map[string]interface{}{
								"defaultRound": 10,
								"addRound":     5,
								"maxRound":     30,
							},
						},
					},
				},
				map[string]interface{}{
					"gameStateType": "GS_161",
					"frameSetting": map[string]interface{}{
						"screenColumn":    6,
						"screenRow":       5,
						"wheelUsePattern": "PositionDependence",
					},
					"tableSetting": map[string]interface{}{
						"tableCount":          1,
						"tableHitProbability": []float64{1}, // 可填 []float64{0.8, 0.2}
						"wheelData": [][]map[string]interface{}{
							{
								map[string]interface{}{
									"wheelLength": 100,
									"noWinIndex":  []int{0},
									"wheelData":   []int{8, 2, 2, 7, 7, 9, 4, 4, 3, 3, 7, 7, 4, 4, 10, 10, 5, 5, 10, 10, 0, 6, 6, 8, 8, 5, 5, 3, 4, 4, 9, 9, 5, 5, 8, 8, 4, 4, 10, 10, 6, 8, 8, 5, 5, 10, 10, 3, 3, 0, 2, 2, 4, 10, 10, 7, 7, 2, 2, 6, 6, 7, 7, 8, 8, 5, 5, 9, 9, 10, 10, 2, 2, 2, 8, 8, 10, 10, 3, 3, 9, 9, 4, 4, 10, 10, 10, 9, 9, 4, 4, 10, 10, 2, 2, 9, 9, 9, 9, 8}, // 数据较长，省略内容
								},
								map[string]interface{}{
									"wheelLength": 100,
									"noWinIndex":  []int{0},
									"wheelData":   []int{9, 9, 7, 7, 7, 8, 10, 10, 5, 5, 7, 7, 3, 3, 8, 8, 5, 5, 10, 10, 0, 7, 7, 4, 4, 7, 7, 7, 4, 4, 9, 9, 2, 2, 10, 10, 7, 7, 9, 9, 8, 8, 5, 5, 6, 6, 5, 5, 9, 9, 3, 3, 8, 8, 7, 7, 9, 4, 4, 10, 10, 0, 8, 8, 5, 5, 6, 6, 10, 10, 3, 3, 9, 9, 8, 8, 4, 4, 10, 10, 9, 9, 5, 5, 9, 9, 5, 5, 8, 8, 10, 10, 0, 7, 7, 4, 4, 9, 9, 6}, // 数据较长，省略内容
								},
								map[string]interface{}{
									"wheelLength": 100,
									"noWinIndex":  []int{0},
									"wheelData":   []int{3, 3, 10, 10, 6, 7, 9, 9, 5, 5, 7, 7, 3, 3, 10, 10, 5, 5, 9, 9, 0, 7, 7, 4, 4, 7, 7, 7, 4, 4, 9, 9, 2, 2, 10, 10, 7, 7, 9, 9, 8, 8, 5, 5, 6, 6, 5, 5, 9, 9, 3, 3, 8, 8, 7, 7, 9, 4, 4, 10, 10, 0, 8, 8, 5, 5, 6, 6, 10, 10, 8, 8, 5, 5, 6, 6, 9, 9, 4, 4, 8, 8, 10, 10, 9, 9, 5, 5, 8, 8, 10, 10, 2, 7, 7, 4, 4, 7, 7, 9}, // 数据较长，省略内容
								},
								map[string]interface{}{
									"wheelLength": 100,
									"noWinIndex":  []int{0},
									"wheelData":   []int{7, 0, 4, 4, 9, 10, 10, 10, 7, 7, 10, 10, 2, 2, 8, 8, 3, 3, 6, 6, 0, 10, 10, 7, 7, 6, 6, 6, 4, 4, 7, 7, 3, 3, 8, 8, 5, 5, 6, 6, 8, 8, 3, 3, 6, 6, 9, 9, 4, 4, 10, 10, 8, 8, 9, 9, 5, 5, 6, 6, 6, 0, 8, 8, 3, 3, 10, 10, 10, 8, 8, 7, 7, 4, 4, 9, 9, 6, 6, 3, 3, 10, 9, 9, 7, 7, 10, 10, 9, 9, 4, 7, 7, 10, 10, 5, 5, 7, 7, 10}, // 数据较长，省略内容
								},
								map[string]interface{}{
									"wheelLength": 100,
									"noWinIndex":  []int{0},
									"wheelData":   []int{9, 9, 6, 6, 6, 5, 0, 9, 6, 6, 10, 10, 6, 6, 7, 7, 10, 10, 5, 5, 0, 8, 8, 6, 6, 8, 8, 8, 4, 4, 9, 9, 5, 5, 9, 9, 10, 10, 8, 8, 7, 7, 5, 5, 4, 4, 8, 8, 10, 10, 2, 2, 3, 3, 0, 4, 4, 6, 6, 9, 9, 0, 8, 8, 5, 5, 9, 9, 10, 10, 4, 4, 6, 6, 3, 3, 10, 10, 8, 8, 9, 9, 10, 10, 5, 5, 0, 9, 9, 4, 4, 10, 10, 7, 7, 2, 2, 6, 6, 6}, // 数据较长，省略内容
								},
								map[string]interface{}{
									"wheelLength": 100,
									"noWinIndex":  []int{0},
									"wheelData":   []int{10, 5, 5, 8, 8, 0, 5, 5, 2, 2, 8, 8, 10, 6, 6, 6, 10, 10, 9, 9, 0, 10, 10, 3, 3, 6, 6, 6, 4, 4, 7, 7, 3, 3, 8, 8, 9, 9, 6, 6, 8, 8, 5, 5, 6, 6, 5, 5, 10, 10, 9, 9, 8, 8, 7, 7, 2, 2, 8, 8, 7, 7, 0, 6, 6, 5, 5, 8, 8, 4, 4, 3, 3, 10, 10, 8, 8, 9, 9, 2, 2, 10, 10, 4, 4, 9, 9, 2, 7, 7, 4, 4, 4, 6, 6, 5, 5, 6, 6, 6}, // 数据较长，省略内容
								},
							},
						},
						"screenControlSetting": map[string]interface{}{
							"scatterId":               0,
							"scatterPatternHitWeight": []int{0, 0, 0, 0, 10000, 77, 5},
							"scatterTargetColumn":     []int{0, 1, 2, 3, 4, 5},
							"repeatScatter":           false,
							"continuous":              false,
						},
					},
					"symbolSetting": map[string]interface{}{
						"symbolCount":     11,
						"symbolAttribute": []string{"FreeGame_01", "BonusGame_01", "M1", "M2", "M3", "M4", "A", "K", "Q", "J", "TE"},
						"payTable": [][]int{
							{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
							{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
							{0, 0, 0, 0, 0, 0, 0, 50, 50, 200, 200, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000},
							{0, 0, 0, 0, 0, 0, 0, 30, 30, 100, 100, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500},
							{0, 0, 0, 0, 0, 0, 0, 20, 20, 80, 80, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300},
							{0, 0, 0, 0, 0, 0, 0, 10, 10, 30, 30, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240},
							{0, 0, 0, 0, 0, 0, 0, 8, 8, 20, 20, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200},
							{0, 0, 0, 0, 0, 0, 0, 5, 5, 10, 10, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160, 160},
							{0, 0, 0, 0, 0, 0, 0, 3, 3, 8, 8, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100},
							{0, 0, 0, 0, 0, 0, 0, 2, 2, 5, 5, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80},
							{0, 0, 0, 0, 0, 0, 0, 1, 1, 3, 3, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40},
						},

						"mixGroupCount": 0,
					},
					"lineSetting": map[string]interface{}{
						"maxBetLine": 0,
					},
					"gameHitPatternSetting": map[string]interface{}{
						"gameHitPattern":    "QuantityGame",
						"maxEliminateTimes": 0,
					},
					"specialFeatureSetting": map[string]interface{}{
						"specialFeatureCount": 3,
						"specialHitInfo": []map[string]interface{}{
							{
								"specialHitPattern": "HP_109",
								"triggerEvent":      "Trigger_01",
								"basePay":           3,
							},
							{
								"specialHitPattern": "HP_110",
								"triggerEvent":      "Trigger_02",
								"basePay":           5,
							},
							{
								"specialHitPattern": "HP_124",
								"triggerEvent":      "Trigger_03",
								"basePay":           100,
							},
						},
					},
					"progressSetting": map[string]interface{}{
						"triggerLimitType": "NoLimit",
						"stepSetting": map[string]interface{}{
							"defaultStep": 1,
							"addStep":     0,
							"maxStep":     1,
						},
						"stageSetting": map[string]interface{}{
							"defaultStage": 1,
							"addStage":     0,
							"maxStage":     1,
						},
						"roundSetting": map[string]interface{}{
							"defaultRound": 1,
							"addRound":     0,
							"maxRound":     1,
						},
					},
					"displaySetting": map[string]interface{}{
						"readyHandSetting": map[string]interface{}{
							"readyHandLimitType": "NoReadyHandLimit",
							"readyHandCount":     1,
							"readyHandType":      []string{"ReadyHand_34"},
						},
					},

					"extendSetting": map[string]interface{}{
						"eliminatedMaxTimes":           999,
						"scatterC1Id":                  0,
						"scatterC2Id":                  1,
						"scatterMultiplier":            []int{2, 3, 5, 8, 10, 12, 15, 18, 20, 25, 35, 50, 100},
						"scatterMultiplierWeight":      []int{100, 100, 1000, 200, 120, 600, 50, 30, 20, 10, 5, 4, 2},
						"scatterMultiplierNoHitWeight": []int{200, 250, 300, 500, 350, 200, 150, 100, 80, 30, 20, 4, 2},
						"triggerRound": map[string]interface{}{
							"Trigger_01": map[string]interface{}{
								"defaultRound": 1,
								"addRound":     0,
								"maxRound":     1,
							},
							"Trigger_02": map[string]interface{}{
								"defaultRound": 1,
								"addRound":     0,
								"maxRound":     1,
							},
							"Trigger_03": map[string]interface{}{
								"defaultRound": 1,
								"addRound":     0,
								"maxRound":     1,
							},
						},
					},
				},
			},

			"doubleGameSetting": map[string]interface{}{
				"doubleRoundUpperLimit": 5,
				"doubleBetUpperLimit":   1000000000,
				"rtp":                   0.96,
				"tieRate":               0.1,
			},

			"boardDisplaySetting": map[string]interface{}{
				"winRankSetting": map[string]interface{}{
					"BigWin":   10,
					"MegaWin":  35,
					"UltraWin": 80,
				},
			},
			"gameFlowSetting": map[string]interface{}{
				"conditionTableWithoutBoardEnd": [][]string{
					{"CD_False", "CD_38", "CD_False", "CD_37", "CD_False"},
					{"CD_False", "CD_False", "CD_12", "CD_False", "CD_False"},
					{"CD_False", "CD_False", "CD_False", "CD_False", "CD_False"},
					{"CD_False", "CD_False", "CD_False", "CD_False", "CD_12"},
					{"CD_False", "CD_False", "CD_False", "CD_False", "CD_False"},
				},
			},

			"reiterateSpinCriterion": map[string]interface{}{
				"oddsIntervalSetting": []map[string]interface{}{
					{"minOdds": 0, "maxOdds": 0.0001, "rejectProb": 0.38},
					{"minOdds": 0.0001, "maxOdds": 1, "rejectProb": 0.65},
					{"minOdds": 1, "maxOdds": 2, "rejectProb": 0.99},
					{"minOdds": 2, "maxOdds": 3, "rejectProb": 0.99},
					{"minOdds": 3, "maxOdds": 4, "rejectProb": 0.9},
					{"minOdds": 4, "maxOdds": 5, "rejectProb": 0.6},
					{"minOdds": 5, "maxOdds": 6, "rejectProb": 0.5},
					{"minOdds": 6, "maxOdds": 7, "rejectProb": 0.2},
					{"minOdds": 50, "maxOdds": 60, "rejectProb": 0.1},
					{"minOdds": 60, "maxOdds": 70, "rejectProb": 0.35},
					{"minOdds": 70, "maxOdds": 80, "rejectProb": 0.4},
					{"minOdds": 80, "maxOdds": 90, "rejectProb": 0.4},
					{"minOdds": 90, "maxOdds": 100, "rejectProb": 0.3},
					{"minOdds": 100, "maxOdds": 110, "rejectProb": 0.35},
					{"minOdds": 110, "maxOdds": 120, "rejectProb": 0.3},
					{"minOdds": 120, "maxOdds": 130, "rejectProb": 0.2},
					{"minOdds": 130, "maxOdds": 140, "rejectProb": 0.2},
					{"minOdds": 140, "maxOdds": 150, "rejectProb": 0.2},
					{"minOdds": 150, "maxOdds": 160, "rejectProb": 0.1},
				},
			},
			"rValue": map[string]interface{}{
				"NoExtraBet":    28934,
				"BuyFeature_01": 28898,
			},
		},
		Denoms:          []int{10},
		DefaultDenomIdx: 0,
		BuyFeature:      true,
		BuyFeatureLimit: math.MaxInt32,
	}

	// 转换为 JSON 字符串
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(jsonBytes)
}
