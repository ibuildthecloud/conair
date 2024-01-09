package reaper

import (
	"errors"
	"io/fs"
	"syscall"
	"time"

	"github.com/prometheus/procfs"
	"github.com/ramr/go-reaper"
	"github.com/sirupsen/logrus"
)

func Start() {
	go reaper.Reap()

	go func() {
		for {
			if err := killDLV(); err != nil && !errors.Is(err, fs.ErrNotExist) {
				logrus.Errorf("failed to find and terminate stale dlv: %v", err)
			}
			time.Sleep(1 * time.Second)
		}
	}()
}

func killDLV() error {
	procs, err := procfs.AllProcs()
	if err != nil {
		return err
	}

	var dlvPid int
	for _, proc := range procs {
		cl, err := proc.CmdLine()
		if err != nil {
			return err
		}
		if len(cl) > 2 && cl[0] == "dlv" && cl[1] == "exec" {
			dlvPid = proc.PID
			break
		}
	}

	if dlvPid == 0 {
		return nil
	}

	for i := 0; i < 10; i++ {
		procs, err = procfs.AllProcs()
		if err != nil {
			return err
		}

		for _, proc := range procs {
			ps, err := proc.Stat()
			if err != nil {
				return err
			}

			if ps.PPID == dlvPid {
				return nil
			}
		}

		time.Sleep(time.Second)
	}

	logrus.Infof("Killing dlv: %d", dlvPid)
	return syscall.Kill(dlvPid, syscall.SIGTERM)
}
