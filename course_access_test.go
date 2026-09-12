package edtechotp

import (
	"context"
	"errors"
	"testing"
	"time"
)

type verifierStub struct{ err error }

func (v verifierStub) VerifyCode(context.Context, string, string, string) error { return v.err }

func TestVerifyLearnerDecision(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		delivery  CourseDelivery
		verifyErr error
		want      string
		wantErr   bool
	}{
		{"active course opens", CourseDelivery{Delivered: true, Deadline: now.Add(time.Hour)}, nil, "open", false},
		{"undelivered course waits", CourseDelivery{Delivered: false, Deadline: now.Add(time.Hour)}, nil, "pending_delivery", false},
		{"late learner is reported", CourseDelivery{Delivered: true, Deadline: now.Add(-time.Minute)}, nil, "deadline_passed", false},
		{"rejected code makes no decision", CourseDelivery{Delivered: true, Deadline: now.Add(time.Hour)}, errors.New("code rejected"), "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := LoginService{Verifier: verifierStub{tc.verifyErr}, Now: func() time.Time { return now }}
			got, err := svc.VerifyLearner(context.Background(), tc.delivery, "+15550102030", "123456", "login-42")
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tc.wantErr)
			}
			if got.Access != tc.want {
				t.Fatalf("access = %q, want %q", got.Access, tc.want)
			}
			if !tc.wantErr && got.Report.Decision != tc.want {
				t.Fatalf("report decision = %q, want %q", got.Report.Decision, tc.want)
			}
		})
	}
}
