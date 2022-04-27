# Flippy Installation on kubernetes cluster

To install flippy in kubernetes cluster following information needed.
1. Create [Kubernetes Service Account](https://kubernetes.io/docs/reference/access-authn-authz/service-accounts-admin/). <br>
Example - <br>
    ```
    ---
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      labels:
        k8s-app: flippy
      name: flippy-service-account
      namespace: istio-system
    ---
    ```
2. Create [Kubernetes Cluster Role & Role Binding](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#role-and-clusterrole)<br>
Example - <br>
    ```
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
      name: flippy-cluster-role
      namespace: istio-system
      labels:
        k8s-app: flippy
    rules:
      - apiGroups: ["*"]
        resources: ["configmaps","pods","namespaces","deployments","replicasets"]
        verbs: ["*"]
      - apiGroups: ["argoproj.io"]
        resources: ["*"]
        verbs: ["*"]
      - apiGroups: ["keikoproj.io"]
        resources: ["*"]
        verbs: ["*"]
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      name: flippy-cluster-role-binding
      namespace: istio-system
      labels:
        k8s-app: flippy
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: flippy-cluster-role
    subjects:
      - kind: ServiceAccount
        name: flippy-service-account
        namespace: istio-system
    ---
    ```

3. Install CRD<br>
    `kubectl apply -f config/crd/bases/keikoproj.io_flippyconfigs.yaml`


<HR>

Feel free to refer sample [Example deployment](../sample/deployment.yaml).
 

