package platform

import (
	"net/smtp"
	"context"
	"errors"
	"time"
	"net"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel"
	"github.com/joho/godotenv"
)

type EmailSenderManager struct {
	Host string
	From string
}

var (
	emailManager *EmailSenderManager
)

const EmailManagerName = "email-manager"

func InitMailManager(ctx context.Context) error {
	tr := otel.Tracer(EmailManagerName)
	dialCtx, span := tr.Start(ctx, fmt.Sprintf("%s.ensureBucketExists", EmailManagerName))
	defer span.End()

	if err := godotenv.Load(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("Unable to load env: %s", err.Error()))
		return err
	}

	host := os.Getenv("SMTP_HOST")
	if host == "" {
		err := errors.New("SMTP_HOST environment variable is empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("Unable to load env: %s", err.Error()))
		return err
	}

	span.SetAttributes(attribute.String("messaging.smtp.host", host))

	dial := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dial.DialContext(dialCtx, "tcp", host)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "SMTP server unreachable")
		return fmt.Errorf("could not connect to SMTP server at %s: %w", host, err)
	}
	conn.Close()

	emailManager = &EmailSenderManager{
		Host: host,
		From: "no-reply@sender.com",
	}

	return nil
}

func SendEmail(to []string, msg []byte) error {
	if emailManager == nil {
		return errors.New("Empty manager")
	}

	return smtp.SendMail(emailManager.Host, nil, emailManager.From, to, msg)
}

