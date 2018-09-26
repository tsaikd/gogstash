// +build !windows

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
	return syscall.ForkExec(os.Args[0], args, execSpec)
}
