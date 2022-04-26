# Contributing to Flippy

We welcome contributions :)


## Setting up for local Development


#### Kubebuilder
Flippy uses [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) for CRD implementation. Kubebuilder is a framework for building Kubernetes APIs using [custom resource definitions (CRDs)](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions).

To understand how kubebuilder you can refer to  [installation](https://book.kubebuilder.io/quick-start.html#installation) guide.

To install CRD - <BR>
`make install`

To run - <BR>
`make run`

#### Argo Rollouts

Flippy also support argo rollouts. [Please install argo rollouts add on.](https://argoproj.github.io/argo-rollouts/installation/#kubectl-plugin-installation)

#### Golang

Flippy is developed on [Golang Version 1.15](https://go.dev/doc/go1.15).

Please install [Golang specific version.](https://go.dev/doc/install)


#### Kubectl
Please install [kubectl tool](https://kubernetes.io/docs/tasks/tools/#kubectl)

## How to run in from local
TBD
