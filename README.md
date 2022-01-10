   * Minikube Kubernetes setup
      * 3 cluster nodes
      * Ingress using minikube addon
         * With wildcard DNS record resolving to the ip of Ingress node. Using BIND instead of Minikube addon ingress-dns   
      * Container registry
      * Local Docker, replace Docker Desktop
      * ArgoCD 
      * RANCHER 2.6

   * Target OS
      * MACOS    
         * Monterey, 128G RAM, XEON 16 cores

# Install
## Install packages on MACOS
```commandline
brew install zsh minikube helm go kustomize cfssl
```
   * Optional
```commandline
brew install rancher-cli
```


## Install Minikube TLS CA certificate on MACOS
   * Go to directory "install/scripts/minikube-certs". Double-click ca.crt and add certificate on "System"
   * Open "Keychain Access", click on "System", double-click "minikubeCA". 
      * On "Trust" set "Always Trust"
   * [Keychain Access screenshot](docs/MacOsKeyChainMinikubeCA.png)

### Conda Issue
   * https://docs.conda.io/projects/conda/en/latest/user-guide/configuration/non-standard-certs.html: CONDA, Using non-standard certificates
````shell
strace -f curl https://minio.minio-tenant-1.svc.cluster.local 2> /tmp/a
grep ssl /tmp/a
````
   * It doesn't use /etc/openssl!!!
```
openat(AT_FDCWD, "/opt/conda/envs/python_3_with_R/bin/../lib/./libssl.so.3", O_RDONLY|O_CLOEXEC) = 3
openat(AT_FDCWD, "/opt/conda/envs/python_3_with_R/ssl/openssl.cnf", O_RDONLY) = 3
openat(AT_FDCWD, "/opt/conda/envs/python_3_with_R/ssl/cacert.pem", O_RDONLY) = 6
write(2, "More details here: https://curl."..., 264More details here: https://curl.se/docs/sslcerts.html
```

   * Append Minikube CA certificate on "/opt/conda/envs/python_3_with_R/ssl/cacert.pem". "strace" works
   * "links" uses /etc/ssl directory
   * aws cli doesn't validate certificate
   * https://medium.com/@iffi33/dealing-with-ssl-authentication-on-a-secure-corporate-network-pip-conda-git-npm-yarn-bower-73e5b93fd4b2

### See also
   * https://minikube.sigs.k8s.io/docs/tutorials/custom_cert_ingress/: How to use custom TLS certificate with ingress addon
   * https://github.com/FiloSottile/mkcert: mkcert is a simple tool for making locally-trusted development certificates. It requires no configuration.


## Build and install management tool
```shell
    mkdir -p $HOME/go
    GOPATH="$HOME/go"
    export PATH=$PATH:"$GOPATH/bin"
    
    cd install/src
    go get
    go install
```

## Configure MACOS NFS server
   * /etc/exports
```
/Users/rogermm/git -alldirs -maproot=rogermm -network 192.168.64.0 -mask 255.255.255.0
/Volumes/data -alldirs -maproot=rogermm -network 192.168.64.0 -mask 255.255.255.0
/Users/rogermm/nfs -alldirs -maproot=rogermm -network 192.168.64.0 -mask 255.255.255.0
```

```commandline
sudo nfsd enable
sudo nfsd checkexports
sudo nfsd restart
```

# Kubernetes dashboard
* https://dashboard.worldl.xpt/

# ArgoCD
   * https://argocd.world.xpt

# Rancher
   * Disabled, waiting Kubernetes 1.22 compatibility for RANCHER 2.6
   * https://rancher.world.xpt

# BIND
   * Restart service
```commandline
sudo brew services restart bind
sudo brew  services info bind
```
   * Debugging
```commandline
tail -f /usr/local/var/log/named/named.log
```
   * Getting MACOS dns configuration
```commandline
scutil --dns
```

# Minikube ingress
![alt text](docs/IngressDiagram.png "Title")

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
```shell
dig @10.96.0.10 kube-dns.kube-system.svc.cluster.local
dig @10.96.0.10 web.default.svc.cluster.local
```
   * Check TCP service/pod TCP connection
```commandline
curl http://web.default.svc.cluster.local:8080
```

## Ingress dns lookups
```commandline
minikube ip
ping anything.worldl.xpt
```

# Minikube Ingress TLS certificate
```shell
kubectl -n kube-system create secret tls mkcert --key $MINIKUBE_HOME/key.pem --cert $MINIKUBE_HOME/cert.pem
```

```commandline
minikube addons configure ingress
```
```commandline
minikube addons disable ingress
minikube addons enable ingress
```

  * Verify if custom certificate was enabled
```shell
kubectl -n ingress-nginx get deployment ingress-nginx-controller -o yaml | grep "kube-system"
```

   * https://minikube.sigs.k8s.io/docs/tutorials/custom_cert_ingress/: How to use custom TLS certificate with ingress addon

# Kubernetes NFS Subdir External Provisioner
## Uninstall
```shell
helm list
helm uninstall nfs-subdir-external-provisioner
```

# See also
   * [Certificates](docs/HowToMakeIngressCertificate.md)
   * [Jetbrains](docs/Jetbrains.md)

# References
   * https://argo-cd.readthedocs.io/en/stable/getting_started/
   * https://rancher.com/docs/rancher/v2.5/en/installation/install-rancher-on-k8s
   * https://www.youtube.com/watch?v=_pUXOn_VRdQ: Nginx Ingress Controller Minikube with dnsmasq
   * https://gist.github.com/davebarnwell/c408533d608bfe24f4f5: Install dnsmasq and configure for *.dev.local domains
   * https://gist.github.com/loa/a88803c5678381eb515ab7f1241199a3: Minikube host networking integration
   * https://kubernetes.io/docs/concepts/storage/volumes/#local:
   * https://vocon-it.com/2018/12/31/kubernetes-6-https-applications-via-ingress-controller-on-minikube/: Kubernetes (6) – HTTPS Applications via Ingress Controller on Minikube 