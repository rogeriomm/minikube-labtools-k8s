# Install CFSSL
```shell
brew install cfssl
```

# Create a Certificate Signing Request

```shell
cat <<EOF | cfssl genkey - | cfssljson -bare server
{
  "hosts": [
    "*.worldl.xpt"
  ],
  "CN": "system:node:my-pod.my-namespace.pod.cluster.local",
  "key": {
    "algo": "ecdsa",
    "size": 256
  },
  "names": [
    {
      "O": "system:nodes"
    }
  ]
}
EOF
```
   * Generated files
      * server-key.pem
      * server.csr

# Create a Certificate Signing Request object to send to the Kubernetes API
```shell
cat <<EOF | kubectl apply -f -
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: my-svc.my-namespace
spec:
  request: $(cat server.csr | base64 | tr -d '\n')
  signerName: kubernetes.io/kubelet-serving
  usages:
  - digital signature
  - key encipherment
  - server auth
EOF
```

# Get the Certificate Signing Request Approved
```shell
kubectl certificate approve my-svc.my-namespace
```

# Download the Certificate and Use It
```shell
kubectl get csr
```

```shell
kubectl get csr my-svc.my-namespace -o jsonpath='{.status.certificate}' \
    | base64 --decode > server.crt
```

   * Show server certificate
```shell
openssl x509 -in server.crt -noout -text
```

   * Move certificates
```shell
mv server.crt server.csr server-key.pem ../install/scripts/ingress-certs
```

# Configure Minikube ingress 
```shell
cd ../install/scripts/ingress-certs
kubectl -n kube-system delete secret mkcert
kubectl -n kube-system create secret tls mkcert --key server-key.pem --cert server.crt
echo "Enter custom cert: kube-system/mkcert" && minikube addons configure ingress
minikube -p cluster2 addons disable ingress
minikube -p cluster2 addons enable ingress
labtools-k8s configure
```
  * On MACOS click on $MINIKUBE_HOME/ca.crt to add Minikube CA certificate on Keychain

# References
   * https://minikube.sigs.k8s.io/docs/tutorials/custom_cert_ingress/: How to use custom TLS certificate with ingress addon
   * https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/: Manage TLS Certificates in a Cluster
   * https://security.stackexchange.com/questions/6873/can-a-wildcard-ssl-certificate-be-issued-for-a-second-level-domain/40481#40481
      * https://cheapsslsecurity.com/p/why-is-my-wildcard-ssl-not-working-on-a-second-level-subdomain/
     