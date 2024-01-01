package engine

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cosmtrek/air/runner"
)

type Options struct {
	AirConfig string `usage:"air config to override defaults" short:"c"`
}

func Run(ctx context.Context, args []string, opt Options) error {
	cfg, err := runner.InitConfig(opt.AirConfig)
	if err != nil {
		return err
	}

	if _, err := exec.LookPath("dlv"); errors.Is(err, exec.ErrNotFound) {
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
	context.AfterFunc(ctx, r.Stop)

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
	defer os.RemoveAll(tmp)

	log.Printf("Using temp dir: %s", tmp)

	var (
		targets []string
		argsBin []string
	)

	for i, arg := range args {
		if arg == "--" {
			argsBin = args[i+1:]
			break
		}
		if _, err := os.Stat(arg); err == nil {
			targets = append(targets, arg)
		} else {
			argsBin = args[i:]
			break
		}
	}

	if len(targets) == 0 {
		targets = []string{"."}
	}

	bin := filepath.Join(tmp, "main")
	cmd := fmt.Sprintf("go build -o %s %s", bin, strings.Join(targets, " "))
	log := filepath.Join(tmp, "build-errors.log")
	fullBin := fmt.Sprintf("dlv exec --continue --accept-multiclient --listen=:2345 --headless=true --api-version=2 --log %s --", bin)

	cfg.Build.ArgsBin = setSliceIfEmpty(cfg.Build.ArgsBin, argsBin)
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
	cfg.Build.RerunDelay = setIfEmpty(cfg.Build.RerunDelay, 5000)

	return nil
}
