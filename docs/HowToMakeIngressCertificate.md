# Install CFSSL
```commandline
brew install cfssl
```

# Create a Certificate Signing Request

```commandline
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

# Create a Certificate Signing Request object to send to the Kubernetes API
```commandline
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
```commandline
kubectl certificate approve my-svc.my-namespace
```

# Download the Certificate and Use It
```commandline
kubectl get csr
```
```commandline
kubectl get csr my-svc.my-namespace -o jsonpath='{.status.certificate}' \
    | base64 --decode > server.crt
```

# Configure Minikube ingress 
```commandline
cd install/scripts/ingress-certs
kubectl -n kube-system delete secret mkcert
kubectl -n kube-system create secret tls mkcert --key server-key.pem --cert server.crt
minikube addons configure ingress
minikube addons disable ingress
minikube addons enable ingress
minikube-labtools-k8s configure
```
* On MACOS click on ca.crt to add certificate on Keychain

# References
   * https://minikube.sigs.k8s.io/docs/tutorials/custom_cert_ingress/: How to use custom TLS certificate with ingress addon
   * https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/: Manage TLS Certificates in a Cluster
