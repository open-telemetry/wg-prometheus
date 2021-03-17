#!/usr/bin/env sh
set -o errexit

cluster_name='kind-collector-load-balancer-1'
read -p "Are you sure you want to delete ${cluster_name} [y/n] " choice
case "$choice" in
  y|Y ) echo "deleting";;
  * ) exit 0;;
esac

kind delete cluster --name ${cluster_name}
