package main

import (
	"github.com/shirou/gopsutil/v3/process"
	"os/exec"
	"runtime"
)

func flushDnsCache() {
	sugar.Info("Flushing host DNS cache...")

	if runtime.GOOS == "linux" {

	} else if runtime.GOOS == "darwin" {
		_, err := exec.Command("dscacheutil", "-flushcache").Output()

		if err != nil {
			sugar.Error(err)
		}

		_, err = exec.Command("sudo", "killall", "-HUP", "mDNSResponder").Output()

		if err != nil {
			sugar.Error(err)
		}
	}
}

func sudoValidateUser() {
	_, err := exec.Command("sudo", "-v").Output()
	if err != nil {
		sugar.Error(err)
	}
}

func CheckProcessRunning(processName string) (bool, string, error) {
	// Get a list of all running processes
	processes, err := process.Processes()
	if err != nil {
		return false, "", err
	}

	// Iterate through all processes to find a match
	for _, p := range processes {
		pname, err := p.Name()
		if err != nil {
			continue
		}

		if pname == processName {
			cmdline, err := p.Cmdline()
			if err != nil {
				return true, "", nil
			}
			return true, cmdline, nil
		}
	}

	return false, "", nil
}
