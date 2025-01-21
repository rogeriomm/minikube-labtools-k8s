   * /etc/pf.conf
```text
#
# Default PF configuration file.
#
# This file contains the main ruleset, which gets automatically loaded
# at startup.  PF will not be automatically enabled, however.  Instead,
# each component which utilizes PF is responsible for enabling and disabling
# PF via -E and -X as documented in pfctl(8).  That will ensure that PF
# is disabled only when the last enable reference is released.
#
# Care must be taken to ensure that the main ruleset does not get flushed,
# as the nested anchors rely on the anchor point defined here. In addition,
# to the anchors loaded by this file, some system services would dynamically 
# insert anchors into the main ruleset. These anchors will be added only when
# the system service is used and would removed on termination of the service.
#
# See pf.conf(5) for syntax.
#

#
# com.apple anchor point
#
# Redirect TCP port 80 to 127.0.0.1 port 8080


scrub-anchor "com.apple/*"
nat-anchor "com.apple/*"
rdr pass on lo0 inet proto tcp from any to any port 80  -> 127.0.0.1 port 8080
rdr pass on lo0 inet proto tcp from any to any port 443 -> 127.0.0.1 port 8081
rdr-anchor "com.apple/*"
dummynet-anchor "com.apple/*"
anchor "com.apple/*"
load anchor "com.apple" from "/etc/pf.anchors/com.apple"
```

```shell
ip=$(minikube -p cluster2 ip)
echo $ip
```

```shell
sudo ifconfig en0 alias 192.168.49.2 255.255.255.0 
```

```shell
ifconfig en0
```
```text
en0: flags=8863<UP,BROADCAST,SMART,RUNNING,SIMPLEX,MULTICAST> mtu 1500
        options=46b<RXCSUM,TXCSUM,VLAN_HWTAGGING,TSO4,TSO6,CHANNEL_IO>
        ether 74:56:3c:65:7b:84
        inet6 fe80::ff:34d9:ce3:b44e%en0 prefixlen 64 secured scopeid 0x4 
        inet 192.168.15.250 netmask 0xffffff00 broadcast 192.168.15.255
        inet6 fd2f:eb79:a8ba:0:4ad:4312:a4b7:c329 prefixlen 64 autoconf secured 
        inet6 2804:1b3:3001:d48e:ce7:6f28:3696:3f31 prefixlen 64 autoconf secured 
        inet6 2804:1b3:3001:d48e:e1cb:be73:fbb:918d prefixlen 64 autoconf temporary 
        inet 192.168.49.2 netmask 0xffffff00 broadcast 255.255.255.0
        nd6 options=201<PERFORMNUD,DAD>
        media: autoselect (1000baseT <full-duplex>)
        status: active

```

```shell
sudo pfctl -f /etc/pf.conf
```

```shell
sudo pfctl -e
```

```shell
docker inspect cluster2 | jq  
```
```json
[
  {
    "Id": "fc303690e554b768b13186ba9941567a2fb99b9e56db2008eabf9132d28a9036",
    "Created": "2024-02-01T20:04:29.161137647Z",
    "Path": "/usr/local/bin/entrypoint",
    "Args": [
      "/sbin/init"
    ],
    "State": {
      "Status": "running",
      "Running": true,
      "Paused": false,
      "Restarting": false,
      "OOMKilled": false,
      "Dead": false,
      "Pid": 7322,
      "ExitCode": 0,
      "Error": "",
      "StartedAt": "2024-02-05T10:39:45.318267066Z",
      "FinishedAt": "2024-02-04T13:21:44.854407807Z"
    },
    "Image": "sha256:dbc648475405a75e8c472743ce721cb0b74db98d9501831a17a27a54e2bd3e47",
    "ResolvConfPath": "/var/lib/docker/containers/fc303690e554b768b13186ba9941567a2fb99b9e56db2008eabf9132d28a9036/resolv.conf",
    "HostnamePath": "/var/lib/docker/containers/fc303690e554b768b13186ba9941567a2fb99b9e56db2008eabf9132d28a9036/hostname",
    "HostsPath": "/var/lib/docker/containers/fc303690e554b768b13186ba9941567a2fb99b9e56db2008eabf9132d28a9036/hosts",
    "LogPath": "/var/lib/docker/containers/fc303690e554b768b13186ba9941567a2fb99b9e56db2008eabf9132d28a9036/fc303690e554b768b13186ba9941567a2fb99b9e56db2008eabf9132d28a9036-json.log",
    "Name": "/cluster2",
    "RestartCount": 0,
    "Driver": "overlay2",
    "Platform": "linux",
    "MountLabel": "",
    "ProcessLabel": "",
    "AppArmorProfile": "",
    "ExecIDs": null,
    "HostConfig": {
      "Binds": [
        "/lib/modules:/lib/modules:ro",
        "cluster2:/var"
      ],
      "ContainerIDFile": "",
      "LogConfig": {
        "Type": "json-file",
        "Config": {}
      },
      "NetworkMode": "cluster2",
      "PortBindings": {
        "22/tcp": [
          {
            "HostIp": "127.0.0.1",
            "HostPort": "0"
          }
        ],
        "2376/tcp": [
          {
            "HostIp": "127.0.0.1",
            "HostPort": "0"
          }
        ],
        "32443/tcp": [
          {
            "HostIp": "127.0.0.1",
            "HostPort": "0"
          }
        ],
        "5000/tcp": [
          {
            "HostIp": "127.0.0.1",
            "HostPort": "0"
          }
        ],
        "8443/tcp": [
          {
            "HostIp": "127.0.0.1",
            "HostPort": "0"
          }
        ]
      },
      "RestartPolicy": {
        "Name": "no",
        "MaximumRetryCount": 0
      },
      "AutoRemove": false,
      "VolumeDriver": "",
      "VolumesFrom": null,
      "ConsoleSize": [
        0,
        0
      ],
      "CapAdd": null,
      "CapDrop": null,
      "CgroupnsMode": "private",
      "Dns": [],
      "DnsOptions": [],
      "DnsSearch": [],
      "ExtraHosts": null,
      "GroupAdd": null,
      "IpcMode": "private",
      "Cgroup": "",
      "Links": null,
      "OomScoreAdj": 0,
      "PidMode": "",
      "Privileged": true,
      "PublishAllPorts": false,
      "ReadonlyRootfs": false,
      "SecurityOpt": [
        "seccomp=unconfined",
        "apparmor=unconfined",
        "label=disable"
      ],
      "Tmpfs": {
        "/run": "",
        "/tmp": ""
      },
      "UTSMode": "",
      "UsernsMode": "",
      "ShmSize": 67108864,
      "Runtime": "runc",
      "Isolation": "",
      "CpuShares": 0,
      "Memory": 0,
      "NanoCpus": 32000000000,
      "CgroupParent": "",
      "BlkioWeight": 0,
      "BlkioWeightDevice": [],
      "BlkioDeviceReadBps": [],
      "BlkioDeviceWriteBps": [],
      "BlkioDeviceReadIOps": [],
      "BlkioDeviceWriteIOps": [],
      "CpuPeriod": 0,
      "CpuQuota": 0,
      "CpuRealtimePeriod": 0,
      "CpuRealtimeRuntime": 0,
      "CpusetCpus": "",
      "CpusetMems": "",
      "Devices": [],
      "DeviceCgroupRules": null,
      "DeviceRequests": null,
      "MemoryReservation": 0,
      "MemorySwap": 0,
      "MemorySwappiness": null,
      "OomKillDisable": null,
      "PidsLimit": null,
      "Ulimits": [],
      "CpuCount": 0,
      "CpuPercent": 0,
      "IOMaximumIOps": 0,
      "IOMaximumBandwidth": 0,
      "MaskedPaths": null,
      "ReadonlyPaths": null
    },
    "GraphDriver": {
      "Data": {
        "LowerDir": "/var/lib/docker/overlay2/d248826cb44085ca3df520bdcfe163c7d270d0757b7429e45445ace45a47316e-init/diff:/var/lib/docker/overlay2/fd3a1b1efc0471d70c927cd2244c37f55c3f7361d2067a53e69b929117f5e6e0/diff",
        "MergedDir": "/var/lib/docker/overlay2/d248826cb44085ca3df520bdcfe163c7d270d0757b7429e45445ace45a47316e/merged",
        "UpperDir": "/var/lib/docker/overlay2/d248826cb44085ca3df520bdcfe163c7d270d0757b7429e45445ace45a47316e/diff",
        "WorkDir": "/var/lib/docker/overlay2/d248826cb44085ca3df520bdcfe163c7d270d0757b7429e45445ace45a47316e/work"
      },
      "Name": "overlay2"
    },
    "Mounts": [
      {
        "Type": "volume",
        "Name": "cluster2",
        "Source": "/var/lib/docker/volumes/cluster2/_data",
        "Destination": "/var",
        "Driver": "local",
        "Mode": "z",
        "RW": true,
        "Propagation": ""
      },
      {
        "Type": "bind",
        "Source": "/lib/modules",
        "Destination": "/lib/modules",
        "Mode": "ro",
        "RW": false,
        "Propagation": "rprivate"
      }
    ],
    "Config": {
      "Hostname": "cluster2",
      "Domainname": "",
      "User": "",
      "AttachStdin": false,
      "AttachStdout": false,
      "AttachStderr": false,
      "ExposedPorts": {
        "22/tcp": {},
        "2376/tcp": {},
        "32443/tcp": {},
        "5000/tcp": {},
        "8443/tcp": {}
      },
      "Tty": true,
      "OpenStdin": false,
      "StdinOnce": false,
      "Env": [
        "container=docker",
        "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
      ],
      "Cmd": null,
      "Image": "gcr.io/k8s-minikube/kicbase:v0.0.42@sha256:d35ac07dfda971cabee05e0deca8aeac772f885a5348e1a0c0b0a36db20fcfc0",
      "Volumes": null,
      "WorkingDir": "/",
      "Entrypoint": [
        "/usr/local/bin/entrypoint",
        "/sbin/init"
      ],
      "MacAddress": "02:42:c0:a8:31:02",
      "OnBuild": null,
      "Labels": {
        "created_by.minikube.sigs.k8s.io": "true",
        "mode.minikube.sigs.k8s.io": "cluster2",
        "name.minikube.sigs.k8s.io": "cluster2",
        "role.minikube.sigs.k8s.io": ""
      },
      "StopSignal": "SIGRTMIN+3"
    },
    "NetworkSettings": {
      "Bridge": "",
      "SandboxID": "6a86e8401e5b7b1883d689dce7fc5fa4dbad605e66d868d33434d12d14e3f648",
      "SandboxKey": "/var/run/docker/netns/6a86e8401e5b",
      "Ports": {
        "22/tcp": [
          {
            "HostIp": "127.0.0.1",
            "HostPort": "50571"
          }
        ],
        "2376/tcp": [
          {
            "HostIp": "127.0.0.1",
            "HostPort": "50572"
          }
        ],
        "32443/tcp": [
          {
            "HostIp": "127.0.0.1",
            "HostPort": "50573"
          }
        ],
        "5000/tcp": [
          {
            "HostIp": "127.0.0.1",
            "HostPort": "50574"
          }
        ],
        "8443/tcp": [
          {
            "HostIp": "127.0.0.1",
            "HostPort": "50575"
          }
        ]
      },
      "HairpinMode": false,
      "LinkLocalIPv6Address": "",
      "LinkLocalIPv6PrefixLen": 0,
      "SecondaryIPAddresses": null,
      "SecondaryIPv6Addresses": null,
      "EndpointID": "",
      "Gateway": "",
      "GlobalIPv6Address": "",
      "GlobalIPv6PrefixLen": 0,
      "IPAddress": "",
      "IPPrefixLen": 0,
      "IPv6Gateway": "",
      "MacAddress": "",
      "Networks": {
        "cluster2": {
          "IPAMConfig": {
            "IPv4Address": "192.168.49.2"
          },
          "Links": null,
          "Aliases": [
            "fc303690e554",
            "cluster2"
          ],
          "MacAddress": "02:42:c0:a8:31:02",
          "NetworkID": "03adc8c0810e83803b5b116afe515476b6bf7c9da7f9916cc3227c08222662e3",
          "EndpointID": "1294889108db9ad03ad1007e0006ff706784ce38f9be9a73d5909501c3899a6d",
          "Gateway": "192.168.49.1",
          "IPAddress": "192.168.49.2",
          "IPPrefixLen": 24,
          "IPv6Gateway": "",
          "GlobalIPv6Address": "",
          "GlobalIPv6PrefixLen": 0,
          "DriverOpts": null,
          "DNSNames": [
            "cluster2",
            "fc303690e554"
          ]
        }
      }
    }
  }
]
```

```shell
SSH_PORT=50571
ssh -p $SSH_PORT -i /Volumes/data/.minikube/machines/cluster2/id_rsa docker@localhost -L 8080:192.168.49.2:80 -L 8081:192.168.49.2:443
```

```shell
ssh -p 50571 -i /Volumes/data/.minikube/machines/cluster2/id_rsa docker@localhost -L 8080:192.168.49.2:80 -L 8081:192.168.49.2:443
```

```shell
nc -v 192.168.49.2 80
```

```shell
nc -v 192.168.49.2 443
```

```shell
minikube -p cluster2 ip
```
```text
192.168.49.2
```