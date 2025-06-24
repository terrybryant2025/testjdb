package main

import (
	"encoding/json"
	"fmt"
	"github.com/xuri/excelize/v2"
	"math/rand"
	"time"
)

var GameList []*RicherSpinResult

type RicherSpinResult struct {
	Winnings     int // 赢得的金额(没有扣除下注额)
	BetChip      int // 下注金额
	BetMulti     int // 下注倍数
	RTP          float64
	GameType     int    // 1普通 2大奖 3特将
	GameTypeName string // 1普通 2大奖 3特将
	Card         *Card
}

func simulateGame(baseBet, baseMulti int) *RicherSpinResult {
	card := NewCard(baseMulti).DealCards().StatisticCard()
	card.CalcBaseScoreAndFreeGame(oddsMap)
	totalScore := card.StartFreeGame()
	rtp := float64(totalScore) / float64(baseBet) * 100

	gameTypeName := ""
	gameType := 0
	if card.FreeGameNum > 0 || rtp >= 3000 { // 巨奖
		gameType = 4
		gameTypeName = "巨奖"
	} else if rtp < 3000 && rtp >= 2000 { // 特大奖
		gameType = 3
		gameTypeName = "特大奖"
	} else if rtp < 2000 && rtp >= 1000 {
		gameType = 2
		gameTypeName = "大奖"
	} else if rtp < 1000 {
		gameType = 1
		gameTypeName = "普通奖"
	}

	return &RicherSpinResult{Winnings: totalScore, BetChip: baseBet, BetMulti: baseMulti, RTP: rtp, GameType: gameType, GameTypeName: gameTypeName, Card: card}
}

func adjustRTP(results []*RicherSpinResult, baseBet, betMulti int, targetRTP float64) []*RicherSpinResult {
	//return results
	for {
		totalBet := (len(results)) * baseBet
		totalScore := 0

		bestWin, superWin, bigWin, normalWin := 0, 0, 0, 0
		best, super, big, n := 0, 0, 0, 0
		for _, res := range results {
			totalScore += res.Winnings
			switch res.GameType {
			case 4:
				bestWin += res.Winnings
				best++
			case 3:
				superWin += res.Winnings
				super++
			case 2:
				bigWin += res.Winnings
				big++
			case 1:
				normalWin += res.Winnings
				n++
			}
		}

		overallRTP := float64(totalScore) / float64(totalBet) * 100

		lowerBound := targetRTP - 0.5
		upperBound := targetRTP + 0.5

		if overallRTP >= lowerBound && overallRTP <= upperBound {
			fmt.Printf("✅ 最终整体 RTP: %.4f%% 在目标范围 [%.2f%%, %.2f%%]\n", overallRTP, lowerBound, upperBound)
			break
		}

		fmt.Printf("🔄 当前 RTP: %.4f%% 不在目标范围，尝试补偿...\n", overallRTP)
		fmt.Printf("🔄 巨奖奖励:%d  场次:%d, 特大奖奖励:%d 场次:%d, 大奖奖励:%d 场次:%d, 普通奖励:%d 场次:%d,  \n", bestWin, best, superWin, super, bigWin, big, normalWin, n)

		newRes := simulateGame(baseBet, betMulti)
		if newRes.RTP > 0 {
			fmt.Println("补偿一个大于0的")
		}
		//results = append(results, newRes)
		newGameType := newRes.GameType

		// 如果超出上限，找一个高 RTP 的替换为低 RTP
		if overallRTP > upperBound {
			// 找到 RTP 最高的那一局替换
			idx := 0
			maxRTP := results[0].RTP
			for i, res := range results {
				if res.RTP > maxRTP && newGameType == res.GameType {
					maxRTP = res.RTP
					idx = i
				}
			}
			if maxRTP > newRes.RTP {
				fmt.Printf("📈 高于上限，替换第 %d 轮 RTP %.2f → 补偿 RTP %.2f\n", idx+1, results[idx].RTP, newRes.RTP)
				results[idx] = newRes
			}

		} else if overallRTP < lowerBound {
			// 找到 RTP 最低的那一局替换
			idx := 0
			minRTP := results[0].RTP
			for i, res := range results {
				if res.RTP < minRTP && newGameType == res.GameType {
					minRTP = res.RTP
					idx = i
				}
			}

			if minRTP < newRes.RTP {
				fmt.Printf("📉 低于下限，替换第 %d 轮 RTP %.2f → 补偿 RTP %.2f\n", idx+1, results[idx].RTP, newRes.RTP)
				results[idx] = newRes
			}
		}
		time.Sleep(100 * time.Millisecond) // 控制频率
	}
	return results
}

func adjustGameType(results []*RicherSpinResult, baseBet, betMulti int, diffBest, diffSuper, diffBig int) []*RicherSpinResult {
	//return results
	if diffSuper == 0 && diffBig == 0 && diffBest == 0 {
		return results
	}

	for {
		newRes := simulateGame(baseBet, betMulti)

		// 大奖太少，增加奖项
		if diffBig > 0 && newRes.GameType == 2 {
			// 缺少大奖，将一个普通奖替换为大奖
			i := rand.Intn(len(results)) // 随机找一个偏移量，直到找到一个普通奖，然后进行替换
			for {
				if i >= len(results) {
					i = 0
				}
				tmp := results[i]
				if tmp.GameType == 1 {
					results[i] = newRes
					break
				}
				i++
			}
			fmt.Printf("🔧 缺少大奖，将第 %d 局普通奖替换为大奖\n", i)
			results[i] = newRes
			diffBig--
		} else if diffSuper > 0 && newRes.GameType == 3 {
			// 缺少特大奖，将一个普通奖替换为特大奖
			i := rand.Intn(len(results)) // 随机找一个偏移量，直到找到一个普通奖，然后进行替换
			for {
				if i >= len(results) {
					i = 0
				}
				tmp := results[i]
				if tmp.GameType == 1 {
					results[i] = newRes
					break
				}
				i++
			}
			fmt.Printf("🔧 缺少特大奖，将第 %d 局普通奖替换为特大奖\n", i+1)
			results[i] = newRes
			diffSuper--
		} else if diffBest > 0 && newRes.GameType == 4 {
			// 缺少巨奖，将一个普通奖替换为巨奖
			i := rand.Intn(len(results)) // 随机找一个偏移量，直到找到一个普通奖，然后进行替换
			for {
				if i >= len(results) {
					i = 0
				}
				tmp := results[i]
				if tmp.GameType == 1 {
					results[i] = newRes
					break
				}
				i++
			}
			fmt.Printf("🔧 缺少巨奖，将第 %d 局普通奖替换为巨奖\n", i)
			results[i] = newRes
			diffBest--
		}

		// 大奖过多，减少奖项
		if diffBig < 0 && newRes.GameType == 1 {
			// 多了大奖，将一个大奖替换为普通奖
			i := rand.Intn(len(results))
			for {
				if i >= len(results) {
					i = 0
				}
				tmp := results[i]
				if tmp.GameType == 2 {
					results[i] = newRes
					break
				}
				i++
			}
			fmt.Printf("🔧 大奖超出，将第 %d 局大奖替换为普通奖\n", i+1)

			results[i] = newRes
			diffBig++
		} else if diffSuper < 0 && newRes.GameType == 1 {
			// 多了特大奖，将一个特大奖替换为普通奖
			i := rand.Intn(len(results))
			for {
				if i >= len(results) {
					i = 0
				}
				tmp := results[i]
				if tmp.GameType == 3 {
					results[i] = newRes
					break
				}
				i++
			}
			fmt.Printf("🔧 特大奖超出，将第 %d 局特大奖替换为普通奖\n", i+1)
			results[i] = newRes
			diffSuper++
		} else if diffBest < 0 && newRes.GameType == 1 {
			// 多了巨奖，将一个巨奖替换为普通奖
			i := rand.Intn(len(results))
			for {
				if i >= len(results) {
					i = 0
				}
				tmp := results[i]
				if tmp.GameType == 4 {
					results[i] = newRes
					break
				}
				i++
			}
			fmt.Printf("🔧 巨奖超出，将第 %d 局巨奖替换为普通奖\n", i+1)
			results[i] = newRes
			diffBest++
		}

		if diffSuper == 0 && diffBig == 0 && diffBest == 0 {
			fmt.Printf("✅ 符合要求，游戏数据生成完毕！\n")
			break
		}
	}

	return results
}

func generateSpinData(targetRTP float64) []*RicherSpinResult {
	rand.Seed(time.Now().UnixNano())
	bet := 50           // 50分
	betMulti := 50 / 50 // 下注金额 / 基数 = 倍数
	var results []*RicherSpinResult

	//
	//gameCount := 100   // 生成游戏场次
	//superRatio := 0.09 // 0.09% 巨奖比例  >60倍
	//bigRatio := 5.0    // 5% 大奖比例  >30倍 <60倍
	//ExpectSuperGameCount := int(math.Ceil((float64(gameCount) * superRatio) / 100))
	//ExpectBigGameCount := int(math.Ceil((float64(gameCount) * bigRatio) / 100))

	ExpectBestGameCount := 3
	ExpectSuperGameCount := 10
	ExpectBigGameCount := 50

	bigGameCount := 0
	SuperGameCount := 0
	BestGameCount := 0
	fmt.Printf("🎲 开始生成 100 局游戏数据，目标 RTP %.2f%%\n", targetRTP)
	for i := 1; i <= 5000; i++ {
		res := simulateGame(bet, betMulti)
		results = append(results, res)
		fmt.Printf("🎮 第 %3d 局: 投入 %3d | 得分 %3d | RTP: %.4f%% | 类型: %s \n", i, bet, res.Winnings, res.RTP, res.GameTypeName)
		if res.GameType == 3 {
			SuperGameCount++
		} else if res.GameType == 2 {
			bigGameCount++
		} else if res.GameType == 4 {
			BestGameCount++
		}
	}

	// 场次补偿
	results = adjustGameType(results, bet, betMulti, ExpectBestGameCount-BestGameCount, ExpectSuperGameCount-SuperGameCount, ExpectBigGameCount-bigGameCount)

	// rtp补偿
	results = adjustRTP(results, bet, betMulti, targetRTP)

	// 输出最终统计
	totalBet := len(results) * bet
	totalScore := 0
	for _, res := range results {
		totalScore += res.Winnings
	}
	finalRTP := float64(totalScore) / float64(totalBet) * 100
	fmt.Printf("\n📊 最终统计数据：共 %d 局 | 总投入: %d | 总得分: %d | 最终 RTP: %.4f%%\n",
		len(results), totalBet, totalScore, finalRTP)

	exportToExcel(results, "gameRecord.xlsx", bet)
	GameList = results
	return results
}

func exportToExcel(results []*RicherSpinResult, filename string, bet int) error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("关闭文件失败:", err)
		}
	}()

	sheet := "游戏记录"
	index, _ := f.NewSheet(sheet)
	f.SetActiveSheet(index)

	// 设置表头
	headers := []string{
		"局数", "投入积分", "得分", "RTP", "游戏类型", "游戏类型名称", "游戏结果",
	}
	for i, h := range headers {
		cell, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, cell+"1", h)
	}

	// 写入数据
	for i, res := range results {
		row := i + 2 // 从第二行开始写入
		res.Card.ToSerializable()
		j, _ := json.Marshal(res)

		f.SetCellValue(sheet, "A"+fmt.Sprintf("%d", row), i+1)
		f.SetCellValue(sheet, "B"+fmt.Sprintf("%d", row), bet)
		f.SetCellValue(sheet, "C"+fmt.Sprintf("%d", row), res.Winnings)
		f.SetCellValue(sheet, "D"+fmt.Sprintf("%d", row), res.RTP)
		f.SetCellValue(sheet, "E"+fmt.Sprintf("%d", row), res.GameType)
		f.SetCellValue(sheet, "F"+fmt.Sprintf("%d", row), res.GameTypeName)
		f.SetCellValue(sheet, "G"+fmt.Sprintf("%d", row), string(j))
	}

	// 自动调整列宽
	for _, col := range []string{"A", "B", "C", "D", "E"} {
		f.SetColWidth(sheet, col, col, 15)
	}

	// 保存文件
	if err := f.SaveAs(filename); err != nil {
		return err
	}

	fmt.Printf("✅ 数据已成功导出到 %s\n", filename)
	return nil
}

// 从 Excel 文件导入数据并保存到 GameList
func importFromExcel(filename string) error {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("关闭文件失败:", err)
		}
	}()

	sheetName := "游戏记录"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("读取工作表失败: %v", err)
	}

	// 跳过表头
	for i, row := range rows {
		if i == 0 {
			continue // 跳过第一行（标题）
		}
		if i == 2170-1 {
			fmt.Println(1)
		}

		if len(row) < 7 {
			continue // 数据不完整，跳过
		}

		// 解析 JSON 字符串为 RicherSpinResult
		var result RicherSpinResult
		if err := json.Unmarshal([]byte(row[6]), &result); // 第7列是 JSON 数据
		err != nil {
			fmt.Printf("解析第 %d 行数据失败: %v\n", i+1, err)
			continue
		}

		// 追加到全局变量 GameList
		GameList = append(GameList, &result)
	}

	fmt.Printf("✅ 成功从 %s 导入 %d 条记录到 GameList\n", filename, len(GameList))
	return nil
}

type SerializableCard struct {
	CardCounts       map[int][]int    `json:"card_counts"`
	ScreenHits       map[int][][]bool `json:"screen_hits"`
	SerialRollCounts map[int]int      `json:"serial_roll_counts"`
}

func (sc *SerializableCard) FromSerializable() *Card {
	card := &Card{
		CardCounts:          make(map[CardVal][]int),
		CardScreenHit:       make(map[CardVal][][]bool),
		CardSerialRollCount: make(map[CardVal]int),
	}

	// cardCounts
	for k, v := range sc.CardCounts {
		cv := CardVal(k)
		card.CardCounts[cv] = v
	}

	// screen hits
	for k, v := range sc.ScreenHits {
		cv := CardVal(k)
		card.CardScreenHit[cv] = v
	}

	// serial roll count
	for k, v := range sc.SerialRollCounts {
		cv := CardVal(k)
		card.CardSerialRollCount[cv] = v
	}

	return card
}
