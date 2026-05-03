package event

import (
	"encoding/json"

	"github.com/IsFariza/ap2-Message-Queue/doctor-service/internal/model"
	"github.com/nats-io/nats.go"
)

type DoctorPublisher struct {
	nc *nats.Conn
}

func NewDoctorPublisher(nc *nats.Conn) *DoctorPublisher {
	return &DoctorPublisher{nc: nc}
}

func (p *DoctorPublisher) PublishDoctorCreated(doc *model.Doctor) error {
	// 1. Define the subject (the "channel")
	subject := "doctor.created"

	// 2. Convert your doctor data to JSON bytes
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	// 3. Send it to NATS
	return p.nc.Publish(subject, data)
}
