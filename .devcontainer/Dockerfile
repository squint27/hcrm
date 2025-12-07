FROM mcr.microsoft.com/devcontainers/go:2-1.25-trixie

RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends curl 

RUN curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)" && \
chmod +x kubebuilder && sudo mv kubebuilder /usr/local/bin/
