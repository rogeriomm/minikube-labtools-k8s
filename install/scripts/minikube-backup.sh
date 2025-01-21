#!/usr/bin/env zsh

PATH=$PATH:.

source minikube-lib.sh

echo "Backup minikube."

are_you_sure "Are you sure you want to do backup?"

sudoValidateUser

minikube-stop.sh

echo "Backuping Minikube on $MINIKUBE_BACKUP ..."
sudo tar -vcf $MINIKUBE_BACKUP/minikube-backup-`date "+%Y%m%d%H%M"`.tar $MINIKUBE_HOME
ls -lah  $MINIKUBE_BACKUP
