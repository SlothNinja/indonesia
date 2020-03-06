package indonesia

type Comb struct {
	N       int
	M       int
	Set     []int
	Current []int
}

func NewComb(n, m int) *Comb {
        if n < 0 || m < 0 || m > n {
                return nil
        }
	s := make([]int, n)
	for i := range s {
		s[i] = i
	}
	c := &Comb{
		N:       n,
		M:       m,
		Set:     s,
		Current: s[:m],
	}
	return c
}

func (c *Comb) Next() *Comb {
	next := &Comb{
		N:       c.N,
		M:       c.M,
		Set:     make([]int, len(c.Set)),
		Current: make([]int, len(c.Current)),
	}
	copy(next.Set, c.Set)
	copy(next.Current, c.Current)
	for i := len(next.Current) - 1; i >= 0; i-- {
		switch {
		case next.Current[i]+1 <= next.N-next.M+i:
			next.Current[i] += 1
			return next
		default:
			for j := i - 1; j >= 0; j-- {
				if v := next.Current[j] + 1; v <= next.N-next.M+j {
					for k := j; k < len(c.Current); k++ {
						next.Current[k] = v
						v += 1
					}
					return next
				}
			}
		}
	}
	c.Current = []int{}
	return c
}
