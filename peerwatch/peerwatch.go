package peerwatch

import (
    "k8s.io/api/core/v1"
    "log"
)


func debugLogf(format string, v ...interface{}) {
    log.Printf(format, v...)
}

func isPodReady(pod *v1.Pod) bool {
    for _, condition := range pod.Status.Conditions {
        if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
            return true
        }
    }
    return false
}
