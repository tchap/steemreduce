package notifications

import (
	"fmt"

	"github.com/go-steem/rpc"
)

type WatchStoriesConfig struct {
	Authors []string `yaml:"authors"`
	Tags    []string `yaml:"tags"`
}

type StoryEvent struct {
	Op      *rpc.CommentOperation
	Content *rpc.Content
}

type StoriesEventMiner struct {
	authors StringSet
	tags    StringSet
}

func newStoriesEventMiner(config *WatchStoriesConfig) *StoriesEventMiner {
	return &StoriesEventMiner{
		authors: MakeStringSet(config.Authors),
		tags:    MakeStringSet(config.Tags),
	}
}

func (miner *StoriesEventMiner) MineEvent(operation *rpc.Operation, content *rpc.Content) interface{} {
	if !content.IsStory() {
		return nil
	}

	op, ok := operation.Body.(*rpc.CommentOperation)
	if !ok {
		return nil
	}

	match := miner.authors.Contains(content.Author)
	if !match {
		for _, tag := range content.JsonMetadata.Tags {
			if miner.tags.Contains(tag) {
				match = true
				break
			}
		}
		if !match {
			return nil
		}
	}

	fmt.Println(" ---> Emitting Story event ...")
	return &StoryEvent{op, content}
}
