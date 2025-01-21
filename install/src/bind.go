package main

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"regexp"
)

type Bind9 struct {
	f *os.File
}

func (bind *Bind9) open() {
	var err error
	bind.f, err = os.OpenFile("/usr/local/etc/bind/zones/db.worldl.xpt", os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func (bind *Bind9) close() {
	err := bind.f.Close()
	if err != nil {
		return
	}
}

func (bind *Bind9) findBindZone(key string) bool {
	scanner := bufio.NewScanner(bind.f)
	r, err := regexp.Compile(key)
	if err != nil {
		log.Fatal(err)
	}

	for scanner.Scan() {
		if r.MatchString(scanner.Text()) {
			return true
		}
	}

	return false
}

func (bind *Bind9) updateK8sIngress(ip string) {
	bind.open()
	defer bind.close()
	found := bind.findBindZone("\\$INCLUDE /usr/local/etc/bind/zones/ingress-k8s.worldl.xpt")
	if !found {
		if _, err := bind.f.WriteString("$INCLUDE /usr/local/etc/bind/zones/ingress-k8s.worldl.xpt\n"); err != nil {
			log.Fatal(err)
		}
	}

	f, err := os.OpenFile("/usr/local/etc/bind/zones/ingress-k8s.worldl.xpt", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.WriteString("*.worldl.xpt. IN A " + ip + "\n")
	if err != nil {
		log.Fatal(err)
	}

	err = f.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func (bind *Bind9) updateResolver(subDomain string, ip string) {
	bind.open()
	defer bind.close()
	subDomainFile := "/usr/local/etc/bind/zones/" + subDomain + ".worldl.xpt"
	found := bind.findBindZone("\\$INCLUDE " + subDomainFile)
	if !found {
		if _, err := bind.f.WriteString("$INCLUDE " + subDomainFile + "\n"); err != nil {
			log.Fatal(err)
		}
	}

	f, err := os.OpenFile(subDomainFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.WriteString("*." + subDomain + ".worldl.xpt. IN A " + ip + "\n")
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.WriteString(subDomain + ".worldl.xpt. IN A " + ip + "\n")
	if err != nil {
		log.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func (bind *Bind9) restartBind() {
	log.Println("Restarting BIND...")
	_, err := exec.Command("sudo", "brew", "services", "restart", "bind").Output()
	//log.Println(string(out))
	if err != nil {
		log.Fatal(err)
	}
}
