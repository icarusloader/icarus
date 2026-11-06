package core

import "icarus/common"

func Run(ms []common.Module, OrderedMods *common.OrderedMods, rd common.RunData) (bool, error) {
	entries := make(map[string][]common.ModEntry, len(ms))
	for _, m := range ms {
		meta := m.Meta()
		entries[meta.Name] = []common.ModEntry{}
	}
	for _, v := range OrderedMods.Order {
		if v.Enabled {
			for _, mn := range v.Pair.Meta.Modules {
				_, ok := entries[mn]
				if ok {
					entries[mn] = append(entries[mn], v)
				}
			}
		}
	}
	block := false
	for _, m := range ms {
		name := m.Meta().Name
		b, err := m.Run(entries[name], rd)
		if err != nil {
			rd.Hooks.WhenError(name, err)
			return block, err
		}
		block = block || b
	}
	return block, nil
}
