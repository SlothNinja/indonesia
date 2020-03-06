package indonesia

import (
	"fmt"
	"html/template"
)

type Company struct {
	g        *Game
	OwnerID  int
	Slot     int
	Deeds    Deeds
	Merged   bool
	ShipType ShipType
	Operated bool
	Zones    Zones
}

func (c *Company) Equal(company *Company) bool {
	if c == nil || company == nil {
		return false
	}
	return c.OwnerID == company.OwnerID && c.Slot == company.Slot
}

type Companies []*Company

func (cs Companies) include(c *Company) bool {
	for _, company := range cs {
		if company.OwnerID == c.OwnerID && company.Slot == c.Slot {
			return true
		}
	}
	return false
}

func (c *Company) Game() *Game {
	return c.g
}

func (c *Company) Init(g *Game) {
	c.g = g
	for _, z := range c.Zones {
		z.Init(g)
	}
}

func (c *Company) Delivered() int {
	d := 0
	if c == nil {
		return 0
	}
	for _, a := range c.Areas() {
		if a.Used {
			d += 1
		}
	}
	return d
}

func (c *Company) Ships() int {
	ships := 0
	for _, area := range c.Areas() {
		ships += c.ShipsIn(area)
	}
	return ships
}

func (c *Company) ShipsIn(a *Area) int {
	ships := 0
	if c.IsProductionCompany() {
		return ships
	}
	for _, shipper := range a.Shippers {
		if shipper.Company() == c {
			ships += 1
		}
	}
	return ships
}

//func (c *Company) resetDelivered() {
//	for _, z := range c.Zones {
//                z.Delivered = 0
//	}
//}

func (c *Company) deliveredAllGoods() bool {
	return c.IsProductionCompany() && c.Delivered() >= len(c.Areas())
}

func (c *Company) Goods() Goods {
	if c == nil {
		return NoGoods
	}
	switch l := len(c.Deeds); {
	case l < 1:
		return NoGoods
	case l == 1:
		return c.Deeds[0].Goods
	default:
		goods := c.Deeds[0].Goods
		for _, d := range c.Deeds[1:] {
			if goods != d.Goods {
				if (goods == Rice && d.Goods == Spice) || (goods == Spice && d.Goods == Rice) {
					return SiapFaji
				} else {
					return NoGoods
				}
			}
		}
		return goods
	}
}

func (c *Company) IsProductionCompany() bool {
	goods := c.Goods()
	return goods != NoGoods && goods != Shipping
}

func (c *Company) IsShippingCompany() bool {
	return c.Goods() == Shipping
}

func (c *Company) Production() int {
	count := 0
	for _, z := range c.Zones {
		count += len(z.AreaIDS)
	}
	return count
}

func (c *Company) MaxShips() int {
	count := 0
	for _, deed := range c.Deeds {
		count += deed.MaxShips[c.g.Era]
	}
	return count
}

func (c *Company) AddShipIn(a *Area) {
	a.AddShip(c)
	c.AddArea(a)
}

func (c *Company) AddArea(a *Area) {
	c.Zones = c.Zones.addZones(newZone(c.g, AreaIDS{a.ID}))
}

//	if !c.Areas().include(a) {
//		for _, zone := range c.Zones {
//			for _, area := range zone.Areas() {
//				if area.AdjacentAreas().include(a) {
//					zone.AreaIDS = append(zone.AreaIDS, a.ID)
//					return
//				}
//			}
//		}
//		c.Zones = append(c.Zones, newZone(c.g, AreaIDS{a.ID}))
//	}
// }

func (c *Company) RemoveArea(a *Area) {
	if c.Areas().include(a) {
		for _, zone := range c.Zones {
			for _, area := range zone.Areas() {
				if area.AdjacentAreas().include(a) {
					zone.AreaIDS = zone.AreaIDS.remove(a.ID)
					return
				}
			}
		}
	}
}

func (c *Company) Areas() Areas {
	var areas Areas
	for _, zone := range c.Zones {
		areas = append(areas, zone.Areas()...)
	}
	return areas
}

func (a *Area) Goods() Goods {
	switch {
	case a.IsLand() && a.Producer != nil:
		return a.Producer.Goods
	case a.IsSea() && len(a.Shippers) > 0:
		return Shipping
	default:
		return NoGoods
	}
}

func (c *Company) ZoneFor(a *Area) *Zone {
	if c == nil || a == nil {
		return nil
	}
	for _, zone := range c.Zones {
		if zone.AreaIDS.include(a.ID) {
			return zone
		}
	}
	return nil
}

var noAcquiredCompanyIndex = -1

func newCompany(g *Game, owner *Player, index int, d *Deed) *Company {
	return &Company{
		g:        g,
		OwnerID:  owner.ID(),
		Slot:     index,
		Deeds:    Deeds{d},
		Merged:   false,
		ShipType: NoShipType,
	}
}

func (c *Company) Owner() *Player {
	if c == nil {
		return nil
	}
	return c.g.PlayerByID(c.OwnerID)
}

func (c *Company) HTML() template.HTML {
	return template.HTML(c.String())
}

func (c *Company) String() string {
	if c == nil {
		return ""
	}
	return fmt.Sprintf("%s %s", c.Province(), c.Goods())
}

func (c *Company) Province() Province {
	if len(c.Deeds) > 0 {
		return c.Deeds[0].Province
	}
	return NoProvince
}

func (g *Game) AllCompaniesOperated() bool {
	for _, p := range g.Players() {
		if p.HasCompanyToOperate() {
			return false
		}
	}
	return true
}

func (g *Game) Companies() Companies {
	var companies Companies
	for _, p := range g.Players() {
		companies = append(companies, p.Companies()...)
	}
	return companies
}

func (g *Game) ShippingCompanies() map[Province]*Company {
	companies := make(map[Province]*Company, 0)
	for _, p := range g.Players() {
		for _, company := range p.Companies() {
			if company.IsShippingCompany() {
				companies[company.Province()] = company
			}
		}
	}
	return companies
}

func (g *Game) resetCompanies() {
	for _, company := range g.Companies() {
		for _, area := range company.Areas() {
			area.Used = false
		}
		company.Operated = false
	}
}

//func (c *Company) maxDelivery() int {
//	if c.Goods() == Shipping {
//		return 0
//	}
//	routes, production := c.deliveryRoutes(), c.Production()
//	c.g.debugf("maxDelivery Routes: %s", routes)
//	switch {
//	case len(routes) < 2:
//		return len(routes)
//	case production < 2:
//		return c.Production()
//	default:
//		areas := c.g.Areas.copy()
//		c.g.debugf("maxDelivery c.Areas: %#v", c.Areas())
//		c.g.debugf("maxDelivery Areas: %#v", areas)
//		count := 0
//		for _, route := range routes {
//			var ok bool
//			if areas, ok = areas.useRoute(route); ok {
//				count += 1
//			}
//			if count == production {
//				return count
//			}
//		}
//		return count
//	}
//}
//
//func (as Areas) useRoute(route *Route) (Areas, bool) {
//	areas := as.copy()
//	for _, aid := range route.AreaIDS {
//		area := areas[aid]
//		if !area.Used {
//			area.Used = true
//		} else {
//			return as, false
//		}
//	}
//	return areas, true
//}

func (a *Area) demands(goods Goods) bool {
	return a.City != nil && a.City.demands(goods)
}

func (c *City) demands(goods Goods) bool {
	return c.Delivered[goods] < c.Size
}

func (c *City) demandFor(goods Goods) int {
	return c.Size - c.Delivered[goods]
}

func (c *City) hasDemandFor(goods Goods) bool {
	return c.demandFor(goods) > 0
}

func (cs Cities) demandFor(goods Goods) int {
	demand := 0
	for _, city := range cs {
		demand += city.demandFor(goods)
	}
	return demand
}

//type Route struct {
//	g       *Game
//	Zone    *Zone
//	Goods   Goods
//	Shipper *Shipper
//	AreaIDS AreaIDS
//}
//
//type Routes []*Route
//
//func (r *Route) lastArea() *Area {
//	return r.g.Areas[r.AreaIDS[len(r.AreaIDS)-1]]
//}
//
//func (r *Route) copy() *Route {
//	l := len(r.AreaIDS)
//	route := &Route{
//		g:       r.g,
//		Zone:    r.Zone,
//		Goods:   r.Goods,
//		Shipper: r.Shipper,
//		AreaIDS: make(AreaIDS, l),
//	}
//	if elements := copy(route.AreaIDS, r.AreaIDS); elements == l {
//		return route
//	} else {
//		return nil
//	}
//}
//
//func (r *Route) include(a *Area) bool {
//	return r.AreaIDS.include(a.ID)
//}
//
//func (r *Route) add(a *Area) *Route {
//	r.AreaIDS = append(r.AreaIDS, a.ID)
//	return r
//}
//
//func (r *Route) String() string {
//	s := ""
//	for _, id := range r.AreaIDS {
//		s += fmt.Sprintf("%d -> ", id)
//	}
//	return s
//}
//
//func (rs Routes) String() string {
//	s := ""
//	for i, r := range rs {
//		s += fmt.Sprintf("Route %d: %s", i, r)
//	}
//	return s
//}

func hasAShipper(a *Area) bool {
	return len(a.Shippers) > 0
}

func (a *Area) hasAShipper() bool {
	return len(a.Shippers) > 0
}

func (a *Area) hasShipper(s *Shipper) bool {
	for _, shipper := range a.Shippers {
		if shipper != nil && s != nil && shipper.equals(s) {
			return true
		}
	}
	return false
}

func (a *Area) hasShippingCapacity() bool {
	return a.Shippers.haveCapacity()
}

func (a *Area) hasShippingCapacityFor(s *Shipper) bool {
	for _, shipper := range a.Shippers {
		if shipper.equals(s) {
			return s.hasCapacity()
		}
	}
	return false
}

func hasShippingCapacity(a *Area) bool {
	return a.hasShippingCapacity()
}

func (s *Shipper) hasCapacity() bool {
	return s.Delivered < s.HullSize()
}

func (ss Shippers) haveCapacity() bool {
	for _, s := range ss {
		if s.hasCapacity() {
			return true
		}
	}
	return false
}

//func (c *Company) deliveryRoutes() Routes {
//	c.g.debugf("enter c.deliveryRoutes")
//	defer c.g.debugf("exit c.deliveryRoutes")
//	if goods := c.Goods(); goods == Shipping {
//		return Routes{}
//	} else {
//		var routes Routes
//		for _, zone := range c.Zones {
//			if newRoutes := zone.deliveryRoutes(goods); newRoutes != nil {
//				routes = append(routes, newRoutes...)
//			}
//		}
//		return routes
//	}
//}
//
//func extendRoute(route *Route, areas Areas) Routes {
//	route.g.debugf("enter extendRoute: route: %s areas: %#v", route, areas)
//	defer route.g.debugf("exit extendRoute")
//	var routes Routes
//	for _, area := range areas {
//		route.g.debugf("Area ID: %d", area.ID)
//		if !route.AreaIDS.include(area.ID) {
//			route.g.debugf("Area %d demands %s %v", area.ID, route.Goods, area.demands(route.Goods))
//			switch {
//			case area.demands(route.Goods):
//				r := route.copy()
//				r.AreaIDS = append(r.AreaIDS, area.ID)
//				routes = append(routes, r)
//			case route.Shipper == nil && area.hasAShipper():
//				for _, shipper := range area.Shippers {
//					r := route.copy()
//					r.AreaIDS = append(r.AreaIDS, area.ID)
//					r.Shipper = shipper
//					if newRoutes := extendRoute(r, area.adjacentAreas()); newRoutes != nil {
//						routes = append(routes, newRoutes...)
//					}
//				}
//			case area.hasShipper(route.Shipper):
//				r := route.copy()
//				r.AreaIDS = append(r.AreaIDS, area.ID)
//				if newRoutes := extendRoute(r, area.adjacentAreas()); newRoutes != nil {
//					for i := 0; i < route.Shipper.HullSize(); i++ {
//						routes = append(routes, newRoutes...)
//					}
//				}
//			}
//		}
//	}
//	return routes
//}
//
//func (g *Game) expandRoutes(rs Routes, goods Goods) Routes {
//	var routes Routes
//	for _, route := range rs {
//		max := 1000
//		for _, aid := range route.AreaIDS {
//			area := g.Areas[aid]
//			g.debugf("Area: %#v", area)
//			g.debugf("len(area.Shippers) > 0 : %v", len(area.Shippers))
//			g.debugf("area.Shippers %v", area.Shippers)
//			switch {
//			case area.Producer != nil:
//				if goods := len(area.GoodsCompany().ZoneFor(area).AreaIDS); goods < max {
//					max = goods
//				}
//			case len(area.Shippers) > 0:
//				for _, shipper := range area.Shippers {
//					if shipper == route.Shipper {
//						if capacity := shipper.HullSize(); capacity < max {
//							max = capacity
//						}
//					}
//				}
//			default:
//				if capacity := area.City.Size - area.City.Delivered[goods]; capacity < max {
//					max = capacity
//				}
//			}
//		}
//		for i := 0; i < max; i++ {
//			routes = append(routes, route)
//		}
//	}
//	return routes
//}

//var companyValues = sslice{"Slot", "Merged", "Operated", "ShipType"}
//
//func adminCompany(g *Game, form url.Values) (string, game.ActionType, error) {
//	if err := g.adminUpdateCompany(companyValues); err != nil {
//		return "indonesia/flash_notice", game.None, err
//	}
//
//	return "", game.Save, nil
//}
//
//func (g *Game) adminUpdateCompany(ss sslice) error {
//	if err := g.validateAdminAction(); err != nil {
//		return err
//	}
//
//	values, err := g.getValues()
//	if err != nil {
//		return err
//	}
//
//	company := g.SelectedCompany()
//	var d *Deed
//	for key := range values {
//		if !ss.include(key) {
//			//			g.debugf("Key: %q", key)
//			var k0, k1 string
//			if keys := strings.Split(key, "-"); len(keys) > 1 {
//				k0, k1 = keys[0], keys[1]
//			} else {
//				k0 = keys[0]
//			}
//			switch k0 {
//			case "AddZone":
//				if values.Get(key) == "true" {
//					company.Zones = append(company.Zones, newZone(g, AreaIDS{}))
//				}
//			case "AddDeed":
//				if v := values.Get(key); v != "none" {
//					d = g.Deeds().get(v)
//				}
//			case "AddAreaZone":
//				if value := values.Get(key); value != "none" {
//					if id, err := strconv.Atoi(value); err == nil {
//						if zindex, err := strconv.Atoi(k1); err == nil {
//							company.Zones[zindex].AreaIDS =
//								append(company.Zones[zindex].AreaIDS, AreaID(id))
//						}
//					}
//				}
//			case "RemoveAreaZone":
//				if value := values.Get(key); value != "none" {
//					if id, err := strconv.Atoi(value); err == nil {
//						if zindex, err := strconv.Atoi(k1); err == nil {
//							zone := company.Zones[zindex]
//							zone.AreaIDS = zone.AreaIDS.remove(AreaID(id))
//							if len(zone.AreaIDS) == 0 {
//								company.removeZoneAt(zindex)
//							}
//						}
//					}
//				}
//			}
//			delete(values, key)
//		}
//	}
//
//	schema.RegisterConverter(ShipType(0), convertShipType)
//	if err := schema.Decode(company, values); err != nil {
//		return err
//	}
//	if d != nil {
//		company.Deeds = append(company.Deeds, d)
//	}
//	return nil
//}

func (c *Company) remove(a *Area) {
	var zones Zones
	for _, zone := range c.Zones {
		zone.AreaIDS = zone.AreaIDS.remove(a.ID)
		//		c.g.debugf("len(zone.AreaIDS)", len(zone.AreaIDS))
		if len(zone.AreaIDS) > 0 {
			zones = append(zones, zone)
		}
	}
	c.Zones = zones
}

func (c *Company) removeZoneAt(i int) {
	c.Zones = append(c.Zones[:i], c.Zones[i+1:]...)
}

func (c *Company) canDeliverGood() bool {
	switch {
	case !c.IsProductionCompany():
		return false
	case c.adjacentShippingCapacity() == 0:
		return false
	default:
		return true
	}
}

func (c *Company) adjacentShippingCapacity() int {
	capacity := 0
	for _, zone := range c.Zones {
		capacity += zone.adjacentShippingCapacity()
	}
	return capacity
}

func (z *Zone) adjacentShippingCapacity() int {
	capacity, goods, hulls := 0, len(z.Areas()), 0
	for _, area := range z.adjacentAreas(hasAShipper) {
		for _, shipper := range area.Shippers {
			hulls += shipper.HullSize()
		}
	}
	if hulls > goods {
		capacity += goods
	} else {
		capacity += hulls
	}
	return capacity
}

func (c *Company) deliveredAdjacentShippingCapacity() bool {
	return c.IsProductionCompany() && c.Delivered() >= c.adjacentShippingCapacity()
}

//func (c *Company) madeAllRequiredDeliveries() bool {
//	return c.IsProductionCompany() && c.Delivered() == c.requiredDeliveries()
//}

//func (c *Company) connectedCityDemand() int {
//	demand, goods := 0, c.Goods()
//	totalDemand := c.g.Cities().demandFor(goods)
//	maxZoneShipCap := c.maxZoneShipCap()
//	cappedDemand := min(totalDemand, maxZoneShipCap, c.Production())
//	c.g.debugf("totalDemand: %d", totalDemand)
//	c.g.debugf("maxZoneShipCap: %d", maxZoneShipCap)
//	c.g.debugf("cappedDemand: %d", cappedDemand)
//	if cappedDemand == 0 {
//		return 0
//	}
//	for _, city := range c.g.Cities() {
//		cityDemand := city.connectedDemandFor(c)
//		demand = min(demand+cityDemand, cappedDemand)
//		c.g.debugf("cityDemand: %v", cityDemand)
//		c.g.debugf("demand: %d", demand)
//		if demand == cappedDemand {
//			return demand
//		}
//	}
//	return demand
//}
//
//func (c *City) connectedDemandFor(company *Company) int {
//	demand := 0
//	goods := company.Goods()
//	company.g.debugf("City in %s", c.a.Province())
//	company.g.debugf("Goods: %s", goods)
//	maxCityDemand := c.demandFor(goods)
//	company.g.debugf("maxCityDemand: %d", maxCityDemand)
//	if maxCityDemand == 0 {
//		return 0
//	}
//	for _, zone := range company.Zones {
//		for _, shipper := range c.shippers() {
//			shipCap := shipper.capacityBetween(zone, c)
//			numGoods := len(zone.AreaIDS)
//			demand = min(demand+min(shipCap, numGoods), maxCityDemand)
//			company.g.debugf("shipCap: %v", shipCap)
//			company.g.debugf("numGoods: %v", numGoods)
//			company.g.debugf("demand: %d", demand)
//			if demand == maxCityDemand {
//				return demand
//			}
//		}
//	}
//	return demand
//}

func (c *Company) maxZoneShipCap() int {
	capacity := 0
	for _, zone := range c.Zones {
		zoneCap := 0
		production := len(zone.AreaIDS)
		for _, area := range zone.AdjacentSeaAreas() {
			for _, shipper := range area.Shippers {
				zoneCap = min(zoneCap+shipper.HullSize(), production)
			}
		}
		capacity += zoneCap
	}
	return capacity
}

//func (c *Company) deliveredConnectedCityDemand() bool {
//	return c.IsProductionCompany() && c.Delivered() >= c.connectedCityDemand()
//}

//func (c *Company) deliveredRequiredDeliveries() bool {
//	return c.IsProductionCompany() && c.Delivered() >= c.requiredDeliveries()
//}

func (c *City) shippers() Shippers {
	var shippers Shippers
	for _, area := range c.a.AdjacentSeaAreas() {
		for _, shipper := range area.Shippers {
			if !shippers.include(shipper) {
				shippers = append(shippers, shipper)
			}
		}
	}
	return shippers
}

func (s *Shipper) capacityBetween(z *Zone, c *City) int {
	zones1 := s.zonesAdjacentToZone(z)
	zones2 := s.zoneAdjacentToCity(c)
	if common := zones1.intersection(zones2); common == nil {
		return 0
	} else {
		//		s.g.debugf("common: %#v", common)
		capacity := 0
		for _, z := range common {
			capacity += z.minCapacityFor(s)
		}
		return capacity
	}
}

func (s *Shipper) zoneAdjacentToCity(c *City) Zones {
	var zones Zones
	if company := s.Company(); company == nil {
		return nil
	} else {
		for _, zone := range company.Zones {
			if zone.adjacentToArea(c.a) {
				zones = append(zones, zone)
			}
		}
	}
	return zones
}

func (s *Shipper) zonesAdjacentToZone(z *Zone) Zones {
	var zones Zones
	if company := s.Company(); company == nil {
		return nil
	} else {
		for _, zone := range company.Zones {
			if zone.adjacentToZone(z) {
				zones = append(zones, zone)
			}
		}
	}
	return zones
}
