package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

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

var g *AviatorGameContext = nil

func main() {

	g = NewGameContext()
	g.Init()
	g.NewGameInit()

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
	g.OnLogin(conn, obj)
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

	fmt.Printf("📨 CallExtension: cmd=%s, reqId=%v, params=%v\n", cmd, params)

	switch cmd {
	case "GEN_HEARTBEAT":
		handleGENHeartbeat(conn, params)
	case "PING_REQUEST":
		//handlePingRequest(conn, params)
	default:
		g.OnRecv(conn, obj)
		//fmt.Printf("⚠️ 未知扩展命令: %s\n", cmd)
	}
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

func handlePingRequest(conn *websocket.Conn, obj map[string]interface{}) {
	p := map[string]interface{}{
		"p": map[string]interface{}{},
		"c": "PING_RESPONSE",
	}
	packet := BuildSFSMessage(13, 1, p)
	conn.WriteMessage(websocket.BinaryMessage, packet)
}
