package notifications

import (
	"bytes"
	"errors"
	"os/exec"
	"text/template"
)

//
// Config
//

type CommandNotifierConfig struct {
	Stories      *Command `yaml:"stories"`
	StoryVotes   *Command `yaml:"story_votes"`
	Comments     *Command `yaml:"comments"`
	CommentVotes *Command `yaml:"comment_votes"`
}

type Command struct {
	Name string   `yaml:"name"`
	Args []string `yaml:"args"`
}

func (cmd *Command) Parse() ([]*template.Template, error) {
	cmdTemplate := make([]*template.Template, 0, 1+len(cmd.Args))

	t, err := template.New("").Parse(cmd.Name)
	if err != nil {
		return nil, err
	}
	cmdTemplate = append(cmdTemplate, t)

	for _, arg := range cmd.Args {
		t, err := template.New("").Parse(arg)
		if err != nil {
			return nil, err
		}
		cmdTemplate = append(cmdTemplate, t)
	}

	return cmdTemplate, nil
}

func (config *CommandNotifierConfig) Validate() error {
	var (
		ok  bool
		err error
	)
	checkCommand := func(cmd *Command, path string) {
		if err != nil {
			return
		}
		if cmd != nil {
			if cmd.Name == "" {
				err = errors.New("key not set: " + path)
			}
			ok = true
		}
	}

	checkCommand(config.Stories, "command.stories")
	checkCommand(config.StoryVotes, "command.story_votes")
	checkCommand(config.Comments, "command.comments")
	checkCommand(config.CommentVotes, "command.comment_votes")

	if err != nil {
		return err
	}
	if !ok {
		return errors.New("command notifier enabled, but no command configured")
	}
	return nil
}

//
// Notifier
//

type CommandNotifier struct {
	stories      []*template.Template
	storyVotes   []*template.Template
	comments     []*template.Template
	commentVotes []*template.Template
}

func NewCommandNotifier(config *CommandNotifierConfig) (*CommandNotifier, error) {
	// Validate.
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Parse.
	var err error
	parseCommand := func(cmd *Command) []*template.Template {
		if err != nil {
			return nil
		}
		cmdTemplate, ex := cmd.Parse()
		if ex != nil {
			err = ex
			return nil
		}
		return cmdTemplate
	}

	notifier := &CommandNotifier{
		stories:      parseCommand(config.Stories),
		storyVotes:   parseCommand(config.StoryVotes),
		comments:     parseCommand(config.Comments),
		commentVotes: parseCommand(config.CommentVotes),
	}
	if err != nil {
		return nil, err
	}

	// Done.
	return notifier, nil
}

func (notifier *CommandNotifier) DispatchNotification(event interface{}) error {
	var (
		cmd []string
		err error
	)
	switch event := event.(type) {
	case *StoryEvent:
		cmd, err = renderCommand(notifier.stories, event)
	case *StoryVoteEvent:
		cmd, err = renderCommand(notifier.storyVotes, event)
	case *CommentEvent:
		cmd, err = renderCommand(notifier.comments, event)
	case *CommentVoteEvent:
		cmd, err = renderCommand(notifier.commentVotes, event)
	}
	if err != nil {
		return err
	}
	return exec.Command(cmd[0], cmd[1:]...).Run()
}

func renderCommand(cmdTemplate []*template.Template, context interface{}) ([]string, error) {
	cmd := make([]string, 0, len(cmdTemplate))
	for _, t := range cmdTemplate {
		var buffer bytes.Buffer
		if err := t.Execute(&buffer, context); err != nil {
			return nil, err
		}
		cmd = append(cmd, buffer.String())
	}
	return cmd, nil
}
