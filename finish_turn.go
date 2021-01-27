package indonesia

import (
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func (client *Client) finish(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		g := gameFrom(c)
		oldCP := g.CurrentPlayer()

		cu, err := client.User.Current(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param(hParam)))
			return
		}

		s, cs, err := client.finishTurn(c, g, cu)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param(hParam)))
			return
		}

		// Game is over if cs != nil
		if cs != nil {
			g.Phase = GameOver
			g.Status = game.Completed
			ks, es := wrap(s.GetUpdate(c, g.UpdatedAt), cs)
			err = client.saveWith(c, g, cu, ks, es)
			if err != nil {
				client.Log.Errorf(err.Error())
				c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param(hParam)))
				return
			}
			err = g.SendEndGameNotifications(c)
			if err != nil {
				client.Log.Warningf(err.Error())
			}
			c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param(hParam)))
			return
		}

		s = s.GetUpdate(c, g.UpdatedAt)
		err = client.saveWith(c, g, cu, []*datastore.Key{s.Key}, []interface{}{s})
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param(hParam)))
			return
		}

		newCP := g.CurrentPlayer()
		if newCP != nil && oldCP.ID() != newCP.ID() {
			err = g.SendTurnNotificationsTo(c, newCP)
			if err != nil {
				client.Log.Warningf(err.Error())
			}
		}

		defer c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param(hParam)))
	}
}

func (client *Client) finishTurn(c *gin.Context, g *Game, cu *user.User) (*user.Stats, []*contest.Contest, error) {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	switch {
	case g.Phase == NewEra:
		return client.newEraFinishTurn(c, g, cu)
	case g.Phase == BidForTurnOrder:
		return client.bidForTurnOrderFinishTurn(c, g, cu)
	case g.Phase == Mergers && g.SubPhase == MBid:
		return client.mergersBidFinishTurn(c, g, cu)
	case g.Phase == Mergers:
		return client.mergersFinishTurn(c, g, cu)
	case g.Phase == Acquisitions:
		return client.acquisitionsFinishTurn(c, g, cu)
	case g.Phase == Research:
		return client.researchFinishTurn(c, g, cu)
	case g.Phase == Operations:
		return client.companyExpansionFinishTurn(c, g, cu)
	case g.Phase == CityGrowth:
		return client.cityGrowthFinishTurn(c, g, cu)
	default:
		return nil, nil, sn.NewVError("Improper Phase for finishing turn.")
	}
}

func (g *Game) validateFinishTurn(c *gin.Context, cu *user.User) (*user.Stats, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp := g.CurrentPlayer()
	s := user.StatsFetched(c)
	switch {
	case s == nil:
		return nil, sn.NewVError("missing stats for player.")
	case !g.IsCurrentPlayer(cu):
		return nil, sn.NewVError("only the current player may finish a turn.")
	case !cp.PerformedAction:
		return nil, sn.NewVError("%s has yet to perform an action.", g.NameFor(cp))
	default:
		return s, nil
	}
}

// ps is an optional parameter.
// If no player is provided, assume current player.
func (g *Game) nextPlayer(cu *user.User, ps ...game.Playerer) *Player {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	nper := g.NextPlayerer(ps...)
	if nper != nil {
		return nper.(*Player)
	}
	return nil
}

func (g *Game) newEraNextPlayer(cu *user.User) *Player {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g.CurrentPlayer().endOfTurnUpdate()
	p := g.nextPlayer(cu)
	for g.Players().anyCanPlaceCity() {
		if p.CanPlaceCity() {
			p.beginningOfTurnReset()
			return p
		}
		p = g.nextPlayer(cu, p)
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

func (client *Client) newEraFinishTurn(c *gin.Context, g *Game, cu *user.User) (*user.Stats, []*contest.Contest, error) {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	s, err := g.validateNewEraFinishTurn(c, cu)
	if err != nil {
		return nil, nil, err
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	np := g.newEraNextPlayer(cu)
	if np == nil {
		for _, p := range g.Players() {
			g.removeUnplayableCityCardsFor(c, p)
		}
		np = g.startBidForTurnOrder(c)
	}
	g.setCurrentPlayers(np)
	return s, nil, nil
}

func (g *Game) validateNewEraFinishTurn(c *gin.Context, cu *user.User) (*user.Stats, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	s, err := g.validateFinishTurn(c, cu)
	if err != nil {
		return nil, err
	}

	if g.Phase != NewEra {
		return nil, sn.NewVError(`expected "New Era" phase but have %q phase.`, g.Phase)
	}
	return s, nil
}

func (g *Game) bidForTurnOrderNextPlayer(cu *user.User, pers ...game.Playerer) *Player {
	g.CurrentPlayer().endOfTurnUpdate()
	p := g.nextPlayer(cu, pers...)
	for !p.Equal(g.Players()[0]) {
		if !p.CanBid() {
			p = g.nextPlayer(cu, p)
		} else {
			return p
		}
	}
	return nil
}

func (client *Client) bidForTurnOrderFinishTurn(c *gin.Context, g *Game, cu *user.User) (*user.Stats, []*contest.Contest, error) {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	s, err := g.validateBidForTurnOrderFinishTurn(c, cu)
	if err != nil {
		return nil, nil, err
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	np := g.bidForTurnOrderNextPlayer(cu)
	if np == nil {
		g.setTurnOrder(c, cu)
		return s, nil, nil
	}
	g.setCurrentPlayers(np)
	return s, nil, nil
}

func (g *Game) validateBidForTurnOrderFinishTurn(c *gin.Context, cu *user.User) (s *user.Stats, err error) {
	if s, err = g.validateFinishTurn(c, cu); g.Phase != BidForTurnOrder {
		err = sn.NewVError(`Expected "Bid For Turn Order" phase but have %q phase.`, g.Phase)
	}
	return
}

func (g *Game) mergersBidNextPlayer(cu *user.User, pers ...game.Playerer) *Player {
	g.CurrentPlayer().endOfTurnUpdate()
	p := g.nextPlayer(cu, pers...)
	for !g.Players().allPassed() {
		if !p.CanBidOnMerger() {
			g.autoPass(p)
			p = g.nextPlayer(cu, p)
		} else {
			return p
		}
	}
	return nil
}

func (client *Client) mergersBidFinishTurn(c *gin.Context, g *Game, cu *user.User) (*user.Stats, []*contest.Contest, error) {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	s, err := g.validateMergersBidFinishTurn(c, cu)
	if err != nil {
		return nil, nil, err
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	np := g.mergersBidNextPlayer(cu)
	if np == nil {
		g.startMergerResolution(c, cu)
		return s, nil, nil
	}
	g.setCurrentPlayers(np)
	return s, nil, nil
}

func (g *Game) validateMergersBidFinishTurn(c *gin.Context, cu *user.User) (s *user.Stats, err error) {
	if s, err = g.validateFinishTurn(c, cu); g.Phase != Mergers {
		err = sn.NewVError(`Expected "Mergers" phase but have %q phase.`, g.Phase)
	}
	return
}

func (g *Game) mergersNextPlayer(cu *user.User, pers ...game.Playerer) *Player {
	g.CurrentPlayer().endOfTurnUpdate()
	p := g.nextPlayer(cu, pers...)
	for !g.Players().allPassed() {
		if !p.CanAnnounceMerger() {
			g.autoPass(p)
			p = g.nextPlayer(cu, p)
		} else {
			return p
		}
	}
	return nil
}

func (client *Client) mergersFinishTurn(c *gin.Context, g *Game, cu *user.User) (*user.Stats, []*contest.Contest, error) {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	s, err := g.validateMergersFinishTurn(c, cu)
	if err != nil {
		return nil, nil, err
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	if g.SubPhase == MSiapFajiCreation {
		announcer := g.PlayerByID(g.Merger.AnnouncerID)
		g.SiapFajiMerger = nil
		g.Merger = nil
		g.setCurrentPlayers(announcer)
		g.beginningOfPhaseReset()
		g.SubPhase = MSelectCompany1
		np := g.mergersNextPlayer(cu)
		if np != nil {
			g.setCurrentPlayers(np)
			return s, nil, nil
		}
		g.startAcquisitions(c, cu)
		return s, nil, nil
	}

	np := g.mergersNextPlayer(cu)
	if np == nil {
		g.startAcquisitions(c, cu)
		return s, nil, nil
	}
	g.setCurrentPlayers(np)
	return s, nil, nil
}

func (g *Game) validateMergersFinishTurn(c *gin.Context, cu *user.User) (*user.Stats, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	s, err := g.validateFinishTurn(c, cu)
	switch {
	case err != nil:
		return nil, err
	case g.Phase != Mergers:
		return nil, sn.NewVError(`Expected "Mergers" phase but have %q phase.`, g.Phase)
	case g.SubPhase == MSiapFajiCreation && g.SiapFajiMerger.GoodsToRemove() > 0:
		return nil, sn.NewVError("you must remove %d more rice/spice", g.SiapFajiMerger.GoodsToRemove())
	case g.SubPhase == MSiapFajiCreation && !g.SiapFajiMerger.Company().Zones.contiguous():
		return nil, sn.NewVError("each zone must be contiguous after removal.")
	default:
		return s, nil
	}
}

func (g *Game) acquisitionsNextPlayer(cu *user.User, pers ...game.Playerer) (p *Player) {
	g.CurrentPlayer().endOfTurnUpdate()
	p = g.nextPlayer(cu, pers...)
	for !g.Players().allPassed() {
		if !p.CanAcquireCompany() {
			g.autoPass(p)
			p = g.nextPlayer(cu, p)
		} else {
			return
		}
	}
	p = nil
	return
}

func (client *Client) acquisitionsFinishTurn(c *gin.Context, g *Game, cu *user.User) (*user.Stats, []*contest.Contest, error) {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	s, err := g.validateAcquisitionsFinishTurn(c, cu)
	if err != nil {
		return nil, nil, err
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	np := g.acquisitionsNextPlayer(cu)
	if np == nil {
		g.startResearch(c)
		return s, nil, nil
	}
	g.setCurrentPlayers(np)
	return s, nil, nil
}

func (g *Game) validateAcquisitionsFinishTurn(c *gin.Context, cu *user.User) (*user.Stats, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	s, err := g.validateFinishTurn(c, cu)
	if err != nil {
		return nil, err
	}
	if g.Phase != Acquisitions {
		return nil, sn.NewVError(`Expected "Acquisitions" phase but have %q phase.`, g.Phase)
	}
	return s, nil
}

func (g *Game) researchNextPlayer(cu *user.User, pers ...game.Playerer) *Player {
	g.CurrentPlayer().endOfTurnUpdate()
	p := g.nextPlayer(cu, pers...)
	for !p.Equal(g.Players()[0]) {
		if !p.CanResearch() {
			g.autoPass(p)
			p = g.nextPlayer(cu, p)
		} else {
			return p
		}
	}
	return nil
}

func (client *Client) researchFinishTurn(c *gin.Context, g *Game, cu *user.User) (*user.Stats, []*contest.Contest, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	s, err := g.validateResearchFinishTurn(c, cu)
	if err != nil {
		return nil, nil, err
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	np := g.researchNextPlayer(cu)
	if np == nil {
		cs, err := client.startOperations(c, g, cu)
		if err != nil {
			return nil, nil, err
		}
		return s, cs, nil
	}
	g.setCurrentPlayers(np)
	return s, nil, nil
}

func (g *Game) validateResearchFinishTurn(c *gin.Context, cu *user.User) (s *user.Stats, err error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	if s, err = g.validateFinishTurn(c, cu); g.Phase != Research {
		err = sn.NewVError(`Expected "Research" phase but have %q phase.`, g.Phase)
	}
	return
}

func (g *Game) companyExpansionNextPlayer(cu *user.User, pers ...game.Playerer) *Player {
	g.CurrentPlayer().endOfTurnUpdate()
	p := g.nextPlayer(cu, pers...)
	g.OverrideDeliveries = -1
	for !g.AllCompaniesOperated() {
		if !p.HasCompanyToOperate() {
			g.autoPass(p)
			p = g.nextPlayer(cu, p)
		} else {
			return p
		}
	}
	return nil
}

func (client *Client) companyExpansionFinishTurn(c *gin.Context, g *Game, cu *user.User) (*user.Stats, []*contest.Contest, error) {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	s, err := g.validateCompanyExpansionFinishTurn(c, cu)
	if err != nil {
		return nil, nil, err
	}

	restful.AddNoticef(c, "%s finished turn.", g.NameFor(g.CurrentPlayer()))

	np := g.companyExpansionNextPlayer(cu)
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

func (g *Game) validateCompanyExpansionFinishTurn(c *gin.Context, cu *user.User) (*user.Stats, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	com := g.SelectedCompany()
	s, err := g.validateFinishTurn(c, cu)
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

func (client *Client) cityGrowthFinishTurn(c *gin.Context, g *Game, cu *user.User) (*user.Stats, []*contest.Contest, error) {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	s, err := g.validateCityGrowthFinishTurn(c, cu)
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

func (g *Game) validateCityGrowthFinishTurn(c *gin.Context, cu *user.User) (s *user.Stats, err error) {
	cmap := g.CityGrowthMap()
	switch s, err = g.validateFinishTurn(c, cu); {
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
