package main

import (
    "com.github/LeonKalderon/k8s-peer-pod-discovery/peerwatch"
    "context"
    "fmt"
    "log"
    "net/http"
    "sort"
)

const Port = 5000

type UrlSet map[string]bool
func (urlSet UrlSet) Keys() []string {
    keys := make([]string, len(urlSet))
    i := 0
    for key := range urlSet {
        keys[i] = key
        i++
    }
    sort.Strings(keys)
    return keys
}
func (urlSet UrlSet) String() string {
    return fmt.Sprintf("%v", urlSet.Keys())
}

var urlSet UrlSet

func Index(w http.ResponseWriter, _ *http.Request) {
    w.WriteHeader(http.StatusOK)
}
//
//func logRequest(handler http.Handler) http.Handler {
//    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//        log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
//        handler.ServeHTTP(w, r)
//    })
//}

func main() {
    ctx := context.Background()
    urlSet = make(UrlSet)

    discoverer := peerwatch.NewPeerPodDiscoverer()
    initialPeers, err := discoverer.Init(
        ctx,
    )
    if err != nil {
       // Setup groupcache with just self as peer
       log.Printf("WARNING: error getting initial pods: %v", err)
       url := fmt.Sprintf("http://0.0.0.0:%d", Port)
       urlSet[url] = true
    } else {
       for _, ip := range initialPeers {
           urlSet[peerwatch.GetPodUrl(ip)] = true
       }
    }
    log.Printf("logMain3: readiness")

    http.HandleFunc("/", Index)
    log.Printf("Listening on port %d...", Port)
    if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", Port), http.DefaultServeMux); err != nil {
        log.Fatalf("error in ListenAndServe: %s", err)
    }
}
