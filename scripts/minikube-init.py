import os
import shutil
from python_hosts import Hosts, HostsEntry
import ipaddress
from rich.traceback import install
from rich.console import Console
from rich.theme import Theme


def minikube_cmd(node: str, cmd: str, profile=None, is_print=False):
    if profile is None:
        s = f"minikube --node={node} {cmd}"
    else:
        s = f"minikube -p {profile} --node={node} {cmd}"
    stream = os.popen(s)
    ret = stream.read()
    if is_print:
        print(ret, end='')
    return ret


def minikube_get_ip(node: str, profile=None) -> str:
    m_ip = minikube_cmd(node, "ip", profile)
    m_ip = m_ip.rstrip()
    try:
        ipaddress.ip_address(m_ip)
    except:
        m_ip = None
    return m_ip


def minikube_get_sshkey(node: str, profile=None) -> str:
    key = minikube_cmd(node, "ssh-key", profile)
    key = key.strip()
    return key


def minikube_get_ingress_node() -> str:
    cmd = "kubectl get pod -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx -l app.kubernetes.io/component=controller -o=jsonpath='{.items[0].spec.nodeName}'"
    stream = os.popen(cmd)
    return stream.read()


def minikube_set_profile(profile: str):
    stream = os.popen(f"minikube profile {profile}")
    print(stream.read(), end='')


def update_host(ip: str, address: []) -> bool:
    shutil.copyfile("/etc/hosts", "/tmp/hosts")
    hosts = Hosts(path="/tmp/hosts")
    hosts.remove_all_matching(name=address)
    argocd_entry = HostsEntry(entry_type='ipv4', address=ip, names=address)
    hosts.add([argocd_entry], force=True)
    hosts.write()
    os.system("sudo echo -n")
    os.system("sudo cp /tmp/hosts /etc/hosts")
    return True


def scp(ip: str, ssh_key: str, file: str):
    cmd = f"scp -o \"StrictHostKeyChecking no\"  -i {ssh_key} {file} docker@{ip}: 2> /dev/null"
    stream = os.popen(cmd)
    k = stream.read()
    print(k, end='')


def cluster2_run(cmd: str):
    minikube_cmd("cluster2-m02", cmd, "cluster2", is_print=True)
    minikube_cmd("cluster2", cmd, "cluster2", is_print=True)


def main():
    # Install Rich
    install()

    custom_theme = Theme({"success": "green", "error": "bold red"})
    console = Console(theme=custom_theme)

    console.print(":ok: Post installation...")

    minikube_set_profile("cluster2")
    minikube_ip1 = minikube_get_ip("cluster2")
    minikube_ip2 = minikube_get_ip("cluster2-m02")
    sshkey1 = minikube_get_sshkey("cluster2")
    sshkey2 = minikube_get_sshkey("cluster2-m02")

    console.print(f"Minikube cluster2 ip: {minikube_ip1} {minikube_ip2}")

    if minikube_ip1 is None or \
            minikube_ip2 is None or \
            sshkey1 is None or \
            sshkey2 is None:
        console.print("Minikube installation failed: invalid minikube ip/ssh-key", style="error")
        return

    # Copy initialization script to minikube nodes
    scp(minikube_ip1, sshkey1, "init.sh")
    scp(minikube_ip2, sshkey2, "init.sh")
    minikube_cmd("cluster2", "ssh \"chmod +x init.sh\"")
    minikube_cmd("cluster2-m02", "ssh \"chmod +x init.sh\"")

    ingress_node = minikube_get_ingress_node()

    console.print(f"Ingress node: {ingress_node}")
    minikube_ip_ingress = minikube_get_ip(ingress_node)

    if not update_host(minikube_ip_ingress, ['argocd.world.xpt', 'rancher.world.xpt']):
        console.print("Minikube installation failed: update hosts", style="error")
        return

    # Execute initialization script on minikube nodes
    minikube_cmd("cluster2", "ssh \"sudo ./init.sh\"", is_print=True)
    minikube_cmd("cluster2-m02", "ssh \"sudo ./init.sh\"", is_print=True)


if __name__ == "__main__":
    main()
