/*
 */
package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	"os/exec"
	"runtime"
)

var kub k8s

func flushDnsCache() {
	log.Println("Flushing host DNS cache...")
	_, err := exec.Command("dscacheutil", "-flushcache").Output()

	if err != nil {
		log.Fatal(err)
	}

	_, err = exec.Command("sudo", "killall", "-HUP", "mDNSResponder").Output()

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

func asdf(version string) {
	_, err := exec.Command("asdf", "global", "kubectl", version).Output()
	if err != nil {
		log.Fatal(err)
	}
}

func configure() {
	bind := Bind9{}

	log.Println("Configure")

	sudoValidateUser()

	kub.kubecfg("cluster2")
	minikubeSetProfile("cluster2")

	ipIngressMinikube := kub.getIngressMinikube()
	log.Println("Minikube ingress ip:", ipIngressMinikube)

	bind.updateK8sIngress(ipIngressMinikube)

	nodeList, err1 := kub.core.Nodes().List(context.TODO(), metav1.ListOptions{})
	if err1 != nil {
		log.Fatal(err1)
	}

	log.Println("Running init script on Minikube nodes")
	for _, node := range nodeList.Items {
		minikubeSsh(node.Name, "[ -f init.sh ] && sudo sh init.sh")
	}

	log.Println("Creating storage on Minikube nodes")

	createPv("cluster")
	createPv("cluster2")

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

	bind := Bind9{}

	kub.kubecfg("cluster2")

	namespace := argv[0]
	svc := argv[1]
	subDomain := argv[2]

	log.Printf("Set ingress on service %s/%s, subdomain %s", namespace, svc, subDomain)
	sudoValidateUser()
	ip := kub.getSvcIp(namespace, svc)
	log.Printf("Update Bind to resolve \"*.%s.worldl.xpt\" and \"%s.worldl.xpt\""+
		" to service \"%s/%s ip %s\"",
		subDomain, subDomain, namespace, svc, ip)
	bind.updateResolver(subDomain, ip)

	bind.restartBind()
	flushDnsCache()
}

func createPv(ctx string) {
	kub.kubecfg(ctx)
	minikubeSetProfile(ctx)

	pvList, err := kub.core.PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for _, pv := range pvList.Items {
		if pv.Spec.StorageClassName == "standard" &&
			*pv.Spec.VolumeMode == "Filesystem" {

			// FIXME check array size
			operator := pv.Spec.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Operator
			node := pv.Spec.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Values[0]
			path := pv.Spec.Local.Path

			if operator == "In" {
				minikubeSsh(node,
					"sudo mkdir -p "+path)
			}
		}
	}
}

func cleanAvailablePv(ctx string) {
	kub.kubecfg(ctx)
	minikubeSetProfile(ctx)

	pvList, err := kub.core.PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for _, pv := range pvList.Items {
		if pv.Status.Phase == "Available" &&
			pv.Spec.StorageClassName == "standard" &&
			*pv.Spec.VolumeMode == "Filesystem" {

			// FIXME check array size
			operator := pv.Spec.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Operator
			node := pv.Spec.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Values[0]
			name := pv.Name
			path := pv.Spec.Local.Path

			if operator == "In" {
				println("Cleanning node:", node, " pv:", name, " path:", path)
				minikubeSsh(node,
					"sudo mkdir -p "+path+
						" && cd "+path+
						" && ls -A1 | sudo xargs rm -rf")
			}
		}
	}
}

func help() {
	fmt.Println("Minikube lab tool")
	fmt.Println("Commands:")
	fmt.Println("  configure          Configure")
	fmt.Println("  set-ingress        Set K8S ingress")
	fmt.Println("  clean-available-pv Clean available pv")
	fmt.Println("  set-context        Set context")
}

func gen_man() {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "my test program",
	}

	header := &doc.GenManHeader{
		Title:   "MINE",
		Section: "3",
	}
	err := doc.GenManTree(cmd, header, "/tmp")
	if err != nil {
		log.Fatal(err)
	}
}

func setContext(ctx string) {
	kub.kubecfg(ctx)
	minikubeSetProfile(ctx)

	switch ctx {
	case "cluster":
		asdf("1.18.14")

	case "cluster2":
		asdf("1.23.12")
	}
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
	case "clean-available-pv":
		if len(args) < 2 {
			help()
			return
		}
		cleanAvailablePv(args[1])
	case "set-context":
		if len(args) < 2 {
			help()
			return
		}
		setContext(args[1])

	default:
		log.Fatal("Invalid command: " + args[0])
		help()
	}
}
