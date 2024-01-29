package main

import (
	"context"
	"fmt"
	"github.com/bitfield/script"
	cp "github.com/otiai10/copy"
	"github.com/pbnjay/memory"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

const Cluster1 = "cluster1"

const Cluster1SvcSubnet = "10.112.0.0/12"

const Cluster2 = "cluster2"

const Cluster2SvcSubnet = "10.96.0.0/12"

const ClustersDomain = "xpt"

const KubernetesVersion1 = "1.28.6"
const KubernetesVersion2 = "1.28.6"

var kub1 = k8s{ctx: Cluster1}
var kub2 = k8s{ctx: Cluster2}
var mkb1 = minikube{Cluster1, true}
var mkb2 = minikube{Cluster2, true}
var bind bind9
var nfsServer nfs
var minikubeK8sPath string
var minikubeHomePath string
var sugar *zap.SugaredLogger

func configureClusters() {
	sugar.Info("Configure clusters")

	nfsServer.reconfigure()

	kub1.connect()
	kub2.connect()

	ipIngressMinikube := kub2.getIngressMinikube()
	sugar.Info("Minikube ingress ip: ", ipIngressMinikube)

	bind.updateK8sIngress(ipIngressMinikube)

	nodeList, err := kub2.core.Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		sugar.Fatal(err)
	}

	sugar.Info("Running init script on Minikube nodes")
	for _, node := range nodeList.Items {
		err := mkb2.ssh(node.Name, "[ -f init.sh ] && sudo sh init.sh")
		if err != nil {
			sugar.Fatal(err)
		}
	}

	kub2.useContext()

	err = kub2.createNamespace("argocd")
	if err != nil {
		sugar.Fatal(err)
	}

	err = kub2.createNamespace("nfs-external-provisioner")
	if err != nil {
		sugar.Fatal(err)
	}

	err = kub2.createNamespace("tunnel")
	if err != nil {
		sugar.Fatal(err)
	}

	err = kub2.ctl("apply", "-f", minikubeK8sPath+"/install/scripts/cluster2/")

	if err != nil {
		sugar.Fatal(err)
	}

	kub1.useContext()

	err = kub1.ctl("apply", "-f", minikubeK8sPath+"/install/scripts/cluster1")

	if err != nil {
		sugar.Fatal(err)
	}

	sugar.Info("Creating storage on Minikube nodes")

	createPv(&mkb1, &kub1)
	createPv(&mkb2, &kub2)

	sugar.Info("Add k8s services route")
	mkb1.minikubeAddIpRoute(Cluster1SvcSubnet)
	mkb2.minikubeAddIpRoute(Cluster2SvcSubnet)

	sugar.Info("Add k8s pods route")
	addPodRoute(&mkb1, &kub1)
	addPodRoute(&mkb2, &kub2)

	bind.restartBind()
	flushDnsCache()

	mkb1.setDockerEnv()
}

func setIngress(argv []string) {
	if len(argv) != 3 {
		sugar.Fatal("Invalid argument", len(argv))
	}

	namespace := argv[0]
	svc := argv[1]
	subDomain := argv[2]

	kub2.connect()

	sugar.Info("Set ingress on service %s/%s, subdomain %s", namespace, svc, subDomain)
	sudoValidateUser()
	ip := kub2.getSvcIp(namespace, svc)
	sugar.Info("Update Bind to resolve \"*.%s.worldl.xpt\" and \"%s.worldl.xpt\""+
		" to service \"%s/%s ip %s\"",
		subDomain, subDomain, namespace, svc, ip)
	bind.updateResolver(subDomain, ip)

	bind.restartBind()
	flushDnsCache()
}

func addPodRoute(mkb *minikube, kub *k8s) {
	nodeList, err := kub.core.Nodes().List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		sugar.Fatal(err)
	}

	for _, pv := range nodeList.Items {
		for _, i := range pv.Spec.PodCIDRs {
			mkb.minikubeAddIpRoute(i)
		}
	}
}

func createPv(mkbPv *minikube, kub *k8s) {
	pvList, err := kub.core.PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		sugar.Info(err)
	}

	for _, pv := range pvList.Items {
		if *pv.Spec.VolumeMode == "Filesystem" &&
			pv.Spec.NodeAffinity != nil &&
			len(pv.Spec.NodeAffinity.Required.NodeSelectorTerms) == 1 {
			operator := pv.Spec.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Operator
			node := pv.Spec.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Values[0]
			path := pv.Spec.Local.Path

			if operator == "In" {
				err = mkbPv.ssh(node,
					"sudo mkdir -p "+path)

				if err != nil {
					sugar.Fatal(err)
				}
			}
		}
	}
}

func cleanAvailablePv(ctx string) {
	var mkb *minikube
	var kub *k8s

	if ctx == mkb1.profile {
		kub = &kub1
		mkb = &mkb1
	} else if ctx == mkb2.profile {
		kub = &kub2
		mkb = &mkb2
	} else {
		sugar.Fatal("Invalid context")
	}

	kub.connect()

	pvList, err := kub.core.PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		sugar.Fatal(err)
	}

	for _, pv := range pvList.Items {
		if pv.Status.Phase == "Available" &&
			(pv.Spec.StorageClassName == "standard-consumer" || pv.Spec.StorageClassName == "minio-local-storage") &&
			*pv.Spec.VolumeMode == "Filesystem" &&
			pv.Spec.NodeAffinity != nil &&
			len(pv.Spec.NodeAffinity.Required.NodeSelectorTerms) == 1 {

			operator := pv.Spec.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Operator
			node := pv.Spec.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Values[0]
			name := pv.Name
			path := pv.Spec.Local.Path

			if operator == "In" {
				println("Cleaning node:", node, " pv:", name, " path:", path)
				mkb.ssh(node,
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
	fmt.Println("  configureClusters  Configure")
	fmt.Println("  set-ingress        Set K8S ingress")
	fmt.Println("  clean-available-pv Clean available pv")
	fmt.Println("  set-context        Set context")
}

func setContext(ctx string) {
	if ctx == mkb1.profile {
		kub1.connect()
		mkb1.applyProfile()
		kub1.useContext()
	} else if ctx == mkb2.profile {
		kub2.connect()
		mkb2.applyProfile()
		kub2.useContext()
	} else {
		sugar.Fatal("Invalid cluster: ", ctx)
	}
}

func showClustersConfiguration() {
	script.Exec("minikube profile list").Stdout()
	mkb1.showPlugins()
	mkb2.showPlugins()
	script.Exec("docker ps").Stdout()
	script.Exec("ip r").Stdout()
}

func startCluster2() {
	var osStr string

	switch runtime.GOOS {
	case "linux":
		osStr = `--driver='docker'`
	case "darwin":
		osStr = `--driver='docker'`
	}

	cmd := `minikube -p ` + mkb2.profile + ` start
	--kubernetes-version="v` + KubernetesVersion2 + `" 
	--dns-domain="` + mkb2.profile + `.` + ClustersDomain + `" 
	--extra-config=kubelet.max-pods=150
	--nodes 4
	--insecure-registry "192.168.0.0/16,10.0.0.0/8"
	--service-cluster-ip-range='` + Cluster2SvcSubnet + `' ` + `--cache-images=true ` +
		osStr

	sugar.Info(cmd)
	status, err := script.Exec(cmd).Stdout()

	if err != nil {
		sugar.Fatal(err, status)
	}

	mkb2.addonEnable([]string{"metrics-server", "ingress", "registry", "registry-aliases",
		"dashboard", "metallb", "storage-provisioner"}, true)
	mkb2.addonEnable([]string{"ingress-dns"}, false)
}

func startCluster1() {
	var osStr string

	switch runtime.GOOS {
	case "linux":
		osStr = `--driver='docker'`
	case "darwin":
		osStr = `--driver='docker'`
	}

	cmd := `minikube -p ` + mkb1.profile + ` start
	--kubernetes-version="v` + KubernetesVersion1 + `" 
	--dns-domain="` + mkb1.profile + `.` + ClustersDomain + `" 
	--nodes 1
	--insecure-registry "192.168.0.0/16,10.0.0.0/8"
	--service-cluster-ip-range='` + Cluster1SvcSubnet + `' ` + `--cache-images=true ` +
		osStr

	sugar.Info(cmd)
	status, err := script.Exec(cmd).Stdout()

	if err != nil {
		sugar.Fatal(err, status)
	}

	mkb1.addonEnable([]string{"metrics-server", "dashboard"}, true)
}

func createClusters() {
	cpus := runtime.NumCPU()
	mkb2.config("cpus", strconv.Itoa(cpus))

	totalRam := memory.TotalMemory()
	mkb2.configSize("memory", (float64(totalRam)/4)*0.65)

	mkb2.config("disk-size", "130G")

	startCluster2()

	mkb1.config("cpus", strconv.Itoa(cpus))
	mkb1.configSize("memory", (float64(totalRam)/3)*0.6)
	mkb1.config("disk-size", "100G")

	startCluster1()
}

func destroyClusters() {
	sugar.Info("Destroying cluster...")

	sugar.Info("Deleting cluster1")

	err := mkb1.stop()

	err = mkb1.delete()
	if err != nil {
		sugar.Error(err)
	}

	sugar.Info("Deleting cluster2")
	err = mkb2.stop()

	err = mkb2.delete()
	if err != nil {
		sugar.Error(err)
	}

	if runtime.GOOS == "linux" {
		sugar.Info("Delete docker networks")
		script.Exec("docker network rm minikube").Stdout()
		script.Exec("docker network rm cluster1").Stdout()
		script.Exec("docker network rm cluster2").Stdout()

		script.Exec("docker ps").Stdout()
	}
}

func installNfsProvisioner() {

	kub2.connect()

	err := kub2.createNamespace("nfs-external-provisioner")

	if err != nil {
		sugar.Fatal(err)
	}

	status, err := script.Exec("helm repo add nfs-subdir-external-provisioner https://kubernetes-sigs.github.io/nfs-subdir-external-provisioner/").Stdout()
	if err != nil {
		sugar.Fatal(err, status)
	}

	status, err = script.Exec("helm repo update").Stdout()
	if err != nil {
		sugar.Fatal(err, status)
	}

	status, err = script.Exec(`helm install --namespace nfs-external-provisioner nfs-subdir-external-provisioner 
                              nfs-subdir-external-provisioner/nfs-subdir-external-provisioner
                               --values ` + minikubeK8sPath + `/install/scripts/helm/nfs-subdir-external-provisioner/values.yaml `).Stdout()
	if err != nil {
		sugar.Fatal(err, status)
	}
}

func initializeClusters() {
	nfsServer.reconfigure()

	// Sync files into minikube: https://minikube.sigs.k8s.io/docs/handbook/filesync/
	err := cp.Copy(minikubeK8sPath+"/install/scripts/files",
		minikubeHomePath+"/files")

	if err != nil {
		sugar.Fatal(err)
	}

	// Copy TLS certificates
	err = cp.Copy(minikubeK8sPath+"/install/scripts/minikube-certs", minikubeHomePath)

	if err != nil {
		sugar.Fatal(err)
	}

	createClusters()

	setupK8sRegistry()

	configureIngress()

	installNfsProvisioner()

	configureClusters()
}

func setupK8sRegistry() {
	kub1.connect()

	// Configure docker internal registry. https://github.com/kameshsampath/minikube-helpers
	// kubectl -n kube-system delete configmap registry-cluster
	kubeSystem := kub1.core.ConfigMaps("kube-system")

	err := kubeSystem.Delete(context.TODO(), "registry-cluster", metav1.DeleteOptions{})

	//kubectl -n kube-system create configmap registry-cluster \
	//        --from-literal=registryAliases=registry.minikube \
	//        --from-literal=registryServiceHost="$(minikube -p $CLUSTER2 ip)" # Internal registry

	// kubectl get -n kube-system configmap registry-cluster -o yaml

	cm := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "registry-cluster"},
		Data: map[string]string{
			"registryAliases":     "registry.minikube",
			"registryServiceHost": mkb2.getIp(),
		}}

	_, err = kubeSystem.Create(context.TODO(), &cm, metav1.CreateOptions{})

	if err != nil {
		sugar.Fatal(err)
	}
}

func startClusters() {
	nfsServer.reconfigure()

	startCluster1()
	startCluster2()

	script.Exec("docker ps").Stdout()
}

func stopClusters() {
	// rm -f "$MINIKUBE_HOME"/docker-env
	mkb2.stop()
	mkb1.stop()

	script.Exec("docker ps").Stdout()
}

func sshCluster() {
	mkb1.loginSsh(1)
}

func configureIngress() {
	kub2.connect()
	kub2.useContext()

	pathCerts := minikubeK8sPath + "/install/scripts/ingress-certs/"
	status, err := script.Exec(`kubectl -n kube-system create secret tls mkcert \
		--key  "` + pathCerts + `/server-key.pem" \
		--cert "` + pathCerts + `/server.crt"`).Stdout()

	if err != nil {
		sugar.Fatal(err, status)
	}

	proc := exec.Command("minikube", "-p", mkb2.profile, "addons", "configure", "ingress")

	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	proc.Stdin = os.Stdin

	fmt.Println("Enter custom cert: kube-system/mkcert")

	if err := proc.Run(); err != nil {
		sugar.Fatal(err)
	}
}

func main() {
	logger, _ := zap.NewDevelopment()

	sugar = logger.Sugar()

	sugar.Info("minikube labtools")

	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		sugar.Fatal("It can run only on macOS or Linux")
	}

	labtoolsK8sPath := os.Getenv("LABTOOLS_K8S")

	if _, err := os.Stat(labtoolsK8sPath); os.IsNotExist(err) {
		sugar.Fatal("The LABTOOLS_K8S environment variable must be defined")
	}

	minikubeK8sPath = os.Getenv("LABTOOLS_K8S") + "/modules/minikube-labtools-k8s/"

	minikubeHomePath = os.Getenv("MINIKUBE_HOME")

	if len(minikubeHomePath) == 0 {
		sugar.Fatal("The MINIKUBE_HOME environment variable must be defined")
	}

	args := os.Args[1:]

	if len(args) < 1 {
		help()
		return
	}

	if runtime.GOOS == "linux" {
		os.Unsetenv("DOCKER_CERT_PATH")
		os.Unsetenv("DOCKER_HOST")
		os.Unsetenv("DOCKER_TLS_VERIFY")
	}

	switch args[0] {
	case "configure-clusters":
		sudoValidateUser()
		configureClusters()

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
	case "destroy-clusters":
		destroyClusters()
	case "create-clusters":
		createClusters()
	case "initialize-clusters":
		initializeClusters()
	case "show-clusters-configuration":
		showClustersConfiguration()
	case "ssh":
		sshCluster()
	case "flush-dns-cache":
		flushDnsCache()
	case "restart-bind":
		sudoValidateUser()
		bind.restartBind()
		flushDnsCache()
	case "start-clusters":
		sudoValidateUser()
		startClusters()
	case "stop-clusters":
		sudoValidateUser()
		stopClusters()
	case "reconfigure-nfs":
		nfsServer.reconfigure()
	default:
		sugar.Fatal("Invalid command: " + args[0])
	}
}
