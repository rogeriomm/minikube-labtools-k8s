#!/usr/bin/env zsh
# TODO https://stackoverflow.com/questions/36917882/how-to-use-pigz-with-tar

PATH=$PATH:.

source minikube-lib.sh

echo "Backup minikube."

are_you_sure "Are you sure you want to do backup?"

sudoValidateUser

mkdir -p /Volumes/backup/minikube-backup

minikube-stop.sh

echo "Backuping Minikube on $MINIKUBE_BACKUP ..."
sudo tar -vcf $MINIKUBE_BACKUP/minikube-backup-`date "+%Y%m%d%H%M"`.tar $MINIKUBE_HOME
ls -lah  $MINIKUBE_BACKUP
echo "Use pigz to compress backup using all CPU cores"

