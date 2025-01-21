package main

import (
	"bufio"
	"github.com/bitfield/script"
	"os"
	"regexp"
	"runtime"
)

type bind9 struct {
	configPath  string
	labBindPath string
	f           *os.File
}

func (bind *bind9) open() {
	var err error

	if runtime.GOOS == "linux" {
		bind.configPath = "/etc/bind"
	} else if runtime.GOOS == "darwin" {
		bind.configPath = "/usr/local/etc/bind"
	}

	if script.IfExists(bind.configPath).Error() != nil {
		sugar.Fatal("BIND9 not installed")
	}

	bind.labBindPath = minikubeK8sPath + "/install/scripts/bind/" + runtime.GOOS + "/docker/"

	if script.IfExists(bind.configPath+"/zones").Error() != nil {
		script.Exec("sudo chown -R mac /etc/bind").Stdout()
		os.MkdirAll(bind.configPath+"/zones", 0777)
	}

	script.Exec("zsh -c 'cp " + bind.labBindPath + "/zones/* " + bind.configPath + "/zones/'").Stdout()

	if runtime.GOOS == "linux" {
		script.Exec("cp " + bind.labBindPath + "/named.conf.local " + bind.configPath).Stdout()
		script.Exec("cp " + bind.labBindPath + "/named.conf.options " + bind.configPath).Stdout()
	} else if runtime.GOOS == "darwin" {
		script.Exec("cp " + bind.labBindPath + "/named.conf.local " + bind.configPath).Stdout()
		script.Exec("cp " + bind.labBindPath + "/named.conf " + bind.configPath).Stdout()
	}

	// Validating Syntax of bind configuration and Zone files
	script.Exec("named-checkconf " + bind.configPath + "/named.conf.local").Stdout()
	script.Exec("named-checkconf " + bind.configPath + "/named.conf").Stdout()

	script.Exec("named-checkzone xxx " + bind.configPath + "/zones/db.worldl.xpt").Stdout()

	// Open configuration file
	bind.f, err = os.OpenFile(bind.configPath+"/zones/db.worldl.xpt", os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		sugar.Fatal(err)
	}
}

func (bind *bind9) close() {
	err := bind.f.Close()
	if err != nil {
		return
	}
}

func (bind *bind9) findBindZone(key string) bool {
	scanner := bufio.NewScanner(bind.f)
	r, err := regexp.Compile(key)
	if err != nil {
		sugar.Fatal(err)
	}

	for scanner.Scan() {
		if r.MatchString(scanner.Text()) {
			return true
		}
	}

	return false
}

func (bind *bind9) updateK8sIngress(ip string) {
	bind.open()
	defer bind.close()
	found := bind.findBindZone("\\$INCLUDE " + bind.configPath + "/zones/ingress-k8s.worldl.xpt")
	if !found {
		if _, err := bind.f.WriteString("$INCLUDE " + bind.configPath + "/zones/ingress-k8s.worldl.xpt\n"); err != nil {
			sugar.Fatal(err)
		}
	}

	f, err := os.OpenFile(bind.configPath+"/zones/ingress-k8s.worldl.xpt", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		sugar.Fatal(err)
	}
	_, err = f.WriteString("*.worldl.xpt. IN A " + ip + "\n")
	if err != nil {
		sugar.Fatal(err)
	}

	err = f.Close()
	if err != nil {
		sugar.Fatal(err)
	}
}

func (bind *bind9) updateResolver(subDomain string, ip string) {
	bind.open()
	defer bind.close()
	subDomainFile := bind.configPath + "/zones/" + subDomain + ".worldl.xpt"
	found := bind.findBindZone("\\$INCLUDE " + subDomainFile)
	if !found {
		if _, err := bind.f.WriteString("$INCLUDE " + subDomainFile + "\n"); err != nil {
			sugar.Fatal(err)
		}
	}

	f, err := os.OpenFile(subDomainFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		sugar.Fatal(err)
	}
	_, err = f.WriteString("*." + subDomain + ".worldl.xpt. IN A " + ip + "\n")
	if err != nil {
		sugar.Fatal(err)
	}
	_, err = f.WriteString(subDomain + ".worldl.xpt. IN A " + ip + "\n")
	if err != nil {
		sugar.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		sugar.Fatal(err)
	}
}

func (bind *bind9) restartBind() {
	sugar.Info("Restarting BIND...")

	if runtime.GOOS == "linux" {
		script.Exec("sudo systemctl restart bind9").Stdout()
	} else if runtime.GOOS == "darwin" {
		script.Exec("sudo brew services restart bind").Stdout()
	}
}
