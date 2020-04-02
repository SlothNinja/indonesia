package indonesia

import (
	"encoding/gob"
	"errors"
	"html/template"
	"sort"

	"github.com/SlothNinja/color"
	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.RegisterName("Player", newPlayer())
}

type Player struct {
	*game.Player
	Log          GameLog
	Rupiah       int
	Bank         int
	Bid          int
	OpIncome     int
	CityCards    CityCards
	Technologies Technologies
	Slots        Slots

	cardsForCurrentEra        CityCards
	canPlaceCity              int
	newCityAreasForCurrentEra Areas
}

func (p *Player) Score() int {
	return p.Rupiah + p.Bank
}

func (p *Player) Game() *Game {
	return p.Player.Game().(*Game)
}

type Players []*Player

func (ps Players) allPassed() bool {
	for _, p := range ps {
		if !p.Passed {
			return false
		}
	}
	return true
}

func (p *Player) canAutoPass() bool { return false }

// sort.Interface interface
func (p Players) Len() int { return len(p) }

func (p Players) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

type ByScore struct{ Players }

func (this ByScore) Less(i, j int) bool {
	return this.Players[i].compareByScore(this.Players[j]) == game.LessThan
}

func (p *Player) compareByScore(player *Player) game.Comparison {
	switch {
	case p.Score() < player.Score():
		return game.LessThan
	case p.Score() > player.Score():
		return game.GreaterThan
	}
	return game.EqualTo
}

//func (g *Game) determinePlaces() []Players {
//	// sort players by score
//	players := g.Players()
//	sort.Sort(Reverse{ByScore{players}})
//	g.setPlayers(players)
//
//	places := make([]Players, 0)
//	for _, p := range g.Players() {
//		places = append(places, Players{p})
//	}
//	return places
//}

func (client Client) determinePlaces(c *gin.Context, g *Game) (contest.Places, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	// sort players by score
	players := g.Players()
	sort.Stable(Reverse{ByScore{players}})
	g.setPlayers(players)

	places := make(contest.Places, 0)
	for i, p1 := range g.Players() {
		rmap := make(contest.ResultsMap, 0)
		results := make(contest.Results, 0)
		for j, p2 := range g.Players() {
			r, err := client.Rating.For(c, p2.User(), g.Type)
			if err != nil {
				return nil, err
			}
			result := &contest.Result{
				GameID: g.ID(),
				Type:   g.Type,
				R:      r.R,
				RD:     r.RD,
			}
			switch c := p1.compareByScore(p2); {
			case i == j:
			case c == game.GreaterThan, c == game.EqualTo && i < j:
				result.Outcome = 1
				results = append(results, result)
			default:
				result.Outcome = 0
				results = append(results, result)
			}
		}
		rmap[p1.User().Key] = results
		places = append(places, rmap)
	}
	return places, nil
}

type ByTurnOrderBid struct{ Players }

func (this ByTurnOrderBid) Less(i, j int) bool {
	return this.Players[i].compareByTurnOrderBid(this.Players[j]) == game.LessThan
}

func (p *Player) compareByTurnOrderBid(player *Player) game.Comparison {
	switch {
	case p.TotalBid() < player.TotalBid():
		return game.LessThan
	case p.TotalBid() > player.TotalBid():
		return game.GreaterThan
	default:
		return game.EqualTo
	}
}

type Reverse struct{ sort.Interface }

func (r Reverse) Less(i, j int) bool { return r.Interface.Less(j, i) }

var NotFound = errors.New("Not Found")

func (p *Player) Init(gr game.Gamer) {
	p.SetGame(gr)

	g, ok := gr.(*Game)
	if !ok {
		return
	}

	for _, company := range p.Companies() {
		company.Init(g)
	}
}

func newPlayer() *Player {
	p := &Player{
		Rupiah: 100,
		Slots:  make(Slots, 5),
		Technologies: Technologies{
			BidMultiplierTech: 1,
			SlotsTech:         1,
			MergersTech:       1,
			ExpansionsTech:    1,
			HullTech:          1,
		},
	}
	for i := range p.Slots {
		p.Slots[i] = new(Slot)
	}
	p.Slots[0].Developed = true
	p.Player = game.NewPlayer()
	return p
}

//func (p *Player) BidMultiplier() int {
//	return p.Technologies[BidMultiplierTech]
//}

func (g *Game) addNewPlayer(u *user.User) {
	p := CreatePlayer(g, u)
	g.Playerers = append(g.Playerers, p)
}

func CreatePlayer(g *Game, u *user.User) *Player {
	p := newPlayer()
	p.SetID(int(len(g.Players())))
	p.SetGame(g)

	colorMap := g.DefaultColorMap()
	p.SetColorMap(make(color.Colors, g.NumPlayers))

	for i := 0; i < g.NumPlayers; i++ {
		index := (i - p.ID()) % g.NumPlayers
		if index < 0 {
			index += g.NumPlayers
		}
		color := colorMap[index]
		p.ColorMap()[i] = color
	}

	return p
}

func (p *Player) beginningOfTurnReset() {
	p.clearActions()
}

func (g *Game) beginningOfPhaseReset() {
	g.SubPhase = NoSubPhase
	for _, p := range g.Players() {
		p.clearActions()
		p.Passed = false
		p.Bid = NoBid
	}
}

func (p *Player) clearActions() {
	p.PerformedAction = false
	p.Log = make(GameLog, 0)
}

func (p *Player) endOfTurnUpdate() {
	p.PerformedAction = false
}

//var playerValues = sslice{"Player.Passed", "Player.PerformedAction", "Player.Score",
//	"Bid", "Rupiah", "Bank", "OpIncome",
//	"Slots.0.Developed", "Slots.1.Developed", "Slots.2.Developed", "Slots.3.Developed", "Slots.4.Developed"}
//
//func adminPlayer(g *Game, form url.Values) (string, game.ActionType, error) {
//	if err := g.adminUpdatePlayer(playerValues); err != nil {
//		return "indonesia/flash_notice", game.None, err
//	}
//
//	return "", game.Save, nil
//}
//
//func (g *Game) adminUpdatePlayer(ss sslice) error {
//	if err := g.validateAdminAction(); err != nil {
//		return err
//	}
//
//	values, err := g.getValues()
//	if err != nil {
//		return err
//	}
//
//	p := g.SelectedPlayer()
//	//	g.debugf("Values: %#v", values)
//	c0Index, c1Index, c2Index, c3Index, c4Index := -10, -10, -10, -10, -10
//	for key := range values {
//		switch {
//		case key == "Technologies":
//			for tech, valueS := range values[key] {
//				if value, err := strconv.Atoi(valueS); err != nil {
//					return err
//				} else {
//					t := Technology(tech + 1)
//					if _, ok := p.Technologies[t]; ok && t < 1 || t > 5 {
//						delete(p.Technologies, t)
//					} else {
//						p.Technologies[t] = value
//					}
//				}
//			}
//			delete(values, key)
//		case key == "Slots.0.Company":
//			// v := values.Get(key)
//			//			g.debugf("Slots.0.Company: %v", v)
//			if index, err := strconv.Atoi(values.Get(key)); err == nil {
//				c0Index = index
//				//				g.debugf("c0Index: %d", c0Index)
//			}
//			delete(values, key)
//		case key == "Slots.1.Company":
//			if index, err := strconv.Atoi(values.Get(key)); err == nil {
//				c1Index = index
//			}
//			delete(values, key)
//		case key == "Slots.2.Company":
//			if index, err := strconv.Atoi(values.Get(key)); err == nil {
//				c2Index = index
//			}
//			delete(values, key)
//		case key == "Slots.3.Company":
//			if index, err := strconv.Atoi(values.Get(key)); err == nil {
//				c3Index = index
//			}
//			delete(values, key)
//		case key == "Slots.4.Company":
//			if index, err := strconv.Atoi(values.Get(key)); err == nil {
//				c4Index = index
//			}
//			delete(values, key)
//		default:
//			if !ss.include(key) {
//				delete(values, key)
//			}
//		}
//	}
//	if err := schema.Decode(p, values); err != nil {
//		return err
//	}
//	//	g.debugf("indices: %d %d %d %d %d", c0Index, c1Index, c2Index, c3Index, c4Index)
//	if c0Index == -1 {
//		p.Slots[0].Company = nil
//	} else if c0Index != -10 {
//		company := g.Companies()[c0Index]
//		//		g.debugf("Slot 1 Company: %s", company)
//		p.Slots[0].Company = company
//	}
//	if c1Index == -1 {
//		p.Slots[1].Company = nil
//	} else if c1Index != -10 {
//		company := g.Companies()[c1Index]
//		//		g.debugf("Slot 2 Company: %s", company)
//		p.Slots[1].Company = company
//	}
//	if c2Index == -1 {
//		p.Slots[2].Company = nil
//	} else if c2Index != -10 {
//		company := g.Companies()[c2Index]
//		//		g.debugf("Slot 3 Company: %s", company)
//		p.Slots[2].Company = company
//	}
//	if c3Index == -1 {
//		p.Slots[3].Company = nil
//	} else if c3Index != -10 {
//		company := g.Companies()[c3Index]
//		//		g.debugf("Slot 4 Company: %s", company)
//		p.Slots[3].Company = company
//	}
//	if c4Index == -1 {
//		p.Slots[4].Company = nil
//	} else if c4Index != -10 {
//		company := g.Companies()[c4Index]
//		//		g.debugf("Slot 5 Company: %s", company)
//		p.Slots[4].Company = company
//	}
//	return nil
//}
//
//func adminPlayerNewCompany(g *Game, form url.Values) (string, game.ActionType, error) {
//	if err := g.validateAdminAction(); err != nil {
//		return "indonesia/flash_notice", game.None, err
//	}
//
//	values, err := g.getValues()
//	if err != nil {
//		return "indonesia/flash_notice", game.None, err
//	}
//
//	p := g.SelectedPlayer()
//	//	g.debugf("Values: %#v", values)
//	var d *Deed
//	slot := -1
//	for key := range values {
//		switch key {
//		case "Slot":
//			if v, err := strconv.Atoi(values.Get(key)); err == nil {
//				slot = v
//			}
//		case "Deed":
//			if v := values.Get(key); v != "none" {
//				d = g.Deeds().get(v)
//				g.CTX().Debugf("Deed: %#v", d)
//			}
//		}
//	}
//
//	if slot != -1 && d != nil {
//		p.Slots[slot].Company = newCompany(g, p, slot, d)
//	}
//	return "", game.Save, nil
//}

func (p *Player) CanClick(a *Area) bool {
	if p == nil {
		return false
	}
	switch g := p.Game(); {
	case !p.User().IsAdminOrCurrent(g.CTX()):
		return false
	case g.Phase == NewEra:
		return p.canClickNewEra(a)
	case g.Phase == Acquisitions && (g.SubPhase == AQInitialProduction || g.SubPhase == AQInitialShip):
		return p.canClickAcquisitions(a)
	case g.Phase == Operations && (g.SubPhase == OPExpansion || g.SubPhase == OPFreeExpansion):
		if p.canExpandShipping() {
			return g.freeShippingExpansionAreas().include(a)
		} else {
			return g.SelectedCompany().ExpansionAreas().include(a)
		}
	default:
		return false
	}
}

func (p *Player) CanSelectCompany(c *Company) bool {
	switch g := p.Game(); {
	case g.Phase == Operations && g.SubPhase == OPSelectCompany && !c.Operated:
		return p.Companies().include(c)
	case g.Phase == Mergers:
		return p.canSelectFirstCompany(c) || p.canSelectSecondCompany(c)
	default:
		return false
	}
}

func (p *Player) CanSelectCard() bool {
	return p != nil && p.Game().Phase == NewEra && p.Game().SubPhase == NESelectCard && !p.PerformedAction
}

func (p *Player) canExpandShipping() bool {
	g := p.Game()
	company := g.SelectedCompany()
	return g.Phase == Operations && g.SubPhase == OPFreeExpansion &&
		company != nil && company.Goods() == Shipping &&
		company.Ships() < company.MaxShips() &&
		g.Expansions < p.Technologies[ExpansionsTech] &&
		!p.PerformedAction
}

func (p *Player) CanPlaceCity() bool {
	if p == nil {
		return false
	}

	switch p.canPlaceCity {
	case 1:
		return false
	case 2:
		return true
	}

	g := p.Game()
	if g.Phase == NewEra &&
		!p.PerformedAction && g.CityStones[0] > 0 &&
		len(p.NewCityAreasForCurrentEra()) > 0 {
		p.canPlaceCity = 2
		return true
	} else {
		p.canPlaceCity = 1
		return false
	}
}

func (p *Player) CanBid() bool {
	return p != nil && p.Game().Phase == BidForTurnOrder &&
		!p.PerformedAction
}

func (p *Player) CanAcquireCompany() bool {
	return p != nil && p.Game().Phase == Acquisitions && p.Game().SubPhase == NoSubPhase &&
		len(p.Game().AvailableDeeds) > 0 && !p.PerformedAction && p.hasEmptySlot()
}

func (p *Player) canPlaceInitialProduct() bool {
	g := p.Game()
	return g.Phase == Acquisitions &&
		g.SubPhase == AQInitialProduction &&
		!p.PerformedAction
}

func (p *Player) canPlaceInitialShip() bool {
	g := p.Game()
	return g.Phase == Acquisitions &&
		g.SubPhase == AQInitialShip &&
		!p.PerformedAction
}

func (p *Player) CanResearch() bool {
	return p != nil && p.Game().Phase == Research && !p.PerformedAction
}

func (p *Player) CanExpandProduction() bool {
	g := p.Game()
	company := g.SelectedCompany()
	return g.Phase == Operations && (g.SubPhase == OPFreeExpansion || g.SubPhase == OPExpansion) &&
		company != nil && company.Goods() != Shipping &&
		!p.PerformedAction
}

func (p *Player) HasCompanyToOperate() bool {
	for _, c := range p.Companies() {
		if !c.Operated {
			return true
		}
	}
	return false
}

func (ps Players) anyCanPlaceCity() bool {
	for _, p := range ps {
		if p.CanPlaceCity() {
			return true
		}
	}
	return false
}

func (p *Player) canClickNewEra(a *Area) bool {
	return p.CanPlaceCity() && p.NewCityAreasForCurrentEra().include(a)
}

func (p *Player) NewCityAreasForCurrentEra() Areas {
	cards := p.CardsForCurrentEra()
	if len(p.newCityAreasForCurrentEra) > 1 {
		return p.newCityAreasForCurrentEra
	}
	p.newCityAreasForCurrentEra = p.Game().newCityAreasFor(cards...)
	return p.newCityAreasForCurrentEra
}

func (p *Player) canClickAcquisitions(a *Area) bool {
	if p == nil {
		return false
	}
	g := p.Game()
	company := g.SelectedCompany()
	return a != nil &&
		(g.SubPhase == AQInitialProduction &&
			a.Producer == nil &&
			a.City == nil &&
			a.IsLand() &&
			company != nil &&
			company.Deeds != nil &&
			company.Deeds[0].Province == a.Province()) ||
		(g.SubPhase == AQInitialShip &&
			company != nil &&
			company.Deeds != nil &&
			a.IsSea() &&
			a.adjacentToProvince(company.Deeds[0].Province))
}

func (p *Player) CanSelectCompanyToOperate() bool {
	if p == nil {
		return false
	}
	return p != nil && p.Game().Phase == Operations && p.Game().SubPhase == OPSelectCompany && !p.PerformedAction
}

func (p *Player) CanSelectGood() bool {
	if p == nil {
		return false
	}
	return p != nil && p.Game().Phase == Operations && p.Game().SubPhase == OPSelectProductionArea &&
		!p.PerformedAction
}

func (p *Player) CanSelectShip() bool {
	if p == nil {
		return false
	}
	g := p.Game()
	return g.Phase == Operations && g.SubPhase == OPSelectShip
}

func (p *Player) CanClickGoodsIn(a *Area) bool {
	if p == nil {
		return false
	}
	g := p.Game()
	switch phase, subphase := g.Phase, g.SubPhase; {
	case phase == Operations && subphase == OPSelectProductionArea:
		company := g.SelectedCompany()
		return company != nil && !a.Used && company.Areas().include(a)
	case phase == Mergers && subphase == MSiapFajiCreation:
		if g.SiapFajiMerger == nil {
			return false
		}
		p := g.PlayerByID(g.SiapFajiMerger.OwnerID)
		if p == nil {
			return false
		}
		index := g.SiapFajiMerger.OwnerSlot - 1
		if index < 0 || index > 4 {
			return false
		}
		slot := p.Slots[index]
		if slot == nil {
			return false
		}
		return slot.Company != nil && slot.Company.Areas().include(a)
	default:
		return false
	}
}

func (p *Player) CanClickShipOf(s *Shipper) bool {
	if p == nil {
		return false
	}
	g := p.Game()
	a := g.SelectedArea()
	return a != nil && g.Phase == Operations &&
		(g.SubPhase == OPSelectShip || g.SubPhase == OPSelectCityOrShip) && s.hasCapacity() &&
		(g.SelectedShippingCompany() == nil || g.SelectedShippingCompany() == s.Company()) &&
		((a.IsLand() && s.a.adjacentToZoneFor(a)) ||
			(a.IsSea() && s.a.adjacentToArea(a)))
}

func (p *Player) CanClickCityIn(a *Area) bool {
	if p == nil {
		return false
	}
	g := p.Game()
	company := g.SelectedCompany()
	return a.City != nil && company != nil && company.IsProductionCompany() &&
		a.City.Delivered[company.Goods()] < a.City.Size && p.CanSelectCityOrShip() &&
		a.adjacentToArea(g.SelectedArea())
}

func (p *Player) CanSelectCityOrShip() bool {
	if p == nil {
		return false
	}
	g := p.Game()
	return g.Phase == Operations && g.SubPhase == OPSelectCityOrShip
}

func (p *Player) cardsForEra(era Era) CityCards {
	if len(p.cardsForCurrentEra) != 0 {
		return p.cardsForCurrentEra
	}
	var cards CityCards
	for _, card := range p.CityCards {
		if card.Era == era {
			cards = append(cards, card)
		}
	}
	p.cardsForCurrentEra = cards
	return cards
}

func (p *Player) CardsForCurrentEra() CityCards {
	return p.cardsForEra(p.Game().Era)
}

func (p *Player) DisplayHand() template.HTML {
	s := restful.HTML("<div id='player-hand-%d'>", p.ID())
	for _, card := range p.CityCards {
		name := card.IDString()
		s += restful.HTML("<div>")
		s += restful.HTML("<img class='card' src='/images/indonesia/city-card-%s.png'/>", name)
		s += restful.HTML("</div>")
	}
	s += restful.HTML("</div>")
	return s
}

func (p *Player) PlayCardDisplay() template.HTML {
	s := restful.HTML("")
	for i, card := range p.CardsForCurrentEra() {
		s += restful.HTML("<div>")
		s += restful.HTML("<img id='card-%d' class='top-padding clickable card' src='/images/indonesia/city-card-%s.png'/>",
			i, card.IDString())
		s += restful.HTML("</div>")
	}
	return s
}

func (p *Player) SelectCompanyDisplay() template.HTML {
	s := restful.HTML("")
	for _, company := range p.Companies() {
		if !company.Operated {
			s += restful.HTML("<div class='pull-left'>")
			s += restful.HTML("<div class='top-padding center' style='width:100px'>Slot %d</div>", company.Slot)
			s += restful.HTML("<div id='company-%d' class='clickable card' style='padding:3px;border:3px solid yellow'>", company.Slot)
			deed := company.Deeds[0]
			s += restful.HTML("<img class='deed' src='/images/indonesia/%s.png'/>", deed.IDString())
			s += restful.HTML("</div></div>")
		}
	}
	return s
}

var bidMultiplier = map[int]int{1: 1, 2: 5, 3: 25, 4: 100, 5: 400}

func (p *Player) Multiplier() int {
	return bidMultiplier[p.Technologies[BidMultiplierTech]]
}

func (p *Player) TotalBid() int {
	return p.Bid * p.Multiplier()
}

func (p *Player) HasSlots(i int) bool {
	return p.Technologies[SlotsTech] >= i
}

func (p *Player) Companies() Companies {
	var companies Companies
	for _, slot := range p.Slots {
		if slot.Developed && slot.Company != nil {
			companies = append(companies, slot.Company)
		}
	}
	return companies
}

func (p *Player) hasShippingCompany() bool {
	for _, c := range p.Companies() {
		if c.IsShippingCompany() {
			return true
		}
	}
	return false
}
