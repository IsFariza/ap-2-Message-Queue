package interfaces

import "github.com/IsFariza/ap2-Message-Queue/doctor-service/internal/model"

type DoctorPublisher interface {
	PublishDoctorCreated(doc *model.Doctor) error
}
