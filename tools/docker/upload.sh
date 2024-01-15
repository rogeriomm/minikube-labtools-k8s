docker save docker-command -o  docker-command.tar
minikube -p cluster2 image load docker-command.tar
