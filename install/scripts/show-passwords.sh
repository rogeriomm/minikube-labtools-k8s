#!/usr/bin/env zsh

rancher_show_password()
{
  echo -n "Rancher password: "

  kubectl get secret --namespace cattle-system bootstrap-secret \
     -o go-template='{{.data.bootstrapPassword|base64decode}}{{"\n"}}'
}

argocd_show_password()
{
  while : ; do
    kubectl -n argocd get secret/argocd-initial-admin-secret 2> /dev/null > /dev/null && break
    sleep 15
  done
  echo -n "ARGOCD admin password: "
  kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
  echo ""
}

kubectx cluster2

rancher_show_password

argocd_show_password
