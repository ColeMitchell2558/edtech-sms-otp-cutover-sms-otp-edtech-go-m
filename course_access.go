package edtechotp

import (
	"context"
	"time"
)

type CodeVerifier interface {
	VerifyCode(ctx context.Context, phone, code, requestID string) error
}

type CourseDelivery struct {
	CourseID  string
	LearnerID string
	Delivered bool
	Deadline  time.Time
}

type EducatorEvent struct {
	CourseID  string    `json:"course_id"`
	LearnerID string    `json:"learner_id"`
	Decision  string    `json:"decision"`
	Occurred  time.Time `json:"occurred_at"`
}

type LoginDecision struct {
	Access string        `json:"access"`
	Report EducatorEvent `json:"educator_report"`
}

type LoginService struct {
	Verifier CodeVerifier
	Now      func() time.Time
}

func (s LoginService) VerifyLearner(ctx context.Context, delivery CourseDelivery, phone, code, requestID string) (LoginDecision, error) {
	if err := s.Verifier.VerifyCode(ctx, phone, code, requestID); err != nil {
		return LoginDecision{}, err
	}

	now := s.Now().UTC()
	access := "open"
	if !delivery.Delivered {
		access = "pending_delivery"
	} else if now.After(delivery.Deadline) {
		access = "deadline_passed"
	}
	return LoginDecision{
		Access: access,
		Report: EducatorEvent{
			CourseID: delivery.CourseID, LearnerID: delivery.LearnerID,
			Decision: access, Occurred: now,
		},
	}, nil
}
