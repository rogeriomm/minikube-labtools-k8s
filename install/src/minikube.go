package main

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os/exec"
)

func minikubeRun(cmd string) {
	fmt.Println(cmd)
}

func minikubeSsh(node string, params string) error {
	out, err := exec.Command("minikube", "--node="+node, "ssh", params).Output()
	if err != nil {
		err = errors.Errorf("Minikube ssh %s %s %v", node, params, err)
		log.Println(err)
		return err
	}
	if len(out) == 0 {
		return nil
	}
	fmt.Println(string(out))

	return nil
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

func minikubeAddIpRoute(cluster string, subnet string) {
	ip := minikubeGetIp(cluster)
	addIpRoute(subnet, ip)
}
