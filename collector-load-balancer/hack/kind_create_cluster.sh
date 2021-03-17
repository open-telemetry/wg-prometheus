#!/usr/bin/env sh
set -o errexit

# From https://kind.sigs.k8s.io/docs/user/local-registry/

# create registry container unless it already exists
reg_name='kind-registry'
reg_port='5000'
echo "create container registry if not exists ${reg_name} localhost:${reg_port}"
running="$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)"
if [ "${running}" != 'true' ]; then
  docker run \
    -d --restart=always -p "127.0.0.1:${reg_port}:5000" --name "${reg_name}" \
    registry:2
fi

# generates cluster config
cluster_name='kind-collector-load-balancer-1'
# create a cluster with the local registry enabled in containerd
# writes the generated config out
cat > "${cluster_name}.yaml" <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${reg_port}"]
    endpoint = ["http://${reg_name}:${reg_port}"]
nodes:
- role: control-plane
- role: worker
- role: worker
- role: worker
EOF

echo "create cluster ${cluster_name}"
kind create cluster --config="${cluster_name}.yaml" --name ${cluster_name}

# connect the registry to the cluster network
# (the network may already be connected)
docker network connect "kind" "${reg_name}" || true

# Document the local registry
# https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/generic/1755-communicating-a-local-registry
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${reg_port}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF