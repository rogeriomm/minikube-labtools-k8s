#!/bin/zsh

minikube -p cluster2 start 

minikube -p cluster  start
minikube -p cluster docker-env > ~/.minikube/docker-env

minikube profile cluster2

