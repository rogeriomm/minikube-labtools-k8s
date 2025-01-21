#!/usr/bin/env zsh

minikube -p cluster stop 
minikube -p cluster2 stop 

rm -f ~/.minikube/docker-env

