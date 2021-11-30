   * Minikube Kubernetes setup
      * 2 cluster nodes
      * Ingress using minikube addon
      * Container registry
      * ArgoCD 
      * RANCHER 2.6
      * Local Docker, replace Docker Desktop

   * Tested only on MAC OS Monterey, 128G RAM, XEON 16 cores

# Preparing
```commandline
brew install kustomize helm minikube pyenv zsh rancher-cli
pyenv install 3.10.0
pyenv global 3.10.0 
```   
```commandline
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

# ArgoCD
   * ArgoCD login
      * ```argocd login argocd.world.xpt:443```

   * ArgoCD web
      * https://argocd.world.xpt

# Rancher
   * https://rancher.world.xpt

## References
   * https://argo-cd.readthedocs.io/en/stable/getting_started/
   * https://rancher.com/docs/rancher/v2.5/en/installation/install-rancher-on-k8s
