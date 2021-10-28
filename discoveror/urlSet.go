package discoveror

import (
	"fmt"
	"sort"
)

const Port = 5000

type UrlSet map[string]string
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
