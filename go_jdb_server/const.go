package main

// currentState Enum
const (
	StartGameState = 0
	BetState       = 1
	FreeGameState  = 2
	EndState       = 3
)

type StateStruct struct {
	GameStateType string
}

var StateMap = map[int]StateStruct{
	StartGameState: {
		GameStateType: "GS_001",
	},
	BetState: {
		GameStateType: "GS_003",
	},
	FreeGameState: {
		GameStateType: "GS_095",
	},
	EndState: {
		GameStateType: "GS_002",
	},
}

var oddsMap = map[CardVal]map[int]int{
	Treasure: {
		3: 0,
		4: 0,
		5: 0,
	},
	AirPlane: {
		3: 75,
		4: 150,
		5: 400,
	},
	Yacht: {
		3: 50,
		4: 150,
		5: 300,
	},
	SportCar: {
		3: 40,
		4: 100,
		5: 250,
	},
	Motorcycle: {
		3: 30,
		4: 100,
		5: 200,
	},
	Ace: {
		3: 15,
		4: 30,
		5: 125,
	},
	King: {
		3: 15,
		4: 30,
		5: 125,
	},
	Queen: {
		3: 10,
		4: 20,
		5: 100,
	},
	Jack: {
		3: 10,
		4: 20,
		5: 100,
	},
}
