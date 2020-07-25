package indonesia

type Producer struct {
	g       *Game
	OwnerID int
	Slot    int
	Goods   Goods
}

func (p *Producer) init(g *Game) {
	p.g = g
}

func (p *Producer) copy() *Producer {
	copied := *p
	return &copied
}

func (p *Producer) Owner() *Player {
	return p.g.PlayerByID(p.OwnerID)
}

func (p *Producer) Company() *Company {
	if owner := p.Owner(); owner == nil {
		return nil
	} else {
		if p.Slot < 1 || p.Slot > 5 {
			return nil
		}
		return owner.Slots[p.Slot-1].Company
	}
}
