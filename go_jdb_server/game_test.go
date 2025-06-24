package main

import (
	"fmt"
	"testing"
)

func TestGame(t *testing.T) {
	card := NewCard(20)
	card.DealCards()                                                   // 发卡
	card.StatisticCard()                                               // 统计卡
	totalScore, cardScore, _ := card.CalcBaseScoreAndFreeGame(oddsMap) // 通过配置计算得分

	fmt.Println("totalScore:", totalScore, "cardScore:", cardScore)
}

func TestCreate(t *testing.T) {
	generateSpinData(96.0) // 设置目标 RTP 为 96%
}
