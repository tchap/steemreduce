package notifications

import (
	"bytes"
	"crypto/tls"
	"errors"
	"strings"
	"text/template"

	"gopkg.in/gomail.v2"
)

//
// Templates
//

// StoryEvent
const storyEventSubjectTemplate = `
[Steemit] @{{.Content.Author}} published "{{.Content.Title}}"`

const storyEventBodyTemplate = `
@{{.Content.Author}} has published the following story:<br />
<br />
{{.Content.Title}}<br />
<br />
You can view the story directly on <a href="https://steemit.com{{.Content.URL}}">Steemit</a>.
`

var (
	storyEventSubject = template.Must(template.New("").Parse(
		strings.TrimSpace(storyEventSubjectTemplate)))
	storyEventBody = template.Must(template.New("").Parse(
		strings.TrimSpace(storyEventBodyTemplate)))
)

// StoryVoteEvent
const storyVoteEventSubjectTemplate = `
[Steemit] @{{.Op.Voter}} voted for @{{.Content.Author}}/{{.Content.Permlink}}`

const storyVoteEventBodyTemplate = `
@{{.Op.Voter}} has cast a vote on @{{.Content.Author}}/{{.Content.Permlink}}.<br />
<br />
Weight: {{.Op.Weight}}<br />
Pending payout: {{.Content.PendingPayoutValue}}<br />
<br />
You can view the story directly on <a href="https://steemit.com{{.Content.URL}}">Steemit</a>.
`

var (
	storyVoteEventSubject = template.Must(template.New("").Parse(
		strings.TrimSpace(storyVoteEventSubjectTemplate)))
	storyVoteEventBody = template.Must(template.New("").Parse(
		strings.TrimSpace(storyVoteEventBodyTemplate)))
)

// CommentEvent
const commentEventSubjectTemplate = `
[Steemit] @{{.Content.Author}} commented on @{{.Content.ParentAuthor}}/{{.Content.ParentPermlink}}`

const commentEventBodyTemplate = `
@{{.Content.Author}} commented on @{{.Content.ParentAuthor}}/{{.Content.ParentPermlink}}.<br />
You can view the comment directly on <a href="https://steemit.com{{.Content.URL}}">Steemit</a>.
`

var (
	commentEventSubject = template.Must(template.New("").Parse(
		strings.TrimSpace(commentEventSubjectTemplate)))
	commentEventBody = template.Must(template.New("").Parse(
		strings.TrimSpace(commentEventBodyTemplate)))
)

// CommentVoteEvent
const commentVoteEventSubjectTemplate = `
[Steemit] @{{.Op.Voter}} voted for @{{.Content.Author}}/{{.Content.Permlink}}`

const commentVoteEventBodyTemplate = `
@{{.Op.Voter}} has cast a vote on @{{.Content.Author}}/{{.Content.Permlink}}.<br />
<br />
Weight: {{.Op.Weight}}</br>
Pending payout: {{.Content.PendingPayoutValue}}<br />
<br />
You can view the comment directly on <a href="https://steemit.com{{.Content.URL}}">Steemit</a>.
`

var (
	commentVoteEventSubject = template.Must(template.New("").Parse(
		strings.TrimSpace(commentVoteEventSubjectTemplate)))
	commentVoteEventBody = template.Must(template.New("").Parse(
		strings.TrimSpace(commentVoteEventBodyTemplate)))
)

//
// Config
//

type EmailNotifierConfig struct {
	SMTPServerHost string   `yaml:"smtp_server_host"`
	SMTPServerPort int      `yaml:"smtp_server_port"`
	SMTPUsername   string   `yaml:"smtp_username"`
	SMTPPassword   string   `yaml:"smtp_password"`
	From           string   `yaml:"from"`
	To             []string `yaml:"to"`
}

func (config *EmailNotifierConfig) Validate() error {
	switch {
	case config.SMTPServerHost == "":
		return errors.New("key not set: email.smtp_server_host")
	case config.SMTPServerPort == 0:
		return errors.New("key not set: email.smtp_server_port")
	case config.SMTPUsername == "":
		return errors.New("key not set: email.smtp_username")
	case config.SMTPPassword == "":
		return errors.New("key not set: email.smtp_password")
	case config.From == "":
		return errors.New("key not set: email.from")
	case len(config.To) == 0:
		return errors.New("array empty: email.to")
	default:
		return nil
	}
}

//
// Notifier
//

type EmailNotifier struct {
	config *EmailNotifierConfig
}

func NewEmailNotifier(config *EmailNotifierConfig) (*EmailNotifier, error) {
	// Validate.
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Done.
	return &EmailNotifier{config}, nil
}

func (notifier *EmailNotifier) DispatchNotification(event interface{}) error {
	var (
		subjectTemplate *template.Template
		bodyTemplate    *template.Template
	)
	switch event.(type) {
	case *StoryEvent:
		subjectTemplate = storyEventSubject
		bodyTemplate = storyEventBody
	case *StoryVoteEvent:
		subjectTemplate = storyVoteEventSubject
		bodyTemplate = storyVoteEventBody
	case *CommentEvent:
		subjectTemplate = commentEventSubject
		bodyTemplate = commentEventBody
	case *CommentVoteEvent:
		subjectTemplate = commentVoteEventSubject
		bodyTemplate = commentVoteEventBody
	}

	var subject bytes.Buffer
	if err := subjectTemplate.Execute(&subject, event); err != nil {
		return err
	}

	var body bytes.Buffer
	if err := bodyTemplate.Execute(&body, event); err != nil {
		return err
	}

	return notifier.send(subject.String(), body.String(), "text/html")
}

func (notifier *EmailNotifier) send(
	subject string,
	body string,
	contentType string,
) error {

	config := notifier.config

	msg := gomail.NewMessage()
	msg.SetHeader("From", config.From)
	msg.SetHeader("To", config.To...)
	msg.SetHeader("Subject", subject)
	msg.SetBody(contentType, body)

	dialer := gomail.NewDialer(
		config.SMTPServerHost, config.SMTPServerPort,
		config.SMTPUsername, config.SMTPPassword)

	dialer.TLSConfig = &tls.Config{ServerName: config.SMTPServerHost}
	return dialer.DialAndSend(msg)
}
