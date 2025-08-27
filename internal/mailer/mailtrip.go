package mailer

import (
	"bytes"
	"errors"
	"html/template"

	gomail "gopkg.in/mail.v2"
)

// mailTrip的客户端
type mailTripClient struct {
	fromEmail string
	apiKey    string
}

// 新建一个MailTrapClient的
func NewMailTrapClient(apiKey string, fromEmail string) (mailTripClient, error) {
	//检验apiKey是否为空
	if apiKey == "" {
		return mailTripClient{}, errors.New("api key is required")
	}
	return mailTripClient{
		apiKey:    apiKey,
		fromEmail: fromEmail,
	}, nil
}

// 发送邮件
func (m mailTripClient) Send(templateFile string, username string, email string, data any, isSandbox bool) (int, error) {
	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return -1, err
	}
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return -1, err
	}
	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return -1, err
	}
	//准备发送邮件
	message := gomail.NewMessage()
	message.SetHeader("From", m.fromEmail)
	message.SetHeader("To", email)
	message.SetHeader("Subject", subject.String())

	message.AddAlternative("text/html", body.String())
	dialer := gomail.NewDialer("live.smtp.mailtrap.io", 587, "api", m.apiKey)
	if err := dialer.DialAndSend(message); err != nil {
		return -1, err
	}
	return 200, nil
}
