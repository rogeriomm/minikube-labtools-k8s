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
		log.Fatal(err)
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

func minikubeGetMainIp() string {
	out, err := exec.Command("minikube", "ip").Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(out)
}

func minikubeAddIpRoute() {
	ip := minikubeGetMainIp()

	out, err := exec.Command("sudo", "route", "-n", "delete", "10.0.0.0/8", ip).Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))

	out, err = exec.Command("sudo", "route", "-n", "add", "10.0.0.0/8", ip).Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))
}
