FROM docker.intuit.com/oicp/standard/golang-1.x/amzn2-golang:1.14

WORKDIR /workspace
USER root
ARG GIT_TOKEN

COPY . ./

# Build
RUN mv install/kubernetes.repo /etc/yum.repos.d/kubernetes.repo && sudo yum install -y kubectl git \
    && git config --global url."https://git:$GIT_TOKEN@github.intuit.com".insteadOf "https://github.intuit.com" \
    && go build -a -o manager main.go \
    && rm -rf /usr/local/go/bin/pkg/mod/cloud.google.com/go@v0.38.0/storage/testdata/dummy_rsa /usr/local/go/bin/pkg/mod/github.com/prometheus/common@v0.4.1/config/testdata/client-no-pass.key /usr/local/go/bin/pkg/mod/github.com/prometheus/common@v0.4.1/config/testdata/self-signed-client.key /usr/local/go/bin/pkg/mod/github.com/prometheus/common@v0.4.1/config/testdata/server.key /usr/local/go/bin/pkg/mod/k8s.io/apiextensions-apiserver@v0.18.6/pkg/cmd/server/testing/testdata/localhost_127.0.0.1_localhost.key /usr/local/go/bin/pkg/mod/k8s.io/apiextensions-apiserver@v0.18.6/test/integration/apiserver.local.config/certificates/apiserver.key /usr/local/go/bin/pkg/mod/k8s.io/client-go@v0.18.8/util/cert/testdata/dontUseThisKey.pem

RUN sudo curl -sSL -o /usr/local/bin/kubectl-argo-rollouts https://github.com/argoproj/argo-rollouts/releases/latest/download/kubectl-argo-rollouts-linux-amd64 \
    && chmod +x /usr/local/bin/kubectl-argo-rollouts

USER appuser

ARG JIRA_PROJECT=https://jira.intuit.com/projects/MESH
ARG DOCKER_IMAGE_NAME=docker.intuit.com/services/mesh/flippy/service/flippy:latest
ARG SERVICE_LINK=https://devportal.intuit.com/app/dp/resource/7209154696006625069/overview

ENTRYPOINT ["./manager"]
