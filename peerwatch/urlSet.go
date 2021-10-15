package peerwatch

import (
	"fmt"
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

func GetPodUrl(podIp string) string {
	return fmt.Sprintf("http://%s:%d", podIp, Port)
}