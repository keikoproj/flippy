FROM golang:1.19.5-alpine3.17 as builder

USER root

RUN apk update && apk upgrade && apk add curl \
  && curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.21.0/bin/linux/amd64/kubectl \
  && chmod +x ./kubectl && mv ./kubectl /usr/local/bin/kubectl \
  && curl -sSL -o /usr/local/bin/kubectl-argo-rollouts https://github.com/argoproj/argo-rollouts/releases/latest/download/kubectl-argo-rollouts-linux-amd64 \
  && chmod +x /usr/local/bin/kubectl-argo-rollouts

RUN kubectl version --client
RUN kubectl argo rollouts version

WORKDIR /workspace
COPY . ./

USER appuser

# Build
RUN CGO_ENABLED=0 GO111MODULE=on go build -a -o manager main.go

ENTRYPOINT ["/manager"]
