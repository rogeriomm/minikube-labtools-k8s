
   * Update modules
```shell
go get -u
```

   * Build
```shell
go build -o $GOPATH/bin/labtools-k8s
```

   * Alias
```shell
alias lab1="labtools-k8s set-context cluster1 && k9s --cluster cluster1 -A"
alias lab2="labtools-k8s set-context cluster2 && k9s --cluster cluster2 -A"
```

   *  Test context switch
```shell
lab1
```

```shell
kubectl version
kubectx
minikube profile
```

```shell
lab2
```

```shell
kubectl version
kubectx
minikube profile
```

   * https://securitynetworkinglinux.wordpress.com/2019/07/03/how-to-masquerade-nat-ipv4-traffic-using-ufw-on-ubuntu-cli/
   * https://github.com/kubernetes/client-go