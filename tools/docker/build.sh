pushd .
cd command
docker build --no-cache . -t registry.minikube/command
popd
