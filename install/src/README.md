
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
alias lab="labtools-k8s set-context cluster && k9s --cluster cluster -A"
alias lab2="labtools-k8s set-context cluster2 && k9s --cluster cluster2 -A"
```

   * Test context switch
```shell
lab
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
