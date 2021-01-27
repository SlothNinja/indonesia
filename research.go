package indonesia

import (
	"encoding/gob"
	"html/template"
	"strings"

	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.Register(new(researchEntry))
}

type Technology int
type Technologies map[Technology]int

const (
	NoTech Technology = iota
	BidMultiplierTech
	SlotsTech
	MergersTech
	ExpansionsTech
	HullTech
)

var technologyStrings = map[Technology]string{
	NoTech:            "None",
	BidMultiplierTech: "Turn Order Bid",
	SlotsTech:         "Slots",
	MergersTech:       "Mergers",
	ExpansionsTech:    "Expansions",
	HullTech:          "Hull",
}

func (t Technology) String() string {
	return technologyStrings[t]
}

func (t Technology) Int() int {
	return int(t)
}

func (t Technology) IDString() string {
	return strings.Replace(strings.ToLower(t.String()), " ", "-", -1)
}

func (p *Player) Expansions() int {
	return p.Technologies[ExpansionsTech]
}

func (g *Game) startResearch(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g.Phase = Research
	g.beginningOfPhaseReset()
	g.setCurrentPlayers(g.Players()[0])
}

func (g *Game) conductResearch(c *gin.Context, cu *user.User) (tmpl string, err error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	var tech Technology

	cp := g.CurrentPlayer()
	switch tech, err = g.validateConductResearch(cu); {
	case err != nil:
	case tech == HullTech:
		g.SubPhase = RSelectPlayer
		tmpl = "indonesia/select_hull_player_dialog"
	case tech == SlotsTech:
		cp.Slots[cp.Technologies[SlotsTech]].Developed = true
		fallthrough
	default:
		cp.Technologies[tech] += 1
		cp.PerformedAction = true

		// Log
		e := g.newResearchEntryFor(cp, nil, tech, cp.Technologies[tech])
		restful.AddNoticef(c, string(e.HTML(c)))
		tmpl = "indonesia/research_update"
	}
	return
}

func (g *Game) validateConductResearch(cu *user.User) (Technology, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, tech, err := g.CurrentPlayer(), g.SelectedTechnology, g.validatePlayerAction(cu)
	switch {
	case err != nil:
		return NoTech, nil
	case tech < BidMultiplierTech || tech > HullTech:
		return NoTech, sn.NewVError("Received invalid for researched technology.")
	case tech != HullTech && cp.Technologies[tech] == 5:
		return NoTech, sn.NewVError("Your %s is already at the maximum level.", tech)
	default:
		return tech, nil
	}
}

type researchEntry struct {
	*Entry
	Technology Technology
	Level      int
}

func (g *Game) newResearchEntryFor(p, op *Player, t Technology, l int) (e *researchEntry) {
	if t == BidMultiplierTech {
		l = bidMultiplier[p.Technologies[BidMultiplierTech]]
	}

	e = &researchEntry{
		Entry:      g.newEntryFor(p),
		Technology: t,
		Level:      l,
	}
	if op != nil {
		e.SetOtherPlayer(op)
	}
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (e *researchEntry) HTML(c *gin.Context) (s template.HTML) {
	g := gameFrom(c)
	n := g.NameByPID(e.PlayerID)
	if e.OtherPlayerID == NoPlayerID {
		return restful.HTML("<div>%s increased %s to %d</div>", n, e.Technology, e.Level)
	} else {
		return restful.HTML("<div>%s increased %s of %s to %d</div>", n, e.Technology,
			g.NameByPID(e.OtherPlayerID), e.Level)
	}
}

func (g *Game) selectHullPlayer(c *gin.Context, cu *user.User) (tmpl string, act game.ActionType, err error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	var p *Player

	if p, err = g.validateSelectHullPlayer(c, cu); err != nil {
		tmpl, act = "indonesia/flash_notice", game.None
		return
	}

	cp := g.CurrentPlayer()
	p.Technologies[HullTech] += 1
	cp.PerformedAction = true

	// Log
	if cp.Equal(p) {
		e := g.newResearchEntryFor(cp, nil, HullTech, p.Technologies[HullTech])
		restful.AddNoticef(c, string(e.HTML(c)))
	} else {
		e := g.newResearchEntryFor(cp, p, HullTech, p.Technologies[HullTech])
		restful.AddNoticef(c, string(e.HTML(c)))
	}
	tmpl, act = "indonesia/research_update", game.Cache
	return
}

func (g *Game) validateSelectHullPlayer(c *gin.Context, cu *user.User) (*Player, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	if !g.IsCurrentPlayer(cu) {
		return nil, sn.NewVError("Only the current player can perform an action.")
	}

	p := g.PlayerBySID(c.PostForm("id"))
	switch {
	case p == nil:
		return nil, sn.NewVError("Received invalid player.")
	case p.Technologies[HullTech] == 5:
		return nil, sn.NewVError("Hull size of %s is already 5.", g.NameFor(p))
	default:
		return p, nil
	}
}
