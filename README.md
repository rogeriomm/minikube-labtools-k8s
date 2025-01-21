
# Docker on MAC OS
```commandline
cat ~/.docker/config.json
```

   * docker login http://registry-1.docker.io/
   * docker login https://registry-1.docker.io/
   * docker login registry-1.docker.io
      *  registry-1.docker.io
   * docker login
      * https://index.docker.io/v1/
   * Internal registry
      * docker login $(minikube -p cluster2 ip):5000
      * docker logout $(minikube -p cluster2 ip):5000
## Configuration
```commandline
minikube 
```
      
## Pushing image
```commandline
docker tag rogermm/spark-base-python:master registry.minikube:5000/rogermm/spark-base-python:master
docker push registry.minikube:5000/rogermm/spark-base-python:master 
```      

# ArgoCD

   * File /etc/hosts
      * ```echo $(minikube -p cluster2 ip) argocd.world.xpt >> /etc/hosts ``` 

   * ArgoCD login
      * ```argocd login argocd.world.xpt:443```

   * ArgoCD web
      * https://argocd.world.xpt



## References
* https://argo-cd.readthedocs.io/en/stable/getting_started/
