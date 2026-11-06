package main

import (
	"icarus/app/cli"
	"icarus/common"
	"os"
)

func main() {
	cli.Main(os.Args, []common.Module{})
}
