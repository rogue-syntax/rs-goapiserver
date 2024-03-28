package mailgun

import (
	"context"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

type MailGunReq struct {
	Domain       string
	APIKey       string
	Sender       string
	Recip        string
	Subject      string
	Body         string //leave blank for template sending
	TemplateName string
	Vars         map[string]string
}

func SendMailGunTemplate(req *MailGunReq) (string, error) {
	// Create an instance of the Mailgun Client
	mg := mailgun.NewMailgun(req.Domain, req.APIKey)

	sender := req.Sender
	subject := req.Subject
	body := req.Body
	recipient := req.Recip

	// The message object allows you to add attachments and Bcc recipients
	message := mg.NewMessage(sender, subject, body, recipient)
	message.SetTemplate(req.TemplateName)

	for k, v := range req.Vars {
		err := message.AddTemplateVariable(k, v)
		if err != nil {
			return "", err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message with a 10 second timeout
	_, id, err := mg.Send(ctx, message)

	if err != nil {
		return "", err
	}

	return id, nil
}
