package main

import (
	"fmt"
	"log"
	"os/exec"
)

func minikubeRun(cmd string) {
	fmt.Println(cmd)
}

func minikubeSsh(node string, parms string) {
	out, err := exec.Command("minikube", "--node="+node, "ssh", parms).Output()
	if err != nil {
		log.Println(err)
	}
	if len(out) == 0 {
		return
	}
	fmt.Println(string(out))
}

func minikubeSetProfile(name string) {
	_, err := exec.Command("minikube", "profile", name).Output()
	if err != nil {
		log.Fatal(err)
	}
}

func minikubeGetIp(name string) string {
	minikubeSetProfile(name)
	out, err := exec.Command("minikube", "ip").Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(out)
}

/*
netstat -anr -f inet
*/
func addIpRoute(subnet string, gateway string) {
	_, err := exec.Command("sudo", "route", "-n", "delete", subnet).Output()
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(string(out))

	_, err = exec.Command("sudo", "route", "-n", "add", subnet, gateway).Output()
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(string(out))
}

func minikubeAddIpRoute() {
	ip := minikubeGetIp("cluster")
	addIpRoute("10.112.0.0/12", ip)
	ip = minikubeGetIp("cluster2")
	addIpRoute("10.96.0.0/12", ip)
}
