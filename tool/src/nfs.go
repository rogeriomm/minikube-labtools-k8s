package main

import (
	"os/exec"
	"os/user"
	"runtime"
	"strings"
)

type nfs struct {
}

func (m *nfs) reconfigure() {
	if runtime.GOOS == "darwin" {
		usr, _ := user.Current()
		if usr.Uid != "0" {
			sugar.Error("Must run as root")
			return
		}

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
