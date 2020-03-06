package indonesia

import (
	"github.com/SlothNinja/restful"
)

type Goods int

const (
	Rice Goods = iota
	Spice
	Rubber
	Oil
	SiapFaji
	Shipping
	NoGoods Goods = -1
)

func (g *Game) ToGoods(i int) Goods {
	return Goods(i)
}

func (g *Game) ToShipType(i int) ShipType {
	return ShipType(i)
}

var goodsStrings = map[Goods]string{
	Rice:     "Rice",
	Spice:    "Spice",
	Rubber:   "Rubber",
	Oil:      "Oil",
	SiapFaji: "Siap Faji",
	Shipping: "Shipping",
	NoGoods:  "None",
}

func (g Goods) String() string {
	return goodsStrings[g]
}

func (g Goods) IDString() string {
	return restful.IDString(g.String())
}

func (g Goods) JSONString() string {
	return restful.JSONString(g.String())
}

var goodsPrice = map[Goods]int{
	Rice:     20,
	Spice:    25,
	Rubber:   30,
	Oil:      40,
	SiapFaji: 35,
	Shipping: 10,
	NoGoods:  0,
}

func (g Goods) Price() int {
	return goodsPrice[g]
}
