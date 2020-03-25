package indonesia

import (
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	stats "github.com/SlothNinja/user-stats"
	"github.com/gin-gonic/gin"
)

func (svr server) finish(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		g := gameFrom(c)
		oldCP := g.CurrentPlayer()

		s, cs, err := g.finishTurn(c)
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param(hParam)))
			return
		}

		// Game is over if cs != nil
		if cs != nil {
			g.Phase = GameOver
			g.Status = game.Completed
			ks, es := wrap(s.GetUpdate(c, g.UpdatedAt), cs)
			err = svr.saveWith(c, g, ks, es)
			if err != nil {
				log.Errorf(err.Error())
				c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param(hParam)))
				return
			}
			err = g.SendEndGameNotifications(c)
			if err != nil {
				log.Warningf(err.Error())
			}
			c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param(hParam)))
			return
		}

		s = s.GetUpdate(c, g.UpdatedAt)
		err = svr.saveWith(c, g, []*datastore.Key{s.Key}, []interface{}{s})
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param(hParam)))
			return
		}

		newCP := g.CurrentPlayer()
		if newCP != nil && oldCP.ID() != newCP.ID() {
			err = g.SendTurnNotificationsTo(c, newCP)
			if err != nil {
				log.Warningf(err.Error())
			}
		}

		defer c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param(hParam)))
	}
}

func (g *Game) finishTurn(c *gin.Context) (s *stats.Stats, cs contest.Contests, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	switch {
	case g.Phase == NewEra:
		s, err = g.newEraFinishTurn(c)
	case g.Phase == BidForTurnOrder:
		s, err = g.bidForTurnOrderFinishTurn(c)
	case g.Phase == Mergers && g.SubPhase == MBid:
		s, err = g.mergersBidFinishTurn(c)
	case g.Phase == Mergers:
		s, err = g.mergersFinishTurn(c)
	case g.Phase == Acquisitions:
		s, err = g.acquisitionsFinishTurn(c)
	case g.Phase == Research:
		s, cs, err = g.researchFinishTurn(c)
	case g.Phase == Operations:
		s, cs, err = g.companyExpansionFinishTurn(c)
	case g.Phase == CityGrowth:
		s, cs, err = g.cityGrowthFinishTurn(c)
	default:
		err = sn.NewVError("Improper Phase for finishing turn.")
	}

	return
}

func (g *Game) validateFinishTurn(c *gin.Context) (s *stats.Stats, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	var cp *Player

	switch cp, s = g.CurrentPlayer(), stats.Fetched(c); {
	case s == nil:
		err = sn.NewVError("missing stats for player.")
	case !g.CUserIsCPlayerOrAdmin(c):
		err = sn.NewVError("Only the current player may finish a turn.")
	case !cp.PerformedAction:
		err = sn.NewVError("%s has yet to perform an action.", g.NameFor(cp))
	}
	return
}

// ps is an optional parameter.
// If no player is provided, assume current player.
func (g *Game) nextPlayer(ps ...game.Playerer) (p *Player) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if nper := g.NextPlayerer(ps...); nper != nil {
		p = nper.(*Player)
	}
	return
}

func (g *Game) newEraNextPlayer(pers ...game.Playerer) (p *Player) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	g.CurrentPlayer().endOfTurnUpdate()
	p = g.nextPlayer(pers...)
	for g.Players().anyCanPlaceCity() {
		if !p.CanPlaceCity() {
			p = g.nextPlayer(p)
		} else {
			p.beginningOfTurnReset()
			return
		}
	}
	return nil
}

func (g *Game) removeUnplayableCityCardsFor(c *gin.Context, p *Player) {
	var newCityCards CityCards
	for _, card := range p.CityCards {
		if card.Era != g.Era {
			newCityCards = append(newCityCards, card)
		} else {
			e := g.newDiscardCityEntryFor(p, card)
			restful.AddNoticef(c, string(e.HTML(c)))
		}
	}
	p.CityCards = newCityCards
}

func (g *Game) newEraFinishTurn(c *gin.Context) (s *stats.Stats, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if s, err = g.validateNewEraFinishTurn(c); err != nil {
		return
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	if np := g.newEraNextPlayer(); np == nil {
		for _, p := range g.Players() {
			g.removeUnplayableCityCardsFor(c, p)
		}
		g.startBidForTurnOrder(c)
	} else {
		g.setCurrentPlayers(np)
	}
	return
}

func (g *Game) validateNewEraFinishTurn(c *gin.Context) (s *stats.Stats, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if s, err = g.validateFinishTurn(c); g.Phase != NewEra {
		err = sn.NewVError(`Expected "New Era" phase but have %q phase.`, g.Phase)
	}
	return
}

func (g *Game) bidForTurnOrderNextPlayer(pers ...game.Playerer) *Player {
	g.CurrentPlayer().endOfTurnUpdate()
	p := g.nextPlayer(pers...)
	for !p.Equal(g.Players()[0]) {
		if !p.CanBid() {
			p = g.nextPlayer(p)
		} else {
			return p
		}
	}
	return nil
}

func (g *Game) bidForTurnOrderFinishTurn(c *gin.Context) (s *stats.Stats, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if s, err = g.validateBidForTurnOrderFinishTurn(c); err != nil {
		return
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	if np := g.bidForTurnOrderNextPlayer(); np == nil {
		g.setTurnOrder(c)
	} else {
		g.setCurrentPlayers(np)
	}
	return
}

func (g *Game) validateBidForTurnOrderFinishTurn(c *gin.Context) (s *stats.Stats, err error) {
	if s, err = g.validateFinishTurn(c); g.Phase != BidForTurnOrder {
		err = sn.NewVError(`Expected "Bid For Turn Order" phase but have %q phase.`, g.Phase)
	}
	return
}

func (g *Game) mergersBidNextPlayer(pers ...game.Playerer) *Player {
	g.CurrentPlayer().endOfTurnUpdate()
	p := g.nextPlayer(pers...)
	for !g.Players().allPassed() {
		if !p.CanBidOnMerger() {
			g.autoPass(p)
			p = g.nextPlayer(p)
		} else {
			return p
		}
	}
	return nil
}

func (g *Game) mergersBidFinishTurn(c *gin.Context) (s *stats.Stats, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if s, err = g.validateMergersBidFinishTurn(c); err != nil {
		return
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	if np := g.mergersBidNextPlayer(); np == nil {
		g.startMergerResolution(c)
	} else {
		g.setCurrentPlayers(np)
	}
	return
}

func (g *Game) validateMergersBidFinishTurn(c *gin.Context) (s *stats.Stats, err error) {
	if s, err = g.validateFinishTurn(c); g.Phase != Mergers {
		err = sn.NewVError(`Expected "Mergers" phase but have %q phase.`, g.Phase)
	}
	return
}

func (g *Game) mergersNextPlayer(pers ...game.Playerer) *Player {
	g.CurrentPlayer().endOfTurnUpdate()
	p := g.nextPlayer(pers...)
	for !g.Players().allPassed() {
		if !p.CanAnnounceMerger() {
			g.autoPass(p)
			p = g.nextPlayer(p)
		} else {
			return p
		}
	}
	return nil
}

func (g *Game) mergersFinishTurn(c *gin.Context) (s *stats.Stats, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if s, err = g.validateMergersFinishTurn(c); err != nil {
		return
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	if g.SubPhase == MSiapFajiCreation {
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
	} else {
		if np := g.mergersNextPlayer(); np == nil {
			g.startAcquisitions(c)
		} else {
			g.setCurrentPlayers(np)
		}
	}
	return
}

func (g *Game) validateMergersFinishTurn(c *gin.Context) (s *stats.Stats, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if s, err = g.validateFinishTurn(c); g.Phase != Mergers {
		err = sn.NewVError(`Expected "Mergers" phase but have %q phase.`, g.Phase)
	}
	return
}

func (g *Game) acquisitionsNextPlayer(pers ...game.Playerer) (p *Player) {
	g.CurrentPlayer().endOfTurnUpdate()
	p = g.nextPlayer(pers...)
	for !g.Players().allPassed() {
		if !p.CanAcquireCompany() {
			g.autoPass(p)
			p = g.nextPlayer(p)
		} else {
			return
		}
	}
	p = nil
	return
}

func (g *Game) acquisitionsFinishTurn(c *gin.Context) (s *stats.Stats, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if s, err = g.validateAcquisitionsFinishTurn(c); err != nil {
		return
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	if np := g.acquisitionsNextPlayer(); np == nil {
		g.startResearch(c)
	} else {
		g.setCurrentPlayers(np)
	}
	return
}

func (g *Game) validateAcquisitionsFinishTurn(c *gin.Context) (*stats.Stats, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	s, err := g.validateFinishTurn(c)
	if err != nil {
		return nil, err
	}
	if g.Phase != Acquisitions {
		return nil, sn.NewVError(`Expected "Acquisitions" phase but have %q phase.`, g.Phase)
	}
	return s, nil
}

func (g *Game) researchNextPlayer(pers ...game.Playerer) *Player {
	g.CurrentPlayer().endOfTurnUpdate()
	p := g.nextPlayer(pers...)
	for !p.Equal(g.Players()[0]) {
		if !p.CanResearch() {
			g.autoPass(p)
			p = g.nextPlayer(p)
		} else {
			return p
		}
	}
	return nil
}

func (g *Game) researchFinishTurn(c *gin.Context) (s *stats.Stats, cs contest.Contests, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if s, err = g.validateResearchFinishTurn(c); err != nil {
		return
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	if np := g.researchNextPlayer(); np == nil {
		cs = g.startOperations(c)
	} else {
		g.setCurrentPlayers(np)
	}
	return
}

func (g *Game) validateResearchFinishTurn(c *gin.Context) (s *stats.Stats, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if s, err = g.validateFinishTurn(c); g.Phase != Research {
		err = sn.NewVError(`Expected "Research" phase but have %q phase.`, g.Phase)
	}
	return
}

func (g *Game) companyExpansionNextPlayer(pers ...game.Playerer) *Player {
	g.CurrentPlayer().endOfTurnUpdate()
	p := g.nextPlayer(pers...)
	g.OverrideDeliveries = -1
	for !g.AllCompaniesOperated() {
		if !p.HasCompanyToOperate() {
			g.autoPass(p)
			p = g.nextPlayer(p)
		} else {
			return p
		}
	}
	return nil
}

func (g *Game) companyExpansionFinishTurn(c *gin.Context) (s *stats.Stats, cs contest.Contests, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if s, err = g.validateCompanyExpansionFinishTurn(c); err != nil {
		return
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	if np := g.companyExpansionNextPlayer(); np == nil {
		cs = g.startCityGrowth(c)
	} else {
		g.Phase = Operations
		g.SubPhase = OPSelectCompany
		g.resetShipping()
		np.beginningOfTurnReset()
		g.setCurrentPlayers(np)
	}
	return
}

func (g *Game) validateCompanyExpansionFinishTurn(c *gin.Context) (*stats.Stats, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	com := g.SelectedCompany()
	s, err := g.validateFinishTurn(c)
	switch {
	case err != nil:
		return nil, err
	case g.Phase != Operations:
		return nil, sn.NewVError("Expected %q phase but have %q phase.", Operations, g.PhaseName())
	case g.SubPhase != OPFreeExpansion && g.SubPhase != OPExpansion:
		return nil, sn.NewVError("Expected an expansion subphase but have %q subphase.", g.SubPhaseName())
	case com == nil:
		return nil, sn.NewVError("You must select a company to operate.")
	case !com.Operated:
		return nil, sn.NewVError("You must operate the selected company.")
	default:
		return s, nil
	}
}

func (g *Game) cityGrowthFinishTurn(c *gin.Context) (s *stats.Stats, cs contest.Contests, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if s, err = g.validateCityGrowthFinishTurn(c); err == nil {
		restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))
		cs = g.startNewEra(c)
	}
	return
}

func (g *Game) validateCityGrowthFinishTurn(c *gin.Context) (s *stats.Stats, err error) {
	cmap := g.CityGrowthMap()
	switch s, err = g.validateFinishTurn(c); {
	case err != nil:
	case g.Phase != CityGrowth:
		err = sn.NewVError("Expected %q phase but have %q phase.", CityGrowth, g.PhaseName())
	case g.C3StonesToUse(cmap) > 0:
		err = sn.NewVError("You did not select enough size 2 cities to grow.")
	case g.C2StonesToUse(cmap) > 0:
		err = sn.NewVError("You did not select enough size 1 cities to grow.")
	}
	return
}
