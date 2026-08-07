package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"decorebator.com/internal/mail"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

type fakeResetPasswordEmailSender struct {
	calls        []string
	deliveryKeys []string
}

func (f *fakeResetPasswordEmailSender) SendResetPasswordEmail(ctx context.Context, email string, deliveryKeys ...string) error {
	f.calls = append(f.calls, email)
	if len(deliveryKeys) != 1 {
		return errors.New("expected one delivery key")
	}
	f.deliveryKeys = append(f.deliveryKeys, deliveryKeys[0])
	if err := ctx.Err(); err != nil {
		return err
	}
	switch email {
	case "provider-failure@example.com":
		return errors.New("provider unavailable")
	case "rollout-disabled@example.com":
		return mail.ErrAuthHardeningWritesDisabled
	case "superseded@example.com":
		return mail.ErrResetDeliverySuperseded
	case "known@example.com", "absent@example.com", "":
		return nil
	default:
		return errors.New("unexpected test address")
	}
}

func TestResetPasswordEmailWorkerTimeout(t *testing.T) {
	if got := (&ResetPasswordEmailWorker{}).Timeout(nil); got != 30*time.Second {
		t.Fatalf("unexpected reset-password email worker timeout: %s", got)
	}
}

func TestResetPasswordEmailWorkerDeliveryOutcomes(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		cancel    bool
		wantError bool
	}{
		{name: "known account", email: "known@example.com"},
		{name: "absent account", email: "absent@example.com"},
		{name: "invalid sentinel", email: ""},
		{name: "provider failure", email: "provider-failure@example.com", wantError: true},
		{name: "rollback cancellation", email: "rollout-disabled@example.com", wantError: true},
		{name: "superseded cancellation", email: "superseded@example.com", wantError: true},
		{name: "cancellation", email: "known@example.com", cancel: true, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sender := &fakeResetPasswordEmailSender{}
			worker := &ResetPasswordEmailWorker{sender: sender}
			ctx, cancel := context.WithCancel(t.Context())
			if test.cancel {
				cancel()
			} else {
				defer cancel()
			}
			err := worker.Work(ctx, &river.Job[ResetPasswordEmailArgs]{
				JobRow: &rivertype.JobRow{ID: 42},
				Args:   ResetPasswordEmailArgs{Email: test.email},
			})
			if test.wantError && err == nil {
				t.Fatal("expected worker error")
			}
			if !test.wantError && err != nil {
				t.Fatalf("unexpected worker error: %v", err)
			}
			if len(sender.calls) != 1 || sender.calls[0] != test.email {
				t.Fatalf("unexpected sender calls: %#v", sender.calls)
			}
			if len(sender.deliveryKeys) != 1 || sender.deliveryKeys[0] != "42" {
				t.Fatalf("unexpected delivery keys: %#v", sender.deliveryKeys)
			}
			if test.email == "rollout-disabled@example.com" {
				var snoozeError *river.JobSnoozeError
				if !errors.As(err, &snoozeError) {
					t.Fatalf("rollout error = %T %v, want JobSnoozeError", err, err)
				}
			}
			if test.email == "superseded@example.com" {
				var cancelError *river.JobCancelError
				if !errors.As(err, &cancelError) {
					t.Fatalf("superseded error = %T %v, want JobCancelError", err, err)
				}
			}
		})
	}
}
