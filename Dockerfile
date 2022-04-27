FROM golang:1.14-alpine3.11 as builder

RUN apk update && apk upgrade && apk add curl \
  && curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.18.0/bin/linux/amd64/kubectl \
  && chmod +x ./kubectl && mv ./kubectl /usr/local/bin/kubectl \
  && curl -sSL -o /usr/local/bin/kubectl-argo-rollouts https://github.com/argoproj/argo-rollouts/releases/latest/download/kubectl-argo-rollouts-linux-amd64 \
  && chmod +x /usr/local/bin/kubectl-argo-rollouts

RUN kubectl version --client
RUN kubectl argo rollouts version

WORKDIR /workspace
COPY . ./

# Build
RUN CGO_ENABLED=0 GO111MODULE=on go build -a -o manager main.go

ENTRYPOINT ["/workspace/manager"]
