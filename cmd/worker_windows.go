// +build windows

package cmd

import (
	"context"
	"os"
	"syscall"
	"unsafe"

	"github.com/tsaikd/gogstash/config/goglog"
	"golang.org/x/sync/errgroup"
)

func startWorker(args []string, attr *syscall.ProcAttr) (pid int, handle uintptr, err error) {
	pid, handle, err = syscall.StartProcess(os.Args[0], args, attr)
	if err != nil {
		goglog.Logger.Errorf("start worker failed: %v", err)
		return
	}
	goglog.Logger.Infof("worker started: %d", pid)
	return
}

func waitWorkers(ctx context.Context, pids []int, handles []uintptr, args []string, attr *syscall.ProcAttr) error {
	// syscall only has `WaitForSingleObject`, but we have to wait multiple processes,
	// so that we find proc `WaitForMultipleObjects` from kernel32.dll.
	// doc: https://docs.microsoft.com/en-us/windows/desktop/api/synchapi/nf-synchapi-waitformultipleobjects
	dll := syscall.MustLoadDLL("kernel32.dll")
	wfmo := dll.MustFindProc("WaitForMultipleObjects")
	for {
		r1, _, err := wfmo.Call(uintptr(len(handles)), uintptr(unsafe.Pointer(&handles[0])), 0, syscall.INFINITE)
		ret := int(r1)
		if ret == syscall.WAIT_FAILED && err != nil {
			goglog.Logger.Errorf("WaitForMultipleObjects() error: %v", err)
			continue
		}
		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}
		if ret >= syscall.WAIT_OBJECT_0 && ret < syscall.WAIT_OBJECT_0+len(handles) {
			i := ret - syscall.WAIT_OBJECT_0
			syscall.CloseHandle(syscall.Handle(handles[i]))
			goglog.Logger.Warnf("worker %d stopped unexpectedly", pids[i])
			// only restart once after stopped unexpectedly
			pid, handle, _ := startWorker(args, attr)
			pids[i] = pid
			handles[i] = handle
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
	handles := make([]uintptr, workerNum)
	for i := 0; i < workerNum; i++ {
		pid, handle, err := startWorker(args, attr)
		if err != nil {
			return err
		}
		pids[i] = pid
		handles[i] = handle
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return waitWorkers(ctx, pids, handles, args, attr)
	})
	return waitSignals(ctx)
}
