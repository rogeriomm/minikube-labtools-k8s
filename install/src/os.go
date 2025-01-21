package main

import (
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
