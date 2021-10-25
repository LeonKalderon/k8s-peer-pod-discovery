package peerwatch

import (
    "fmt"
    v1 "k8s.io/api/core/v1"
    "sort"
)

// podSet will hold set of ready pods' ip addresses
type podSet map[string]bool

func (podSet podSet) Keys() []string {
    keys := make([]string, len(podSet))
    i := 0
    for key := range podSet {
        keys[i] = key
        i++
    }
    sort.Strings(keys)
    return keys
}

func (podSet podSet) String() string {
    return fmt.Sprintf("%v", podSet.Keys())
}

func isPodReady(pod *v1.Pod) bool {
    for _, condition := range pod.Status.Conditions {
        if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
            return true
        }
    }
    return false
}