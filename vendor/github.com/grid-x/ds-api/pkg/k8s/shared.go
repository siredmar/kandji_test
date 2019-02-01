package k8s

import (
	"time"

	cache "k8s.io/client-go/tools/cache"
)

// Informer represents an informer as used by K8s
type Informer interface {
	AddEventHandlerWithResyncPeriod(cache.ResourceEventHandler, time.Duration)
}
