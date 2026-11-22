// the core module code
package module

import (
	"errors"
	"icarus/common"
)

type Module struct{}

func (Module) Meta() common.ModuleMeta {
	return common.ModuleMeta{
		Name: "basic",
	}
}

func (Module) Run(entries []common.ModEntry, rd common.RunData) (bool, error) {
	block := false // NOT rd.Block
	len := len(entries)
	overlay, ok := rd.Config.Flags["basic"]["overlay"].(bool)
	if !ok {
		return block, errors.New(`Config.Flags["basic"]["overlay"] not found or invalid`)
	}
	dst, ok := rd.Config.Flags["basic"]["destination"].(string)
	if !ok {
		return block, errors.New(`Config.Flags["basic"]["destination"] not found or invalid`)
	}
	lt, ok := rd.Config.Flags["basic"]["link-type"].(string)
	if !ok {
		lt = "default"
	}
	if len <= 0 {
		return block, common.ErrMissing
	}
	m, err := build(entries)
	if err != nil {
		return block, err
	}
	if overlay {
		if err = applyOverlay(rd, m, dst, LinkType(lt)); err != nil {
			return block, err
		}
	} else {
		return block, errors.New(`raw mode isn't supported yet`)
	}
	return block, nil
}
