package indonesia

import "github.com/SlothNinja/game"

const (
	// Phases of Games
	NoPhase game.Phase = iota * 100
	AnnounceWinners
	GameOver
	EndGame
	AwaitPlayerInput
	Setup
	StartGame

	// Game Specific Phases
	NewEra
	BidForTurnOrder
	Mergers
	Acquisitions
	Research
	Operations
	CompanyExpansion
	CityGrowth
)

var PhaseNames = game.PhaseNameMap{
	NoPhase:          "None",
	AnnounceWinners:  "Announce Winners",
	GameOver:         "Game Over",
	EndGame:          "End Of Game",
	AwaitPlayerInput: "Await Player Input",
	Setup:            "Setup",
	StartGame:        "Start Game",

	NewEra:           "New Era",
	BidForTurnOrder:  "Bid For Turn Order",
	Mergers:          "Mergers",
	Acquisitions:     "Acquisitions",
	Research:         "Research",
	Operations:       "Operations",
	CompanyExpansion: "Company Expansion",
	CityGrowth:       "City Growth",
}

func (g *Game) PhaseNames() game.PhaseNameMap {
	return PhaseNames
}

func (g *Game) PhaseName() string {
	return PhaseNames[g.Phase]
}

// NoPhase
const (
	NoSubPhase game.SubPhase = iota
)

// New Era (NE)
const (
	NESelectCard game.SubPhase = game.SubPhase(NewEra + iota + 1)
)

// Bid For Turn Order

// Mergers SubPhases (M)
const (
	MSelectCompany1 game.SubPhase = game.SubPhase(Mergers + iota + 1)
	MSelectCompany2
	MBid
	MResolution
	MSiapFajiCreation
)

// Acquisitions SubPhases (AQ)
const (
	AQInitialProduction game.SubPhase = game.SubPhase(Acquisitions + iota + 1)
	AQInitialShip
)

// Research (R)
const (
	RSelectPlayer game.SubPhase = game.SubPhase(Research + iota + 1)
)

// Operations SubPhases (OP)
const (
	OPSelectCompany game.SubPhase = game.SubPhase(Operations + iota + 1)
	OPReceiveIncome
	OPExpansion
	OPFreeExpansion
	OPSelectProductionArea
	OPSelectShip
	OPSelectCityOrShip
	OPSelectGoods
)

// City Growth

var SubPhaseNames = game.SubPhaseNameMap{
	NoSubPhase:             "None",
	NESelectCard:           "New Era: Select Card",
	MSelectCompany1:        "Mergers: Select First Company",
	MSelectCompany2:        "Mergers: Select Second Company",
	MBid:                   "Mergers: Bid",
	MResolution:            "Mergers: Resolution",
	MSiapFajiCreation:      "Mergers: Siap Faji Creation",
	AQInitialProduction:    "Acquisitions: Production",
	AQInitialShip:          "Acquisitions: Ship",
	RSelectPlayer:          "Resolution: Select Player",
	OPSelectCompany:        "Operations: Selected Company",
	OPReceiveIncome:        "Operations: Receive Income",
	OPExpansion:            "Operations: Expansion",
	OPFreeExpansion:        "Operations: Free Expansion",
	OPSelectProductionArea: "Operations: Select Production Area",
	OPSelectShip:           "Operations: Select Ship",
	OPSelectCityOrShip:     "Operations: Select City Or Ship",
	OPSelectGoods:          "Operations: Select Goods",
}

func (g *Game) SubPhaseNames() game.SubPhaseNameMap {
	return SubPhaseNames
}

func (g *Game) SubPhaseName() string {
	return SubPhaseNames[g.SubPhase]
}
