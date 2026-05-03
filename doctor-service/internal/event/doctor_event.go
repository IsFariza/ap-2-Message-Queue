package event

import (
	"encoding/json"

	"github.com/IsFariza/ap2-Message-Queue/doctor-service/internal/model"
	"github.com/nats-io/nats.go"
)

type doctorPublisher struct {
	nc *nats.Conn
}

func NewDoctorPublisher(nc *nats.Conn) *doctorPublisher {
	return &doctorPublisher{nc: nc}
}

func (p *doctorPublisher) PublishDoctorCreated(doc *model.Doctor) error {

	subject := "doctor.created"

	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	return p.nc.Publish(subject, data)
}
