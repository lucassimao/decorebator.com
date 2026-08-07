package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"decorebator.com/internal/mail"
	"github.com/riverqueue/river"
)

const ResetPasswordEmailQueue = "reset_password_email"

type ResetPasswordEmailArgs struct {
	Email string `json:"email"`
}

func (ResetPasswordEmailArgs) Kind() string { return "reset_password_email" }

type ResetPasswordEmailWorker struct {
	river.WorkerDefaults[ResetPasswordEmailArgs]
	sender resetPasswordEmailSender
}

type resetPasswordEmailSender interface {
	SendResetPasswordEmail(context.Context, string, ...string) error
}

func NewResetPasswordEmailWorker(mailService *mail.MailService) *ResetPasswordEmailWorker {
	return &ResetPasswordEmailWorker{sender: mailService}
}

func (w *ResetPasswordEmailWorker) Timeout(*river.Job[ResetPasswordEmailArgs]) time.Duration {
	return 30 * time.Second
}

func (w *ResetPasswordEmailWorker) Work(ctx context.Context, job *river.Job[ResetPasswordEmailArgs]) error {
	err := w.sender.SendResetPasswordEmail(ctx, job.Args.Email, strconv.FormatInt(job.ID, 10))
	if errors.Is(err, mail.ErrAuthHardeningWritesDisabled) {
		return river.JobSnooze(time.Minute)
	}
	if errors.Is(err, mail.ErrResetDeliverySuperseded) {
		return river.JobCancel(err)
	}
	return err
}
