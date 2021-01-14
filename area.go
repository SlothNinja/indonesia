package indonesia

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/SlothNinja/color"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

type Province int
type Provinces []Province

const (
	NoProvince Province = iota
	Aceh
	Bali
	Bengkulu
	Halmahera
	KalimantanBarat
	KalimantanSelatan
	KalimantanTengah
	KalimantanTimur
	Jambi
	JawaBarat
	JawaTengah
	JawaTimur
	Lampung
	Maluku
	NusaTenggaraBarat
	NusaTenggaraTimur
	Papua
	Riau
	Sarawak
	SulawesiSelatan
	SulawesiTengah
	SulawesiTenggara
	SulawesiUtara
	SumateraBarat
	SumateraSelatan
	SumateraUtara
)

var provinceIDStrings = map[Province]string{
	NoProvince:        "None",
	Aceh:              "Aceh",
	Bali:              "Bali",
	Bengkulu:          "Bengkulu",
	Halmahera:         "Halmahera",
	KalimantanBarat:   "Kalimantan Barat",
	KalimantanSelatan: "Kalimantan Selatan",
	KalimantanTengah:  "Kalimantan Tengah",
	KalimantanTimur:   "Kalimantan Timur",
	Jambi:             "Jambi",
	JawaBarat:         "Jawa Barat",
	JawaTengah:        "Jawa Tengah",
	JawaTimur:         "Jawa Timur",
	Lampung:           "Lampung",
	Maluku:            "Maluku",
	NusaTenggaraBarat: "Nusa Tenggara Barat",
	NusaTenggaraTimur: "Nusa Tenggara Timur",
	Papua:             "Papua",
	Riau:              "Riau",
	Sarawak:           "Sarawak",
	SulawesiSelatan:   "Sulawesi Selatan",
	SulawesiTengah:    "Sulawesi Tengah",
	SulawesiTenggara:  "Sulawesi Tenggara",
	SulawesiUtara:     "Sulawesi Utara",
	SumateraBarat:     "Sumatera Barat",
	SumateraSelatan:   "Sumatera Selatan",
	SumateraUtara:     "Sumatera Utara",
}

func (p Province) String() string {
	return provinceIDStrings[p]
}

func (p Province) LString() string {
	return strings.ToLower(p.String())
}

func (p Province) IDString() string {
	return strings.Replace(p.LString(), " ", "-", -1)
}

func (p Province) Int() int {
	return int(p)
}

var provinceMap = map[AreaID]Province{
	Aceh0: Aceh,
	Aceh1: Aceh,
	Aceh2: Aceh,
	Aceh3: Aceh,

	SumateraUtara4: SumateraUtara,
	SumateraUtara5: SumateraUtara,
	SumateraUtara6: SumateraUtara,
	SumateraUtara7: SumateraUtara,

	Riau8:  Riau,
	Riau9:  Riau,
	Riau10: Riau,
	Riau11: Riau,
	Riau12: Riau,

	SumateraBarat13: SumateraBarat,
	SumateraBarat14: SumateraBarat,
	SumateraBarat15: SumateraBarat,
	SumateraBarat16: SumateraBarat,

	Jambi17: Jambi,
	Jambi18: Jambi,
	Jambi19: Jambi,

	Bengkulu20: Bengkulu,
	Bengkulu21: Bengkulu,
	Bengkulu22: Bengkulu,

	SumateraSelatan23: SumateraSelatan,
	SumateraSelatan24: SumateraSelatan,
	SumateraSelatan25: SumateraSelatan,
	SumateraSelatan26: SumateraSelatan,
	SumateraSelatan27: SumateraSelatan,
	SumateraSelatan28: SumateraSelatan,
	SumateraSelatan29: SumateraSelatan,

	Lampung30: Lampung,
	Lampung31: Lampung,
	Lampung32: Lampung,
	Lampung33: Lampung,

	JawaBarat34: JawaBarat,
	JawaBarat35: JawaBarat,
	JawaBarat36: JawaBarat,
	JawaBarat37: JawaBarat,
	JawaBarat38: JawaBarat,
	JawaBarat39: JawaBarat,
	JawaBarat40: JawaBarat,

	JawaTengah41: JawaTengah,
	JawaTengah42: JawaTengah,
	JawaTengah43: JawaTengah,
	JawaTimur44:  JawaTimur,
	JawaTimur45:  JawaTimur,
	JawaTimur46:  JawaTimur,
	JawaTimur47:  JawaTimur,
	JawaTimur48:  JawaTimur,
	JawaTimur49:  JawaTimur,

	Bali50: Bali,
	Bali51: Bali,

	NusaTenggaraBarat52: NusaTenggaraBarat,
	NusaTenggaraBarat53: NusaTenggaraBarat,
	NusaTenggaraTimur54: NusaTenggaraTimur,
	NusaTenggaraTimur55: NusaTenggaraTimur,
	NusaTenggaraTimur56: NusaTenggaraTimur,
	NusaTenggaraTimur57: NusaTenggaraTimur,
	NusaTenggaraTimur58: NusaTenggaraTimur,
	NusaTenggaraTimur59: NusaTenggaraTimur,

	Maluku60: Maluku,
	Maluku61: Maluku,
	Maluku62: Maluku,
	Maluku63: Maluku,
	Maluku64: Maluku,
	Maluku65: Maluku,
	Maluku66: Maluku,
	Maluku67: Maluku,
	Maluku68: Maluku,

	Papua69: Papua,
	Papua70: Papua,
	Papua71: Papua,
	Papua72: Papua,
	Papua73: Papua,
	Papua74: Papua,
	Papua75: Papua,

	Halmahera76: Halmahera,
	Halmahera77: Halmahera,
	Halmahera78: Halmahera,
	Halmahera79: Halmahera,
	Halmahera80: Halmahera,
	Halmahera81: Halmahera,

	SulawesiTenggara82: SulawesiTenggara,
	SulawesiTenggara83: SulawesiTenggara,
	SulawesiTenggara84: SulawesiTenggara,
	SulawesiSelatan85:  SulawesiSelatan,
	SulawesiSelatan86:  SulawesiSelatan,
	SulawesiSelatan87:  SulawesiSelatan,

	SulawesiTengah88: SulawesiTengah,
	SulawesiTengah89: SulawesiTengah,
	SulawesiTengah90: SulawesiTengah,
	SulawesiTengah91: SulawesiTengah,
	SulawesiTengah92: SulawesiTengah,

	SulawesiUtara93: SulawesiUtara,
	SulawesiUtara94: SulawesiUtara,

	KalimantanTimur95: KalimantanTimur,
	KalimantanTimur96: KalimantanTimur,
	KalimantanTimur97: KalimantanTimur,
	KalimantanTimur98: KalimantanTimur,
	KalimantanTimur99: KalimantanTimur,

	KalimantanSelatan100: KalimantanSelatan,
	KalimantanSelatan101: KalimantanSelatan,
	KalimantanSelatan102: KalimantanSelatan,

	KalimantanTengah103: KalimantanTengah,
	KalimantanTengah104: KalimantanTengah,
	KalimantanTengah105: KalimantanTengah,
	KalimantanTengah106: KalimantanTengah,
	KalimantanTengah107: KalimantanTengah,

	KalimantanBarat108: KalimantanBarat,
	KalimantanBarat109: KalimantanBarat,
	KalimantanBarat110: KalimantanBarat,

	Sarawak111: Sarawak,
	Sarawak112: Sarawak,
	Sarawak113: Sarawak,
	Sarawak114: Sarawak,

	Sea115: Aceh,
	Sea116: SumateraBarat,
	Sea117: JawaBarat,
	Sea118: JawaTimur,
	Sea119: NusaTenggaraBarat,
	Sea120: NusaTenggaraTimur,
	Sea121: Maluku,
	Sea122: Papua,
	Sea123: Papua,
	Sea124: Halmahera,
	Sea125: SulawesiUtara,
	Sea126: Sarawak,
	Sea127: SumateraUtara,
	Sea128: Riau,
	Sea129: SumateraSelatan,
	Sea130: KalimantanBarat,
	Sea131: Lampung,
	Sea132: JawaTengah,
	Sea133: KalimantanSelatan,
	Sea134: SulawesiTenggara,
	Sea135: SulawesiTengah,

	JawaTengah136: JawaTengah,
	JawaTengah137: JawaTengah,
}

func (a *Area) Province() Province {
	return provinceMap[a.ID]
}

type AreaID int
type AreaIDS []AreaID
type Areas []*Area

func (as AreaIDS) remove(aid AreaID) AreaIDS {
	aids := as
	for i, a := range as {
		if a == aid {
			return aids.removeAt(i)
		}
	}
	return as
}

func (as AreaIDS) removeAt(i int) AreaIDS {
	return append(as[:i], as[i+1:]...)
}

func (as Areas) IDS() AreaIDS {
	ids := make(AreaIDS, len(as))
	for i, a := range as {
		ids[i] = a.ID
	}
	return ids
}

func (as AreaIDS) same(aids AreaIDS) bool {
	if len(as) != len(aids) {
		return false
	}
	for _, aid := range aids {
		if !as.include(aid) {
			return false
		}
	}
	return true
}

func (ids AreaIDS) include(aid AreaID) bool {
	for _, id := range ids {
		if id == aid {
			return true
		}
	}
	return false
}

func (ids AreaIDS) addUnique(aids ...AreaID) AreaIDS {
	for _, aid := range aids {
		if !ids.include(aid) {
			ids = append(ids, aid)
		}
	}
	return ids
}

func (g *Game) adjacentAreaIDS(aid AreaID) (ids AreaIDS) {
	ids = adjacentAreasMap[aid]
	if g.Version == 2 {
		switch aid {
		case JawaBarat37:
			ids = AreaIDS{JawaBarat36, JawaBarat38, JawaBarat39, JawaTengah136, Sea131}
		case JawaBarat38:
			ids = AreaIDS{JawaBarat37, JawaBarat39, JawaTengah136, Sea117, Sea118}
		case JawaTengah41:
			ids = AreaIDS{}
		case JawaTengah42:
			ids = AreaIDS{JawaTengah137, JawaTengah43, JawaTimur44, JawaTimur49, Sea132}
		case JawaTengah43:
			ids = AreaIDS{JawaTengah137, JawaTengah42, JawaTimur49, Sea118}
		case JawaTengah136:
			ids = AreaIDS{JawaBarat37, JawaBarat38, JawaTengah137, Sea118, Sea131}
		case JawaTengah137:
			ids = AreaIDS{JawaTengah136, JawaTengah42, JawaTengah43, Sea118, Sea131, Sea132}
		case Sea118:
			ids = AreaIDS{JawaBarat38, JawaTengah136, JawaTengah137, JawaTengah43, JawaTimur46,
				JawaTimur47, JawaTimur48, JawaTimur49, Bali50, Sea117, Sea119}
		case Sea131:
			ids = AreaIDS{Lampung32, Lampung33, JawaBarat34, JawaBarat35, JawaBarat36,
				JawaBarat37, JawaTengah136, JawaTengah137, Sea117, Sea129, Sea130, Sea132}
		case Sea132:
			ids = AreaIDS{JawaTengah137, JawaTengah42, JawaTimur44, JawaTimur45, JawaTimur46,
				JawaTimur47, JawaTimur48, Bali50, Sea119, Sea130, Sea131, Sea133}
		}
	}
	return
}

func (g *Game) adjacentTo(ids AreaIDS, aid AreaID) bool {
	aids := g.adjacentAreaIDS(aid)
	for _, id := range ids {
		if aids.include(id) {
			return true
		}
	}
	return false
}

func (g *Game) contiguous(ids AreaIDS) bool {
	switch l := len(ids); {
	case l < 1:
		return false
	case l == 1:
		return true
	default:
		cids := make(AreaIDS, l-1)
		copy(cids, ids[1:])
		return g.contiguousRecursive(AreaIDS{ids[0]}, cids)
	}
}

func (g *Game) contiguousRecursive(ids, aids AreaIDS) bool {
	l := len(aids)
	for _, id := range aids {
		if g.adjacentTo(ids, id) {
			if l == 1 {
				return true
			}
			as := append(ids, id)
			is := aids.remove(id)
			return g.contiguousRecursive(as, is)
		}
	}
	return false
}

const (
	Aceh0 AreaID = iota
	Aceh1
	Aceh2
	Aceh3

	SumateraUtara4
	SumateraUtara5
	SumateraUtara6
	SumateraUtara7

	Riau8
	Riau9
	Riau10
	Riau11
	Riau12

	SumateraBarat13
	SumateraBarat14
	SumateraBarat15
	SumateraBarat16

	Jambi17
	Jambi18
	Jambi19

	Bengkulu20
	Bengkulu21
	Bengkulu22

	SumateraSelatan23
	SumateraSelatan24
	SumateraSelatan25
	SumateraSelatan26
	SumateraSelatan27
	SumateraSelatan28
	SumateraSelatan29

	Lampung30
	Lampung31
	Lampung32
	Lampung33

	JawaBarat34
	JawaBarat35
	JawaBarat36
	JawaBarat37
	JawaBarat38
	JawaBarat39
	JawaBarat40

	JawaTengah41
	JawaTengah42
	JawaTengah43

	JawaTimur44
	JawaTimur45
	JawaTimur46
	JawaTimur47
	JawaTimur48
	JawaTimur49

	Bali50
	Bali51

	NusaTenggaraBarat52
	NusaTenggaraBarat53

	NusaTenggaraTimur54
	NusaTenggaraTimur55
	NusaTenggaraTimur56
	NusaTenggaraTimur57
	NusaTenggaraTimur58
	NusaTenggaraTimur59

	Maluku60
	Maluku61
	Maluku62
	Maluku63
	Maluku64
	Maluku65
	Maluku66
	Maluku67
	Maluku68

	Papua69
	Papua70
	Papua71
	Papua72
	Papua73
	Papua74
	Papua75

	Halmahera76
	Halmahera77
	Halmahera78
	Halmahera79
	Halmahera80
	Halmahera81

	SulawesiTenggara82
	SulawesiTenggara83
	SulawesiTenggara84

	SulawesiSelatan85
	SulawesiSelatan86
	SulawesiSelatan87

	SulawesiTengah88
	SulawesiTengah89
	SulawesiTengah90
	SulawesiTengah91
	SulawesiTengah92

	SulawesiUtara93
	SulawesiUtara94

	KalimantanTimur95
	KalimantanTimur96
	KalimantanTimur97
	KalimantanTimur98
	KalimantanTimur99

	KalimantanSelatan100
	KalimantanSelatan101
	KalimantanSelatan102

	KalimantanTengah103
	KalimantanTengah104
	KalimantanTengah105
	KalimantanTengah106
	KalimantanTengah107

	KalimantanBarat108
	KalimantanBarat109
	KalimantanBarat110

	Sarawak111
	Sarawak112
	Sarawak113
	Sarawak114

	Sea115
	Sea116
	Sea117
	Sea118
	Sea119
	Sea120
	Sea121
	Sea122
	Sea123
	Sea124
	Sea125
	Sea126
	Sea127
	Sea128
	Sea129
	Sea130
	Sea131
	Sea132
	Sea133
	Sea134
	Sea135

	JawaTengah136
	JawaTengah137

	NoArea    AreaID = -1
	sourceAID AreaID = -10
	targetAID AreaID = -20
	LandFirst AreaID = Aceh0
	LandLast  AreaID = Sarawak114
	SeaFirst  AreaID = Sea115
	SeaLast   AreaID = Sea135
)

func (g *Game) landIDS() (ids AreaIDS) {
	ids = landIDS
	if g.Version == 2 {
		ids = append(ids, JawaTengah136, JawaTengah137)
	}
	return ids
}

func (g *Game) isLandID(id AreaID) bool {
	return g.landIDS().include(id)
}

func (g *Game) areaIDS() (ids AreaIDS) {
	ids = append(landIDS, seaIDS...)
	if g.Version == 2 {
		ids = append(ids, JawaTengah136, JawaTengah137)
	}
	return
}

func (g *Game) isSeaID(id AreaID) bool {
	return seaIDS.include(id)
}

func (g *Game) seaAreas() (as Areas) {
	as = make(Areas, len(seaIDS))
	for i, id := range seaIDS {
		as[i] = g.GetArea(id)
	}
	return
}

func (g *Game) landAreas() (as Areas) {
	ids := g.landIDS()
	as = make(Areas, len(ids))
	for i, id := range ids {
		as[i] = g.GetArea(id)
	}
	return
}

// func (a *Area) GoodsColor() color.Color {
// 	company := a.GoodsCompany()
// 	if company == nil {
// 		return color.Black
// 	}
// 	company.g = a.g
// 	owner := company.Owner()
// 	if owner == nil {
// 		return color.Black
// 	}
// 	return owner.Color()
// }

func (g *Game) GoodsColor(a *Area, cu *user.User) color.Color {
	company := a.GoodsCompany()
	if company == nil {
		return color.Black
	}
	company.g = g
	owner := company.Owner()
	if owner == nil {
		return color.Black
	}
	return g.Color(owner, cu)
}

var landIDS = AreaIDS{Aceh0, Aceh1, Aceh2, Aceh3, SumateraUtara4, SumateraUtara5, SumateraUtara6, SumateraUtara7,
	Riau8, Riau9, Riau10, Riau11, Riau12, SumateraBarat13, SumateraBarat14, SumateraBarat15, SumateraBarat16,
	Jambi17, Jambi18, Jambi19, Bengkulu20, Bengkulu21, Bengkulu22, SumateraSelatan23, SumateraSelatan24,
	SumateraSelatan25, SumateraSelatan26, SumateraSelatan27, SumateraSelatan28, SumateraSelatan29,
	Lampung30, Lampung31, Lampung32, Lampung33, JawaBarat34, JawaBarat35, JawaBarat36, JawaBarat37,
	JawaBarat38, JawaBarat39, JawaBarat40, JawaTengah41, JawaTengah42, JawaTengah43,
	JawaTimur44, JawaTimur45, JawaTimur46, JawaTimur47, JawaTimur48, JawaTimur49,
	Bali50, Bali51, NusaTenggaraBarat52, NusaTenggaraBarat53, NusaTenggaraTimur54, NusaTenggaraTimur55,
	NusaTenggaraTimur56, NusaTenggaraTimur57, NusaTenggaraTimur58, NusaTenggaraTimur59, Maluku60,
	Maluku61, Maluku62, Maluku63, Maluku64, Maluku65, Maluku66, Maluku67, Maluku68,
	Papua69, Papua70, Papua71, Papua72, Papua73, Papua74, Papua75, Halmahera76, Halmahera77,
	Halmahera78, Halmahera79, Halmahera80, Halmahera81, SulawesiTenggara82, SulawesiTenggara83,
	SulawesiTenggara84, SulawesiSelatan85, SulawesiSelatan86, SulawesiSelatan87, SulawesiTengah88,
	SulawesiTengah89, SulawesiTengah90, SulawesiTengah91, SulawesiTengah92, SulawesiUtara93,
	SulawesiUtara94, KalimantanTimur95, KalimantanTimur96, KalimantanTimur97, KalimantanTimur98,
	KalimantanTimur99, KalimantanSelatan100, KalimantanSelatan101, KalimantanSelatan102, KalimantanTengah103,
	KalimantanTengah104, KalimantanTengah105, KalimantanTengah106, KalimantanTengah107, KalimantanBarat108,
	KalimantanBarat109, KalimantanBarat110, Sarawak111, Sarawak112, Sarawak113, Sarawak114}

var seaIDS = AreaIDS{Sea115, Sea116, Sea117, Sea118, Sea119, Sea120, Sea121, Sea122, Sea123, Sea124,
	Sea125, Sea126, Sea127, Sea128, Sea129, Sea130, Sea131, Sea132, Sea133, Sea134, Sea135}

type Area struct {
	g        *Game
	ID       AreaID
	City     *City
	Producer *Producer
	Shippers Shippers
	Used     bool
}

func (c *City) copy() *City {
	city := &City{Size: c.Size}
	copy(city.Delivered, c.Delivered)
	return city
}

func (g *Game) ProducedGoods() []bool {
	produced := make([]bool, 5)
	for _, company := range g.Companies() {
		if goods := company.Goods(); goods >= 0 && goods < 5 {
			produced[goods] = true
		}
	}
	return produced
}

func (a *Area) hasCity() bool {
	return a.City != nil
}

func (a *Area) AddProducer(c *Company) {
	a.Producer = &Producer{
		g:       c.g,
		OwnerID: c.OwnerID,
		Slot:    c.Slot,
		Goods:   c.Goods(),
	}
}

func (a *Area) AddShip(c *Company) {
	a.Shippers = append(a.Shippers, &Shipper{
		g:        c.g,
		OwnerID:  c.OwnerID,
		Slot:     c.Slot,
		ShipType: c.ShipType,
	})
}

func (a *Area) Key(c *gin.Context) template.HTML {
	admin := game.AdminFrom(c)
	if admin {
		return restful.HTML("admin-area-%d", a.ID)
	} else {
		return restful.HTML("area-%d", a.ID)
	}
}

func (a *Area) GoodsCompany() *Company {
	if producer := a.Producer; producer == nil {
		return nil
	} else {
		return producer.Company()
	}
}

func (c *Company) HullSize() int {
	if c == nil {
		return 0
	}
	if c.Goods() != Shipping {
		return 0
	}
	if owner := c.Owner(); owner == nil {
		return 0
	} else {
		return owner.Technologies[HullTech]
	}
}

func (a *Area) Tooltip() template.HTML {
	if a.Producer != nil {
		owner := a.g.PlayerByID(a.Producer.OwnerID)
		slot := a.Producer.Slot
		return restful.HTML(`{%q:%q, %q:"%d"}`, "owner", a.g.NameFor(owner), "slot", slot)
	} else {
		return restful.HTML("")
	}
}

func (a *Area) ShipTip(t ShipType) template.HTML {
	if len(a.Shippers) > 0 {
		for _, shipper := range a.Shippers {
			if shipper.ShipType == t {
				owner := a.g.PlayerByID(shipper.OwnerID)
				slot := shipper.Slot
				hull := shipper.HullSize()
				delivered := shipper.Delivered
				return restful.HTML("{%q:%q, %q:\"%d\", %q:\"%d\", %q:\"%d\"}",
					"owner", a.g.NameFor(owner),
					"slot", slot,
					"hull", hull,
					"delivered", delivered)
			}
		}
	}
	return restful.HTML("")
}

func (a *Area) CityTip() template.HTML {
	if a.City == nil {
		return restful.HTML("")
	}
	d := a.City.Delivered
	tip := restful.HTML("{%q:%q", "province", a.Province())
	count := 0
	for goodsInt, produced := range a.g.ProducedGoods() {
		if produced {
			goods := Goods(goodsInt)
			tip += restful.HTML(", %q:%d", goods.JSONString(), d[goods])
			count += 1
		}
	}
	return tip + restful.HTML("}")
}

func (as Areas) ids() string {
	s := ""
	for _, a := range as {
		s += fmt.Sprintf("%d:", a.ID)
	}
	return s
}

func (a *Area) Game() *Game {
	return a.g
}

func (g *Game) SelectedArea() *Area {
	return g.GetArea(g.SelectedAreaID)
}

func (g *Game) SelectedArea2() *Area {
	return g.GetArea(g.SelectedArea2ID)
}

func (g *Game) SelectedGoodsArea() *Area {
	return g.GetArea(g.SelectedGoodsAreaID)
}

func (g *Game) OldSelectedArea() *Area {
	return g.GetArea(g.OldSelectedAreaID)
}

func (g *Game) GetArea(aid AreaID) (a *Area) {
	if g.isLandID(aid) || g.isSeaID(aid) {
		a = g.Areas[aid]
	}
	return
}

func (a *Area) init(g *Game) {
	a.g = g
	if a.Producer != nil {
		a.Producer.init(g)
	}
	for _, shipper := range a.Shippers {
		shipper.init(g, a)
	}
	if a.hasCity() {
		a.City.init(a)
	}
}

const NoSlot = 0

func (g *Game) newArea(id AreaID) *Area {
	return &Area{g: g, ID: id}
}

func (g *Game) createAreas() {
	ids := g.areaIDS()

	g.Areas = make(Areas, len(ids))
	for i, id := range ids {
		g.Areas[i] = g.newArea(id)
	}
}

func (g *Game) initAreas() {
	for _, a := range g.Areas {
		a.init(g)
	}
}

func (a *Area) adjacentAreaHasCompetingCompanyFor(c *Company) bool {
	goods := c.Goods()
	for _, area := range a.AdjacentAreas() {
		company := area.GoodsCompany()
		if company != nil && company != c && company.Goods() == goods {
			return true
		}
	}
	return false
}

func (a *Area) IsLand() bool {
	if a.g.Version == 2 {
		return (a.ID >= LandFirst && a.ID <= LandLast) || a.ID == JawaTengah136 || a.ID == JawaTengah137
	}
	return a.ID >= LandFirst && a.ID <= LandLast
}

func (a *Area) IsSea() bool {
	return a.ID >= SeaFirst && a.ID <= SeaLast
}

func (a *Area) onShore() bool {
	return a.IsLand() && a.AdjacentSeaAreas() != nil
}

type addAreaTest func(a *Area) bool

func (a *Area) AdjacentAreas() Areas {
	return a.adjacentAreas()
}

func (a *Area) adjacentAreas(tests ...addAreaTest) Areas {
	var areas Areas
	for _, id := range a.g.adjacentAreaIDS(a.ID) {
		if area := a.g.Areas[id]; area.add(tests...) {
			areas = append(areas, a.g.Areas[id])
		}
	}
	return areas
}

func (a *Area) add(tests ...addAreaTest) bool {
	for _, test := range tests {
		if test(a) == false {
			return false
		}
	}
	return true
}

func isSea(a *Area) bool {
	return a.IsSea()
}

func isLand(a *Area) bool {
	return a.IsLand()
}

func hasCity(a *Area) bool {
	return a.hasCity()
}

func (a *Area) AdjacentSeaAreas() Areas {
	return a.adjacentAreas(isSea)
}

func (a *Area) AdjacentLandAreas() Areas {
	return a.adjacentAreas(isLand)
}

func (a *Area) AdjacentCityAreas() Areas {
	return a.adjacentAreas(hasCity)
}

func (a *Area) adjacentToArea(area *Area) bool {
	return a.AdjacentAreas().include(area)
}

func (a *Area) adjacentToProvince(p Province) bool {
	for _, area := range a.g.areasInProvince(p) {
		if a.adjacentToArea(area) {
			return true
		}
	}
	return false
}

func (a *Area) adjacentToZoneFor(area *Area) bool {
	if a.g != nil && a.g.SelectedCompany() != nil {
		for _, za := range a.g.SelectedCompany().ZoneFor(area).Areas() {
			if za.adjacentToArea(a) {
				return true
			}
		}
	}
	return false
}

var adjacentAreasMap = map[AreaID]AreaIDS{
	Aceh0: AreaIDS{Aceh1, Aceh2, Sea115},
	Aceh1: AreaIDS{Aceh0, Aceh2, SumateraUtara4, Sea115, Sea127},
	Aceh2: AreaIDS{Aceh0, Aceh1, Aceh3, SumateraUtara6, Sea115},
	Aceh3: AreaIDS{Aceh2, SumateraUtara7, Sea115},

	SumateraUtara4: AreaIDS{Aceh1, SumateraUtara5, SumateraUtara6, Sea127},
	SumateraUtara5: AreaIDS{SumateraUtara4, SumateraUtara6, Riau8, Riau12, Sea127},
	SumateraUtara6: AreaIDS{Aceh2, SumateraUtara4, SumateraUtara5, SumateraUtara7, Riau12,
		SumateraBarat13, Sea115, Sea116},
	SumateraUtara7: AreaIDS{Aceh3, SumateraUtara6, SumateraBarat16, Sea116},

	Riau8:  AreaIDS{SumateraUtara5, Riau10, Riau11, Riau12, Sea127, Sea128},
	Riau9:  AreaIDS{Riau10, Sea128},
	Riau10: AreaIDS{Riau8, Riau9, Riau11, SumateraBarat14, Jambi17, Jambi18, Sea128},
	Riau11: AreaIDS{Riau8, Riau10, Riau12, SumateraBarat13, SumateraBarat14},
	Riau12: AreaIDS{SumateraUtara5, SumateraUtara6, Riau8, Riau11, SumateraBarat13},

	SumateraBarat13: AreaIDS{SumateraUtara6, Riau11, Riau12, SumateraBarat14, Sea116},
	SumateraBarat14: AreaIDS{Riau10, Riau11, SumateraBarat13, SumateraBarat15, Jambi17, Bengkulu20, Sea116},
	SumateraBarat15: AreaIDS{SumateraBarat14, SumateraBarat16, Sea116},
	SumateraBarat16: AreaIDS{SumateraUtara7, SumateraBarat15, Sea116},

	Jambi17: AreaIDS{Riau10, SumateraBarat14, Jambi18, Bengkulu20, Bengkulu21, SumateraSelatan23},
	Jambi18: AreaIDS{Riau10, Jambi17, Jambi19, SumateraSelatan24, Sea128},
	Jambi19: AreaIDS{Jambi18, SumateraSelatan24, Sea128, Sea129},

	Bengkulu20: AreaIDS{SumateraBarat14, Jambi17, Bengkulu21, Sea116},
	Bengkulu21: AreaIDS{Jambi17, Bengkulu20, Bengkulu22, SumateraSelatan23, SumateraSelatan29, Sea116},
	Bengkulu22: AreaIDS{Bengkulu21, SumateraSelatan29, Lampung30, Sea116, Sea117},

	SumateraSelatan23: AreaIDS{Jambi17, Bengkulu21, SumateraSelatan24, SumateraSelatan28, SumateraSelatan29},
	SumateraSelatan24: AreaIDS{Jambi18, Jambi19, SumateraSelatan23, SumateraSelatan27, SumateraSelatan28, Sea129},
	SumateraSelatan25: AreaIDS{SumateraSelatan26, SumateraSelatan27, Sea129},
	SumateraSelatan26: AreaIDS{SumateraSelatan25, Sea130},
	SumateraSelatan27: AreaIDS{SumateraSelatan24, SumateraSelatan25, SumateraSelatan28, Lampung31,
		Lampung32, Sea129},
	SumateraSelatan28: AreaIDS{SumateraSelatan23, SumateraSelatan24, SumateraSelatan27, SumateraSelatan29,
		Lampung31},
	SumateraSelatan29: AreaIDS{Bengkulu21, Bengkulu22, SumateraSelatan23, SumateraSelatan28, Lampung30},

	Lampung30: AreaIDS{Bengkulu22, SumateraSelatan29, Lampung31, Lampung33, Sea117},
	Lampung31: AreaIDS{SumateraSelatan28, SumateraSelatan27, Lampung30, Lampung32, Lampung33},
	Lampung32: AreaIDS{SumateraSelatan27, Lampung31, Lampung33, Sea129, Sea131},
	Lampung33: AreaIDS{Lampung30, Lampung31, Lampung32, Sea117, Sea131},

	JawaBarat34: AreaIDS{JawaBarat35, JawaBarat40, Sea117, Sea131},
	JawaBarat35: AreaIDS{JawaBarat34, JawaBarat36, JawaBarat40, Sea131},
	JawaBarat36: AreaIDS{JawaBarat35, JawaBarat37, JawaBarat39, JawaBarat40, Sea131},
	JawaBarat37: AreaIDS{JawaBarat36, JawaBarat38, JawaBarat39, JawaTengah41, Sea131},
	JawaBarat38: AreaIDS{JawaBarat37, JawaBarat39, JawaTengah41, Sea117, Sea118},
	JawaBarat39: AreaIDS{JawaBarat36, JawaBarat37, JawaBarat38, JawaBarat40, Sea117},
	JawaBarat40: AreaIDS{JawaBarat34, JawaBarat35, JawaBarat36, JawaBarat39, Sea117},

	JawaTengah41: AreaIDS{JawaBarat37, JawaBarat38, JawaTengah42, JawaTengah43, Sea118, Sea131, Sea132},
	JawaTengah42: AreaIDS{JawaTengah41, JawaTengah43, JawaTimur44, JawaTimur49, Sea132},
	JawaTengah43: AreaIDS{JawaTengah41, JawaTengah42, JawaTimur49, Sea118},

	JawaTimur44: AreaIDS{JawaTengah42, JawaTimur45, JawaTimur48, JawaTimur49, Sea132},
	JawaTimur45: AreaIDS{JawaTimur44, Sea132},
	JawaTimur46: AreaIDS{JawaTimur47, Bali50, Sea118, Sea132},
	JawaTimur47: AreaIDS{JawaTimur46, JawaTimur48, Sea118, Sea132},
	JawaTimur48: AreaIDS{JawaTimur44, JawaTimur47, JawaTimur49, Sea118, Sea132},
	JawaTimur49: AreaIDS{JawaTengah42, JawaTengah43, JawaTimur44, JawaTimur48, Sea118},

	Bali50: AreaIDS{JawaTimur46, Bali51, Sea118, Sea119, Sea132},
	Bali51: AreaIDS{Bali50, NusaTenggaraBarat52, Sea119},

	NusaTenggaraBarat52: AreaIDS{Bali51, NusaTenggaraBarat53, Sea119},
	NusaTenggaraBarat53: AreaIDS{NusaTenggaraBarat52, NusaTenggaraTimur55, Sea119},

	NusaTenggaraTimur54: AreaIDS{NusaTenggaraTimur55, Sea119},
	NusaTenggaraTimur55: AreaIDS{NusaTenggaraBarat53, NusaTenggaraTimur54, NusaTenggaraTimur56, Sea120},
	NusaTenggaraTimur56: AreaIDS{NusaTenggaraTimur55, NusaTenggaraTimur58, Sea120},
	NusaTenggaraTimur57: AreaIDS{NusaTenggaraTimur58, Sea120},
	NusaTenggaraTimur58: AreaIDS{NusaTenggaraTimur56, NusaTenggaraTimur57, NusaTenggaraTimur59, Maluku60, Sea120},
	NusaTenggaraTimur59: AreaIDS{NusaTenggaraTimur58, Sea120},

	Maluku60: AreaIDS{NusaTenggaraTimur58, Maluku61, Sea121},
	Maluku61: AreaIDS{Maluku60, Maluku62, Sea122},
	Maluku62: AreaIDS{Maluku61, Papua70, Sea122},
	Maluku63: AreaIDS{Maluku64, Sea121},
	Maluku64: AreaIDS{Maluku63, Maluku65, Sea121},
	Maluku65: AreaIDS{Maluku64, Maluku66, Sea121},
	Maluku66: AreaIDS{Maluku65, Maluku67, Sea121},
	Maluku67: AreaIDS{Maluku66, Maluku68, Sea124},
	Maluku68: AreaIDS{Maluku67, Sea124},

	Papua69: AreaIDS{Papua70, Papua75, Sea122},
	Papua70: AreaIDS{Maluku62, Papua69, Papua71, Papua75, Sea122},
	Papua71: AreaIDS{Papua70, Papua72, Papua75, Sea122, Sea123},
	Papua72: AreaIDS{Papua71, Papua73, Papua74, Sea123},
	Papua73: AreaIDS{Papua72, Sea123},
	Papua74: AreaIDS{Papua72, Sea123},
	Papua75: AreaIDS{Papua69, Papua70, Papua71, Sea123},

	Halmahera76: AreaIDS{Halmahera77, Halmahera78, Sea124},
	Halmahera77: AreaIDS{Halmahera76, Halmahera78, Sea124},
	Halmahera78: AreaIDS{Halmahera76, Halmahera77, Halmahera79, Halmahera81, Sea124},
	Halmahera79: AreaIDS{Halmahera78, Halmahera80, Sea124},
	Halmahera80: AreaIDS{Halmahera79, Sea124},
	Halmahera81: AreaIDS{Halmahera78, Sea124},

	SulawesiTenggara82: AreaIDS{SulawesiTenggara83, Sea134},
	SulawesiTenggara83: AreaIDS{SulawesiTenggara82, SulawesiTenggara84, Sea134},
	SulawesiTenggara84: AreaIDS{SulawesiTenggara83, SulawesiSelatan86, SulawesiTengah90, Sea134, Sea135},

	SulawesiSelatan85: AreaIDS{SulawesiSelatan86, Sea134},
	SulawesiSelatan86: AreaIDS{SulawesiTenggara84, SulawesiSelatan85, SulawesiSelatan87, SulawesiTengah89,
		SulawesiTengah90, Sea125, Sea134},
	SulawesiSelatan87: AreaIDS{SulawesiSelatan86, SulawesiTengah89, Sea125},

	SulawesiTengah88: AreaIDS{SulawesiTengah89, SulawesiUtara93, Sea125, Sea135},
	SulawesiTengah89: AreaIDS{SulawesiSelatan86, SulawesiSelatan87, SulawesiTengah88, SulawesiTengah90,
		Sea125, Sea135},
	SulawesiTengah90: AreaIDS{SulawesiTenggara84, SulawesiSelatan86, SulawesiTengah89, SulawesiTengah91,
		SulawesiTengah92, Sea135},
	SulawesiTengah91: AreaIDS{SulawesiTengah90, Sea135},
	SulawesiTengah92: AreaIDS{SulawesiTengah90, SulawesiUtara93, Sea135},

	SulawesiUtara93: AreaIDS{SulawesiTengah88, SulawesiTengah92, SulawesiUtara94, Sea125, Sea135},
	SulawesiUtara94: AreaIDS{SulawesiUtara93, Sea135},

	KalimantanTimur95: AreaIDS{KalimantanTimur96, KalimantanTimur97, Sarawak112, Sarawak114, Sea125},
	KalimantanTimur96: AreaIDS{KalimantanTimur95, KalimantanTimur97, KalimantanTimur98, Sea125},
	KalimantanTimur97: AreaIDS{KalimantanTimur95, KalimantanTimur96, KalimantanTimur98, KalimantanTengah103,
		KalimantanBarat108, Sarawak111, Sarawak112},
	KalimantanTimur98: AreaIDS{KalimantanTimur96, KalimantanTimur97, KalimantanTimur99, KalimantanTengah103, KalimantanTengah107, Sea125},
	KalimantanTimur99: AreaIDS{KalimantanTimur98, KalimantanSelatan100, KalimantanTengah107, Sea125, Sea133},

	KalimantanSelatan100: AreaIDS{KalimantanTimur99, KalimantanSelatan101, KalimantanSelatan102, KalimantanTengah107, Sea133},
	KalimantanSelatan101: AreaIDS{KalimantanSelatan100, KalimantanSelatan102, Sea133},
	KalimantanSelatan102: AreaIDS{KalimantanSelatan100, KalimantanSelatan101, KalimantanTengah107, Sea133},

	KalimantanTengah103: AreaIDS{KalimantanTimur97, KalimantanTimur98, KalimantanTengah104, KalimantanTengah105,
		KalimantanTengah106, KalimantanTengah107, KalimantanBarat108},
	KalimantanTengah104: AreaIDS{KalimantanTengah103, KalimantanTengah105, KalimantanTengah107, Sea133},
	KalimantanTengah105: AreaIDS{KalimantanTengah103, KalimantanTengah104, KalimantanTengah106, Sea130, Sea133},
	KalimantanTengah106: AreaIDS{KalimantanTengah103, KalimantanTengah105, KalimantanBarat109, Sea130},
	KalimantanTengah107: AreaIDS{KalimantanTimur98, KalimantanTimur99, KalimantanSelatan100, KalimantanSelatan102,
		KalimantanTengah103, KalimantanTengah104, Sea133},

	KalimantanBarat108: AreaIDS{KalimantanTimur97, KalimantanTengah103, KalimantanBarat109, Sarawak111},
	KalimantanBarat109: AreaIDS{KalimantanTengah106, KalimantanBarat108, KalimantanBarat110, Sarawak111, Sea130},
	KalimantanBarat110: AreaIDS{KalimantanBarat109, Sarawak111, Sea126, Sea130},

	Sarawak111: AreaIDS{KalimantanTimur97, KalimantanBarat108, KalimantanBarat109, KalimantanBarat110,
		Sarawak112, Sea126},
	Sarawak112: AreaIDS{KalimantanTimur95, KalimantanTimur97, Sarawak111, Sarawak113, Sarawak114, Sea126},
	Sarawak113: AreaIDS{Sarawak112, Sarawak114, Sea126},
	Sarawak114: AreaIDS{KalimantanTimur95, Sarawak112, Sarawak113, Sea125},

	Sea115: AreaIDS{Aceh0, Aceh1, Aceh2, Aceh3, SumateraUtara6, Sea116, Sea126, Sea127},
	Sea116: AreaIDS{SumateraUtara6, SumateraUtara7, SumateraBarat13, SumateraBarat14, SumateraBarat15,
		SumateraBarat16, Bengkulu20, Bengkulu21, Bengkulu22, Sea115, Sea117},
	Sea117: AreaIDS{Bengkulu22, Lampung30, Lampung33, JawaBarat34, JawaBarat38, JawaBarat39, JawaBarat40,
		Sea116, Sea118, Sea131},
	Sea118: AreaIDS{JawaBarat38, JawaTengah41, JawaTengah43, JawaTimur46, JawaTimur47, JawaTimur48, JawaTimur49,
		Bali50, Sea117, Sea119},
	Sea119: AreaIDS{Bali50, Bali51, NusaTenggaraBarat52, NusaTenggaraBarat53, NusaTenggaraTimur54,
		Sea118, Sea120, Sea132, Sea133, Sea134},
	Sea120: AreaIDS{NusaTenggaraTimur55, NusaTenggaraTimur56, NusaTenggaraTimur57, NusaTenggaraTimur58,
		NusaTenggaraTimur59, Sea119, Sea121, Sea134},
	Sea121: AreaIDS{Maluku60, Maluku63, Maluku64, Maluku65, Maluku66, Sea120, Sea122, Sea123, Sea124, Sea134},
	Sea122: AreaIDS{Maluku61, Maluku62, Papua69, Papua70, Papua71, Sea121, Sea123},
	Sea123: AreaIDS{Papua71, Papua72, Papua73, Papua74, Papua75, Sea121, Sea122, Sea124},
	Sea124: AreaIDS{Maluku67, Maluku68, Halmahera76, Halmahera77, Halmahera78, Halmahera79, Halmahera80,
		Halmahera81, Sea121, Sea123, Sea125, Sea134, Sea135},
	Sea125: AreaIDS{SulawesiSelatan86, SulawesiSelatan87, SulawesiTengah88, SulawesiTengah89, SulawesiUtara93,
		KalimantanTimur95, KalimantanTimur96, KalimantanTimur98, KalimantanTimur99, Sarawak114, Sea124,
		Sea126, Sea133, Sea134, Sea135},
	Sea126: AreaIDS{KalimantanBarat110, Sarawak111, Sarawak112, Sarawak113, Sea115, Sea125, Sea127,
		Sea128, Sea130},
	Sea127: AreaIDS{Aceh1, SumateraUtara4, SumateraUtara5, Riau8, Sea115, Sea126, Sea128},
	Sea128: AreaIDS{Riau8, Riau9, Riau10, Jambi18, Jambi19, Sea126, Sea127, Sea129, Sea130},
	Sea129: AreaIDS{Jambi19, SumateraSelatan24, SumateraSelatan25, SumateraSelatan27, Lampung32, Sea128,
		Sea130, Sea131},
	Sea130: AreaIDS{SumateraSelatan26, KalimantanTengah105, KalimantanTengah106, KalimantanBarat109,
		KalimantanBarat110, Sea126, Sea128, Sea129, Sea131, Sea132, Sea133},
	Sea131: AreaIDS{Lampung32, Lampung33, JawaBarat34, JawaBarat35, JawaBarat36, JawaBarat37, JawaTengah41,
		Sea117, Sea129, Sea130, Sea132},
	Sea132: AreaIDS{JawaTengah41, JawaTengah42, JawaTimur44, JawaTimur45, JawaTimur46, JawaTimur47, JawaTimur48,
		Bali50, Sea119, Sea130, Sea131, Sea133},
	Sea133: AreaIDS{KalimantanTimur99, KalimantanSelatan100, KalimantanSelatan101, KalimantanSelatan102,
		KalimantanTengah104, KalimantanTengah105, KalimantanTengah107,
		Sea119, Sea125, Sea130, Sea132, Sea134},
	Sea134: AreaIDS{SulawesiTenggara82, SulawesiTenggara83, SulawesiTenggara84, SulawesiSelatan85,
		SulawesiSelatan86, Sea119, Sea120, Sea121, Sea124, Sea125, Sea133, Sea135},
	Sea135: AreaIDS{SulawesiTenggara84, SulawesiTengah88, SulawesiTengah89, SulawesiTengah90, SulawesiTengah91,
		SulawesiTengah92, SulawesiUtara93, SulawesiUtara94, Sea124, Sea125, Sea134},
}

func (as Areas) include(area *Area) bool {
	for _, a := range as {
		if a.ID == area.ID {
			return true
		}
	}
	return false
}

func (as Areas) exclude(areas Areas) Areas {
	var filtered Areas
	for _, a := range as {
		if !areas.include(a) {
			filtered = append(filtered, a)
		}
	}
	return filtered
}

//var baliAreaIDS = AreaIDS{Bali50, Bali51}
//
//var jawaBaratAreaIDS = AreaIDS{JawaBarat34, JawaBarat35, JawaBarat36, JawaBarat37, JawaBarat38, JawaBarat39,
//	JawaBarat40}
//
//var jawaTengahAreaIDS = AreaIDS{JawaTengah41, JawaTengah42, JawaTengah43}
//
//var jawaTimurAreaIDS = AreaIDS{JawaTimur44, JawaTimur45, JawaTimur46, JawaTimur47, JawaTimur48, JawaTimur49}
//
//var sulawesiSelatanAreaIDS = AreaIDS{SulawesiSelatan85, SulawesiSelatan86, SulawesiSelatan87}
//
//var sulawesiUtaraAreaIDS = AreaIDS{SulawesiUtara93, SulawesiUtara94}
//
//var sumateraSelatanAreaIDS = AreaIDS{SumateraSelatan23, SumateraSelatan24, SumateraSelatan25, SumateraSelatan26,
//	SumateraSelatan27, SumateraSelatan28, SumateraSelatan29}

func (c *Company) ExpansionAreas() Areas {
	if c == nil {
		return nil
	}
	if c.IsProductionCompany() {
		return c.Zones.Areas().expandAreasFor(c)
	}
	var expansionAreas Areas
	for _, area := range c.Zones.Areas() {
		for _, a := range area.AdjacentSeaAreas() {
			if !expansionAreas.include(a) {
				expansionAreas = append(expansionAreas, a)
			}
		}
	}
	return expansionAreas
}

//g.RequiredExpansions = min(cp.Technologies[ExpansionsTech], len(company.ExpansionAreas()))

func (as Areas) expandAreasFor(c *Company) Areas {
	var expansionAreas Areas
	for _, area := range as {
		for _, a := range area.AdjacentLandAreas() {
			if !a.hasProducer() && !a.hasCity() && !a.adjacentAreaHasCompetingCompanyFor(c) &&
				!expansionAreas.include(a) && !as.include(a) {
				expansionAreas = append(expansionAreas, a)
			}
		}
	}
	return expansionAreas
}

func (c *Company) requiredExpansions() int {
	result := 0
	maxExpansions := c.g.CurrentPlayer().Technologies[ExpansionsTech]
	areas := c.Zones.Areas()
	startSize := len(areas)
	expandAreas := areas.expandAreasFor(c)
	areas = append(areas, expandAreas...)
	//	c.g.debugf("before for areas: %s", areas.ids())
	for {
		l := len(areas) - startSize
		//		c.g.debugf("l: %d result: %d maxExpansions: %d", l, result, maxExpansions)
		if l >= maxExpansions {
			return maxExpansions
		}
		if l == result {
			return l
		}
		result, areas = l, append(areas, areas.expandAreasFor(c)...)
		//		c.g.debugf("expansionAreas: %s", areas.ids())
	}
	return result
}

func (a *Area) hasProducer() bool {
	return a.Producer != nil
}

func (g *Game) freeShippingExpansionAreas() Areas {
	var expansionAreas Areas
	company := g.SelectedCompany()
	if company == nil {
		return nil
	}
	areas := company.Zones.Areas()
	expansionAreas = append(expansionAreas, areas...)
	for _, area := range areas {
		for _, a := range area.AdjacentAreas() {
			if a.IsSea() && !expansionAreas.include(a) {
				expansionAreas = append(expansionAreas, a)
			}
		}
	}
	return expansionAreas
}

func (g *Game) adminArea(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	err := g.validateAdminAction(cu)
	if err != nil {
		return "indonesia/flash_notice", game.None, err
	}

	obj := struct {
		RemoveProducer bool   `form:"RemoveProducer"`
		AddProducerFor string `form:"AddProducerFor"`
	}{}

	err = c.ShouldBind(&obj)
	if err != nil {
		return "indonesia/flash_notice", game.None, err
	}
	log.Debugf("obj: %#v", obj)

	area := g.SelectedArea()
	log.Debugf("area: %#v", area)

	if obj.RemoveProducer {
		company := area.Producer.Company()
		area.Producer = nil
		if company != nil {
			company.RemoveArea(area)
		}
	}

	if obj.AddProducerFor != "none" && obj.AddProducerFor != "" {
		var company *Company
		splits := strings.Split(obj.AddProducerFor, "-")
		if len(splits) == 2 {
			p := g.PlayerBySID(splits[0])
			slot := -1
			v, err := strconv.Atoi(splits[1])
			if err != nil {
				return "indonesia/flash_notice", game.None, err
			}
			slot = v
			company = p.Slots[slot-1].Company
			area.AddProducer(company)
			company.AddArea(area)
		}
	}

	return "", game.Save, err
}

var imageMapArea = map[AreaID]string{
	Aceh0:               "91,192,85,184,79,181,67,168,67,162,58,157,53,144,50,142,50,137,46,134,46,115,50,110,72,109,76,114,82,117,93,129,107,129,110,132,123,132,123,138,120,151,115,164,105,179,97,187",
	Aceh1:               "196,242,191,231,177,211,156,190,130,173,114,165,120,152,123,139,123,132,130,128,144,129,148,133,148,136,161,133,164,128,175,141,184,146,193,162,193,170,198,170,202,174,207,174,206,215",
	Aceh2:               "112,215,92,191,105,180,114,165,135,175,156,190,176,210,191,231,196,243,184,267,185,277,182,281,188,300,186,304,174,293,174,279,170,274,171,266,169,266,165,261,162,261,155,253,155,250,147,246,141,231,136,227,131,215",
	Aceh3:               "113,295,105,295,95,286,93,285,89,285,84,281,75,281,74,271,71,268,73,267,73,264,82,264,87,271,92,271,93,278,100,278,109,286,113,286",
	SumateraUtara4:      "296,277,265,311,194,247,206,215,206,197,217,197,217,201,219,201,221,204,227,205,231,214,244,222,248,222,250,225,254,225,265,237,269,237,272,244,286,250,290,259,296,263,300,271",
	SumateraUtara5:      "389,333,377,345,367,371,359,376,350,388,265,310,296,277,303,275,304,278,310,275,310,282,326,306,342,314,338,302,350,298,357,298,357,304,366,315,373,309,378,314,378,310,387,309,391,315,391,323,384,330,377,330,378,333",
	SumateraUtara6:      "278,422,271,419,265,419,262,416,258,420,254,419,254,396,243,392,243,375,236,357,236,351,230,342,230,336,232,336,232,332,227,331,223,322,213,314,205,314,205,310,188,301,181,280,185,277,184,266,194,247,350,389,347,393,318,404,305,405",
	SumateraUtara7:      "216,475,207,471,150,382,137,353,142,336,150,333,159,333,198,367,222,389,235,409,239,417,239,425,236,432,231,451,227,462,223,469",
	Riau8:               "389,333,400,344,405,344,403,337,415,336,422,341,424,340,428,344,429,360,427,366,427,369,437,368,437,364,446,360,463,376,463,380,454,378,459,383,459,387,452,384,443,384,444,386,449,386,460,399,460,404,446,413,420,421,401,424,376,421,376,397,372,379,367,371,377,345",
	Riau9:               "493,400,479,396,470,390,468,375,470,356,482,342,498,334,512,334,533,342,556,358,567,379,573,409,571,438,564,456,548,471,535,478,524,478,515,470,510,457,507,440,507,423,504,412",
	Riau10:              "498,431,484,430,484,434,482,434,479,440,495,440,495,446,491,451,483,451,484,453,480,459,478,458,478,464,475,468,485,475,487,482,446,496,443,498,428,498,422,496,421,421,454,411,464,401,463,398,466,390,470,390,471,397,478,398,485,403,487,409,495,413",
	Riau11:              "421,421,423,497,382,488,372,479,368,462,372,447,376,421,401,423",
	Riau12:              "367,371,373,380,376,397,376,421,372,446,326,433,325,423,319,419,317,403,347,393,360,375",
	SumateraBarat13:     "320,492,317,488,314,476,303,465,303,460,300,453,286,446,286,442,282,436,285,433,279,422,306,405,318,404,319,418,325,424,326,433,372,446,369,462,370,470,352,474,336,482",
	SumateraBarat14:     "373,604,369,597,365,585,356,581,344,561,340,551,340,537,330,524,330,516,327,512,327,504,324,504,320,500,320,496,317,496,319,492,351,474,370,470,372,479,382,488,416,496,428,498,428,513,421,548,413,564",
	SumateraBarat15:     "322,625,309,617,292,595,281,580,267,554,265,542,275,533,282,533,285,538,310,564,319,576,326,587,332,605,330,618,327,623,327,625",
	SumateraBarat16:     "259,532,251,532,246,526,242,526,238,523,237,518,222,499,226,494,226,483,234,483,237,480,241,482,241,486,245,490,246,500,248,506,255,506,255,514,258,517,259,523,255,523",
	Jambi17:             "444,588,430,577,408,570,413,565,421,549,428,515,428,498,443,498,446,496,448,510,453,530,463,547,472,558,478,564,493,572,472,580",
	Jambi18:             "533,508,525,511,515,521,506,536,502,550,501,560,501,569,494,571,478,564,464,548,453,531,446,496,487,483,497,483,498,486,503,483,511,487,515,487,518,489,525,487,529,487,529,498,532,502",
	Jambi19:             "553,555,549,561,501,569,502,549,506,535,515,520,526,509,533,508,533,520,536,525,536,529,536,533,543,537,542,540,542,543,547,541,550,541,553,545",
	Bengkulu20:          "413,635,399,628,373,603,407,570,431,577,436,583",
	Bengkulu21:          "484,647,472,650,451,661,441,673,416,652,416,647,412,643,413,635,436,583,471,608",
	Bengkulu22:          "518,696,496,698,486,709,482,708,479,705,475,705,475,701,473,701,464,692,447,685,441,673,451,660,472,650,484,647,488,659",
	SumateraSelatan23:   "497,596,494,610,485,620,477,624,471,608,443,588,474,579,494,571,497,588",
	SumateraSelatan24:   "587,623,535,598,497,596,497,587,494,570,501,568,549,561,553,556,555,559,559,562,568,563,570,559,579,563,592,563,592,572,595,575,595,580,601,580,602,591,605,596,596,609",
	SumateraSelatan25:   "637,601,619,594,613,594,607,582,607,575,604,565,604,551,590,550,584,546,575,550,566,547,566,537,580,529,580,527,577,527,583,517,591,517,597,521,597,514,604,514,604,509,613,517,617,517,617,526,620,530,620,543,627,559,633,567,639,567,658,575,658,578,650,581,644,591,646,595,652,595,656,593,660,596,660,603,654,599,650,606,638,606",
	SumateraSelatan26:   "697,614,696,606,694,601,693,596,698,594,694,589,697,582,698,575,703,570,713,575,717,574,720,577,719,578,722,582,727,582,734,594,727,600,727,605,720,607,716,613,714,609,710,605,709,608",
	SumateraSelatan27:   "607,695,563,692,563,677,569,658,577,638,605,596,610,596,616,600,616,613,612,617,602,639,609,645,609,651,606,657,605,681,609,688",
	SumateraSelatan28:   "563,691,555,691,548,693,546,677,538,660,516,633,493,611,497,596,534,598,587,623,577,639,568,660,563,677",
	SumateraSelatan29:   "548,693,518,696,488,659,477,624,485,620,494,611,516,633,538,662,546,678",
	Lampung30:           "512,742,509,739,507,732,499,728,502,725,493,713,485,713,485,710,496,698,547,694,545,707,539,720,532,729,521,738",
	Lampung31:           "584,736,535,726,545,708,547,693,574,692,581,707,584,719",
	Lampung32:           "605,740,585,736,585,718,580,705,574,692,606,695,604,701,609,708,605,716",
	Lampung33:           "545,767,534,767,534,759,531,759,524,752,520,752,520,744,513,742,522,737,535,725,604,740,605,747,602,752,601,758,597,763,596,759,590,759,589,755,586,755,574,742,568,745,569,750,567,755,570,759,563,759,557,751,541,744,536,744,536,747,546,758",
	JawaBarat34:         "626,825,623,825,619,820,614,817,604,817,601,820,597,816,587,821,582,821,580,818,572,818,572,812,580,807,580,812,585,814,596,796,600,796,602,787,603,772,611,769,610,766,614,766,617,769,622,769,625,766,630,770,645,770",
	JawaBarat35:         "673,794,663,798,653,799,636,798,645,770,649,769,651,773,662,773,662,767,667,765,670,765,674,769,677,765,681,765,690,777,684,787",
	JawaBarat36:         "702,831,688,825,680,820,669,809,663,797,673,794,684,788,691,777,693,776,697,780,705,781,708,777,712,785,716,785,719,788,724,788,729,785,734,785,739,793",
	JawaBarat37:         "775,820,764,848,722,836,702,832,739,794,747,796,750,807,750,816,766,817,772,813",
	JawaBarat38:         "743,867,739,875,724,870,713,872,711,867,706,866,722,836,744,842,764,848,757,868",
	JawaBarat39:         "722,836,705,866,700,864,694,856,683,856,676,853,642,852,680,820,689,826,702,832",
	JawaBarat40:         "680,820,642,853,637,848,634,848,631,845,630,838,641,832,638,824,627,825,636,798,654,799,663,798,670,809",
	JawaTengah41:        "841,854,841,870,839,884,835,890,828,886,822,887,794,872,793,867,758,868,775,821,784,819,795,814,795,820,805,825,814,818,818,818,829,822,835,814,838,814",
	JawaTengah42:        "888,829,888,845,885,858,842,855,838,815,835,815,844,795,854,791,864,795,864,803,868,807,879,807,883,803,897,810",
	JawaTengah43:        "885,868,886,904,882,905,875,898,862,898,852,893,839,894,834,890,839,885,841,868,841,855,884,857",
	JawaTimur44:         "934,855,917,854,888,846,888,828,897,810,911,810,919,820,943,818,945,828,942,832,948,841,950,839,956,844,952,853,950,853,950,855",
	JawaTimur45:         "992,843,972,840,955,836,949,836,946,831,953,818,1012,818,1023,823,1015,828,1009,829,1009,837,998,833,992,842",
	JawaTimur46:         "1008,913,1007,912,1002,914,997,910,1008,864,1013,865,1020,861,1026,865,1030,865,1040,874,1040,888,1037,893,1037,904,1034,909,1039,912,1039,917,1047,922,1047,927,1039,930,1036,924,1036,921,1025,920,1023,916,1020,922,1015,916",
	JawaTimur47:         "1003,886,997,910,990,906,986,907,986,914,981,915,976,907,980,906,980,902,965,901,960,909,955,897,955,884,957,861,976,870,992,864,1008,865",
	JawaTimur48:         "954,883,955,897,958,909,945,909,929,902,924,905,922,902,916,902,916,861,917,854,949,855,949,857,956,861",
	JawaTimur49:         "917,854,916,863,916,902,899,902,897,904,893,904,893,900,888,901,888,904,886,904,884,859,889,846,901,850",
	Bali50:              "1077,883,1077,922,1074,916,1070,916,1064,909,1049,909,1042,902,1041,890,1054,888,1057,892,1071,892",
	Bali51:              "1105,900,1106,907,1099,914,1096,913,1086,918,1086,923,1079,924,1082,928,1082,932,1073,932,1077,927,1077,884",
	NusaTenggaraBarat52: "1144,936,1139,940,1136,936,1130,936,1126,932,1124,932,1124,935,1121,936,1118,932,1111,932,1111,927,1119,926,1125,917,1125,908,1140,891,1150,900,1157,900,1161,902,1159,911,1159,915,1155,921,1151,926,1151,930,1156,936",
	NusaTenggaraBarat53: "1183,948,1175,948,1164,945,1160,935,1165,929,1164,911,1177,910,1185,903,1196,911,1200,907,1207,907,1214,920,1223,925,1230,920,1234,919,1240,922,1235,912,1227,913,1215,905,1215,901,1209,897,1213,892,1221,892,1224,888,1227,888,1232,891,1235,891,1242,903,1247,903,1247,900,1253,896,1262,899,1262,913,1265,911,1265,903,1272,895,1275,898,1280,898,1280,904,1284,910,1281,915,1282,919,1285,919,1289,915,1290,922,1287,929,1275,929,1278,932,1266,932,1263,928,1253,932,1248,932,1248,925,1249,919,1242,927,1235,932,1232,931,1228,936,1210,935,1203,943,1189,944,1185,941",
	NusaTenggaraTimur54: "1313,986,1289,986,1282,982,1275,974,1278,971,1278,966,1282,966,1287,961,1314,961,1320,964,1327,958,1334,964,1345,964,1346,975,1359,975,1367,988,1367,993,1372,992,1377,996,1373,1010,1367,1014,1362,1012,1359,1017,1350,1012,1344,1013,1334,1006,1330,998,1321,995",
	NusaTenggaraTimur55: "1429,928,1426,937,1421,932,1410,932,1406,927,1405,936,1380,936,1374,931,1373,937,1369,932,1361,928,1354,928,1349,931,1343,931,1338,929,1333,932,1326,932,1320,923,1315,928,1310,928,1310,926,1312,923,1309,918,1304,919,1303,926,1299,924,1299,914,1302,908,1304,910,1310,910,1311,917,1315,917,1315,919,1320,919,1320,915,1324,913,1323,908,1331,908,1341,899,1353,899,1359,896,1362,899,1371,899,1387,908,1398,910,1407,914,1414,914,1414,911,1432,910,1438,908,1438,911,1446,915,1449,918,1462,919,1462,911,1469,909,1473,903,1481,899,1482,897,1476,897,1476,894,1480,884,1491,891,1491,901,1497,895,1509,895,1512,900,1514,894,1519,899,1524,899,1531,893,1539,897,1530,901,1525,910,1519,908,1522,916,1515,916,1510,921,1508,917,1500,916,1507,911,1507,905,1493,911,1488,916,1481,916,1477,921,1466,921,1455,925,1443,925,1440,928",
	NusaTenggaraTimur56: "1594,901,1586,915,1573,919,1555,922,1538,921,1532,913,1535,903,1551,891,1568,887,1583,886,1593,890,1595,894",
	NusaTenggaraTimur57: "1622,886,1624,879,1624,875,1632,866,1650,867,1653,862,1664,858,1671,865,1675,865,1680,871,1665,872,1661,877,1658,883,1652,884,1649,879,1631,879",
	NusaTenggaraTimur58: "1657,941,1643,941,1633,947,1633,952,1619,952,1617,957,1609,956,1598,961,1586,940,1588,919,1598,917,1620,917,1624,913,1647,913,1653,909,1663,911,1669,909,1675,909,1684,901,1691,905,1698,905,1703,909,1690,925,1686,925,1677,930,1663,929,1662,935",
	NusaTenggaraTimur59: "1585,939,1597,959,1593,968,1587,981,1581,989,1581,993,1570,1000,1570,1004,1565,1008,1550,1009,1545,1013,1534,1021,1514,1021,1520,1010,1524,1008,1526,1005,1525,1001,1521,1001,1520,983,1523,980,1528,969,1535,968,1550,954,1556,954,1562,950,1572,949,1579,942",
	Maluku60:            "1791,917,1760,923,1737,920,1722,910,1712,896,1708,880,1717,857,1738,837,1758,826,1777,821,1802,825,1813,829,1829,836,1845,846,1855,856,1860,867,1861,873,1859,883,1852,890,1835,901,1818,909",
	Maluku61:            "1963,849,1930,884,1907,903,1892,914,1883,919,1873,919,1873,882,1878,865,1906,833,1925,820,1941,816,1957,823,1963,831,1966,842",
	Maluku62:            "2055,833,1998,800,1975,777,1968,759,1970,742,1981,725,2005,713,2034,708,2057,708,2082,713,2097,724,2105,735,2110,752,2111,777,2107,795,2100,811,2092,821,2080,829,2071,832",
	Maluku63:            "1886,629,1889,632,1889,648,1881,648,1871,640,1868,641,1863,637,1845,630,1867,602,1869,605,1876,608,1876,616,1879,624,1886,623",
	Maluku64:            "1844,630,1835,621,1825,622,1820,618,1820,625,1801,626,1793,621,1794,595,1797,596,1809,593,1811,589,1819,589,1829,593,1835,593,1844,601,1856,601,1864,597,1866,602",
	Maluku65:            "1793,622,1787,621,1787,613,1778,613,1775,619,1775,625,1768,625,1765,629,1758,626,1752,613,1744,613,1744,620,1738,628,1737,633,1731,628,1734,624,1734,620,1728,613,1746,598,1746,593,1785,592,1790,589,1796,595,1793,595",
	Maluku66:            "1698,637,1691,641,1688,640,1684,645,1674,645,1671,649,1656,645,1645,636,1641,637,1638,627,1630,622,1630,609,1638,605,1641,608,1645,606,1680,604,1682,608,1691,613,1687,617,1687,620,1698,620",
	Maluku67:            "1641,568,1629,568,1629,558,1621,546,1621,539,1606,538,1602,535,1602,529,1653,529,1653,531,1647,534,1641,534,1631,539,1630,551",
	Maluku68:            "1581,538,1578,542,1565,542,1562,533,1562,526,1566,521,1577,521,1587,529,1587,525,1595,525,1599,535,1588,533",
	Papua69:             "2358,655,2358,892,2343,893,2338,898,2329,896,2327,890,2321,892,2314,901,2309,899,2309,909,2300,911,2298,906,2289,906,2289,902,2280,911,2274,906,2250,906,2240,911,2248,893,2247,885,2256,871,2260,870,2262,861,2279,849,2291,849,2291,845,2302,846,2298,839,2309,838,2290,824,2310,819,2304,812,2296,812,2296,809,2298,809,2284,792,2282,776,2272,764,2276,751,2272,751,2269,755,2264,755,2264,740,2254,736,2247,725,2235,721,2225,710,2241,653,2309,656",
	Papua70:             "2233,680,2225,709,2221,714,2217,714,2216,708,2197,711,2185,702,2176,697,2163,698,2161,690,2155,686,2141,687,2125,681,2113,684,2099,670,2088,670,2087,664,2085,664,2088,654,2080,659,2093,632,2168,647,2241,653",
	Papua71:             "2103,612,2080,660,2074,653,2070,653,2070,661,2060,651,2063,648,2062,644,2057,648,2046,641,2044,644,2034,630,2022,642,2023,647,2016,660,2022,666,2026,666,2034,674,2027,671,2022,671,2015,663,2003,660,2000,664,1994,664,1987,652,1988,645,1985,639,1992,639,1991,626,1988,621,1982,621,1977,612,1978,608,1972,602,1964,602,1962,597,1944,597,1950,588,1952,588,1957,580,1968,580,1979,585,1985,589,1992,587,1992,581,2001,571,2010,565,2017,565,2017,573,2033,573,2036,579,2037,572,2043,578,2043,571,2040,570,2043,565,2049,565,2047,560,2058,555,2072,579,2078,589,2078,582,2075,576,2080,572,2086,572,2087,581,2091,595,2096,592,2096,603",
	Papua72:             "2059,556,2047,560,2043,560,2043,553,2039,556,2035,556,2032,561,2027,556,2000,555,1997,560,1990,557,1982,564,1977,555,1973,555,1969,560,1964,560,1957,553,1954,553,1943,536,1947,527,1941,526,1944,519,1935,519,1925,513,1921,516,1916,510,1911,516,1907,516,1903,512,1893,512,1896,506,1895,503,1889,503,1880,495,1880,490,1878,487,1879,483,1884,483,1884,480,1867,481,1867,475,1872,470,1880,469,1882,473,1886,471,1889,471,1888,479,1895,479,1900,482,1900,492,1898,498,1905,498,1908,492,1908,474,1918,470,1926,470,1942,467,1948,455,1954,454,1962,447,1984,446,2001,452,2005,456,2011,455,2023,467,2026,466,2032,471,2057,471,2064,476,2057,484,2060,487,2063,495,2071,502,2067,510,2068,517,2063,523,2061,523",
	Papua73:             "1845,554,1826,545,1817,526,1818,505,1830,473,1846,444,1864,425,1890,416,1907,421,1933,442,1934,448,1929,453,1917,457,1901,458,1879,461,1862,471,1855,481,1854,490,1855,496,1860,506,1868,518,1877,534,1878,542,1878,552,1873,555,1864,557",
	Papua74:             "2120,530,2106,521,2098,512,2089,497,2088,485,2095,470,2106,462,2123,456,2140,455,2155,459,2174,472,2188,483,2199,497,2210,517,2212,524,2209,531,2206,536,2195,538,2176,545,2159,546,2138,539",
	Papua75:             "2357,655,2306,656,2241,653,2169,647,2093,632,2103,612,2103,619,2111,618,2119,624,2125,620,2129,623,2134,622,2134,614,2154,597,2154,587,2157,586,2158,580,2165,580,2171,576,2172,567,2176,563,2175,558,2180,553,2188,553,2194,556,2208,548,2214,549,2220,547,2221,542,2214,537,2222,524,2230,524,2260,508,2267,521,2272,520,2301,529,2313,539,2322,539,2343,560,2353,559,2358,562",
	Halmahera76:         "1727,527,1723,527,1716,531,1705,523,1704,513,1717,506,1720,501,1726,508,1735,510,1740,516,1745,516,1748,521,1748,521,1748,527,1748,527,1748,527,1748,527,1748,527,1737,527",
	Halmahera77:         "1735,472,1727,482,1722,476,1722,473,1717,475,1713,477,1708,472,1708,464,1705,464,1701,456,1692,456,1692,445,1698,440,1701,447,1706,445,1708,449,1712,445,1720,452,1720,466,1727,466",
	Halmahera78:         "1764,481,1752,481,1750,476,1739,469,1732,449,1723,445,1722,411,1719,412,1715,407,1715,399,1712,396,1712,386,1720,382,1712,376,1720,375,1723,379,1732,379,1733,376,1755,382,1755,384,1752,384,1752,387,1757,389,1763,390,1768,393,1772,393,1776,397,1776,405,1782,408,1786,420,1781,415,1767,411,1764,408,1754,409,1746,403,1739,403,1734,405,1734,431,1737,435,1738,446,1750,466", ^Halmahera79: "1721,375,1711,376,1711,369,1705,374,1704,358,1712,343,1712,332,1718,326,1718,320,1734,304,1744,304,1739,313,1739,318,1731,321,1731,324,1735,324,1740,329,1741,345,1734,360,1734,364,1729,364,1726,370,1721,370",
	Halmahera79:          "1721,375,1711,376,1711,369,1705,374,1704,358,1712,343,1712,332,1718,326,1718,320,1734,304,1744,304,1739,313,1739,318,1731,321,1731,324,1735,324,1740,329,1741,345,1734,360,1734,364,1729,364,1726,370,1721,370",
	Halmahera80:          "1762,315,1753,314,1749,303,1749,294,1760,282,1768,282,1772,279,1773,283,1776,286,1776,291,1772,293,1772,300,1767,314",
	Halmahera81:          "1755,382,1732,376,1737,372,1737,367,1746,364,1746,362,1743,361,1746,348,1750,344,1753,344,1757,340,1767,340,1771,337,1776,339,1777,345,1780,346,1781,353,1776,360,1776,368,1762,377,1755,376",
	SulawesiTenggara82:   "1506,740,1496,753,1485,759,1473,762,1456,762,1441,758,1426,752,1417,745,1410,733,1413,722,1425,712,1437,710,1449,707,1450,697,1458,690,1475,685,1483,681,1488,674,1485,664,1483,659,1485,650,1496,647,1510,648,1517,657,1519,669,1519,689,1518,707,1515,721",
	SulawesiTenggara83:   "1472,661,1484,664,1484,680,1451,680,1442,686,1440,696,1445,704,1419,705,1413,702,1410,692,1409,667,1415,665,1415,660,1409,655,1405,655,1413,642,1438,632,1452,632,1457,642,1458,642,1467,649,1472,647,1473,651,1470,654",
	SulawesiTenggara84:   "1450,631,1436,633,1422,639,1413,644,1405,654,1398,649,1390,647,1373,626,1382,620,1381,613,1385,611,1384,592,1399,585,1415,560,1431,564,1439,577,1439,582,1449,592,1452,592,1453,599,1455,605,1459,605,1459,610,1453,614,1453,617,1459,622,1456,626,1452,624",
	SulawesiSelatan85:    "1352,689,1356,693,1352,705,1349,705,1349,714,1345,721,1357,742,1357,748,1350,749,1350,742,1337,748,1327,748,1320,754,1311,750,1305,752,1305,746,1299,746,1294,736,1294,729,1299,724,1298,719,1302,719,1306,707,1304,702,1314,696,1335,690,1352,688",
	SulawesiSelatan86:    "1399,585,1386,592,1379,588,1384,582,1376,581,1375,576,1362,586,1342,596,1345,599,1345,604,1352,610,1352,688,1336,690,1319,695,1313,697,1304,703,1305,696,1311,680,1309,672,1309,661,1311,655,1302,644,1302,626,1343,549,1416,561",
	SulawesiSelatan87:    "1303,626,1343,549,1293,541,1292,550,1286,558,1287,566,1283,574,1268,580,1268,591,1271,596,1271,604,1265,604,1265,610,1270,616,1270,620,1274,626,1274,630,1280,627,1295,629,1298,624",
	SulawesiTengah88:     "1396,403,1394,407,1387,404,1385,407,1379,402,1373,402,1372,399,1356,399,1341,418,1341,422,1336,434,1327,434,1327,424,1323,419,1327,414,1327,401,1334,397,1332,392,1336,385,1339,387,1347,383,1347,368,1352,369,1353,379,1362,382,1371,370,1374,370,1374,362,1377,358,1375,353,1383,351",
	SulawesiTengah89:     "1372,553,1298,541,1302,533,1298,525,1298,493,1305,490,1305,483,1312,471,1317,471,1320,467,1324,473,1324,466,1319,459,1318,448,1323,439,1320,438,1315,435,1315,428,1318,424,1321,424,1326,433,1336,436,1339,446,1339,464,1347,478,1353,477,1363,484,1363,484,1360,489,1358,495,1356,510,1356,521,1359,533,1364,543,1369,548",
	SulawesiTengah90:     "1452,470,1463,509,1455,515,1452,523,1435,523,1430,534,1421,537,1413,534,1409,534,1406,531,1406,538,1420,552,1426,552,1431,563,1373,554,1370,549,1363,543,1358,531,1355,521,1355,508,1358,491,1362,487,1367,492,1367,496,1371,497,1371,505,1375,500,1384,506,1394,504,1397,492,1403,492,1413,474,1421,474,1422,477,1433,477,1435,481,1444,478,1447,470",
	SulawesiTengah91:     "1504,460,1508,463,1516,463,1515,467,1518,474,1514,482,1513,489,1505,489,1498,484,1497,478,1484,478,1477,491,1475,500,1467,501,1467,507,1464,508,1462,509,1452,470,1473,470,1479,473,1485,470,1486,468,1478,468",
	SulawesiTengah92:     "1446,429,1460,428,1467,434,1463,439,1456,445,1446,445,1439,443,1438,435",
	SulawesiUtara93:      "1541,411,1503,412,1498,408,1498,402,1495,400,1442,399,1438,402,1426,402,1422,400,1411,399,1407,403,1404,400,1401,403,1396,403,1382,351,1392,351,1400,360,1406,354,1411,354,1415,360,1415,367,1425,368,1434,363,1437,366,1448,367,1454,372,1473,371,1483,382,1491,375,1501,375,1507,370,1511,375,1515,374,1521,377,1539,379",
	SulawesiUtara94:      "1539,378,1558,371,1562,362,1567,358,1575,359,1575,357,1572,356,1572,353,1576,346,1579,347,1579,329,1584,333,1592,333,1592,337,1596,340,1589,345,1589,351,1581,372,1579,390,1577,390,1577,396,1569,399,1569,404,1563,404,1560,408,1545,406,1544,411,1541,411",
	KalimantanTimur95:    "1160,210,1189,221,1189,224,1198,225,1201,229,1202,233,1196,240,1196,244,1193,245,1194,249,1201,248,1204,251,1210,252,1208,258,1213,256,1209,264,1215,264,1212,270,1215,271,1219,279,1230,286,1230,293,1234,294,1234,302,1227,306,1222,314,1201,313,1166,308,1131,299,1073,280,1091,230,1122,211",
	KalimantanTimur96:    "1254,348,1260,348,1264,357,1273,360,1272,364,1276,367,1283,367,1283,372,1271,384,1265,380,1258,385,1254,381,1251,384,1237,380,1231,374,1231,385,1226,381,1215,389,1190,390,1160,389,1105,380,1131,298,1166,307,1221,315,1226,324,1234,324,1235,332",
	KalimantanTimur97:    "1131,298,1105,380,1000,353,968,350,940,366,931,331,940,316,951,306,973,298,1053,291,1074,280",
	KalimantanTimur98:    "1203,435,1103,446,1096,426,1031,380,1006,383,1006,367,1001,352,1103,379,1158,389,1194,390,1217,389,1211,403,1205,403,1206,415,1203,418,1203,425,1207,429",
	KalimantanTimur99:    "1202,454,1209,451,1210,464,1206,470,1209,472,1207,477,1202,477,1199,480,1193,475,1186,483,1186,486,1177,496,1171,500,1165,500,1165,505,1158,505,1155,509,1153,521,1142,525,1142,529,1147,529,1147,545,1144,546,1142,547,1145,550,1134,552,1097,511,1112,468,1103,446,1202,436,1200,441,1203,445",
	KalimantanSelatan100: "1133,551,1145,549,1145,550,1151,550,1154,547,1158,554,1154,572,1141,572,1141,597,1134,603,1127,594,1078,576,1078,565,1082,535,1096,511",
	KalimantanSelatan101: "1100,584,1128,595,1134,604,1129,612,1137,611,1137,618,1135,623,1137,628,1135,633,1136,633,1136,650,1124,659,1125,645,1122,641,1122,634,1113,636,1104,645,1072,654,1054,666,1051,665",
	KalimantanSelatan102: "1100,584,1051,665,1048,665,1048,656,1044,654,1044,636,1041,633,1041,622,1078,576",
	KalimantanTengah103:  "940,366,970,350,1000,352,1006,367,1006,384,1031,381,1096,427,1098,434,1087,436,1078,439,1071,441,1060,445,1051,451,1043,456,1034,464,1028,472,1024,481,1020,481,1012,483,1003,487,975,571,948,561,926,545,890,515,892,509,913,472,943,377",
	//KalimantanTengah103:  "940,366,970,350,1000,352,1006,367,1006,384,1031,381,1096,427,1111,469,1096,512,1082,536,1078,567,1078,576,1041,622,1041,628,1036,628,1021,618,1016,621,1017,579,1033,506,1036,492,1029,484,1020,481,1012,483,1003,487,975,571,948,561,926,545,890,515,892,509,913,472,943,377",
	KalimantanTengah104: "1016,577,1016,620,1010,628,998,628,997,612,995,609,989,608,989,610,984,617,978,617,977,607,971,606,966,598,1003,487,1020,480,1030,484,1037,492",
	KalimantanTengah105: "973,572,965,600,965,609,951,615,945,625,931,622,926,616,917,623,935,552,948,561",
	KalimantanTengah106: "934,552,916,625,908,632,901,632,903,623,900,614,903,601,897,594,904,584,892,552,890,515",
	KalimantanTengah107: "1098,435,1112,468,1096,511,1082,535,1078,564,1078,577,1041,622,1040,630,1020,618,1010,628,1015,618,1016,576,1037,492,1030,483,1024,481,1028,471,1036,462,1043,455,1055,447,1062,444,1073,441,1084,436",
	KalimantanBarat108:  "930,332,943,378,913,472,891,511,878,491,869,461,864,426,868,398,881,348,904,348",
	KalimantanBarat109:  "862,349,881,349,875,369,868,399,865,428,869,461,879,491,891,511,890,516,892,552,904,585,892,601,886,602,884,596,879,596,864,606,854,605,848,601,837,601,836,593,826,601,817,601,816,590,819,589,817,581,814,581,822,570,816,568,815,544,812,536,806,534,802,528,809,524,809,510,813,505,810,499,807,499,807,491,798,490,791,497,790,487,795,487,795,483,790,483,790,480,784,480,782,467,786,465,791,469,791,464,788,461,782,461,781,458,774,458,773,439,777,439,777,428,820,389,836,367,847,343",
	KalimantanBarat110:  "847,343,835,368,820,389,805,404,783,425,775,422,769,412,762,410,763,391,759,387,759,381,766,376,766,367,763,364,767,360,771,349,771,340,773,340,787,326,787,317",
	Sarawak111:          "964,300,951,305,939,317,932,330,904,347,862,350,789,318,790,308,808,292,811,293,845,262,858,266,870,265,876,274,881,275,892,272,916,247,930,256,945,271,957,287",
	Sarawak112:          "1074,280,1053,291,974,298,964,301,955,285,944,269,928,255,916,247,920,236,932,230,934,221,942,224,952,211,959,209,959,200,979,197,989,184,995,184,1003,169,1012,169,1012,151,1015,151,1030,158,1048,161,1068,161,1084,154,1121,211,1090,231",
	Sarawak113:          "1096,85,1101,98,1103,112,1102,125,1100,134,1095,142,1092,147,1087,151,1084,154,1069,161,1046,162,1029,159,1015,151,1037,140,1080,109,1080,98",
	Sarawak114:          "1138,64,1157,69,1191,64,1210,72,1219,66,1246,72,1249,77,1258,80,1280,95,1288,103,1288,113,1296,115,1297,124,1320,137,1325,149,1327,157,1322,162,1322,180,1318,184,1320,188,1316,192,1316,195,1314,199,1308,196,1307,198,1289,207,1275,205,1266,209,1255,209,1226,218,1220,218,1216,221,1207,221,1206,226,1197,222,1188,222,1159,209,1120,211,1084,154,1093,146,1100,134,1103,123,1103,112,1101,98,1096,86,1106,72,1115,65",
	Sea115:              "256,14,330,82,298,99,252,116,172,139,164,129,161,134,148,136,148,132,144,129,130,128,123,132,109,132,106,129,93,129,81,116,75,114,72,110,50,109,46,116,46,133,50,137,50,141,53,144,58,157,66,162,67,167,79,181,86,185,112,216,131,216,136,228,141,232,147,247,155,250,156,254,162,261,166,261,168,266,170,266,170,274,173,278,174,293,185,304,188,301,206,310,206,314,214,314,223,322,228,331,232,332,232,335,231,335,230,338,189,327,152,327,128,336,110,351,97,376,88,412,79,447,46,478,23,446,24,14",
	Sea116:              "328,784,291,740,261,711,231,690,178,642,115,569,46,478,78,447,97,375,110,350,129,334,154,327,189,327,230,338,230,341,236,350,236,358,243,374,243,392,254,397,254,420,258,421,262,416,264,420,272,420,279,423,285,433,283,437,285,441,286,447,301,454,303,462,303,465,313,476,317,488,319,492,317,495,319,496,319,500,324,504,326,504,327,513,330,516,330,524,340,536,340,552,356,581,365,585,373,604,398,628,413,636,412,643,416,647,416,652,441,674,447,685,460,690,447,707,416,735,383,757",
	Sea117:              "674,1110,477,1110,479,1099,479,1079,472,1048,457,1007,414,918,374,851,328,784,383,758,416,734,448,707,460,690,464,692,473,701,474,701,474,704,479,704,485,709,485,713,494,713,502,724,500,727,507,731,509,737,513,742,520,744,520,752,524,752,531,759,534,759,535,766,541,767,541,776,550,793,565,806,573,812,571,812,572,817,580,818,581,820,589,821,597,816,600,819,606,817,615,818,619,820,623,825,637,824,641,831,630,838,630,843,635,848,638,848,641,852,676,852,683,857,694,857,700,865,712,868,715,875,698,917,680,972,666,1028,663,1072,666,1091",
	Sea118:              "1049,1109,674,1111,666,1091,663,1073,666,1028,679,973,699,913,715,875,715,873,723,870,738,875,744,867,794,868,795,873,821,887,828,887,840,894,853,894,862,898,875,898,881,903,882,905,889,905,888,901,893,901,893,904,898,904,899,902,922,902,924,905,929,902,945,909,959,909,965,901,980,902,982,905,977,906,981,915,986,914,985,907,989,906,1002,913,1006,913,1014,916,1020,922,1023,917,1026,920,1036,921,1036,924,1039,929,1047,928,1047,923,1039,917,1038,913,1033,910,1037,904,1037,893,1038,890,1042,897,1041,906,1056,927,1056,955",
	Sea119:              "1430,1108,1048,1108,1052,1053,1056,961,1056,927,1041,905,1042,896,1038,891,1041,888,1042,887,1064,803,1186,821,1325,835,1305,868,1295,893,1293,911,1296,926,1308,940,1324,947,1338,948,1367,963,1390,980,1410,1006,1421,1031,1429,1062,1431,1086",
	Sea120:              "1691,843,1700,873,1710,931,1722,1009,1731,1075,1733,1094,1733,1111,1430,1109,1431,1087,1429,1060,1422,1032,1411,1007,1390,979,1366,962,1339,947,1323,946,1307,939,1295,927,1293,910,1295,892,1306,866,1325,834,1421,841,1488,843,1606,846",
	Sea121:              "1846,1111,1734,1110,1731,1088,1700,873,1690,836,1678,802,1658,763,1627,721,1573,668,1544,652,1607,602,1655,577,1733,552,1809,539,1814,552,1827,567,1837,573,1858,578,1871,586,1893,604,1906,619,1919,638,1925,652,1929,670,1933,685,1942,700,1949,707,1937,723,1916,755,1902,782,1890,808,1879,836,1873,857,1864,891,1855,924,1847,987,1845,1030,1844,1073",
	Sea122:              "2358,892,2343,894,2337,899,2328,895,2328,889,2318,894,2313,902,2309,900,2308,910,2300,911,2298,906,2290,906,2290,903,2280,911,2274,906,2250,906,2240,911,2248,894,2248,885,2256,873,2259,871,2263,860,2278,850,2290,850,2291,846,2302,846,2297,839,2310,839,2290,825,2310,819,2304,813,2296,813,2296,808,2301,808,2282,789,2282,776,2272,763,2277,751,2272,751,2270,756,2266,756,2265,741,2255,736,2247,725,2234,721,2225,710,2221,714,2216,714,2216,708,2197,711,2185,702,2176,698,2164,698,2161,690,2155,686,2141,687,2125,681,2112,683,2098,670,2088,670,2088,664,2085,664,2089,654,2081,659,2071,653,2071,660,2060,651,2063,648,2062,644,2057,649,2046,642,2044,644,2034,630,2021,641,2023,646,2017,659,2022,665,2027,665,2034,675,2026,671,2023,671,2014,663,2003,660,2000,664,1994,664,1970,684,1949,706,1932,731,1916,756,1892,804,1878,836,1866,878,1856,919,1847,985,1844,1040,1846,1111,2357,1110",
	Sea123:              "1971,683,1949,707,1941,699,1932,685,1925,651,1916,634,1904,617,1892,602,1857,577,1837,572,1826,566,1813,551,1805,522,1804,500,1808,468,1814,447,1829,419,1852,400,1878,388,1901,382,1930,378,1969,377,2026,382,2072,390,2166,401,2261,407,2358,411,2358,562,2353,559,2343,559,2320,540,2313,540,2300,529,2272,520,2267,520,2260,508,2229,524,2221,524,2214,537,2220,542,2220,548,2213,550,2207,549,2194,557,2187,554,2179,554,2173,559,2174,563,2171,567,2172,576,2165,581,2158,581,2157,586,2153,588,2153,597,2133,614,2132,623,2129,624,2125,621,2120,624,2111,619,2103,619,2103,612,2096,603,2097,592,2090,594,2086,579,2086,570,2080,570,2075,575,2078,581,2078,589,2059,555,2047,560,2049,565,2043,565,2040,570,2043,570,2043,577,2037,571,2036,578,2033,572,2017,573,2016,565,2009,565,2000,571,1991,581,1991,587,1985,589,1969,581,1957,581,1954,587,1950,587,1943,598,1960,597,1963,602,1973,603,1977,609,1977,613,1981,622,1989,621,1991,627,1992,639,1984,640,1988,646,1988,651,1994,664",
	Sea124:              "1542,651,1606,601,1629,589,1629,589,1655,576,1719,556,1719,556,1733,551,1810,538,1805,521,1803,498,1808,465,1814,444,1830,417,1852,398,1878,387,1901,381,1927,377,1968,375,2026,381,2069,388,2165,400,2358,410,2358,12,1630,13,1636,53,1639,105,1638,173,1634,231,1627,282,1613,344,1604,371,1585,421,1574,471,1538,516,1518,555,1506,606,1507,640",
	Sea125:              "1054,13,1631,14,1636,58,1639,103,1638,181,1636,218,1633,247,1627,282,1610,284,1588,292,1570,309,1556,326,1545,346,1539,364,1537,378,1522,378,1516,374,1512,374,1506,371,1500,375,1490,374,1484,382,1474,370,1452,371,1449,366,1437,366,1435,362,1424,367,1415,367,1415,361,1411,355,1406,355,1401,359,1392,351,1375,352,1377,358,1374,364,1374,371,1371,371,1361,383,1353,379,1352,369,1347,369,1347,383,1339,386,1336,386,1332,392,1334,397,1326,401,1326,414,1322,418,1326,424,1326,431,1325,432,1321,425,1319,425,1315,428,1315,434,1319,438,1323,439,1323,441,1319,448,1319,458,1323,465,1325,467,1324,474,1320,467,1317,471,1312,471,1305,483,1305,491,1298,494,1298,527,1302,533,1299,540,1292,540,1291,550,1287,558,1288,566,1283,574,1269,581,1269,593,1271,595,1271,605,1266,605,1266,612,1271,617,1271,622,1275,627,1275,631,1280,629,1295,629,1297,625,1303,628,1302,644,1311,656,1309,661,1309,672,1311,679,1305,697,1305,703,1275,715,1247,731,1224,748,1206,770,1197,724,1197,650,1197,601,1193,559,1172,507,1165,507,1165,499,1171,500,1186,487,1186,482,1193,476,1199,480,1203,476,1207,476,1209,471,1206,470,1210,464,1209,452,1203,454,1203,446,1199,440,1206,430,1203,424,1203,418,1206,414,1206,403,1211,402,1218,386,1225,381,1230,385,1230,375,1236,380,1252,384,1255,382,1258,385,1266,379,1272,384,1284,371,1284,367,1275,367,1272,363,1272,359,1265,356,1261,348,1255,347,1237,332,1234,326,1228,325,1221,314,1228,306,1234,302,1234,294,1230,292,1230,285,1216,277,1215,271,1213,270,1216,263,1209,263,1213,256,1209,257,1210,252,1205,251,1203,247,1195,248,1193,245,1195,244,1195,238,1203,233,1199,225,1189,225,1189,221,1197,221,1205,224,1207,220,1217,220,1220,218,1227,218,1255,209,1265,209,1274,204,1289,206,1308,197,1314,199,1319,192,1321,188,1318,184,1322,181,1322,162,1327,156,1324,143,1321,137,1297,124,1296,115,1288,113,1288,103,1276,92,1257,79,1249,77,1245,72,1220,65,1211,73,1191,64,1157,70,1137,63,1115,64,1106,73,1097,85,1060,36",
	Sea126:              "1054,14,1060,35,1072,54,1096,85,1079,100,1079,109,1033,142,1012,153,1012,170,1003,170,996,183,989,184,983,193,978,197,959,202,958,210,952,212,941,224,935,222,931,231,921,236,916,247,891,271,882,275,876,274,871,264,858,267,846,261,813,293,807,291,790,308,790,316,787,317,787,326,774,340,766,340,725,331,686,317,609,281,548,245,435,166,330,82,256,14",
	Sea127:              "548,245,494,300,461,328,428,349,428,344,425,340,422,341,415,336,402,337,405,344,399,344,389,333,379,333,377,330,385,330,392,322,392,315,387,309,379,309,378,312,374,309,366,315,355,304,356,299,351,299,338,302,342,313,326,307,310,282,310,275,305,277,302,276,296,276,300,272,295,261,290,259,285,249,273,244,269,237,264,236,253,224,249,224,248,221,245,222,232,214,227,205,221,204,220,200,216,200,216,197,206,197,207,174,202,174,198,170,193,169,193,163,184,146,173,139,254,115,298,99,331,82,435,166",
	Sea128:              "671,477,640,478,609,484,584,495,561,507,546,519,536,530,536,525,532,521,533,503,529,498,529,488,525,487,519,490,516,485,510,486,504,483,499,486,498,483,487,483,485,474,475,468,478,465,478,458,480,458,484,453,482,451,491,451,495,446,495,440,479,439,482,433,484,433,484,430,499,431,495,413,487,409,485,403,478,398,472,397,471,391,466,391,462,397,464,401,459,404,459,398,449,386,445,386,444,384,453,384,459,387,459,383,455,378,462,379,462,375,445,360,437,363,437,368,426,368,429,359,429,348,462,327,496,299,548,245,609,281,687,317,681,351,676,402",
	Sea129:              "698,696,656,697,620,703,608,707,604,702,609,689,604,681,606,656,609,651,609,645,603,640,613,616,616,613,615,600,610,596,606,597,602,590,602,581,595,581,595,575,591,572,591,563,578,563,570,558,568,563,558,562,553,557,553,545,550,541,547,541,542,542,543,537,536,533,536,529,546,520,561,507,608,484,639,478,671,477,672,537,676,583",
	Sea130:              "880,763,795,737,787,726,757,708,725,699,699,696,676,582,672,536,671,476,681,349,687,318,725,331,765,340,770,340,770,349,767,360,763,365,766,367,766,376,760,382,760,388,763,391,762,410,770,413,777,423,783,426,777,429,777,440,773,440,773,458,780,458,783,461,788,461,791,464,791,468,786,464,782,468,784,481,790,481,790,484,794,484,795,482,795,487,790,487,790,498,798,490,807,491,807,499,809,499,813,505,809,513,809,525,802,529,806,534,812,537,816,545,816,568,822,571,814,581,817,582,820,589,816,591,816,600,825,601,836,593,837,601,848,601,853,604,864,605,879,596,883,596,886,601,891,601,897,594,903,600,900,613,903,622,900,630,908,632,926,616,931,621,923,626,908,654,894,689,886,724",
	Sea131:              "814,819,805,824,796,821,796,815,784,820,776,821,771,813,766,816,750,816,750,807,747,796,739,794,733,785,728,785,722,789,719,789,715,785,712,785,708,777,704,781,697,781,693,777,690,777,681,765,677,765,674,768,670,765,666,765,661,769,661,773,652,773,649,769,630,770,625,766,621,770,618,770,615,765,611,765,611,769,603,773,603,787,598,793,599,798,596,798,587,812,586,814,580,813,580,807,574,811,564,806,550,793,541,776,541,767,546,767,546,758,535,747,535,743,541,743,555,751,562,759,571,760,567,754,570,751,566,746,574,742,586,755,588,754,589,759,596,759,596,763,602,759,602,751,605,748,605,716,609,707,621,702,658,696,698,696,727,699,759,709,788,726,795,738,797,754,806,775,810,792,813,807",
	Sea132:              "1064,803,1042,887,1040,887,1041,874,1030,865,1026,865,1021,861,1012,866,992,865,977,870,950,858,950,853,952,853,957,844,950,838,948,841,943,832,946,829,944,824,943,818,919,820,911,810,898,811,884,803,879,807,867,807,864,803,864,795,854,792,844,795,831,823,819,818,814,819,812,807,806,775,796,754,796,736,881,764,981,788",
	Sea133:              "1186,821,1131,814,1064,803,978,788,880,764,886,723,894,688,907,653,923,627,931,621,942,624,946,624,951,614,965,609,965,600,967,599,970,606,976,607,979,618,985,616,989,612,989,609,995,609,997,612,997,627,1010,628,1020,618,1037,628,1041,629,1041,632,1045,636,1045,652,1049,656,1049,665,1054,666,1072,654,1094,648,1104,645,1113,636,1118,636,1122,635,1122,641,1125,645,1125,658,1134,651,1137,650,1137,633,1134,632,1138,628,1134,623,1137,618,1137,610,1130,612,1130,610,1134,605,1134,601,1141,596,1141,572,1154,572,1158,553,1154,547,1150,551,1146,551,1143,548,1144,546,1148,546,1147,529,1141,529,1141,526,1154,523,1155,511,1158,506,1172,507,1192,558,1197,603,1197,652,1196,691,1198,727,1205,770,1194,791,1188,808",
	Sea134:              "1452,632,1507,641,1544,653,1573,668,1600,695,1627,722,1658,764,1679,805,1691,840,1691,843,1600,846,1324,835,1185,821,1188,807,1195,787,1206,770,1224,749,1250,730,1276,715,1305,703,1306,708,1301,718,1298,719,1298,724,1296,731,1301,746,1305,746,1305,751,1312,750,1318,754,1329,748,1337,748,1346,743,1350,742,1351,749,1356,748,1357,742,1352,733,1345,721,1349,714,1349,705,1352,705,1356,694,1353,689,1352,611,1346,604,1346,599,1343,597,1361,587,1375,576,1377,581,1384,583,1379,589,1384,592,1384,611,1381,614,1381,620,1374,627,1390,648,1397,648,1403,654,1409,655,1415,659,1415,665,1410,667,1410,692,1414,701,1420,705,1445,704,1440,696,1443,686,1450,681,1481,681,1484,679,1484,664,1473,661,1471,654,1474,652,1473,648,1466,649,1460,641,1456,642",
	Sea135:              "1538,378,1558,372,1561,363,1566,359,1572,359,1576,359,1576,357,1572,356,1572,352,1576,347,1579,347,1580,330,1587,332,1593,332,1593,336,1597,340,1590,344,1590,351,1581,373,1579,389,1577,391,1577,396,1571,399,1571,402,1564,403,1560,407,1546,407,1544,411,1504,412,1499,408,1498,402,1495,399,1443,399,1437,403,1428,403,1424,400,1411,399,1408,404,1404,399,1400,403,1397,403,1394,407,1387,403,1384,407,1378,402,1374,403,1372,400,1356,399,1342,417,1342,422,1336,434,1336,440,1338,447,1338,463,1347,479,1353,477,1363,483,1363,487,1367,492,1367,497,1371,498,1371,505,1375,500,1386,506,1392,506,1396,493,1403,493,1412,475,1420,475,1421,478,1433,478,1435,482,1438,480,1444,477,1447,470,1473,470,1479,474,1485,470,1485,468,1478,468,1505,459,1508,463,1516,463,1514,467,1518,475,1514,481,1513,489,1500,484,1498,480,1496,478,1484,479,1476,492,1474,500,1470,500,1467,501,1467,507,1456,514,1451,523,1437,523,1430,533,1419,538,1414,535,1410,535,1407,532,1406,537,1420,553,1425,553,1439,577,1439,582,1449,591,1453,592,1453,599,1455,605,1460,606,1460,611,1453,615,1453,618,1454,618,1459,622,1455,628,1453,625,1453,631,1451,632,1507,640,1506,606,1518,555,1537,518,1575,471,1586,420,1605,369,1621,313,1627,282,1611,283,1587,292,1555,328,1543,347,1538,363",

	JawaTengah136: "757,867,774,818,783,819,793,816,800,822,797,833,798,846,798,860,796,869",
	JawaTengah137: "838,891,842,853,837,811,831,822,818,819,806,821,802,821,798,838,801,855,796,871,822,887",
}

func (a *Area) Coords() template.HTML {
	if a.g.Version == 2 && a.ID == JawaTengah41 {
		return restful.HTML("")
	}
	return restful.HTML("%s", imageMapArea[a.ID])
}
