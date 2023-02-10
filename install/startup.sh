#!/bin/sh

echo "Wait for Kubernetes context to be set and ready"

max_retry_count=5
kube_system_pod_count=$(kubectl get pod -n kube-system | wc -l)
retry_count=1


while [[ $kube_system_pod_count -le 0 ]] && [[ $retry_count -le $max_retry_count ]]
do
    sleep 3
    kube_system_pod_count=$(kubectl get pod -n kube-system | wc -l)
    retry_count=$(($retry_count+1))
done

if [[ $kube_system_pod_count -le 0 ]]
then
    echo "Failed to get Kubernetes context"
    exit 1
fi

/workspace/manager &