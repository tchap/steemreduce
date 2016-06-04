package notifications

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

//
// Slack webhook payload
//

type Payload struct {
	Text        string        `json:"text,omitempty"`
	Attachments []*Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Fallback   string   `json:"fallback"`
	Color      string   `json:"color,omitempty"`
	Pretext    string   `json:"pretext,omitempty"`
	AuthorName string   `json:"author_name,omitempty"`
	AuthorLink string   `json:"author_link,omitempty"`
	AuthorIcon string   `json:"author_icon,omitempty"`
	Title      string   `json:"title,omitempty"`
	TitleLink  string   `json:"title_link,omitempty"`
	Text       string   `json:"text,omitempty"`
	Fields     []*Field `json:"fields,omitempty"`
	ImageURL   string   `json:"image_url,omitempty"`
	ThumbURL   string   `json:"thumb_url,omitempty"`
	Footer     string   `json:"footer,omitempty"`
	FooterIcon string   `json:"footer_icon,omitempty"`
	Timestamp  uint64   `json:"ts,omitempty"`
}

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short,omitempty"`
}

//
// Config
//

type SlackNotifierConfig struct {
	WebhookURL string `yaml:"webhook_url"`
}

func (config *SlackNotifierConfig) Validate() error {
	// Make sure the webhook URL is not empty.
	if config.WebhookURL == "" {
		return errors.New("key not set: slack.webhook_url")
	}

	// Make the webhook URL is a valid URL.
	if _, err := url.Parse(config.WebhookURL); err != nil {
		return errors.New("not a valid URL: slack.webhook_url")
	}

	// Cool.
	return nil
}

//
// Notifier
//

type SlackNotifier struct {
	config *SlackNotifierConfig
}

func NewSlackNotifier(config *SlackNotifierConfig) (*SlackNotifier, error) {
	// Validate.
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Done.
	return &SlackNotifier{config}, nil
}

func (notifier *SlackNotifier) DispatchNotification(event interface{}) error {
	var (
		payload *Payload
		err     error
	)
	switch event := event.(type) {
	case *StoryEvent:
		payload, err = renderStoryEvent(event)
	case *StoryVoteEvent:
		payload, err = renderStoryVoteEvent(event)
	case *CommentEvent:
		payload, err = renderCommentEvent(event)
	case *CommentVoteEvent:
		payload, err = renderCommentVoteEvent(event)
	}
	if err != nil {
		return err
	}

	return notifier.send(payload)
}

func (notifier *SlackNotifier) send(payload *Payload) error {
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return err
	}

	_, err := http.Post(notifier.config.WebhookURL, "application/json", &body)
	return err
}

//
// Rendering
//

func makeMessage(attachment *Attachment) *Payload {
	return &Payload{
		Attachments: []*Attachment{attachment},
	}
}

// StoryEvent

func renderStoryEvent(event *StoryEvent) (*Payload, error) {
	c := event.Content
	r := bufio.NewReader(strings.NewReader(c.Body))

	summary, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}

	return makeMessage(&Attachment{
		Fallback:  fmt.Sprintf(`@%v has published "%v".`, c.Author, c.Title),
		Color:     "#00C957",
		Pretext:   fmt.Sprintf("@%v has published or updated a story.", c.Author),
		Title:     c.Title,
		TitleLink: "https://steemit.com" + c.URL,
		Fields: []*Field{
			{
				Title: "Summary",
				Value: summary,
			},
			{
				Title: "Tags",
				Value: fmt.Sprintf("%v", c.JsonMetadata.Tags),
			},
		},
		ThumbURL: "https://steemit.com/images/favicons/favicon-96x96.png",
	}), nil
}

// StoryVoteEvent

func renderStoryVoteEvent(event *StoryVoteEvent) (*Payload, error) {
	o := event.Op
	c := event.Content

	evt := fmt.Sprintf("@%v cast a vote on a story by @%v.", o.Voter, o.Author)

	return makeMessage(&Attachment{
		Fallback:  evt,
		Color:     "#BDFCC9",
		Pretext:   evt,
		Title:     c.Title,
		TitleLink: "https://steemit.com" + c.URL,
		Fields: []*Field{
			{
				Title: "Vote Weight",
				Value: fmt.Sprintf("%v", o.Weight),
				Short: true,
			},
			{
				Title: "Story Pending Payout",
				Value: c.PendingPayoutValue,
				Short: true,
			},
		},
	}), nil
}

// CommentEvent

func renderCommentEvent(event *CommentEvent) (*Payload, error) {
	c := event.Content

	commentLines := make([]string, 0, 5)
	scanner := bufio.NewScanner(strings.NewReader(c.Body))
	for scanner.Scan() {
		commentLines = append(commentLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	extractLines := commentLines
	if len(extractLines) > 5 {
		extractLines = extractLines[:5]
	}

	extract := strings.Join(extractLines, "\n")
	if len(commentLines) > 5 {
		extract += fmt.Sprintf("\n<https://steemit.com%v|Read more...>", c.URL)
	}

	evt := fmt.Sprintf("@%v commented on @%v/%v", c.Author, c.ParentAuthor, c.ParentPermlink)
	pre := fmt.Sprintf("@%v <https://steemit.com%v|commented> on @%v/%v",
		c.Author, c.URL, c.ParentAuthor, c.ParentPermlink)

	return makeMessage(&Attachment{
		Fallback: evt,
		Color:    "#FF9912",
		Pretext:  pre,
		Fields: []*Field{
			{
				Title: "Comment Body",
				Value: extract,
			},
		},
	}), nil
}

// CommentVoteEvent

func renderCommentVoteEvent(event *CommentVoteEvent) (*Payload, error) {
	o := event.Op
	c := event.Content

	evt := fmt.Sprintf("@%v cast a vote on comment @%v/%v", o.Voter, o.Author, o.Permlink)

	return makeMessage(&Attachment{
		Fallback:  evt,
		Color:     "#FFEBCD",
		Pretext:   evt,
		Title:     fmt.Sprintf("@%v/%v", c.Author, c.Permlink),
		TitleLink: "https://steemit.com" + c.URL,
		Fields: []*Field{
			{
				Title: "Vote Weight",
				Value: fmt.Sprintf("%v", o.Weight),
				Short: true,
			},
			{
				Title: "Comment Pending Payout",
				Value: c.PendingPayoutValue,
				Short: true,
			},
		},
	}), nil
}
