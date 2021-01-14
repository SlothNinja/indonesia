package indonesia

import (
	"html/template"

	"github.com/SlothNinja/color"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/user"
)

type Shipper struct {
	g         *Game
	a         *Area
	OwnerID   int
	Slot      int
	ShipType  ShipType
	Delivered int
}

type Shippers []*Shipper

func (ss Shippers) includeCompany(c *Company) bool {
	for _, s := range ss {
		if s.Company() == c {
			return true
		}
	}
	return false
}

func (ss Shippers) include(shipper *Shipper) bool {
	for _, s := range ss {
		if s.equals(shipper) {
			return true
		}
	}
	return false
}

func (s *Shipper) init(g *Game, a *Area) {
	s.g = g
	s.a = a
}

// func (s *Shipper) Color() color.Color {
// 	if owner := s.Owner(); owner == nil {
// 		return color.Black
// 	} else {
// 		return owner.Color()
// 	}
// }

func (g *Game) ShipColor(s *Shipper, cu *user.User) color.Color {
	owner := s.Owner()
	if owner == nil {
		return color.Black
	}
	return g.Color(owner, cu)
}

func (s *Shipper) equals(shipper *Shipper) bool {
	return s.OwnerID == shipper.OwnerID && s.Slot == shipper.Slot
}

func (s *Shipper) copy() *Shipper {
	copied := *s
	return &copied
}

func (ss Shippers) copy() Shippers {
	var copied Shippers
	copy(copied, ss)
	for i, s := range copied {
		copied[i] = s.copy()
	}
	return copied
}

func (s *Shipper) Owner() *Player {
	return s.g.PlayerByID(s.OwnerID)
}

func (s *Shipper) Company() *Company {
	owner := s.Owner()
	if owner == nil {
		return nil
	}
	if company := owner.Slots[s.Slot-1].Company; company == nil || company.Goods() != Shipping {
		return nil
	} else {
		return company
	}
}

func (s *Shipper) HullSize() int {
	if company := s.Company(); company == nil {
		return 0
	} else {
		return company.HullSize()
	}
}

func (s *Shipper) ShipTip() template.HTML {
	return restful.HTML("{%q:%q, %q:\"%d\", %q:\"%d\", %q:\"%d\"}",
		"owner", s.g.NameByPID(s.OwnerID),
		"slot", s.Slot,
		"hull", s.HullSize(),
		"delivered", s.Delivered)
}

func (g *Game) SelectedShipper() *Shipper {
	if index, area := g.SelectedShipperIndex, g.SelectedArea2(); area == nil {
		return nil
	} else {
		return area.Shippers[index]
	}
}

func (s *Shipper) shipsInArea(a *Area) int {
	ships := 0
	if a.IsLand() {
		return ships
	}
	for _, shipper := range a.Shippers {
		if s.equals(shipper) {
			ships += 1
		}
	}
	return ships
}

func (s *Shipper) Province() Province {
	return s.Company().Province()
}
