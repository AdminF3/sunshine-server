package services

import (
	"net/mail"
	"testing"

	"stageai.tech/sunshine/sunshine/config"
	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
	"github.com/matcornic/hermes/v2"
)

func TestSendToFile(t *testing.T) {
	var (
		cfg    = config.Load()
		mailer = NewMailer(cfg.General, cfg.Mail, SendToFile)
		to     = mail.Address{
			Name:    "John Doe",
			Address: "john_doe@example.com",
		}

		h = hermes.Email{
			Body: hermes.Body{
				Name: to.Name,
				Intros: []string{
					"Welcome to Sunshine!",
				},
				Outros: []string{
					"Bye",
				},
			},
		}
	)

	if err := mailer.Send([]mail.Address{to}, "Welcome", h); err != nil {
		t.Fatal(err)
	}
}

func TestNewUserEmail(t *testing.T) {
	cfg := config.Load()
	mailer := NewMailer(cfg.General, cfg.Mail, SendToFile)
	u := models.User{
		Name:  "John Doe",
		Email: "john@doe.org",
	}

	NewUserEmail(mailer, u, uuid.New())
}
