/*

 */
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
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
	bind.f.Close()
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
	f.WriteString("*.worldl.xpt. IN A " + ip + "\n")
	f.Close()
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
	f.WriteString("*." + subDomain + ".worldl.xpt. IN A " + ip + "\n")
	f.WriteString(subDomain + ".worldl.xpt. IN A " + ip + "\n")
	f.Close()
}

type k8s struct {
	core v1.CoreV1Interface
}

func (k *k8s) kubecfg() {
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	k.core = clientset.CoreV1()
}

func (k *k8s) getNodeIngress() string {
	pod := k.core.Pods("ingress-nginx")
	opts := metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=ingress-nginx",
	}

	p, err := pod.List(context.TODO(), opts)
	if err != nil {
		log.Fatal(err)
	}
	return p.Items[0].Spec.NodeName
}

func (k *k8s) getIngressMinikube() string {
	node := k.getNodeIngress()

	n, err := k.core.Nodes().Get(context.TODO(), node, metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}

	ip := n.Status.Addresses[0].Address

	return ip
}

func (k *k8s) getSvcIp(namespace string, name string) string {
	svc, err := k.core.Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})

	if err != nil {
		log.Fatal(err)
	}

	return svc.Spec.ClusterIP
}

func restartBind() {
	log.Println("Restarting BIND...")
	_, err := exec.Command("sudo", "brew", "services", "restart", "bind").Output()
	//log.Println(string(out))
	if err != nil {
		log.Fatal(err)
	}
}

func flushDnsCache() {
	log.Println("Flushing host DNS cache...")
	_, err := exec.Command("dscacheutil", "-flushcache").Output()

	if err != nil {
		log.Fatal(err)
	}
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

	out, err := exec.Command("sudo", "route", "-n", "add", "10.0.0.0/8", ip).Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))
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
	}

	minikubeSetProfile("cluster2")

	log.Println("Add Minikube IP route")
	minikubeAddIpRoute()

	restartBind()
	flushDnsCache()
}

func setIngress(namespace string, svc string, subDomain string) {
	k8s := k8s{}
	bind := Bind9{}

	k8s.kubecfg()

	log.Printf("Set ingress on service %s/%s, subdomain %s", namespace, svc, subDomain)
	sudoValidateUser()
	ip := k8s.getSvcIp(namespace, svc)
	log.Printf("Update Bind to resolve *.%s.worldl.xpt to service %s/%s ip %s", subDomain, namespace, svc, ip)
	bind.updateResolver(subDomain, ip)

	restartBind()
	flushDnsCache()
}

func main() {
	log.Println("Minikube lab tool")
	args := os.Args[1:]

	if len(args) < 1 {
		log.Println("Invalid argument")
		return
	}

	if args[0] == "configure" {
		configure()
	} else if args[0] == "set-ingress" {
		if len(args) != 4 {
			fmt.Println(len(args))
			log.Fatal("Invalid argument")
		}
		setIngress(args[1], args[2], args[3])
	} else {
		log.Fatal("Invalid argument")
	}
}
