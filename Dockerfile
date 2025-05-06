FROM golang:1.24-alpine3.21 AS builder

USER root

WORKDIR /workspace
COPY . ./

# Build
RUN CGO_ENABLED=0 GO111MODULE=on go build -a -o manager main.go

FROM alpine:latest

RUN adduser --disabled-password --gecos '' appuser

# Add ARG declarations to receive build args
ARG CREATED
ARG VERSION
ARG REVISION

RUN apk update && apk upgrade && apk add curl \
  && curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.21.0/bin/linux/amd64/kubectl \
  && chmod +x ./kubectl && mv ./kubectl /usr/local/bin/kubectl \
  && curl -sSL -o /usr/local/bin/kubectl-argo-rollouts https://github.com/argoproj/argo-rollouts/releases/latest/download/kubectl-argo-rollouts-linux-amd64 \
  && chmod +x /usr/local/bin/kubectl-argo-rollouts

RUN kubectl version --client
RUN kubectl argo rollouts version

WORKDIR /workspace
COPY --from=0 /workspace/manager ./
COPY --from=0 /workspace/install/startup.sh ./

USER appuser

#CMD exec /bin/sh -c "trap : TERM INT; sleep infinity & wait"
ENTRYPOINT ["./startup.sh"]

LABEL org.opencontainers.image.source="https://github.com/keikoproj/flippy"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.created="${CREATED}"
LABEL org.opencontainers.image.revision="${REVISION}"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.url="https://github.com/keikoproj/flippy/blob/master/README.md"
LABEL org.opencontainers.image.description="A Kubernetes controller for creating and managing worker node instance groups across multiple providers"