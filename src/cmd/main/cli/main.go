package main

import (
	"icarus/app/cli"
	"icarus/common"
	basic "icarus/modules/basic/module"
	"os"
)

func main() {
	cli.Main(os.Args, []common.Module{basic.Module{}})
}
