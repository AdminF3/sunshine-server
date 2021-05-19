package services

import (
	"fmt"
	"net/mail"
	"net/smtp"
	"os"
	"path"
	"strings"
	"time"

	"stageai.tech/sunshine/sunshine/config"

	"github.com/google/uuid"
	"github.com/jordan-wright/email"
	"github.com/matcornic/hermes/v2"
)

const copyrightFmt = "Copyright @ %d %s. All rights reserved."

// SendMail is a stateless function that sends email with given settings.
type SendMail func(*email.Email, string, smtp.Auth) error

// Mailer send emails.
type Mailer interface {
	// URL returns the base URL
	URL() string

	// Send email to given recipients using settings in the Mailer.
	Send(to []mail.Address, subject string, he hermes.Email) error
}

type mailer struct {
	auth smtp.Auth
	send SendMail
	gen  hermes.Hermes
	from string
	host string
}

// NewMailer creates new Mailer value.
func NewMailer(general config.General, mail config.Mail, send SendMail) Mailer {
	return mailer{
		auth: smtp.PlainAuth("",
			mail.Username,
			mail.Password,
			mail.Host,
		),
		gen: hermes.Hermes{
			Theme: new(Custom),
			Product: hermes.Product{
				Name: general.Name,
				Link: general.URL,
				Logo: general.Logo,
			},
		},
		host: fmt.Sprintf("%s:%d", mail.Host, mail.Port),
		from: mail.From,
		send: send,
	}
}

// URL returns the base URL of mailer.
func (m mailer) URL() string {
	return m.gen.Product.Link
}

// Send email to given recipients with body using settings in the Mailer.
func (m mailer) Send(to []mail.Address, subject string, he hermes.Email) error {
	var recipients = make([]string, len(to))

	m.gen.Product.Copyright = fmt.Sprintf(
		copyrightFmt, time.Now().Year(), m.gen.Product.Name)
	text, err := m.gen.GeneratePlainText(he)
	if err != nil {
		return err
	}

	html, err := m.gen.GenerateHTML(he)
	if err != nil {
		return err
	}

	for i, t := range to {
		recipients[i] = t.String()
	}

	email := &email.Email{
		From:    m.from,
		To:      recipients,
		Subject: subject,
		Text:    []byte(text),
		HTML:    []byte(html),
	}

	return m.send(email, m.host, m.auth)
}

// Send calls email.Email.Send(addr, auth)
func Send(e *email.Email, addr string, auth smtp.Auth) error {
	return e.Send(addr, auth)
}

// SendToFile writes email to a file in given path instead of actually sending it.
func SendToFile(e *email.Email, dir string, _ smtp.Auth) error {
	var (
		filename = fmt.Sprintf("%d_%s.eml", time.Now().Unix(), uuid.New())
		dirname  = strings.Split(dir, ":")[0]
		filepath = path.Join(dirname, filename)
	)

	if err := os.MkdirAll(dirname, os.ModePerm); err != nil {
		return err
	}

	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := e.Bytes()
	if err != nil {
		return err
	}

	_, err = f.Write(b)
	return err
}

//go:generate mockgen -package=mocks -self_package=stageai.tech/sunshine/sunshine/mocks -destination=./../mocks/services_mailer.go -write_package_comment=false stageai.tech/sunshine/sunshine/services Mailer
