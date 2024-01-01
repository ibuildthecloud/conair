package main

import (
	"github.com/acorn-io/cmd"
	"github.com/ibuildthecloud/conair/pkg/cli"
	"github.com/ramr/go-reaper"
)

func main() {
	go reaper.Reap()
	cmd.Main(cli.New())
}
