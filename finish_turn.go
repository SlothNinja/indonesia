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

func (client Client) finish(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		g := gameFrom(c)
		oldCP := g.CurrentPlayer()

		s, cs, err := client.finishTurn(c, g)
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
			err = client.saveWith(c, g, ks, es)
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
		err = client.saveWith(c, g, []*datastore.Key{s.Key}, []interface{}{s})
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

func (client Client) finishTurn(c *gin.Context, g *Game) (*stats.Stats, contest.Contests, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	switch {
	case g.Phase == NewEra:
		return client.newEraFinishTurn(c, g)
	case g.Phase == BidForTurnOrder:
		return client.bidForTurnOrderFinishTurn(c, g)
	case g.Phase == Mergers && g.SubPhase == MBid:
		return client.mergersBidFinishTurn(c, g)
	case g.Phase == Mergers:
		return client.mergersFinishTurn(c, g)
	case g.Phase == Acquisitions:
		return client.acquisitionsFinishTurn(c, g)
	case g.Phase == Research:
		return client.researchFinishTurn(c, g)
	case g.Phase == Operations:
		return client.companyExpansionFinishTurn(c, g)
	case g.Phase == CityGrowth:
		return client.cityGrowthFinishTurn(c, g)
	default:
		return nil, nil, sn.NewVError("Improper Phase for finishing turn.")
	}
}

func (g *Game) validateFinishTurn(c *gin.Context) (*stats.Stats, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	cp := g.CurrentPlayer()
	s := stats.Fetched(c)
	switch {
	case s == nil:
		return nil, sn.NewVError("missing stats for player.")
	case !g.CUserIsCPlayerOrAdmin(c):
		return nil, sn.NewVError("only the current player may finish a turn.")
	case !cp.PerformedAction:
		return nil, sn.NewVError("%s has yet to perform an action.", g.NameFor(cp))
	default:
		return s, nil
	}
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

func (client Client) newEraFinishTurn(c *gin.Context, g *Game) (*stats.Stats, contest.Contests, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	s, err := g.validateNewEraFinishTurn(c)
	if err != nil {
		return nil, nil, err
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	np := g.newEraNextPlayer()
	if np == nil {
		for _, p := range g.Players() {
			g.removeUnplayableCityCardsFor(c, p)
		}
		g.startBidForTurnOrder(c)
	}
	g.setCurrentPlayers(np)
	return s, nil, nil
}

func (g *Game) validateNewEraFinishTurn(c *gin.Context) (*stats.Stats, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	s, err := g.validateFinishTurn(c)
	if err != nil {
		return nil, err
	}

	if g.Phase != NewEra {
		return nil, sn.NewVError(`expected "New Era" phase but have %q phase.`, g.Phase)
	}
	return s, nil
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

func (client Client) bidForTurnOrderFinishTurn(c *gin.Context, g *Game) (*stats.Stats, contest.Contests, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	s, err := g.validateBidForTurnOrderFinishTurn(c)
	if err != nil {
		return nil, nil, err
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	np := g.bidForTurnOrderNextPlayer()
	if np == nil {
		g.setTurnOrder(c)
		return s, nil, nil
	}
	g.setCurrentPlayers(np)
	return s, nil, nil
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

func (client Client) mergersBidFinishTurn(c *gin.Context, g *Game) (*stats.Stats, contest.Contests, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	s, err := g.validateMergersBidFinishTurn(c)
	if err != nil {
		return nil, nil, err
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	np := g.mergersBidNextPlayer()
	if np == nil {
		g.startMergerResolution(c)
		return s, nil, nil
	}
	g.setCurrentPlayers(np)
	return s, nil, nil
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

func (client Client) mergersFinishTurn(c *gin.Context, g *Game) (*stats.Stats, contest.Contests, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	s, err := g.validateMergersFinishTurn(c)
	if err != nil {
		return nil, nil, err
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	if g.SubPhase == MSiapFajiCreation {
		announcer := g.PlayerByID(g.Merger.AnnouncerID)
		g.Merger = nil
		g.setCurrentPlayers(announcer)
		g.beginningOfPhaseReset()
		g.SubPhase = MSelectCompany1
		np := g.mergersNextPlayer()
		if np != nil {
			g.setCurrentPlayers(np)
			return s, nil, nil
		}
		g.startAcquisitions(c)
		return s, nil, nil
	}

	np := g.mergersNextPlayer()
	if np == nil {
		g.startAcquisitions(c)
		return s, nil, nil
	}
	g.setCurrentPlayers(np)
	return s, nil, nil
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

func (client Client) acquisitionsFinishTurn(c *gin.Context, g *Game) (*stats.Stats, contest.Contests, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	s, err := g.validateAcquisitionsFinishTurn(c)
	if err != nil {
		return nil, nil, err
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	np := g.acquisitionsNextPlayer()
	if np == nil {
		g.startResearch(c)
		return s, nil, nil
	}
	g.setCurrentPlayers(np)
	return s, nil, nil
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

func (client Client) researchFinishTurn(c *gin.Context, g *Game) (*stats.Stats, contest.Contests, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	s, err := g.validateResearchFinishTurn(c)
	if err != nil {
		return nil, nil, err
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	np := g.researchNextPlayer()
	if np == nil {
		cs, err := client.startOperations(c, g)
		if err != nil {
			return nil, nil, err
		}
		return s, cs, nil
	}
	g.setCurrentPlayers(np)
	return s, nil, nil
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

func (client Client) companyExpansionFinishTurn(c *gin.Context, g *Game) (*stats.Stats, contest.Contests, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	s, err := g.validateCompanyExpansionFinishTurn(c)
	if err != nil {
		return nil, nil, err
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	np := g.companyExpansionNextPlayer()
	if np == nil {
		cs, err := client.startCityGrowth(c, g)
		if err != nil {
			return nil, nil, err
		}
		return s, cs, nil
	}

	g.Phase = Operations
	g.SubPhase = OPSelectCompany
	g.resetShipping()
	np.beginningOfTurnReset()
	g.setCurrentPlayers(np)
	return s, nil, nil
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

func (client Client) cityGrowthFinishTurn(c *gin.Context, g *Game) (*stats.Stats, contest.Contests, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	s, err := g.validateCityGrowthFinishTurn(c)
	if err != nil {
		return nil, nil, err
	}
	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))
	cs, err := client.startNewEra(c, g)
	if err != nil {
		return nil, nil, err
	}

	return s, cs, err
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
