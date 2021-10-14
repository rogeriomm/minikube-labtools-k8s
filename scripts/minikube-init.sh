#~/bin/zsh

#minikube delete

minikube config set cpus 14
minikube config set memory 80g
minikube config set disk-size 100g

minikube start --driver='hyperkit'

minikube addons enable ingress
minikube addons enable ingress-dns
minikube addons enable registry 
minikube addons enable dashboard 
minikube addons list 

minikube docker-env > ~/.minikube/docker-env

