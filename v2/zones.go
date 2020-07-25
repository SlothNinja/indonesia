package indonesia

type Zone struct {
	g       *Game
	AreaIDS AreaIDS
}

func (z *Zone) id() int {
	if len(z.AreaIDS) > 0 {
		return int(z.AreaIDS[0])
	}
	return -1
}

type Zones []*Zone

func (z *Zone) Init(g *Game) {
	z.g = g
}

func newZone(g *Game, ids AreaIDS) *Zone {
	return &Zone{g: g, AreaIDS: ids}
}

func (z *Zone) Areas() (as Areas) {
	for _, id := range z.AreaIDS {
		if a := z.g.GetArea(id); a != nil {
			as = append(as, a)
		}
	}
	return
}

func (z *Zone) AdjacentSeaAreas() Areas {
	return z.adjacentAreas(isSea)
}

func (z *Zone) adjacentAreas(tests ...addAreaTest) Areas {
	var areas Areas
	for _, area := range z.Areas() {
		for _, a := range area.adjacentAreas(tests...) {
			if !areas.include(a) {
				areas = append(areas, a)
			}
		}
	}
	return areas
}

func (z *Zone) adjacentToArea(area *Area) bool {
	return z.adjacentAreas().include(area)
}

func (z *Zone) adjacentToZone(zone *Zone) bool {
	for _, area := range zone.Areas() {
		if z.adjacentToArea(area) {
			return true
		}
	}
	return false
}

func (z *Zone) same(zone *Zone) bool {
	return z.AreaIDS.same(zone.AreaIDS)
}

func (zs Zones) include(zone *Zone) bool {
	for _, z := range zs {
		if z.same(zone) {
			return true
		}
	}
	return false
}

func (z *Zone) overlaps(zone *Zone) bool {
	for _, aid := range zone.AreaIDS {
		if z.AreaIDS.include(aid) {
			return true
		}
	}
	return false
}

func (zs Zones) intersection(zones Zones) Zones {
	var common Zones
	for _, zone := range zones {
		if zs.include(zone) {
			common = append(common, zone)
		}
	}
	return common
}

func (zs Zones) addZone(zone *Zone) Zones {
	var zonesToMerge, zones Zones
	for _, z := range zs {
		if z.overlaps(zone) || z.adjacentToZone(zone) {
			zonesToMerge = append(zonesToMerge, z)
		} else {
			zones = append(zones, z)
		}
	}
	if len(zonesToMerge) == 0 {
		return append(zs, zone)
	}
	for _, z := range zonesToMerge {
		zone.AreaIDS = zone.AreaIDS.addUnique(z.AreaIDS...)
	}
	return append(zones, zone)
}

func (zs Zones) addZones(zones ...*Zone) Zones {
	for _, zone := range zones {
		zs = zs.addZone(zone)
	}
	return zs
}

func (z *Zone) Goods() Goods {
	if area := z.g.Areas[z.AreaIDS[0]]; area != nil {
		return area.Goods()
	} else {
		return NoGoods
	}
}

func (z Zones) Areas() Areas {
	var areas Areas
	for _, zone := range z {
		areas = append(areas, zone.Areas()...)
	}
	return areas
}

func (z *Zone) minCapacityFor(s *Shipper) int {
	minShips := 10
	for _, area := range z.Areas() {
		ships := s.shipsInArea(area)
		if ships < minShips {
			minShips = ships
		}
	}
	return minShips * s.HullSize()
}

func (z *Zone) contiguous() bool {
	return z.g.contiguous(z.AreaIDS)
	//return z.AreaIDS.contiguous()
}

func (zs Zones) contiguous() bool {
	for _, z := range zs {
		if !z.contiguous() {
			return false
		}
	}
	return true
}
