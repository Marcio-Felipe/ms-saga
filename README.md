# Go Saga Pattern with 4 Microservices, RabbitMQ, Prometheus, and Grafana

This project contains **four Go microservices** orchestrated with an event-driven **Saga Pattern** over RabbitMQ.

## Services

1. **order-service**
   - Creates `order.created` events on a timer.
   - Listens for `shipping.scheduled` (saga completion).
2. **payment-service**
   - Consumes `order.created`.
   - Publishes `payment.completed`.
3. **inventory-service**
   - Consumes `payment.completed`.
   - Publishes `inventory.reserved`.
4. **shipping-service**
   - Consumes `inventory.reserved`.
   - Publishes `shipping.scheduled`.

All services expose:
- `GET /health`
- `GET /metrics` (Prometheus format)

## Architecture

Event flow:

`order.created -> payment.completed -> inventory.reserved -> shipping.scheduled`

RabbitMQ exchange:
- `saga.events` (`topic`)

## Run the full stack

```bash
docker compose up --build
```

Endpoints:
- RabbitMQ UI: http://localhost:15672 (`guest` / `guest`)
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (default `admin` / `admin`)

## Useful Prometheus queries

```promql
saga_events_published_total
saga_events_consumed_total
saga_events_failed_total
```

## Local Go checks

```bash
go test ./...
```
