package indonesia

import "github.com/SlothNinja/color"

type City struct {
	a         *Area
	Size      int
	Delivered []int
	Grew      bool
}
type Cities []*City

func (g *Game) Cities() Cities {
	var cities Cities
	for _, a := range g.landAreas() {
		if a.City != nil {
			cities = append(cities, a.City)
		}
	}
	return cities
}

func (g *Game) resetCities() {
	for _, c := range g.Cities() {
		c.Delivered = defaultDeliveredGoods()
		c.Grew = false
	}
}

func (c *City) id() int {
	if c.a != nil {
		return int(c.a.ID)
	}
	return -1
}

func (cs Cities) include(city *City) bool {
	for _, c := range cs {
		if c == city {
			return true
		}
	}
	return false
}

func (c *City) init(a *Area) {
	c.a = a
}

func (c *City) CanGrow() (r bool) {
	if c.Size == 3 || c.Grew {
		return
	}

	for i, b := range c.a.g.ProducedGoods() {
		if b && c.Delivered[i] != c.Size {
			return
		}
	}
	return true
}

func (c *City) Area() *Area {
	if c == nil {
		return nil
	}
	return c.a
}

func (c *City) Province() Province {
	if c == nil || c.a == nil {
		return NoProvince
	}
	return c.a.Province()
}

func (c *City) Color() color.Color {
	switch c.Size {
	case 1:
		return color.Green
	case 2:
		return color.Yellow
	case 3:
		return color.Red
	default:
		return color.None
	}
}

func (c *City) adjacentAreas(tests ...addAreaTest) Areas {
	return c.a.adjacentAreas(tests...)
}

func newCity(a *Area) *City {
	return &City{a: a, Size: 1, Delivered: defaultDeliveredGoods()}
}

func defaultDeliveredGoods() []int {
	return []int{0, 0, 0, 0, 0}
}
