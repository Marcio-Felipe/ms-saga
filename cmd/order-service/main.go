package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"ms-saga/pkg/messaging"
	"ms-saga/pkg/telemetry"
)

func main() {
	service := "order-service"
	rabbitURL := getenv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
	port := getenv("HTTP_PORT", "8081")

	reg := prometheus.NewRegistry()
	reg.MustRegister(prometheus.NewGoCollector(), prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	metrics := telemetry.NewServiceMetrics(service, reg)

	bus := messaging.MustConnect(rabbitURL)
	defer bus.Close()

	go startHTTP(port, reg)
	go publishOrders(bus, metrics, service)
	consumeSagaCompletion(bus, metrics)
}

func publishOrders(bus *messaging.Bus, metrics *telemetry.ServiceMetrics, service string) {
	ticker := time.NewTicker(12 * time.Second)
	defer ticker.Stop()

	for {
		sagaID := uuid.NewString()
		orderID := uuid.NewString()
		evt := messaging.Event{
			SagaID:      sagaID,
			EventType:   "order.created",
			Service:     service,
			OrderID:     orderID,
			Correlation: sagaID,
			OccurredAt:  time.Now().UTC(),
			Payload: map[string]string{
				"customer": "sample-customer",
				"amount":   "99.99",
			},
		}
		if err := bus.Publish("order.created", evt); err != nil {
			metrics.Failed.Inc()
			log.Printf("publish order.created failed: %v", err)
		} else {
			metrics.Published.Inc()
			log.Printf("published order.created saga=%s order=%s", sagaID, orderID)
		}
		<-ticker.C
	}
}

func consumeSagaCompletion(bus *messaging.Bus, metrics *telemetry.ServiceMetrics) {
	msgs, err := bus.Consume("order-service.shipping-scheduled", "shipping.scheduled")
	if err != nil {
		log.Fatalf("consume shipping.scheduled: %v", err)
	}

	for msg := range msgs {
		evt, err := messaging.DecodeEvent(msg.Body)
		if err != nil {
			metrics.Failed.Inc()
			_ = msg.Nack(false, false)
			continue
		}
		metrics.Consumed.Inc()
		log.Printf("saga completed for order=%s saga=%s status=%s", evt.OrderID, evt.SagaID, evt.Payload["status"])
		_ = msg.Ack(false)
	}
}

func startHTTP(port string, reg *prometheus.Registry) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	log.Printf("order-service metrics on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func getenv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
