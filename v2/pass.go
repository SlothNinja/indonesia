package indonesia

import (
	"encoding/gob"
	"html/template"

	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.Register(new(passEntry))
	gob.Register(new(autoPassEntry))
}

func (g *Game) pass(c *gin.Context, cu *user.User) (tmpl string, act game.ActionType, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if err = g.validatePass(c, cu); err != nil {
		log.Errorf(err.Error())
		tmpl, act = "indonesia/flash_notice", game.None
		return
	}

	cp := g.CurrentPlayer()
	cp.Passed = true
	cp.PerformedAction = true

	// Log Pass
	e := g.newPassEntryFor(cp)
	restful.AddNoticef(c, string(e.HTML(c)))

	tmpl, act = "indonesia/pass_update", game.Cache
	return
}

func (g *Game) validatePass(c *gin.Context, cu *user.User) (err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if err = g.validatePlayerAction(c, cu); err != nil {
		return
	}

	switch {
	case g.Phase == Acquisitions && g.SubPhase != NoSubPhase:
		err = sn.NewVError("You can not pass in SubPhase: %v", g.SubPhaseName())
	case g.Phase == Mergers && g.SubPhase != MSelectCompany1:
		err = sn.NewVError("You cannot pass in SubPhase: %v", g.SubPhaseName())
	case g.Phase != Acquisitions && g.Phase != Mergers:
		err = sn.NewVError("You cannot pass in Phase: %v", g.PhaseName())
	}
	return
}

type passEntry struct {
	*Entry
}

func (g *Game) newPassEntryFor(p *Player) (e *passEntry) {
	e = &passEntry{
		Entry: g.newEntryFor(p),
	}
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (e *passEntry) HTML(c *gin.Context) template.HTML {
	g := gameFrom(c)
	return restful.HTML("%s passed.", g.NameByPID(e.PlayerID))
}

func (g *Game) autoPass(p *Player) {
	p.PerformedAction = true
	p.Passed = true
	g.newAutoPassEntryFor(p)
}

type autoPassEntry struct {
	*Entry
}

func (g *Game) newAutoPassEntryFor(p *Player) (e *autoPassEntry) {
	e = new(autoPassEntry)
	e.Entry = g.newEntryFor(p)
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (e *autoPassEntry) HTML(c *gin.Context) template.HTML {
	g := gameFrom(c)
	return restful.HTML("System auto passed for %s.", g.NameByPID(e.PlayerID))
}
