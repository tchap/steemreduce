package notifications

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const Id = "notifications"

const DataDirectoryEnvironmentKey = "STEEMREDUCE_PARAMS_DATA_DIR"

var DefaultDataDirectoryPath = filepath.Join("steemreduce_data", Id)

const ConfigFilename = "config.yml"

type Config struct {
	Watch struct {
		Stories      WatchStoriesConfig      `yaml:"stories"`
		StoryVotes   WatchStoryVotesConfig   `yaml:"story_votes"`
		Comments     WatchCommentsConfig     `yaml:"comments"`
		CommentVotes WatchCommentVotesConfig `yaml:"comment_votes"`
	} `yaml:"watch"`
	EnabledNotifications []string               `yaml:"enabled_notifications"`
	Command              *CommandNotifierConfig `yaml:"command"`
	Email                *EmailNotifierConfig   `yaml:"email"`
	Slack                *SlackNotifierConfig   `yaml:"slack"`
}

func (config *Config) Validate() error {
	var ok bool
	for _, v := range config.EnabledNotifications {
		switch v {
		case "command":
			if config.Command == nil {
				return errors.New("key not set: command")
			}
			ok = true
		case "email":
			if config.Email == nil {
				return errors.New("key not set: email")
			}
			ok = true
		case "slack":
			if config.Slack == nil {
				return errors.New("key not set: slack")
			}
			ok = true
		default:
			return errors.New("enabled_notifications: unknown notifier: " + v)
		}
	}
	if !ok {
		return errors.New("enabled_notifications: no known notifier specified")
	}
	return nil
}

func loadConfig() (*Config, error) {
	// Get the data directory path from the environment.
	dataDirectoryPath := os.Getenv(DataDirectoryEnvironmentKey)
	if dataDirectoryPath == "" {
		dataDirectoryPath = DefaultDataDirectoryPath
	}

	// Open the state file.
	configPath := filepath.Join(dataDirectoryPath, ConfigFilename)
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Unmarshal the state data.
	var config Config
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	// Validate.
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Return the config object.
	return &config, nil
}
