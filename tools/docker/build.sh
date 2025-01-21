pushd .
cd command
docker build --no-cache . -t registry.minikube/command
popd
docker push registry.minikube/command
