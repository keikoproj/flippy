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
   When enabled, Flippy processes namespaces one at a time in sorted order. Within a namespace each object is
   restarted and then waited on until it reports healthy or exhausts its retries before the next object is
   restarted, and the next namespace starts only after the current one is fully processed. The health wait
   is forced even if the `RestartObjects[].StatusCheckConfig.CheckStatus` field is `false`; `MaxRetry` and
   `RetryDuration` are taken from that config when set, otherwise default to 10 retries every 30 seconds.
   Flippy also pauses one `RetryDuration` after each restart before the first status poll so a stale
   "Healthy" status cannot release the gate. If the restart command itself fails (for example the object was
   deleted), the wait is skipped for that object. Because the wait is sequential, a staggered run can keep a
   single reconcile busy for a long time; this is expected.
   Any value other than `true`/`1` (including unset or invalid) leaves the default behavior unchanged.
   This slows down the overall rollout, so enable it only on clusters that need it.


<HR>

Feel free to refer sample [Example deployment](../sample/deployment.yaml).

Feel free to refer sample [Flippy Config](../sample/sample.yaml)
 

