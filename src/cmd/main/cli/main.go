package main

import (
	"os"

	"github.com/icarusloader/icarus/app/cli"
	"github.com/icarusloader/icarus/common"
	basic "github.com/icarusloader/icarus/modules/basic/module"
)

func main() {
	cli.Main(os.Args, []common.Module{basic.Module{}})
}
