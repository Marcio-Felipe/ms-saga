package messaging

import (
	"encoding/json"
	"time"
)

// Event is the payload shared by all services in the saga.
type Event struct {
	SagaID      string            `json:"saga_id"`
	EventType   string            `json:"event_type"`
	Service     string            `json:"service"`
	OrderID     string            `json:"order_id"`
	Payload     map[string]string `json:"payload"`
	OccurredAt  time.Time         `json:"occurred_at"`
	Correlation string            `json:"correlation_id"`
}

func (e Event) Encode() ([]byte, error) {
	return json.Marshal(e)
}

func DecodeEvent(raw []byte) (Event, error) {
	var evt Event
	err := json.Unmarshal(raw, &evt)
	return evt, err
}
