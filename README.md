

   * docker login http://registry-1.docker.io/
   * docker login https://registry-1.docker.io/
   * docker login registry-1.docker.io
      *  registry-1.docker.io
   * docker login
      * https://index.docker.io/v1/
   * Internal registry
      * docker login 192.168.64.5:500
      * docker logout 192.168.64.5:5000
     

# ArgoCD
   * https://argo-cd.readthedocs.io/en/stable/getting_started/

   * https://argocd.world.xpt

```commandline
argocd login argocd.world.xpt:443
```

```commandline
echo $(minikube -p cluster2 ip) argocd.world.xpt >> /etc/hosts
echo $(minikube -p cluster2 ip) zeppelin.world.xpt >> /etc/hosts
```
