package main

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type nfs struct {
}

func (m *nfs) reconfigure() {
	if runtime.GOOS != "darwin" {
		return
	}

	if os.Geteuid() != 0 {
		exePath, err := os.Executable()
		if err != nil {
			return
		}

		out, err := exec.Command("sudo", "-E", exePath, "reconfigure-nfs").CombinedOutput()
		if err != nil {
			sugar.Error("Error ", out)
			return
		}
		//sugar.Info(string(out))
	} else {
		running, param, err := CheckProcessRunning("nfsd")
		if err != nil {
			sugar.Error(err)
			return
		}

		if running && !strings.Contains(param, "-N") {
			sugar.Info("Restarting NFSD")

			out, err := exec.Command("nfsd", "stop").CombinedOutput()

			if err != nil {
				sugar.Error(err, string(out))
				return
			}
			running = false
		}

		if !running {
			cmd := exec.Command("nfsd", "-N")

			err = cmd.Start()

			if err != nil {
				sugar.Error(err)
			}
		}
	}
}
