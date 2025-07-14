package main

import "math/rand"

var WheelData = [][]WheelSlot{

	{
		{
			WheelLength: 48,
			NoWinIndex:  []int{0},
			WheelData:   []int{3, 7, 4, 0, 5, 6, 0, 8, 8, 4, 0, 7, 6, 5, 0, 3, 0, 7, 6, 4, 3, 5, 0, 8, 8, 3, 4, 7, 0, 6, 5, 0, 4, 7, 6, 0, 3, 5, 0, 8, 0, 6, 8, 7, 0, 5, 4, 3},
		},
		{
			WheelLength: 14,
			NoWinIndex:  []int{0},
			WheelData:   []int{8, 5, 0, 4, 6, 3, 1, 7, 0, 8, 6, 6, 6, 7},
		},
		{
			WheelLength: 48,
			NoWinIndex:  []int{0},
			WheelData:   []int{8, 3, 4, 7, 0, 6, 5, 0, 4, 7, 6, 0, 3, 5, 0, 8, 0, 6, 8, 7, 0, 5, 4, 3, 3, 7, 4, 0, 5, 6, 0, 8, 8, 4, 0, 7, 6, 5, 0, 3, 0, 7, 6, 4, 3, 5, 0, 8},
		},
	},
	{
		{
			WheelLength: 1,
			NoWinIndex:  []int{0},
			WheelData:   []int{0},
		},
		{
			WheelLength: 35,
			NoWinIndex:  []int{0},
			WheelData:   []int{6, 1, 8, 3, 1, 8, 7, 1, 4, 1, 8, 6, 7, 1, 5, 6, 7, 1, 5, 8, 7, 1, 8, 6, 1, 8, 7, 1, 7, 5, 1, 8, 6, 0, 6},
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
}

// 符号赔付
var payTable = [][]int{
	{0, 0, 200},
	{0, 0, 0},
	{0, 0, 0},
	{0, 0, 30},
	{0, 0, 20},
	{0, 0, 15},
	{0, 0, 8},
	{0, 0, 5},
	{0, 0, 2},
	{0, 0, 0},
}

func GetRoundWin(screen [][]int) (int, int) {
	if len(screen) != 3 || len(screen[0]) == 0 || len(screen[1]) == 0 || len(screen[2]) == 0 {
		return 0, -1 // 防御性判断
	}

	a, b, c := screen[0][0], screen[1][0], screen[2][0]

	// 判断是否三个元素一样（考虑百搭0）
	if match, value := isThreeMatch(a, b, c); match {

		return payTable[value][2], value // 返回三连的具体数字 对应赔付表的值
	}

	return 0, -1 // 没匹配成功
}

// 判断是否三个相等，允许0作为任意符号
func isThreeMatch(a, b, c int) (bool, int) {
	// 统计非0符号
	nonZero := []int{}
	for _, v := range []int{a, b, c} {
		if v != 0 {
			nonZero = append(nonZero, v)
		}
	}

	if len(nonZero) == 0 {
		return true, 0 // 全是百搭，默认算作“3个0”
	}

	first := nonZero[0]
	for _, v := range nonZero {
		if v != first {
			return false, -1 // 非百搭符号不一致
		}
	}
	return true, first
}

func ReSpinClumnTwo(issuper bool) (int, []int) {

	column1 := WheelData[0][1].WheelData
	if issuper {
		column1 = WheelData[2][1].WheelData
	}
	wheel := column1
	index := rand.Intn(len(wheel))
	indexes := index
	screen := wheel[index]
	length := len(wheel)
	dampInfos := []int{
		wheel[(indexes-1+length)%length],
		wheel[(indexes+1)%length],
	}

	return screen, dampInfos
}

func getSpecialDataWithDampInfo(CanWin bool) ([][]int, [][]int) {
	screen := make([][]int, 3)
	for i := range screen {
		screen[i] = make([]int, 1)
	}
	dampInfos := make([][]int, 3)
	column1 := WheelData[2][1].WheelData

	// 首次抽取

	index := rand.Intn(len(column1))
	indexes := index
	screen[1][0] = column1[index]

	// 如果不允许中奖 且 当前组合可能中奖 → 修改其中一列打散
	if !CanWin {

		wheel := column1
		// oldIndex := indexes
		// oldSymbol := wheel[oldIndex]

		// 保证替换值不同（避免同值）
		for {
			newIndex := rand.Intn(len(wheel))
			newSymbol := wheel[newIndex]
			if 2 == newSymbol {
				indexes = newIndex
				screen[1][0] = newSymbol
				break
			}
		}
	}

	// 构造 damp 信息（前后符号）

	idx := indexes
	length := len(column1)
	dampInfos[1] = []int{
		column1[(idx-1+length)%length],
		column1[(idx+1)%length],
	}
	dampInfos[0] = []int{
		0, 0,
	}
	dampInfos[2] = []int{
		0, 0,
	}
	screen[0][0] = 0
	screen[2][0] = 0
	return screen, dampInfos
}

func GetSreenResult(CanWin bool) ([][]int, [][]int) {
	return getRandomDataWithDampInfo(CanWin)
}

func getRandomDataWithDampInfo(CanWin bool) ([][]int, [][]int) {
	screen := make([][]int, 3)
	for i := range screen {
		screen[i] = make([]int, 1)
	}

	dampInfos := make([][]int, 3)

	column0 := WheelData[0][0].WheelData
	column1 := WheelData[0][1].WheelData
	column2 := WheelData[0][2].WheelData
	columns := [][]int{column0, column1, column2}

	indexes := make([]int, 3) // 保存每列的随机下标

	// 首次抽取
	for i := 0; i < 3; i++ {
		wheel := columns[i]
		index := rand.Intn(len(wheel))
		indexes[i] = index
		screen[i][0] = wheel[index]
	}

	// 如果不允许中奖 且 当前组合可能中奖 → 修改其中一列打散
	if !CanWin && violatesNoWinRule(screen) {
		// 随机选择一列修改
		col := rand.Intn(3)
		wheel := columns[col]
		oldIndex := indexes[col]
		oldSymbol := wheel[oldIndex]

		// 保证替换值不同（避免同值）
		for {
			newIndex := rand.Intn(len(wheel))
			newSymbol := wheel[newIndex]
			if newIndex != oldIndex && oldSymbol != newSymbol {
				indexes[col] = newIndex

				// 50% 概率强制设置为 symbol 9
				if rand.Float64() < 0.5 {
					screen[col][0] = 9
				} else {
					screen[col][0] = newSymbol
				}
				break
			}
		}
	}

	// 构造 damp 信息（前后符号）
	for i := 0; i < 3; i++ {
		wheel := columns[i]
		idx := indexes[i]
		length := len(wheel)
		dampInfos[i] = []int{
			wheel[(idx-1+length)%length],
			wheel[(idx+1)%length],
		}
	}
	// screen = [][]int{{0}, {1}, {0}}
	return screen, dampInfos
}

// 判断不能中奖规则
func violatesNoWinRule(screen [][]int) bool {
	a, b, c := screen[0][0], screen[1][0], screen[2][0]

	if a == b && b == c {
		return true
	}

	wildCount := 0
	for _, col := range screen {
		if col[0] == 0 {
			wildCount++
		}
	}
	if wildCount >= 2 {
		return true
	}

	if wildCount == 1 {
		if (a == 0 && b == c) || (b == 0 && a == c) || (c == 0 && a == b) {
			return true
		}
	}

	if a == 0 && c == 0 && b == 1 {
		return true
	}

	return false
}
