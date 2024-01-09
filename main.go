package main

import (
	"github.com/acorn-io/cmd"
	"github.com/ibuildthecloud/conair/pkg/cli"
	"github.com/ibuildthecloud/conair/pkg/reaper"
)

func main() {
	reaper.Start()
	cmd.Main(cli.New())
}
