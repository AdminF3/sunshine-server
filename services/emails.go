package services

import (
	"fmt"
	"log"
	"net/mail"

	"stageai.tech/sunshine/sunshine/models"

	"github.com/google/uuid"
	"github.com/matcornic/hermes/v2"
)

const (
	welcomeIntro   = "Welcome to Sunshine! We're very excited to have you on board."
	fpIntro        = "You have received this email because a password reset request for your account was received."
	fpOutro        = "If you did not request a password reset, no further action is required on your part."
	fpInstructions = "Click the button below to reset your password:"
	cuInstructions = "Click the button below to confirm your account. This validation link expires in 48 hours:"
	cuOutro        = `
Data Privacy

You have received this email because you would like to register to the SunSHiNE Platform.

The information transmitted is intended for the person or entity to which it is addressed and may contain confidential, privileged or copyrighted material. If you receive this in error, please contact the sender and delete the material from any computer.

We work hard to keep your personal data secure, which includes regularly reviewing our privacy notice. When there’s an important change we’ll remind you to take a look, so you’re aware how we use your data and what your options are. Please review the latest privacy notice.
`
)

func NewUserEmail(mailer Mailer, user models.User, token uuid.UUID) {
	var err = mailer.Send(
		[]mail.Address{{Name: user.Name, Address: user.Email}},
		"Welcome to Sunshine!",
		hermes.Email{
			Body: hermes.Body{
				Name:   user.Name,
				Intros: []string{welcomeIntro},
				Actions: []hermes.Action{
					{
						Instructions: cuInstructions,
						Button: hermes.Button{
							Color:     "#DC4D2F",
							TextColor: "#FFFFFF",
							Text:      "Confirm your account",
							Link: fmt.Sprintf(
								"%s/confirm_user/%s",
								mailer.URL(),
								token,
							),
						},
					},
				},
				Outros: []string{cuOutro},
			},
		},
	)
	log.Printf("Sending email for user %s on create: %v", user.Email, err)
}

func ForgottenPasswordEmail(mailer Mailer, user models.User, token uuid.UUID) {
	var err = mailer.Send(
		[]mail.Address{{Name: user.Name, Address: user.Email}},
		"Forgotten Password",
		hermes.Email{
			Body: hermes.Body{
				Name:   user.Name,
				Intros: []string{fpIntro},
				Actions: []hermes.Action{
					{
						Instructions: fpInstructions,
						Button: hermes.Button{
							Color:     "#DC4D2F",
							TextColor: "#FFFFFF",
							Text:      "Reset your password",
							Link: fmt.Sprintf(
								"%s/reset_password/%s",
								mailer.URL(),
								token,
							),
						},
					},
				},
				Outros: []string{fpOutro},
			},
		},
	)
	log.Printf("Sending email for forgotten password with token %s: %v", token, err)
}
