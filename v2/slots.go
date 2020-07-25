package indonesia

type Slot struct {
	Developed bool
	Company   *Company
}

type Slots []*Slot

func (s *Slot) Empty() bool {
	return s.Developed && s.Company == nil
}

func (p *Player) hasEmptySlot() bool {
	s, _ := p.getEmptySlot()
	return s != nil
}

func (p *Player) getEmptySlot() (*Slot, int) {
	for i, s := range p.Slots {
		if s.Empty() {
			return s, i + 1
		}
	}
	return nil, NoSlot
}
