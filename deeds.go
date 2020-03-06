package indonesia

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"strings"

	"github.com/SlothNinja/restful"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.Register(new(removeDeedsEntry))
}

type MaxShips map[Era]int

func (m MaxShips) String() string {
	return fmt.Sprintf("%d %d %d", m[EraA], m[EraB], m[EraC])
}

type Deed struct {
	Era      Era
	Province Province
	Goods    Goods
	MaxShips MaxShips
}

func newDeed(era Era, province Province, goods Goods, maxShips MaxShips) *Deed {
	return &Deed{Era: era, Province: province, Goods: goods, MaxShips: maxShips}
}

func maxShips(a, b, c int) MaxShips {
	return MaxShips{EraA: a, EraB: b, EraC: c}
}

func (d *Deed) IDString() string {
	return fmt.Sprintf("%s-%s-%s", d.Era, d.Province.IDString(), d.Goods.IDString())
}

func (d *Deed) Tip() template.HTML {
	return restful.HTML("{%q:%q, %q:%q, %q:%q}", "province", d.Province, "goods", d.Goods, "capacity", d.MaxShips)
}

func (g *Game) SelectedDeed() *Deed {
	if index := g.SelectedDeedIndex; index < 0 || index >= len(g.AvailableDeeds) {
		return nil
	} else {
		return g.AvailableDeeds[index]
	}
}

type Deeds []*Deed

var NoDeed = -1

func (ds Deeds) remove(d1 *Deed) Deeds {
	for i, d2 := range ds {
		if d1.Era == d2.Era && d1.Province == d2.Province && d1.Goods == d2.Goods {
			return ds.removeAt(i)
		}
	}
	return ds
}

func (ds Deeds) removeAt(i int) Deeds {
	return append(ds[:i], ds[i+1:]...)
}

func (g *Game) Deeds() Deeds {
	var ds Deeds
	for _, deeds := range deedsMap {
		ds = append(ds, deeds...)
	}
	return ds
}

func (ds Deeds) get(s string) *Deed {
	ss := strings.Split(s, "-")
	l := len(ss)
	if l < 3 {
		return nil
	}

	era, province, goods := ss[0], strings.Join(ss[1:l-1], "-"), ss[l-1]

	var deeds Deeds
	switch era {
	case EraA.String():
		deeds = deedsMap[EraA]
	case EraB.String():
		deeds = deedsMap[EraB]
	case EraC.String():
		deeds = deedsMap[EraC]
	default:
		return nil
	}

	for _, d := range deeds {
		if d.Province.IDString() == province && d.Goods.IDString() == goods {
			return d
		}
	}
	return nil
}

var deedsMap = map[Era]Deeds{
	EraA: Deeds{
		newDeed(EraA, Bali, Rice, nil),
		newDeed(EraA, Halmahera, Shipping, maxShips(3, 4, 5)),
		newDeed(EraA, Halmahera, Spice, nil),
		newDeed(EraA, JawaBarat, Rice, nil),
		newDeed(EraA, JawaTimur, Shipping, maxShips(2, 3, 3)),
		newDeed(EraA, Lampung, Shipping, maxShips(2, 3, 4)),
		newDeed(EraA, Maluku, Spice, nil),
		newDeed(EraA, SulawesiSelatan, Shipping, maxShips(3, 3, 4)),
	},
	EraB: Deeds{
		newDeed(EraB, Aceh, Rice, nil),
		newDeed(EraB, JawaBarat, Shipping, maxShips(0, 4, 5)),
		newDeed(EraB, JawaTengah, Spice, nil),
		newDeed(EraB, KalimantanBarat, Rubber, nil),
		newDeed(EraB, KalimantanTimur, Rice, nil),
		newDeed(EraB, Riau, Rubber, nil),
		newDeed(EraB, SulawesiTengah, Spice, nil),
		newDeed(EraB, SumateraBarat, Rubber, nil),
		newDeed(EraB, SumateraUtara, Shipping, maxShips(0, 4, 5)),
	},
	EraC: Deeds{
		newDeed(EraC, KalimantanSelatan, Oil, nil),
		newDeed(EraC, Maluku, Oil, nil),
		newDeed(EraC, Papua, Oil, nil),
		newDeed(EraC, Papua, Rubber, nil),
		newDeed(EraC, Sarawak, Oil, nil),
		newDeed(EraC, SulawesiTenggara, Rice, nil),
		newDeed(EraC, SumateraSelatan, Spice, nil),
	},
}

func deedsFor(era Era) Deeds {
	return deedsMap[era]
}

func (ds Deeds) Types() int {
	gm := make(map[Goods]bool, 0)
	for _, d := range ds {
		gm[d.Goods] = true
	}
	return len(gm)
}

func (ds Deeds) RemoveUnstartable(g *Game) Deeds {
	var deeds, removed Deeds
	for _, d := range ds {
		if d.Goods == Shipping {
			deeds = append(deeds, d)
		} else {
			if g.canStartGoodsInProvince(d.Goods, d.Province) {
				deeds = append(deeds, d)
			} else {
				removed = append(removed, d)
			}
		}
	}
	if len(removed) > 0 {
		g.newRemoveDeedsEntry(g.Era, removed)
	}
	return deeds
}

func (g *Game) canStartGoodsInProvince(goods Goods, p Province) bool {
	for _, area := range g.areasInProvince(p) {
		if !area.hasProducer() && !area.hasCity() && !area.adjacentAreaHasGoods(goods) {
			return true
		}
	}
	return false
}

func (a *Area) adjacentAreaHasGoods(g Goods) bool {
	for _, area := range a.AdjacentLandAreas() {
		if area.hasProducer() && area.Producer.Goods == g {
			return true
		}
	}
	return false
}

type removeDeedsEntry struct {
	*Entry
	Era   Era
	Deeds Deeds
}

func (g *Game) newRemoveDeedsEntry(era Era, deeds Deeds) (e *removeDeedsEntry) {
	e = &removeDeedsEntry{
		Entry: g.newEntry(),
		Era:   era,
		Deeds: deeds,
	}
	g.Log = append(g.Log, e)
	return
}

func (e *removeDeedsEntry) HTML(c *gin.Context) template.HTML {
	var s template.HTML
	for _, deed := range e.Deeds {
		s += restful.HTML("<div>No area in which to start a %s company in %s.  Deed discarded.</div>",
			deed.Goods, deed.Province)
	}
	return s
}
