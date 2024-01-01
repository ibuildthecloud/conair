package cli

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/acorn-io/cmd"
	"github.com/cosmtrek/air/runner"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	return cmd.Command(&ConAir{})
}

type ConAir struct {
	AirConfig string `usage:"air config to override defaults"`
}

func (c *ConAir) Run(cmd *cobra.Command, args []string) error {
	cfg, err := runner.InitConfig(c.AirConfig)
	if err != nil {
		return err
	}

	if _, err := exec.LookPath("dlv"); errors.Is(err, fs.ErrNotExist) {
		cmd := exec.Command("go", "install", "github.com/go-delve/delve/cmd/dlv@latest")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	if err := setDefaults(cfg, args); err != nil {
		return err
	}

	r, err := runner.NewEngineWithConfig(cfg, false)
	if err != nil {
		return err
	}
	context.AfterFunc(cmd.Context(), r.Stop)

	r.Run()
	return nil
}

func setSliceIfEmpty[T any](currentValue, newValue []T) []T {
	if len(currentValue) > 0 {
		return currentValue
	}
	return newValue
}

func setIfEmpty[T comparable](currentValue, newValue T) T {
	var unset T
	if currentValue == unset {
		return newValue
	}
	return currentValue
}

func setDefaults(cfg *runner.Config, args []string) error {
	tmp, err := os.MkdirTemp("", "air-run")
	if err != nil {
		return err
	}

	bin := filepath.Join(tmp, "main")
	cmd := fmt.Sprintf("go build -o %s .", bin)
	log := filepath.Join(tmp, "build-errors.log")
	fullBin := fmt.Sprintf("dlv exec --continue --accept-multiclient --listen=:2345 --headless=true --api-version=2 --log %s --", bin)

	cfg.Build.ArgsBin = setSliceIfEmpty(cfg.Build.ArgsBin, args)
	cfg.Build.Bin = setIfEmpty(cfg.Build.Bin, bin)
	cfg.Build.Cmd = setIfEmpty(cfg.Build.Cmd, cmd)
	cfg.Build.Delay = setIfEmpty(cfg.Build.Delay, 1000)
	cfg.Build.ExcludeDir = setSliceIfEmpty(cfg.Build.ExcludeDir, []string{"assets", "tmp", "testdata"})
	cfg.Build.ExcludeRegex = setSliceIfEmpty(cfg.Build.ExcludeDir, []string{"_test.go"})
	cfg.Build.FullBin = setIfEmpty(cfg.Build.FullBin, fullBin)
	cfg.Build.KillDelay = setIfEmpty(cfg.Build.KillDelay, time.Second)
	cfg.Build.Log = setIfEmpty(cfg.Build.Log, log)
	cfg.Build.SendInterrupt = true
	cfg.Build.StopOnError = true
	cfg.Build.Rerun = true
	cfg.Build.RerunDelay = setIfEmpty(cfg.Build.RerunDelay, 2000)

	return nil
}
