#!/bin/bash

if [ -f ".values.conf" ]; then
  source "$PWD/.values.conf"
fi

#NAME="lab"
#sudo mkdir -p "/$NAME"
#sudo umount "/$NAME" 2> /dev/null
#sudo mount.cifs "\\\\192.168.0.201\\$NAME" -o user=lab,pass=c3RvcG1lbm93Cg "/$NAME"

sudo mkdir -p /Users/$HOST_USERNAME/git
sudo mkdir -p /Volumes/data
sudo mount -a
