package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"ms-saga/pkg/messaging"
	"ms-saga/pkg/telemetry"
)

func main() {
	service := "shipping-service"
	rabbitURL := getenv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
	port := getenv("HTTP_PORT", "8084")

	reg := prometheus.NewRegistry()
	reg.MustRegister(prometheus.NewGoCollector(), prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	metrics := telemetry.NewServiceMetrics(service, reg)

	bus := messaging.MustConnect(rabbitURL)
	defer bus.Close()

	go startHTTP(port, reg)
	consumeInventoryReserved(bus, metrics, service)
}

func consumeInventoryReserved(bus *messaging.Bus, metrics *telemetry.ServiceMetrics, service string) {
	msgs, err := bus.Consume("shipping-service.inventory-reserved", "inventory.reserved")
	if err != nil {
		log.Fatalf("consume inventory.reserved: %v", err)
	}

	for msg := range msgs {
		evt, err := messaging.DecodeEvent(msg.Body)
		if err != nil {
			metrics.Failed.Inc()
			_ = msg.Nack(false, false)
			continue
		}

		metrics.Consumed.Inc()
		log.Printf("scheduling shipment order=%s saga=%s", evt.OrderID, evt.SagaID)
		time.Sleep(500 * time.Millisecond)

		next := messaging.Event{
			SagaID:      evt.SagaID,
			EventType:   "shipping.scheduled",
			Service:     service,
			OrderID:     evt.OrderID,
			Correlation: evt.Correlation,
			OccurredAt:  time.Now().UTC(),
			Payload: map[string]string{
				"status": "completed",
			},
		}

		if err := bus.Publish("shipping.scheduled", next); err != nil {
			metrics.Failed.Inc()
			_ = msg.Nack(false, true)
			continue
		}
		metrics.Published.Inc()
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
	log.Printf("shipping-service metrics on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func getenv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
