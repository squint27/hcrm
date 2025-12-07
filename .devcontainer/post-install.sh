s#!/bin/bash
set -x

curl -Lo ./kind https://kind.sigs.k8s.io/dl/latest/kind-linux-$(go env GOARCH)
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/linux/$(go env GOARCH)
chmod +x kubebuilder
sudo mv kubebuilder /usr/local/bin/

KUBECTL_VERSION=$(curl -L -s https://dl.k8s.io/release/stable.txt)
curl -LO "https://dl.k8s.io/release/$KUBECTL_VERSION/bin/linux/$(go env GOARCH)/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/kubectl

docker network create -d=bridge --subnet=172.19.0.0/24 kind

kind version
kubebuilder version
docker --version
go version
kubectl version --client
