package common

import (
	"errors"
)

type OrderedMods struct {
	Order []ModEntry
	Paths map[string]string
}

func (om *OrderedMods) AddToTop(m ModEntry) {
	_, ok := om.Paths[m.Pair.Meta.Domain]
	if ok {
		for i, v := range om.Order {
			if v.Pair.Meta.Domain == m.Pair.Meta.Domain {
				om.Order = append(om.Order[:i], om.Order[i+1:]...)
				break
			}
		}
	}
	om.Order = append(om.Order, m)
	om.Paths[m.Pair.Meta.Domain] = m.Pair.Path
}

func (om *OrderedMods) Move(d string, to int) bool {
	if to < 0 || to >= len(om.Order) {
		return false
	}
	for i, v := range om.Order {
		if v.Pair.Meta.Domain == d {
			if i == to {
				return true
			}
			if i < to {
				copy(om.Order[i:to], om.Order[i+1:to+1])
			} else {
				copy(om.Order[to+1:i+1], om.Order[to:i])
			}
			om.Order[to] = v
			return true
		}
	}
	return false
}

func (om *OrderedMods) Remove(d string) bool {
	_, ok := om.Paths[d]
	if ok {
		for i, v := range om.Order {
			if v.Pair.Meta.Domain == d {
				om.Order = append(om.Order[:i], om.Order[i+1:]...)
				return true
			}
		}
	}
	return false
}

var ErrMissing = errors.New("no mods given for this specific module")
