package indonesia

import (
	"encoding/gob"
	"html/template"
	"strconv"

	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.Register(new(announceMergerEntry))
	gob.Register(new(mergerBidEntry))
	gob.Register(new(mergerResolutionEntry))
	gob.Register(new(removeRiceSpiceEntry))
}

func (g *Game) startMergers(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	g.Phase = Mergers
	g.setCurrentPlayers(g.Players()[0])
	g.beginningOfPhaseReset()

	// Reset Merger Data
	for _, company := range g.Companies() {
		company.Merged = false
	}

	g.Merger, g.SiapFajiMerger = nil, nil

	g.SubPhase = MSelectCompany1
	cp := g.CurrentPlayer()
	if !cp.CanAnnounceMerger() {
		g.autoPass(cp)
		if np := g.mergersNextPlayer(); np == nil {
			g.startAcquisitions(c)
		} else {
			g.setCurrentPlayers(np)
			if g.SubPhase == MSiapFajiCreation {
				g.beginningOfPhaseReset()
				g.SubPhase = MSelectCompany1
			}
		}
	}
}

func (g *Game) selectCompany1(c *gin.Context) (string, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	com, err := g.validateSelectCompany1(c)
	if err != nil {
		return "indonesia/flash_notice", err
	}

	cp := g.CurrentPlayer()
	g.beginningOfPhaseReset()
	g.Merger = newMerger(g)
	g.Merger.setCompany1(com)
	g.Merger.AnnouncerID = cp.ID()

	// Log
	e := g.newAnnounceMergerEntryFor(cp, com)
	restful.AddNoticef(c, string(e.HTML(c)))

	// Next SubPhase
	g.SubPhase = MSelectCompany2
	return "indonesia/announce_merger1_update", nil
}

func (g *Game) validateSelectCompany1(c *gin.Context) (*Company, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	com, err := g.SelectedCompany(), g.validatePlayerAction(c)
	switch {
	case err != nil:
		return nil, err
	case com == nil:
		return nil, sn.NewVError("Missing company selection.")
	case g.Phase != Mergers:
		return nil, sn.NewVError("Expected %q phase but has %q phase.", PhaseNames[Mergers], PhaseNames[g.Phase])
	case g.SubPhase != MSelectCompany1:
		return nil, sn.NewVError("Expected %q subphase but has %q subphase.", SubPhaseNames[MSelectCompany1], SubPhaseNames[g.SubPhase])
	default:
		return com, nil
	}
}

type announceMergerEntry struct {
	*Entry
	Company1 Company
	Company2 Company
}

func (g *Game) newAnnounceMergerEntryFor(p *Player, c1 *Company) (e *announceMergerEntry) {
	e = &announceMergerEntry{
		Entry:    g.newEntryFor(p),
		Company1: *c1,
	}
	e.Company2.OwnerID = NoPlayerID
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (p *Player) updateAnnounceMergerEntry(c2 *Company) *announceMergerEntry {
	g := p.Game()
	pIndex := len(p.Log) - 1
	gIndex := len(g.Log) - 1
	e := p.Log[pIndex].(*announceMergerEntry)
	e.Company2 = *c2
	p.Log[pIndex] = e
	g.Log[gIndex] = e
	return e
}
func (e *announceMergerEntry) HTML(c *gin.Context) (s template.HTML) {
	g := gameFrom(c)
	n := g.NameByPID(e.PlayerID)
	s = restful.HTML("<div>%s announces a merger of:</div>", n)
	if owner1 := g.PlayerByID(e.Company1.OwnerID); owner1 != nil {
		s += restful.HTML("<div>%s's %s %s company having %d deeds",
			g.NameFor(owner1), e.Company1.Province(), e.Company1.Goods(), len(e.Company1.Deeds))
	}
	if owner2 := g.PlayerByID(e.Company2.OwnerID); owner2 != nil {
		s += restful.HTML("; and </div>")
		s += restful.HTML("<div>%s's %s %s company having %d deeds.</div>",
			g.NameFor(owner2), e.Company2.Province(), e.Company2.Goods(), len(e.Company2.Deeds))
	} else {
		s += restful.HTML(".</div>")
	}
	return
}

func (g *Game) selectCompany2(c *gin.Context) (string, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	com, err := g.validateSelectCompany2(c)
	if err != nil {
		return "indonesia/flash_notice", err
	}

	e := g.CurrentPlayer().updateAnnounceMergerEntry(com)
	restful.AddNoticef(c, string(e.HTML(c)))
	g.SubPhase = MBid
	g.Merger.setCompany2(com)
	return "indonesia/announce_merger2_update", nil
}

func (g *Game) validateSelectCompany2(c *gin.Context) (*Company, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	com, err := g.SelectedCompany(), g.validatePlayerAction(c)
	switch {
	case err != nil:
		return nil, err
	case com == nil:
		return nil, sn.NewVError("Missing company selection.")
	case g.Phase != Mergers:
		return nil, sn.NewVError("Expected %q phase but has %q phase.", PhaseNames[Mergers], PhaseNames[g.Phase])
	case g.SubPhase != MSelectCompany2:
		return nil, sn.NewVError("Expected %q subphase but has %q subphase.", SubPhaseNames[MSelectCompany2], SubPhaseNames[g.SubPhase])
	default:
		return com, nil
	}
}

func (g *Game) mergerBid(c *gin.Context) (tmpl string, act game.ActionType, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	var bid int
	if bid, err = g.validateMergerBid(c); err != nil {
		tmpl, act = "indonesia/flash_notice", game.None
		return
	}

	cp := g.CurrentPlayer()
	if bid != NoBid {
		g.Merger.setBid(cp, bid)
	} else {
		cp.Passed = true
	}
	cp.PerformedAction = true

	e := g.newMergerBidEntryFor(cp, bid)
	restful.AddNoticef(c, string(e.HTML(c)))
	tmpl, act = "indonesia/merger_bid_update", game.Cache
	return
}

func (p *Player) CanBidNone() bool {
	return !(p.Game().Merger.AnnouncerID == p.ID() && p.Game().Merger.CurrentBid == 0)
}

func (g *Game) validateMergerBid(c *gin.Context) (bid int, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	bid = NoBid
	if err = g.validatePlayerAction(c); err != nil {
		return
	}

	cp := g.CurrentPlayer()
	if bidValue := c.PostForm("bid"); bidValue == "none" {
		if g.Merger.AnnouncerID == cp.ID() && g.Merger.CurrentBid == 0 {
			err = sn.NewVError("You must bid at least the nominal value of Rp %d in order to announce the merger.", g.Merger.NominalBid())
		}
	} else if bid, err = strconv.Atoi(bidValue); err == nil {
		switch {
		case bid > cp.Rupiah:
			err = sn.NewVError("You bid more than you have.")
		case bid < g.Merger.CurrentBid:
			err = sn.NewVError("You can't bid less than current bid.")
		case !g.Merger.BidsFor(cp).include(bid):
			err = sn.NewVError("Bid must be equal to nominal value + multiple of goods/ships.")
		}
	}
	return
}

type mergerBidEntry struct {
	*Entry
	Bid int
}

func (g *Game) newMergerBidEntryFor(p *Player, bid int) (e *mergerBidEntry) {
	e = &mergerBidEntry{
		Entry: g.newEntryFor(p),
		Bid:   bid,
	}
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (e *mergerBidEntry) HTML(c *gin.Context) (s template.HTML) {
	g := gameFrom(c)
	n := g.NameByPID(e.PlayerID)
	if e.Bid == NoBid {
		s = restful.HTML("<div>%s did not bid on the announced merger.</div>", n)
	} else {
		s = restful.HTML("<div>%s bid %d on the announced merger.</div>", n, e.Bid)
	}
	return
}

func (m *Merger) CurrentBidder() *Player {
	return m.g.PlayerByID(m.CurrentBidderID)
}

type Merger struct {
	g               *Game
	Owner1ID        int
	Owner1Slot      int
	Owner2ID        int
	Owner2Slot      int
	CurrentBid      int
	CurrentBidderID int
	AnnouncerID     int
}

type SiapFajiMerger struct {
	g          *Game
	OwnerID    int
	OwnerSlot  int
	Production int
}

func (g *Game) newSiapFajiMerger(c *Company) {
	g.SiapFajiMerger = &SiapFajiMerger{
		g:          g,
		OwnerID:    c.OwnerID,
		OwnerSlot:  c.Slot,
		Production: c.Production() / 2,
	}
}

func (s *SiapFajiMerger) owner() *Player {
	return s.g.PlayerByID(s.OwnerID)
}

func (s *SiapFajiMerger) Company() *Company {
	if owner := s.owner(); owner == nil {
		return nil
	} else {
		return owner.Slots[s.OwnerSlot-1].Company
	}
}

func (s *SiapFajiMerger) GoodsToRemove() int {
	c := s.Company()
	if c != nil {
		return c.Production() - s.Production
	}
	return 0
}

func (s *SiapFajiMerger) init(g *Game) {
	s.g = g
}

func newMerger(g *Game) *Merger {
	return &Merger{
		g:               g,
		Owner1ID:        NoPlayerID,
		Owner2ID:        NoPlayerID,
		CurrentBidderID: NoPlayerID,
	}
}

func (m *Merger) IsShippingCompany() bool {
	return m.Company1().IsShippingCompany() && m.Company2().IsShippingCompany()
}

func (m *Merger) IsProductionCompany() bool {
	return m.Company1().IsProductionCompany() && m.Company2().IsProductionCompany()
}

func (m *Merger) NominalBid() int {
	if inc := m.BidIncrement(); inc > 0 {
		return inc * m.Price()
	}
	return -1
}

func (m *Merger) Ships() int {
	return m.Company1().Ships() + m.Company2().Ships()
}

func (m *Merger) Production() int {
	return m.Company1().Production() + m.Company2().Production()
}

func (m *Merger) BidIncrement() int {
	if m.IsShippingCompany() {
		return m.Ships()
	}
	if m.IsProductionCompany() {
		return m.Production()
	}
	return 0
}

func (m *Merger) Price() int {
	if m.IsShippingCompany() {
		return m.Company1().Goods().Price()
	}
	if m.IsProductionCompany() {
		g1, g2 := m.Company1().Goods(), m.Company2().Goods()
		if (g1 == Rice && g2 == Spice) || (g1 == Spice && g2 == Rice) {
			return 25
		}
		if g1 == g2 {
			return g1.Price()
		}
	}
	return 0
}

type bids []int

func (bs bids) include(bid int) bool {
	for _, b := range bs {
		if b == bid {
			return true
		}
	}
	return false
}

func (m *Merger) BidsFor(p *Player) bids {
	inc := m.BidIncrement()
	min := m.CurrentBid + inc
	if m.CurrentBid == 0 {
		min = m.NominalBid()
	}
	bids := make(bids, 0)
	for i := min; i <= p.Rupiah; i += inc {
		bids = append(bids, i)
	}
	return bids
}

func (m *Merger) Owner1() *Player {
	return m.g.PlayerByID(m.Owner1ID)
}

func (m *Merger) Company1() *Company {
	if owner1, slot := m.Owner1(), m.Owner1Slot; owner1 == nil || slot < 1 || slot > 5 {
		return nil
	} else {
		return owner1.Slots[slot-1].Company
	}
}

func (m *Merger) Owner2() *Player {
	return m.g.PlayerByID(m.Owner2ID)
}

func (m *Merger) Company2() *Company {
	if owner2, slot := m.Owner2(), m.Owner2Slot; owner2 == nil || slot < 1 || slot > 5 {
		return nil
	} else {
		return owner2.Slots[slot-1].Company
	}
}

func (m *Merger) Owner1Share() int {
	switch c1 := m.Company1(); {
	case c1.IsProductionCompany():
		return c1.Production() * m.CurrentBid / m.BidIncrement()
	case c1.IsShippingCompany():
		return c1.Ships() * m.CurrentBid / m.BidIncrement()
	default:
		return 0
	}
}

func (m *Merger) Owner2Share() int {
	switch c2 := m.Company2(); {
	case c2.IsProductionCompany():
		return c2.Production() * m.CurrentBid / m.BidIncrement()
	case c2.IsShippingCompany():
		return c2.Ships() * m.CurrentBid / m.BidIncrement()
	default:
		return 0
	}
}

func (m *Merger) setCompany1(c *Company) {
	m.Owner1ID, m.Owner1Slot = c.OwnerID, c.Slot
}

func (m *Merger) setCompany2(c *Company) {
	m.Owner2ID, m.Owner2Slot = c.OwnerID, c.Slot
}

func (m *Merger) setBid(p *Player, bid int) {
	m.CurrentBidderID, m.CurrentBid = p.ID(), bid
}

func (p *Player) canSelectFirstCompany(c *Company) bool {
	g := p.Game()
	return g.Merger == nil && len(mergeableCompaniesFor(p)[c]) > 0
}

func (p *Player) canSelectSecondCompany(c *Company) bool {
	g := p.Game()
	m := g.Merger
	if m == nil {
		return false
	}
	c1 := m.Company1()
	if c1 == nil {
		return false
	}
	c2 := m.Company2()
	if c2 != nil {
		return false
	}
	mergeableCompanies := mergeableCompaniesFor(p)
	companies := mergeableCompanies[c1]
	if companies.include(c) {
		return true
	}
	return false
}

func (p *Player) CanBidOnMerger() bool {
	if p == nil {
		return false
	}
	g := p.Game()
	m, cs := g.Merger, p.Companies()
	if m == nil {
		return false
	}
	c1, c2 := m.Company1(), m.Company2()
	return g.Phase == Mergers && g.SubPhase == MBid &&
		(p.hasEmptySlot() || cs.include(c1) || cs.include(c2)) &&
		!p.Passed && m.CurrentBidder() != nil && !p.Equal(m.CurrentBidder()) &&
		p.Rupiah >= m.CurrentBid+m.BidIncrement()
}

func (p *Player) CanAnnounceMerger() bool {
	return p != nil && p.Game().Phase == Mergers && p.Game().SubPhase == MSelectCompany1 &&
		!p.PerformedAction && p.canMergeCompanies()
}

func (p *Player) CanAnnounceSecondCompany() bool {
	return p != nil && p.Game().Phase == Mergers && p.Game().SubPhase == MSelectCompany2 &&
		!p.PerformedAction && p.canMergeCompanies()
}

func (c *Company) compatableWith(company *Company) bool {
	switch c.g.Era {
	case EraA:
		return c.Goods() == company.Goods()
	default:
		return (c.Goods() == company.Goods()) ||
			(c.Goods() == Rice && company.Goods() == Spice) ||
			(c.Goods() == Spice && company.Goods() == Rice)
	}
}

func (p *Player) canMergeCompanies() bool {
	return len(mergeableCompaniesFor(p)) > 0
}

type CompanyMap map[*Company]Companies

func mergeableCompaniesFor(p *Player) CompanyMap {
	cmap := make(CompanyMap, 0)
	if p.Technologies[MergersTech] < 2 {
		return cmap
	}

	g := p.Game()
	for _, c1 := range g.Companies() {
		margin := p.Technologies[MergersTech] - len(c1.Deeds)
		if !c1.Merged && margin > 0 {
			for _, c2 := range g.Companies() {
				m := &Merger{
					g:          g,
					Owner1ID:   c1.OwnerID,
					Owner1Slot: c1.Slot,
					Owner2ID:   c2.OwnerID,
					Owner2Slot: c2.Slot,
				}
				if !c2.Merged && c1 != c2 &&
					c1.compatableWith(c2) && len(c2.Deeds) <= margin &&
					(p.hasEmptySlot() || p.owns(c1) || p.owns(c2)) &&
					p.Rupiah >= m.NominalBid() {
					cmap[c1] = append(cmap[c1], c2)
				}
			}
		}
	}
	return cmap
}

func (p *Player) owns(c *Company) bool {
	return p.ID() == c.OwnerID
}

func (g *Game) startMergerResolution(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	g.Phase = Mergers
	g.SubPhase = MResolution
	g.payOwnersOf(c, g.Merger)
	c1Goods := g.Merger.Company1().Goods()
	com := g.Merger.execute()

	// If merged company is newly created Siap Faji company handle differently.
	if c1Goods != SiapFaji && com.Goods() == SiapFaji {
		g.newSiapFajiMerger(com)
		g.siapFajiCreation(c)
		if !g.SiapFajiMerger.CanEndRiceSpiceRemoval() {
			return
		}
		g.SiapFajiMerger.Company().toSiapFaji()
	}

	// Continue with next player or start next phase
	announcer := g.PlayerByID(g.Merger.AnnouncerID)
	g.Merger = nil
	g.setCurrentPlayers(announcer)
	g.beginningOfPhaseReset()
	g.SubPhase = MSelectCompany1
	if np := g.mergersNextPlayer(); np != nil {
		g.setCurrentPlayers(np)
	} else {
		g.startAcquisitions(c)
	}
}

func (g *Game) payOwnersOf(c *gin.Context, m *Merger) {
	bidder, owner1, owner2 := m.CurrentBidder(), m.Owner1(), m.Owner2()
	bidder.Rupiah -= m.CurrentBid
	r1, r2 := m.Owner1Share(), m.Owner2Share()
	owner1.Rupiah += r1
	owner2.Rupiah += r2

	e := g.newMergerResolutionEntryFor(bidder, r1, r2)
	restful.AddNoticef(c, string(e.HTML(c)))
}

func (m *Merger) execute() *Company {
	owner, owner1, owner2 := m.CurrentBidder(), m.Owner1(), m.Owner2()
	c1, slot1, c2, slot2 := m.Company1(), m.Company1().Slot, m.Company2(), m.Company2().Slot

	// Remove companies from current owners to ensure bidder has empty slot
	owner1.Slots[slot1-1].Company = nil
	owner2.Slots[slot2-1].Company = nil

	// Get empty slot
	slot, index := owner.getEmptySlot()
	slot.Company = &Company{
		g:        m.g,
		Deeds:    append(c1.Deeds, c2.Deeds...),
		Merged:   true,
		OwnerID:  owner.ID(),
		Slot:     index,
		ShipType: c1.ShipType,
	}

	// Merge Zones/Areas
	if slot.Company.IsProductionCompany() {
		slot.Company.Zones = append(c1.Zones, c2.Zones...)
	} else {
		slot.Company.Zones = c1.Zones.addZones(c2.Zones...)
	}

	// Update Areas
	for _, area := range slot.Company.Areas() {
		if slot.Company.IsProductionCompany() {
			area.Producer.OwnerID = slot.Company.OwnerID
			area.Producer.Slot = slot.Company.Slot
		} else {
			for _, shipper := range area.Shippers {
				if (shipper.OwnerID == owner1.ID() && shipper.Slot == slot1) ||
					(shipper.OwnerID == owner2.ID() && shipper.Slot == slot2) {
					shipper.OwnerID = owner.ID()
					shipper.Slot = index
					shipper.ShipType = c1.ShipType
				}
			}
		}
	}
	return slot.Company
}

type mergerResolutionEntry struct {
	*Entry
	Bid      int
	Owner1ID int
	Rupiah1  int
	Owner2ID int
	Rupiah2  int
}

func (g *Game) newMergerResolutionEntryFor(p *Player, r1, r2 int) (e *mergerResolutionEntry) {
	m := g.Merger
	e = &mergerResolutionEntry{
		Entry:    g.newEntryFor(p),
		Bid:      m.CurrentBid,
		Owner1ID: m.Owner1ID,
		Rupiah1:  r1,
		Owner2ID: m.Owner2ID,
		Rupiah2:  r2,
	}
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (e *mergerResolutionEntry) HTML(c *gin.Context) (s template.HTML) {
	g := gameFrom(c)
	n := g.NameByPID(e.PlayerID)
	owner1, owner2 := g.PlayerByID(e.Owner1ID), g.PlayerByID(e.Owner2ID)
	s = restful.HTML("<div>%s bought the merged company for %d Rupiah.</div>", n, e.Bid)
	s += restful.HTML("<div>%s received %d Rupiah.</div>", g.NameFor(owner1), e.Rupiah1)
	s += restful.HTML("<div>%s received %d Rupiah.</div>", g.NameFor(owner2), e.Rupiah2)
	return
}

func (g *Game) siapFajiCreation(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	g.SubPhase = MSiapFajiCreation
	owner := g.SiapFajiMerger.owner()
	g.setCurrentPlayers(owner)
	owner.PerformedAction = false
	g.SiapFajiMerger.removeAreasAdjacentCompetitor()
}

func (m *SiapFajiMerger) CanEndRiceSpiceRemoval() bool {
	return m.GoodsToRemove() <= 0 && m.Company().Zones.contiguous()
}

func (p *Player) CanCreateSiapFaji() bool {
	return p != nil && p.Game().Phase == Mergers && p.Game().SubPhase == MSiapFajiCreation &&
		p.Game().SiapFajiMerger != nil && !p.PerformedAction && p.IsCurrentPlayer() &&
		p.ID() == p.Game().SiapFajiMerger.OwnerID
}

func (m *SiapFajiMerger) removeAreasAdjacentCompetitor() {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	comp := m.Company()
	if comp == nil {
		return
	}
	for _, a := range comp.Areas() {
		if a.adjacentAreaHasCompetingCompanyFor(comp) {
			comp.remove(a)
		}
	}
}

func (g *Game) removeRiceSpice(c *gin.Context) (string, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	m, a, err := g.validateRemoveRiceSpice(c)
	if err != nil {
		return "indonesia/flash_notice", err
	}

	cp := g.CurrentPlayer()
	com := m.Company()
	goods := a.Producer.Goods
	com.remove(a)

	// Log
	e := g.newRemoveRiceSpiceEntryFor(cp, a, goods)
	restful.AddNoticef(c, string(e.HTML(c)))

	// Return if more goods to remove
	if !m.CanEndRiceSpiceRemoval() {
		return "indonesia/remove_goods_update", nil
	}

	m.Company().toSiapFaji()

	// Merge zones
	if len(com.Zones) > 1 {
		com.Zones = Zones{com.Zones[0]}.addZones(com.Zones[1:]...)
	}

	// Reset game state for next merger round.
	cp.PerformedAction = true
	return "indonesia/remove_goods_update", nil
}

func (comp *Company) toSiapFaji() {
	for _, a := range comp.Areas() {
		a.Producer.Goods = SiapFaji
	}
}

func (g *Game) validateRemoveRiceSpice(c *gin.Context) (*SiapFajiMerger, *Area, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	m, a, err := g.SiapFajiMerger, g.SelectedArea(), g.validatePlayerAction(c)
	switch {
	case err != nil:
		return nil, nil, err
	case g.SiapFajiMerger == nil:
		return nil, nil, sn.NewVError("No Siap Faji Merger defined.")
	case g.SiapFajiMerger.Company() == nil:
		return nil, nil, sn.NewVError("No Siap Faji Merger company.")
	case g.SelectedArea() == nil:
		return nil, nil, sn.NewVError("No area selected.")
	case g.SiapFajiMerger.Company().Goods() != SiapFaji:
		return nil, nil, sn.NewVError("Wrong goods for Siap Faji Merger company.")
	case g.SelectedArea().Goods() != Rice && g.SelectedArea().Goods() != Spice:
		return nil, nil, sn.NewVError("Selected area does not have rice or spice.")
	case !g.SiapFajiMerger.Company().Areas().include(g.SelectedArea()):
		return nil, nil, sn.NewVError("Selected Area not part of Siap Faji Merger company.")
	case g.Phase != Mergers:
		return nil, nil, sn.NewVError("Expected %q phase but has %q phase.",
			PhaseNames[Mergers], PhaseNames[g.Phase])
	case g.SubPhase != MSiapFajiCreation:
		return nil, nil, sn.NewVError("Expected %q subphase but has %q subphase.",
			SubPhaseNames[MSiapFajiCreation], SubPhaseNames[g.SubPhase])
	default:
		return m, a, nil
	}
}

type removeRiceSpiceEntry struct {
	*Entry
	AreaID AreaID
	Goods  Goods
}

func (g *Game) newRemoveRiceSpiceEntryFor(p *Player, a *Area, goods Goods) (e *removeRiceSpiceEntry) {
	e = &removeRiceSpiceEntry{
		Entry:  g.newEntryFor(p),
		AreaID: a.ID,
		Goods:  goods,
	}
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (e *removeRiceSpiceEntry) HTML(c *gin.Context) template.HTML {
	g := gameFrom(c)
	return restful.HTML("<div>%s removed %s from %s</div>", g.NameByPID(e.PlayerID), e.Goods, g.Areas[e.AreaID].Province())
}
