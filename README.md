# k8s-peer-pod-discovery

This is a proof of concept, showing how the Kubernetes `watch` API can be used to keep an up-to-date list of all peer pods.

## Running

```bash
minikube start
minikube dashboard --url
```

```
$ eval $(minikube docker-env)
$ docker build -t pod-listener:0.0.13 .

$ kubectl create clusterrolebinding default-admin --clusterrole cluster-admin --serviceaccount=default:default
$ helm install k8s-peer-pod-discovery helm-chart/
$ helm upgrade k8s-peer-pod-discovery helm-chart/
$ kubectl scale --replicas=3 deployment/k8s-peer-pod-discovery helm-chart/
$ export NODE_PORT=$(kubectl get --namespace default -o jsonpath="{.spec.ports[0].nodePort}" services peer-aware-groupcache)
$ export NODE_IP=$(kubectl get nodes --namespace default -o jsonpath="{.items[0].status.addresses[0].address}")
$ for i in `seq 0 9`; do echo 1234$i; curl http://$NODE_IP:$NODE_PORT/factors?n=1234$i; done
```

## Development

Notes to self about how to publish new versions of this.
