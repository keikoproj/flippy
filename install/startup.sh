#!/bin/sh

echo "Wait for Kubernetes context to be set and ready"

kubectl_command="kubectl get pods -o custom-columns=NAME:metadata.name -n kube-system --no-headers"

kube_system_pod_count=$($kubectl_command | wc -l)
retry_count=1

if [[ $MAX_RETRY_COUNT -eq "" ]]
then
    echo "Setting MAX_RETRY_COUNT to default"
    MAX_RETRY_COUNT=5
fi

if [[ $RETRY_AFTER -eq "" ]]
then
    echo "Setting RETRY_AFTER to default"
    RETRY_AFTER=60
fi

while [[ $kube_system_pod_count -le 0 ]] && [[ $retry_count -le $MAX_RETRY_COUNT ]]
do
    sleep $RETRY_AFTER
    kube_system_pod_count=$($kubectl_command | wc -l)
    echo "Re-Checking Kubernetes context . . . [$retry_count]"
    retry_count=$(($retry_count+1))
done

if [[ $kube_system_pod_count -le 0 ]]
then
    echo "[error] Failed to get Kubernetes context"
    exit 1
else
    echo "[success] Found Kubernetes context."
    kubectl cluster-info
fi

/workspace/manager &