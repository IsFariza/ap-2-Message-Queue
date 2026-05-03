package grpc

import (
	"time"

	pb "github.com/IsFariza/ap2-Message-Queue/appointment-service/appointment_proto"
	"github.com/IsFariza/ap2-Message-Queue/appointment-service/internal/model"
)

func toDomain(req *pb.CreateAppointmentRequest) *model.Appointment {
	return &model.Appointment{
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		DoctorID:    req.GetDoctorId(),
	}
}
func toProto(appt *model.Appointment) *pb.AppointmentResponse {
	return &pb.AppointmentResponse{
		Id:          appt.ID,
		Title:       appt.Title,
		Description: appt.Description,
		DoctorId:    appt.DoctorID,
		Status:      string(appt.Status),
		CreatedAt:   appt.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   appt.UpdatedAt.Format(time.RFC3339),
	}
}
