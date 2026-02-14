package telemetry

import "github.com/prometheus/client_golang/prometheus"

type ServiceMetrics struct {
	Published prometheus.Counter
	Consumed  prometheus.Counter
	Failed    prometheus.Counter
}

func NewServiceMetrics(service string, reg prometheus.Registerer) *ServiceMetrics {
	published := prometheus.NewCounter(prometheus.CounterOpts{
		Name:        "saga_events_published_total",
		Help:        "Total saga events published by service",
		ConstLabels: prometheus.Labels{"service": service},
	})
	consumed := prometheus.NewCounter(prometheus.CounterOpts{
		Name:        "saga_events_consumed_total",
		Help:        "Total saga events consumed by service",
		ConstLabels: prometheus.Labels{"service": service},
	})
	failed := prometheus.NewCounter(prometheus.CounterOpts{
		Name:        "saga_events_failed_total",
		Help:        "Total saga event processing failures",
		ConstLabels: prometheus.Labels{"service": service},
	})

	reg.MustRegister(published, consumed, failed)
	return &ServiceMetrics{Published: published, Consumed: consumed, Failed: failed}
}
