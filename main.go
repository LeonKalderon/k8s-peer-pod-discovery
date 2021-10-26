package main

import (
    "com.github/LeonKalderon/k8s-peer-pod-discovery/peerwatch"
    "context"
    "fmt"
    "log"
    "net/http"
)

func Index(w http.ResponseWriter, _ *http.Request) {
    w.WriteHeader(http.StatusOK)
}

const Port = 5000

func main() {
    ctx := context.Background()
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    discoveror := peerwatch.NewPeerPodDiscoverer()
    go discoveror.Run(ctx)
    
    http.HandleFunc("/readinessProbe", Index)
    http.HandleFunc("/livenessProbe", Index)
    http.HandleFunc("/informer/list", discoveror.List)
    log.Printf("Listening on port %d...", Port)
    if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", Port), http.DefaultServeMux); err != nil {
        log.Fatalf("error in ListenAndServe: %s", err)
    }
    ctx.Done()
}
