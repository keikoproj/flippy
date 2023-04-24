#!/bin/sh

echo "Wait for Kubernetes context to be set and ready"

kubectl_command="kubectl get pods -o custom-columns=NAME:metadata.name --no-headers"

pod_count=$($kubectl_command | wc -l)
retry_count=1

if [[ -z $MAX_RETRY_COUNT ]]
then
    echo "Setting MAX_RETRY_COUNT to default"
    MAX_RETRY_COUNT=5
fi

if [[ -z $RETRY_AFTER ]]
then
    echo "Setting RETRY_AFTER to default"
    RETRY_AFTER=60
fi

while [[ $pod_count -le 0 ]] && [[ $retry_count -le $MAX_RETRY_COUNT ]]
do
    sleep $RETRY_AFTER
    pod_count=$($kubectl_command | wc -l)
    echo "Re-Checking Kubernetes context . . . [$retry_count]"
    retry_count=$(($retry_count+1))
done

if [[ $pod_count -le 0 ]]
then
    echo "[error] Failed to get Kubernetes context"
    exit 1
else
    echo "[success] Found Kubernetes context."
fi

/workspace/manager