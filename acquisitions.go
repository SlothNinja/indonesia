package indonesia

import (
	"encoding/gob"
	"html/template"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.Register(new(acquiredCompanyEntry))
}

func (g *Game) startAcquisitions(c *gin.Context, cu *user.User) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	g.Phase = Acquisitions
	g.beginningOfPhaseReset()
	if np := g.acquisitionsNextPlayer(cu, g.Players()[g.NumPlayers-1]); np == nil {
		g.startResearch(c)
	} else {
		g.setCurrentPlayers(np)
	}
}

func (g *Game) SelectedCompany() *Company {
	if p, slot := g.SelectedPlayer(), g.SelectedSlot; p == nil || slot == NoSlot || slot < 1 || slot > 5 {
		return nil
	} else {
		return p.Slots[slot-1].Company
	}
}

func (g *Game) SelectedShippingCompany() *Company {
	return g.ShippingCompanies()[g.SelectedShippingProvince]
}

func (g *Game) acquireCompany(c *gin.Context, cu *user.User) (tmpl string, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	var (
		s              *Slot
		d              *Deed
		sIndex, dIndex int
	)

	if s, sIndex, d, dIndex, err = g.validateAcquireCompany(c, cu); err != nil {
		tmpl = "indonesia/flash_notice"
		return
	}

	cp := g.CurrentPlayer()
	s.Company = newCompany(g, cp, sIndex, d)

	// Cache SelectedSlot, SelectedPlayerID so SelectedCompany works.
	g.setSelectedPlayer(cp)
	g.SelectedSlot = sIndex
	g.AvailableDeeds = g.AvailableDeeds.removeAt(dIndex)
	if s.Company.IsProductionCompany() {
		g.SubPhase = AQInitialProduction
	} else {
		s.Company.ShipType = g.getAvailableShipType()
		g.SubPhase = AQInitialShip
	}
	tmpl = "indonesia/acquire_company_update"
	return
}

func (g *Game) validateAcquireCompany(c *gin.Context, cu *user.User) (s *Slot, sIndex int, d *Deed, dIndex int, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if err = g.validatePlayerAction(cu); err != nil {
		return
	}

	cp := g.CurrentPlayer()
	s, sIndex = cp.getEmptySlot()
	d, dIndex = g.SelectedDeed(), g.SelectedDeedIndex

	switch {
	case d == nil:
		err = sn.NewVError("You must select deed.")
	case s == nil:
		err = sn.NewVError("You do not have a free slot for the company.")
	}
	return
}

func (g *Game) placeInitialProduct(c *gin.Context, cu *user.User) (string, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	a, com, err := g.validateplaceInitialProduct(c, cu)
	if err != nil {
		return "indonesia/flash_notice", err
	}

	cp := g.CurrentPlayer()
	cp.PerformedAction = true
	a.AddProducer(com)
	com.AddArea(a)

	// Log placement
	e := g.newAcquiredCompanyEntryFor(cp, com)
	restful.AddNoticef(c, string(e.HTML(c)))

	// Reset SubPhase
	g.SubPhase = NoSubPhase
	return "indonesia/placed_product_update", nil
}

func (g *Game) validateplaceInitialProduct(c *gin.Context, cu *user.User) (*Area, *Company, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	a, com, err := g.SelectedArea(), g.SelectedCompany(), g.validatePlayerAction(cu)
	switch {
	case err != nil:
		return nil, nil, err
	case c == nil:
		return nil, nil, sn.NewVError("You must acquire a company first.")
	case a == nil:
		return nil, nil, sn.NewVError("You must select an area for the %s token.", com.Goods())
	case !a.IsLand():
		return nil, nil, sn.NewVError("You must select a land area for the initial %s token.", com.Goods())
	case com.Deeds[0].Province != a.Province():
		return nil, nil, sn.NewVError("You must select a land area in the %s province for the initial %s token.", com.Deeds[0].Province, com.Goods())
	case a.City != nil:
		return nil, nil, sn.NewVError("You can not place a %s token in an area having a city.", com.Goods())
	case a.Producer != nil:
		return nil, nil, sn.NewVError("You can not place a %s token in an area already having goods token.", com.Goods())
	case a.adjacentAreaHasCompetingCompanyFor(com):
		return nil, nil, sn.NewVError("You can not place a %s token adjacent to an area having %s token.", com.Goods(), com.Goods())
	default:
		return a, com, err
	}
}

type acquiredCompanyEntry struct {
	*Entry
	Deed Deed
}

func (g *Game) newAcquiredCompanyEntryFor(p *Player, company *Company) (e *acquiredCompanyEntry) {
	e = &acquiredCompanyEntry{
		Entry: g.newEntryFor(p),
		Deed:  *(company.Deeds[0]),
	}
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (e *acquiredCompanyEntry) HTML(c *gin.Context) template.HTML {
	g := gameFrom(c)
	return restful.HTML("<div>%s started a %s company in the %s province.</div>",
		g.NameByPID(e.PlayerID), e.Deed.Goods, e.Deed.Province)
}

func (g *Game) placeInitialShip(c *gin.Context, cu *user.User) (string, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	a, com, err := g.validateplaceInitialShip(c, cu)
	if err != nil {
		return "indonesia/flash_notice", err
	}

	cp := g.CurrentPlayer()
	cp.PerformedAction = true
	com.AddShipIn(a)

	// Log placement
	e := g.newAcquiredCompanyEntryFor(cp, com)
	restful.AddNoticef(c, string(e.HTML(c)))

	// Reset SubPhase
	g.SubPhase = NoSubPhase
	return "indonesia/placed_product_update", nil
}

func (g *Game) validateplaceInitialShip(c *gin.Context, cu *user.User) (*Area, *Company, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	err := g.validatePlayerAction(cu)
	if err != nil {
		return nil, nil, err
	}

	a, com := g.SelectedArea(), g.SelectedCompany()
	switch {
	case com == nil:
		return nil, nil, sn.NewVError("You must acquire a company first.")
	case a == nil:
		return nil, nil, sn.NewVError("You must select an area for the %s token.", com.Goods)
	case !a.IsSea():
		return nil, nil, sn.NewVError("You must select a sea area.")
	case !a.adjacentToProvince(com.Deeds[0].Province):
		return nil, nil, sn.NewVError("You must select a sea are adjacent to the %s province.", com.Deeds[0].Province)
	default:
		return a, com, nil
	}
}
