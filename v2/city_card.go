package indonesia

import (
	"html/template"

	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
)

type CityCard struct {
	Era  Era
	Type int
}

type CityCards []*CityCard

func newADeck() CityCards {
	return CityCards{
		&CityCard{Era: EraA, Type: 1},
		&CityCard{Era: EraA, Type: 2},
		&CityCard{Era: EraA, Type: 3},
		&CityCard{Era: EraA, Type: 4},
		&CityCard{Era: EraA, Type: 5},
	}
}

func newBDeck() CityCards {
	return CityCards{
		&CityCard{Era: EraB, Type: 1},
		&CityCard{Era: EraB, Type: 2},
		&CityCard{Era: EraB, Type: 3},
		&CityCard{Era: EraB, Type: 4},
		&CityCard{Era: EraB, Type: 5},
	}
}

func newCDeck() CityCards {
	return CityCards{
		&CityCard{Era: EraC, Type: 1},
		&CityCard{Era: EraC, Type: 2},
		&CityCard{Era: EraC, Type: 3},
		&CityCard{Era: EraC, Type: 4},
		&CityCard{Era: EraC, Type: 5},
	}
}

func (c *CityCard) IDString() template.HTML {
	return restful.HTML("%s-%d", c.Era, c.Type)
}

func (cs *CityCards) draw() *CityCard {
	var card *CityCard
	*cs, card = cs.drawS()
	return card
}

func (cs CityCards) drawS() (CityCards, *CityCard) {
	i := sn.MyRand.Intn(len(cs))
	card := cs[i]
	cards := cs.removeAt(i)
	return cards, card
}

func (cs *CityCards) append(cards ...*CityCard) {
	*cs = cs.appendS(cards...)
}

func (cs CityCards) appendS(cards ...*CityCard) CityCards {
	if len(cards) == 0 {
		return cs
	}
	return append(cs, cards...)
}

func (cs CityCards) removeAt(i int) CityCards {
	return append(cs[:i], cs[i+1:]...)
}

func (g *Game) dealCityCards() {
	a := newADeck()
	b := newBDeck()
	c := newCDeck()
	for _, p := range g.Players() {
		if len(g.Players()) == 2 {
			p.CityCards = CityCards{a.draw(), a.draw(), b.draw(), b.draw(), c.draw(), c.draw()}
		} else {
			p.CityCards = CityCards{a.draw(), b.draw(), c.draw()}
		}
	}
}

var cityCardProvinces = map[Era]map[int]Provinces{
	EraA: {
		1: Provinces{JawaBarat, JawaTengah, SumateraSelatan},
		2: Provinces{JawaTimur, SulawesiSelatan, SumateraSelatan},
		3: Provinces{Bali, JawaTengah, SulawesiUtara},
		4: Provinces{JawaBarat, SulawesiSelatan, SulawesiUtara},
		5: Provinces{Bali, JawaBarat, JawaTimur},
	},
	EraB: {
		1: Provinces{Aceh, SumateraUtara, Bengkulu},
		2: Provinces{SumateraBarat, Lampung, KalimantanSelatan},
		3: Provinces{Aceh, Lampung, Maluku},
		4: Provinces{SumateraBarat, Bengkulu, JawaBarat},
		5: Provinces{SumateraUtara, KalimantanSelatan, Maluku},
	},
	EraC: {
		1: Provinces{Jambi, SulawesiTengah, NusaTenggaraTimur},
		2: Provinces{JawaBarat, NusaTenggaraTimur, Halmahera},
		3: Provinces{NusaTenggaraBarat, Halmahera, Papua},
		4: Provinces{Sarawak, SulawesiTengah, Papua},
		5: Provinces{Jambi, Sarawak, NusaTenggaraBarat},
	},
}

func (g *Game) areasInProvince(p Province) Areas {
	var areas Areas
	for _, area := range g.landAreas() {
		if area.Province() == p && (g.Version != 2 || area.ID != JawaTengah41) {
			areas = append(areas, area)
		}
	}
	return areas
}

func (g *Game) cityCardAreasForCard(c *CityCard) Areas {
	var areas Areas
	for _, p := range cityCardProvinces[c.Era][c.Type] {
		for _, area := range g.areasInProvince(p) {
			if !areas.include(area) {
				areas = append(areas, area)
			}
		}

	}
	return areas
}

func (g *Game) newCityAreasFor(cs ...*CityCard) Areas {
	var areas Areas
	for _, c := range cs {
		for _, area := range g.cityCardAreasForCard(c) {
			if !areas.include(area) && area.canHaveNewCity() {
				areas = append(areas, area)
			}
		}
	}
	return areas
}

func (a *Area) canHaveNewCity() bool {
	return a.Producer == nil && a.onShore() && !a.cityInProvince()
}

func (a *Area) cityInProvince() bool {
	return a.g.cityInProvince(a.Province())
}

func (g *Game) cityInProvince(p Province) bool {
	for _, a := range g.areasInProvince(p) {
		if a.City != nil {
			return true
		}
	}
	return false
}
