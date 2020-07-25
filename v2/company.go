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

func (c *Company) remove(a *Area) {
	var zones Zones
	for _, zone := range c.Zones {
		a.Producer = nil
		zone.AreaIDS = zone.AreaIDS.remove(a.ID)
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
