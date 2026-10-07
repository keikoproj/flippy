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

4. (Optional) Enable staggered restarts<br>
   By default Flippy restarts all matched objects across all namespaces back to back. On clusters where
   related namespaces must not rotate at the same time (for example a multi-shard tenant), set
   `FLIPPY_STAGGER_NAMESPACES` to `true` on the Flippy container:
    ```
    env:
      - name: FLIPPY_STAGGER_NAMESPACES
        value: "true"
      - name: FLIPPY_STAGGER_MAX_PARALLEL_GROUPS   # optional, default 5
        value: "5"
    ```
   When enabled:
   - Namespaces are grouped by the `iks.intuit.com/service-asset-alias` **annotation** on the Namespace.
     Namespaces without that annotation each form their own group.
   - Groups are processed in parallel, at most `FLIPPY_STAGGER_MAX_PARALLEL_GROUPS` at a time (default 5;
     unset, invalid or non-positive values use the default).
   - Within a group, namespaces are processed one at a time in sorted order.
   - Within a namespace, all Deployments are processed first, then all Rollouts, one object at a time.
     Objects that were healthy before the restart are waited on until they report healthy or exhaust their
     retries before the next object is restarted. The wait is forced even if
     `RestartObjects[].StatusCheckConfig.CheckStatus` is `false`; `MaxRetry` and `RetryDuration` are taken
     from that config when set, otherwise default to 10 retries every 30 seconds. Flippy pauses one
     `RetryDuration` after each restart before the first status poll so a stale "Healthy" status cannot release
     the gate. If the restart command itself fails (for example the object was deleted) the wait is skipped.
   - Objects that were already unhealthy before Flippy touched them are restarted without waiting, as in
     the default mode.
   - `PostFilterRestarts` (for example `istio-ingressgateway`) still run first, before any group starts.

   Any value of `FLIPPY_STAGGER_NAMESPACES` other than `true`/`1` (including unset or invalid) leaves the
   default behavior unchanged. A staggered run can keep a single reconcile busy for a long time; this is
   expected. Enable it only on clusters that need it.

<HR>

Feel free to refer sample [Example deployment](../sample/deployment.yaml).

Feel free to refer sample [Flippy Config](../sample/sample.yaml)
 

