package peerwatch

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"log"
	"net/http"
	"os"
	"time"
)

type (
	PeerPodDiscoverer interface {
		Run(ctx context.Context)
		List(w http.ResponseWriter, r *http.Request)
	}

	peerPodDiscoverer struct {
		kubeClient *kubernetes.Clientset
		urlSet     UrlSet
		thisIP     string
		informer   cache.SharedIndexInformer
	}
)

func NewPeerPodDiscoverer() PeerPodDiscoverer {
	ctx := context.Background()

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	thisIP := os.Getenv("POD_IP")
	nameSpace := os.Getenv("NAMESPACE")
	appName := os.Getenv("DEPLOYMENT_NAME")

	log.Println("IP: " + thisIP)
	log.Println("NAMESPACE: " + nameSpace)
	log.Println("APP_NAME: " + appName)

	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", appName),
	}

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return kubeClient.CoreV1().Pods(nameSpace).List(ctx, listOptions)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return kubeClient.CoreV1().Pods(nameSpace).Watch(ctx, listOptions)
			},
		},
		nil,
		time.Second,
		cache.Indexers{},
	)

	return &peerPodDiscoverer{
		kubeClient: kubeClient,
		urlSet:     make(UrlSet),
		thisIP:     thisIP,
		informer:   informer,
	}
}

func (d peerPodDiscoverer) Run(ctx context.Context) {
	d.informer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    d.onAdd,
			DeleteFunc: d.onDelete,
			UpdateFunc: d.onUpdate,
		},
	)

	stopper := make(chan struct{})
	go d.informer.Run(stopper)
	select {
	case <-ctx.Done():
		close(stopper)
	}
}

func (d peerPodDiscoverer) onAdd(obj interface{}) {
	pod := obj.(*v1.Pod)
	if isPodReady(pod) && pod.Status.PodIP != d.thisIP {
		d.urlSet[pod.Status.PodIP] = fmt.Sprint(pod.Status.Conditions)
	}
}

func (d peerPodDiscoverer) onDelete(obj interface{}) {
	pod := obj.(*v1.Pod)
	delete(d.urlSet, pod.Status.PodIP)
}

func (d peerPodDiscoverer) onUpdate(oldObj, newObj interface{}) {
	newPod := newObj.(*v1.Pod)
	if isPodReady(newPod) && newPod.Status.PodIP != d.thisIP {
		d.urlSet[newPod.Status.PodIP] = fmt.Sprint(newPod.Status.Conditions)
	}
}

func (d peerPodDiscoverer) List(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("List Informer"))
	_, _ = w.Write([]byte(fmt.Sprintf("%v \n", d.urlSet.String())))
	w.WriteHeader(http.StatusOK)
}
