package peerwatch

import (
	"context"
	"errors"
	"fmt"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"os"
)

type (
	PeerPodDiscoverer interface {
		Init(
			ctx context.Context,
		) ([]string, error)
		GetPods() podSet
	}

	peerPodDiscoverer struct {
		kubeClient *kubernetes.Clientset
		podSet     podSet
		thisIP     string
	}
)

func NewPeerPodDiscoverer() PeerPodDiscoverer {
	// Setup Kube api connection, using the in-cluster config. This assumes the app is running in a pod.
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return &peerPodDiscoverer{
		kubeClient: kubeClient,
		podSet:     make(podSet),
	}
}

// Init initializes the peerwatch library, returning the initial set of pod ips
// and then continually monitors for changes, notifying notifyFunc whenever a pod
// change occurs.
//
// myIp is the IP of the current pod
// listOptions will be used in the calls to Kubernetes API, to filter to desired pods (e.g. by LabelSelector)
// f is a NotifyFunc that lets you do whatever you want with the incoming pod change events. Note this will be called in goroutines so should include thread-safe logic.
// debugMode controls whether to log debug messages or not
func (d peerPodDiscoverer) Init(
	ctx context.Context,
) ([]string, error) {

	// TODO create config for all env
	d.thisIP = os.Getenv("POD_IP")
	// TODO get app name from ENV
	listOptions := metav1.ListOptions{LabelSelector: "app=k8s-peer-pod-discoverer"}

	// Fetch initial pods from API
	err := d.initialPods(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("could not get initial pod list: %v", err)
	}
	if len(d.podSet) <= 0 {
		return nil, errors.New("no pods detected, not even self")
	}
	podIps := d.podSet.Keys()

	// Start monitoring for pod transitions, to keep pool up to date
	go d.monitorPodState(ctx, listOptions, d.podSet)

	return podIps, nil
}

func (d peerPodDiscoverer) GetPods() podSet {
	panic(nil)
}

func (d peerPodDiscoverer) initialPods(
	ctx context.Context,
	listOptions metav1.ListOptions,
) error {
	pods, err := d.kubeClient.CoreV1().Pods("default").List(ctx, listOptions)
	if err != nil {
		return err
	}
	log.Printf("Log1: %s", err)

	d.podSet[d.thisIP] = true
	log.Printf("Log2: %s", d.podSet.String())
	for _, pod := range pods.Items {
		podIp := pod.Status.PodIP
		if isPodReady(&pod) && podIp != d.thisIP {
			d.podSet[podIp] = true
		}
	}
	log.Printf("Log3: %s", d.podSet.String())

	return nil
}

func (d peerPodDiscoverer) monitorPodState(
	ctx context.Context,
	listOptions metav1.ListOptions,
	initialPods podSet,
) {
	// When a kube pod is ADDED or DELETED, it goes through several changes which issue MODIFIED events.
	// By watching these MODIFIED events for times when we see a given podIp associated with a Pod READY condition
	// set to true or false, we can keep track of all pod ip addresses which are ready to receive connections.
	d.podSet = initialPods
	debugLogf("Initial pod list = %v", d.podSet)
	log.Printf("Log4: %s", d.podSet.String())

	// begin watch API call
	watchInterface, err := d.kubeClient.CoreV1().Pods("default").Watch(ctx, listOptions)
	if err != nil {
		debugLogf("WARNING: error watching pods: %v", err)
		return
	}

	// React to watch result channel
	ch := watchInterface.ResultChan()
	for event := range ch {
		pod, ok := event.Object.(*v1.Pod)
		if !ok {
			debugLogf("WARNING: got non-pod object from pod watching: %v", event.Object)
			continue
		}

		podName := pod.Name
		podIp := pod.Status.PodIP
		podReady := isPodReady(pod)
		log.Printf("Log4: %s", podIp)

		// Log raw event stream to debug log
		switch event.Type {
		case "ADDED":
			debugLogf("ADDED pod %s with ip %s. Ready = %v", podName, podIp, podReady)
		case "MODIFIED":
			debugLogf("MODIFIED pod %s with ip %s. Ready = %v", podName, podIp, podReady)
		case "DELETED":
			debugLogf("DELETED pod %s with ip %s. Ready = %v", podName, podIp, podReady)
		}

		// Main events we care about: MODIFIED including a PodIp other than current pod's IP
		if event.Type == "MODIFIED" && podIp != "" && podIp != d.thisIP {
			if podReady && !d.podSet[podIp] {
				debugLogf("Newly ready pod %s @ %s", podName, podIp)
				d.podSet[podIp] = true
				d.podSet[GetPodUrl(podIp)] = true
			} else if !podReady && d.podSet[podIp] {
				debugLogf("Newly disappeared pod %s @ %s", podName, podIp)
				delete(d.podSet, podIp)
			} else {
				continue // no change to pod list
			}
		}
	}
}
