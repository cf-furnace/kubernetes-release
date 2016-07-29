#!/bin/bash

set -e

# Install certstrap
if ! command -v certstrap >/dev/null 2>&1; then
  go get -v github.com/square/certstrap
fi

script_dir=$(cd "$(dirname "$0")" && pwd)
depot_path="${script_dir}/../certs"

mkdir -p ${depot_path}
function join_array { local IFS=","; echo "$*"; }

#############################################################################
## Kubernetes
#############################################################################
# CA for TLS and client authentication
certstrap --depot-path ${depot_path} init --passphrase '' --common-name 'Kubernetes CA'
mv -f ${depot_path}/Kubernetes_CA.crt ${depot_path}/kube-ca.crt
mv -f ${depot_path}/Kubernetes_CA.crl ${depot_path}/kube-ca.crl
mv -f ${depot_path}/Kubernetes_CA.key ${depot_path}/kube-ca.key

server_cn=kube-apiserver.service.cf.internal
server_alternate_names=("$server_cn" "*.$server_cn" "kubernetes" "kubernetes.default" "kubernetes.default.svc" "kubernetes.default.svc.cluster.local")
server_ips=("10.244.24.2" "10.254.0.1") # host IP and cluster IP
certstrap --depot-path ${depot_path} request-cert --passphrase '' --common-name $server_cn --domain $(join_array ${server_alternate_names[@]}) --ip $(join_array ${server_ips[@]})
certstrap --depot-path ${depot_path} sign $server_cn --CA kube-ca
mv -f ${depot_path}/$server_cn.key ${depot_path}/kube-apiserver.key
mv -f ${depot_path}/$server_cn.csr ${depot_path}/kube-apiserver.csr
mv -f ${depot_path}/$server_cn.crt ${depot_path}/kube-apiserver.crt

# Client certificate to distribute to nodes
server_cn=kubelet
certstrap --depot-path ${depot_path} request-cert --passphrase '' --common-name $server_cn
certstrap --depot-path ${depot_path} sign $server_cn --CA kube-ca

# Admin client certificate to distribute to nodes
server_cn=kube-admin
certstrap --depot-path ${depot_path} request-cert --passphrase '' --common-name $server_cn
certstrap --depot-path ${depot_path} sign $server_cn --CA kube-ca
