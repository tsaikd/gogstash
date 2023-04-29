package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/shirou/gopsutil/v3/process"
)

const randomPostfixLength = 4
const executableNamePrefix = "gogstash_test"
const terminationDuration = 100 * time.Millisecond

func TestProgramTermination(t *testing.T) {
	// Start the main program in a separate process
	randBytes := make([]byte, randomPostfixLength)
	if _, err := rand.Read(randBytes); err != nil {
		t.Fatal(err)
	}
	executablePath := fmt.Sprintf("./%s_%x", executableNamePrefix, randBytes)

	if _, err := exec.Command("go", "build", "-o", executablePath).CombinedOutput(); err != nil {
		t.Fatalf("Failed to build the program: %v", err)
	}
	defer func() {
		os.Remove(executablePath)
	}()

	cmd := exec.Command(executablePath, "--config", "./testdata/config.yaml")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start the program: %v", err)
	}

	// Get the process ID of the main process
	mainPID := int32(cmd.Process.Pid)

	// Find all child processes of the main process
	childProcesses, err := waitChildProcesses(mainPID, 1*time.Second)
	t.Logf("Children are %s", strings.Join(lo.Map(childProcesses, func(v int32, _ int) string { return strconv.Itoa(int(v)) }), ", "))
	if err != nil {
		t.Fatalf("Failed to get child processes: %v", err)
	}
	if len(childProcesses) < 1 {
		t.Fatal("Cannot test if there are no child processes")
	}
	// Cleanup created child processes
	defer func() {
		// Recalling `Processes` to get a fresh information from the OS. Important to force `process.NewProcess` to return valid information.
		_, err = process.Processes()
		if err != nil {
			t.Fatalf("Failed to update processes information: %v", err)
		}

		for _, pid := range childProcesses {
			cp, err := process.NewProcess(pid)
			if err == nil {
				err = cp.Kill()
				if err != nil {
					t.Fatalf("Cannot terminate process with PID %d", pid)
				}
			}
		}
	}()

	// Send a SIGINT signal (Ctrl+C) to the main process
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("Failed to send SIGINT signal: %v", err)
	}

	// Wait for a reasonable time for the processes to terminate
	time.Sleep(terminationDuration)

	// Check whether all child processes have terminated.
	for _, pid := range childProcesses {
		exist, err := process.PidExists(pid)
		if err != nil {
			t.Fatal(err)
		}
		if exist {
			t.Errorf("Process with PID %d did not terminate.", pid)
		}
	}
}

func getChildProcesses(parentPID int32) ([]int32, error) {
	var childPIDs []int32
	allProcesses, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("Failed to get all processes: %v", err)
	}
	for _, p := range allProcesses {
		ppid, perr := p.Ppid()
		if perr != nil {
			// Process without ppid is irrelevant
			continue
		}
		if ppid == parentPID {
			childPIDs = append(childPIDs, p.Pid)
		}
	}
	return childPIDs, nil
}

func waitChildProcesses(parentPID int32, timeout time.Duration) ([]int32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	interval := 50 * time.Millisecond
	timer := time.NewTimer(interval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		case <-timer.C:
			childPIDs, err := getChildProcesses(parentPID)
			if err != nil {
				return nil, err
			}
			if len(childPIDs) > 0 {
				return childPIDs, nil
			}
			timer.Reset(interval)
		}
	}
}
