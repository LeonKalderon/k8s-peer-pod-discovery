package main

import (
    "context"
    "fmt"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/watch"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/cache"
    "log"
    "net/http"
    "time"
)

const Port = 5000

func Index(w http.ResponseWriter, _ *http.Request) {
    w.WriteHeader(http.StatusOK)
}

func onAdd(obj interface{}) {
    //log.Println("ADDED")
    //pod := obj.(metav1.Object)
    //log.Println(pod.GetName())
}

func onDelete(obj interface{}) {
    //log.Println("DELETED")
    //pod := obj.(metav1.Object)
    //log.Println(pod.GetName())
}

func onUpdate(oldObj, newObj interface{}) {
    //log.Println("UPDATED")
    //pod := newObj.(metav1.Object)
    //log.Println(pod.GetName())
}

var informer cache.SharedInformer

func main() {
    ctx := context.Background()

    config, err := rest.InClusterConfig()
    if err != nil {
        panic(err)
    }
    kubeClient, err := kubernetes.NewForConfig(config)
    if err != nil {
        panic(err)
    }
    nameSpace := "default"
    listOptions := metav1.ListOptions{LabelSelector: "app=k8s-peer-pod-discovery"}
    informer = cache.NewSharedIndexInformer(
        &cache.ListWatch{
            ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
                return kubeClient.CoreV1().Pods(nameSpace).List(ctx, listOptions)
            },
            WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
                return kubeClient.CoreV1().Pods(nameSpace).Watch(ctx, listOptions)
            },
        },
        nil,
        time.Nanosecond,
        cache.Indexers{},
    )

    informer.AddEventHandler(
        cache.ResourceEventHandlerFuncs{
            AddFunc:    onAdd,
            DeleteFunc: onDelete,
            UpdateFunc: onUpdate,
        },
    )

    stopper := make(chan struct{})
    defer close(stopper)
    go informer.Run(stopper)

    http.HandleFunc("/readinessProbe", Index)
    http.HandleFunc("/informer/list", List)
    log.Printf("Listening on port %d...", Port)
    if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", Port), http.DefaultServeMux); err != nil {
        log.Fatalf("error in ListenAndServe: %s", err)
    }
}

func List(w http.ResponseWriter, r *http.Request) {
    _, _ = w.Write([]byte("List Informer"))
    _, _ = w.Write([]byte(fmt.Sprintf("%v", informer.GetStore().ListKeys())))
    w.WriteHeader(http.StatusOK)
}
