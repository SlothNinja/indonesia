package indonesia

import (
	"encoding/gob"
	"html/template"

	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.Register(new(selectCompanyEntry))
	gob.Register(new(deliveredGoodEntry))
	gob.Register(new(receiveIncomeEntry))
	gob.Register(make(ShipperIncomeMap, 0))
	gob.Register(new(expandProductionEntry))
	gob.Register(new(expandShippingEntry))
	gob.Register(new(stopExpandingEntry))
}

type ShipperIncomeMap map[int]int

func (m ShipperIncomeMap) OtherShips(pid int) int {
	ships := 0
	for id, s := range m {
		if id != pid {
			ships += s
		}
	}
	return ships
}

func (m ShipperIncomeMap) OwnShips(pid int) int {
	ships := 0
	for id, s := range m {
		if id == pid {
			ships += s
		}
	}
	return ships
}

func (client Client) startOperations(c *gin.Context, g *Game) (contest.Contests, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	np := g.companyExpansionNextPlayer()
	if np == nil {
		return client.startCityGrowth(c, g)
	}

	g.beginningOfPhaseReset()
	g.Phase = Operations
	g.SubPhase = OPSelectCompany
	g.resetShipping()
	g.resetOpIncome()
	g.setCurrentPlayers(np)
	g.OverrideDeliveries = -1
	return nil, nil
}

func (g *Game) resetOpIncome() {
	for _, p := range g.Players() {
		p.OpIncome = 0
	}
}

func (g *Game) selectCompany(c *gin.Context) (string, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	com, err := g.validateSelectCompany(c)
	switch {
	case err != nil:
		return "indonesia/flash_notice", err
	case com.IsShippingCompany():
		g.SubPhase = OPFreeExpansion
		if cp := g.CurrentPlayer(); cp.canExpandShipping() {

			// Log
			e := g.newSelectCompanyEntryFor(g.CurrentPlayer(), com, 0)
			restful.AddNoticef(c, string(e.HTML(c)))
			return "indonesia/select_company_update", nil
		} else {
			cp.PerformedAction = true
			com.Operated = true
			e := g.newSelectCompanyEntryFor(g.CurrentPlayer(), com, 0)
			restful.AddNoticef(c, string(e.HTML(c)))
			return "indonesia/completed_expansion_dialog", nil
		}
	default:
		e := g.newSelectCompanyEntryFor(g.CurrentPlayer(), com, 0)
		restful.AddNoticef(c, string(e.HTML(c)))
		if g.OverrideDeliveries > -1 {
			g.RequiredDeliveries = g.OverrideDeliveries
		} else {
			g.RequiredDeliveries, g.ProposedPath = com.maxFlow()
		}
		if g.RequiredDeliveries > 0 {
			g.SubPhase = OPSelectProductionArea
			g.ShipperIncomeMap = make(ShipperIncomeMap, 0)
			return "indonesia/select_company_update", nil
		} else {
			return g.startCompanyExpansion(c), nil
		}
	}
}

func (g *Game) validateSelectCompany(c *gin.Context) (*Company, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	com, err := g.SelectedCompany(), g.validatePlayerAction(c)
	if err != nil {
		return nil, err
	}
	if com == nil {
		return nil, sn.NewVError("Missing company selection.")
	}
	return com, nil
}

type selectCompanyEntry struct {
	*Entry
	Company Company
	Deliver int
}

func (g *Game) newSelectCompanyEntryFor(p *Player, c *Company, d int) (e *selectCompanyEntry) {
	e = &selectCompanyEntry{
		Entry:   g.newEntryFor(p),
		Company: *c,
		Deliver: d,
	}
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (e *selectCompanyEntry) HTML(c *gin.Context) template.HTML {
	g := gameFrom(c)
	company := e.Company
	name := g.NameByPID(e.PlayerID)
	return restful.HTML("<div>%s selected the %s company to operate.</div>", name, company.String())
}

func (g *Game) selectGood(c *gin.Context) (tmpl string, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	var a *Area
	if a, err = g.validateSelectGood(c); err != nil {
		tmpl = "indonesia/flash_notice"
		return
	}

	g.SubPhase = OPSelectShip
	g.SelectedGoodsAreaID = a.ID
	g.SelectedShippingProvince = NoProvince
	a.Used = true
	from := sourceFID
	to := toFlowID(a.ID)
	g.CustomPath = g.CustomPath.addFlow(from, to)
	tmpl = "indonesia/select_good_update"
	return
}

const (
	shipInput int = iota + 1
	shipOutput
)

func (fp flowMatrix) addFlow(source, target FlowID) flowMatrix {
	var flowPath flowMatrix
	if fp == nil {
		flowPath = make(flowMatrix, 0)
	} else {
		flowPath = fp
	}
	if flowPath[source] == nil {
		flowPath[source] = make(subflow, 0)
	}
	if flowPath[target] == nil {
		flowPath[target] = make(subflow, 0)
	}
	flowPath[source][target] += 1
	flowPath[target][source] -= 1
	return flowPath
}

func (g *Game) validateSelectGood(c *gin.Context) (*Area, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	err := g.validatePlayerAction(c)
	if err != nil {
		return nil, err
	}

	com := g.SelectedCompany()
	a := g.SelectedArea()
	zone := com.ZoneFor(a)

	switch {
	case com == nil:
		return nil, sn.NewVError("You must select company to operate.")
	case a == nil:
		return nil, sn.NewVError("You must select a good area.")
	case zone == nil:
		return nil, sn.NewVError("You must select a good in a production zone of the company.")
	case a.Used:
		return nil, sn.NewVError("The selected area has already delivered its goods.")
	default:
		return a, nil
	}
}

const InvalidUsedShips = -1

func (g *Game) selectShip(c *gin.Context) (tmpl string, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	var (
		old, area *Area
		shipper   *Shipper
		incomeMap ShipperIncomeMap
	)

	if old, area, shipper, incomeMap, err = g.validateSelectShip(c); err != nil {
		tmpl = "indonesia/flash_notice"
		return
	}
	area.Used = true
	shipper.Delivered += 1
	if g.ShipsUsed == InvalidUsedShips {
		g.ShipsUsed = 1
	} else {
		g.ShipsUsed += 1
	}

	province := shipper.Province()
	g.SelectedShippingProvince = province

	incomeMap[shipper.OwnerID] += 1

	g.SelectedAreaID, g.SelectedArea2ID, g.OldSelectedAreaID = area.ID, NoArea, old.ID
	fromAID, fromFID := old.ID, FlowID{AreaID: old.ID}
	toFID := FlowID{AreaID: area.ID}
	if old.IsSea() {
		inputFID := FlowID{
			AreaID:   fromAID,
			PID:      shipper.OwnerID,
			Index:    g.SelectedShipper2Index,
			IO:       shipInput,
			Province: province,
		}
		outputFID := FlowID{
			AreaID:   fromAID,
			PID:      shipper.OwnerID,
			Index:    g.SelectedShipper2Index,
			IO:       shipOutput,
			Province: province,
		}
		g.CustomPath = g.CustomPath.addFlow(inputFID, outputFID)
		fromFID = outputFID
	}
	if area.IsSea() {
		toFID = FlowID{
			AreaID:   area.ID,
			PID:      shipper.OwnerID,
			Index:    g.SelectedShipperIndex,
			IO:       shipInput,
			Province: province,
		}
	}
	g.SelectedShipper2Index = g.SelectedShipperIndex
	g.CustomPath = g.CustomPath.addFlow(fromFID, toFID)
	g.SubPhase = OPSelectCityOrShip
	tmpl = "indonesia/select_ship_update"
	return
}

func (g *Game) validateSelectShip(c *gin.Context) (*Area, *Area, *Shipper, ShipperIncomeMap, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	com := g.SelectedCompany()
	shipper := g.SelectedShipper()
	shippingCompany := g.SelectedShippingCompany()
	old, area := g.SelectedArea(), g.SelectedArea2()
	incomeMap := g.ShipperIncomeMap

	err := g.validatePlayerAction(c)
	switch {
	case err != nil:
		return nil, nil, nil, nil, err
	case g.Phase != Operations:
		return nil, nil, nil, nil, sn.NewVError("Expected %q phase, have %q phase.", PhaseNames[Operations], g.PhaseName())
	case !(g.SubPhase == OPSelectShip || g.SubPhase == OPSelectCityOrShip):
		return nil, nil, nil, nil, sn.NewVError("Expected %q or %q subphase, have %q subphase.",
			SubPhaseNames[OPSelectShip], SubPhaseNames[OPSelectCityOrShip], g.SubPhaseName())
	case c == nil:
		return nil, nil, nil, nil, sn.NewVError("You must select company to operate.")
	case g.ShipperIncomeMap == nil:
		return nil, nil, nil, nil, sn.NewVError("Missing temp value for income map.")
	case g.SubPhase == OPSelectShip &&
		(old == nil || area == nil || !com.ZoneFor(old).adjacentToArea(area)):
		return nil, nil, nil, nil, sn.NewVError("You must select a ship adjacent to the previously selected area.")
	case shipper == nil:
		return nil, nil, nil, nil, sn.NewVError("You must select a valid ship adjacent to the previously selected area.")
	case shipper.Delivered+1 > shipper.HullSize():
		return nil, nil, nil, nil, sn.NewVError("The selected ship has already reached its hull limit.")
	case shippingCompany != nil && !(shippingCompany.OwnerID == shipper.OwnerID && shippingCompany.Slot == shipper.Slot):
		return nil, nil, nil, nil, sn.NewVError("You must select a ship of the same shipping company.")
	default:
		return old, area, shipper, incomeMap, nil
	}
}

func (g *Game) selectCityOrShip(c *gin.Context) (tmpl string, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	switch area2 := g.SelectedArea2(); {
	case area2 == nil:
		tmpl = "indonesia/flash_notice"
		err = sn.NewVError("You must select an area having a city or boat.")
	case area2.IsLand():
		tmpl, err = g.selectCity(c)
	case area2.IsSea():
		tmpl, err = g.selectShip(c)
	default:
		tmpl = "indonesia/flash_notice"
		err = sn.NewVError("Unexpectant value for area received.")
	}
	return
}

func (g *Game) selectCity(c *gin.Context) (tmpl string, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	var (
		city                     *City
		company, shippingCompany *Company
		from, to                 Province
		used                     int
	)

	if city, company, from, to, shippingCompany, used, err = g.validateSelectCity(c); err != nil {
		tmpl = "indonesia/flash_notice"
		return
	}

	city.Delivered[company.Goods()] += 1
	inputFID := toFlowID(g.SelectedAreaID, shippingCompany.OwnerID, g.SelectedShipper2Index, shipInput,
		shippingCompany.Province().Int())
	outputFID := toFlowID(g.SelectedAreaID, shippingCompany.OwnerID, g.SelectedShipper2Index, shipOutput,
		shippingCompany.Province().Int())
	g.CustomPath = g.CustomPath.addFlow(inputFID, outputFID)

	inputFID, outputFID = outputFID, toFlowID(g.SelectedGoodsAreaID)
	g.CustomPath = g.CustomPath.addFlow(inputFID, outputFID)

	inputFID, outputFID = outputFID, targetFID
	g.CustomPath = g.CustomPath.addFlow(inputFID, outputFID)

	// Log
	e := g.newDeliveredGoodEntryFor(g.CurrentPlayer(), company.Goods(), from, to, shippingCompany.OwnerID, used)
	restful.AddNoticef(c, string(e.HTML(c)))
	if company.Delivered() == g.RequiredDeliveries {
		tmpl, err = g.receiveIncome(c)
	} else {
		g.SubPhase = OPSelectProductionArea
		g.resetShipper()
		tmpl = "indonesia/select_city_update"
	}
	return
}

func (g *Game) resetShipper() {
	g.ShippingCompanyOwnerID, g.ShippingCompanySlot, g.ShipsUsed = NoPlayerID, NoSlot, InvalidUsedShips
}

// func (g *Game) validateSelectCity(c *gin.Context) (city *City, c *Company, from Province, to Province, sc *Company, used int, err error) {
func (g *Game) validateSelectCity(c *gin.Context) (*City, *Company, Province, Province, *Company, int, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	com, sc := g.SelectedCompany(), g.SelectedShippingCompany()
	a := g.SelectedArea()
	a2 := g.SelectedArea2()
	goodsArea := g.SelectedGoodsArea()

	err := g.validatePlayerAction(c)
	switch {
	case err != nil:
		return nil, nil, 0, 0, nil, 0, err
	case g.Phase != Operations:
		return nil, nil, 0, 0, nil, 0, sn.NewVError("Expected %q phase, have %q phase.", PhaseNames[Operations], g.PhaseName())
	case g.SubPhase != OPSelectCityOrShip:
		return nil, nil, 0, 0, nil, 0, sn.NewVError("Expected %q subphase, have %q subphase.", SubPhaseNames[OPSelectCityOrShip], g.SubPhaseName())
	case c == nil:
		return nil, nil, 0, 0, nil, 0, sn.NewVError("You must select company to operate.")
	case goodsArea == nil:
		return nil, nil, 0, 0, nil, 0, sn.NewVError("Missing selected goods area.")
	case a == nil || a2 == nil || !a2.adjacentToArea(a):
		return nil, nil, 0, 0, nil, 0, sn.NewVError("You must select a ship adjacent to the previously selected area.")
	case a2.City == nil:
		return nil, nil, 0, 0, nil, 0, sn.NewVError("You must select an area with a city.")
	case goodsArea.Province() == NoProvince:
		return nil, nil, 0, 0, nil, 0, sn.NewVError("Invalid 'From' province. Undo turn and try again.")
	case a2.City.Delivered[com.Goods()] >= a2.City.Size:
		return nil, nil, 0, 0, nil, 0, sn.NewVError("City has already received its allotment of %s.", com.Goods())
	case g.ShipsUsed == InvalidUsedShips:
		return nil, nil, 0, 0, nil, 0, sn.NewVError("Missing temp value for used ships.")
	case sc == nil:
		return nil, nil, 0, 0, nil, 0, sn.NewVError("Missing temp value for shipping company owner.")
	default:
		// func (g *Game) validateSelectCity(c *gin.Context) (city *City, c *Company, from Province, to Province, sc *Company, used int, err error) {
		return a2.City, com, goodsArea.Province(), a2.Province(), sc, g.ShipsUsed, nil
	}
}

type deliveredGoodEntry struct {
	*Entry
	Goods     Goods
	From      Province
	To        Province
	ShipsUsed int
}

func (g *Game) newDeliveredGoodEntryFor(p *Player, goods Goods, from, to Province, ownerID, used int) (e *deliveredGoodEntry) {
	e = &deliveredGoodEntry{
		Entry:     g.newEntryFor(p),
		Goods:     goods,
		From:      from,
		To:        to,
		ShipsUsed: used,
	}
	e.OtherPlayerID = ownerID
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (e *deliveredGoodEntry) HTML(c *gin.Context) template.HTML {
	g := gameFrom(c)
	return restful.HTML("<div>%s delivered %s from the %s province to the city in the %s province using %d ships of %s.</div>", g.NameByPID(e.PlayerID), e.Goods, e.From, e.To, e.ShipsUsed, g.NameByPID(e.OtherPlayerID))
}

func (g *Game) receiveIncome(c *gin.Context) (string, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	g.SubPhase = OPReceiveIncome
	com, incomeMap, err := g.validateReceiveIncome(c)
	if err != nil {
		return "indonesia/flash_notice", err
	}
	cp := g.CurrentPlayer()
	otherShips := incomeMap.OtherShips(cp.ID())
	income := com.Delivered()*com.Goods().Price() - (otherShips * 5)
	cp.Rupiah += income
	cp.OpIncome += income
	if otherShips != 0 {
		for pid, count := range incomeMap {
			if pid != cp.ID() {
				income := 5 * count
				p := g.PlayerByID(pid)
				p.Rupiah += income
				p.OpIncome += income
			}
		}
	}

	// Log
	e := g.newReceiveIncomeEntryFor(g.CurrentPlayer(), com.Delivered(), com.Goods(), incomeMap)
	restful.AddNoticef(c, string(e.HTML(c)))
	return g.startCompanyExpansion(c), nil
}

// func (g *Game) validateReceiveIncome(c *gin.Context) (c *Company, incomeMap ShipperIncomeMap, err error) {
func (g *Game) validateReceiveIncome(c *gin.Context) (*Company, ShipperIncomeMap, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	com := g.SelectedCompany()
	incomeMap := g.ShipperIncomeMap
	err := g.validatePlayerAction(c)
	switch {
	case err != nil:
		return nil, nil, err
	case com == nil:
		return nil, nil, sn.NewVError("Missing selected company.")
	case g.ShipperIncomeMap == nil:
		return nil, nil, sn.NewVError("Missing income map.")
	default:
		return com, incomeMap, nil
	}
}

type receiveIncomeEntry struct {
	*Entry
	Delivered     int
	Goods         Goods
	ShipperIncome ShipperIncomeMap
}

func (g *Game) newReceiveIncomeEntryFor(p *Player, delivered int, goods Goods, incomeMap ShipperIncomeMap) (e *receiveIncomeEntry) {
	e = &receiveIncomeEntry{
		Entry:         g.newEntryFor(p),
		Delivered:     delivered,
		Goods:         goods,
		ShipperIncome: incomeMap,
	}
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (e *receiveIncomeEntry) HTML(c *gin.Context) (s template.HTML) {
	otherShips := e.ShipperIncome.OtherShips(e.PlayerID)
	rupiah := e.Delivered*e.Goods.Price() - (otherShips * 5)
	g := gameFrom(c)
	s = restful.HTML("<div>%s received %d rupiah for selling %d %s (%d &times; %d %s - 5 &times; %d ships)</div>",
		g.NameByPID(e.PlayerID), rupiah, e.Delivered, e.Goods, e.Goods.Price(), e.Delivered, e.Goods, otherShips)
	if otherShips != 0 {
		for pid, count := range e.ShipperIncome {
			if pid != e.PlayerID {
				s += restful.HTML("<div>%s received %d rupiah for %d ships used to transport %s.</div>",
					g.NameByPID(pid), 5*count, count, e.Goods)
			}
		}
	}
	return
}

func (g *Game) startCompanyExpansion(c *gin.Context) (tmpl string) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if company := g.SelectedCompany(); company.deliveredAllGoods() {
		g.SubPhase = OPFreeExpansion
		cp := g.CurrentPlayer()
		//g.RequiredExpansions = min(cp.Technologies[ExpansionsTech], len(company.ExpansionAreas()))
		g.RequiredExpansions = company.requiredExpansions()
		if g.RequiredExpansions == 0 {
			company.Operated = true
			cp.PerformedAction = true
			tmpl = "indonesia/completed_expansion_update"
		}
	} else {
		g.SubPhase = OPExpansion
	}
	tmpl = "indonesia/select_city_update"
	return
}

// func (g *Game) stopExpanding(c *gin.Context) (tmpl string, act game.ActionType, err error) {
func (g *Game) stopExpanding(c *gin.Context) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	com, err := g.validateStopExpanding(c)
	if err != nil {
		return "indonesia/flash_notice", game.None, err
	}

	cp := g.CurrentPlayer()
	cp.PerformedAction = true
	com.Operated = true

	// Log
	e := g.newStopExpandingEntryFor(g.CurrentPlayer())
	restful.AddNoticef(c, string(e.HTML(c)))
	return "indonesia/stop_expanding_update", game.Cache, nil
}

// func (g *Game) validateStopExpanding(c *gin.Context) (c *Company, err error) {
func (g *Game) validateStopExpanding(c *gin.Context) (*Company, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	com, err := g.SelectedCompany(), g.validatePlayerAction(c)
	switch {
	case err != nil:
		return nil, err
	case com == nil:
		return nil, sn.NewVError("Missing selected company.")
	case com.IsProductionCompany() && g.SubPhase == OPFreeExpansion:
		return nil, sn.NewVError("You can not stop expanding.")
	default:
		return com, nil
	}
}

type stopExpandingEntry struct {
	*Entry
}

func (g *Game) newStopExpandingEntryFor(p *Player) (e *stopExpandingEntry) {
	e = &stopExpandingEntry{Entry: g.newEntryFor(p)}
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (e *stopExpandingEntry) HTML(c *gin.Context) template.HTML {
	g := gameFrom(c)
	return restful.HTML("<div>%s stopped expanding selected company.</div>", g.NameByPID(e.PlayerID))
}

// func (g *Game) expandProduction(c *gin.Context) (tmpl string, err error) {
func (g *Game) expandProduction(c *gin.Context) (string, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	a, com, err := g.validateExpandProduction(c)
	if err != nil {
		return "indonesia/flash_notice", err
	}

	g.Expansions += 1
	a.AddProducer(com)
	com.AddArea(a)
	cp := g.CurrentPlayer()

	// Log
	if g.SubPhase == OPExpansion {
		expense := com.Goods().Price()
		cp.Rupiah -= expense
		cp.OpIncome -= expense
	}
	e := g.newExpandProductionEntryFor(cp, com.Goods(), a.Province(), g.SubPhase == OPFreeExpansion)
	restful.AddNoticef(c, string(e.HTML(c)))
	if g.Expansions == g.RequiredExpansions || cp.RemainingExpansions() == 0 {
		com.Operated = true
		cp.PerformedAction = true
		return "indonesia/completed_expansion_update", nil
	}
	return "indonesia/company_expansion_dialog", nil
}

func (p *Player) RemainingExpansions() int {
	if p.Game().SubPhase == OPFreeExpansion {
		return p.Game().RequiredExpansions - p.Game().Expansions
	}
	return p.Technologies[ExpansionsTech] - p.Game().Expansions
}

// func (g *Game) validateExpandProduction(c *gin.Context) (a *Area, c *Company, err error) {
func (g *Game) validateExpandProduction(c *gin.Context) (*Area, *Company, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	cp := g.CurrentPlayer()
	a, com, err := g.SelectedArea(), g.SelectedCompany(), g.validatePlayerAction(c)
	switch {
	case err != nil:
		return nil, nil, err
	case com == nil:
		return nil, nil, sn.NewVError("Missing selected company.")
	case a == nil:
		return nil, nil, sn.NewVError("Missing selected area.")
	case g.SubPhase == OPExpansion && cp.Rupiah < com.Goods().Price():
		return nil, nil, sn.NewVError("You do not have %d rupiah to pay for expansion.", com.Goods().Price())
	case !com.ExpansionAreas().include(a):
		return nil, nil, sn.NewVError("Selected area is not a valid expansion area.")
	case cp.RemainingExpansions() == 0:
		return nil, nil, sn.NewVError("You have already performed the allotted number of expansions.")
	default:
		return a, com, nil
	}
}

type expandProductionEntry struct {
	*Entry
	Goods    Goods
	Province Province
	Paid     int
}

func (g *Game) newExpandProductionEntryFor(p *Player, goods Goods, province Province, free bool) (e *expandProductionEntry) {
	e = &expandProductionEntry{
		Entry:    g.newEntryFor(p),
		Goods:    goods,
		Province: province,
	}
	if !free {
		e.Paid = goods.Price()
	}
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (e *expandProductionEntry) HTML(c *gin.Context) (s template.HTML) {
	g := gameFrom(c)
	n := g.NameByPID(e.PlayerID)
	if e.Paid == 0 {
		s = restful.HTML("<div>%s freely expanded the selected %s company to an area in the %s province.</div>", n, e.Goods, e.Province)
	} else {
		s = restful.HTML("<div>%s paid %d to expand the selected %s company to an area in the %s province.</div>", n, e.Paid, e.Goods, e.Province)
	}
	return
}

// func (g *Game) expandShipping(c *gin.Context) (tmpl string, err error) {
func (g *Game) expandShipping(c *gin.Context) (string, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	a, com, err := g.validateExpandShipping(c)
	if err != nil {
		return "indonesia/flash_notice", err
	}

	g.Expansions += 1
	com.Operated = true
	com.AddShipIn(a)
	cp := g.CurrentPlayer()

	// Log
	e := g.newExpandShippingEntryFor(g.CurrentPlayer(), com, a)
	restful.AddNoticef(c, string(e.HTML(c)))
	if g.Expansions < cp.Technologies[ExpansionsTech] && com.Ships() < com.MaxShips() {
		return "indonesia/select_shipping_area_update", nil
	}
	cp.PerformedAction = true
	return "indonesia/completed_expansion_update", nil
}

// func (g *Game) validateExpandShipping(c *gin.Context) (a *Area, c *Company, err error) {
func (g *Game) validateExpandShipping(c *gin.Context) (*Area, *Company, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	a := g.SelectedArea()
	com := g.SelectedCompany()
	maxShips := com.MaxShips()
	cp := g.CurrentPlayer()

	err := g.validatePlayerAction(c)
	switch {
	case err != nil:
		return nil, nil, err
	case com == nil:
		return nil, nil, sn.NewVError("Missing selected company.")
	case a == nil:
		return nil, nil, sn.NewVError("Missing selected area.")
	case !g.freeShippingExpansionAreas().include(a):
		return nil, nil, sn.NewVError("Selected area is not a valid expansion area.")
	case g.Expansions >= cp.Technologies[ExpansionsTech]:
		return nil, nil, sn.NewVError("You have already performed the allotted number of expansion.")
	case com.Ships() >= maxShips:
		return nil, nil, sn.NewVError("The selected company is already at it's ship limit of %d for the era.", maxShips)
	case com.MaxShips() == com.Ships():
		return nil, nil, sn.NewVError("The selected shipping company has already expanded to its ship limit for the era.")
	default:
		return a, com, nil
	}
}

type expandShippingEntry struct {
	*Entry
	Company Company
	Area    Area
}

func (g *Game) newExpandShippingEntryFor(p *Player, company *Company, area *Area) (e *expandShippingEntry) {
	e = &expandShippingEntry{
		Entry:   g.newEntryFor(p),
		Company: *company,
		Area:    *area,
	}
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (e *expandShippingEntry) HTML(c *gin.Context) template.HTML {
	g := gameFrom(c)
	return restful.HTML("<div>%s freely expanded the %s company to a sea area near the %s province.</div>",
		g.NameByPID(e.PlayerID), e.Company.String(), e.Area.Province().String())
}

func (g *Game) resetShipping() {
	for _, a := range g.seaAreas() {
		a.Used = false
		for _, s := range a.Shippers {
			s.Delivered = 0
		}
	}
}

func (p *Player) CanFreeExpansion() bool {
	g := p.Game()
	return g.Phase == Operations && g.SubPhase == OPFreeExpansion
}

// func (g *Game) acceptProposedFlow(c *gin.Context) (tmpl string, act game.ActionType, err error) {
func (g *Game) acceptProposedFlow(c *gin.Context) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	com, err := g.validateAcceptProposedFlow(c)
	if err != nil {
		return "indonesia/flash_notice", game.None, err
	}
	com.Operated = true

	g.ShipperIncomeMap = g.ProposedShips(g.ProposedPath)
	for aid, v := range g.ProposedCities() {
		g.GetArea(aid).City.Delivered[com.Goods()] += v
	}
	for fid, v := range g.ProposedPath[sourceFID] {
		count := 0
		for _, a := range com.ZoneFor(g.GetArea(fid.AreaID)).Areas() {
			a.Used = true
			if count += 1; count == v {
				break
			}
		}
	}

	tmpl, err := g.receiveIncome(c)
	if err != nil {
		return tmpl, game.None, err
	}
	return tmpl, game.Cache, nil
}

// func (g *Game) validateAcceptProposedFlow(c *gin.Context) (c *Company, err error) {
func (g *Game) validateAcceptProposedFlow(c *gin.Context) (*Company, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	com, err := g.SelectedCompany(), g.validatePlayerAction(c)
	switch {
	case err != nil:
		return nil, err
	case com == nil:
		return nil, sn.NewVError("Missing selected company.")
	case com.Delivered() != 0:
		return nil, sn.NewVError("You can not accept proposed deliveries.")
	default:
		return com, nil
	}
}
