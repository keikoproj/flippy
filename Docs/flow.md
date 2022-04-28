# Flippy Operator Flow

![Block Diagram](FlowDiagram.jpeg)
<HR>
### Description for Flippy Config

##### 1. ProcessFilter:
This section dictates condition(s) on which desire docker image needs to compared.<br>
Example -
```
  ProcessFilter:
    Containers:
      - istio-proxy
    Labels:
      "istio-injection": "enabled"
    Annotations:
      "sidecar.istio.io/inject": "true"
```
This will filter pod(s)/deployment(s) which  has container name `istio-proxy`, contains label `"istio-injection": "enabled"` and has annotation `"sidecar.istio.io/inject": "true"`
##### 2. ImageList
This section dictates desire docker image(s). Flippy will mark deployment(s)/pod(s) for restart if container image doesn't match any of the below.<br>
Example -
  ```
  ImageList:
    - docker.intuit.com/strategic/services/service-mesh/service/proxyv2:mesh-1662wasmpoc-cf2b1
    - docker.intuit.com/strategic/services/service-mesh/service/proxyv2:patch-1.10.42-boo

  ```    
This will filter pod(s)/deployment(s) which doesn't have any of mentioned docker images for condition specified in #1
##### 3. Preconditions:
This section dictates wait condition before processing any restart. <br>
Example -
  ```
  Preconditions:
    - K8S:
        Type: Deployment
        Name: istiod
        Namespace: istio-system
      StatusCheckConfig:
        CheckStatus: true
        MaxRetry: 10
        RetryDuration: 30
  ```
This will wait & check `istiod` deployment under `istio-system` to be healthy before procedding any restart.
##### 4. PostFilterRestarts:
This section dictates action which needs to be perform before restarting any deployment(s) from generated list. <br>
Example -
  ```
  PostFilterRestarts:
    - K8S:
        Type: Deployment
        Name: istio-ingressgateway
        Namespace: istio-system
      StatusCheckConfig:
        CheckStatus: true
        MaxRetry: 10
        RetryDuration: 30
  ```
This will restart `istio-ingressgateway` deployement under `istio-system` namespace before processing any restart(s) from generated list.
##### 4. RestartObjects:
This section dictates what are the kubenetes object needs to be asserted for generating list.<br>
Example -
  ```
  RestartObjects:
    - Type: Deployment
      StatusCheckConfig:
        CheckStatus: false
        MaxRetry: 10
        RetryDuration: 30
    - Type: ArgoRollout
      StatusCheckConfig:
        CheckStatus: false
        MaxRetry: 10
        RetryDuration: 30
  ```
This will watch all Kubernetes Deployment and Argo Rollouts which matches  #1 & #2

 Feel free to refer sample [Flippy Config](../sample/sample.yaml)
