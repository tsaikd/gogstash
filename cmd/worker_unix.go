//go:build !windows
// +build !windows

package cmd

import (
	"context"
	"os"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/tsaikd/gogstash/config/goglog"
)

func startWorker(args []string, attr *syscall.ProcAttr) (pid int, err error) {
	pid, err = syscall.ForkExec(args[0], args, attr)
	if err != nil {
		goglog.Logger.Errorf("start worker error: %v", err)
		return
	}
	goglog.Logger.Infof("worker started: %d", pid)
	return
}

func waitWorkers(ctx context.Context, pids []int, args []string, attr *syscall.ProcAttr) error {
	var ws syscall.WaitStatus
	for {
		// wait for any child process
		pid, err := syscall.Wait4(-1, &ws, 0, nil)
		if err != nil {
			goglog.Logger.Errorf("wait4() error: %v", err)
			continue
		}
		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}
		for i, p := range pids {
			// match our worker's pid
			if p == pid {
				goglog.Logger.Warnf("worker %d stopped unexpectedly (wstatus: %d)", pid, ws)
				// only restart once after stopped unexpectedly
				pid, _ = startWorker(args, attr)
				pids[i] = pid
				break
			}
		}
	}
}

func startWorkers(ctx context.Context, workerNum int) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	attr := &syscall.ProcAttr{
		Env:   os.Environ(),
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
	}
	args := append([]string{os.Args[0], WorkerModule.Use}, os.Args[1:]...)

	pids := make([]int, workerNum)
	for i := 0; i < workerNum; i++ {
		pid, err := startWorker(args, attr)
		if err != nil {
			return err
		}
		pids[i] = pid
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return waitWorkers(ctx, pids, args, attr)
	})

	signal := waitSignals(ctx)
	if signal != nil {
		for _, pid := range pids {
			p, err := os.FindProcess(pid)
			if err == nil {
				_ = p.Signal(signal)
			}
		}
	}

	return nil
}
