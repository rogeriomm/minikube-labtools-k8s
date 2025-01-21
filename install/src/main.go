/*

 */
package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

func flushDnsCache() {
	log.Println("Flushing host DNS cache...")
	_, err := exec.Command("dscacheutil", "-flushcache").Output()

	if err != nil {
		log.Fatal(err)
	}
}

func sudoValidateUser() {
	_, err := exec.Command("sudo", "-v").Output()
	if err != nil {
		log.Fatal(err)
	}
}

func configure() {
	k8s := &k8s{}
	bind := Bind9{}

	log.Println("Configure")

	sudoValidateUser()

	k8s.kubecfg()
	ipIngressMinikube := k8s.getIngressMinikube()
	log.Println("Minikube ingress ip:", ipIngressMinikube)

	bind.updateK8sIngress(ipIngressMinikube)

	nodeList, err1 := k8s.core.Nodes().List(context.TODO(), metav1.ListOptions{})
	if err1 != nil {
		log.Fatal(err1)
	}

	log.Println("Running init script on Minikube nodes")
	for _, node := range nodeList.Items {
		minikubeSsh(node.Name, "[ -f init.sh ] && sudo sh init.sh")
	}

	log.Println("Creating storage on Minikube nodes")
	var pv int64 = 1
	for _, node := range nodeList.Items {
		minikubeSsh(node.Name, "sudo mkdir -p /data/local-storage/pv000"+strconv.FormatInt(pv, 10))
		pv++
		minikubeSsh(node.Name, "sudo mkdir -p /data/local-storage/pv000"+strconv.FormatInt(pv, 10))
		pv++
		minikubeSsh(node.Name, "sudo mkdir -p /data/standard-storage")
	}

	minikubeSetProfile("cluster2")

	log.Println("Add Minikube IP route")
	minikubeAddIpRoute()

	bind.restartBind()
	flushDnsCache()
}

func setIngress(argv []string) {
	if len(argv) != 3 {
		fmt.Println(len(argv))
		log.Fatal("Invalid argument")
	}

	k8s := k8s{}
	bind := Bind9{}

	k8s.kubecfg()

	namespace := argv[0]
	svc := argv[1]
	subDomain := argv[2]

	log.Printf("Set ingress on service %s/%s, subdomain %s", namespace, svc, subDomain)
	sudoValidateUser()
	ip := k8s.getSvcIp(namespace, svc)
	log.Printf("Update Bind to resolve \"*.%s.worldl.xpt\" and \"%s.worldl.xpt\""+
		" to service \"%s/%s ip %s\"",
		subDomain, subDomain, namespace, svc, ip)
	bind.updateResolver(subDomain, ip)

	bind.restartBind()
	flushDnsCache()
}

func help() {
	fmt.Println("Minikube lab tool")
	fmt.Println("Commands:")
	fmt.Println("   configure      Configure")
	fmt.Println("   set-ingress    Set K8S ingress")
}

func main() {
	log.Println("Minikube lab tool")

	if runtime.GOOS != "darwin" {
		log.Fatal("It can run only on MACOS")
	}

	args := os.Args[1:]

	if len(args) < 1 {
		help()
		return
	}

	switch args[0] {
	case "configure":
		configure()
	case "set-ingress":
		setIngress(args[1:])
	default:
		log.Fatal("Invalid command: " + args[0])
		help()
	}
}
