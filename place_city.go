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
	gob.Register(new(placeCityEntry))
	gob.Register(new(discardCityEntry))
}

func (g *Game) placeCity(c *gin.Context, cu *user.User) (tmpl string, err error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	var (
		a      *Area
		c0, c1 *CityCard
	)

	if a, c0, c1, err = g.validatePlaceCity(c, cu); err != nil {
		tmpl = "indonesia/flash_notice"
		return
	}

	cp := g.CurrentPlayer()
	cp.PerformedAction = true
	switch {
	case c0 != nil && c1 == nil:
		a.City = newCity(a)
		g.CityStones[0] -= 1
		cp.CityCards = cp.CityCards[1:]
		// Log placement
		e := g.newPlaceCityEntryFor(cp, c0)
		restful.AddNoticef(c, string(e.HTML(c)))
		tmpl = "indonesia/place_city_update"
	case c0 == nil && c1 != nil:
		a.City = newCity(a)
		g.CityStones[0] -= 1
		cp.CityCards[1] = cp.CityCards[0]
		cp.CityCards = cp.CityCards[1:]
		// Log placement
		e := g.newPlaceCityEntryFor(cp, c1)
		restful.AddNoticef(c, string(e.HTML(c)))
		tmpl = "indonesia/place_city_update"
	default:
		cp.PerformedAction = false
		g.SubPhase = NESelectCard
		tmpl = "indonesia/select_card_dialog"
	}
	return
}

func (p *Player) cardsFor(a *Area) (c0, c1 *CityCard) {
	g := p.Game()
	switch cards := p.CardsForCurrentEra(); len(cards) {
	case 0:
		return nil, nil
	case 1:
		return cards[0], nil
	case 2:
		if g.newCityAreasFor(cards[0]).include(a) {
			c0 = cards[0]
		}
		if g.newCityAreasFor(cards[1]).include(a) {
			c1 = cards[1]
		}
	}
	return
}

func (g *Game) validatePlaceCity(c *gin.Context, cu *user.User) (a *Area, c0, c1 *CityCard, err error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	a = g.SelectedArea()
	cp := g.CurrentPlayer()
	c0, c1 = cp.cardsFor(a)

	switch err = g.validatePlayerAction(cu); {
	case err != nil:
	case a == nil:
		err = sn.NewVError("You must select an area.")
	case c0 == nil && c1 == nil:
		err = sn.NewVError("You don't have a city card for the selected area.")
	}
	return
}

type placeCityEntry struct {
	*Entry
	Area Area
	Card CityCard
}

func (g *Game) newPlaceCityEntryFor(p *Player, c *CityCard) (e *placeCityEntry) {
	area := g.SelectedArea()
	e = &placeCityEntry{
		Entry: g.newEntryFor(p),
		Area:  *area,
		Card:  *c,
	}
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (e *placeCityEntry) HTML(c *gin.Context) (s template.HTML) {
	g := gameFrom(c)
	s = restful.HTML("<div>%s used the following card to place city in %s.</div>",
		g.NameByPID(e.PlayerID), e.Area.Province())
	s += restful.HTML("<div class='top-padding'><img class='card' src='/images/indonesia/city-card-%s.png'/></div>", e.Card.IDString())
	return
}

func (g *Game) playCard(c *gin.Context, cu *user.User) (tmpl string, err error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	//	g.debugf("Play Card")
	var index int
	if index, err = g.validatePlayCard(c, cu); err != nil {
		tmpl = "indonesia/flash_notice"
		return
	}

	cp := g.CurrentPlayer()
	card := cp.CityCards[index]
	area := g.SelectedArea()
	area.City = newCity(area)
	g.CityStones[0] -= 1

	switch index {
	case 0:
		cp.CityCards = cp.CityCards[1:]
	default:
		cp.CityCards[1] = cp.CityCards[0]
		cp.CityCards = cp.CityCards[1:]
	}

	g.SubPhase = NoSubPhase
	cp.PerformedAction = true

	// Log placement
	e := g.newPlaceCityEntryFor(cp, card)
	restful.AddNoticef(c, string(e.HTML(c)))

	tmpl = "indonesia/place_city_update"
	return
}

func (g *Game) validatePlayCard(c *gin.Context, cu *user.User) (index int, err error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	index = g.SelectedCardIndex
	if err = g.validatePlayerAction(cu); g.SelectedCardIndex < 0 || g.SelectedCardIndex > 1 {
		err = sn.NewVError("Recieved invalid card index.")
	}
	return
}

type discardCityEntry struct {
	*Entry
	Card CityCard
}

func (g *Game) newDiscardCityEntryFor(p *Player, c *CityCard) (e *discardCityEntry) {
	e = &discardCityEntry{
		Entry: g.newEntryFor(p),
		Card:  *c,
	}
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (e *discardCityEntry) HTML(c *gin.Context) (s template.HTML) {
	g := gameFrom(c)
	s = restful.HTML("<div>%s unable to use the following card to place a city.</div>",
		g.NameByPID(e.PlayerID))
	s += restful.HTML("<div class='top-padding'><img class='card' src='/images/indonesia/city-card-%s.png'/></div>",
		e.Card.IDString())
	return
}
