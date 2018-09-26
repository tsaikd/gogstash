// +build windows

package cmd

import (
	"os"
	"syscall"
)

func startWorker() (int, error) {
	execSpec := &syscall.ProcAttr{
		Env:   os.Environ(),
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
	}
	args := append(os.Args, "--follower")
	pid, handle, err := syscall.StartProcess(os.Args[0], args, execSpec)
	if err == nil {
		// succeed
		syscall.CloseHandle(syscall.Handle(handle))
	}
	return pid, err
}
