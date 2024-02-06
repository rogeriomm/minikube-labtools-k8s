package main

import (
	"fmt"
	"github.com/bitfield/script"
	"github.com/inhies/go-bytesize"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type minikube struct {
	profile string
	verbose bool
}

func (m *minikube) addonEnable(addons []string, enable bool) {
	for i := 0; i < len(addons); i++ {
		if enable {
			script.Exec("minikube -p " + m.profile + " addons enable " + addons[i]).Stdout()
		} else {
			script.Exec("minikube -p " + m.profile + " addons disable " + addons[i]).Stdout()
		}

	}
}

func (m *minikube) ssh(node string, params string) error {
	out, err := exec.Command("minikube", "-p", m.profile, "--node="+node, "ssh", params).CombinedOutput()
	if err != nil {
		err = errors.Errorf("Minikube ssh %s %s %v", node, params, err)
		sugar.Error(err, string(out))
	}

	return err
}

func (m *minikube) loginSsh(node int) {
	script.Exec("minikube -p " + m.profile + " ssh").Stdout()
}

func (m *minikube) setProfile(name string) {
	sugar.Debug("Minikube: set profile " + name)
	m.profile = name
}

func (m *minikube) applyProfile() {
	sugar.Debug("Minikube apply profile " + m.profile)

	_, err := exec.Command("minikube", "profile", m.profile).Output()
	if err != nil {
		sugar.Fatal(err)
	}
}

func (m *minikube) start() error {
	sugar.Info("Minikube " + m.profile + ": start")

	proc := exec.Command("minikube", "-p", m.profile, "start")

	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	proc.Stdin = os.Stdin

	err := proc.Run()

	if err == nil {
		ip := m.getIp()

		_, err = exec.Command("sudo", "ifconfig",
			"en0", "alias", ip, "255.255.255.0").Output()
		if err == nil {
			_, err = exec.Command("sudo", "pfctl", "-e").Output()
		}
	}

	return err
}

func (m *minikube) stop() error {
	sugar.Info("Minikube " + m.profile + ": stop")

	proc := exec.Command("minikube", "-p", m.profile, "stop")

	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	proc.Stdin = os.Stdin

	return proc.Run()
}

func (m *minikube) delete() error {
	sugar.Info("Minikube " + m.profile + ": delete")

	proc := exec.Command("minikube", "-p", m.profile, "delete")

	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	proc.Stdin = os.Stdin

	return proc.Run()
}

func (m *minikube) getIp() string {
	out, err := exec.Command("minikube", "-p", m.profile, "ip").Output()
	if err != nil {
		sugar.Error(err)
	}
	ip := strings.Replace(string(out), "\n", "", -1)
	return ip
}

func (m *minikube) addIpRoute(subnet string, gateway string) {
	var out []byte
	var err error

	if runtime.GOOS == "darwin" {
		out, err = exec.Command("sudo", "route", "-n", "delete", subnet).Output()
	} else {
		out, err = exec.Command("sudo", "route", "delete", "-net", subnet).Output()
	}

	if err != nil {
		//sugar.Error(err)
	}

	if runtime.GOOS == "darwin" {
		out, err = exec.Command("sudo", "route", "-n", "add", subnet, gateway).Output()
	} else {
		out, err = exec.Command("sudo", "route", "add", "-net", subnet, "gw", gateway).Output()
	}

	if err != nil {
		sugar.Error(out, err)
	}
	sugar.Debug(string(out))
}

func (m *minikube) minikubeAddIpRoute(subnet string) {
	ip := m.getIp()
	m.addIpRoute(subnet, ip)
}

func (m *minikube) showPlugins() error {
	out, err := exec.Command("minikube", "-p", m.profile, "addons", "list").Output()
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func (m *minikube) configSize(key string, value float64) error {
	_, err := exec.Command("minikube", "-p", m.profile, "config", "set",
		key, bytesize.New(value).Format("%.0f ", "", false)).Output()

	if err != nil {
		return err
	}
	return nil
}

func (m *minikube) config(key string, value string) error {
	_, err := exec.Command("minikube", "-p", m.profile, "config", "set", key, value).Output()
	if err != nil {
		return err
	}
	return nil
}

func (m *minikube) setDockerEnv() error {
	out, err := exec.Command("minikube", "-p", m.profile, "docker-env").Output()
	if err != nil {
		return err
	}

	err = os.WriteFile(minikubeHomePath+"/docker-env", out, 0644)

	if err != nil {
		sugar.Debug(err)
	}

	return err
}
