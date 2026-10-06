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

4. (Optional) Enable namespace-level staggering<br>
   By default Flippy restarts all matched objects across all namespaces back to back. On clusters where many
   mesh-enabled namespaces must not rotate at the same time (for example multi-shard tenants), set the
   `FLIPPY_STAGGER_NAMESPACES` environment variable on the Flippy container to `true`:
    ```
    env:
      - name: FLIPPY_STAGGER_NAMESPACES
        value: "true"
    ```
   When enabled, Flippy processes namespaces one at a time in sorted order and only moves to the next namespace
   after every restarted object in the current namespace reports healthy or exhausts its retries. The health wait
   is forced even if the `RestartObjects[].StatusCheckConfig.CheckStatus` field is `false`; `MaxRetry` and
   `RetryDuration` are taken from that config when set, otherwise default to 10 retries every 30 seconds.
   Any value other than `true`/`1` (including unset or invalid) leaves the default behavior unchanged.
   This slows down the overall rollout, so enable it only on clusters that need it.


<HR>

Feel free to refer sample [Example deployment](../sample/deployment.yaml).

Feel free to refer sample [Flippy Config](../sample/sample.yaml)
 

