import os
import sys
from python_hosts import Hosts, HostsEntry
import ipaddress


def get_minikube_ip(profile: str) -> str:
    stream = os.popen(f"minikube -p {profile} ip")
    m_ip = stream.read()
    m_ip = m_ip.rstrip()
    try:
        ip = ipaddress.ip_address(m_ip)
    except:
        m_ip = None
    return m_ip


def update_host(ip: str, address: str) -> bool:
    hosts = Hosts()
    hosts.remove_all_matching(name=address)
    argocd_entry = HostsEntry(entry_type='ipv4', address=ip, names=[address])
    hosts.add([argocd_entry], force=True)
    try:
        hosts.write()
    except:
        return False
    return True


def main():
    minikube_ip = get_minikube_ip("cluster")
    minikube_ip2 = get_minikube_ip("cluster2")

    if minikube_ip is None or minikube_ip2 is None:
        print("Minikube instalation failed")
        return

    if not update_host(minikube_ip2, 'argocd.world.xpt'):
        print("Minikube instalation failed")
        return


if __name__ == "__main__":
    main()
