package indonesia

import "fmt"

type FlowID struct {
	AreaID   AreaID
	PID      int
	Index    int
	IO       int
	Province Province
}
type FlowIDS []FlowID

var sourceFID = toFlowID(sourceAID)
var targetFID = toFlowID(targetAID)
var noFID = toFlowID(NoArea)

func (fids FlowIDS) include(fid FlowID) bool {
	for _, id := range fids {
		if id == fid {
			return true
		}
	}
	return false
}

func (fids FlowIDS) Strings() (s string) {
	s += "{ "
	for _, fid := range fids {
		s += fid.String()
	}
	return s + " }"
}

type subflow map[FlowID]int
type flowMatrix map[FlowID]subflow

func (fm flowMatrix) String() (s string) {
	for fid, sf := range fm {
		s += fmt.Sprintf("%s: %s\n", fid, sf)
	}
	return s
}

func (fid FlowID) String() string {
	return fmt.Sprintf("[%d %d %d %d %d]", fid.AreaID, fid.PID, fid.Index, fid.IO, fid.Province.Int())
}

func (sf subflow) String() (s string) {
	s += "{ "
	for fid, flow := range sf {
		s += fmt.Sprintf("%s:%d ", fid, flow)
	}
	return s + " }"
}

// Sufficiently large number that it does not limit max flow
const infinity = 10000

func (c *Company) maxFlow() (int, flowMatrix) {
	flow1, fm1 := c.maxFlow2()
	//	c.g.debugf("initial flow: %d delivered: %s", flow1, fm1)
	if cp := c.g.CurrentPlayer(); cp.hasShippingCompany() {
		hullsizes := c.g.getHullSizes()
		for _, p := range c.g.Players() {
			if pid := p.ID(); pid == cp.ID() {
				p.Technologies[HullTech] = hullsizes[pid]
			} else {
				p.Technologies[HullTech] = -1
			}
		}
		for i := 0; i < 5; i++ {
			for _, p := range c.g.Players() {
				if pid := p.ID(); pid == cp.ID() {
					p.Technologies[HullTech] = hullsizes[pid]
				} else {
					p.Technologies[HullTech] = min(p.Technologies[HullTech]+1, hullsizes[pid])
				}
				//				c.g.debugf("pid: %d hullsize: %d", p.ID(), p.Technologies[HullTech])
			}
			flow2, fm2 := c.maxFlow2()
			//			c.g.debugf("flow1: %d fm1: %s", flow1, fm1)
			//			c.g.debugf("ProposedShips: %#v", c.g.ProposedShips(fm1))
			//			c.g.debugf("flow2: %d fm2: %s", flow2, fm2)
			//			c.g.debugf("ProposedShips: %#v", c.g.ProposedShips(fm2))
			if flow2 == flow1 {
				c.g.restoreHullSizes(hullsizes)
				return flow2, fm2
			}
		}
		c.g.restoreHullSizes(hullsizes)
	}
	return flow1, fm1
}

func (g *Game) getHullSizes() map[int]int {
	hs := make(map[int]int, 0)
	for _, p := range g.Players() {
		hs[p.ID()] = p.Technologies[HullTech]
	}
	return hs
}

func (g *Game) restoreHullSizes(hs map[int]int) {
	for pid, size := range hs {
		g.PlayerByID(pid).Technologies[HullTech] = size
	}
}

// Edmonds Karp
func (c *Company) maxFlow2() (flow int, fm flowMatrix) {
	fm = make(flowMatrix, 0)
	for newFlow, parentTable := c.search(fm); newFlow > 0; newFlow, parentTable = c.search(fm) {
		flow += newFlow

		// Backtrach search and write flow
		v := targetFID
		for v != sourceFID {
			u := parentTable[v]
			if _, ok := fm[u]; !ok {
				fm[u] = make(subflow, 0)
			}
			fm[u][v] += newFlow

			if _, ok := fm[v]; !ok {
				fm[v] = make(subflow, 0)
			}
			fm[v][u] -= newFlow
			v = u
		}
		//		c.g.debugf("flow: %d fm: %s", flow, fm)
	}
	return flow, fm
}

type parentTable map[FlowID]FlowID
type foundCapTo map[FlowID]int

func (c *Company) search(fm flowMatrix) (int, parentTable) {
	parentTable := make(parentTable, 0)
	foundCapTo := make(foundCapTo, 0)

	// make sure source is not rediscovered
	parentTable[sourceFID] = noFID
	foundCapTo[sourceFID] = infinity

	q := make(queue, 0)
	q.push(sourceFID)
	for len(q) > 0 {
		u := q.pop()
		ids := c.neighboringFIDSFor(u)
		//		c.g.debugf("\n\n=============\nneighboringFIDSFor(%d): %s", u, ids)
		for _, v := range ids {

			// If there is residual capacity and v is not seen before in search
			capBetween := c.capBetween(u, v)
			//			c.g.debugf("\n\ncapBetween(%d, %d): %d", u, v, capBetween)
			residualCap := capBetween - fm[u][v]
			//			c.g.debugf("residualCapBetween(%d, %d): %d", u, v, residualCap)
			if _, seen := parentTable[v]; residualCap > 0 && !seen {
				parentTable[v] = u
				//                                c.g.debugf("parentTable[%s]: %#v", v, parentTable[v])
				foundCapTo[v] = min(foundCapTo[u], residualCap)
				if v != targetFID {
					q.push(v)
				} else {
					return foundCapTo[targetFID], parentTable
				}
			}
		}

	}
	return 0, parentTable
}

type queue FlowIDS

func (q queue) popS() (FlowID, queue) {
	if len(q) > 0 {
		return q[0], q[1:]
	}
	return noFID, q
}

func (q *queue) pop() FlowID {
	id, nq := q.popS()
	*q = nq
	return id
}

func (q queue) pushS(id FlowID) queue {
	return append(q, id)
}

func (q *queue) push(id FlowID) {
	nq := q.pushS(id)
	*q = nq
}

//const (
//	AreaIDMask   FlowID = PIDMask * 1000
//	PIDMask             = IndexMask * 10
//	IndexMask           = IOMask * 100
//	IOMask              = ProvinceMask * 10
//	ProvinceMask        = 1000
//)

//type decodedFlowID struct {
//	AreaID   AreaID
//	PID      int
//	Index    int
//	IO       int
//	Province Province
//}
//
//func decodeFlowID(flowID FlowID) *decodedFlowID {
//	decodedFlowID := new(decodedFlowID)
//
//	decodedFlowID.AreaID = AreaID(flowID / AreaIDMask)
//	PIDIndexIOProvince := (flowID % AreaIDMask)
//
//	decodedFlowID.PID = int(PIDIndexIOProvince / PIDMask)
//	IndexIOProvince := (PIDIndexIOProvince % PIDMask)
//
//	decodedFlowID.Index = int(IndexIOProvince / IndexMask)
//	IOProvince := (IndexIOProvince % IndexMask)
//
//	decodedFlowID.IO = int(IOProvince / IOMask)
//	decodedFlowID.Province = Province(IOProvince % IOMask)
//
//	return decodedFlowID
//}

func toFlowID(aid AreaID, args ...int) FlowID {
	if len(args) == 4 {
		return FlowID{aid, args[0], args[1], args[2], Province(args[3])}
	}
	return FlowID{aid, -1, -1, -1, -1}
}

//func toFlowID(aid AreaID, pid, index, io int, province Province) FlowID {
//	return (FlowID(aid) * AreaIDMask) + (FlowID(pid) * PIDMask) + (FlowID(index) * IndexMask) +
//		(FlowID(io) * IOMask) + FlowID(province)
//}

//func (g *Game) getAreaWithIOProvincePIDIndexIO(id AreaID) (*Area, io, Province, int, int, int) {
//	if id.isLandID() || id.isSeaID() {
//		return g.Areas[id], NoPlayerID, -1, -1
//	}
//	i := int(id)
//	io := (i / 10000000)
//	indexPIDAID := (i % 10000000)
//	index := (indexPIDAID / 100000) - 1
//	pidAID := (indexPIDAID % 100000)
//	pid := pidAID / 1000
//	aid := AreaID(pidAID % 1000)
//	if aid.isSeaID() {
//		return g.Areas[aid], pid, index, io
//	}
//	return nil, NoPlayerID, -1, -1
//}

func (c *Company) capBetween(from, to FlowID) int {
	fromArea := c.g.GetArea(from.AreaID)
	toArea := c.g.GetArea(to.AreaID)
	shippers := c.g.ShippingCompanies()
	switch {
	case from == sourceFID && toArea.IsLand() && toArea.hasProducer():
		zone := c.ZoneFor(toArea)
		return len(zone.AreaIDS)
	case fromArea == nil:
		return 0
	case fromArea.IsLand() && fromArea.hasCity() && to == targetFID:
		demand := fromArea.City.demandFor(c.Goods())
		return demand
	case toArea == nil:
		return 0
	case fromArea.IsLand() && fromArea.hasProducer() &&
		toArea.IsSea() && toArea.hasAShipper():
		production := len(c.ZoneFor(fromArea).AreaIDS)
		shipper := shippers[to.Province]
		return min(shipper.HullSize(), production)
	case fromArea.IsSea() && fromArea.hasAShipper() &&
		toArea.IsLand() && toArea.hasCity():
		demand := toArea.City.demandFor(c.Goods())
		shipper := shippers[from.Province]
		capacity := min(shipper.HullSize(), demand)
		return capacity
	case fromArea.IsSea() && fromArea == toArea:
		if from.IO == shipInput && to.IO == shipOutput {
			shipper := shippers[from.Province]
			return shipper.HullSize()
		}
		return 0
	case fromArea.IsSea() && fromArea != toArea && (from.IO == shipInput || to.IO == shipOutput):
		return 0
	case fromArea.IsSea() && fromArea.hasAShipper() &&
		toArea.IsSea() && toArea.hasAShipper() && to.Province == from.Province:
		shipper := shippers[from.Province]
		return shipper.HullSize()
	default:
		return 0
	}
}

func (c *Company) neighboringFIDSFor(from FlowID) (fids FlowIDS) {
	area := c.g.GetArea(from.AreaID)
	switch {
	case from == sourceFID:
		for _, zone := range c.Zones {
			fids = append(fids, toFlowID(zone.AreaIDS[0]))
		}
	case from == targetFID:
		for _, city := range c.g.Cities() {
			fids = append(fids, toFlowID(city.a.ID))
		}
	case area.hasCity():
		for _, a := range area.AdjacentSeaAreas() {
			for i, shipper := range a.Shippers {
				fid := toFlowID(a.ID, shipper.OwnerID, i, shipOutput, shipper.Province().Int())
				fids = append(fids, fid)
			}
		}
		fids = append(fids, targetFID)
	case area.hasProducer():
		for _, a := range c.ZoneFor(area).adjacentAreas(hasAShipper) {
			for i, shipper := range a.Shippers {
				fid := toFlowID(a.ID, shipper.OwnerID, i, shipInput, shipper.Province().Int())
				fids = append(fids, fid)
			}
		}
		fids = append(fids, sourceFID)
	case area.IsSea():
		for _, a := range area.adjacentAreas() {
			switch {
			case a.hasCity() && from.IO == shipOutput:
				fids = append(fids, toFlowID(a.ID))
			case a.hasProducer() && from.IO == shipInput && c.ZoneFor(a) != nil:
				fids = append(fids, toFlowID(c.ZoneFor(a).AreaIDS[0]))
			case a.IsSea() && from.IO == shipOutput:
				for i, shipper := range a.Shippers {
					if from.Province == shipper.Province() {
						fid := toFlowID(a.ID, from.PID, i, shipInput, from.Province.Int())
						if !fids.include(fid) {
							fids = append(fids, fid)
						}
					}
				}
				fid := toFlowID(from.AreaID, from.PID, from.Index, shipInput, from.Province.Int())
				if !fids.include(fid) {
					fids = append(fids, fid)
				}
			case a.IsSea() && from.IO == shipInput:
				for i, shipper := range a.Shippers {
					if from.Province == shipper.Province() {
						fid := toFlowID(a.ID, from.PID, i, shipOutput, from.Province.Int())
						if !fids.include(fid) {
							fids = append(fids, fid)
						}
					}
				}
				fid := toFlowID(from.AreaID, from.PID, from.Index, shipOutput, from.Province.Int())
				if !fids.include(fid) {
					fids = append(fids, fid)
				}
			}
		}
	}
	return fids
}

func (g *Game) ProposedShips(fm flowMatrix) ShipperIncomeMap {
	ships := make(ShipperIncomeMap, 0)
	for from, sf := range fm {
		if g.isSeaID(from.AreaID) && from.IO == shipInput {
			for to, v := range sf {
				if v > 0 && to.IO == shipOutput {
					ships[from.PID] += v
				}
			}
		}
	}
	return ships
}

func (g *Game) usedOtherShips(p *Player, fm flowMatrix) int {
	count := 0
	for pid, v := range g.ProposedShips(fm) {
		if pid != p.ID() {
			count += v
		}
	}
	return count
}

func (g *Game) ProposedCities() map[AreaID]int {
	cities := make(map[AreaID]int, 0)
	for from, v := range g.ProposedPath[targetFID] {
		cities[from.AreaID] = v * -1
	}
	return cities
}

func (fm flowMatrix) usesOtherPlayerShips(p *Player) bool {
	g := p.Game()
	for _, sf := range fm {
		for from, v := range sf {
			if v > 0 && g.isSeaID(from.AreaID) && from.PID != p.ID() {
				return true
			}
		}
	}
	return false
}
