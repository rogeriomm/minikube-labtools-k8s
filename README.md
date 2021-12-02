   * Minikube Kubernetes setup
      * 3 cluster nodes
      * Ingress using minikube addon
      * Container registry
      * ArgoCD 
      * RANCHER 2.6
      * Local Docker, replace Docker Desktop

   * Tested only on MAC OS Monterey, 128G RAM, XEON 16 cores

# Preparing
```commandline
brew install kustomize helm minikube pyenv zsh rancher-cli dsnmasq
pyenv install 3.10.0
pyenv global 3.10.0 
pip install kubernetes rich 
```

## MAC OS NFS server
   * /etc/exports
```text
/Users/rogermm/git -maproot=rogermm -rw -network 192.168.64.0 -mask 255.255.255.0
/Volumes/data -maproot=rogermm -rw -network 192.168.64.0 -mask 255.255.255.0
```
```commandline
sudo nfsd enable
sudo nfsd restart
```

## Dnsmasq
   * /usr/local/etc/dnsmasq.d/world.xpt.conf
```text
address=/.world.xpt/192.168.64.118
```

```commandline
scutil --dns
```
```commandline
sudo brew services restart dnsmasq
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

# References
   * https://argo-cd.readthedocs.io/en/stable/getting_started/
   * https://rancher.com/docs/rancher/v2.5/en/installation/install-rancher-on-k8s
   * https://www.youtube.com/watch?v=_pUXOn_VRdQ: Nginx Ingress Controller Minikube with dnsmasq
   * https://gist.github.com/davebarnwell/c408533d608bfe24f4f5: Install dnsmasq and configure for *.dev.local domains
