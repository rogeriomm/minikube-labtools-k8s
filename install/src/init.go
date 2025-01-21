package main

// Last stable kuberbetes version supported by RANCHER 2.6
const KUBERNETES_VERSION = "1.21.6"

func argocd_show_password() {

}

func rancher_show_password() {

}

func cluster1_create() {
	minikubeSetProfile("cluster")

	minikubeRun("config set cpus 14")
	minikubeRun("config set memory 16g")
	minikubeRun("config set disk-size 100g")
	minikubeRun("config view")

	minikubeRun("start --kubernetes-version=\"v" + KUBERNETES_VERSION +
		"\" --nodes 1 --driver='hyperkit' --insecure-registry \"192.168.64.0/24,10.0.0.0/8\"")

	//minikube -p cluster docker-env > "$MINIKUBE_HOME"/docker-env
}

func cluster2_create() {
	minikubeSetProfile("cluster2")

	minikubeRun("config set cpus 28")
	minikubeRun("config set memory 80g")
	minikubeRun("config set disk-size 100g")
	minikubeRun("config view")

	minikubeRun("start --kubernetes-version=\"v" + KUBERNETES_VERSION +
		"\" --nodes 3 --driver='hyperkit' --insecure-registry \"192.168.64.0/24,10.0.0.0/8\"")

	minikubeRun("addons enable ingress")
	minikubeRun("addons disable ingress-dns")
	minikubeRun("addons enable registry")
	minikubeRun("addons enable registry-aliases")
	minikubeRun("addons enable dashboard")
	minikubeRun("addons enable metrics-server")
	minikubeRun("addons disable registry-creds")
	minikubeRun("addons enable metallb")
	minikubeRun("addons disable storage-provisioner")

	minikubeRun("addons list")
}

func clusters_start() {

	minikubeRun("minikube -p cluster2 start --embed-certs")
	//minikube -p cluster docker-env > "$MINIKUBE_HOME"/docker-env

	minikubeSetProfile("cluster2")
}

func clusters_post_start() {
	//minikube-labtools-k8s configure

	argocd_show_password()

	rancher_show_password()
}

func argocd_setup() {
	//kubectx cluster2
	//kubectl create namespace argocd
	//kubectl apply -n argocd -f argocd-install.yaml
	//kubectl apply -f argocd-ingress.yaml
}

func internal_registry_setup() {
	// Configure docker internal registry
	//kubectx cluster # Our docker engine used by MAC OS
	//kubectl -n kube-system delete configmap registry-cluster || echo -n
	//kubectl -n kube-system create configmap registry-cluster \
	//	--from-literal=registryAliases=registry.minikube \
	//	--from-literal=registryServiceHost=$(minikube -p cluster2 ip) # Internal registry
	//kubectl -n kube-system apply -f node-etc-hosts-update.yaml
}

// https://rancher.com/docs/rancher/v2.5/en/installation/install-rancher-on-k8s/
func rancher_setup() {
	//helm repo add rancher-stable https://releases.rancher.com/server-charts/stable
	//helm repo update
	//kubectl create namespace cattle-system
	//helm install rancher rancher-stable/rancher \
	//	--namespace cattle-system \
	//	--set hostname=rancher.worldl.xpt \
	//	--set replicas=2 \
	//	--set ingress.tls.source=secret
}

func are_you_sure() {

}
