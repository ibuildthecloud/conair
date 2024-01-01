package cli

import (
	"github.com/acorn-io/cmd"
	"github.com/ibuildthecloud/conair/pkg/engine"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	return cmd.Command(&ConAir{})
}

type ConAir struct {
	engine.Options
}

func (c *ConAir) Run(cmd *cobra.Command, args []string) error {
	return engine.Run(cmd.Context(), args, c.Options)
}
