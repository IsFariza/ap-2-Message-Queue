package event

import (
	"encoding/json"

	"github.com/IsFariza/ap2-Message-Queue/appointment-service/internal/model"
	"github.com/IsFariza/ap2-Message-Queue/appointment-service/internal/model/interfaces"
	"github.com/nats-io/nats.go"
)

type appointmentPublisher struct {
	nc *nats.Conn
}

func NewAppointmentPublisher(nc *nats.Conn) interfaces.AppointmentPublisher {
	return &appointmentPublisher{nc: nc}
}
func (p *appointmentPublisher) PublishCreated(appt *model.Appointment) error {
	data, err := json.Marshal(appt)
	if err != nil {
		return err
	}
	return p.nc.Publish("appointments.created", data)
}

func (p *appointmentPublisher) PublishStatusUpdated(id string, oldS, newS model.Status) {

	payload := map[string]interface{}{
		"id":         id,
		"old_status": oldS,
		"new_status": newS,
	}
	data, _ := json.Marshal(payload)
	p.nc.Publish("appointments.status_updated", data)
}
