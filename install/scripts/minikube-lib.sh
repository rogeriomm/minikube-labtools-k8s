# https://itnext.io/goodbye-docker-desktop-hello-minikube-3649f2a1c469
# brew install docker-credential-helper

#
# https://minikube.sigs.k8s.io/docs/handbook/config/#selecting-a-kubernetes-version
# https://github.com/kubernetes/minikube/blob/master/pkg/minikube/constants/constants.go
# NewestKubernetesVersion = "v1.25.2"
# OldestKubernetesVersion = "v1.16.0"
#
KUBERNETES_VERSION_1="1.18.14"
KUBERNETES_VERSION_2="1.23.12"
#
#
#

MINIKUBE_HOME="${MINIKUBE_HOME:-${HOME}/.minikube}"
MINIKUBE_FILES=$MINIKUBE_HOME/files
MINIKUBE_ETC=$MINIKUBE_FILES/etc
#MINIKUBE_CERTS=$MINIKUBE_FILES/certs

minikube_check_config()
{
  if [ ! -d "$MINIKUBE_HOME" ]; then
    echo "Invalid Minikube home directory: $MINIKUBE_HOME"
    exit 1
  fi
}

install_kubectl()
{
  if ! which asdf ; then
    brew install asdf
  fi

  asdf install kubectl $KUBERNETES_VERSION_1
  asdf install kubectl $KUBERNETES_VERSION_2

  if ! which kubectx ; then
    brew install kubectx
  fi

  if ! which kubens ; then
    brew install kubens
  fi

  if ! which k9s ; then
    brew install k9s
  fi

  if ! which minikube ; then
    brew install minikube
  fi
}

cluster1_create()
{
  PROFILE=cluster

  minikube -p $PROFILE config set cpus 4
  minikube -p $PROFILE config set memory 16g
  minikube -p $PROFILE config set disk-size 130g
  minikube -p $PROFILE config view

  minikube -p $PROFILE start \
           --kubernetes-version="v${KUBERNETES_VERSION_1}" \
           --service-cluster-ip-range='10.112.0.0/12' \   # 0x0a 0x70 0x00 0x00
           --dns-domain='cluster1.local' \
           --nodes 1 --driver='hyperkit' --insecure-registry "192.168.64.0/24,10.0.0.0/8"

  minikube -p cluster docker-env > "$MINIKUBE_HOME"/docker-env

  source "$MINIKUBE_HOME"/docker-env
}

cluster2_create()
{
  PROFILE=cluster2

  minikube -p $PROFILE config set cpus 22
  minikube -p $PROFILE config set memory 80g
  minikube -p $PROFILE config set disk-size 100g
  minikube -p $PROFILE config view

  minikube -p $PROFILE start \
           --kubernetes-version="v${KUBERNETES_VERSION_2}" \
           --service-cluster-ip-range='10.96.0.0/12' \   # 0x0a 0x60 0x00 0x00
           --dns-domain='cluster.local' \
           --extra-config=kubelet.max-pods=100 \
           --nodes 3 --driver='hyperkit' --insecure-registry "192.168.64.0/24,10.0.0.0/8"

  minikube -p $PROFILE addons enable ingress
  minikube -p $PROFILE addons disable ingress-dns
  minikube -p $PROFILE addons enable registry
  minikube -p $PROFILE addons enable registry-aliases
  minikube -p $PROFILE addons enable dashboard
  minikube -p $PROFILE addons enable metrics-server
  minikube -p $PROFILE addons disable registry-creds
  minikube -p $PROFILE addons enable metallb
  minikube -p $PROFILE addons disable storage-provisioner

  minikube -p $PROFILE addons list

  cp minikube-certs/{ca.crt,ca.key,ca.pem,cert.pem,key.pem} "$MINIKUBE_HOME"
}

clusters_stop()
{
  rm -f "$MINIKUBE_HOME"/docker-env
  minikube -p cluster stop
  minikube -p cluster2 stop
}

clusters_start()
{
  set -x
  set -e

  ip=$(ifconfig en0 | grep 'inet ' | awk '{print $2}')

  if ! dig @$ip www.google.com; then
    echo "Local DNS server isn't working. Verify your local DNS configuration"
    exit 1
  fi

  minikube -p cluster2 start --embed-certs

  minikube -p cluster start --embed-certs
  minikube -p cluster docker-env > "$MINIKUBE_HOME"/docker-env
  source "$MINIKUBE_HOME"/docker-env

  minikube profile cluster2
}

clusters_post_start()
{
  labtools-k8s configure

  argocd_show_password

  #rancher_show_password

  kubectl get pv
}

minikube_get_host_ip()
{
  ip=$(ifconfig bridge100 | grep "inet " | awk -F ' ' '{print $2}')
  echo "$ip"
}

init_ingress()
{
  kubectl -n kube-system create secret tls mkcert \
      --key  "ingress-certs/server-key.pem" \
      --cert "ingress-certs/server.crt"
  echo "Enter custom cert: kube-system/mkcert"
  minikube addons configure ingress
}

argocd_setup()
{
  kubectx cluster2
  kubectl create namespace argocd
  kubectl apply -n argocd -f yaml2/argocd-install.yaml
  kubectl apply -f yaml2/argocd-ingress.yaml
}

argocd_show_password()
{
  kubectx cluster2

  while : ; do
    kubectl -n argocd get secret/argocd-initial-admin-secret 2> /dev/null > /dev/null && break
    sleep 20
  done
  set +x
  echo -n "ARGOCD admin password: "
  kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
  echo ""
  set -x
}

internal_registry_setup()
{
  # Configure docker internal registry
  kubectx cluster # Our docker engine used by MAC OS
  kubectl -n kube-system delete configmap registry-cluster || echo -n
  kubectl -n kube-system create configmap registry-cluster \
                  --from-literal=registryAliases=registry.minikube \
                  --from-literal=registryServiceHost="$(minikube -p cluster2 ip)" # Internal registry
  kubectl -n kube-system apply -f yaml2/node-etc-hosts-update.yaml
}

# https://rancher.com/docs/rancher/v2.5/en/installation/install-rancher-on-k8s/
rancher_setup()
{
  helm repo add rancher-stable https://releases.rancher.com/server-charts/stable
  helm repo update
  kubectl create namespace cattle-system
  helm install rancher rancher-stable/rancher \
     --namespace cattle-system \
     --set hostname=rancher.worldl.xpt \
     --set replicas=2 \
     --set ingress.tls.source=secret
}

rancher_show_password()
{
  set +x
  echo -n "Rancher password: "
  kubectl get secret --namespace cattle-system bootstrap-secret \
     -o go-template='{{.data.bootstrapPassword|base64decode}}{{"\n"}}'
  set -x
}

copy_cert()
{
  cp -f minikube-certs/* "$MINIKUBE_HOME"
}

create_mounts()
{
  {
    echo "# Added by script #"
    echo "host.minikube.internal:/Users/${USERNAME}/git /Users/${USERNAME}/git nfs nfsvers=3 0 0"
    echo "###################"
  } >> "${MINIKUBE_ETC}/fstab"
}

k8s_nfs_provisioner_setup()
{
  helm repo add nfs-subdir-external-provisioner https://kubernetes-sigs.github.io/nfs-subdir-external-provisioner/
  helm repo update
  kubectl create ns nfs-external-provisioner
  helm install --namespace nfs-external-provisioner nfs-subdir-external-provisioner \
      nfs-subdir-external-provisioner/nfs-subdir-external-provisioner \
      --values nfs-subdir-external-provisioner/values.yaml
}

are_you_sure()
{
  # shellcheck disable=SC1049
  if [ "$1" = "" ]; then
    read -r -q "REPLY?Initialize Minikube(y/n)? "
  else
    read -r -q "REPLY?$1(y/n)? "
  fi
  printf "\n"

  if [ "$REPLY" = "n" ]; then
    exit
  fi
}

sudoValidateUser()
{
  sudo -v
}

init()
{
  echo "Installation..."
  set -x
  set -e

  install_kubectl

  # Delete all clusters
  minikube -p cluster delete
  minikube -p cluster2 delete

  # Delete all files from Minikube home
  rm -rf "$MINIKUBE_HOME"
  mkdir -p "$MINIKUBE_HOME"

  cp -r ./files "$MINIKUBE_HOME"
  echo "HOST_USERNAME=${USERNAME}" > "${MINIKUBE_FILES}/home/docker/.values.conf"

  # Copy TLS certificates
  copy_cert

  # Create mounts, edit fstab
  create_mounts

  cluster1_create
  cluster2_create

  # Setup internal registry
  internal_registry_setup

  # Set current minikube profile
  minikube profile cluster2
  kubectx cluster2
  asdf global kubectl $KUBERNETES_VERSION_2

  # Setup Kubernetes NFS Subdir External Provisioner
  k8s_nfs_provisioner_setup

  init_ingress

  kubectl apply -f yaml2/persistent-volumes.yaml
  kubectl apply -f yaml2/dashboard-ingress.yaml

  # Setup Argocd
  argocd_setup

  # RANCHER setup
  #rancher_setup

  clusters_start

  clusters_post_start
}
