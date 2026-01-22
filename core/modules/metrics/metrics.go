package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metrics for observability
var (
	// User management metrics
	UserCreationCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "admin_user_creation_total",
		Help: "Total number of users created by admins",
	})

	UserDeletionCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_user_deletion_total",
		Help: "Total number of users deleted by admins",
	}, []string{"reason"})

	TokenRevocationCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "token_revocation_total",
		Help: "Total number of token revocations",
	}, []string{"trigger"})
)
