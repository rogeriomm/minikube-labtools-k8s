PATH=$PATH:/usr/local/bin
mac_ip=$(ifconfig bridge100 | grep "inet " | awk -F ' ' '{print $2}')
minikube mount --ip $mac_ip /Users/rogermm/git/dataops/bdmm/labtools-k8s:/labtools-k8s
