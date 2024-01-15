#!/bin/bash

service tinyproxy start
service danted start
service named start
service ssh start

echo -e "nameserver 127.0.0.1\nsearch cluster2.xpt cluster1.xpt" | tee /etc/resolv.conf

if [ -f /tmp/conf/start.sh ] ; then
   sh  /tmp/conf/start.sh
fi

sleep infinity
