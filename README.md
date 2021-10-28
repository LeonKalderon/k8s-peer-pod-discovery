# k8s-peer-pod-discovery

This is a proof of concept project, showing how the Kubernetes `informer` and `cache` API can be used to keep an up-to-date list of all peer pods.

The ultimate aim of this package is to enable the usage of [groupcache](https://github.com/golang/groupcache) inside a kubernetes cluster.

## Running locally

### Prerequisites:
- [Minikube](https://minikube.sigs.k8s.io/docs/start/)
- [Kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm](https://helm.sh/docs/intro/install/)


### Minikube
- Start minikube
```bash
minikube start
```
- Access the UI
```bash
minikube dashboard
```
- Point your terminal to use the docker daemon inside minikube
```bash
eval $(minikube docker-env)
```
- Build the docker image
```bash
docker build -t pod-listener:0.0.1 .
```
- Add permissions to the pods to access k8s-api
```bash
kubectl create clusterrolebinding default-admin --clusterrole cluster-admin --serviceaccount=default:default
```
- Create the deployment
```bash
 helm install k8s-peer-pod-discovery helm-chart/
```
- Scale up the number of pods
```bash
kubectl scale --replicas=3 deployment/k8s-peer-pod-discovery helm-chart/
``` 
- Get the pod IPs
```bash
export NODE_PORT=$(kubectl get --namespace default -o jsonpath="{.spec.ports[0].nodePort}" services peer-aware-groupcache)
export NODE_IP=$(kubectl get nodes --namespace default -o jsonpath="{.items[0].status.addresses[0].address}")
for i in `seq 0 9`; do echo 1234$i; curl http://$NODE_IP:$NODE_PORT/informer/list; done
```
