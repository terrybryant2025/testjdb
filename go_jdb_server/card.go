package main

import (
	"fmt"
	"math/rand"
)

type CardVal int

const (
	Richer     CardVal = 0 // 富豪
	Treasure   CardVal = 1 // 宝箱
	AirPlane   CardVal = 2 // 飞机
	Yacht      CardVal = 3 //游艇
	SportCar   CardVal = 4 // 跑车
	Motorcycle CardVal = 5 // 摩托
	Ace        CardVal = 6 // A
	King       CardVal = 7 // K
	Queen      CardVal = 8 //  Q
	Jack       CardVal = 9 // J
)

type PosIdx struct {
	X, Y int
}

type Card struct {
	Rolls    [5][3]CardVal // 每个滚筒上面3张卡片
	BetMulti int           // 下注倍数

	//cardIdx             map[CardVal][]PosIdx // 卡片出现的位置

	CardCounts          map[CardVal][]int    // 记录卡片 每个轮出现的次数
	CardScreenHit       map[CardVal][][]bool // 卡片出现的位置
	CardSerialRollCount map[CardVal]int      // 记录卡片连续出现在滚轮的次数(2个滚轮最多是2)
	BaseScore           int

	FreeGameNum        int     // 免费游戏次数
	FreeGameCard       []*Card // 免费游戏结果
	FreeGameTotalScore int     // 免费游戏盈利
}

func NewCard(betMulti int) *Card {
	card := Card{}
	card.FreeGameCard = make([]*Card, 0)
	card.BetMulti = betMulti

	card.CardCounts = make(map[CardVal][]int, 3)
	card.CardCounts = make(map[CardVal][]int, 3)
	card.CardSerialRollCount = make(map[CardVal]int, 3)
	card.CardScreenHit = make(map[CardVal][][]bool, 3)
	//card.cardIdx = make(map[CardVal][]PosIdx, 3)

	return &card
}

// DealCards 发卡函数
func (d *Card) DealCards() *Card {
	// 模拟5轮滚筒
	for i := 0; i < 5; i++ {
		d.generateRoundCards(i)
		fmt.Println("轮次:", i, ":", d.Rolls[i])
	}

	// 模拟固定值
	//d.Rolls = [5][3]CardVal{
	//	{Yacht, AirPlane, King},
	//	{AirPlane, AirPlane, King},
	//	{Richer, Jack, Richer},
	//	{Richer, King, Jack},
	//	{Yacht, Yacht, AirPlane},
	//}

	//d.Rolls = [5][3]CardVal{
	//	{AirPlane, Ace, Motorcycle},
	//	{SportCar, Ace, Ace},
	//	{King, Ace, Ace},
	//	{Richer, King, Ace},
	//	{Treasure, Treasure, Motorcycle},
	//}

	//d.Rolls = [5][3]CardVal{
	//	{King, Motorcycle, King},
	//	{Ace, Treasure, King},
	//	{Ace, King, Richer},
	//	{Yacht, King, Jack},
	//	{Yacht, Treasure, AirPlane},
	//}

	//d.Rolls = [5][3]CardVal{
	//	{SportCar, AirPlane, Yacht},
	//	{Richer, Ace, Richer},
	//	{King, Richer, Yacht},
	//	{SportCar, AirPlane, Yacht},
	//	{Treasure, Treasure, Motorcycle},
	//}

	//d.Rolls = [5][3]CardVal{
	//	{Ace, Ace, King},
	//	{Richer, Ace, Richer},
	//	{Yacht, Motorcycle, King},
	//	{Richer, Queen, Yacht},
	//	{King, Treasure, AirPlane},
	//}
	return d
}

func (d *Card) generateRoundCards(round int) {
	for j := 0; j < 3; j++ {
		if round == 0 {
			// 其他卡片（ID = 1到9）随机抽取
			d.Rolls[round][j] = CardVal(rand.Intn(8) + 1)
		} else {
			// 0-9
			d.Rolls[round][j] = CardVal(rand.Intn(10))
		}
	}
}

// StatisticCard 统计卡片
func (d *Card) StatisticCard() *Card {
	// 第一只轮 记录需要可能出现获得积分的卡 (最左只有3张，所以只会存在三种卡片满足得分)
	for i := 0; i < 3; i++ {
		cardVal := d.Rolls[0][i]
		if _, ok := d.CardSerialRollCount[cardVal]; !ok {
			d.CardCounts[cardVal] = make([]int, 5)
		}
		d.CardCounts[cardVal][0]++
		d.CardSerialRollCount[cardVal] = 1
		//d.cardIdx[cardVal] = []PosIdx{{0, i}}
		d.setScreenHit(cardVal, 0, i, true)

	}

	// 遍历其他轮上的卡片
	for x := 1; x < 5; x++ {
		for y := 0; y < 3; y++ {
			compareCard := d.Rolls[x][y]
			for card := range d.CardSerialRollCount {

				if compareCard == Richer && card != Treasure {
					d.setCardStatisticMap(x, y, card, card) // 万能卡（ID=0）替代其他卡片
					continue
				}

				if _, ok := d.CardSerialRollCount[card]; ok {
					d.setCardStatisticMap(x, y, compareCard, card) // 其他卡片
					continue
				}
			}
		}
	}
	return d
}

// setCardStatisticMap 设置游戏的统计map
func (d *Card) setCardStatisticMap(x, y int, compareCard, cardVal CardVal) {
	if compareCard != cardVal { // 两张卡不匹配
		return
	}

	if count, ok := d.CardSerialRollCount[cardVal]; !ok || count < x { // 不存 或者 不是连续的，那么不在统计
		return
	}

	d.CardSerialRollCount[cardVal] = x + 1 // 一个轮里有多个时，只记录一次。
	d.CardCounts[cardVal][x]++
	//d.cardIdx[cardVal] = append(d.cardIdx[cardVal], PosIdx{x, y})
	d.setScreenHit(cardVal, x, y, true)
}

// CalcBaseScoreAndFreeGame 计算初始局积分积分和获得的免费游戏次数
func (d *Card) CalcBaseScoreAndFreeGame(odds map[CardVal]map[int]int) (int, map[CardVal]int, int) {
	data := make(map[CardVal]int)

	for cardVal, rollCount := range d.CardSerialRollCount {
		if rollCount >= 3 { // 连续大于3轮才能获得积分
			if cardVal == Treasure && rollCount != 5 {
				continue
			}
			combTotal := 1 // 多少种组合
			for _, count := range d.CardCounts[cardVal] {
				if count > 0 {
					combTotal *= count
				}
			}
			data[cardVal] = odds[cardVal][rollCount] * combTotal * d.BetMulti
		}
	}

	totalScore := 0
	for _, val := range data {
		totalScore += val
	}

	// 判断是否命中免费游戏
	freeGameNum := d.addFreeGame()
	d.BaseScore = totalScore

	return d.BaseScore, data, freeGameNum
}

func (d *Card) addFreeGame() int {
	if count, ok := d.CardSerialRollCount[Treasure]; ok && count == 5 {
		d.FreeGameNum += 12
		return d.FreeGameNum
	}
	return 0
}

func (d *Card) StartFreeGame() int {
	for i := 0; i < d.FreeGameNum; i++ {
		newCard := NewCard(d.BetMulti).DealCards().StatisticCard()
		d.FreeGameCard = append(d.FreeGameCard, newCard)

		freeGameScore, _, freeGameNum := newCard.CalcBaseScoreAndFreeGame(oddsMap)
		d.FreeGameNum += freeGameNum
		d.FreeGameTotalScore += freeGameScore
		fmt.Printf("🎮 第 %d 小局: 投入 %3d | 得分 %3d \n", i, d.BetMulti, freeGameScore)
	}
	return d.FreeGameTotalScore + d.BaseScore
}

func (d *Card) setScreenHit(cardVal CardVal, x, y int, res bool) {
	hits, ok := d.CardScreenHit[cardVal]
	if !ok {
		hits = make([][]bool, 5)
		for i := range hits {
			hits[i] = make([]bool, 3) // 每个滚筒有3个位置
		}
	}

	hits[x][y] = res
	d.CardScreenHit[cardVal] = hits
}

func (d *Card) ToSerializable() {
	//d.ExportCardCounts = make(map[int][]int)
	//d.ExportCardScreenHit = make(map[int][][]bool)
	//d.ExportCardSerialRollCount = make(map[string]int)
	//
	//// cardCounts
	//for k, v := range d.CardCounts {
	//	//sc.CardCounts[fmt.Sprintf("%d", k)] = v
	//	d.ExportCardCounts[int(k)] = v
	//}
	//
	//// cardScreenHit
	//for k, v := range d.CardScreenHit {
	//	//sc.ScreenHits[fmt.Sprintf("%d", k)] = v
	//	d.ExportCardScreenHit[int(k)] = v
	//}
	//
	//// cardSerialRollCount
	//for k, v := range d.CardSerialRollCount {
	//	d.ExportCardSerialRollCount[fmt.Sprintf("%d", k)] = v
	//}
}
