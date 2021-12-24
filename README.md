   * Minikube Kubernetes setup
      * 3 cluster nodes
      * Ingress using minikube addon
         * With wildcard DNS record resolving to the ip of Ingress node. Using Dnsmasq instead of Minikube addon ingress-dns   
      * Container registry
      * ArgoCD 
      * RANCHER 2.6
      * Local Docker, replace Docker Desktop

   * Tested only on MAC OS Monterey, 128G RAM, XEON 16 cores

# Install
## Install brew packages
```commandline
brew install kustomize helm minikube zsh rancher-cli go
```

## Build and install tool
```commandline
    mkdir -p $HOME/go
    GOPATH="$HOME/go"
    export PATH=$PATH:"$GOPATH/bin"
    
    cd install/src
    go install
```

## Configure MAC OS NFS server
   * /etc/exports
```text
/Users/rogermm/git -maproot=rogermm -rw -network 192.168.64.0 -mask 255.255.255.0
/Volumes/data -maproot=rogermm -rw -network 192.168.64.0 -mask 255.255.255.0
```
```commandline
sudo nfsd enable
sudo nfsd restart
```

# ArgoCD
   * ArgoCD login
```commandline
   kubectl port-forward svc/argocd-server -n argocd 8080:443
   argocd login --insecure --username admin localhost:8080
```

   * ArgoCD web
      * https://argocd.world.xpt

# Rancher
   * https://rancher.world.xpt

## BIND
```commandline
sudo brew services restart bind
sudo brew  services info bind
```
   * Debugging
```commandline
tail -f /usr/local/var/log/named/named.log
```

```commandline
scutil --dns
```

# Post install checklist
## *.cluster.local dns lookups and service/pods connection on host
```commandline
kubectl create namespace default
kubectl create deployment web --image=gcr.io/google-samples/hello-app:1.0 -n default
kubectl expose deployment web --type=NodePort --port=8080 -n default
```

```text
$ kubectl get svc/web
NAME   TYPE       CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE
web    NodePort   10.109.115.245   <none>        8080:32742/TCP   90s
```
   * Check dns cluster lookup
```commandline
dig @10.96.0.10 kube-dns.kube-system.svc.cluster.local
dig @10.96.0.10 web.default.svc.cluster.local
```
   * Check TCP service/pod TCP connection
```commandline
curl http://web.default.svc.cluster.local:8080
```
## Ingress dns lookups
```commandline
ping anything.worldl.xpt
```

# References
   * https://argo-cd.readthedocs.io/en/stable/getting_started/
   * https://rancher.com/docs/rancher/v2.5/en/installation/install-rancher-on-k8s
   * https://www.youtube.com/watch?v=_pUXOn_VRdQ: Nginx Ingress Controller Minikube with dnsmasq
   * https://gist.github.com/davebarnwell/c408533d608bfe24f4f5: Install dnsmasq and configure for *.dev.local domains
   * https://gist.github.com/loa/a88803c5678381eb515ab7f1241199a3: Minikube host networking integration
   * https://kubernetes.io/docs/concepts/storage/volumes/#local: 