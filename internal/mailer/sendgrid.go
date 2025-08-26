package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *sendgrid.Client
}

// 新建mail发送器
func NewSendgrid(apiKey string, fromEmail string) *SendGridMailer {
	client := sendgrid.NewSendClient(apiKey)
	return &SendGridMailer{
		fromEmail: fromEmail,
		apiKey:    apiKey,
		client:    client,
	}
}

// 发送邮件
func (m *SendGridMailer) Send(templateFile string,
	username string,
	email string,
	data any,
	isSandbox bool) error {
	//发邮件的地址
	from := mail.NewEmail(FromName, m.fromEmail)
	//收邮件
	to := mail.NewEmail(username, email)

	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}
	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return err
	}
	message := mail.NewSingleEmail(from, subject.String(), to, "", body.String())
	//设置沙盒模式
	message.SetMailSettings(&mail.MailSettings{
		SandboxMode: &mail.Setting{
			Enable: &isSandbox,
		},
	})
	for i := 0; i < maxRetries; i++ {
		response, err := m.client.Send(message)
		if err != nil {
			log.Printf("Failed to send email to %v attempt &d of %d\n", email, i+1, maxRetries)
			log.Printf("Err is %v", err)
			//退避重试
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		log.Printf("Email send with status code %v", response.StatusCode)
		return nil
	}
	return fmt.Errorf("failed to send email  after %d attempts", maxRetries)
}
