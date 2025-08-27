package mailer

import "embed"

const (
	FromName            = "looksaw"
	maxRetries          = 3
	UserWelcomeTemplate = "user_invitation.tmpl"
)

//go:embed "templates"
var FS embed.FS

type Client interface {
	Send(templateFile string,
		username string,
		email string,
		data any,
		isSandbox bool) (int, error)
}
