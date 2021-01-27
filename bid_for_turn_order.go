package indonesia

import (
	"encoding/gob"
	"html/template"
	"sort"
	"strconv"

	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.Register(new(bidEntry))
	gob.Register(new(turnOrderEntry))
}

const NoBid = -1

func (g *Game) startBidForTurnOrder(c *gin.Context) *Player {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g.Phase = BidForTurnOrder
	return g.Players()[0]
}

func (g *Game) placeTurnOrderBid(c *gin.Context, cu *user.User) (tmpl string, act game.ActionType, err error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	if err = g.validateBid(c, cu); err != nil {
		tmpl, act = "indonesia/flash_notice", game.None
		return
	}

	cp := g.CurrentPlayer()
	cp.Bank += cp.Bid
	cp.Rupiah -= cp.Bid
	cp.PerformedAction = true

	// Log placement
	e := g.newBidEntryFor(cp)
	restful.AddNoticef(c, string(e.HTML(c)))
	tmpl, act = "indonesia/turn_order_bid_update", game.Cache
	return
}

func (g *Game) validateBid(c *gin.Context, cu *user.User) (err error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	switch err = g.validatePlayerAction(cu); {
	case err != nil:
	default:
		cp := g.CurrentPlayer()
		switch cp.Bid, err = strconv.Atoi(c.PostForm("Bid")); {
		case err != nil:
		case cp.Bid > cp.Rupiah:
			err = sn.NewVError("You bid more than you have.")
		case cp.Bid < 0:
			err = sn.NewVError("You can't bid less than zero.")
		}
	}
	return
}

type bidEntry struct {
	*Entry
	Bid           int
	BidMultiplier int
}

func (g *Game) newBidEntryFor(p *Player) (e *bidEntry) {
	e = &bidEntry{
		Entry:         g.newEntryFor(p),
		Bid:           p.Bid,
		BidMultiplier: p.Multiplier(),
	}
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (e *bidEntry) HTML(c *gin.Context) template.HTML {
	g := gameFrom(c)
	return restful.HTML("<div>%s bid %d &times; %d for a total bid of %d</div>",
		g.NameByPID(e.PlayerID), e.Bid, e.BidMultiplier, e.Bid*e.BidMultiplier)
}

func (g *Game) setTurnOrder(c *gin.Context, cu *user.User) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	com, n := make([]int, g.NumPlayers), make([]int, g.NumPlayers)
	for i, p := range g.Players() {
		com[i] = p.ID()
	}

	ps := g.Players()
	b := make([]int, g.NumPlayers)
	sort.Sort(Reverse{ByTurnOrderBid{ps}})
	g.setPlayers(ps)
	cp := g.Players()[0]
	g.setCurrentPlayers(cp)

	// Log new order
	for i, p := range g.Players() {
		pid := p.ID()
		n[i] = pid
		b[pid] = p.TotalBid()
	}
	g.newTurnOrderEntry(com, n, b)
	g.startMergers(c, cu)
}

type turnOrderEntry struct {
	*Entry
	Current []int
	New     []int
	Bids    []int
}

func (g *Game) newTurnOrderEntry(c, n, b []int) {
	e := &turnOrderEntry{
		Entry:   g.newEntry(),
		Current: c,
		New:     n,
		Bids:    b,
	}
	g.Log = append(g.Log, e)
}

func (e *turnOrderEntry) HTML(c *gin.Context) template.HTML {
	g := gameFrom(c)
	s := restful.HTML("<div><table class='strippedDataTable'><thead><tr><th>Player</th><th>Bid</th></tr></thead><tbody>")
	for _, pid := range e.Current {
		s += restful.HTML("<tr><td>%s</td><td>%d</td></tr>", g.NameByPID(pid), e.Bids[pid])
	}
	s += restful.HTML("</tbody></table></div>")
	names := make([]string, g.NumPlayers)
	for i, pid := range e.New {
		names[i] = g.NameByPID(pid)
	}
	s += restful.HTML("<div class='top-padding'>New Turn Order: %s.</div>", restful.ToSentence(names))
	return s
}
