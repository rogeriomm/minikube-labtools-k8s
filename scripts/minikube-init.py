import os
import shutil
from python_hosts import Hosts, HostsEntry
import ipaddress
from kubernetes import client, config
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
    cmd = "kubectl get pod -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx " + \
          "-l app.kubernetes.io/component=controller -o=jsonpath='{.items[0].spec.nodeName}'"
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


def main() -> int:
    # Install Rich
    install()

    custom_theme = Theme({"success": "green", "error": "bold red"})
    console = Console(theme=custom_theme)

    console.print(":ok: Post installation...")

    # Configs can be set in Configuration class directly or using helper utility
    config.load_kube_config()

    kbct = client.CoreV1Api()

    ret = kbct.list_node()

    minikube_set_profile("cluster2")

    nodes = {}
    for i in ret.items:
        node = i.metadata.name
        address = i.status.addresses[0].address
        sshkey = minikube_get_sshkey(i.metadata.name)
        nodes[node] = (address, sshkey)

    for node in nodes:
        val = nodes[node]
        console.print(node, val)
        if val[0] == '' or val[1] == '':
            console.print(f"Minikube installation failed: invalid minikube ip/ssh-key node {node}", style="error")
            return

    # Execute initialization script on minikube nodes
    console.print("Running init script on nodes...")
    cmd = "ssh \"[ -f init.sh ] && sudo sh init.sh\""

    for node in nodes:
        minikube_cmd(node, cmd, is_print=True)

    ingress_node = minikube_get_ingress_node()

    console.print(f"Ingress node: {ingress_node}")
    minikube_ip_ingress = minikube_get_ip(ingress_node)

    with open('/usr/local/etc/dnsmasq.d/worldl.xpt.conf', 'w') as f:
        f.write(f"address=/.worldl.xpt/{minikube_ip_ingress}")

    # Setup resolver to use Dnsmasq local dns server for all *.worldl.xpt ingress lookups
    with open('/etc/resolver/worldl.xpt', 'w') as f:
        f.write("nameserver 127.0.0.1")

    # Setup resolver to use Kubernetes dns server for all *.cluster.local lookups
    with open('/etc/resolver/cluster.local', 'w') as f:
        f.write("nameserver 10.96.0.10")

    with open('/usr/local/etc/dnsmasq.conf', 'w') as f:
        f.write("listen-address=0.0.0.0")

    # Reload Dnsmasq configuration and clear cache
    assert os.system("sudo brew services restart dnsmasq") == 0
    assert os.system("dscacheutil -flushcache") == 0

    # Host integration
    assert os.system("route -n add 10.0.0.0/8 $(minikube ip)") == 0  # Pods, services

    return 0


if __name__ == "__main__":
    exit(main())
